import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'

import type { DashboardStats } from '@/types'
import DashboardView from '../DashboardView.vue'

const { getSnapshotV2, getModelStats, getUserUsageTrend, getUserSpendingRanking, listUsage } = vi.hoisted(() => ({
  getSnapshotV2: vi.fn(),
  getModelStats: vi.fn(),
  getUserUsageTrend: vi.fn(),
  getUserSpendingRanking: vi.fn(),
  listUsage: vi.fn()
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    dashboard: {
      getSnapshotV2,
      getModelStats,
      getUserUsageTrend,
      getUserSpendingRanking
    },
    usage: {
      list: listUsage
    }
  }
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError: vi.fn()
  })
}))

vi.mock('vue-router', () => ({
  useRouter: () => ({
    push: vi.fn()
  })
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key
    })
  }
})

const formatLocalDate = (date: Date): string => {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

const createDashboardStats = (): DashboardStats => ({
  total_users: 0,
  today_new_users: 0,
  active_users: 0,
  hourly_active_users: 0,
  stats_updated_at: '',
  stats_stale: false,
  total_api_keys: 0,
  active_api_keys: 0,
  total_accounts: 0,
  normal_accounts: 0,
  error_accounts: 0,
  ratelimit_accounts: 0,
  overload_accounts: 0,
  total_requests: 0,
  total_input_tokens: 0,
  total_output_tokens: 0,
  total_cache_creation_tokens: 0,
  total_cache_read_tokens: 0,
  total_tokens: 0,
  total_cost: 0,
  total_actual_cost: 0,
  today_requests: 0,
  today_input_tokens: 0,
  today_output_tokens: 0,
  today_cache_creation_tokens: 0,
  today_cache_read_tokens: 0,
  today_tokens: 0,
  today_cost: 0,
  today_actual_cost: 0,
  average_duration_ms: 0,
  uptime: 0,
  rpm: 0,
  tpm: 0
})

describe('admin DashboardView', () => {
  beforeEach(() => {
    setActivePinia(createPinia())

    getSnapshotV2.mockReset()
    getModelStats.mockReset()
    getUserUsageTrend.mockReset()
    getUserSpendingRanking.mockReset()
    listUsage.mockReset()

    getSnapshotV2.mockResolvedValue({
      stats: createDashboardStats(),
      trend: [],
      models: []
    })
    getUserUsageTrend.mockResolvedValue({
      trend: [],
      start_date: '',
      end_date: '',
      granularity: 'hour'
    })
    getModelStats.mockResolvedValue({
      models: [],
      start_date: '1970-01-01',
      end_date: ''
    })
    getUserSpendingRanking.mockResolvedValue({
      ranking: [],
      total_actual_cost: 0,
      total_requests: 0,
      total_tokens: 0,
      start_date: '',
      end_date: ''
    })
    listUsage.mockResolvedValue({
      items: [],
      total: 0,
      page: 1,
      page_size: 10,
      total_pages: 0
    })
  })

  it('uses last 24 hours as default dashboard range', async () => {
    mount(DashboardView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          LoadingSpinner: true,
          Icon: true,
          DateRangePicker: true,
          Select: true,
          ModelDistributionChart: true,
          TokenUsageTrend: true,
          UsageTable: true,
          Line: true
        }
      }
    })

    await flushPromises()

    const now = new Date()
    const yesterday = new Date(now.getTime() - 24 * 60 * 60 * 1000)

    expect(getSnapshotV2).toHaveBeenCalledTimes(1)
    expect(getSnapshotV2).toHaveBeenCalledWith(expect.objectContaining({
      start_date: formatLocalDate(yesterday),
      end_date: formatLocalDate(now),
      granularity: 'hour',
      include_model_stats: false
    }))
  })

  it('loads model distribution over the full retained history', async () => {
    mount(DashboardView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          LoadingSpinner: true,
          Icon: true,
          DateRangePicker: true,
          Select: true,
          ModelDistributionChart: true,
          TokenUsageTrend: true,
          UsageTable: true,
          Line: true
        }
      }
    })

    await flushPromises()

    expect(getModelStats).toHaveBeenCalledTimes(1)
    expect(getModelStats).toHaveBeenCalledWith({
      start_date: '1970-01-01',
      end_date: formatLocalDate(new Date()),
      model_source: 'requested'
    })
  })

  it('does not render the Quick Actions section', async () => {
    const wrapper = mount(DashboardView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          LoadingSpinner: true,
          Icon: true,
          DateRangePicker: true,
          Select: true,
          ModelDistributionChart: true,
          TokenUsageTrend: true,
          UsageTable: true,
          Line: true
        }
      }
    })

    await flushPromises()

    expect(wrapper.text()).not.toContain('admin.dashboard.quickActions')
  })

  it('loads the latest persisted request logs', async () => {
    mount(DashboardView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          LoadingSpinner: true,
          Icon: true,
          DateRangePicker: true,
          Select: true,
          ModelDistributionChart: true,
          TokenUsageTrend: true,
          UsageTable: true,
          Line: true
        }
      }
    })

    await flushPromises()

    expect(listUsage).toHaveBeenCalledTimes(1)
    expect(listUsage).toHaveBeenCalledWith({
      page: 1,
      page_size: 10,
      exact_total: false,
      sort_by: 'created_at',
      sort_order: 'desc'
    })
  })

  it('shows account IDs without request IDs in recent requests', async () => {
    const wrapper = mount(DashboardView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          LoadingSpinner: true,
          Icon: true,
          DateRangePicker: true,
          Select: true,
          ModelDistributionChart: true,
          TokenUsageTrend: true,
          UsageTable: {
            name: 'UsageTable',
            props: ['columns', 'accountIdOnly'],
            template: '<div data-testid="recent-requests-table" />'
          },
          Line: true
        }
      }
    })

    await flushPromises()

    const table = wrapper.getComponent({ name: 'UsageTable' })
    expect(table.props('accountIdOnly')).toBe(true)
    expect(table.props('columns')).toEqual(expect.arrayContaining([{ key: 'account', label: 'admin.usage.account' }]))
    expect(table.props('columns')).not.toEqual(expect.arrayContaining([{ key: 'request_id', label: expect.any(String) }]))
  })
})
