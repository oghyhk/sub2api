package service

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

func openAIWeeklyUsageExtra(percent float64, resetAt time.Time) map[string]any {
	return map[string]any{
		"codex_7d_used_percent":  percent,
		"codex_7d_reset_at":      resetAt.Format(time.RFC3339),
		"codex_usage_updated_at": time.Now().UTC().Format(time.RFC3339),
	}
}

func antigravityWeeklyUsageExtra(percent float64, resetAt time.Time) map[string]any {
	return map[string]any{
		antigravityGeminiWeeklyUsedPercentKey: percent,
		antigravityGeminiWeeklyResetAtKey:     resetAt.Format(time.RFC3339),
		antigravityWeeklyUsageUpdatedAtKey:    time.Now().UTC().Format(time.RFC3339),
	}
}

func TestWeeklyWarmupUtilizationUsesKnownDashboardWindows(t *testing.T) {
	now := time.Now().UTC()

	tests := []struct {
		name    string
		account *Account
		model   string
		warmup  bool
		known   bool
	}{
		{
			name:    "missing OpenAI snapshot is unknown",
			account: &Account{Platform: PlatformOpenAI},
		},
		{
			name:    "OpenAI zero percent warms",
			account: &Account{Platform: PlatformOpenAI, Extra: openAIWeeklyUsageExtra(0, now.Add(7*24*time.Hour))},
			warmup:  true,
			known:   true,
		},
		{
			name:    "OpenAI usage below one percent remains warm-up",
			account: &Account{Platform: PlatformOpenAI, Extra: openAIWeeklyUsageExtra(0.01, now.Add(7*24*time.Hour))},
			warmup:  true,
			known:   true,
		},
		{
			name:    "OpenAI usage at one percent is not warm-up",
			account: &Account{Platform: PlatformOpenAI, Extra: openAIWeeklyUsageExtra(1, now.Add(7*24*time.Hour))},
			known:   true,
		},
		{
			name:    "expired OpenAI window returns to warm-up",
			account: &Account{Platform: PlatformOpenAI, Extra: openAIWeeklyUsageExtra(80, now.Add(-time.Minute))},
			warmup:  true,
			known:   true,
		},
		{
			name:    "Antigravity Gemini zero percent warms",
			account: &Account{Platform: PlatformAntigravity, Extra: antigravityWeeklyUsageExtra(0, now.Add(7*24*time.Hour))},
			model:   "gemini-3.1-pro",
			warmup:  true,
			known:   true,
		},
		{
			name:    "Antigravity Gemini usage below one percent remains warm-up",
			account: &Account{Platform: PlatformAntigravity, Extra: antigravityWeeklyUsageExtra(0.01, now.Add(7*24*time.Hour))},
			model:   "gemini-3.1-pro",
			warmup:  true,
			known:   true,
		},
		{
			name:    "Antigravity Gemini usage at one percent is not warm-up",
			account: &Account{Platform: PlatformAntigravity, Extra: antigravityWeeklyUsageExtra(1, now.Add(7*24*time.Hour))},
			model:   "gemini-3.1-pro",
			known:   true,
		},
		{
			name: "Antigravity Claude uses third-party weekly bucket",
			account: &Account{Platform: PlatformAntigravity, Extra: map[string]any{
				antigravityThirdPartyWeeklyUsedKey:    0,
				antigravityThirdPartyWeeklyResetAtKey: now.Add(7 * 24 * time.Hour).Format(time.RFC3339),
			}},
			model:  "claude-sonnet-4-6",
			warmup: true,
			known:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, known := weeklyWarmupUtilization(tt.account, tt.model, now)
			require.Equal(t, tt.known, known)
			require.Equal(t, tt.warmup, isWeeklyWarmupAccount(tt.account, tt.model, now))
		})
	}
}

func TestAntigravityWeeklyUsageExtraUpdatesPersistsDashboardBuckets(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	info := &UsageInfo{
		UpdatedAt: &now,
		AntigravityQuota: map[string]*AntigravityModelQuota{
			"gemini-weekly": {Utilization: 0, ResetTime: "2026-08-01T00:00:00Z"},
			"3p-weekly":     {Utilization: 42, ResetTime: "2026-08-02T00:00:00Z"},
		},
	}

	updates := antigravityWeeklyUsageExtraUpdates(info, now.Add(time.Hour))
	require.Equal(t, 0, updates[antigravityGeminiWeeklyUsedPercentKey])
	require.Equal(t, 42, updates[antigravityThirdPartyWeeklyUsedKey])
	require.Equal(t, "2026-08-01T00:00:00Z", updates[antigravityGeminiWeeklyResetAtKey])
	require.Equal(t, "2026-08-02T00:00:00Z", updates[antigravityThirdPartyWeeklyResetAtKey])
	require.Equal(t, now.Format(time.RFC3339), updates[antigravityWeeklyUsageUpdatedAtKey])
}

func TestOpenAIAdvancedSchedulerWeeklyWarmupOverridesSticky(t *testing.T) {
	resetOpenAIAdvancedSchedulerSettingCacheForTest()
	now := time.Now().UTC()
	groupID := int64(7401)
	sticky := Account{
		ID: 74011, Platform: PlatformOpenAI, Type: AccountTypeOAuth,
		Status: StatusActive, Schedulable: true, Concurrency: 1, GroupIDs: []int64{groupID},
		Extra: openAIWeeklyUsageExtra(5, now.Add(7*24*time.Hour)),
	}
	warmup := Account{
		ID: 74012, Platform: PlatformOpenAI, Type: AccountTypeOAuth,
		Status: StatusActive, Schedulable: true, Concurrency: 1, Priority: 99, GroupIDs: []int64{groupID},
		Extra: openAIWeeklyUsageExtra(0, now.Add(7*24*time.Hour)),
	}
	cache := &schedulerTestGatewayCache{sessionBindings: map[string]int64{"openai:warm-session": sticky.ID}}
	svc := &OpenAIGatewayService{
		accountRepo:        schedulerTestOpenAIAccountRepo{accounts: []Account{sticky, warmup}},
		cache:              cache,
		cfg:                &config.Config{},
		rateLimitService:   newOpenAIAdvancedSchedulerRateLimitService("true"),
		concurrencyService: NewConcurrencyService(schedulerTestConcurrencyCache{}),
	}

	selection, decision, err := svc.SelectAccountWithScheduler(
		context.Background(), &groupID, "", "warm-session", "gpt-5.6-sol", nil,
		OpenAIUpstreamTransportAny, false,
	)
	require.NoError(t, err)
	require.Equal(t, warmup.ID, selection.Account.ID)
	require.Equal(t, openAIAccountScheduleLayerWeeklyWarmup, decision.Layer)
	selection.ReleaseFunc()
}

func TestOpenAILegacySchedulerWeeklyWarmupOverridesSticky(t *testing.T) {
	resetOpenAIAdvancedSchedulerSettingCacheForTest()
	now := time.Now().UTC()
	groupID := int64(7403)
	sticky := Account{
		ID: 74031, Platform: PlatformOpenAI, Type: AccountTypeOAuth,
		Status: StatusActive, Schedulable: true, Concurrency: 1, GroupIDs: []int64{groupID},
		Extra: openAIWeeklyUsageExtra(5, now.Add(7*24*time.Hour)),
	}
	warmup := Account{
		ID: 74032, Platform: PlatformOpenAI, Type: AccountTypeOAuth,
		Status: StatusActive, Schedulable: true, Concurrency: 1, Priority: 99, GroupIDs: []int64{groupID},
		Extra: openAIWeeklyUsageExtra(0, now.Add(7*24*time.Hour)),
	}
	cache := &schedulerTestGatewayCache{sessionBindings: map[string]int64{"openai:legacy-warm-session": sticky.ID}}
	cfg := &config.Config{}
	cfg.Gateway.Scheduling.LoadBatchEnabled = true
	svc := &OpenAIGatewayService{
		accountRepo:        schedulerTestOpenAIAccountRepo{accounts: []Account{sticky, warmup}},
		cache:              cache,
		cfg:                cfg,
		rateLimitService:   newOpenAIAdvancedSchedulerRateLimitService("false"),
		concurrencyService: NewConcurrencyService(schedulerTestConcurrencyCache{}),
	}

	selection, _, err := svc.SelectAccountWithScheduler(
		context.Background(), &groupID, "", "legacy-warm-session", "gpt-5.6-sol", nil,
		OpenAIUpstreamTransportAny, false,
	)
	require.NoError(t, err)
	require.Equal(t, warmup.ID, selection.Account.ID)
	selection.ReleaseFunc()
}

func TestOpenAIAdvancedSchedulerReturnsToStickyAfterOnePercentUsage(t *testing.T) {
	resetOpenAIAdvancedSchedulerSettingCacheForTest()
	now := time.Now().UTC()
	groupID := int64(7404)
	sticky := Account{
		ID: 74041, Platform: PlatformOpenAI, Type: AccountTypeOAuth,
		Status: StatusActive, Schedulable: true, Concurrency: 1, Priority: 99, GroupIDs: []int64{groupID},
		Extra: openAIWeeklyUsageExtra(5, now.Add(7*24*time.Hour)),
	}
	other := Account{
		ID: 74042, Platform: PlatformOpenAI, Type: AccountTypeOAuth,
		Status: StatusActive, Schedulable: true, Concurrency: 1, GroupIDs: []int64{groupID},
		Extra: openAIWeeklyUsageExtra(1, now.Add(7*24*time.Hour)),
	}
	cache := &schedulerTestGatewayCache{sessionBindings: map[string]int64{"openai:one-percent-sticky-session": sticky.ID}}
	svc := &OpenAIGatewayService{
		accountRepo:        schedulerTestOpenAIAccountRepo{accounts: []Account{sticky, other}},
		cache:              cache,
		cfg:                &config.Config{},
		rateLimitService:   newOpenAIAdvancedSchedulerRateLimitService("true"),
		concurrencyService: NewConcurrencyService(schedulerTestConcurrencyCache{}),
	}

	selection, decision, err := svc.SelectAccountWithScheduler(
		context.Background(), &groupID, "", "one-percent-sticky-session", "gpt-5.6-sol", nil,
		OpenAIUpstreamTransportAny, false,
	)
	require.NoError(t, err)
	require.Equal(t, sticky.ID, selection.Account.ID)
	require.Equal(t, openAIAccountScheduleLayerSessionSticky, decision.Layer)
	selection.ReleaseFunc()
}

func TestOpenAILegacySchedulerReturnsToStickyAfterWarmup(t *testing.T) {
	resetOpenAIAdvancedSchedulerSettingCacheForTest()
	now := time.Now().UTC()
	groupID := int64(7402)
	sticky := Account{
		ID: 74021, Platform: PlatformOpenAI, Type: AccountTypeOAuth,
		Status: StatusActive, Schedulable: true, Concurrency: 1, Priority: 99, GroupIDs: []int64{groupID},
		Extra: openAIWeeklyUsageExtra(5, now.Add(7*24*time.Hour)),
	}
	other := Account{
		ID: 74022, Platform: PlatformOpenAI, Type: AccountTypeOAuth,
		Status: StatusActive, Schedulable: true, Concurrency: 1, GroupIDs: []int64{groupID},
		Extra: openAIWeeklyUsageExtra(1, now.Add(7*24*time.Hour)),
	}
	cache := &schedulerTestGatewayCache{sessionBindings: map[string]int64{"openai:sticky-session": sticky.ID}}
	cfg := &config.Config{}
	cfg.Gateway.Scheduling.LoadBatchEnabled = true
	svc := &OpenAIGatewayService{
		accountRepo:        schedulerTestOpenAIAccountRepo{accounts: []Account{sticky, other}},
		cache:              cache,
		cfg:                cfg,
		rateLimitService:   newOpenAIAdvancedSchedulerRateLimitService("false"),
		concurrencyService: NewConcurrencyService(schedulerTestConcurrencyCache{}),
	}

	selection, _, err := svc.SelectAccountWithScheduler(
		context.Background(), &groupID, "", "sticky-session", "gpt-5.6-sol", nil,
		OpenAIUpstreamTransportAny, false,
	)
	require.NoError(t, err)
	require.Equal(t, sticky.ID, selection.Account.ID)
	selection.ReleaseFunc()
}
