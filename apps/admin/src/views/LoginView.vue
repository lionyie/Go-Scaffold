<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'

import { notifyApi } from '@scaffold/ui'

import { errorMessage } from '../api/errors'
import { useAuthStore } from '../stores/auth'

interface ProviderInfo {
  name: string
  displayName: string
}

const auth = useAuthStore()
const router = useRouter()
const route = useRoute()
const { t } = useI18n()

const form = reactive({ username: 'admin', password: 'admin123' })
const loading = ref(false)
const showPassword = ref(false)
const providers = ref<ProviderInfo[]>([])

async function submit(): Promise<void> {
  loading.value = true
  try {
    await auth.login(form.username, form.password)
    if (auth.user?.mustChangePassword) {
      await router.replace('/profile')
      return
    }
    const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : '/dashboard'
    await router.replace(redirect)
  } catch (error) {
    notifyApi.error(errorMessage(error, (key) => t(key)))
  } finally {
    loading.value = false
  }
}

async function startOAuth(provider: string): Promise<void> {
  try {
    const response = await fetch(`/api/v1/auth/oauth/${encodeURIComponent(provider)}/start`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: '{}',
    })
    if (!response.ok) throw new Error(await response.text())
    const payload = (await response.json()) as { authUrl?: string }
    if (!payload.authUrl) throw new Error('missing authUrl')
    window.location.href = payload.authUrl
  } catch (error) {
    notifyApi.error(errorMessage(error, (key) => t(key)))
  }
}

onMounted(async () => {
  try {
    const response = await fetch('/api/v1/auth/oauth/providers')
    if (!response.ok) return
    const payload = (await response.json()) as { providers?: ProviderInfo[] }
    providers.value = payload.providers ?? []
  } catch {
    providers.value = []
  }
})
</script>

<template>
  <div class="login-shell">
    <div class="blob blob--one" />
    <div class="blob blob--two" />
    <div class="blob blob--three" />
    <v-container class="login-container">
      <v-row align="center" justify="center" class="fill-height">
        <v-col cols="12" md="6" class="d-none d-md-flex">
          <div class="hero-copy">
            <v-chip color="primary" variant="tonal" class="mb-6 hero-chip">
              <v-icon start icon="mdi-hexagon-multiple" />
              {{ t('login.heroChip') }}
            </v-chip>
            <h1 class="hero-title">
              {{ t('login.heroTitleA') }}
              <span class="gradient-text">{{ t('login.heroTitleB') }}</span>{{ t('login.heroTitleC') }}
            </h1>
            <p class="hero-subtitle">{{ t('login.heroSubtitle') }}</p>
            <div class="hero-pills">
              <v-chip color="primary" variant="tonal" prepend-icon="mdi-shield-check">
                {{ t('login.heroRbac') }}
              </v-chip>
              <v-chip color="secondary" variant="tonal" prepend-icon="mdi-puzzle-outline">
                {{ t('login.heroExtensions') }}
              </v-chip>
              <v-chip color="accent" variant="tonal" prepend-icon="mdi-earth">
                {{ t('login.heroOauth') }}
              </v-chip>
            </div>
          </div>
        </v-col>

        <v-col cols="12" md="5" lg="4">
          <v-card class="login-card glass">
            <div class="login-card__glow" />
            <v-card-title class="text-h5 font-weight-bold px-8 pt-7">
              {{ t('login.title') }}
            </v-card-title>
            <v-card-text class="px-8 pb-0 pt-1 text-medium-emphasis">
              {{ t('login.subtitle') }}
            </v-card-text>

            <v-form class="px-8 pt-5" @submit.prevent="submit">
              <v-text-field
                v-model="form.username"
                :label="t('login.username')"
                prepend-inner-icon="mdi-account-outline"
                autocomplete="username"
                class="mb-3"
              />
              <v-text-field
                v-model="form.password"
                :label="t('login.password')"
                :type="showPassword ? 'text' : 'password'"
                prepend-inner-icon="mdi-lock-outline"
                :append-inner-icon="showPassword ? 'mdi-eye-off-outline' : 'mdi-eye-outline'"
                autocomplete="current-password"
                class="mb-4"
                @click:append-inner="showPassword = !showPassword"
              />
              <v-btn type="submit" color="primary" size="large" block :loading="loading">
                {{ t('login.submit') }}
              </v-btn>
            </v-form>

            <div v-if="providers.length > 0" class="px-8 pt-5 pb-2">
              <v-divider class="mb-4">
                <span class="text-caption text-medium-emphasis">{{ t('login.oauth') }}</span>
              </v-divider>
              <v-btn
                v-for="provider in providers"
                :key="provider.name"
                variant="outlined"
                block
                class="mb-2 oauth-btn"
                @click="startOAuth(provider.name)"
              >
                <v-icon start icon="mdi-login-variant" />
                {{ provider.displayName }}
              </v-btn>
            </div>
            <div class="pb-7" />
          </v-card>
        </v-col>
      </v-row>
    </v-container>
  </div>
</template>

<style scoped>
.login-shell {
  position: relative;
  min-height: 100vh;
  display: grid;
  place-items: center;
  overflow: hidden;
  background:
    radial-gradient(900px 520px at 80% -10%, rgba(var(--v-theme-secondary), 0.22), transparent 60%),
    radial-gradient(760px 420px at 0% 100%, rgba(var(--v-theme-primary), 0.22), transparent 60%),
    rgb(var(--v-theme-background));
}
.login-container {
  position: relative;
  z-index: 2;
  width: 100%;
}
.hero-copy {
  max-width: 560px;
}
.hero-chip {
  backdrop-filter: blur(10px);
}
.hero-title {
  margin: 0 0 18px;
  font-size: clamp(2rem, 3.8vw, 3.2rem);
  line-height: 1.08;
  font-weight: 850;
  letter-spacing: -0.035em;
  white-space: pre-line;
}
.hero-subtitle {
  margin: 0 0 26px;
  max-width: 480px;
  color: rgb(var(--v-theme-on-surface-variant));
  font-size: 1.02rem;
  line-height: 1.7;
}
.hero-pills {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}
.login-card {
  position: relative;
  overflow: hidden;
  border-radius: 28px;
  box-shadow: 0 40px 90px -40px rgba(15, 23, 42, 0.55);
  animation: card-in 0.6s cubic-bezier(0.22, 1, 0.36, 1) both;
}
.login-card__glow {
  position: absolute;
  inset: -40% -20% auto;
  height: 240px;
  background: radial-gradient(circle, rgba(var(--v-theme-primary), 0.28), transparent 68%);
  pointer-events: none;
}
.oauth-btn {
  text-transform: none;
}
.blob--one {
  top: -120px;
  right: -80px;
  width: 420px;
  height: 420px;
  background: rgba(99, 102, 241, 0.45);
}
.blob--two {
  bottom: -160px;
  left: -120px;
  width: 480px;
  height: 480px;
  background: rgba(6, 182, 212, 0.38);
  animation-delay: -4s;
}
.blob--three {
  top: 40%;
  left: 45%;
  width: 260px;
  height: 260px;
  background: rgba(244, 114, 182, 0.26);
  animation-delay: -8s;
}
</style>
