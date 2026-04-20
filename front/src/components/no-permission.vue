<template>
  <div class="no-permission-page">
    <img class="no-permission-img" src="/images/403.png" alt="403">
    <div class="no-permission-title">{{ t('components.permission.noPermission') }}</div>
    <bk-button
      theme="primary"
      class="no-permission-btn"
      :loading="loading"
      @click="handleApply"
    >
      {{ t('components.permission.apply') }}
    </bk-button>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';

import type { PageAuthItem } from '@/constants/auth';
import { useAuthStore } from '@/stores/auth';
import { usePermissionStore } from '@/stores/permission';

const props = defineProps<{
  authItems: PageAuthItem[];
  /** biz 资源的业务范围（biz 级权限时使用） */
  bizScope?: string | number | Array<string | number>;
  /** 非 biz 资源的具体资源 ID（如 package 的 release_type、networkarea 的 areaId） */
  resourceId?: string | number;
}>();

const { t } = useI18n();
const authStore = useAuthStore();
const permissionStore = usePermissionStore();
const loading = ref(false);

async function handleApply() {
  if (!props.authItems?.length) return;
  loading.value = true;
  try {
    await authStore.batchVerify(props.authItems, props.bizScope, props.resourceId);
    const detail = authStore.getPermissionDetail();
    if (detail?.actions?.length) {
      permissionStore.showDialog(detail);
    }
  } finally {
    loading.value = false;
  }
}
</script>

<style lang="postcss" scoped>
.no-permission-page {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  min-height: 400px;
}

.no-permission-img {
  width: 200px;
  height: auto;
}

.no-permission-title {
  margin-top: 20px;
  font-size: 22px;
  font-weight: 400;
  color: #63656e;
}

.no-permission-btn {
  margin-top: 24px;
}
</style>

