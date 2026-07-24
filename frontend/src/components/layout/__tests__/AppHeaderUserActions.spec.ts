import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const componentPath = resolve(dirname(fileURLToPath(import.meta.url)), '../AppHeader.vue')
const source = readFileSync(componentPath, 'utf8')

describe('AppHeader normal-user actions', () => {
  it('exposes Guide, Docs, and Pricing only through the normal-user action group', () => {
    expect(source).toContain('v-if="isNormalUser"')
    expect(source).toContain('data-testid="user-guide-action"')
    expect(source).toContain('data-testid="user-docs-action"')
    expect(source).toContain('data-testid="user-pricing-action"')
    expect(source).toContain('to="/docs"')
    expect(source).toContain('to="/pricing"')
    expect(source).toContain('@click="handleReplayGuide"')
    expect(source).toContain('const isNormalUser = computed(() => Boolean(user.value) && !authStore.isAdmin)')
  })
})
