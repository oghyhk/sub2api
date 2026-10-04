package service

import (
	"context"
	"math"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/usagestats"
)

// Antigravity 额度窗口的展示口径修复：
//
// Google 的 retrieveUserQuotaSummary / fetchAvailableModels 配额窗口会在滚动
// 重置后清空已用记录（窗口惰性重建、空窗口 resetTime 随探测漂移），因此紧随
// 用量高峰之后，面板会显示 0% 与实际用量不符。这里以本网关 usage_logs 的
// 真实用量（滚动 5h / 7d 成本）为准计算展示百分比，并在上游窗口仍有效时用
// 上游 remainingFraction 反推各账号的窗口额度上限做校准；账号窗口刚重建导致
// 上游读数为 0 时，回退到同层级其他账号的中位上限。模型限流标记（429 写入
// 的 model_rate_limits）仍然把对应家族的 5h 窗口钉在 100% 直到恢复时间。

const (
	antigravityQuotaWindow5h = 5 * time.Hour
	antigravityQuotaWindow7d = 7 * 24 * time.Hour

	// 上游读数只有在窗口内确实有用量时才可校准（排除 0%/100% 边界与噪声）。
	antigravityCalibrationMinUsedPct = 2
	antigravityCalibrationMaxUsedPct = 98

	antigravityCalibrationTTL = 24 * time.Hour
)

type antigravityQuotaWindowKey struct {
	family string // gemini | 3p
	window string // 5h | 7d
}

func antigravityWindowKey5h(family string) antigravityQuotaWindowKey {
	return antigravityQuotaWindowKey{family: family, window: "5h"}
}

func antigravityWindowKey7d(family string) antigravityQuotaWindowKey {
	return antigravityQuotaWindowKey{family: family, window: "7d"}
}

type antigravityLimitSample struct {
	limit      float64
	observedAt time.Time
}

type antigravityQuotaCalibration struct {
	mu         sync.Mutex
	perAccount map[int64]map[antigravityQuotaWindowKey]antigravityLimitSample
}

var antigravityQuotaCalibrationState = &antigravityQuotaCalibration{
	perAccount: make(map[int64]map[antigravityQuotaWindowKey]antigravityLimitSample),
}

func resetAntigravityQuotaCalibrationForTest() {
	antigravityQuotaCalibrationState.mu.Lock()
	defer antigravityQuotaCalibrationState.mu.Unlock()
	antigravityQuotaCalibrationState.perAccount = make(map[int64]map[antigravityQuotaWindowKey]antigravityLimitSample)
}

func (c *antigravityQuotaCalibration) observe(accountID int64, key antigravityQuotaWindowKey, limit float64, now time.Time) {
	if c == nil || accountID <= 0 || limit <= 0 || math.IsNaN(limit) || math.IsInf(limit, 0) {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	slots, ok := c.perAccount[accountID]
	if !ok {
		slots = make(map[antigravityQuotaWindowKey]antigravityLimitSample, 4)
		c.perAccount[accountID] = slots
	}
	slots[key] = antigravityLimitSample{limit: limit, observedAt: now}
}

func (c *antigravityQuotaCalibration) limitFor(accountID int64, key antigravityQuotaWindowKey, now time.Time) (float64, bool) {
	if c == nil {
		return 0, false
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	if slots, ok := c.perAccount[accountID]; ok {
		if sample, ok := slots[key]; ok && now.Sub(sample.observedAt) < antigravityCalibrationTTL {
			return sample.limit, true
		}
	}

	var candidates []float64
	for id, slots := range c.perAccount {
		if id == accountID {
			continue
		}
		sample, ok := slots[key]
		if !ok || now.Sub(sample.observedAt) >= antigravityCalibrationTTL {
			continue
		}
		candidates = append(candidates, sample.limit)
	}
	if len(candidates) == 0 {
		return 0, false
	}
	sort.Float64s(candidates)
	mid := len(candidates) / 2
	if len(candidates)%2 == 0 {
		return (candidates[mid-1] + candidates[mid]) / 2, true
	}
	return candidates[mid], true
}

// isAntigravityGeminiModelName 判断 usage_logs 的请求模型是否属于 Gemini 家族。
func isAntigravityGeminiModelName(model string) bool {
	return strings.HasPrefix(strings.ToLower(strings.TrimSpace(model)), "gemini")
}

// splitAntigravityFamilyCosts 把模型统计按 Gemini / 其他（3p）家族拆成两组成本。
func splitAntigravityFamilyCosts(stats []usagestats.ModelStat) (geminiCost float64, otherCost float64) {
	for _, stat := range stats {
		if isAntigravityGeminiModelName(stat.Model) {
			geminiCost += stat.AccountCost
		} else {
			otherCost += stat.AccountCost
		}
	}
	return geminiCost, otherCost
}

// antigravityWindowDisplayUsed 计算展示用百分比：真实滚动用量（校准上限换算）
// 与上游读数、限流钉底三者取最大值。limit<=0 表示无校准，仅回退上游读数。
func antigravityWindowDisplayUsed(providerUsedPct int, stampFloor bool, localCost, limit float64) int {
	used := float64(providerUsedPct)
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
	return int(math.Round(used))
}

// antigravityCalibrationLimitFromSample 由上游剩余比例反推窗口上限；读数不在
// 可校准区间或本地无成本时返回 false。
func antigravityCalibrationLimitFromSample(providerUsedPct int, localCost float64) (float64, bool) {
	if providerUsedPct <= antigravityCalibrationMinUsedPct || providerUsedPct >= antigravityCalibrationMaxUsedPct {
		return 0, false
	}
	if localCost <= 0 {
		return 0, false
	}
	return localCost / (float64(providerUsedPct) / 100), true
}

func (s *AccountUsageService) antigravityRollingFamilyCosts(ctx context.Context, accountID int64, now time.Time) (gemini5h, other5h, gemini7d, other7d float64) {
	if s == nil || s.usageLogRepo == nil || accountID <= 0 {
		return 0, 0, 0, 0
	}
	stats5h, err := s.usageLogRepo.GetModelStatsWithFilters(ctx, now.Add(-antigravityQuotaWindow5h), now, 0, 0, accountID, 0, nil, nil, nil)
	if err == nil {
		gemini5h, other5h = splitAntigravityFamilyCosts(stats5h)
	}
	stats7d, err := s.usageLogRepo.GetModelStatsWithFilters(ctx, now.Add(-antigravityQuotaWindow7d), now, 0, 0, accountID, 0, nil, nil, nil)
	if err == nil {
		gemini7d, other7d = splitAntigravityFamilyCosts(stats7d)
	}
	return gemini5h, other5h, gemini7d, other7d
}

// antigravityScopeFamily 返回条目所属家族（gemini / 3p），周窗口条目同样归属
// 各自家族，只是对应 7d 窗口。
func antigravityScopeFamily(name string) string {
	if antigravityGeminiScopeKey(name) {
		return "gemini"
	}
	if antigravityClaudeScopeKey(name) {
		return "3p"
	}
	return ""
}

// applyAntigravityQuotaDisplayOverlay 在配额耗尽钉底之上，叠加基于真实滚动
// 用量的展示百分比。返回浅拷贝，不修改入参（调用方可能持有缓存对象）。
func (s *AccountUsageService) applyAntigravityQuotaDisplayOverlay(ctx context.Context, usage *UsageInfo, account *Account) *UsageInfo {
	if usage == nil || account == nil {
		return usage
	}
	usage = applyAntigravityQuotaExhaustionOverlay(usage, account)
	if s == nil {
		return usage
	}
	now := time.Now()
	gemini5hCost, other5hCost, gemini7dCost, other7dCost := s.antigravityRollingFamilyCosts(ctx, account.ID, now)
	return antigravityComposeDisplayOverlay(usage, account, now, gemini5hCost, other5hCost, gemini7dCost, other7dCost)
}

func antigravityComposeDisplayOverlay(usage *UsageInfo, account *Account, now time.Time, gemini5hCost, other5hCost, gemini7dCost, other7dCost float64) *UsageInfo {
	if usage == nil || account == nil {
		return usage
	}
	type windowSample struct {
		key       antigravityQuotaWindowKey
		entry     *AntigravityModelQuota
		localCost float64
	}
	samples := map[antigravityQuotaWindowKey]*windowSample{}
	for name, quota := range usage.AntigravityQuota {
		if quota == nil {
			continue
		}
		family := antigravityScopeFamily(name)
		if family == "" {
			continue
		}
		key := antigravityWindowKey5h(family)
		localCost := gemini5hCost
		if family == "3p" {
			localCost = other5hCost
		}
		if isAntigravityWeeklyWindowKey(name) {
			key = antigravityWindowKey7d(family)
			localCost = gemini7dCost
			if family == "3p" {
				localCost = other7dCost
			}
		}
		sample, ok := samples[key]
		if !ok {
			sample = &windowSample{key: key, localCost: localCost}
			samples[key] = sample
		}
		if sample.entry == nil || quota.Utilization > sample.entry.Utilization {
			entryCopy := *quota
			sample.entry = &entryCopy
		}
	}

	displayUsed := map[antigravityQuotaWindowKey]int{}
	for key, sample := range samples {
		if sample.entry == nil {
			continue
		}
		providerUsed := sample.entry.Utilization
		if limit, ok := antigravityCalibrationLimitFromSample(providerUsed, sample.localCost); ok {
			antigravityQuotaCalibrationState.observe(account.ID, key, limit, now)
		}
		limit := 0.0
		if calibrated, ok := antigravityQuotaCalibrationState.limitFor(account.ID, key, now); ok {
			limit = calibrated
		}
		displayUsed[key] = antigravityWindowDisplayUsed(providerUsed, providerUsed >= 100, sample.localCost, limit)
	}
	if len(displayUsed) == 0 {
		return usage
	}

	out := *usage
	out.AntigravityQuota = make(map[string]*AntigravityModelQuota, len(usage.AntigravityQuota))
	for name, quota := range usage.AntigravityQuota {
		out.AntigravityQuota[name] = quota
		family := antigravityScopeFamily(name)
		if family == "" || quota == nil {
			continue
		}
		key := antigravityWindowKey5h(family)
		if isAntigravityWeeklyWindowKey(name) {
			key = antigravityWindowKey7d(family)
		}
		used, ok := displayUsed[key]
		if !ok || used <= quota.Utilization {
			continue
		}
		clamped := *quota
		clamped.Utilization = used
		out.AntigravityQuota[name] = &clamped
	}
	return &out
}
