package service

import (
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/usagestats"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSplitAnthropicClassCosts(t *testing.T) {
	stats := []usagestats.ModelStat{
		{Model: "claude-sonnet-4-6", AccountCost: 10},
		{Model: "claude-fable-5", AccountCost: 7},
		{Model: "claude-opus-4-6", AccountCost: 5},
		{Model: "claude-haiku-4-5", AccountCost: 2},
	}
	total, sonnet, fable := splitAnthropicClassCosts(stats)
	assert.Equal(t, 24.0, total)
	assert.Equal(t, 10.0, sonnet)
	assert.Equal(t, 7.0, fable)
}

func TestAnthropicScopeClass(t *testing.T) {
	assert.Equal(t, "claude:sonnet", anthropicScopeClass("claude-sonnet-4-6"))
	assert.Equal(t, "claude:fable", anthropicScopeClass("claude-fable-5"))
	assert.Equal(t, "claude", anthropicScopeClass("claude-opus-4-6-thinking"))
	assert.Equal(t, "", anthropicScopeClass("gemini-3.8-flash"))
	assert.Equal(t, "", anthropicScopeClass("antigravity:gemini"))
}

func TestAnthropicWindowStampActive(t *testing.T) {
	resetAt := time.Now().Add(time.Hour)
	account := accountWithRateLimits(map[string]any{
		"claude-sonnet-4-6": rateLimitEntry(resetAt),
	})
	assert.True(t, anthropicWindowStampActive(account, "claude"), "generic windows see any claude stamp")
	assert.True(t, anthropicWindowStampActive(account, "claude:sonnet"))
	assert.False(t, anthropicWindowStampActive(account, "claude:fable"))

	fable := accountWithRateLimits(map[string]any{
		"claude-fable-5": rateLimitEntry(resetAt),
	})
	assert.True(t, anthropicWindowStampActive(fable, "claude:fable"))
	assert.False(t, anthropicWindowStampActive(fable, "claude:sonnet"))

	generic := accountWithRateLimits(map[string]any{
		"claude-opus-4-6": rateLimitEntry(resetAt),
	})
	assert.True(t, anthropicWindowStampActive(generic, "claude"))
	assert.False(t, anthropicWindowStampActive(generic, "claude:sonnet"))

	expired := accountWithRateLimits(map[string]any{
		"claude-sonnet-4-6": rateLimitEntry(time.Now().Add(-time.Minute)),
	})
	assert.False(t, anthropicWindowStampActive(expired, "claude"))
}

func TestAnthropicComposeDisplayOverlay(t *testing.T) {
	t.Run("rolling cost raises 5h and class windows after window roll", func(t *testing.T) {
		resetAntigravityQuotaCalibrationForTest()
		now := time.Now()

		warm := &UsageInfo{
			FiveHour:       &UsageProgress{Utilization: 50},
			SevenDay:       &UsageProgress{Utilization: 20},
			SevenDaySonnet: &UsageProgress{Utilization: 40},
			SevenDayFable:  &UsageProgress{Utilization: 10},
		}
		out := anthropicComposeDisplayOverlay(warm, &Account{ID: 1}, now, 25, 40, 32, 9)
		require.NotNil(t, out)
		// 校准上限：5h=50、7d=200、sonnet 7d=80、fable 7d=90。

		rolled := &UsageInfo{
			FiveHour:       &UsageProgress{Utilization: 0},
			SevenDay:       &UsageProgress{Utilization: 0},
			SevenDaySonnet: &UsageProgress{Utilization: 0},
			SevenDayFable:  &UsageProgress{Utilization: 0},
		}
		out = anthropicComposeDisplayOverlay(rolled, &Account{ID: 2}, now, 30, 150, 40, 18)
		require.NotNil(t, out)
		assert.Equal(t, 60.0, out.FiveHour.Utilization, "30/50 = 60% of calibrated 5h limit")
		assert.Equal(t, 75.0, out.SevenDay.Utilization, "150/200 = 75% of calibrated 7d limit")
		assert.Equal(t, 50.0, out.SevenDaySonnet.Utilization, "40/80 = 50% of calibrated sonnet limit")
		assert.Equal(t, 20.0, out.SevenDayFable.Utilization, "18/90 = 20% of calibrated fable limit")

		assert.Equal(t, 0.0, rolled.FiveHour.Utilization, "source must not be mutated")
	})

	t.Run("stamp floors shared windows and its class window", func(t *testing.T) {
		resetAntigravityQuotaCalibrationForTest()
		now := time.Now()
		usage := &UsageInfo{
			FiveHour:       &UsageProgress{Utilization: 40},
			SevenDay:       &UsageProgress{Utilization: 80},
			SevenDaySonnet: &UsageProgress{Utilization: 90},
			SevenDayFable:  &UsageProgress{Utilization: 10},
		}
		account := accountWithRateLimits(map[string]any{
			"claude-fable-5": rateLimitEntry(time.Now().Add(time.Hour)),
		})
		out := anthropicComposeDisplayOverlay(usage, account, now, 5, 10, 8, 2)
		require.NotNil(t, out)
		assert.Equal(t, 100.0, out.FiveHour.Utilization, "any claude stamp floors the shared 5h window")
		assert.Equal(t, 100.0, out.SevenDay.Utilization, "any claude stamp floors the shared 7d window")
		assert.Equal(t, 90.0, out.SevenDaySonnet.Utilization, "fable stamp does not floor the sonnet window")
		assert.Equal(t, 100.0, out.SevenDayFable.Utilization, "fable stamp floors the fable window")
	})

	t.Run("lower local readings never reduce provider values", func(t *testing.T) {
		resetAntigravityQuotaCalibrationForTest()
		now := time.Now()
		usage := &UsageInfo{SevenDay: &UsageProgress{Utilization: 80}}
		out := anthropicComposeDisplayOverlay(usage, accountWithRateLimits(map[string]any{}), now, 1, 2, 1, 1)
		require.NotNil(t, out)
		assert.Equal(t, 80.0, out.SevenDay.Utilization)
	})

	t.Run("no calibration falls back to provider values", func(t *testing.T) {
		resetAntigravityQuotaCalibrationForTest()
		now := time.Now()
		usage := &UsageInfo{FiveHour: &UsageProgress{Utilization: 45}}
		out := anthropicComposeDisplayOverlay(usage, &Account{ID: 4}, now, 100, 100, 100, 100)
		require.NotNil(t, out)
		assert.Equal(t, 45.0, out.FiveHour.Utilization)
	})
}
