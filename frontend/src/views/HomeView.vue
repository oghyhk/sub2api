<template>
  <!-- Custom Home Content: Full Page Mode -->
  <div v-if="homeContent" class="min-h-screen">
    <!-- iframe mode -->
    <iframe
      v-if="isHomeContentUrl"
      :src="homeContent.trim()"
      class="h-screen w-full border-0"
      allowfullscreen
    ></iframe>
    <!-- HTML mode - SECURITY: homeContent is admin-only setting, XSS risk is acceptable -->
    <div v-else v-html="homeContent"></div>
  </div>

  <!-- Default Home Page -->
  <div
    v-else
    class="min-h-screen bg-[#f5f5f7] dark:bg-[#000000]"
  >
    <!-- Header -->
    <header class="px-6 py-4">
      <nav class="mx-auto flex max-w-[980px] items-center justify-between">
        <!-- Logo -->
        <div class="flex min-w-0 items-center gap-2.5">
          <div class="h-10 w-10 overflow-hidden rounded-xl shadow-md">
            <img :src="siteLogo || '/logo.svg'" :alt="siteName" class="h-full w-full object-contain" />
          </div>
          <span class="hidden truncate text-sm font-semibold text-[#1d1d1f] dark:text-[#f5f5f7] sm:block">{{ siteName }}</span>
        </div>

        <!-- Nav Actions -->
        <div class="flex items-center gap-3">
          <!-- Language Switcher -->
          <LocaleSwitcher />

          <!-- Doc Link -->
          <a
            v-if="docUrl"
            :href="docUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="flex h-11 w-11 items-center justify-center rounded-lg text-gray-500 transition-colors hover:bg-gray-100 hover:text-gray-700 focus:outline-none focus:ring-2 focus:ring-[#0071e3]/40 dark:text-dark-400 dark:hover:bg-dark-800 dark:hover:text-white"
            :title="t('home.viewDocs')"
          >
            <Icon name="book" size="md" />
          </a>

          <!-- Theme Toggle -->
          <button
            @click="toggleTheme"
            class="flex h-11 w-11 items-center justify-center rounded-lg text-gray-500 transition-colors hover:bg-gray-100 hover:text-gray-700 focus:outline-none focus:ring-2 focus:ring-[#0071e3]/40 dark:text-dark-400 dark:hover:bg-dark-800 dark:hover:text-white"
            :aria-label="isDark ? t('home.switchToLight') : t('home.switchToDark')"
            :title="isDark ? t('home.switchToLight') : t('home.switchToDark')"
          >
            <Icon v-if="isDark" name="sun" size="md" />
            <Icon v-else name="moon" size="md" />
          </button>

          <!-- Login / Dashboard Button -->
          <router-link
            v-if="isAuthenticated"
            :to="dashboardPath"
            class="inline-flex min-h-11 items-center gap-2 rounded-lg bg-[#1d1d1f] px-3 transition-colors hover:bg-black focus:outline-none focus:ring-2 focus:ring-[#0071e3]/40 dark:bg-[#f5f5f7] dark:text-[#1d1d1f]"
          >
            <span
              class="flex h-6 w-6 items-center justify-center rounded-full bg-[#0071e3] text-[10px] font-semibold text-white"
            >
              {{ userInitial }}
            </span>
            <span class="text-xs font-medium text-white dark:text-[#1d1d1f]">{{ t('home.dashboard') }}</span>
            <Icon name="externalLink" size="xs" class="text-gray-400 dark:text-[#6e6e73]" />
          </router-link>
          <router-link
            v-else
            to="/login"
            class="inline-flex min-h-11 items-center rounded-lg bg-[#1d1d1f] px-4 text-xs font-medium text-white transition-colors hover:bg-black focus:outline-none focus:ring-2 focus:ring-[#0071e3]/40 dark:bg-[#f5f5f7] dark:text-[#1d1d1f]"
          >
            {{ t('home.login') }}
          </router-link>
        </div>
      </nav>
    </header>

    <!-- Main Content -->
    <main class="flex-1 px-6">
      <div class="mx-auto max-w-[980px]">
        <!-- Hero Section -->
        <section class="pb-12 pt-16 text-center md:pb-16 md:pt-24 lg:pb-20 lg:pt-32">
          <h1 class="mb-4 text-[40px] font-semibold leading-[1.05] text-[#1d1d1f] dark:text-[#f5f5f7] md:text-[56px] lg:text-[64px]">
            {{ t('productOnboarding.hero.title', { count: PRODUCT_MODELS.length }) }}
          </h1>
          <p class="mx-auto mb-8 max-w-[680px] text-lg leading-relaxed text-[#6e6e73] dark:text-[#98989d] md:text-xl">
            {{ t('productOnboarding.hero.subtitle') }}
          </p>
          <router-link
            :to="primaryActionPath"
            class="inline-flex min-h-[44px] items-center justify-center rounded-[8px] bg-[#0071e3] px-6 text-sm font-medium text-white transition-all duration-200 hover:bg-[#0077ed] focus:outline-none focus:ring-2 focus:ring-[#0071e3]/40"
          >
            {{ isAuthenticated ? t('home.goToDashboard') : t('productOnboarding.cta.button') }}
          </router-link>
        </section>

        <!-- Model Selector -->
        <section class="border-t border-[#d2d2d7] py-12 dark:border-[#2c2c2e] md:py-16">
          <div class="mb-8 text-center">
            <h2 class="mb-2 text-[28px] font-semibold text-[#1d1d1f] dark:text-[#f5f5f7] md:text-[36px]">
              {{ t('productOnboarding.models.title') }}
            </h2>
            <p class="text-sm text-[#6e6e73] dark:text-[#98989d] md:text-base">
              {{ t('productOnboarding.models.subtitle') }}
            </p>
          </div>
          <div class="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-5">
            <button
              v-for="m in PRODUCT_MODELS"
              :key="m.id"
              @click="selectedModelId = m.id"
              :aria-pressed="selectedModelId === m.id"
              class="flex min-h-[44px] cursor-pointer flex-col items-center justify-center rounded-[8px] border p-5 text-center transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-[#0071e3]/40"
              :class="selectedModelId === m.id
                ? `${m.accent.selectedBorderClass} ${m.accent.selectedBgClass}`
                : 'border-[#d2d2d7] bg-white hover:border-[#a1a1a6] dark:border-[#2c2c2e] dark:bg-[#1c1c1e] dark:hover:border-[#3a3a3c]'"
            >
              <span
                class="mb-1.5 text-xs font-semibold uppercase tracking-wider"
                :class="m.accent.badgeTextClass"
              >{{ m.accent.providerLabel }}</span>
              <span class="mb-1 text-sm font-medium text-[#1d1d1f] dark:text-[#f5f5f7]">{{ m.displayName }}</span>
            </button>
          </div>

          <!-- Selected Model Detail Exposure -->
          <div
            v-if="selectedModelDetails"
            class="mt-6 rounded-[8px] border border-[#d2d2d7] bg-white p-5 dark:border-[#2c2c2e] dark:bg-[#1c1c1e]"
          >
            <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
              <div class="flex items-center gap-2.5">
                <span class="h-2.5 w-2.5 rounded-full" :class="selectedModelDetails.accent.dotClass"></span>
                <span class="text-sm font-semibold text-[#1d1d1f] dark:text-[#f5f5f7]">{{ selectedModelDetails.displayName }}</span>
                <span class="rounded bg-[#f5f5f7] px-2 py-0.5 font-mono text-xs text-[#6e6e73] dark:bg-[#2c2c2e] dark:text-[#98989d]">
                  {{ selectedModelDetails.protocol }}
                </span>
              </div>
              <div class="font-mono text-xs text-[#6e6e73] dark:text-[#98989d]">
                <span class="rounded bg-[#f5f5f7] px-2 py-1 select-all dark:bg-[#2c2c2e]">{{ selectedModelDetails.endpointPath }}</span>
              </div>
            </div>
            <p class="mt-3 text-xs leading-relaxed text-[#6e6e73] dark:text-[#98989d]">
              {{ getModelPurpose(selectedModelDetails, locale) }}
            </p>
          </div>
        </section>

        <!-- Three Setup Steps -->
        <section class="border-t border-[#d2d2d7] py-12 dark:border-[#2c2c2e] md:py-16">
          <h2 class="mb-10 text-center text-[28px] font-semibold text-[#1d1d1f] dark:text-[#f5f5f7] md:text-[36px]">
            {{ t('productOnboarding.steps.title') }}
          </h2>
          <div class="grid grid-cols-1 gap-8 md:grid-cols-3 md:gap-6">
            <div v-for="(step, index) in stepsData" :key="index" class="text-center">
              <div class="mx-auto mb-4 flex h-10 w-10 items-center justify-center rounded-full bg-[#f5f5f7] dark:bg-[#2c2c2e]">
                <span class="text-sm font-semibold text-[#6e6e73] dark:text-[#98989d]">{{ index + 1 }}</span>
              </div>
              <h3 class="mb-2 text-base font-semibold text-[#1d1d1f] dark:text-[#f5f5f7]">{{ t(step.titleKey) }}</h3>
              <p class="mx-auto max-w-[280px] text-sm leading-relaxed text-[#6e6e73] dark:text-[#98989d]">{{ t(step.descKey) }}</p>
            </div>
          </div>
        </section>

        <!-- CTA Section -->
        <section class="border-t border-[#d2d2d7] py-12 text-center dark:border-[#2c2c2e] md:py-16 lg:py-20">
          <h2 class="mb-6 text-[28px] font-semibold text-[#1d1d1f] dark:text-[#f5f5f7] md:text-[36px]">
            {{ t('productOnboarding.cta.title') }}
          </h2>
          <router-link
            :to="primaryActionPath"
            class="inline-flex min-h-[44px] items-center justify-center rounded-[8px] bg-[#0071e3] px-6 text-sm font-medium text-white transition-all duration-200 hover:bg-[#0077ed] focus:outline-none focus:ring-2 focus:ring-[#0071e3]/40"
          >
            {{ isAuthenticated ? t('home.goToDashboard') : t('productOnboarding.cta.button') }}
          </router-link>
        </section>
      </div>
    </main>

    <!-- Footer -->
    <footer class="border-t border-[#d2d2d7] px-6 py-8 dark:border-[#2c2c2e]">
      <div class="mx-auto flex max-w-[980px] flex-col items-center justify-center gap-4 text-center sm:flex-row sm:text-left">
        <p class="text-sm text-[#6e6e73] dark:text-[#98989d]">
          &copy; {{ currentYear }} {{ siteName }}. {{ t('home.footer.allRightsReserved') }}
        </p>
        <div class="flex items-center gap-4">
          <a
            v-if="docUrl"
            :href="docUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="text-sm text-[#6e6e73] transition-colors hover:text-[#1d1d1f] dark:text-[#98989d] dark:hover:text-[#f5f5f7]"
          >
            {{ t('home.docs') }}
          </a>
          <a
            :href="githubUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="text-sm text-[#6e6e73] transition-colors hover:text-[#1d1d1f] dark:text-[#98989d] dark:hover:text-[#f5f5f7]"
          >
            GitHub
          </a>
        </div>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore, useAppStore } from '@/stores'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import Icon from '@/components/icons/Icon.vue'
import { sanitizeUrl } from '@/utils/url'
import { PRODUCT_MODELS, getModelById, getModelPurpose, DEFAULT_MODEL_ID } from '@/constants/models'

const { t, locale } = useI18n()

const authStore = useAuthStore()
const appStore = useAppStore()

// Site settings - directly from appStore (already initialized from injected config)
const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'Sub2API')
const siteLogo = computed(() => sanitizeUrl(appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '', { allowRelative: true, allowDataUrl: true }))
const docUrl = computed(() => sanitizeUrl(appStore.cachedPublicSettings?.doc_url || appStore.docUrl || ''))
const homeContent = computed(() => appStore.cachedPublicSettings?.home_content || '')

// Check if homeContent is a URL (for iframe display)
const isHomeContentUrl = computed(() => {
  const content = homeContent.value.trim()
  return content.startsWith('http://') || content.startsWith('https://')
})

// Theme
const isDark = ref(document.documentElement.classList.contains('dark'))

// GitHub URL
const githubUrl = 'https://github.com/oghyhk/sub2api'

// Auth state
const isAuthenticated = computed(() => authStore.isAuthenticated)
const isAdmin = computed(() => authStore.isAdmin)
const dashboardPath = computed(() => isAdmin.value ? '/admin/dashboard' : '/dashboard')
const primaryActionPath = computed(() => {
  if (isAuthenticated.value) return dashboardPath.value
  return appStore.cachedPublicSettings?.registration_enabled === false ? '/login' : '/register'
})
const userInitial = computed(() => {
  const user = authStore.user
  if (!user || !user.email) return ''
  return user.email.charAt(0).toUpperCase()
})

// Current year for footer
const currentYear = computed(() => new Date().getFullYear())

// Onboarding state
const selectedModelId = ref(DEFAULT_MODEL_ID)
const selectedModelDetails = computed(() => getModelById(selectedModelId.value))

const stepsData = [
  { titleKey: 'productOnboarding.steps.step1.title', descKey: 'productOnboarding.steps.step1.desc' },
  { titleKey: 'productOnboarding.steps.step2.title', descKey: 'productOnboarding.steps.step2.desc' },
  { titleKey: 'productOnboarding.steps.step3.title', descKey: 'productOnboarding.steps.step3.desc' }
]

// Toggle theme
function toggleTheme() {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}

// Initialize theme
function initTheme() {
  const savedTheme = localStorage.getItem('theme')
  if (
    savedTheme === 'dark' ||
    (!savedTheme && window.matchMedia('(prefers-color-scheme: dark)').matches)
  ) {
    isDark.value = true
    document.documentElement.classList.add('dark')
  }
}

onMounted(() => {
  initTheme()

  // Check auth state
  authStore.checkAuth()

  // Ensure public settings are loaded (will use cache if already loaded from injected config)
  if (!appStore.publicSettingsLoaded) {
    appStore.fetchPublicSettings()
  }
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
