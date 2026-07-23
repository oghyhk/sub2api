<template>
  <div class="flex min-h-screen items-center justify-center bg-[#f5f5f7] p-4 dark:bg-[#000000]">
    <div class="w-full max-w-md">
      <!-- Logo/Brand -->
      <div class="mb-10 text-center">
        <template v-if="settingsLoaded">
          <div
            class="mb-4 inline-flex h-14 w-14 items-center justify-center overflow-hidden rounded-xl"
          >
            <img :src="siteLogo || '/logo.svg'" alt="Logo" class="h-full w-full object-contain" />
          </div>
          <h1 class="mb-1 text-2xl font-semibold tracking-tight text-[#1d1d1f] dark:text-[#f5f5f7]">
            {{ siteName }}
          </h1>
          <p class="text-sm text-[#6e6e73] dark:text-[#98989d]">
            {{ siteSubtitle }}
          </p>
        </template>
      </div>

      <!-- Card Container -->
      <div class="rounded-[8px] border border-[#d2d2d7] bg-white p-8 shadow-sm dark:border-[#2c2c2e] dark:bg-[#1c1c1e]">
        <slot />
      </div>

      <!-- Footer Links -->
      <div class="mt-6 text-center text-sm text-[#6e6e73] dark:text-[#98989d]">
        <slot name="footer" />
      </div>

      <!-- Copyright -->
      <div class="mt-8 text-center text-xs text-[#6e6e73] dark:text-[#98989d]">
        &copy; {{ currentYear }} {{ siteName }}. All rights reserved.
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useAppStore } from '@/stores'
import { sanitizeUrl } from '@/utils/url'

const appStore = useAppStore()

const siteName = computed(() => appStore.siteName || 'Sub2API')
const siteLogo = computed(() => sanitizeUrl(appStore.siteLogo || '', { allowRelative: true, allowDataUrl: true }))
const siteSubtitle = computed(() => appStore.cachedPublicSettings?.site_subtitle || 'Subscription to API Conversion Platform')
const settingsLoaded = computed(() => appStore.publicSettingsLoaded)

const currentYear = computed(() => new Date().getFullYear())

onMounted(() => {
  appStore.fetchPublicSettings()
})
</script>

<style scoped>
@media (prefers-reduced-motion: reduce) {
  *,
  *::before,
  *::after {
    transition-duration: 1ms !important;
    animation-duration: 1ms !important;
  }
}
</style>
