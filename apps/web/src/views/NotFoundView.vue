<script setup lang="ts">
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
</script>

<template>
  <div class="status-shell">
    <div class="blob blob--one" />
    <div class="status-card glass">
      <div class="status-code">404</div>
      <h1>{{ t('common.empty') }}</h1>
      <v-btn color="primary" to="/" prepend-icon="mdi-arrow-left">{{ t('nav.home') }}</v-btn>
    </div>
  </div>
</template>

<style scoped>
.status-shell {
  position: relative;
  min-height: 70vh;
  display: grid;
  place-items: center;
  overflow: hidden;
}
.status-card {
  position: relative;
  z-index: 2;
  max-width: 460px;
  padding: 48px 40px;
  border-radius: 28px;
  text-align: center;
}
.status-code {
  font-size: clamp(4rem, 10vw, 7rem);
  font-weight: 900;
  line-height: 1;
  letter-spacing: -0.06em;
  background: linear-gradient(120deg, rgb(var(--v-theme-primary)), rgb(var(--v-theme-secondary)));
  -webkit-background-clip: text;
  background-clip: text;
  color: transparent;
}
h1 {
  margin: 12px 0 24px;
  font-size: 1.3rem;
}
.blob--one {
  top: -100px;
  right: 25%;
  width: 420px;
  height: 420px;
  background: rgba(99, 102, 241, 0.22);
}
</style>
