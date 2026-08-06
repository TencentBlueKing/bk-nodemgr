<template>
  <div class="no-permission-page">
    <bk-exception
      class="exception-wrap-item exception-part"
      type="403"
      scene="part"
      :title="title"
      :description="description"
    >
      <!-- action 模式：带申请按钮 -->
      <bk-button
        v-if="type === 'action'"
        theme="primary"
        :loading="loading"
        @click="handleApply"
      >
        {{ t('components.permission.apply') }}
      </bk-button>
    </bk-exception>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue';
import { useI18n } from 'vue-i18n';

import type { PageAuthItem } from '@/constants/auth';
import { useAuthStore } from '@/stores/auth';
import { usePermissionStore } from '@/stores/permission';

const props = withDefaults(defineProps<{
  /** 展示类型：action=无操作权限(带申请按钮)，biz=无业务权限(带提示文字)，plugin=无插件包权限 */
  type?: 'action' | 'biz' | 'plugin';
  authItems?: PageAuthItem[];
  /** biz 资源的业务范围（biz 级权限时使用） */
  bizScope?: string | number | Array<string | number>;
  /** 非 biz 资源的具体资源 ID（如 package 的 release_type、networkarea 的 areaId） */
  resourceId?: string | number;
}>(), {
  type: 'action',
  authItems: () => [],
});

const { t } = useI18n();
const authStore = useAuthStore();
const permissionStore = usePermissionStore();
const loading = ref(false);

const title = computed(() => {
  if (props.type === 'biz') return t('components.permission.noBizPermission');
  if (props.type === 'plugin') return t('components.permission.noPluginPermission');
  return t('components.permission.noPermission');
});

const description = computed(() => {
  if (props.type === 'biz') return t('components.permission.noBizPermissionTip');
  if (props.type === 'plugin') return t('components.permission.noPluginPermissionTip');
  return '';
});

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
  align-items: center;
  justify-content: center;
  height: 100%;
  min-height: 400px;
}

.exception-wrap-item.exception-part {
  width: auto;

  :deep(.bk-exception-text) {
    .bk-exception-title {
      font-size: 20px;
      font-weight: 700;
      color: #313238;
    }

    .bk-exception-description {
      margin-top: 8px;
      font-size: 14px;
      color: #979ba5;
    }
  }
}
</style>
