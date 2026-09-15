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

// 操作记录页覆盖全部包类型：申请历史查看权限时逐资源鉴权
const ALL_RELEASE_TYPES = ['agent', 'proxy', 'cert', 'bintool', 'plugin_bintool', 'plugin'];

const currentRouteName = computed(() => (typeof route.name === 'string' ? route.name : ''));

// 当前子路由对应的 action；未配置则放行
const currentAction = computed<string | undefined>(() => MENU_ROUTE_AUTH_MAP[currentRouteName.value]);

// 当前子路由对应的资源 ID（release_type）；
// operationRecords 传全部包类型数组（申请权限时携带全部资源实例）
const currentResourceId = computed<string | string[] | undefined>(() => {
  if (currentRouteName.value === 'operationRecords') return ALL_RELEASE_TYPES;
  return ROUTE_RELEASE_TYPE_MAP[currentRouteName.value];
});

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

  // 插件包管理路由级不限制（资源 id 实际是插件名，不是固定 "plugin"）
  // 由子页 plugin-package-manage.vue 基于 package_manage 资源（插件名）判断全无权限占位
  if (currentRouteName.value === 'pluginPackageMng') return true;

  // 操作记录：拥有任意包类型的历史查看权限即放行（资源为数组，不能直接按单资源比对）
  if (currentRouteName.value === 'operationRecords') {
    return authStore.hasAuthorizedResource(action);
  }

  return authStore.hasAuthorizedResource(action, currentResourceId.value);
});
</script>
