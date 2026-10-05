package service

import (
	"context"
	"math"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/usagestats"
)

// Antigravity 额度窗口的展示口径（2026-10-05 实测后重写）：
//
//   - 百分比只展示上游自己的读数（不再用本地成本反推窗口上限：实测同一窗口内
//     上游读数与本网关成本并不成比例，任何反推出的百分比都是臆造的）。
//   - 上游读数按四舍五入展示，非零用量至少显示 1%（原先 int() 截断会把 0.9% 显示为 0%）。
//   - remainingFraction=1 的空窗口，其 resetTime 是“探测时刻 + 窗口长度”的占位值，
//     每次探测都会漂移，展示时清空，不再显示假的倒计时。
//   - Claude(3p) 窗口在真实用量后仍报告空窗口（上游不计量），展示为“无数据”而非 0%。
//   - 429 “Individual quota reached” 写入的 model_rate_limits 仍把对应家族 5h 窗口
//     钉在 100% 直到恢复时间（applyAntigravityQuotaExhaustionOverlay）。
//   - 每个窗口附带本网关 usage_logs 的真实用量（请求数 / tokens / 成本），有真实
//     上游窗口时按窗口起点对齐，否则取滚动 5h / 7d。

const (
	antigravityQuotaWindow5h = 5 * time.Hour
	antigravityQuotaWindow7d = 7 * 24 * time.Hour

	antigravityLocalUsageGemini5h = "gemini_5h"
	antigravityLocalUsageGemini7d = "gemini_7d"
	antigravityLocalUsageClaude5h = "claude_5h"
	antigravityLocalUsageClaude7d = "claude_7d"
)

// antigravityLocalUsageWindow 描述一个展示窗口与其对应的上游 bucket。
type antigravityLocalUsageWindow struct {
	key     string
	family  string // gemini | 3p
	window  time.Duration
	buckets []string
}

var antigravityLocalUsageWindows = []antigravityLocalUsageWindow{
	{key: antigravityLocalUsageGemini5h, family: "gemini", window: antigravityQuotaWindow5h, buckets: []string{"gemini-5h", "gemini:5h"}},
	{key: antigravityLocalUsageGemini7d, family: "gemini", window: antigravityQuotaWindow7d, buckets: []string{"gemini-weekly", "gemini:weekly"}},
	{key: antigravityLocalUsageClaude5h, family: "3p", window: antigravityQuotaWindow5h, buckets: []string{"3p-5h", "claude:5h"}},
	{key: antigravityLocalUsageClaude7d, family: "3p", window: antigravityQuotaWindow7d, buckets: []string{"3p-weekly", "claude:weekly"}},
}

// isAntigravityGeminiModelName 判断 usage_logs 的请求模型是否属于 Gemini 家族。
func isAntigravityGeminiModelName(model string) bool {
	return strings.HasPrefix(strings.ToLower(strings.TrimSpace(model)), "gemini")
}

// antigravityFamilyWindowStats 汇总模型统计中属于指定家族（gemini / 3p）的用量。
func antigravityFamilyWindowStats(stats []usagestats.ModelStat, family string) *WindowStats {
	out := &WindowStats{}
	for _, stat := range stats {
		if isAntigravityGeminiModelName(stat.Model) != (family == "gemini") {
			continue
		}
		out.Requests += stat.Requests
		out.Tokens += stat.TotalTokens
		out.Cost += stat.AccountCost
		out.StandardCost += stat.Cost
		out.UserCost += stat.ActualCost
	}
	return out
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

// antigravityDisplayQuotaEntry 生成单个配额条目的展示副本。
func antigravityDisplayQuotaEntry(name string, quota *AntigravityModelQuota) *AntigravityModelQuota {
	out := *quota
	if out.UsedPercent > 0 {
		rounded := math.Round(out.UsedPercent)
		if rounded < 1 {
			rounded = 1
		}
		if rounded > 100 {
			rounded = 100
		}
		if int(rounded) > out.Utilization {
			out.Utilization = int(rounded)
		}
	}
	if out.Empty {
		out.ResetTime = ""
		if antigravityScopeFamily(name) == "3p" {
			out.Unmetered = true
		}
	}
	return &out
}

// antigravityComposeDisplayOverlay 返回 usage 的浅拷贝，配额条目替换为展示副本；
// 不修改入参（调用方可能持有缓存对象）。
func antigravityComposeDisplayOverlay(usage *UsageInfo) *UsageInfo {
	if usage == nil {
		return nil
	}
	out := *usage
	if len(usage.AntigravityQuota) == 0 {
		return &out
	}
	out.AntigravityQuota = make(map[string]*AntigravityModelQuota, len(usage.AntigravityQuota))
	for name, quota := range usage.AntigravityQuota {
		if quota == nil {
			out.AntigravityQuota[name] = nil
			continue
		}
		out.AntigravityQuota[name] = antigravityDisplayQuotaEntry(name, quota)
	}
	return &out
}

// antigravityLocalWindowStart 计算本地用量统计的起点：上游窗口有真实重置时间时
// 对齐到 “重置时间 - 窗口长度”，否则为滚动窗口 now - 窗口长度。
func antigravityLocalWindowStart(quota map[string]*AntigravityModelQuota, w antigravityLocalUsageWindow, now time.Time) time.Time {
	start := now.Add(-w.window)
	for _, name := range w.buckets {
		entry := quota[name]
		if entry == nil {
			continue
		}
		if entry.Empty || entry.ResetTime == "" {
			return start
		}
		resetAt, err := time.Parse(time.RFC3339, entry.ResetTime)
		if err != nil || !resetAt.After(now) {
			return start
		}
		if aligned := resetAt.Add(-w.window); aligned.After(start) && !aligned.After(now) {
			return aligned
		}
		return start
	}
	return start
}

// antigravityLocalUsage 查询本网关 usage_logs 在各展示窗口内的真实用量。
func (s *AccountUsageService) antigravityLocalUsage(ctx context.Context, accountID int64, quota map[string]*AntigravityModelQuota, now time.Time) map[string]*WindowStats {
	if s == nil || s.usageLogRepo == nil || accountID <= 0 {
		return nil
	}
	statsByStart := make(map[time.Time][]usagestats.ModelStat, 3)
	out := make(map[string]*WindowStats, len(antigravityLocalUsageWindows))
	for _, w := range antigravityLocalUsageWindows {
		start := antigravityLocalWindowStart(quota, w, now)
		stats, ok := statsByStart[start]
		if !ok {
			var err error
			stats, err = s.usageLogRepo.GetModelStatsWithFilters(ctx, start, now, 0, 0, accountID, 0, nil, nil, nil)
			if err != nil {
				continue
			}
			statsByStart[start] = stats
		}
		out[w.key] = antigravityFamilyWindowStats(stats, w.family)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// applyAntigravityQuotaDisplayOverlay 依次叠加配额耗尽钉底与展示口径修正，并附带
// 本网关真实用量。返回浅拷贝，不修改入参（调用方可能持有缓存对象）。
func (s *AccountUsageService) applyAntigravityQuotaDisplayOverlay(ctx context.Context, usage *UsageInfo, account *Account) *UsageInfo {
	if usage == nil || account == nil {
		return usage
	}
	out := antigravityComposeDisplayOverlay(applyAntigravityQuotaExhaustionOverlay(usage, account))
	if len(out.AntigravityQuota) > 0 {
		out.AntigravityLocalUsage = s.antigravityLocalUsage(ctx, account.ID, out.AntigravityQuota, time.Now())
	}
	return out
}

// antigravityMeteredQuota 过滤掉上游不计量的条目，供聚合摘要求均值。
func antigravityMeteredQuota(quota map[string]*AntigravityModelQuota) map[string]*AntigravityModelQuota {
	if len(quota) == 0 {
		return quota
	}
	out := make(map[string]*AntigravityModelQuota, len(quota))
	for name, entry := range quota {
		if entry == nil || entry.Unmetered {
			continue
		}
		out[name] = entry
	}
	return out
}

// addWindowStats 把 src 累加到 dst（dst 为 nil 时新建）。
func addWindowStats(dst, src *WindowStats) *WindowStats {
	if src == nil {
		return dst
	}
	if dst == nil {
		dst = &WindowStats{}
	}
	dst.Requests += src.Requests
	dst.Tokens += src.Tokens
	dst.Cost += src.Cost
	dst.StandardCost += src.StandardCost
	dst.UserCost += src.UserCost
	return dst
}
