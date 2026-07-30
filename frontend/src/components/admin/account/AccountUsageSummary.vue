<template>
  <div class="mb-4 rounded-xl border border-gray-200 bg-white p-4 shadow-sm dark:border-dark-700 dark:bg-dark-800">
    <!-- Header / Context Bar -->
    <div class="mb-3 flex flex-wrap items-center justify-between gap-2">
      <div class="flex items-center gap-2">
        <h3 class="text-xs font-bold uppercase tracking-wider text-gray-700 dark:text-gray-200">
          {{ t('admin.accounts.summary.title') }}
        </h3>
        <span
          v-if="summary && !loading && !error"
          class="inline-flex items-center gap-1 rounded-full bg-gray-100 px-2 py-0.5 text-[11px] font-medium text-gray-600 dark:bg-dark-700 dark:text-gray-300"
          :title="t('admin.accounts.summary.explanationTooltip')"
        >
          <span>{{ t('admin.accounts.summary.sampledAccounts', { count: summary.eligible_accounts }) }}</span>
          <span v-if="summary.failed_accounts > 0" class="text-amber-600 dark:text-amber-400">
            ({{ t('admin.accounts.summary.failedAccounts', { count: summary.failed_accounts }) }})
          </span>
        </span>
      </div>

      <div class="flex items-center gap-2">
        <span v-if="formattedUpdatedAt" class="text-[11px] text-gray-400 dark:text-gray-500">
          {{ t('admin.accounts.summary.updatedAt', { time: formattedUpdatedAt }) }}
        </span>
        <button
          type="button"
          class="rounded p-1 text-gray-400 hover:bg-gray-100 hover:text-gray-600 dark:hover:bg-dark-700 dark:hover:text-gray-200"
          :disabled="loading"
          :title="t('admin.accounts.summary.refresh')"
          @click="emit('refresh')"
        >
          <Icon name="refresh" size="xs" :class="{ 'animate-spin': loading }" />
        </button>
      </div>
    </div>

    <!-- Error State -->
    <div
      v-if="error && !loading"
      class="flex items-center justify-between rounded-lg border border-red-200 bg-red-50 p-3 text-xs text-red-700 dark:border-red-900/40 dark:bg-red-900/20 dark:text-red-300"
    >
      <span>{{ error }}</span>
      <button
        type="button"
        class="rounded bg-red-100 px-2 py-1 font-medium text-red-800 hover:bg-red-200 dark:bg-red-800/40 dark:text-red-200 dark:hover:bg-red-800/60"
        @click="emit('refresh')"
      >
        {{ t('admin.accounts.summary.retry') }}
      </button>
    </div>

    <!-- Loading Skeleton State -->
    <div v-else-if="loading && !summary" class="grid grid-cols-1 gap-4 md:grid-cols-2">
      <div v-for="i in 2" :key="i" class="space-y-3 rounded-lg border border-gray-100 bg-gray-50/50 p-3 dark:border-dark-700/60 dark:bg-dark-900/40">
        <div class="h-4 w-24 animate-pulse rounded bg-gray-200 dark:bg-dark-700"></div>
        <div class="space-y-2">
          <div v-for="j in 2" :key="j" class="space-y-1">
            <div class="flex justify-between">
              <div class="h-3 w-16 animate-pulse rounded bg-gray-200 dark:bg-dark-700"></div>
              <div class="h-3 w-12 animate-pulse rounded bg-gray-200 dark:bg-dark-700"></div>
            </div>
            <div class="h-2.5 w-full animate-pulse rounded-full bg-gray-200 dark:bg-dark-700"></div>
          </div>
        </div>
      </div>
    </div>

    <!-- Empty State (No accounts) -->
    <div
      v-else-if="summary && summary.eligible_accounts === 0"
      class="py-4 text-center text-xs text-gray-400 dark:text-gray-500"
    >
      {{ t('admin.accounts.summary.noEligibleAccounts') }}
    </div>

    <!-- Data Display Grid -->
    <div v-else-if="summary" class="grid grid-cols-1 gap-4 md:grid-cols-2">
      <!-- Gemini Provider Card -->
      <div class="rounded-lg border border-emerald-100 bg-emerald-50/30 p-3.5 dark:border-emerald-900/30 dark:bg-emerald-950/10">
        <div class="mb-2.5 flex items-center justify-between">
          <div class="flex items-center gap-1.5 font-semibold text-emerald-800 dark:text-emerald-300 text-xs">
            <span class="inline-block h-2 w-2 rounded-full bg-emerald-500"></span>
            {{ t('admin.accounts.usageWindow.gemini') }}
          </div>
          <span class="text-[10px] text-gray-400 dark:text-gray-500">
            {{ t('admin.accounts.summary.providerSubtitle') }}
          </span>
        </div>

        <div class="space-y-3">
          <!-- Gemini 5h -->
          <div class="space-y-1">
            <div class="flex items-center justify-between text-xs">
              <span class="font-medium text-gray-700 dark:text-gray-300 flex items-center gap-1.5">
                {{ t('admin.accounts.summary.fiveHourWindow') }}
                <span v-if="summary.gemini_5h.sample_count > 0" class="text-[10px] text-gray-400 font-normal">
                  ({{ t('admin.accounts.summary.sampleCount', { count: summary.gemini_5h.sample_count }) }})
                </span>
              </span>
              <span :class="['font-semibold text-xs', getPercentTextClass(summary.gemini_5h.utilization)]">
                {{ formatPercent(summary.gemini_5h.utilization) }}
              </span>
            </div>
            <div
              class="h-2.5 w-full overflow-hidden rounded-full bg-gray-200 dark:bg-dark-700"
              role="progressbar"
              :aria-valuenow="summary.gemini_5h.utilization ?? 0"
              aria-valuemin="0"
              aria-valuemax="100"
              :aria-label="t('admin.accounts.usageWindow.gemini5hHint')"
            >
              <div
                v-if="summary.gemini_5h.utilization !== null"
                :class="['h-full transition-all duration-500', getBarColorClass(summary.gemini_5h.utilization, 'emerald')]"
                :style="{ width: `${clampPercent(summary.gemini_5h.utilization)}%` }"
              ></div>
            </div>
          </div>

          <!-- Gemini 7d -->
          <div class="space-y-1">
            <div class="flex items-center justify-between text-xs">
              <span class="font-medium text-gray-700 dark:text-gray-300 flex items-center gap-1.5">
                {{ t('admin.accounts.summary.sevenDayWindow') }}
                <span v-if="summary.gemini_7d.sample_count > 0" class="text-[10px] text-gray-400 font-normal">
                  ({{ t('admin.accounts.summary.sampleCount', { count: summary.gemini_7d.sample_count }) }})
                </span>
              </span>
              <span :class="['font-semibold text-xs', getPercentTextClass(summary.gemini_7d.utilization)]">
                {{ formatPercent(summary.gemini_7d.utilization) }}
              </span>
            </div>
            <div
              class="h-2.5 w-full overflow-hidden rounded-full bg-gray-200 dark:bg-dark-700"
              role="progressbar"
              :aria-valuenow="summary.gemini_7d.utilization ?? 0"
              aria-valuemin="0"
              aria-valuemax="100"
              :aria-label="t('admin.accounts.usageWindow.gemini7dHint')"
            >
              <div
                v-if="summary.gemini_7d.utilization !== null"
                :class="['h-full transition-all duration-500', getBarColorClass(summary.gemini_7d.utilization, 'emerald')]"
                :style="{ width: `${clampPercent(summary.gemini_7d.utilization)}%` }"
              ></div>
            </div>
          </div>
        </div>
      </div>

      <!-- Claude Provider Card -->
      <div class="rounded-lg border border-amber-100 bg-amber-50/30 p-3.5 dark:border-amber-900/30 dark:bg-amber-950/10">
        <div class="mb-2.5 flex items-center justify-between">
          <div class="flex items-center gap-1.5 font-semibold text-amber-800 dark:text-amber-300 text-xs">
            <span class="inline-block h-2 w-2 rounded-full bg-amber-500"></span>
            {{ t('admin.accounts.usageWindow.claude') }}
          </div>
          <span class="text-[10px] text-gray-400 dark:text-gray-500">
            {{ t('admin.accounts.summary.providerSubtitle') }}
          </span>
        </div>

        <div class="space-y-3">
          <!-- Claude 5h -->
          <div class="space-y-1">
            <div class="flex items-center justify-between text-xs">
              <span class="font-medium text-gray-700 dark:text-gray-300 flex items-center gap-1.5">
                {{ t('admin.accounts.summary.fiveHourWindow') }}
                <span v-if="summary.claude_5h.sample_count > 0" class="text-[10px] text-gray-400 font-normal">
                  ({{ t('admin.accounts.summary.sampleCount', { count: summary.claude_5h.sample_count }) }})
                </span>
              </span>
              <span :class="['font-semibold text-xs', getPercentTextClass(summary.claude_5h.utilization)]">
                {{ formatPercent(summary.claude_5h.utilization) }}
              </span>
            </div>
            <div
              class="h-2.5 w-full overflow-hidden rounded-full bg-gray-200 dark:bg-dark-700"
              role="progressbar"
              :aria-valuenow="summary.claude_5h.utilization ?? 0"
              aria-valuemin="0"
              aria-valuemax="100"
              :aria-label="t('admin.accounts.usageWindow.claude5hHint')"
            >
              <div
                v-if="summary.claude_5h.utilization !== null"
                :class="['h-full transition-all duration-500', getBarColorClass(summary.claude_5h.utilization, 'amber')]"
                :style="{ width: `${clampPercent(summary.claude_5h.utilization)}%` }"
              ></div>
            </div>
          </div>

          <!-- Claude 7d -->
          <div class="space-y-1">
            <div class="flex items-center justify-between text-xs">
              <span class="font-medium text-gray-700 dark:text-gray-300 flex items-center gap-1.5">
                {{ t('admin.accounts.summary.sevenDayWindow') }}
                <span v-if="summary.claude_7d.sample_count > 0" class="text-[10px] text-gray-400 font-normal">
                  ({{ t('admin.accounts.summary.sampleCount', { count: summary.claude_7d.sample_count }) }})
                </span>
              </span>
              <span :class="['font-semibold text-xs', getPercentTextClass(summary.claude_7d.utilization)]">
                {{ formatPercent(summary.claude_7d.utilization) }}
              </span>
            </div>
            <div
              class="h-2.5 w-full overflow-hidden rounded-full bg-gray-200 dark:bg-dark-700"
              role="progressbar"
              :aria-valuenow="summary.claude_7d.utilization ?? 0"
              aria-valuemin="0"
              aria-valuemax="100"
              :aria-label="t('admin.accounts.usageWindow.claude7dHint')"
            >
              <div
                v-if="summary.claude_7d.utilization !== null"
                :class="['h-full transition-all duration-500', getBarColorClass(summary.claude_7d.utilization, 'amber')]"
                :style="{ width: `${clampPercent(summary.claude_7d.utilization)}%` }"
              ></div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { AntigravityUsageSummary } from '@/types'
import Icon from '@/components/icons/Icon.vue'

const props = defineProps<{
  summary?: AntigravityUsageSummary | null
  loading?: boolean
  error?: string | null
}>()

const emit = defineEmits<{
  (e: 'refresh'): void
}>()

const { t } = useI18n()

const formattedUpdatedAt = computed(() => {
  if (!props.summary?.updated_at) return ''
  try {
    const d = new Date(props.summary.updated_at)
    return d.toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit', second: '2-digit' })
  } catch {
    return props.summary.updated_at
  }
})

function clampPercent(val: number | null): number {
  if (val === null || val === undefined) return 0
  return Math.min(Math.max(val, 0), 100)
}

function formatPercent(val: number | null): string {
  if (val === null || val === undefined) return t('admin.accounts.summary.unavailable')
  const rounded = Math.round(val * 10) / 10
  return `${rounded}%`
}

function getBarColorClass(val: number | null, defaultTheme: 'emerald' | 'amber'): string {
  if (val === null || val === undefined) return 'bg-gray-300 dark:bg-gray-600'
  if (val >= 100) return 'bg-red-500'
  if (val >= 80) return 'bg-amber-500'
  if (defaultTheme === 'emerald') return 'bg-emerald-500'
  return 'bg-amber-500'
}

function getPercentTextClass(val: number | null): string {
  if (val === null || val === undefined) return 'text-gray-400 dark:text-gray-500 font-normal'
  if (val >= 100) return 'text-red-600 dark:text-red-400'
  if (val >= 80) return 'text-amber-600 dark:text-amber-400'
  return 'text-gray-700 dark:text-gray-300'
}
</script>
