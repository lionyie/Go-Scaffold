<script setup lang="ts">
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
</script>

<template>
  <el-result icon="info" title="404" :sub-title="t('common.empty')">
    <template #extra>
      <router-link to="/dashboard">
        <el-button type="primary">{{ t('nav.dashboard') }}</el-button>
      </router-link>
    </template>
  </el-result>
</template>
