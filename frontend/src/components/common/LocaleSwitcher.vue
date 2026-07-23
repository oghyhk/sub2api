<template>
  <button
    type="button"
    :disabled="switching"
    :aria-label="nextLocale.ariaLabel"
    :title="nextLocale.ariaLabel"
    class="flex min-h-11 cursor-pointer items-center gap-1.5 rounded-lg px-2.5 text-sm font-medium text-gray-600 transition-colors hover:bg-gray-100 focus:outline-none focus:ring-2 focus:ring-primary-500/40 disabled:cursor-wait disabled:opacity-60 dark:text-gray-300 dark:hover:bg-dark-700"
    @click="toggleLocale"
  >
    <Icon name="globe" size="sm" />
    <span>{{ nextLocale.label }}</span>
  </button>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import { setLocale } from '@/i18n'

const { locale } = useI18n()
const switching = ref(false)

const nextLocale = computed(() => {
  if (locale.value === 'en') {
    return { code: 'zh', label: '中文', ariaLabel: 'Switch to Chinese' }
  }

  return { code: 'en', label: 'EN', ariaLabel: 'Switch to English' }
})

async function toggleLocale() {
  if (switching.value) return

  switching.value = true
  try {
    await setLocale(nextLocale.value.code)
  } finally {
    switching.value = false
  }
}
</script>
