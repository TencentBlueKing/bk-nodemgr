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
  if (!mainStore.businessList.length || !authStore.authorizedLoaded) return false;
  const authorizedBizIds = authStore.getAuthorizedBizIds('biz_access');
  if (authorizedBizIds === null) return false;
  return !mainStore.businessList.some(biz => authorizedBizIds.includes(String(biz.bk_biz_id)));
});
</script>
