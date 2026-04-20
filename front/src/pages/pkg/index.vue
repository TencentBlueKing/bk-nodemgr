<template>
  <NoPermission
    v-if="!hasCurrentAuth"
    :auth-items="currentAuthItems"
    :resource-id="currentResourceId"
  />
  <RouterView v-else />
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { useRoute } from 'vue-router';

import NoPermission from '@/components/no-permission.vue';
import type { PageAuthItem } from '@/constants/auth';
import { MENU_ROUTE_AUTH_MAP } from '@/constants/auth';
import { useAuthStore } from '@/stores/auth';

const route = useRoute();
const authStore = useAuthStore();

// 子路由名 → release_type（package 资源 ID）映射
const ROUTE_RELEASE_TYPE_MAP: Record<string, string> = {
  agentPackageMng: 'agent',
  proxyPackageMng: 'proxy',
  certPackageMng: 'cert',
  bintoolPackageMng: 'bintool',
  plugin_bintoolPackageMng: 'plugin_bintool',
  pluginPackageMng: 'plugin',
};

const currentRouteName = computed(() => (typeof route.name === 'string' ? route.name : ''));

// 当前子路由对应的 action；未配置则放行
const currentAction = computed<string | undefined>(() => MENU_ROUTE_AUTH_MAP[currentRouteName.value]);

// 当前子路由对应的资源 ID（release_type）；无映射（如 operationRecords）则不传资源粒度
const currentResourceId = computed<string | undefined>(() => ROUTE_RELEASE_TYPE_MAP[currentRouteName.value]);

const currentAuthItems = computed<PageAuthItem[]>(() => {
  const action = currentAction.value;
  if (!action) return [];
  return [{ id: action, action, resourceType: 'package', routes: [] }];
});

const hasCurrentAuth = computed(() => {
  const action = currentAction.value;
  if (!action) return true; // 无需鉴权的子路由直接放行
  // authorized 尚未加载 → 先放行，避免闪烁
  if (!authStore.authorizedMap[action]) return true;
  return authStore.hasAuthorizedResource(action, currentResourceId.value);
});
</script>
