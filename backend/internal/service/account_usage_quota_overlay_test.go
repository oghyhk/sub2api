package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func accountWithRateLimits(limits map[string]any) *Account {
	return &Account{Extra: map[string]any{modelRateLimitsKey: limits}}
}

func rateLimitEntry(resetAt time.Time) map[string]any {
	return map[string]any{
		"rate_limited_at":     time.Now().UTC().Format(time.RFC3339),
		"rate_limit_reset_at": resetAt.UTC().Format(time.RFC3339),
	}
}

func sampleQuotaMap() map[string]*AntigravityModelQuota {
	return map[string]*AntigravityModelQuota{
		"gemini-5h":                {Utilization: 45, ResetTime: "2026-10-04T09:36:19Z"},
		"gemini-weekly":            {Utilization: 9, ResetTime: "2026-10-08T13:05:29Z"},
		"gemini-3.8-flash-high":    {Utilization: 45, ResetTime: "2026-10-04T09:36:19Z"},
		"3p-5h":                    {Utilization: 0, ResetTime: "2026-10-04T11:30:09Z"},
		"3p-weekly":                {Utilization: 0, ResetTime: "2026-10-11T06:30:09Z"},
		"claude-opus-4-6-thinking": {Utilization: 0, ResetTime: "2026-10-04T11:30:08Z"},
		"gpt-oss-120b-medium":      {Utilization: 0, ResetTime: "2026-10-04T11:30:08Z"},
	}
}

func TestApplyAntigravityQuotaExhaustionOverlay(t *testing.T) {
	t.Run("gemini quota exhaustion clamps gemini 5h windows only", func(t *testing.T) {
		resetAt := time.Now().Add(2 * time.Hour).Truncate(time.Second)
		account := accountWithRateLimits(map[string]any{
			"antigravity:gemini":    rateLimitEntry(resetAt),
			"gemini-3.8-flash-high": rateLimitEntry(resetAt),
		})
		usage := &UsageInfo{AntigravityQuota: sampleQuotaMap()}

		out := applyAntigravityQuotaExhaustionOverlay(usage, account)
		require.NotNil(t, out)

		wantReset := resetAt.UTC().Format(time.RFC3339)
		assert.Equal(t, 100, out.AntigravityQuota["gemini-5h"].Utilization)
		assert.Equal(t, wantReset, out.AntigravityQuota["gemini-5h"].ResetTime)
		assert.Equal(t, 100, out.AntigravityQuota["gemini-3.8-flash-high"].Utilization)

		assert.Equal(t, 9, out.AntigravityQuota["gemini-weekly"].Utilization)
		assert.Equal(t, 0, out.AntigravityQuota["3p-5h"].Utilization)
		assert.Equal(t, 0, out.AntigravityQuota["claude-opus-4-6-thinking"].Utilization)
		assert.Equal(t, 0, out.AntigravityQuota["gpt-oss-120b-medium"].Utilization)

		assert.Equal(t, 45, usage.AntigravityQuota["gemini-5h"].Utilization,
			"cached source map must not be mutated")
		assert.NotSame(t, usage.AntigravityQuota["gemini-5h"], out.AntigravityQuota["gemini-5h"])
	})

	t.Run("claude quota exhaustion clamps 3p 5h windows only", func(t *testing.T) {
		resetAt := time.Now().Add(90 * time.Minute).Truncate(time.Second)
		account := accountWithRateLimits(map[string]any{
			"claude-opus-4-6-thinking": rateLimitEntry(resetAt),
		})
		usage := &UsageInfo{AntigravityQuota: sampleQuotaMap()}

		out := applyAntigravityQuotaExhaustionOverlay(usage, account)
		require.NotNil(t, out)

		assert.Equal(t, 100, out.AntigravityQuota["3p-5h"].Utilization)
		assert.Equal(t, resetAt.UTC().Format(time.RFC3339), out.AntigravityQuota["3p-5h"].ResetTime)
		assert.Equal(t, 100, out.AntigravityQuota["claude-opus-4-6-thinking"].Utilization)
		assert.Equal(t, 45, out.AntigravityQuota["gemini-5h"].Utilization)
		assert.Equal(t, 0, out.AntigravityQuota["3p-weekly"].Utilization)
	})

	t.Run("AICredits exhaustion clamps both families", func(t *testing.T) {
		resetAt := time.Now().Add(3 * time.Hour).Truncate(time.Second)
		account := accountWithRateLimits(map[string]any{
			creditsExhaustedKey: rateLimitEntry(resetAt),
		})
		usage := &UsageInfo{AntigravityQuota: sampleQuotaMap()}

		out := applyAntigravityQuotaExhaustionOverlay(usage, account)
		require.NotNil(t, out)

		assert.Equal(t, 100, out.AntigravityQuota["gemini-5h"].Utilization)
		assert.Equal(t, 100, out.AntigravityQuota["3p-5h"].Utilization)
		assert.Equal(t, 9, out.AntigravityQuota["gemini-weekly"].Utilization)
		assert.Equal(t, 0, out.AntigravityQuota["3p-weekly"].Utilization)
	})

	t.Run("expired stamps are ignored", func(t *testing.T) {
		account := accountWithRateLimits(map[string]any{
			"antigravity:gemini":    rateLimitEntry(time.Now().Add(-time.Minute)),
			"gemini-3.8-flash-high": rateLimitEntry(time.Now().Add(-time.Hour)),
		})
		usage := &UsageInfo{AntigravityQuota: sampleQuotaMap()}

		out := applyAntigravityQuotaExhaustionOverlay(usage, account)
		require.NotNil(t, out)
		assert.Equal(t, 45, out.AntigravityQuota["gemini-5h"].Utilization)
		assert.Equal(t, "2026-10-04T09:36:19Z", out.AntigravityQuota["gemini-5h"].ResetTime)
	})

	t.Run("no stamps returns usage unchanged", func(t *testing.T) {
		account := accountWithRateLimits(map[string]any{})
		usage := &UsageInfo{AntigravityQuota: sampleQuotaMap()}

		out := applyAntigravityQuotaExhaustionOverlay(usage, account)
		assert.Same(t, usage, out)
	})

	t.Run("nil usage or account is passthrough", func(t *testing.T) {
		account := accountWithRateLimits(map[string]any{})
		assert.Nil(t, applyAntigravityQuotaExhaustionOverlay(nil, account))
		usage := &UsageInfo{AntigravityQuota: sampleQuotaMap()}
		assert.Same(t, usage, applyAntigravityQuotaExhaustionOverlay(usage, nil))
	})
}

func TestActiveModelRateLimitResetAt(t *testing.T) {
	soon := time.Now().Add(30 * time.Minute).Truncate(time.Second)
	latest := time.Now().Add(3 * time.Hour).Truncate(time.Second)
	account := accountWithRateLimits(map[string]any{
		"antigravity:gemini":    rateLimitEntry(soon),
		"gemini-3.8-flash-high": rateLimitEntry(latest),
		"gemini-3.7-flash-high": rateLimitEntry(time.Now().Add(-time.Minute)),
		"claude-sonnet-4-6":     rateLimitEntry(latest),
	})

	gemReset := account.activeModelRateLimitResetAt(antigravityGeminiScopeKey)
	require.NotNil(t, gemReset)
	assert.True(t, latest.Equal(*gemReset), "want latest active gemini reset, got %v", gemReset)

	claudeReset := account.activeModelRateLimitResetAt(antigravityClaudeScopeKey)
	require.NotNil(t, claudeReset)
	assert.True(t, latest.Equal(*claudeReset))

	none := accountWithRateLimits(map[string]any{
		"gemini-3.8-flash-high": rateLimitEntry(time.Now().Add(-time.Minute)),
	})
	assert.Nil(t, none.activeModelRateLimitResetAt(antigravityGeminiScopeKey))
}
