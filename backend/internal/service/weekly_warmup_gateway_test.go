//go:build unit

package service

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	"github.com/stretchr/testify/require"
)

func TestAntigravityWeeklyWarmupOverridesThenRestoresSticky(t *testing.T) {
	now := time.Now().UTC()
	ctx := context.WithValue(context.Background(), ctxkey.ForcePlatform, PlatformAntigravity)

	newService := func(warmupPercent float64) (*GatewayService, *mockGatewayCacheForPlatform) {
		accounts := []Account{
			{
				ID: 75001, Platform: PlatformAntigravity, Type: AccountTypeOAuth,
				Status: StatusActive, Schedulable: true, Concurrency: 5,
				Extra: antigravityWeeklyUsageExtra(5, now.Add(7*24*time.Hour)),
			},
			{
				ID: 75002, Platform: PlatformAntigravity, Type: AccountTypeOAuth,
				Status: StatusActive, Schedulable: true, Concurrency: 5, Priority: 99,
				Extra: antigravityWeeklyUsageExtra(warmupPercent, now.Add(7*24*time.Hour)),
			},
		}
		repo := &mockAccountRepoForPlatform{accounts: accounts, accountsByID: make(map[int64]*Account)}
		for i := range repo.accounts {
			repo.accountsByID[repo.accounts[i].ID] = &repo.accounts[i]
		}
		cache := &mockGatewayCacheForPlatform{sessionBindings: map[string]int64{"weekly-session": 75001}}
		cfg := testConfig()
		cfg.Gateway.Scheduling.LoadBatchEnabled = true
		return &GatewayService{
			accountRepo:        repo,
			cache:              cache,
			cfg:                cfg,
			concurrencyService: NewConcurrencyService(&mockConcurrencyCache{}),
		}, cache
	}

	t.Run("zero percent overrides sticky", func(t *testing.T) {
		svc, cache := newService(0)
		selection, err := svc.SelectAccountWithLoadAwareness(ctx, nil, "weekly-session", "", nil, "", 0)
		require.NoError(t, err)
		require.Equal(t, int64(75002), selection.Account.ID)
		require.Equal(t, int64(75002), cache.sessionBindings["weekly-session"])
		selection.ReleaseFunc()
	})

	t.Run("one percent restores sticky", func(t *testing.T) {
		svc, _ := newService(1)
		selection, err := svc.SelectAccountWithLoadAwareness(ctx, nil, "weekly-session", "", nil, "", 0)
		require.NoError(t, err)
		require.Equal(t, int64(75001), selection.Account.ID)
		selection.ReleaseFunc()
	})
}
