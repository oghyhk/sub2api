import { flushPromises, mount } from '@vue/test-utils'
import { ref } from 'vue'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import SetupDocsView from '../SetupDocsView.vue'

const { copyToClipboardMock } = vi.hoisted(() => ({
  copyToClipboardMock: vi.fn().mockResolvedValue(true),
}))

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
    }),
  }
})

vi.mock('@/components/layout/AppLayout.vue', () => ({
  default: { template: '<div><slot /></div>' },
}))

vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({
    copied: ref(false),
    copyToClipboard: copyToClipboardMock,
  }),
}))

describe('SetupDocsView', () => {
  beforeEach(() => {
    copyToClipboardMock.mockClear()
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
      ok: true,
      text: () => Promise.resolve(`# Sub2API General API Setup

\`https://api.k2598.com/v1\`

\`https://api.k2598.com/antigravity/v1beta\`
`),
    }))
  })

  it('renders the public guide and copies its agent-fetchable URL', async () => {
    const wrapper = mount(SetupDocsView, {
      global: {
        stubs: {
          Icon: { template: '<span />' },
        },
      },
    })

    await flushPromises()

    expect(fetch).toHaveBeenCalledWith('/docs/setup.md', {
      headers: { Accept: 'text/markdown' },
    })
    expect(wrapper.get('[data-testid="setup-doc-content"]').text()).toContain(
      'https://api.k2598.com/antigravity/v1beta',
    )

    await wrapper.get('[data-testid="copy-agent-guide-link"]').trigger('click')
    expect(copyToClipboardMock).toHaveBeenCalledWith(
      `${window.location.origin}/docs/setup.md`,
      'setupDocs.linkCopied',
    )
  })
})
