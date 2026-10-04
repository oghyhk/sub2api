package service

import (
	"context"
	"math"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/usagestats"
)

// Anthropic（Claude）用量窗口的展示口径修复，与 antigravity_quota_display.go
// 同一口径：上游 /api/oauth/usage 与 session_window 都是快照窗口，滚动重置后
// 已用记录清零；这里用 usage_logs 的真实滚动成本 + 上游读数校准上限，展示值取
// 上游读数、真实滚动用量、限流钉底三者的最大值。

// splitAnthropicClassCosts 按 Claude 模型子类拆分窗口成本。
func splitAnthropicClassCosts(stats []usagestats.ModelStat) (total, sonnet, fable float64) {
	for _, stat := range stats {
		total += stat.AccountCost
		name := strings.ToLower(stat.Model)
		if strings.Contains(name, "sonnet") {
			sonnet += stat.AccountCost
		}
		if strings.Contains(name, "fable") {
			fable += stat.AccountCost
		}
	}
	return total, sonnet, fable
}

// anthropicScopeClass 把 Claude 命名空间下的模型/限流 key 归类到用量窗口：
// claude:sonnet / claude:fable / claude（通用）；非 claude 命名返回空串。
func anthropicScopeClass(scope string) string {
	name := strings.ToLower(strings.TrimSpace(scope))
	if !strings.HasPrefix(name, "claude") {
		return ""
	}
	switch {
	case strings.Contains(name, "sonnet"):
		return "claude:sonnet"
	case strings.Contains(name, "fable"):
		return "claude:fable"
	default:
		return "claude"
	}
}

// anthropicWindowStampActive 判断目标窗口是否被生效中的模型限流标记钉住：
// 通用 claude 标记钉住全量窗口；子类标记只钉对应子类窗口。
func anthropicWindowStampActive(account *Account, class string) bool {
	if account == nil {
		return false
	}
	return account.activeModelRateLimitResetAt(func(scope string) bool {
		scopeClass := anthropicScopeClass(scope)
		if scopeClass == "" {
			return false
		}
		if class == "claude" {
			return true
		}
		return scopeClass == class
	}) != nil
}

// anthropicWindowDisplayUsed 与 antigravityWindowDisplayUsed 同口径，保留
// 上游的浮点百分比精度。
func anthropicWindowDisplayUsed(providerUsed float64, stampFloor bool, localCost, limit float64) float64 {
	used := providerUsed
	if stampFloor {
		used = 100
	}
	if limit > 0 && localCost > 0 {
		if localUsed := localCost / limit * 100; localUsed > used {
			used = localUsed
		}
	}
	if used < 0 {
		used = 0
	}
	if used > 100 {
		used = 100
	}
	return math.Round(used*10) / 10
}

func (s *AccountUsageService) anthropicRollingClassCosts(ctx context.Context, accountID int64, now time.Time) (total5h, total7d, sonnet7d, fable7d float64) {
	if s == nil || s.usageLogRepo == nil || accountID <= 0 {
		return 0, 0, 0, 0
	}
	if stats, err := s.usageLogRepo.GetModelStatsWithFilters(ctx, now.Add(-antigravityQuotaWindow5h), now, 0, 0, accountID, 0, nil, nil, nil); err == nil {
		total5h, _, _ = splitAnthropicClassCosts(stats)
	}
	if stats, err := s.usageLogRepo.GetModelStatsWithFilters(ctx, now.Add(-antigravityQuotaWindow7d), now, 0, 0, accountID, 0, nil, nil, nil); err == nil {
		total7d, sonnet7d, fable7d = splitAnthropicClassCosts(stats)
	}
	return total5h, total7d, sonnet7d, fable7d
}

// applyAnthropicUsageDisplayOverlay 叠加基于真实滚动用量的展示百分比。返回
// 浅拷贝，不修改入参（调用方可能持有缓存对象）。
func (s *AccountUsageService) applyAnthropicUsageDisplayOverlay(ctx context.Context, usage *UsageInfo, account *Account) *UsageInfo {
	if usage == nil || account == nil || s == nil {
		return usage
	}
	now := time.Now()
	total5h, total7d, sonnet7d, fable7d := s.anthropicRollingClassCosts(ctx, account.ID, now)
	return anthropicComposeDisplayOverlay(usage, account, now, total5h, total7d, sonnet7d, fable7d)
}

type anthropicWindowTarget struct {
	key        antigravityQuotaWindowKey
	progress   *UsageProgress
	localCost  float64
	stampFloor bool
	set        func(*UsageProgress)
}

func anthropicComposeDisplayOverlay(usage *UsageInfo, account *Account, now time.Time, total5h, total7d, sonnet7d, fable7d float64) *UsageInfo {
	if usage == nil || account == nil {
		return usage
	}
	out := *usage

	targets := []anthropicWindowTarget{
		{
			key: antigravityWindowKey5h("claude"), progress: usage.FiveHour, localCost: total5h,
			stampFloor: anthropicWindowStampActive(account, "claude"),
			set:        func(p *UsageProgress) { out.FiveHour = p },
		},
		{
			key: antigravityWindowKey7d("claude"), progress: usage.SevenDay, localCost: total7d,
			stampFloor: anthropicWindowStampActive(account, "claude"),
			set:        func(p *UsageProgress) { out.SevenDay = p },
		},
		{
			key: antigravityWindowKey7d("claude:sonnet"), progress: usage.SevenDaySonnet, localCost: sonnet7d,
			stampFloor: anthropicWindowStampActive(account, "claude:sonnet"),
			set:        func(p *UsageProgress) { out.SevenDaySonnet = p },
		},
		{
			key: antigravityWindowKey7d("claude:fable"), progress: usage.SevenDayFable, localCost: fable7d,
			stampFloor: anthropicWindowStampActive(account, "claude:fable"),
			set:        func(p *UsageProgress) { out.SevenDayFable = p },
		},
	}

	for _, target := range targets {
		if target.progress == nil {
			continue
		}
		providerUsed := target.progress.Utilization
		if limit, ok := antigravityCalibrationLimitFromSample(int(math.Round(providerUsed)), target.localCost); ok {
			antigravityQuotaCalibrationState.observe(account.ID, target.key, limit, now)
		}
		limit := 0.0
		if calibrated, ok := antigravityQuotaCalibrationState.limitFor(account.ID, target.key, now); ok {
			limit = calibrated
		}
		used := anthropicWindowDisplayUsed(providerUsed, target.stampFloor, target.localCost, limit)
		if used <= providerUsed {
			continue
		}
		raised := *target.progress
		raised.Utilization = used
		target.set(&raised)
	}
	return &out
}
