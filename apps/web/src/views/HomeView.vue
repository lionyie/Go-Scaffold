<script setup lang="ts">
import { useI18n } from 'vue-i18n'

import { PageHeader } from '@scaffold/ui'

const { t } = useI18n()
</script>

<template>
  <div class="page-card">
    <PageHeader :title="t('home.title')" :description="t('home.subtitle')" />
    <el-row :gutter="16">
      <el-col :span="8">
        <el-card shadow="never">
          <h3>{{ t('home.apps') }}</h3>
          <p>{{ t('home.appsHint') }}</p>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="never">
          <h3>{{ t('home.extensions') }}</h3>
          <p>{{ t('home.extensionsHint') }}</p>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="never">
          <h3>{{ t('home.admin') }}</h3>
          <el-link href="/admin/" type="primary">{{ t('home.adminLink') }}</el-link>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>
