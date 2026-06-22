<!--
 * @Author: hyfyahuang
 * @Date: 2026-05-18
 * @Description: 统一的用户名展示组件
 * 根据租户模式自动选择展示方式：
 * - 无租户模式：直接显示用户名文本
 * - 单租户/多租户模式：使用 bk-user-display-name 组件
 * 注意：若传入的 name 为空或类似模板语法，直接展示原始值而非通过 Web Component 转换
-->
<template>
  <span v-if="isPlainText">{{ name || '--' }}</span>
  <bk-user-display-name v-else ref="customEl" :user-id="name"></bk-user-display-name>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue';

// 确保 bk-user-display-name 自定义元素已注册
import BkUserDisplayName from '@blueking/bk-user-display-name';
// 避免 TypeScript 未使用变量警告
void BkUserDisplayName;

const props = defineProps<{
  name: string;
}>();

const customEl = ref<HTMLElement | null>(null);

const hasApigwBaseUrl = computed(() => window.PROJECT_CONFIG.BK_USER_WEB_URL);
const tenant = computed(() => window.PROJECT_CONFIG.BK_TENANT);

// 当无租户环境、name 为空、或 name 是未渲染的模板字符串时，直接显示原文
const isPlainText = computed(() => {
  if (!hasApigwBaseUrl.value || !tenant.value) return true;
  const val = props.name;
  if (!val || !val.trim()) return true;
  // 包含 {{ }} 模板语法的字符串（后端未渲染），直接展示原文
  if (/\{\{.*?\}\}/.test(val)) return true;
  return false;
});

// 手动同步 user-id 属性，确保 Web Component 的 attributeChangedCallback 触发
watch(() => props.name, (newName) => {
  if (customEl.value) {
    const el = customEl.value as HTMLElement;
    const nextVal = newName || '';
    if (el.getAttribute('user-id') !== nextVal) {
      el.setAttribute('user-id', nextVal);
    }
  }
}, { immediate: true });
</script>
