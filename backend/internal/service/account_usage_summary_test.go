package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractAntigravityBucketUtilization(t *testing.T) {
	t.Run("Gemini 5h canonical and fallback", func(t *testing.T) {
		// Canonical takes precedence
		quotaCanonical := map[string]*AntigravityModelQuota{
			"gemini-5h":        {Utilization: 30, ResetTime: "2026-03-01T10:00:00Z"},
			"gemini-pro-agent": {Utilization: 80, ResetTime: "2026-03-01T12:00:00Z"},
		}
		util, ok := ExtractGemini5hUtilization(quotaCanonical)
		assert.True(t, ok)
		assert.Equal(t, 30.0, util)

		// Fallback uses maximum utilization across fallback models
		quotaFallback := map[string]*AntigravityModelQuota{
			"gemini-pro-agent":     {Utilization: 45},
			"gemini-3.1-pro-high":  {Utilization: 75},
			"gemini-3-flash-agent": {Utilization: 20},
		}
		util, ok = ExtractGemini5hUtilization(quotaFallback)
		assert.True(t, ok)
		assert.Equal(t, 75.0, util)

		// Missing returns false
		quotaNone := map[string]*AntigravityModelQuota{
			"unrelated-model": {Utilization: 50},
		}
		_, ok = ExtractGemini5hUtilization(quotaNone)
		assert.False(t, ok)
	})

	t.Run("Claude 5h canonical and fallback", func(t *testing.T) {
		quotaCanonical := map[string]*AntigravityModelQuota{
			"claude:5h":         {Utilization: 10},
			"claude-sonnet-4-5": {Utilization: 90},
		}
		util, ok := ExtractClaude5hUtilization(quotaCanonical)
		assert.True(t, ok)
		assert.Equal(t, 10.0, util)

		quotaFallback := map[string]*AntigravityModelQuota{
			"claude-fable-5":    {Utilization: 25},
			"claude-sonnet-4-5": {Utilization: 60},
			"claude-opus-4-6":   {Utilization: 40},
		}
		util, ok = ExtractClaude5hUtilization(quotaFallback)
		assert.True(t, ok)
		assert.Equal(t, 60.0, util)
	})

	t.Run("Gemini 7d and Claude 7d canonical", func(t *testing.T) {
		quota := map[string]*AntigravityModelQuota{
			"gemini-weekly": {Utilization: 15},
			"3p-weekly":     {Utilization: 85},
		}
		utilGem, okGem := ExtractGemini7dUtilization(quota)
		assert.True(t, okGem)
		assert.Equal(t, 15.0, utilGem)

		utilClaude, okClaude := ExtractClaude7dUtilization(quota)
		assert.True(t, okClaude)
		assert.Equal(t, 85.0, utilClaude)
	})
}

type stubAccountRepoForSummary struct {
	accounts []Account
}

func (r *stubAccountRepoForSummary) ClearAntigravityQuotaScopes(ctx context.Context, id int64) error {
	return nil
}
func (r *stubAccountRepoForSummary) ClearModelRateLimits(ctx context.Context, id int64) error {
	return nil
}
func (r *stubAccountRepoForSummary) SetRateLimited(ctx context.Context, id int64, resetAt time.Time) error {
	return nil
}
func (r *stubAccountRepoForSummary) SetModelRateLimit(ctx context.Context, id int64, scope string, resetAt time.Time, reason ...string) error {
	return nil
}
func (r *stubAccountRepoForSummary) SetOverloaded(ctx context.Context, id int64, until time.Time) error {
	return nil
}
func (r *stubAccountRepoForSummary) SetTempUnschedulable(ctx context.Context, id int64, until time.Time, reason string) error {
	return nil
}
func (r *stubAccountRepoForSummary) ClearTempUnschedulable(ctx context.Context, id int64) error {
	return nil
}
func (r *stubAccountRepoForSummary) ClearRateLimit(ctx context.Context, id int64) error {
	return nil
}
func (r *stubAccountRepoForSummary) UpdateSessionWindow(ctx context.Context, id int64, start, end *time.Time, status string) error {
	return nil
}
func (r *stubAccountRepoForSummary) UpdateSessionWindowEnd(ctx context.Context, id int64, end time.Time) error {
	return nil
}
func (r *stubAccountRepoForSummary) UpdateExtra(ctx context.Context, id int64, updates map[string]any) error {
	return nil
}
func (r *stubAccountRepoForSummary) IncrementQuotaUsed(ctx context.Context, id int64, amount float64) error {
	return nil
}
func (r *stubAccountRepoForSummary) ResetQuotaUsed(ctx context.Context, id int64) error {
	return nil
}
func (r *stubAccountRepoForSummary) RevertProxyFallback(ctx context.Context, accountID int64) error {
	return nil
}
func (r *stubAccountRepoForSummary) ListShadowsByParent(ctx context.Context, parentID int64) ([]*Account, error) {
	return nil, nil
}
func (r *stubAccountRepoForSummary) ListModelAvailabilityCandidates(ctx context.Context, groupID *int64, platforms []string, includeGrouped bool) ([]Account, error) {
	return nil, nil
}
func (r *stubAccountRepoForSummary) ListSchedulableByPlatforms(ctx context.Context, platforms []string) ([]Account, error) {
	return nil, nil
}
func (r *stubAccountRepoForSummary) ListSchedulableByGroupIDAndPlatforms(ctx context.Context, groupID int64, platforms []string) ([]Account, error) {
	return nil, nil
}
func (r *stubAccountRepoForSummary) ListSchedulableUngroupedByPlatform(ctx context.Context, platform string) ([]Account, error) {
	return nil, nil
}
func (r *stubAccountRepoForSummary) ListSchedulableUngroupedByPlatforms(ctx context.Context, platforms []string) ([]Account, error) {
	return nil, nil
}
func (r *stubAccountRepoForSummary) BulkUpdate(ctx context.Context, ids []int64, updates AccountBulkUpdate) (int64, error) {
	return 0, nil
}
func (r *stubAccountRepoForSummary) Create(ctx context.Context, account *Account) error {
	return nil
}
func (r *stubAccountRepoForSummary) GetByID(ctx context.Context, id int64) (*Account, error) {
	for i := range r.accounts {
		if r.accounts[i].ID == id {
			return &r.accounts[i], nil
		}
	}
	return nil, errors.New("account not found")
}
func (r *stubAccountRepoForSummary) GetByIDs(ctx context.Context, ids []int64) ([]*Account, error) {
	var out []*Account
	for i := range r.accounts {
		for _, id := range ids {
			if r.accounts[i].ID == id {
				out = append(out, &r.accounts[i])
			}
		}
	}
	return out, nil
}
func (r *stubAccountRepoForSummary) ExistsByID(ctx context.Context, id int64) (bool, error) {
	for i := range r.accounts {
		if r.accounts[i].ID == id {
			return true, nil
		}
	}
	return false, nil
}
func (r *stubAccountRepoForSummary) GetByCRSAccountID(ctx context.Context, crsAccountID string) (*Account, error) {
	return nil, nil
}
func (r *stubAccountRepoForSummary) FindByExtraField(ctx context.Context, key string, value any) ([]Account, error) {
	return nil, nil
}
func (r *stubAccountRepoForSummary) ListCRSAccountIDs(ctx context.Context) (map[string]int64, error) {
	return map[string]int64{}, nil
}
func (r *stubAccountRepoForSummary) Update(ctx context.Context, account *Account) error { return nil }
func (r *stubAccountRepoForSummary) Delete(ctx context.Context, id int64) error         { return nil }
func (r *stubAccountRepoForSummary) List(ctx context.Context, params pagination.PaginationParams) ([]Account, *pagination.PaginationResult, error) {
	return r.ListWithFilters(ctx, params, "", "", "", "", 0, "")
}
func (r *stubAccountRepoForSummary) ListWithFilters(ctx context.Context, params pagination.PaginationParams, platform, accountType, status, search string, groupID int64, privacyMode string) ([]Account, *pagination.PaginationResult, error) {
	var filtered []Account
	for _, acc := range r.accounts {
		if platform != "" && acc.Platform != platform {
			continue
		}
		if accountType != "" && acc.Type != accountType {
			continue
		}
		if status != "" && acc.Status != status {
			continue
		}
		filtered = append(filtered, acc)
	}

	total := int64(len(filtered))
	offset := params.Offset()
	limit := params.Limit()

	if offset >= len(filtered) {
		return []Account{}, &pagination.PaginationResult{Total: total, Page: params.Page, PageSize: params.PageSize}, nil
	}
	end := offset + limit
	if end > len(filtered) {
		end = len(filtered)
	}

	return filtered[offset:end], &pagination.PaginationResult{
		Total:    total,
		Page:     params.Page,
		PageSize: params.PageSize,
	}, nil
}
func (r *stubAccountRepoForSummary) ListAllWithFilters(ctx context.Context, platform, accountType, status, search string, groupID int64, privacyMode string) ([]Account, error) {
	accs, _, err := r.ListWithFilters(ctx, pagination.PaginationParams{Page: 1, PageSize: 10000}, platform, accountType, status, search, groupID, privacyMode)
	return accs, err
}
func (r *stubAccountRepoForSummary) ListByGroup(ctx context.Context, groupID int64) ([]Account, error) {
	return nil, nil
}
func (r *stubAccountRepoForSummary) ListActive(ctx context.Context) ([]Account, error) { return nil, nil }
func (r *stubAccountRepoForSummary) ListByPlatform(ctx context.Context, platform string) ([]Account, error) {
	return nil, nil
}
func (r *stubAccountRepoForSummary) UpdateLastUsed(ctx context.Context, id int64) error { return nil }
func (r *stubAccountRepoForSummary) BatchUpdateLastUsed(ctx context.Context, updates map[int64]time.Time) error {
	return nil
}
func (r *stubAccountRepoForSummary) SetError(ctx context.Context, id int64, errorMsg string) error {
	return nil
}
func (r *stubAccountRepoForSummary) ClearError(ctx context.Context, id int64) error { return nil }
func (r *stubAccountRepoForSummary) SetSchedulable(ctx context.Context, id int64, schedulable bool) error {
	return nil
}
func (r *stubAccountRepoForSummary) AutoPauseExpiredAccounts(ctx context.Context, now time.Time) (int64, error) {
	return 0, nil
}
func (r *stubAccountRepoForSummary) BindGroups(ctx context.Context, accountID int64, groupIDs []int64) error {
	return nil
}
func (r *stubAccountRepoForSummary) ListSchedulable(ctx context.Context) ([]Account, error) {
	return nil, nil
}
func (r *stubAccountRepoForSummary) ListSchedulableByGroupID(ctx context.Context, groupID int64) ([]Account, error) {
	return nil, nil
}
func (r *stubAccountRepoForSummary) ListSchedulableByPlatform(ctx context.Context, platform string) ([]Account, error) {
	return nil, nil
}
func (r *stubAccountRepoForSummary) ListSchedulableByGroupIDAndPlatform(ctx context.Context, groupID int64, platform string) ([]Account, error) {
	return nil, nil
}
func TestGetAntigravityUsageSummary(t *testing.T) {
	t.Run("Calculates arithmetic mean, handles missing windows and partial account failures", func(t *testing.T) {
		repo := &stubAccountRepoForSummary{
			accounts: []Account{
				// Eligible account 1
				{ID: 1, Platform: PlatformAntigravity, Type: AccountTypeOAuth, Status: StatusActive, Schedulable: true, Credentials: map[string]any{"access_token": "test"}},
				// Eligible account 2
				{ID: 2, Platform: PlatformAntigravity, Type: AccountTypeOAuth, Status: StatusActive, Schedulable: true, Credentials: map[string]any{"access_token": "test"}},
				// Eligible account 3 (failed query)
				{ID: 3, Platform: PlatformAntigravity, Type: AccountTypeOAuth, Status: StatusActive, Schedulable: true, Credentials: map[string]any{"access_token": "test"}},
				// Eligible account 4 (partial windows)
				{ID: 4, Platform: PlatformAntigravity, Type: AccountTypeOAuth, Status: StatusActive, Schedulable: true, Credentials: map[string]any{"access_token": "test"}},
				// Excluded: unschedulable
				{ID: 5, Platform: PlatformAntigravity, Type: AccountTypeOAuth, Status: StatusActive, Schedulable: false, Credentials: map[string]any{"access_token": "test"}},
				// Excluded: non-OAuth
				{ID: 6, Platform: PlatformAntigravity, Type: AccountTypeAPIKey, Status: StatusActive, Schedulable: true, Credentials: map[string]any{"access_token": "test"}},
				// Excluded: non-Antigravity
				{ID: 7, Platform: PlatformOpenAI, Type: AccountTypeOAuth, Status: StatusActive, Schedulable: true, Credentials: map[string]any{"access_token": "test"}},
				// Excluded: disabled
				{ID: 8, Platform: PlatformAntigravity, Type: AccountTypeOAuth, Status: StatusDisabled, Schedulable: true, Credentials: map[string]any{"access_token": "test"}},
			},
		}

		cache := NewUsageCache()
		// Populate antigravity cache for accounts 1, 2, 4
		cache.antigravityCache.Store(int64(1), &antigravityUsageCache{
			usageInfo: &UsageInfo{
				AntigravityQuota: map[string]*AntigravityModelQuota{
					"gemini-5h":     {Utilization: 20},
					"gemini-weekly": {Utilization: 10},
					"3p-5h":         {Utilization: 50},
					"3p-weekly":     {Utilization: 30},
				},
			},
			timestamp: time.Now(),
		})
		cache.antigravityCache.Store(int64(2), &antigravityUsageCache{
			usageInfo: &UsageInfo{
				AntigravityQuota: map[string]*AntigravityModelQuota{
					"gemini-5h":     {Utilization: 40},
					"gemini-weekly": {Utilization: 20},
					"3p-5h":         {Utilization: 70},
					"3p-weekly":     {Utilization: 40},
				},
			},
			timestamp: time.Now(),
		})
		// Account 3 represents a degraded/failed usage fetch
		cache.antigravityCache.Store(int64(3), &antigravityUsageCache{
			usageInfo: &UsageInfo{
				Error: "upstream timeout",
			},
			timestamp: time.Now(),
		})
		cache.antigravityCache.Store(int64(4), &antigravityUsageCache{
			usageInfo: &UsageInfo{
				AntigravityQuota: map[string]*AntigravityModelQuota{
					"gemini-5h":     {Utilization: 60},
					"gemini-weekly": {Utilization: 30},
					// 3p-5h and 3p-weekly missing for account 4
				},
			},
			timestamp: time.Now(),
		})

		service := NewAccountUsageService(
			repo,
			nil,
			nil,
			nil,
			NewAntigravityQuotaFetcher(nil),
			nil,
			nil,
			nil,
			cache,
			nil,
			nil,
		)

		summary, err := service.GetAntigravityUsageSummary(context.Background(), true)
		require.NoError(t, err)
		require.NotNil(t, summary)

		// 4 total eligible accounts (1, 2, 3, 4)
		assert.Equal(t, 4, summary.EligibleAccounts)
		// Account 3 failed
		assert.Equal(t, 1, summary.FailedAccounts)

		// Gemini 5h: (20 + 40 + 60) / 3 = 40.0, sample_count = 3
		require.NotNil(t, summary.Gemini5h)
		require.NotNil(t, summary.Gemini5h.Utilization)
		assert.Equal(t, 40.0, *summary.Gemini5h.Utilization)
		assert.Equal(t, 3, summary.Gemini5h.SampleCount)

		// Gemini 7d: (10 + 20 + 30) / 3 = 20.0, sample_count = 3
		require.NotNil(t, summary.Gemini7d)
		require.NotNil(t, summary.Gemini7d.Utilization)
		assert.Equal(t, 20.0, *summary.Gemini7d.Utilization)
		assert.Equal(t, 3, summary.Gemini7d.SampleCount)

		// Claude 5h: (50 + 70) / 2 = 60.0, sample_count = 2
		require.NotNil(t, summary.Claude5h)
		require.NotNil(t, summary.Claude5h.Utilization)
		assert.Equal(t, 60.0, *summary.Claude5h.Utilization)
		assert.Equal(t, 2, summary.Claude5h.SampleCount)

		// Claude 7d: (30 + 40) / 2 = 35.0, sample_count = 2
		require.NotNil(t, summary.Claude7d)
		require.NotNil(t, summary.Claude7d.Utilization)
		assert.Equal(t, 35.0, *summary.Claude7d.Utilization)
		assert.Equal(t, 2, summary.Claude7d.SampleCount)
	})

	t.Run("All windows missing returns null utilization and sample count 0", func(t *testing.T) {
		repo := &stubAccountRepoForSummary{
			accounts: []Account{
				{ID: 10, Platform: PlatformAntigravity, Type: AccountTypeOAuth, Status: StatusActive, Schedulable: true, Credentials: map[string]any{"access_token": "test"}},
			},
		}

		cache := NewUsageCache()
		cache.antigravityCache.Store(int64(10), &antigravityUsageCache{
			usageInfo: &UsageInfo{
				AntigravityQuota: map[string]*AntigravityModelQuota{},
			},
			timestamp: time.Now(),
		})

		service := NewAccountUsageService(
			repo,
			nil,
			nil,
			nil,
			NewAntigravityQuotaFetcher(nil),
			nil,
			nil,
			nil,
			cache,
			nil,
			nil,
		)

		summary, err := service.GetAntigravityUsageSummary(context.Background(), true)
		require.NoError(t, err)

		assert.Equal(t, 1, summary.EligibleAccounts)
		assert.Equal(t, 0, summary.FailedAccounts)

		assert.Nil(t, summary.Gemini5h.Utilization)
		assert.Equal(t, 0, summary.Gemini5h.SampleCount)

		assert.Nil(t, summary.Gemini7d.Utilization)
		assert.Equal(t, 0, summary.Gemini7d.SampleCount)

		assert.Nil(t, summary.Claude5h.Utilization)
		assert.Equal(t, 0, summary.Claude5h.SampleCount)

		assert.Nil(t, summary.Claude7d.Utilization)
		assert.Equal(t, 0, summary.Claude7d.SampleCount)
	})

	t.Run("Pages through accounts across multiple pages", func(t *testing.T) {
		var accounts []Account
		cache := NewUsageCache()
		for i := int64(1); i <= 55; i++ {
			accounts = append(accounts, Account{
				ID:          i,
				Platform:    PlatformAntigravity,
				Type:        AccountTypeOAuth,
				Status:      StatusActive,
				Schedulable: true,
				Credentials: map[string]any{"access_token": "test"},
			})
			cache.antigravityCache.Store(i, &antigravityUsageCache{
				usageInfo: &UsageInfo{
					AntigravityQuota: map[string]*AntigravityModelQuota{
						"gemini-5h": {Utilization: int(i)},
					},
				},
				timestamp: time.Now(),
			})
		}

		repo := &stubAccountRepoForSummary{accounts: accounts}
		service := NewAccountUsageService(
			repo,
			nil,
			nil,
			nil,
			NewAntigravityQuotaFetcher(nil),
			nil,
			nil,
			nil,
			cache,
			nil,
			nil,
		)

		summary, err := service.GetAntigravityUsageSummary(context.Background(), true)
		require.NoError(t, err)
		assert.Equal(t, 55, summary.EligibleAccounts)
		assert.Equal(t, 55, summary.Gemini5h.SampleCount)
		// Sum of 1..55 is 1540. Mean is 1540/55 = 28.0
		require.NotNil(t, summary.Gemini5h.Utilization)
		assert.Equal(t, 28.0, *summary.Gemini5h.Utilization)
	})
}
