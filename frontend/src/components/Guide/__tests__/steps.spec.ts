import { describe, expect, it } from 'vitest'
import { getUserSteps } from '../steps'

describe('normal-user onboarding steps', () => {
  it('continues from key creation through Set Up and the setup handoff', () => {
    const steps = getUserSteps((key) => key)
    const elements = steps.map((step) => step.element)

    expect(elements).toContain('[data-tour="keys-create-btn"]')
    expect(elements).toContain('[data-tour="key-form-submit"]')
    expect(elements).toContain('[data-tour="key-setup-btn"]')
    expect(elements).toContain('[data-testid="setup-handoff"]')
    expect(elements.indexOf('[data-tour="key-setup-btn"]')).toBeGreaterThan(
      elements.indexOf('[data-tour="key-form-submit"]'),
    )
  })
})
