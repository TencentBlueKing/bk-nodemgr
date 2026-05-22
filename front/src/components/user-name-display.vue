<!--
 * @Author: hyfyahuang
 * @Date: 2026-05-18
 * @Description: 统一的用户名展示组件
 * 根据租户模式自动选择展示方式：
 * - 无租户模式：直接显示用户名文本
 * - 单租户/多租户模式：使用 bk-user-display-name 组件
 -->
<template>
  <span v-if="!hasApigwBaseUrl || !tenant">{{ name || '--' }}</span>
  <bk-user-display-name v-else :user-id="name"></bk-user-display-name>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import BkUserDisplayName from '@blueking/bk-user-display-name';

const props = defineProps<{
  name: string;
}>();

const hasApigwBaseUrl = computed(() => window.PROJECT_CONFIG.BK_USER_WEB_URL);
const tenant = computed(() => window.PROJECT_CONFIG.BK_TENANT);
</script>
