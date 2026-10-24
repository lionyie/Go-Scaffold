<script setup lang="ts">
defineProps<{
  title: string
  description?: string
}>()
</script>

<template>
  <div class="page-header">
    <div class="page-header__copy">
      <h1 class="page-header__title">{{ title }}</h1>
      <p v-if="description" class="page-header__description">{{ description }}</p>
    </div>
    <div class="page-header__actions">
      <slot name="actions" />
    </div>
  </div>
</template>

<style scoped>
.page-header {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 20px;
  margin-bottom: 24px;
  animation: header-in 0.45s cubic-bezier(0.22, 1, 0.36, 1) both;
}
.page-header__title {
  margin: 0;
  font-size: clamp(1.5rem, 2.4vw, 2rem);
  font-weight: 800;
  letter-spacing: -0.03em;
  background: linear-gradient(120deg, rgb(var(--v-theme-primary)), rgb(var(--v-theme-secondary)) 55%, rgb(var(--v-theme-accent)));
  -webkit-background-clip: text;
  background-clip: text;
  color: transparent;
}
.page-header__description {
  margin: 6px 0 0;
  color: rgb(var(--v-theme-on-surface-variant));
  font-size: 0.9rem;
}
.page-header__actions {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}
@keyframes header-in {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
</style>
