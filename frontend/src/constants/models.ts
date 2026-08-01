export interface ModelAccent {
  providerLabel: string
  badgeTextClass: string
  selectedBorderClass: string
  selectedBgClass: string
  dotClass: string
  accentHex: string
}

export interface ModelItem {
  id: string
  displayName: string
  providerFamily: 'gpt' | 'gemini'
  purposeEn: string
  purposeZh: string
  protocol: string
  endpointPath: string
  contextWindow: number
  inputModalities: readonly string[]
  outputModalities: readonly string[]
  referencePricing?: {
    inputPerMillion: number | null
    outputPerMillion: number | null
    cacheReadPerMillion: number | null
    source: string
  }
  accent: ModelAccent
}

// Add, remove, or replace models here; onboarding and user surfaces update automatically.
export const PRODUCT_MODELS: readonly ModelItem[] = [
  {
    id: 'gemini-3.6-flash',
    displayName: 'Gemini 3.6 Flash',
    providerFamily: 'gemini',
    purposeEn: 'Ultra-fast multimodal reasoning with 1M context (high effort).',
    purposeZh: '超高速多模态推理，支持 100 万上下文（High Effort）',
    protocol: 'Google Gemini API',
    endpointPath: '/antigravity/v1beta/models/gemini-3.6-flash:generateContent',
    contextWindow: 1048576,
    inputModalities: ['text', 'image', 'pdf'],
    outputModalities: ['text'],
    referencePricing: { inputPerMillion: null, outputPerMillion: null, cacheReadPerMillion: null, source: 'Account pricing' },
    accent: {
      providerLabel: 'Gemini',
      badgeTextClass: 'text-[#2563eb]',
      selectedBorderClass: 'border-[#2563eb]',
      selectedBgClass: 'bg-[#2563eb]/5 dark:bg-[#2563eb]/10',
      dotClass: 'bg-[#2563eb]',
      accentHex: '#2563eb'
    }
  },
  {
    id: 'gemini-3.1-flash-lite',
    displayName: 'Gemini 3.1 Flash-Lite',
    providerFamily: 'gemini',
    purposeEn: 'Efficient multimodal work for high-throughput and low-latency tasks.',
    purposeZh: '适合高吞吐、低延迟任务的高效多模态模型',
    protocol: 'Google Gemini API',
    endpointPath: '/antigravity/v1beta/models/gemini-3.1-flash-lite:generateContent',
    contextWindow: 1048576,
    inputModalities: ['text', 'image', 'pdf'],
    outputModalities: ['text'],
    referencePricing: { inputPerMillion: null, outputPerMillion: null, cacheReadPerMillion: null, source: 'Account pricing' },
    accent: {
      providerLabel: 'Gemini',
      badgeTextClass: 'text-[#0f766e]',
      selectedBorderClass: 'border-[#0f766e]',
      selectedBgClass: 'bg-[#0f766e]/5 dark:bg-[#0f766e]/10',
      dotClass: 'bg-[#0f766e]',
      accentHex: '#0f766e'
    }
  },
  {
    id: 'claude-opus-4-6',
    displayName: 'Claude Opus 4.6',
    providerFamily: 'gpt',
    purposeEn: 'Anthropic flagship frontier reasoning model (high effort).',
    purposeZh: 'Anthropic 旗舰前沿推理模型（High Effort）',
    protocol: 'OpenAI-compatible API',
    endpointPath: '/v1/chat/completions',
    contextWindow: 1048576,
    inputModalities: ['text', 'image'],
    outputModalities: ['text'],
    referencePricing: { inputPerMillion: null, outputPerMillion: null, cacheReadPerMillion: null, source: 'Account pricing' },
    accent: {
      providerLabel: 'Claude',
      badgeTextClass: 'text-[#d97706]',
      selectedBorderClass: 'border-[#d97706]',
      selectedBgClass: 'bg-[#d97706]/5 dark:bg-[#d97706]/10',
      dotClass: 'bg-[#d97706]',
      accentHex: '#d97706'
    }
  }
] as const

export const DEFAULT_MODEL_ID = PRODUCT_MODELS[0].id

export function getModelById(id: string): ModelItem {
  return PRODUCT_MODELS.find((m) => m.id === id) || PRODUCT_MODELS[0]
}

export function getModelPurpose(model: ModelItem, locale: string): string {
  return locale.toLowerCase().startsWith('zh') ? model.purposeZh : model.purposeEn
}
