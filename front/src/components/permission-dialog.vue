<template>
  <Dialog
    class="permissions-dialog-cls"
    :is-show="permissionStore.visible"
    :width="740"
    theme="primary"
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

import type { PermissionAction } from '@/stores/permission';
import { usePermissionStore } from '@/stores/permission';

const { t } = useI18n();
const permissionStore = usePermissionStore();

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

    // instances 是层级路径数组，每一层显示为 type_name: id，用斜杠连接
    const renderedPath = instances.map((instance) => {
      const displayValue = formatResourceInstance(instance);
      return `${instance.type_name}: ${displayValue}`;
    }).join('/');

    return renderedPath;
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
  if (applyUrl.value) {
    const newWindow = window.open(applyUrl.value, '_blank', 'noopener,noreferrer');
    if (newWindow) {
      // 打开成功：关闭权限弹窗，弹出刷新提醒
      permissionStore.hideDialog();
      InfoBox({
        title: t('components.permission.refreshReminder.title'),
        subTitle: t('components.permission.refreshReminder.content'),
        confirmText: t('components.permission.refreshReminder.refresh'),
        cancelText: t('components.permission.refreshReminder.close'),
        onConfirm: () => {
          window.location.reload();
        },
      });
    } else {
      // 弹窗被浏览器拦截：保留原权限弹窗，提示用户
      Message({
        theme: 'warning',
        message: t('components.permission.popupBlocked'),
        width: 562,
      });
    }
  }
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
