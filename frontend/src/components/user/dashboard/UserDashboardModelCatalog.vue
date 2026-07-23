<template>
  <section class="border-y border-[#d2d2d7] py-5 dark:border-[#2c2c2e]">
    <!-- Header -->
    <div class="mb-4 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
      <div>
        <h2 class="text-lg font-semibold text-[#1d1d1f] dark:text-[#f5f5f7]">
          {{ t('dashboard.modelCatalog.title') }}
        </h2>
        <p class="text-xs text-[#6e6e73] dark:text-[#98989d]">
          {{ t('dashboard.modelCatalog.subtitle') }}
        </p>
      </div>

      <router-link
        v-if="totalKeys > 0"
        to="/keys"
        class="inline-flex min-h-[44px] items-center justify-center gap-2 rounded-[8px] bg-[#0071e3] px-4 text-xs font-medium text-white transition-all duration-200 hover:bg-[#0077ed] focus:outline-none focus:ring-2 focus:ring-[#0071e3]/40"
      >
        <Icon name="key" size="sm" />
        <span>{{ totalKeys === 0 ? t('dashboard.modelCatalog.createFirstKey') : t('dashboard.modelCatalog.manageKeys') }}</span>
      </router-link>
    </div>

    <!-- Zero Keys Callout -->
    <div
      v-if="totalKeys === 0"
      class="mb-4 flex flex-col items-start justify-between gap-3 rounded-[8px] border border-amber-200 bg-amber-50/50 p-4 dark:border-amber-900/40 dark:bg-amber-950/20 sm:flex-row sm:items-center"
    >
      <div class="flex items-center gap-3">
        <div class="flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-full bg-amber-100 text-amber-600 dark:bg-amber-900/40 dark:text-amber-400">
          <Icon name="key" size="md" />
        </div>
        <p class="text-xs font-medium text-[#1d1d1f] dark:text-[#f5f5f7]">
          {{ t('dashboard.modelCatalog.noKeysNotice') }}
        </p>
      </div>
      <router-link
        to="/keys"
        class="inline-flex min-h-[44px] min-w-[120px] items-center justify-center rounded-[8px] bg-[#0071e3] px-4 text-xs font-medium text-white transition-colors hover:bg-[#0077ed]"
      >
        {{ t('dashboard.modelCatalog.createFirstKey') }}
      </router-link>
    </div>

    <div class="mb-3 flex flex-col gap-2 sm:flex-row">
      <label class="relative flex-1">
        <span class="sr-only">{{ t('dashboard.modelCatalog.search') }}</span>
        <Icon name="search" size="sm" class="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-[#6e6e73] dark:text-[#98989d]" />
        <input
          v-model="searchQuery"
          type="search"
          :placeholder="t('dashboard.modelCatalog.searchPlaceholder')"
          class="min-h-[44px] w-full rounded-[8px] border border-[#d2d2d7] bg-white pl-9 pr-3 text-sm text-[#1d1d1f] outline-none focus:border-[#0071e3] focus:ring-2 focus:ring-[#0071e3]/20 dark:border-[#2c2c2e] dark:bg-[#1c1c1e] dark:text-[#f5f5f7]"
        />
      </label>
      <select
        v-model="providerFilter"
        :aria-label="t('dashboard.modelCatalog.providerFilter')"
        class="min-h-[44px] rounded-[8px] border border-[#d2d2d7] bg-white px-3 text-sm text-[#1d1d1f] outline-none focus:border-[#0071e3] focus:ring-2 focus:ring-[#0071e3]/20 dark:border-[#2c2c2e] dark:bg-[#1c1c1e] dark:text-[#f5f5f7]"
      >
        <option value="all">{{ t('dashboard.modelCatalog.allProviders') }}</option>
        <option value="gpt">GPT</option>
        <option value="gemini">Gemini</option>
      </select>
    </div>

    <!-- Model catalog -->
    <div class="grid grid-cols-1 gap-2.5 sm:grid-cols-2 lg:grid-cols-5">
      <button
        v-for="m in filteredModels"
        :key="m.id"
        @click="selectedModelId = m.id"
        :aria-pressed="selectedModelId === m.id"
        class="flex min-h-[44px] cursor-pointer flex-col items-center justify-center rounded-[8px] border p-3 text-center transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-[#0071e3]/40"
        :class="selectedModelId === m.id
          ? `${m.accent.selectedBorderClass} ${m.accent.selectedBgClass}`
          : 'border-[#d2d2d7] bg-white hover:border-[#a1a1a6] dark:border-[#2c2c2e] dark:bg-[#1c1c1e] dark:hover:border-[#3a3a3c]'"
      >
        <span
          class="text-[10px] font-semibold uppercase tracking-wider"
          :class="m.accent.badgeTextClass"
        >
          {{ m.accent.providerLabel }}
        </span>
        <span class="mt-0.5 text-xs font-medium text-[#1d1d1f] dark:text-[#f5f5f7]">
          {{ m.displayName }}
        </span>
      </button>
    </div>
    <p v-if="filteredModels.length === 0" class="py-4 text-sm text-[#6e6e73] dark:text-[#98989d]">
      {{ t('dashboard.modelCatalog.noMatches') }}
    </p>

    <!-- Selected Model Detail Workspace -->
    <div
      v-if="selectedModel"
      class="mt-4 border-l-2 bg-[#f5f5f7] p-4 dark:bg-[#1c1c1e]"
      :style="{ borderColor: selectedModel.accent.accentHex }"
    >
      <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div class="flex min-w-0 items-center gap-2">
          <span class="h-2.5 w-2.5 rounded-full" :class="selectedModel.accent.dotClass"></span>
          <span class="text-sm font-semibold text-[#1d1d1f] dark:text-[#f5f5f7]">{{ selectedModel.displayName }}</span>
          <span class="truncate rounded bg-white px-2 py-0.5 font-mono text-xs text-[#6e6e73] dark:bg-[#2c2c2e] dark:text-[#98989d]">{{ selectedModel.protocol }}</span>
        </div>
        <div class="flex items-center gap-2 font-mono text-xs text-[#6e6e73] dark:text-[#98989d]">
          <span class="rounded bg-white px-2 py-1 select-all dark:bg-[#2c2c2e]">{{ selectedModel.endpointPath }}</span>
          <button
            @click="copyEndpoint(selectedModel.endpointPath)"
            :aria-label="copied ? t('dashboard.modelCatalog.endpointCopied') : t('dashboard.modelCatalog.copyEndpoint')"
            class="flex min-h-[44px] min-w-[44px] items-center justify-center rounded-[6px] border border-[#d2d2d7] bg-white p-2 text-[#6e6e73] transition-colors hover:text-[#1d1d1f] dark:border-[#2c2c2e] dark:bg-[#2c2c2e] dark:text-[#98989d] dark:hover:text-[#f5f5f7]"
            :title="copied ? t('dashboard.modelCatalog.endpointCopied') : t('dashboard.modelCatalog.copyEndpoint')"
          >
            <Icon v-if="copied" name="check" size="xs" class="text-emerald-500" />
            <Icon v-else name="copy" size="xs" />
          </button>
        </div>
      </div>
      <p class="mt-2 text-xs leading-relaxed text-[#6e6e73] dark:text-[#98989d]">
        {{ getModelPurpose(selectedModel, locale) }}
      </p>
      <dl class="mt-3 grid grid-cols-2 gap-x-4 gap-y-2 text-xs text-[#6e6e73] dark:text-[#98989d] sm:grid-cols-4">
        <div><dt class="font-medium text-[#1d1d1f] dark:text-[#f5f5f7]">{{ t('dashboard.modelCatalog.context') }}</dt><dd>{{ formatContext(selectedModel.contextWindow) }}</dd></div>
        <div><dt class="font-medium text-[#1d1d1f] dark:text-[#f5f5f7]">{{ t('dashboard.modelCatalog.input') }}</dt><dd>{{ selectedModel.inputModalities.join(', ') }}</dd></div>
        <div><dt class="font-medium text-[#1d1d1f] dark:text-[#f5f5f7]">{{ t('dashboard.modelCatalog.output') }}</dt><dd>{{ selectedModel.outputModalities.join(', ') }}</dd></div>
        <div><dt class="font-medium text-[#1d1d1f] dark:text-[#f5f5f7]">{{ t('dashboard.modelCatalog.referencePricing') }}</dt><dd>{{ selectedModel.referencePricing?.source }}</dd></div>
      </dl>
      <p class="mt-2 text-[11px] text-[#6e6e73] dark:text-[#98989d]">{{ t('dashboard.modelCatalog.pricingNote') }}</p>
    </div>
  </section>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import { PRODUCT_MODELS, getModelById, getModelPurpose, DEFAULT_MODEL_ID } from '@/constants/models'

withDefaults(
  defineProps<{
    totalKeys?: number
  }>(),
  {
    totalKeys: 0
  }
)

const { t, locale } = useI18n()
const selectedModelId = ref(DEFAULT_MODEL_ID)
const copied = ref(false)
const searchQuery = ref('')
const providerFilter = ref<'all' | 'gpt' | 'gemini'>('all')
const filteredModels = computed(() => {
  const query = searchQuery.value.trim().toLowerCase()
  return PRODUCT_MODELS.filter((model) => {
    const providerMatch = providerFilter.value === 'all' || model.providerFamily === providerFilter.value
    const textMatch = !query || [model.id, model.displayName, model.protocol, getModelPurpose(model, locale.value)].some((value) => value.toLowerCase().includes(query))
    return providerMatch && textMatch
  })
})
const selectedModel = computed(() => {
  const selected = filteredModels.value.find((model) => model.id === selectedModelId.value)
  return selected || filteredModels.value[0] || getModelById(selectedModelId.value)
})

function formatContext(value: number): string {
  return `${new Intl.NumberFormat(locale.value).format(value)} tokens`
}

async function copyEndpoint(path: string) {
  try {
    await navigator.clipboard.writeText(path)
    copied.value = true
    setTimeout(() => {
      copied.value = false
    }, 2000)
  } catch (e) {
    console.error('Failed to copy endpoint:', e)
  }
}
</script>
