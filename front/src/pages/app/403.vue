<template>
  <div class="forbidden-page">
    <img class="forbidden-img" src="/images/403.png" alt="403">
    <div class="forbidden-title">{{ t('components.permission.noPermission') }}</div>
    <Table
      v-if="displayActions.length"
      class="forbidden-table"
      align="left"
      :columns="columns"
      :data="displayActions"
      show-overflow-tooltip
      row-hover="auto"
    />
    <div class="forbidden-actions">
      <bk-button
        v-if="canOpenPermissionDialog"
        theme="primary"
        @click="handleApply"
      >
        {{ t('components.permission.apply') }}
      </bk-button>
      <bk-button @click="handleGoBack">
        {{ t('action.back') }}
      </bk-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Table } from 'bkui-vue';
import { computed, h } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';

import { useAuthStore } from '@/stores/auth';
import { usePermissionStore } from '@/stores/permission';

const { t } = useI18n();
const router = useRouter();
const authStore = useAuthStore();
const permissionStore = usePermissionStore();

const permissionDetail = computed(() => authStore.getPermissionDetail());
const deniedActionIds = computed(() => authStore.getDeniedActionIds());
// 只取与被拒 action 匹配的明细：permissionDetail 可能残留其它 action（如单元查看）的 verify 结果，
// 直接使用会导致「去申请」弹出无关权限的申请弹窗
const relatedActions = computed(() => (permissionDetail.value?.actions ?? [])
  .filter(action => deniedActionIds.value.includes(action.id)));
const displayActions = computed(() => deniedActionIds.value.map((actionId) => {
  const matchedAction = permissionDetail.value?.actions?.find(action => action.id === actionId);

  return {
    actionId,
    actionName: matchedAction?.name || actionId,
  };
}));

const canOpenPermissionDialog = computed(() => relatedActions.value.length > 0);

const columns = computed(() => [
  {
    label: t('components.permission.requiredPermissions'),
    width: 700,
    render: ({ data }: { data: { actionId: string; actionName: string } }) => h('span', {}, data.actionName),
  },
]);

function handleApply() {
  if (relatedActions.value.length) {
    permissionStore.showDialog({ ...permissionDetail.value!, actions: relatedActions.value });
  }
}

function handleGoBack() {
  if (window.history.length > 1) {
    router.back();
  } else {
    router.push({ name: 'agent' });
  }
}
</script>

<style lang="postcss" scoped>
.forbidden-page {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  min-height: 400px;
}

.forbidden-img {
  width: 200px;
  height: auto;
}

.forbidden-title {
  margin-top: 20px;
  font-size: 22px;
  font-weight: 400;
  color: #63656e;
}

.forbidden-desc {
  margin-top: 10px;
  font-size: 14px;
  color: #979ba5;
}

.forbidden-table {
  width: 700px;
  max-height: 240px;
  margin-top: 20px;
  overflow-y: auto;
}

.forbidden-actions {
  display: flex;
  gap: 8px;
  margin-top: 24px;
}
</style>
