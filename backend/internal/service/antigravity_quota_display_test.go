package service

import (
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/usagestats"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSplitAntigravityFamilyCosts(t *testing.T) {
	stats := []usagestats.ModelStat{
		{Model: "gemini-3.8-flash", AccountCost: 10},
		{Model: "gemini-3.1-flash-lite", AccountCost: 5},
		{Model: "claude-opus-5-5", AccountCost: 7},
		{Model: "gpt-oss-120b-medium", AccountCost: 3},
	}
	gemini, other := splitAntigravityFamilyCosts(stats)
	assert.Equal(t, 15.0, gemini)
	assert.Equal(t, 10.0, other)
}

func TestAntigravityWindowDisplayUsed(t *testing.T) {
	assert.Equal(t, 45, antigravityWindowDisplayUsed(45, false, 0, 0))
	assert.Equal(t, 100, antigravityWindowDisplayUsed(45, true, 0, 0))
	assert.Equal(t, 60, antigravityWindowDisplayUsed(45, false, 30, 50))
	assert.Equal(t, 45, antigravityWindowDisplayUsed(45, false, 10, 50))
	assert.Equal(t, 45, antigravityWindowDisplayUsed(45, false, 30, 0))
	assert.Equal(t, 100, antigravityWindowDisplayUsed(45, false, 90, 50))
	assert.Equal(t, 0, antigravityWindowDisplayUsed(-5, false, 0, 0))
}

func TestAntigravityCalibrationLimitFromSample(t *testing.T) {
	limit, ok := antigravityCalibrationLimitFromSample(50, 25)
	require.True(t, ok)
	assert.Equal(t, 50.0, limit)

	_, ok = antigravityCalibrationLimitFromSample(1, 25)
	assert.False(t, ok, "near-empty window cannot calibrate")
	_, ok = antigravityCalibrationLimitFromSample(100, 25)
	assert.False(t, ok, "exhausted window cannot calibrate")
	_, ok = antigravityCalibrationLimitFromSample(50, 0)
	assert.False(t, ok, "no local cost cannot calibrate")
}

func TestAntigravityQuotaCalibrationLimitFor(t *testing.T) {
	resetAntigravityQuotaCalibrationForTest()
	now := time.Now()
	key := antigravityWindowKey5h("gemini")

	_, ok := antigravityQuotaCalibrationState.limitFor(1, key, now)
	assert.False(t, ok)

	antigravityQuotaCalibrationState.observe(1, key, 100, now)
	limit, ok := antigravityQuotaCalibrationState.limitFor(1, key, now)
	require.True(t, ok)
	assert.Equal(t, 100.0, limit)

	antigravityQuotaCalibrationState.observe(2, key, 200, now)
	antigravityQuotaCalibrationState.observe(4, key, 300, now)
	limit, ok = antigravityQuotaCalibrationState.limitFor(3, key, now)
	require.True(t, ok)
	assert.Equal(t, 200.0, limit, "cold account borrows the median of other accounts")

	_, ok = antigravityQuotaCalibrationState.limitFor(3, key, now.Add(25*time.Hour))
	assert.False(t, ok, "borrowed samples expire after the TTL")

	antigravityQuotaCalibrationState.observe(1, key, 999, now.Add(-25*time.Hour))
	limit, ok = antigravityQuotaCalibrationState.limitFor(1, key, now)
	require.True(t, ok)
	assert.Equal(t, 250.0, limit, "stale own sample is skipped in favor of fresh fallback")
}

func TestAntigravityComposeDisplayOverlay(t *testing.T) {
	t.Run("rolling cost raises 5h and weekly independently after window roll", func(t *testing.T) {
		resetAntigravityQuotaCalibrationForTest()
		now := time.Now()

		// 校准样本：上游窗口仍有效（2%~98%）时反推上限。
		warm := &UsageInfo{AntigravityQuota: map[string]*AntigravityModelQuota{
			"gemini-5h":     {Utilization: 50},
			"gemini-weekly": {Utilization: 20},
			"3p-5h":         {Utilization: 25},
			"3p-weekly":     {Utilization: 10},
		}}
		out := antigravityComposeDisplayOverlay(warm, &Account{ID: 1}, now, 25, 5, 40, 18)
		require.NotNil(t, out)
		// 上限：gemini 5h=50, gemini 7d=200, 3p 5h=20, 3p 7d=180。

		// 窗口重建后的账号：上游读数为 0，但真实滚动成本仍在。
		usage := &UsageInfo{AntigravityQuota: map[string]*AntigravityModelQuota{
			"gemini-5h":     {Utilization: 0, ResetTime: "2026-10-04T13:18:32Z"},
			"gemini-weekly": {Utilization: 0, ResetTime: "2026-10-11T05:10:45Z"},
			"3p-5h":         {Utilization: 0, ResetTime: "2026-10-04T13:38:18Z"},
			"3p-weekly":     {Utilization: 0, ResetTime: "2026-10-11T08:38:18Z"},
		}}
		out = antigravityComposeDisplayOverlay(usage, &Account{ID: 2}, now, 30, 1, 150, 2)
		require.NotNil(t, out)
		assert.Equal(t, 60, out.AntigravityQuota["gemini-5h"].Utilization, "30/50 = 60% of calibrated 5h limit")
		assert.Equal(t, 75, out.AntigravityQuota["gemini-weekly"].Utilization, "150/200 = 75% of calibrated 7d limit")
		assert.Equal(t, 5, out.AntigravityQuota["3p-5h"].Utilization, "1/20 = 5% of calibrated 5h limit")
		assert.Equal(t, 1, out.AntigravityQuota["3p-weekly"].Utilization, "2/180 rounds to 1% of calibrated 7d limit")

		assert.Equal(t, 0, usage.AntigravityQuota["gemini-5h"].Utilization, "source must not be mutated")
	})

	t.Run("raise only and stamp floor preserved", func(t *testing.T) {
		resetAntigravityQuotaCalibrationForTest()
		now := time.Now()
		usage := &UsageInfo{AntigravityQuota: map[string]*AntigravityModelQuota{
			"gemini-5h":     {Utilization: 100, ResetTime: "2026-10-04T07:50:52Z"},
			"gemini-weekly": {Utilization: 80, ResetTime: "2026-10-08T13:05:29Z"},
		}}

		out := antigravityComposeDisplayOverlay(usage, &Account{ID: 3}, now, 5, 0, 10, 0)
		require.NotNil(t, out)
		assert.Equal(t, 100, out.AntigravityQuota["gemini-5h"].Utilization)
		assert.Equal(t, 80, out.AntigravityQuota["gemini-weekly"].Utilization, "lower local readings never reduce provider values")
	})

	t.Run("no calibration falls back to provider values", func(t *testing.T) {
		resetAntigravityQuotaCalibrationForTest()
		now := time.Now()
		usage := &UsageInfo{AntigravityQuota: map[string]*AntigravityModelQuota{
			"gemini-5h": {Utilization: 45, ResetTime: "2026-10-04T09:36:19Z"},
		}}
		out := antigravityComposeDisplayOverlay(usage, &Account{ID: 4}, now, 100, 0, 100, 0)
		require.NotNil(t, out)
		assert.Equal(t, 45, out.AntigravityQuota["gemini-5h"].Utilization)
	})
}
