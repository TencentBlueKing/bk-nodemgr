<template>
  <div class="p-[24px]">
    <div class="flex items-center justify-between mb-[16px]">
      <span class="text-[12px] text-[#63656e] break-all">Workflow ID: {{ workflowId }}</span>
      <Button :loading="loading" @click="getOperations">{{ $t('schedulerWorkflow.refresh') }}</Button>
    </div>
    <Alert v-if="failed" theme="danger" class="mb-[16px]" :title="$t('schedulerWorkflow.loadFailed')" />
    <Loading :loading="loading" :title="$t('table.loading')">
      <Table
        :data="operations"
        :pagination="pagination"
        :max-height="320"
        :empty-text="failed ? $t('schedulerWorkflow.loadFailed') : $t('table.empty')"
        show-overflow-tooltip
        @page-limit-change="handleLimitChange"
        @page-value-change="handlePageChange"
      >
        <TableColumn field="operation_id" title="Operation ID" min-width="260">
          <template #default="{ row }">
            <Button text theme="primary" @click="selectOperation(row.operation_id)">{{ row.operation_id }}</Button>
          </template>
        </TableColumn>
        <TableColumn :title="$t('schedulerWorkflow.executionStatus')" min-width="110">
          <template #default="{ row }">
            {{ stateText(row.latest_oper_inst_brief_data?.life_cycle?.state) }}
          </template>
        </TableColumn>
        <TableColumn field="create_time" :title="$t('schedulerWorkflow.createTime')" min-width="180">
          <template #default="{ row }">{{ row.create_time ? formatTimeByTimezone(row.create_time) : '--' }}</template>
        </TableColumn>
        <TableColumn field="operator" :title="$t('schedulerWorkflow.operator')" min-width="120">
          <template #default="{ row }"><UserNameDisplay :name="row.operator || ''" /></template>
        </TableColumn>
        <TableColumn field="parent_operation_id" :title="$t('schedulerWorkflow.parentOperation')" min-width="260" />
        <TableColumn field="parent_oper_inst_id" :title="$t('schedulerWorkflow.parentInstance')" min-width="260" />
      </Table>
    </Loading>
  </div>
</template>

<script setup lang="ts">
import { Alert, Button, Loading } from 'bkui-vue';
import { onBeforeUnmount, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';

import { Table, TableColumn } from '@blueking/table';

import type { ScheduleWorkflowOperationListRespScheduleWorkflowOperation } from '@/@types/schedule_workflow';
import { ScheduleWorkflowService } from '@/api/modules/schedule_workflow';
import { formatTimeByTimezone } from '@/common/util';
import UserNameDisplay from '@/components/user-name-display.vue';

const props = defineProps<{ workflowId: string }>();
const { t, te } = useI18n();
const router = useRouter();
const stateText = (state?: string) => (state && te(`schedulerWorkflow.states.${state}`)
  ? t(`schedulerWorkflow.states.${state}`) : state || '--');
const pagination = reactive({ count: 0, limit: 20, current: 1, remote: true });
type ScheduledOperation = ScheduleWorkflowOperationListRespScheduleWorkflowOperation;
const operations = ref<ScheduledOperation[]>([]);
const loading = ref(false);
const failed = ref(false);
let operationRequestId = 0;

const getOperations = async () => {
  operationRequestId += 1;
  const requestId = operationRequestId;
  loading.value = true;
  failed.value = false;
  try {
    const res = await ScheduleWorkflowService.ScheduleWorkflowOperationList<unknown, {
      total: number; operations: ScheduledOperation[];
    }>({
      workflow_id: props.workflowId,
      page: { limit: pagination.limit, offset: (pagination.current - 1) * pagination.limit },
    });
    if (requestId !== operationRequestId) return;
    operations.value = res.operations || [];
    pagination.count = res.total;
  } catch {
    if (requestId !== operationRequestId) return;
    operations.value = [];
    pagination.count = 0;
    failed.value = true;
  } finally {
    if (requestId === operationRequestId) loading.value = false;
  }
};
const selectOperation = (operationId: string) => router.push({
  name: 'schedulerDetail',
  params: { workflowId: props.workflowId, operationId },
});
const handleLimitChange = (limit: number) => {
  pagination.limit = limit;
  pagination.current = 1;
  getOperations();
};
const handlePageChange = (page: number) => {
  pagination.current = page;
  getOperations();
};
watch(() => props.workflowId, () => {
  pagination.current = 1;
  getOperations();
}, { immediate: true });
onBeforeUnmount(() => {
  operationRequestId += 1;
});
</script>
