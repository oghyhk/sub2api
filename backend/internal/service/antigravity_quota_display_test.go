package service

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/usagestats"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type antigravityDisplayUsageLogRepo struct {
	UsageLogRepository
	stats  []usagestats.ModelStat
	starts []time.Time
}

func (r *antigravityDisplayUsageLogRepo) GetModelStatsWithFilters(_ context.Context, startTime, _ time.Time, _, _, _, _ int64, _ *int16, _ *bool, _ *int8) ([]usagestats.ModelStat, error) {
	r.starts = append(r.starts, startTime)
	return r.stats, nil
}

func TestNewAntigravityModelQuota(t *testing.T) {
	q := newAntigravityModelQuota(0.9952381, "2026-10-05T07:35:39Z")
	assert.Equal(t, 0, q.Utilization, "scheduler-facing utilization keeps truncation")
	assert.InDelta(t, 0.47619, q.UsedPercent, 1e-4)
	assert.False(t, q.Empty)

	empty := newAntigravityModelQuota(1, "2026-10-12T03:26:25Z")
	assert.True(t, empty.Empty, "remainingFraction=1 is an empty window with a placeholder reset")
	assert.Equal(t, 0.0, empty.UsedPercent)

	clamped := newAntigravityModelQuota(-0.2, "")
	assert.Equal(t, 100, clamped.Utilization)
}

func TestAntigravityComposeDisplayOverlay(t *testing.T) {
	usage := &UsageInfo{AntigravityQuota: map[string]*AntigravityModelQuota{
		// 实测 2026-10-05：#5 gemini-5h remaining=0.5575459
		"gemini-5h":     newAntigravityModelQuota(0.5575459, "2026-10-05T04:45:07Z"),
		"gemini-weekly": newAntigravityModelQuota(0.996, "2026-10-11T05:11:10Z"),
		// Claude 窗口在真实用量后仍是 remaining=1 + 占位重置时间
		"3p-5h":                    newAntigravityModelQuota(1, "2026-10-05T08:26:35Z"),
		"3p-weekly":                newAntigravityModelQuota(1, "2026-10-12T03:26:35Z"),
		"claude-opus-4-6-thinking": newAntigravityModelQuota(1, "2026-10-05T08:26:36Z"),
		"gemini-3.5-flash-lite":    newAntigravityModelQuota(1, "2026-10-05T08:26:36Z"),
	}}

	out := antigravityComposeDisplayOverlay(usage)
	require.NotNil(t, out)

	assert.Equal(t, 44, out.AntigravityQuota["gemini-5h"].Utilization, "44.25% rounds to 44")
	assert.Equal(t, "2026-10-05T04:45:07Z", out.AntigravityQuota["gemini-5h"].ResetTime, "real reset kept")
	assert.Equal(t, 1, out.AntigravityQuota["gemini-weekly"].Utilization, "0.4% shows as 1%, never 0%")

	for _, name := range []string{"3p-5h", "3p-weekly", "claude-opus-4-6-thinking"} {
		entry := out.AntigravityQuota[name]
		assert.True(t, entry.Unmetered, "%s: empty Claude window is unmetered", name)
		assert.Empty(t, entry.ResetTime, "%s: placeholder countdown removed", name)
	}
	lite := out.AntigravityQuota["gemini-3.5-flash-lite"]
	assert.False(t, lite.Unmetered, "an empty Gemini window is a real 0%")
	assert.Equal(t, 0, lite.Utilization)
	assert.Empty(t, lite.ResetTime)

	// 缓存对象不被修改
	assert.Equal(t, 0, usage.AntigravityQuota["gemini-weekly"].Utilization)
	assert.Equal(t, "2026-10-12T03:26:35Z", usage.AntigravityQuota["3p-weekly"].ResetTime)
	assert.False(t, usage.AntigravityQuota["3p-weekly"].Unmetered)
}

func TestAntigravityDisplayOverlayKeepsExhaustionStamp(t *testing.T) {
	reset := time.Now().Add(2 * time.Hour).UTC().Truncate(time.Second)
	account := &Account{ID: 9, Platform: PlatformAntigravity, Extra: map[string]any{
		"model_rate_limits": map[string]any{
			"claude-opus-4-6-thinking": map[string]any{
				"rate_limited_at":     time.Now().Add(-time.Minute).UTC().Format(time.RFC3339),
				"rate_limit_reset_at": reset.Format(time.RFC3339),
			},
		},
	}}
	usage := &UsageInfo{AntigravityQuota: map[string]*AntigravityModelQuota{
		"3p-5h":     newAntigravityModelQuota(1, "2026-10-05T08:26:35Z"),
		"3p-weekly": newAntigravityModelQuota(1, "2026-10-12T03:26:35Z"),
	}}

	out := antigravityComposeDisplayOverlay(applyAntigravityQuotaExhaustionOverlay(usage, account))
	require.NotNil(t, out)
	five := out.AntigravityQuota["3p-5h"]
	assert.Equal(t, 100, five.Utilization, "upstream 429 pins the 5h window")
	assert.False(t, five.Unmetered)
	assert.Equal(t, reset.Format(time.RFC3339), five.ResetTime)
	assert.True(t, out.AntigravityQuota["3p-weekly"].Unmetered, "weekly window is not stamped")
}

func TestAntigravityLocalWindowStart(t *testing.T) {
	now := time.Date(2026, 10, 5, 3, 26, 0, 0, time.UTC)
	w5h := antigravityLocalUsageWindows[0]

	aligned := antigravityLocalWindowStart(map[string]*AntigravityModelQuota{
		"gemini-5h": {UsedPercent: 44, ResetTime: "2026-10-05T04:45:07Z"},
	}, w5h, now)
	assert.Equal(t, time.Date(2026, 10, 4, 23, 45, 7, 0, time.UTC), aligned, "aligned to reset - 5h")

	trailing := antigravityLocalWindowStart(map[string]*AntigravityModelQuota{
		"gemini-5h": {Empty: true},
	}, w5h, now)
	assert.Equal(t, now.Add(-5*time.Hour), trailing, "empty window falls back to trailing 5h")

	expired := antigravityLocalWindowStart(map[string]*AntigravityModelQuota{
		"gemini-5h": {UsedPercent: 10, ResetTime: "2026-10-05T03:00:00Z"},
	}, w5h, now)
	assert.Equal(t, now.Add(-5*time.Hour), expired)
}

func TestApplyAntigravityQuotaDisplayOverlayAttachesLocalUsage(t *testing.T) {
	repo := &antigravityDisplayUsageLogRepo{stats: []usagestats.ModelStat{
		{Model: "gemini-3.8-flash", Requests: 61, TotalTokens: 1_000_000, AccountCost: 1.5, Cost: 1.5, ActualCost: 1.5},
		{Model: "gemini-3.5-flash-lite", Requests: 10, TotalTokens: 10_000, AccountCost: 0.1, Cost: 0.1, ActualCost: 0.1},
		{Model: "claude-opus-4-6", Requests: 22, TotalTokens: 41_726, AccountCost: 0.42, Cost: 0.42, ActualCost: 0.42},
	}}
	svc := &AccountUsageService{usageLogRepo: repo}
	reset := time.Now().Add(80 * time.Minute).UTC().Truncate(time.Second)
	usage := &UsageInfo{AntigravityQuota: map[string]*AntigravityModelQuota{
		"gemini-5h":     newAntigravityModelQuota(0.55, reset.Format(time.RFC3339)),
		"gemini-weekly": newAntigravityModelQuota(0.9, "2026-10-11T05:11:10Z"),
		"3p-5h":         newAntigravityModelQuota(1, "2026-10-05T08:26:35Z"),
		"3p-weekly":     newAntigravityModelQuota(1, "2026-10-12T03:26:35Z"),
	}}

	out := svc.applyAntigravityQuotaDisplayOverlay(context.Background(), usage, &Account{ID: 5, Platform: PlatformAntigravity})
	require.NotNil(t, out)
	require.NotNil(t, out.AntigravityLocalUsage)

	gem := out.AntigravityLocalUsage[antigravityLocalUsageGemini5h]
	require.NotNil(t, gem)
	assert.Equal(t, int64(71), gem.Requests)
	assert.InDelta(t, 1.6, gem.Cost, 1e-9)

	claude7d := out.AntigravityLocalUsage[antigravityLocalUsageClaude7d]
	require.NotNil(t, claude7d)
	assert.Equal(t, int64(22), claude7d.Requests, "Claude 7d shows real gateway usage even though the meter is unmetered")
	assert.InDelta(t, 0.42, claude7d.Cost, 1e-9)

	assert.Nil(t, usage.AntigravityLocalUsage, "cached source untouched")
	assert.LessOrEqual(t, len(repo.starts), 4)
	assert.Contains(t, repo.starts, reset.Add(-5*time.Hour), "gemini 5h stats aligned to the provider window")
}

func TestAntigravityMeteredQuotaAndAddWindowStats(t *testing.T) {
	quota := map[string]*AntigravityModelQuota{
		"3p-5h":     {Unmetered: true},
		"gemini-5h": {Utilization: 10},
	}
	metered := antigravityMeteredQuota(quota)
	assert.Len(t, metered, 1)
	_, ok := metered["3p-5h"]
	assert.False(t, ok)
	_, _, ok = ExtractClaude5hUtilization(metered)
	assert.False(t, ok, "aggregate skips an unmetered Claude window instead of averaging it as 0%")

	total := addWindowStats(nil, &WindowStats{Requests: 2, Cost: 1})
	total = addWindowStats(total, &WindowStats{Requests: 3, Cost: 0.5})
	total = addWindowStats(total, nil)
	assert.Equal(t, int64(5), total.Requests)
	assert.InDelta(t, 1.5, total.Cost, 1e-9)
}
