<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'

const { t } = useI18n()
const router = useRouter()

const features = computed(() => [
  {
    icon: 'mdi-shield-lock-outline',
    title: t('home.featureRbacTitle'),
    text: t('home.featureRbacText'),
    color: 'primary',
  },
  {
    icon: 'mdi-puzzle-outline',
    title: t('home.featureExtTitle'),
    text: t('home.featureExtText'),
    color: 'secondary',
  },
  {
    icon: 'mdi-devices',
    title: t('home.featureMultiTitle'),
    text: t('home.featureMultiText'),
    color: 'accent',
  },
])
</script>

<template>
  <div class="home">
    <div class="blob blob--one" />
    <div class="blob blob--two" />

    <section class="hero">
      <v-chip color="primary" variant="tonal" class="mb-5" prepend-icon="mdi-star-four-points-outline">
        {{ t('app.tagline') }}
      </v-chip>
      <h1 class="hero__title">
        {{ t('home.heroA') }}
        <span class="gradient-text">{{ t('home.heroB') }}</span>{{ t('home.heroC') }}
      </h1>
      <p class="hero__subtitle">{{ t('home.heroSubtitle') }}</p>
      <div class="hero__actions">
        <v-btn color="primary" size="large" rounded="pill" prepend-icon="mdi-rocket-launch-outline" @click="router.push('/plugins')">
          {{ t('nav.plugins') }}
        </v-btn>
        <v-btn size="large" rounded="pill" variant="outlined" href="/admin/" prepend-icon="mdi-shield-account">
          {{ t('home.adminLink') }}
        </v-btn>
      </div>
    </section>

    <v-row class="mt-10">
      <v-col v-for="feature in features" :key="feature.title" cols="12" md="4">
        <v-card class="feature-card hover-lift" height="100%">
          <v-card-text class="pa-7">
            <v-avatar :color="feature.color" variant="tonal" size="52" class="mb-5">
              <v-icon :icon="feature.icon" size="26" />
            </v-avatar>
            <h3>{{ feature.title }}</h3>
            <p>{{ feature.text }}</p>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </div>
</template>

<style scoped>
.home {
  position: relative;
  overflow: hidden;
}
.hero {
  position: relative;
  z-index: 2;
  max-width: 820px;
  padding: 34px 0 10px;
}
.hero__title {
  margin: 0 0 20px;
  font-size: clamp(2.2rem, 4.2vw, 3.6rem);
  line-height: 1.06;
  font-weight: 900;
  letter-spacing: -0.045em;
  white-space: pre-line;
}
.hero__subtitle {
  margin: 0 0 28px;
  max-width: 640px;
  color: rgb(var(--v-theme-on-surface-variant));
  font-size: 1.05rem;
  line-height: 1.75;
}
.hero__actions {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}
.feature-card {
  border-radius: 24px;
}
.feature-card h3 {
  margin: 0 0 8px;
  font-size: 1.05rem;
  font-weight: 750;
}
.feature-card p {
  margin: 0;
  color: rgb(var(--v-theme-on-surface-variant));
  line-height: 1.65;
  font-size: 0.92rem;
}
.blob--one {
  top: -160px;
  right: -120px;
  width: 520px;
  height: 520px;
  background: rgba(99, 102, 241, 0.34);
}
.blob--two {
  top: 180px;
  left: -160px;
  width: 460px;
  height: 460px;
  background: rgba(6, 182, 212, 0.28);
  animation-delay: -5s;
}
</style>
