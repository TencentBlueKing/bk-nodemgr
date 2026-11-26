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
    <div v-if="selectedOperation" class="mt-[24px]">
      <div class="text-[12px] mb-[12px] break-all">Operation ID: {{ selectedOperation }}</div>
      <Alert v-if="instanceFailed" theme="danger" class="mb-[12px]" :title="$t('schedulerWorkflow.loadFailed')" />
      <Loading :loading="instanceLoading" :title="$t('table.loading')">
        <div class="flex gap-[12px] items-center mb-[16px]">
          <span class="text-[14px]">{{ $t('schedulerWorkflow.instance') }}</span>
          <Select v-model="selectedInstanceId" class="flex-1" :clearable="false" :placeholder="$t('table.empty')">
            <Select.Option
              v-for="instance in instances"
              :key="instance.oper_inst_id"
              :value="instance.oper_inst_id"
              :label="`${instance.oper_inst_id} (${stateText(instance.life_cycle?.state)})`"
            />
          </Select>
          <Button :loading="instanceLoading" @click="getInstances">{{ $t('schedulerWorkflow.refresh') }}</Button>
        </div>
        <SchedulerLog v-if="selectedInstance" :key="selectedInstance.oper_inst_id" :instance="selectedInstance" />
        <div v-else class="py-[24px] text-center text-[#979ba5]">{{ $t('table.empty') }}</div>
      </Loading>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Alert, Button, Loading, Select } from 'bkui-vue';
import { computed, onBeforeUnmount, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';

import { Table, TableColumn } from '@blueking/table';

import SchedulerLog from './scheduler-log.vue';

import type { ScheduleWorkflowOperationListRespScheduleWorkflowOperation } from '@/@types/schedule_workflow';
import type { WorflowOperationInstanceData } from '@/@types/workflow';
import { ScheduleWorkflowService } from '@/api/modules/schedule_workflow';
import { formatTimeByTimezone } from '@/common/util';
import UserNameDisplay from '@/components/user-name-display.vue';

const props = defineProps<{ workflowId: string }>();
const { t, te } = useI18n();
const stateText = (state?: string) => (state && te(`schedulerWorkflow.states.${state}`)
  ? t(`schedulerWorkflow.states.${state}`) : state || '--');
const pagination = reactive({ count: 0, limit: 20, current: 1, remote: true });
type ScheduledOperation = ScheduleWorkflowOperationListRespScheduleWorkflowOperation;
const operations = ref<ScheduledOperation[]>([]);
const loading = ref(false);
const failed = ref(false);
const selectedOperation = ref('');
const instances = ref<WorflowOperationInstanceData[]>([]);
const instanceLoading = ref(false);
const instanceFailed = ref(false);
const selectedInstanceId = ref('');
const selectedInstance = computed(() => instances.value.find(item => item.oper_inst_id === selectedInstanceId.value));
let operationRequestId = 0;
let instanceRequestId = 0;

const getOperations = async () => {
  operationRequestId += 1;
  const requestId = operationRequestId;
  loading.value = true;
  failed.value = false;
  selectedOperation.value = '';
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
const getInstances = async () => {
  instanceRequestId += 1;
  const requestId = instanceRequestId;
  const operationId = selectedOperation.value;
  instances.value = [];
  selectedInstanceId.value = '';
  instanceFailed.value = false;
  instanceLoading.value = false;
  if (!operationId) return;
  instanceLoading.value = true;
  try {
    const res = await ScheduleWorkflowService.ScheduleWorkflowOperationInstanceList<unknown, {
      total: number; oper_inst_data: WorflowOperationInstanceData[];
    }>({ operation_id: [operationId] });
    if (requestId !== instanceRequestId) return;
    instances.value = res.oper_inst_data || [];
    selectedInstanceId.value = instances.value[instances.value.length - 1]?.oper_inst_id || '';
  } catch {
    if (requestId !== instanceRequestId) return;
    instanceFailed.value = true;
  } finally {
    if (requestId === instanceRequestId) instanceLoading.value = false;
  }
};
const selectOperation = (operationId: string) => {
  if (selectedOperation.value === operationId) {
    getInstances();
    return;
  }
  selectedOperation.value = operationId;
};
const handleLimitChange = (limit: number) => {
  pagination.limit = limit;
  pagination.current = 1;
  getOperations();
};
const handlePageChange = (page: number) => {
  pagination.current = page;
  getOperations();
};
watch(selectedOperation, getInstances);
watch(() => props.workflowId, () => {
  pagination.current = 1;
  getOperations();
}, { immediate: true });
onBeforeUnmount(() => {
  operationRequestId += 1;
  instanceRequestId += 1;
});
</script>
