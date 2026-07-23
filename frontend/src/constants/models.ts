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
  accent: ModelAccent
}

// Add, remove, or replace models here; onboarding and user surfaces update automatically.
export const PRODUCT_MODELS: readonly ModelItem[] = [
  {
    id: 'gpt-5.6-sol',
    displayName: 'GPT-5.6 Sol',
    providerFamily: 'gpt',
    purposeEn: 'Flagship reasoning for complex analysis, architecture, and difficult coding work.',
    purposeZh: '旗舰推理模型，适用于复杂分析、系统架构和高难度编码工作',
    protocol: 'OpenAI-compatible API',
    endpointPath: '/v1/chat/completions',
    accent: {
      providerLabel: 'GPT',
      badgeTextClass: 'text-[#2b8a5e]',
      selectedBorderClass: 'border-[#2b8a5e]',
      selectedBgClass: 'bg-[#2b8a5e]/5 dark:bg-[#2b8a5e]/10',
      dotClass: 'bg-[#2b8a5e]',
      accentHex: '#2b8a5e'
    }
  },
  {
    id: 'gpt-5.6-terra',
    displayName: 'GPT-5.6 Terra',
    providerFamily: 'gpt',
    purposeEn: 'Balanced general-purpose model optimized for everyday production workloads.',
    purposeZh: '全能型通用模型，为日常生产负载进行优化',
    protocol: 'OpenAI-compatible API',
    endpointPath: '/v1/chat/completions',
    accent: {
      providerLabel: 'GPT',
      badgeTextClass: 'text-[#2b8a5e]',
      selectedBorderClass: 'border-[#2b8a5e]',
      selectedBgClass: 'bg-[#2b8a5e]/5 dark:bg-[#2b8a5e]/10',
      dotClass: 'bg-[#2b8a5e]',
      accentHex: '#2b8a5e'
    }
  },
  {
    id: 'gpt-5.6-luna',
    displayName: 'GPT-5.6 Luna',
    providerFamily: 'gpt',
    purposeEn: 'Fast, lightweight model for high-throughput and latency-sensitive applications.',
    purposeZh: '快速轻量模型，适用于高吞吐和低延迟场景',
    protocol: 'OpenAI-compatible API',
    endpointPath: '/v1/chat/completions',
    accent: {
      providerLabel: 'GPT',
      badgeTextClass: 'text-[#2b8a5e]',
      selectedBorderClass: 'border-[#2b8a5e]',
      selectedBgClass: 'bg-[#2b8a5e]/5 dark:bg-[#2b8a5e]/10',
      dotClass: 'bg-[#2b8a5e]',
      accentHex: '#2b8a5e'
    }
  },
  {
    id: 'gemini-3.1-pro',
    displayName: 'Gemini 3.1 Pro',
    providerFamily: 'gemini',
    purposeEn: "Google's most capable multimodal model with native tool use and 1M context.",
    purposeZh: 'Google 最强多模态模型，原生工具调用，支持 100 万上下文',
    protocol: 'Google Gemini API',
    endpointPath: '/antigravity/v1beta/models/gemini-3.1-pro:generateContent',
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
    id: 'gemini-3.6-flash',
    displayName: 'Gemini 3.6 Flash',
    providerFamily: 'gemini',
    purposeEn: 'Ultra-fast inference for real-time streaming and high-volume tasks.',
    purposeZh: '超高速推理，适用于实时流式和高并发任务',
    protocol: 'Google Gemini API',
    endpointPath: '/antigravity/v1beta/models/gemini-3.6-flash:generateContent',
    accent: {
      providerLabel: 'Gemini',
      badgeTextClass: 'text-[#2563eb]',
      selectedBorderClass: 'border-[#2563eb]',
      selectedBgClass: 'bg-[#2563eb]/5 dark:bg-[#2563eb]/10',
      dotClass: 'bg-[#2563eb]',
      accentHex: '#2563eb'
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
