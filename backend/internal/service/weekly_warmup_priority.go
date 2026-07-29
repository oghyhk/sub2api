package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"
)

const (
	weeklyWarmupRefreshInterval = 2 * time.Minute
	// Weekly usage remains in warm-up until the provider reports at least 1%.
	// Keep this threshold centralized so scheduler paths cannot drift apart.
	weeklyWarmupMinimumUsedPercent = 1.0

	antigravityGeminiWeeklyUsedPercentKey = "antigravity_gemini_weekly_used_percent"
	antigravityGeminiWeeklyResetAtKey     = "antigravity_gemini_weekly_reset_at"
	antigravityThirdPartyWeeklyUsedKey    = "antigravity_3p_weekly_used_percent"
	antigravityThirdPartyWeeklyResetAtKey = "antigravity_3p_weekly_reset_at"
	antigravityWeeklyUsageUpdatedAtKey    = "antigravity_weekly_usage_updated_at"
)

func weeklyWarmupUtilization(account *Account, requestedModel string, now time.Time) (float64, bool) {
	if account == nil {
		return 0, false
	}

	switch account.Platform {
	case PlatformOpenAI:
		progress := buildCodexUsageProgressFromExtra(account.Extra, "7d", now)
		if progress == nil {
			return 0, false
		}
		return progress.Utilization, true
	case PlatformAntigravity:
		usedKey, resetKey := antigravityWeeklyExtraKeys(requestedModel)
		usedRaw, ok := account.Extra[usedKey]
		if !ok {
			return 0, false
		}
		utilization := parseExtraFloat64(usedRaw)
		if resetRaw, exists := account.Extra[resetKey]; exists {
			if resetAt, err := parseTime(fmt.Sprint(resetRaw)); err == nil && !now.Before(resetAt) {
				utilization = 0
			}
		}
		return utilization, true
	default:
		return 0, false
	}
}

func isWeeklyWarmupAccount(account *Account, requestedModel string, now time.Time) bool {
	utilization, known := weeklyWarmupUtilization(account, requestedModel, now)
	return known && utilization < weeklyWarmupMinimumUsedPercent
}

// weeklyUsageLess reports whether a has lower known weekly usage than b.
// Known usage is preferred over unknown usage; unknown values never outrank a
// measured value. Callers use this as a tie-breaker after priority/load gates.
func weeklyUsageLess(a, b *Account, requestedModel string, now time.Time) bool {
	aUsage, aKnown := weeklyWarmupUtilization(a, requestedModel, now)
	bUsage, bKnown := weeklyWarmupUtilization(b, requestedModel, now)
	if aKnown != bKnown {
		return aKnown
	}
	return aKnown && aUsage < bUsage
}

func partitionWeeklyWarmupAccounts(accounts []*Account, requestedModel string, now time.Time) ([]*Account, []*Account) {
	warmup := make([]*Account, 0, len(accounts))
	regular := make([]*Account, 0, len(accounts))
	for _, account := range accounts {
		if isWeeklyWarmupAccount(account, requestedModel, now) {
			warmup = append(warmup, account)
			continue
		}
		regular = append(regular, account)
	}
	return warmup, regular
}

func antigravityWeeklyExtraKeys(requestedModel string) (usedPercentKey, resetAtKey string) {
	model := strings.ToLower(strings.TrimSpace(requestedModel))
	if strings.Contains(model, "claude") {
		return antigravityThirdPartyWeeklyUsedKey, antigravityThirdPartyWeeklyResetAtKey
	}
	return antigravityGeminiWeeklyUsedPercentKey, antigravityGeminiWeeklyResetAtKey
}

func antigravityWeeklyUsageExtraUpdates(info *UsageInfo, now time.Time) map[string]any {
	if info == nil || len(info.AntigravityQuota) == 0 {
		return nil
	}

	updates := make(map[string]any, 5)
	writeBucket := func(bucketID, usedKey, resetKey string) {
		quota := info.AntigravityQuota[bucketID]
		if quota == nil {
			return
		}
		updates[usedKey] = quota.Utilization
		if strings.TrimSpace(quota.ResetTime) != "" {
			updates[resetKey] = quota.ResetTime
		}
	}

	writeBucket("gemini-weekly", antigravityGeminiWeeklyUsedPercentKey, antigravityGeminiWeeklyResetAtKey)
	writeBucket("3p-weekly", antigravityThirdPartyWeeklyUsedKey, antigravityThirdPartyWeeklyResetAtKey)
	if len(updates) == 0 {
		return nil
	}
	if info.UpdatedAt != nil {
		now = *info.UpdatedAt
	}
	updates[antigravityWeeklyUsageUpdatedAtKey] = now.UTC().Format(time.RFC3339)
	return updates
}

func (s *GatewayService) scheduleAntigravityWeeklyUsageRefresh(accounts []Account, requestedModel string, now time.Time) {
	if s == nil || s.accountUsageService == nil {
		return
	}
	for i := range accounts {
		account := &accounts[i]
		if account.Platform != PlatformAntigravity {
			continue
		}
		utilization, known := weeklyWarmupUtilization(account, requestedModel, now)
		if known && utilization > 0 {
			continue
		}
		if !s.markWeeklyWarmupRefresh(account.ID, now) {
			continue
		}
		accountID := account.ID
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
			defer cancel()
			if _, err := s.accountUsageService.GetUsage(ctx, accountID, true); err != nil {
				slog.Warn("weekly warm-up usage refresh failed", "account_id", accountID, "error", err)
			}
		}()
	}
}

func (s *GatewayService) markWeeklyWarmupRefresh(accountID int64, now time.Time) bool {
	if s == nil || accountID <= 0 {
		return false
	}
	s.weeklyWarmupRefreshMu.Lock()
	defer s.weeklyWarmupRefreshMu.Unlock()
	if s.weeklyWarmupRefreshAt == nil {
		s.weeklyWarmupRefreshAt = make(map[int64]time.Time)
	}
	if last := s.weeklyWarmupRefreshAt[accountID]; !last.IsZero() && now.Sub(last) < weeklyWarmupRefreshInterval {
		return false
	}
	s.weeklyWarmupRefreshAt[accountID] = now
	return true
}
