<template>
  <NoPermission type="biz" v-if="hasNoBizPermission" />
  <RouterView v-else />
</template>

<script setup lang="ts">
import { computed } from 'vue';

import NoPermission from '@/components/no-permission.vue';
import { useAuthStore } from '@/stores/auth';
import { useMainStore } from '@/stores/main';

const authStore = useAuthStore();
const mainStore = useMainStore();

/** 判断是否所有业务都没有 biz_access 权限 */
const hasNoBizPermission = computed(() => {
  // 业务列表为空 或 权限未加载完成时，不展示占位页
  if (!mainStore.businessList.length || !authStore.authorizedLoaded) return false;
  // biz_access isAny=true 表示全部有权限
  const authorizedBizIds = authStore.getAuthorizedBizIds('biz_access');
  if (authorizedBizIds === null) return false; // null = 全部有权限
  // 检查是否有任何一个业务有权限
  return !mainStore.businessList.some(biz => authorizedBizIds.includes(String(biz.bk_biz_id)));
});
</script>
