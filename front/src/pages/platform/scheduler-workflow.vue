<template>
  <div class="p-[24px] h-full flex flex-col">
    <div class="flex items-center justify-between gap-[16px] mb-[16px]">
      <Button :loading="loading" @click="getWorkflows">{{ $t('schedulerWorkflow.refresh') }}</Button>
      <SearchSelect
        v-model="searchValues"
        class="w-[640px] bg-white"
        :data="searchOptions"
        :unique-select="true"
        :placeholder="$t('schedulerWorkflow.searchPlaceholder')"
      />
    </div>
    <Alert v-if="failed" theme="danger" class="mb-[16px]" :title="$t('schedulerWorkflow.loadFailed')" />
    <Loading :loading="loading" :title="$t('table.loading')" class="flex-1 overflow-auto">
      <Table
        :data="workflows"
        :pagination="pagination"
        :max-height="maxHeight"
        :empty-text="failed ? $t('schedulerWorkflow.loadFailed') : $t('table.empty')"
        show-overflow-tooltip
        @page-limit-change="handleLimitChange"
        @page-value-change="handlePageChange"
      >
        <TableColumn field="workflow_name" :title="$t('schedulerWorkflow.name')" min-width="220" fixed="left" />
        <TableColumn field="workflow_id" title="Workflow ID" min-width="280" />
        <TableColumn field="interval" :title="$t('schedulerWorkflow.interval')" min-width="140" />
        <TableColumn field="enabled" :title="$t('schedulerWorkflow.scheduleStatus')" min-width="100">
          <template #default="{ row }">
            <Tag v-if="row.enabled" theme="success">{{ $t('schedulerWorkflow.enabled') }}</Tag>
            <Tag v-else>{{ $t('schedulerWorkflow.disabled') }}</Tag>
          </template>
        </TableColumn>
        <TableColumn field="operator" :title="$t('schedulerWorkflow.operator')" min-width="140">
          <template #default="{ row }"><UserNameDisplay :name="row.operator || ''" /></template>
        </TableColumn>
        <TableColumn field="operate_time" :title="$t('schedulerWorkflow.operateTime')" min-width="180">
          <template #default="{ row }">{{ row.operate_time ? formatTimeByTimezone(row.operate_time) : '--' }}</template>
        </TableColumn>
        <TableColumn :title="$t('table.action')" min-width="280" fixed="right">
          <template #default="{ row }">
            <Button theme="primary" text @click="showHistory(row)">{{ $t('schedulerWorkflow.history') }}</Button>
            <Button theme="primary" text class="ml-[16px]" @click="showIntervalDialog(row)">
              {{ $t('schedulerWorkflow.editInterval') }}
            </Button>
          </template>
        </TableColumn>
      </Table>
    </Loading>
    <Sideslider
      v-model:is-show="historyVisible"
      :title="historyTitle"
      :width="1200"
      render-directive="if"
    >
      <SchedulerHistory
        v-if="historyVisible && selectedWorkflow"
        :key="selectedWorkflow.workflow_id"
        :workflow-id="selectedWorkflow.workflow_id"
      />
    </Sideslider>
    <SchedulerIntervalDialog
      v-if="intervalWorkflow"
      v-model:is-show="intervalVisible"
      :workflow="intervalWorkflow"
      @updated="getWorkflows"
    />
  </div>
</template>

<script setup lang="ts">
import { Alert, Button, Loading, SearchSelect, Sideslider, Tag } from 'bkui-vue';
import { computed, onBeforeUnmount, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute } from 'vue-router';

import { Table, TableColumn } from '@blueking/table';

import SchedulerHistory from './scheduler-history.vue';
import SchedulerIntervalDialog from './scheduler-interval-dialog.vue';

import type { ScheduleWorkflowListRespScheduleWorkflow } from '@/@types/schedule_workflow';
import { ScheduleWorkflowService } from '@/api/modules/schedule_workflow';
import { formatTimeByTimezone } from '@/common/util';
import UserNameDisplay from '@/components/user-name-display.vue';
import { useMainStore } from '@/stores/main';

type ScheduledWorkflow = ScheduleWorkflowListRespScheduleWorkflow;
type SearchField = 'workflow_id' | 'workflow_name' | 'operator';
interface SearchValue {
  id: SearchField;
  name: string;
  values: { id: string; name: string }[];
}

const { t } = useI18n();
const route = useRoute();
const mainStore = useMainStore();
const maxHeight = computed(() => mainStore.windowInnerHeight - 220 - (mainStore.noticeShow ? 40 : 0));
const pagination = reactive({ count: 0, limit: 50, current: 1, remote: true });
const workflows = ref<ScheduledWorkflow[]>([]);
const loading = ref(false);
const failed = ref(false);
const searchValues = ref<SearchValue[]>([]);
const searchOptions = computed(() => [
  { id: 'workflow_name', name: t('schedulerWorkflow.name'), multiple: true },
  { id: 'workflow_id', name: 'Workflow ID', multiple: true },
  { id: 'operator', name: t('schedulerWorkflow.operator'), multiple: true },
]);
const intervalVisible = ref(false);
const intervalWorkflow = ref<ScheduledWorkflow>();
const historyVisible = ref(false);
const selectedWorkflow = ref<Pick<ScheduledWorkflow, 'workflow_id' | 'workflow_name'>>();
const historyTitle = computed(() => `${t('schedulerWorkflow.history')} - ${
  selectedWorkflow.value?.workflow_name || selectedWorkflow.value?.workflow_id || ''}`);
let requestId = 0;

const getWorkflows = async () => {
  requestId += 1;
  const currentRequest = requestId;
  loading.value = true;
  failed.value = false;
  try {
    const conditions: Partial<Record<SearchField, string[]>> = {};
    searchValues.value.forEach((item) => {
      if (item.values?.length) conditions[item.id] = item.values.map(value => value.id);
    });
    const res = await ScheduleWorkflowService.ScheduleWorkflowList<unknown, {
      total: number; items: ScheduledWorkflow[];
    }>({
      page: { limit: pagination.limit, offset: (pagination.current - 1) * pagination.limit },
      exact_include_conditions: conditions,
    });
    if (currentRequest !== requestId) return;
    workflows.value = res.items || [];
    pagination.count = res.total;
  } catch {
    if (currentRequest !== requestId) return;
    workflows.value = [];
    pagination.count = 0;
    failed.value = true;
  } finally {
    if (currentRequest === requestId) loading.value = false;
  }
};
const handleLimitChange = (limit: number) => {
  pagination.limit = limit;
  pagination.current = 1;
  getWorkflows();
};
const handlePageChange = (page: number) => {
  pagination.current = page;
  getWorkflows();
};
const showIntervalDialog = (workflow: ScheduledWorkflow) => {
  intervalWorkflow.value = workflow;
  intervalVisible.value = true;
};
const showHistory = (workflow: ScheduledWorkflow) => {
  selectedWorkflow.value = workflow;
  historyVisible.value = true;
};
watch(searchValues, () => {
  pagination.current = 1;
  getWorkflows();
}, { deep: true, immediate: true });
watch(() => route.query.workflow_id, (id) => {
  if (typeof id !== 'string' || !id) return;
  selectedWorkflow.value = { workflow_id: id, workflow_name: '' };
  historyVisible.value = true;
}, { immediate: true });
onBeforeUnmount(() => {
  requestId += 1;
});
</script>
