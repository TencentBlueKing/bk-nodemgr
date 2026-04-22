<template>
  <Dialog
    class="permissions-dialog-cls"
    :is-show="permissionStore.visible"
    :width="740"
    theme="primary"
    :z-index="9000"
    @closed="permissionStore.hideDialog()"
  >
    <img class="no-permission-img" src="/images/403.png" alt="403">
    <div class="no-permission-text">{{ t('components.permission.noPermission') }}</div>
    <Table
      class="mt20 no-permission-table"
      align="left"
      :columns="columns"
      :data="actions"
      show-overflow-tooltip
      row-hover="auto"
    />
    <template #footer>
      <div class="flex justify-end gap-[8px]">
        <bk-button
          theme="primary"
          :disabled="!applyUrl"
          @click="handleApply"
        >
          {{ t('components.permission.apply') }}
        </bk-button>
        <bk-button @click="permissionStore.hideDialog()">{{ t('action.cancel') }}</bk-button>
      </div>
    </template>
  </Dialog>
</template>

<script setup lang="ts">
import { Dialog, InfoBox, Message, Table } from 'bkui-vue';
import { computed, h } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';

import type { PermissionAction } from '@/stores/permission';
import { usePermissionStore } from '@/stores/permission';
import { useAuthStore } from '@/stores/auth';

const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const permissionStore = usePermissionStore();
const authStore = useAuthStore();

const systemName = computed(() => permissionStore.data?.system_name ?? '');
const actions = computed(() => permissionStore.data?.actions ?? []);
const applyUrl = computed(() => permissionStore.data?.apply_url ?? '');

function formatResourceInstance(instance: {
  type_name: string;
  id: string;
  name: string;
}): string {
  const trimmedName = instance.name?.trim();
  const trimmedID = instance.id?.trim();

  if (trimmedName && trimmedID && trimmedName !== trimmedID) {
    return `${trimmedName} (${trimmedID})`;
  }

  if (trimmedName) {
    return trimmedName;
  }

  if (trimmedID) {
    return trimmedID;
  }

  return instance.type_name;
}

function formatRelatedResources(data: PermissionAction): string {
  const relatedResourceTypes = data?.related_resource_types;
  if (!relatedResourceTypes?.length) {
    return '--';
  }

  return relatedResourceTypes.map((resource) => {
    const instances = resource.instances ?? [];
    if (!instances.length) {
      return resource.type_name;
    }

    const renderedInstances = instances.map(formatResourceInstance).join('、');
    return `${resource.type_name}: ${renderedInstances}`;
  }).join('；');
}

const columns = computed(() => [
  {
    label: t('components.permission.system'),
    width: 150,
    render: () => h('span', {}, systemName.value),
  },
  {
    label: t('components.permission.requiredPermissions'),
    field: 'name',
    width: 200,
  },
  {
    label: t('components.permission.relatedResources'),
    width: 342,
    render: ({ data }: { data: PermissionAction }) => {
      return h('span', { class: 'permission-related-resources' }, formatRelatedResources(data));
    },
  },
]);

function handleApply() {
  if (!applyUrl.value) return;
  const newWindow = window.open(applyUrl.value, '_blank', 'noopener,noreferrer');
  // 无论是否被浏览器拦截，统一关闭权限弹窗并弹出刷新提醒
  // 拦截时额外用 Message 告知用户手动打开
  permissionStore.hideDialog();
  if (!newWindow) {
    Message({
      theme: 'warning',
      message: t('components.permission.popupBlocked'),
      width: 562,
    });
  }
  InfoBox({
    title: t('components.permission.refreshReminder.title'),
    subTitle: t('components.permission.refreshReminder.content'),
    confirmText: t('components.permission.refreshReminder.refresh'),
    cancelText: t('components.permission.refreshReminder.close'),
    onConfirm: () => {
      // 在 403 页面点"刷新"时，直接 reload 会一直停留在 403。
      // 跳 403 时 router 已将原目标路由的 fullPath 写入 query.from，
      // 这里用 replace 导航回去，beforeEach 会重新 verify 权限；
      // 若用户已授权即可正常展示，未授权仍会拦截回 403（符合预期）。
      if (route.name === '403') {
        // 清除权限缓存，确保 beforeEach 一定会重新调 batchVerify 而非使用旧的 false 缓存
        authStore.reset();
        const from = route.query?.from;
        if (typeof from === 'string' && from && from !== route.fullPath) {
          router.replace(from);
          return;
        }
      }
      window.location.reload();
    },
  });
}
</script>

<style lang="postcss">
.permissions-dialog-cls {
  .bk-dialog-header {
    padding: 0 !important;
  }

  .bk-modal-content {
    overflow: hidden !important;
  }

  .no-permission-table {
    max-height: 200px;
    overflow-y: auto;

    .permission-related-resources {
      display: inline-block;
      line-height: 20px;
      white-space: normal;
      word-break: break-word;
      text-align: left;
    }
  }

  text-align: center;

  .no-permission-img {
    display: block;
    margin: 0 auto;
    height: 100px;
  }

  .no-permission-text {
    color: #63656e;
    font-size: 22px;
    font-weight: 400;
    margin: 15px 0 30px;
  }
}
</style>
