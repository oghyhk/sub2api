import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { nextTick, ref } from 'vue'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string) => key,
    locale: ref('en')
  })
}))

import UserDashboardModelCatalog from '../UserDashboardModelCatalog.vue'

describe('UserDashboardModelCatalog', () => {
  it('supports search and provider filtering while exposing comparison details', async () => {
    const wrapper = mount(UserDashboardModelCatalog, {
      props: { totalKeys: 1 },
      global: {
        stubs: {
          Icon: { template: '<span />' },
          RouterLink: { template: '<a><slot /></a>' }
        }
      }
    })

    expect(wrapper.text()).toContain('GPT-5.6 Sol')
    expect(wrapper.text()).toContain('Gemini 3.6 Flash')
    expect(wrapper.text()).toContain('358,000 tokens')
    expect(wrapper.text()).toContain('dashboard.modelCatalog.pricingNote')

    await wrapper.get('select').setValue('gemini')
    await nextTick()
    expect(wrapper.text()).not.toContain('GPT-5.6 Sol')
    expect(wrapper.text()).toContain('Gemini 3.1 Pro')

    await wrapper.get('input[type="search"]').setValue('flash')
    await nextTick()
    expect(wrapper.text()).not.toContain('Gemini 3.1 Pro')
    expect(wrapper.text()).toContain('Gemini 3.6 Flash')
  })
})
