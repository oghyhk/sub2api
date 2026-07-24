<template>
  <AppLayout>
    <div class="space-y-5">
      <section class="rounded-xl border border-amber-200 bg-amber-50 px-4 py-3 dark:border-amber-800/60 dark:bg-amber-950/20">
        <div class="flex items-start gap-3">
          <Icon name="infoCircle" size="md" class="mt-0.5 flex-shrink-0 text-amber-700 dark:text-amber-300" />
          <div>
            <h2 class="text-sm font-semibold text-amber-900 dark:text-amber-100">
              {{ t('pricingPage.referenceTitle') }}
            </h2>
            <p class="mt-1 text-sm leading-6 text-amber-900/80 dark:text-amber-100/80">
              {{ t('pricingPage.referenceDescription') }}
            </p>
          </div>
        </div>
      </section>

      <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div class="relative w-full sm:max-w-sm">
          <Icon name="search" size="md" class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
          <input
            v-model="searchQuery"
            type="search"
            class="input pl-10"
            :placeholder="t('pricingPage.searchPlaceholder')"
          />
        </div>
        <button type="button" class="btn btn-secondary" :disabled="loading" @click="loadPricing">
          <Icon name="refresh" size="sm" class="mr-2" :class="{ 'animate-spin': loading }" />
          {{ t('common.refresh') }}
        </button>
      </div>

      <section class="overflow-hidden rounded-xl border border-gray-200 bg-white shadow-sm dark:border-dark-700 dark:bg-dark-900">
        <div
          v-if="loading"
          class="flex min-h-72 items-center justify-center"
          data-testid="pricing-loading"
        >
          <span class="h-8 w-8 animate-spin rounded-full border-2 border-gray-200 border-b-primary-600 dark:border-dark-700 dark:border-b-primary-400"></span>
        </div>

        <div
          v-else-if="filteredModels.length === 0"
          class="px-6 py-16 text-center text-sm text-gray-500 dark:text-gray-400"
        >
          {{ t('pricingPage.noMatches') }}
        </div>

        <article
          v-for="model in filteredModels"
          v-else
          :key="model.id"
          class="border-b border-gray-200 p-4 last:border-b-0 dark:border-dark-700 sm:p-5"
          :data-testid="`pricing-model-${model.id}`"
        >
          <div class="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
            <div class="min-w-0">
              <div class="flex flex-wrap items-center gap-2">
                <h2 class="font-mono text-base font-semibold text-gray-950 dark:text-white">
                  {{ model.id }}
                </h2>
                <span
                  class="rounded-md px-2 py-0.5 text-xs font-semibold"
                  :class="model.providerFamily === 'gpt'
                    ? 'bg-emerald-50 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300'
                    : 'bg-blue-50 text-blue-700 dark:bg-blue-900/30 dark:text-blue-300'"
                >
                  {{ model.accent.providerLabel }}
                </span>
                <span class="text-xs text-gray-500 dark:text-gray-400">
                  {{ formatContext(model.contextWindow) }} {{ t('pricingPage.context') }}
                </span>
              </div>
              <p class="mt-1 max-w-3xl text-sm leading-6 text-gray-600 dark:text-gray-300">
                {{ getModelPurpose(model, locale) }}
              </p>
            </div>
            <code class="flex-shrink-0 text-xs text-gray-500 dark:text-gray-400">{{ model.endpointPath }}</code>
          </div>

          <div
            v-if="offersByModel[model.id]?.length"
            class="mt-4 space-y-3"
          >
            <div
              v-for="offer in offersByModel[model.id]"
              :key="offer.key"
              class="rounded-lg border border-gray-200 dark:border-dark-700"
            >
              <div class="flex flex-wrap items-center justify-between gap-2 border-b border-gray-200 px-3 py-2 dark:border-dark-700">
                <div class="text-sm font-medium text-gray-900 dark:text-white">
                  {{ offer.channelName }}
                  <span class="ml-1 text-xs font-normal text-gray-500 dark:text-gray-400">
                    {{ offer.platform }}
                  </span>
                </div>
                <div class="text-xs text-gray-500 dark:text-gray-400">
                  {{ offer.groupNames.join(', ') || t('pricingPage.noGroup') }}
                </div>
              </div>

              <div v-if="offer.pricing" class="grid grid-cols-2 gap-px bg-gray-200 dark:bg-dark-700 sm:grid-cols-4">
                <PricingMetric
                  :label="t('availableChannels.pricing.inputPrice')"
                  :value="priceDisplay(offer.pricing.input_price, offer.pricing.billing_mode)"
                />
                <PricingMetric
                  :label="t('availableChannels.pricing.outputPrice')"
                  :value="priceDisplay(offer.pricing.output_price, offer.pricing.billing_mode)"
                />
                <PricingMetric
                  :label="t('availableChannels.pricing.cacheReadPrice')"
                  :value="priceDisplay(offer.pricing.cache_read_price, offer.pricing.billing_mode)"
                />
                <PricingMetric
                  :label="t('availableChannels.pricing.cacheWritePrice')"
                  :value="priceDisplay(offer.pricing.cache_write_price, offer.pricing.billing_mode)"
                />
              </div>
              <div v-else class="px-3 py-4 text-sm text-gray-500 dark:text-gray-400">
                {{ t('pricingPage.unavailable') }}
              </div>

              <div
                v-if="offer.pricing?.intervals?.length"
                class="border-t border-gray-200 px-3 py-2 text-xs text-gray-600 dark:border-dark-700 dark:text-gray-300"
              >
                <span class="font-semibold">{{ t('availableChannels.pricing.intervals') }}:</span>
                <span
                  v-for="(interval, index) in offer.pricing.intervals"
                  :key="index"
                  class="ml-2 inline-block"
                >
                  {{ interval.tier_label || formatRange(interval.min_tokens, interval.max_tokens) }}
                  {{ t('availableChannels.pricing.inputPrice') }}
                  {{ priceDisplay(interval.input_price, offer.pricing.billing_mode) }},
                  {{ t('availableChannels.pricing.outputPrice') }}
                  {{ priceDisplay(interval.output_price, offer.pricing.billing_mode) }}
                </span>
              </div>
            </div>
          </div>

          <div
            v-else
            class="mt-4 rounded-lg border border-dashed border-gray-300 px-4 py-4 text-sm text-gray-500 dark:border-dark-600 dark:text-gray-400"
          >
            {{ t('pricingPage.unavailable') }}
          </div>
        </article>
      </section>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, defineComponent, h, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import userChannelsAPI, {
  type UserAvailableChannel,
  type UserSupportedModelPricing,
} from '@/api/channels'
import { BILLING_MODE_IMAGE, BILLING_MODE_PER_REQUEST } from '@/constants/channel'
import { PRODUCT_MODELS, getModelPurpose } from '@/constants/models'
import { formatScaled } from '@/utils/pricing'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'

interface PricingOffer {
  key: string
  channelName: string
  platform: string
  groupNames: string[]
  pricing: UserSupportedModelPricing | null
}

const PricingMetric = defineComponent({
  props: {
    label: { type: String, required: true },
    value: { type: String, required: true },
  },
  setup(props) {
    return () => h('div', { class: 'bg-white px-3 py-3 dark:bg-dark-900' }, [
      h('div', { class: 'text-xs text-gray-500 dark:text-gray-400' }, props.label),
      h('div', { class: 'mt-1 font-mono text-sm font-semibold text-gray-900 dark:text-white' }, props.value),
    ])
  },
})

const { t, locale } = useI18n()
const appStore = useAppStore()
const channels = ref<UserAvailableChannel[]>([])
const loading = ref(false)
const searchQuery = ref('')

const offersByModel = computed<Record<string, PricingOffer[]>>(() => {
  const result: Record<string, PricingOffer[]> = Object.fromEntries(
    PRODUCT_MODELS.map((model) => [model.id, []]),
  )

  for (const channel of channels.value) {
    for (const section of channel.platforms) {
      for (const supportedModel of section.supported_models) {
        if (!result[supportedModel.name]) continue
        result[supportedModel.name].push({
          key: `${channel.name}-${section.platform}-${supportedModel.name}-${result[supportedModel.name].length}`,
          channelName: channel.name,
          platform: section.platform,
          groupNames: section.groups.map((group) => group.name),
          pricing: supportedModel.pricing,
        })
      }
    }
  }

  return result
})

const filteredModels = computed(() => {
  const query = searchQuery.value.trim().toLowerCase()
  if (!query) return PRODUCT_MODELS
  return PRODUCT_MODELS.filter((model) =>
    [
      model.id,
      model.displayName,
      model.providerFamily,
      getModelPurpose(model, locale.value),
      model.protocol,
    ].some((value) => value.toLowerCase().includes(query)),
  )
})

function formatContext(tokens: number): string {
  return new Intl.NumberFormat(locale.value).format(tokens)
}

function priceDisplay(value: number | null, billingMode: string): string {
  if (value == null) return t('pricingPage.unavailableShort')
  const perRequest = billingMode === BILLING_MODE_PER_REQUEST || billingMode === BILLING_MODE_IMAGE
  return perRequest
    ? `${formatScaled(value, 1)} ${t('availableChannels.pricing.unitPerRequest')}`
    : `${formatScaled(value, 1_000_000)} ${t('availableChannels.pricing.unitPerMillion')}`
}

function formatRange(min: number, max: number | null): string {
  return `(${min}, ${max == null ? '∞' : max}]`
}

async function loadPricing() {
  loading.value = true
  try {
    channels.value = await userChannelsAPI.getAvailable()
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('pricingPage.loadFailed')))
  } finally {
    loading.value = false
  }
}

onMounted(loadPricing)
</script>
