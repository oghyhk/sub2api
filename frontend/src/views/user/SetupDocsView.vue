<template>
  <AppLayout>
    <div class="mx-auto max-w-5xl space-y-6">
      <section class="rounded-xl border border-primary-200 bg-primary-50 p-5 dark:border-primary-800/60 dark:bg-primary-950/30">
        <div class="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
          <div class="min-w-0">
            <div class="flex items-center gap-2 text-primary-700 dark:text-primary-300">
              <Icon name="book" size="md" />
              <h2 class="text-base font-semibold">{{ t('setupDocs.agentTitle') }}</h2>
            </div>
            <p class="mt-2 max-w-3xl text-sm leading-6 text-primary-900/80 dark:text-primary-100/80">
              {{ t('setupDocs.agentDescription') }}
            </p>
            <code class="mt-3 block break-all rounded-lg bg-white/80 px-3 py-2 text-xs text-gray-800 ring-1 ring-primary-200 dark:bg-black/30 dark:text-gray-100 dark:ring-primary-800">
              {{ agentGuideUrl }}
            </code>
          </div>
          <button
            type="button"
            class="btn btn-primary flex-shrink-0"
            data-testid="copy-agent-guide-link"
            @click="copyAgentLink"
          >
            <Icon :name="copied ? 'check' : 'copy'" size="sm" class="mr-2" />
            {{ copied ? t('setupDocs.copied') : t('setupDocs.copyLink') }}
          </button>
        </div>
      </section>

      <section
        v-if="loading"
        class="flex min-h-72 items-center justify-center rounded-xl border border-gray-200 bg-white dark:border-dark-700 dark:bg-dark-900"
      >
        <span class="h-8 w-8 animate-spin rounded-full border-2 border-gray-200 border-b-primary-600 dark:border-dark-700 dark:border-b-primary-400"></span>
      </section>

      <section
        v-else-if="loadError"
        class="rounded-xl border border-red-200 bg-red-50 p-6 text-red-700 dark:border-red-500/30 dark:bg-red-500/10 dark:text-red-200"
      >
        <h2 class="font-semibold">{{ t('setupDocs.loadFailed') }}</h2>
        <p class="mt-2 text-sm">{{ t('setupDocs.loadFailedDescription') }}</p>
        <button type="button" class="btn btn-secondary mt-4" @click="loadGuide">
          {{ t('common.retry') }}
        </button>
      </section>

      <article
        v-else
        class="setup-doc-content rounded-xl border border-gray-200 bg-white p-5 shadow-sm dark:border-dark-700 dark:bg-dark-900 sm:p-8"
        data-testid="setup-doc-content"
        v-html="renderedGuide"
      ></article>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { marked } from 'marked'
import DOMPurify from 'dompurify'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import { useClipboard } from '@/composables/useClipboard'

const { t } = useI18n()
const { copied, copyToClipboard } = useClipboard()
const guideMarkdown = ref('')
const loading = ref(true)
const loadError = ref(false)

marked.setOptions({ breaks: true, gfm: true })

const agentGuideUrl = computed(() => `${window.location.origin}/docs/setup.md`)
const renderedGuide = computed(() =>
  DOMPurify.sanitize(marked.parse(guideMarkdown.value) as string),
)

async function copyAgentLink() {
  await copyToClipboard(agentGuideUrl.value, t('setupDocs.linkCopied'))
}

async function loadGuide() {
  loading.value = true
  loadError.value = false
  try {
    const response = await fetch('/docs/setup.md', {
      headers: { Accept: 'text/markdown' },
    })
    if (!response.ok) {
      throw new Error(`Setup guide returned ${response.status}`)
    }
    guideMarkdown.value = await response.text()
  } catch (error) {
    console.error('Failed to load setup guide:', error)
    loadError.value = true
  } finally {
    loading.value = false
  }
}

onMounted(loadGuide)
</script>

<style scoped>
.setup-doc-content {
  line-height: 1.7;
  overflow-wrap: anywhere;
}

.setup-doc-content :deep(h1) {
  @apply mb-4 text-2xl font-bold text-gray-950 dark:text-white sm:text-3xl;
}

.setup-doc-content :deep(h2) {
  @apply mb-3 mt-8 border-b border-gray-200 pb-2 text-xl font-semibold text-gray-950 dark:border-dark-700 dark:text-white;
}

.setup-doc-content :deep(h3) {
  @apply mb-2 mt-6 text-base font-semibold text-gray-900 dark:text-white;
}

.setup-doc-content :deep(p) {
  @apply mb-4 text-sm text-gray-700 dark:text-gray-300;
}

.setup-doc-content :deep(ul),
.setup-doc-content :deep(ol) {
  @apply mb-4 space-y-1 pl-6 text-sm text-gray-700 dark:text-gray-300;
}

.setup-doc-content :deep(ul) {
  @apply list-disc;
}

.setup-doc-content :deep(ol) {
  @apply list-decimal;
}

.setup-doc-content :deep(blockquote) {
  @apply my-5 border-l-4 border-primary-400 bg-primary-50 px-4 py-3 text-sm text-primary-900 dark:border-primary-600 dark:bg-primary-950/30 dark:text-primary-100;
}

.setup-doc-content :deep(code) {
  @apply rounded bg-gray-100 px-1.5 py-0.5 font-mono text-xs text-gray-900 dark:bg-dark-800 dark:text-gray-100;
}

.setup-doc-content :deep(pre) {
  @apply my-4 overflow-x-auto rounded-lg bg-gray-950 p-4 text-sm text-gray-100;
}

.setup-doc-content :deep(pre code) {
  @apply bg-transparent p-0 text-inherit;
}

.setup-doc-content :deep(table) {
  @apply my-4 block w-full overflow-x-auto border-collapse text-sm;
}

.setup-doc-content :deep(th) {
  @apply whitespace-nowrap border border-gray-200 bg-gray-50 px-3 py-2 text-left font-semibold dark:border-dark-700 dark:bg-dark-800;
}

.setup-doc-content :deep(td) {
  @apply border border-gray-200 px-3 py-2 dark:border-dark-700;
}
</style>
