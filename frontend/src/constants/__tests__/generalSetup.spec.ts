import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'
import {
  buildGeneralSetupMarkdown,
  getProductModelIds,
  resolveGeneralSetupUrls,
} from '../generalSetup'

describe('general setup guide', () => {
  it('resolves the distinct OpenAI-compatible and native Gemini bases', () => {
    expect(resolveGeneralSetupUrls('https://api.k2598.com/v1')).toEqual({
      serviceRoot: 'https://api.k2598.com',
      openAIBase: 'https://api.k2598.com/v1',
      nativeGeminiBase: 'https://api.k2598.com/antigravity/v1beta',
    })
  })

  it('keeps the public fetchable Markdown file aligned with the product registry', () => {
    const currentDir = dirname(fileURLToPath(import.meta.url))
    const publicGuide = readFileSync(resolve(currentDir, '../../../public/docs/setup.md'), 'utf8')
    const generatedGuide = buildGeneralSetupMarkdown('https://api.k2598.com')

    expect(getProductModelIds('gemini')).toEqual([
      'gemini-3.6-flash',
      'gemini-3.1-flash-lite',
    ])
    expect(publicGuide).toContain('claude-opus-4-6')
    expect(publicGuide).toContain('gemini-3.6-flash')
    expect(publicGuide).toContain('gemini-3.1-flash-lite')
    expect(publicGuide).toContain('Authorization: Bearer YOUR_API_KEY')
    expect(publicGuide).toContain('x-goog-api-key: YOUR_API_KEY')
    expect(generatedGuide).toContain('gemini-3.1-flash-lite')
    expect(generatedGuide).toContain('Authorization: Bearer YOUR_API_KEY')
  })
})
