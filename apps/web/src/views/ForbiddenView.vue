<script setup lang="ts">
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
</script>

<template>
  <el-result icon="warning" title="403" :sub-title="t('errors.FORBIDDEN')">
    <template #extra>
      <router-link to="/dashboard">
        <el-button type="primary">{{ t('nav.dashboard') }}</el-button>
      </router-link>
    </template>
  </el-result>
</template>
