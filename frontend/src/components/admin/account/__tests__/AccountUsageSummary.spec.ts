import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import AccountUsageSummary from '../AccountUsageSummary.vue'
import type { AntigravityUsageSummary } from '@/types'

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, params?: Record<string, unknown>) => {
        if (params && 'count' in params) {
          return `${key}:${params.count}`
        }
        if (params && 'time' in params) {
          return `${key}:${params.time}`
        }
        return key
      }
    })
  }
})

describe('AccountUsageSummary.vue', () => {
  const mockSummary: AntigravityUsageSummary = {
    eligible_accounts: 10,
    failed_accounts: 1,
    gemini_5h: { utilization: 42.5, sample_count: 9 },
    gemini_7d: { utilization: 15.0, sample_count: 9 },
    claude_5h: { utilization: 88.0, sample_count: 8 },
    claude_7d: { utilization: null, sample_count: 0 },
    updated_at: '2026-07-30T15:00:00Z'
  }

  it('renders four stable bar positions with provider grouping', () => {
    const wrapper = mount(AccountUsageSummary, {
      props: {
        summary: mockSummary,
        loading: false,
        error: null
      }
    })

    const progressBars = wrapper.findAll('[role="progressbar"]')
    expect(progressBars.length).toBe(4)

    // Gemini 5h: 42.5%
    expect(progressBars[0].attributes('aria-valuenow')).toBe('42.5')
    expect(wrapper.text()).toContain('42.5%')

    // Gemini 7d: 15%
    expect(progressBars[1].attributes('aria-valuenow')).toBe('15')
    expect(wrapper.text()).toContain('15%')

    // Claude 5h: 88%
    expect(progressBars[2].attributes('aria-valuenow')).toBe('88')
    expect(wrapper.text()).toContain('88%')

    // Claude 7d: null (unavailable state, preserved position)
    expect(progressBars[3].attributes('aria-valuenow')).toBe('0')
    expect(wrapper.text()).toContain('admin.accounts.summary.unavailable')
  })

  it('renders loading state with skeletons', () => {
    const wrapper = mount(AccountUsageSummary, {
      props: {
        summary: null,
        loading: true,
        error: null
      }
    })

    expect(wrapper.findAll('.animate-pulse').length).toBeGreaterThan(0)
    expect(wrapper.findAll('[role="progressbar"]').length).toBe(0)
  })

  it('renders error state and emits refresh on retry', async () => {
    const wrapper = mount(AccountUsageSummary, {
      props: {
        summary: null,
        loading: false,
        error: 'Network Error'
      }
    })

    expect(wrapper.text()).toContain('Network Error')
    const retryBtn = wrapper.find('button')
    expect(retryBtn.exists()).toBe(true)

    await retryBtn.trigger('click')
    expect(wrapper.emitted('refresh')).toBeTruthy()
  })

  it('renders empty state when no eligible accounts exist', () => {
    const emptySummary: AntigravityUsageSummary = {
      eligible_accounts: 0,
      failed_accounts: 0,
      gemini_5h: { utilization: null, sample_count: 0 },
      gemini_7d: { utilization: null, sample_count: 0 },
      claude_5h: { utilization: null, sample_count: 0 },
      claude_7d: { utilization: null, sample_count: 0 },
      updated_at: '2026-07-30T15:00:00Z'
    }

    const wrapper = mount(AccountUsageSummary, {
      props: {
        summary: emptySummary,
        loading: false,
        error: null
      }
    })

    expect(wrapper.text()).toContain('admin.accounts.summary.noEligibleAccounts')
  })

  it('emits refresh event when refresh button is clicked', async () => {
    const wrapper = mount(AccountUsageSummary, {
      props: {
        summary: mockSummary,
        loading: false,
        error: null
      }
    })

    const refreshBtn = wrapper.find('button[title="admin.accounts.summary.refresh"]')
    expect(refreshBtn.exists()).toBe(true)

    await refreshBtn.trigger('click')
    expect(wrapper.emitted('refresh')).toBeTruthy()
  })
})
