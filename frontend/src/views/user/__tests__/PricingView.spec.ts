import { flushPromises, mount } from '@vue/test-utils'
import { ref } from 'vue'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import PricingView from '../PricingView.vue'

const { getAvailableMock, showErrorMock } = vi.hoisted(() => ({
  getAvailableMock: vi.fn(),
  showErrorMock: vi.fn(),
}))

vi.mock('@/api/channels', () => {
  return {
    default: { getAvailable: getAvailableMock },
  }
})

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({ showError: showErrorMock }),
}))

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
      locale: ref('en'),
    }),
  }
})

vi.mock('@/components/layout/AppLayout.vue', () => ({
  default: { template: '<div><slot /></div>' },
}))

describe('PricingView', () => {
  beforeEach(() => {
    getAvailableMock.mockReset()
    showErrorMock.mockReset()
    getAvailableMock.mockResolvedValue([
      {
        name: 'Primary',
        description: '',
        platforms: [
          {
            platform: 'openai',
            groups: [{ name: 'Standard' }],
            supported_models: [
              {
                name: 'gpt-5.6-sol',
                platform: 'openai',
                pricing: {
                  billing_mode: 'token',
                  input_price: 0.000003,
                  output_price: 0.000012,
                  cache_write_price: null,
                  cache_read_price: 0.0000003,
                  image_input_price: null,
                  image_output_price: null,
                  per_request_price: null,
                  intervals: [],
                },
              },
            ],
          },
        ],
      },
    ])
  })

  it('shows exactly the five product models and keeps pricing dimensions separate', async () => {
    const wrapper = mount(PricingView, {
      global: {
        stubs: {
          Icon: { template: '<span />' },
        },
      },
    })

    await flushPromises()

    for (const model of [
      'gpt-5.6-sol',
      'gpt-5.6-terra',
      'gpt-5.6-luna',
      'gemini-3.1-pro',
      'gemini-3.6-flash',
    ]) {
      expect(wrapper.find(`[data-testid="pricing-model-${model}"]`).exists()).toBe(true)
    }

    const sol = wrapper.get('[data-testid="pricing-model-gpt-5.6-sol"]').text()
    expect(sol).toContain('availableChannels.pricing.inputPrice')
    expect(sol).toContain('$3 availableChannels.pricing.unitPerMillion')
    expect(sol).toContain('availableChannels.pricing.outputPrice')
    expect(sol).toContain('$12 availableChannels.pricing.unitPerMillion')
    expect(sol).toContain('availableChannels.pricing.cacheReadPrice')
    expect(sol).toContain('$0.3 availableChannels.pricing.unitPerMillion')
    expect(sol).toContain('availableChannels.pricing.cacheWritePrice')
    expect(sol).toContain('pricingPage.unavailableShort')

    expect(wrapper.get('[data-testid="pricing-model-gemini-3.1-pro"]').text()).toContain(
      'pricingPage.unavailable',
    )
  })
})
