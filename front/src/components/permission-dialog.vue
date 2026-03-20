<template>
  <Dialog
    class="permissions-dialog-cls"
    :is-show="permissionStore.visible"
    :width="740"
    theme="primary"
    @closed="permissionStore.hideDialog()"
  >
    <img class="no-permission-img" src="/images/403.png" alt="403">
    <div class="no-permission-text">没有权限访问或操作此资源</div>
    <Table
      class="mt20 no-permission-table"
      align="left"
      :columns="columns"
      :data="actions"
      show-overflow-tooltip
      row-hover="auto"
    />
    <template #footer>
      <bk-button
        class="mr10"
        theme="primary"
        :disabled="!applyUrl"
        @click="handleApply"
      >
        去申请
      </bk-button>
      <bk-button @click="permissionStore.hideDialog()">取消</bk-button>
    </template>
  </Dialog>
</template>

<script setup lang="ts">
import { Dialog, Table } from 'bkui-vue';
import { computed, h } from 'vue';

import type { PermissionAction } from '@/stores/permission';
import { usePermissionStore } from '@/stores/permission';

const permissionStore = usePermissionStore();

const systemName = computed(() => permissionStore.data?.system_name ?? '');
const actions = computed(() => permissionStore.data?.actions ?? []);
const applyUrl = computed(() => permissionStore.data?.apply_url ?? '');

const columns = computed(() => [
  {
    label: '系统',
    width: 150,
    render: () => h('span', {}, systemName.value),
  },
  {
    label: '需要申请的权限',
    field: 'name',
    width: 200,
  },
  {
    label: '关联的资源实例',
    width: 342,
    render: ({ data }: { data: PermissionAction }) => {
      const relatedResourceTypes = data?.related_resource_types;
      if (!relatedResourceTypes?.length) {
        return h('span', {}, '--');
      }
      return h('span', {}, relatedResourceTypes.map(resource => resource.type_name).join('、'));
    },
  },
]);

function handleApply() {
  if (applyUrl.value) {
    window.open(applyUrl.value, '_blank', 'noopener,noreferrer');
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
