<template>
  <div class="flex flex-col h-full min-h-0 min-w-0">
    <div class="flex shrink-0 items-center justify-between mb-[12px] text-[12px]">
      <span>{{ $t('schedulerWorkflow.executionStatus') }}: {{ stateText(instance.life_cycle?.state) }}</span>
      <Button :loading="loading" @click="getLogs">{{ $t('schedulerWorkflow.refreshLogs') }}</Button>
    </div>
    <Alert v-if="failed" theme="danger" class="shrink-0 mb-[12px]" :title="$t('schedulerWorkflow.loadFailed')" />
    <Loading :loading="loading" :title="$t('table.loading')" class="flex-1 min-h-0 min-w-0">
      <div class="flex h-full min-h-0 min-w-0 border border-[#dcdee5]">
        <div class="w-[44%] min-w-0 shrink-0 overflow-auto bg-[#f5f7fa]">
          <Table
            :data="stepRows"
            :empty-text="$t('table.empty')"
            :auto-resize="false"
            class="w-full h-full cursor-pointer"
            @row-click="selectStep"
          >
            <TableColumn :title="$t('schedulerWorkflow.step')" :min-width="160">
              <template #default="{ row }">
                <Button
                  text
                  :class="{ 'text-[#3a84ff]': selectedStep === row.key, 'text-[#c4c6cc]': row.state === 'pending' }"
                  :aria-pressed="selectedStep === row.key"
                  @click="selectedStep = row.key"
                >
                  <span v-if="row.key !== null">{{ row.index }}. </span>{{ row.name }}
                </Button>
              </template>
            </TableColumn>
            <TableColumn :title="$t('schedulerWorkflow.executionStatus')" :min-width="110">
              <template #default="{ row }">
                <div class="flex items-center gap-[6px]" :style="{ color: statusColor(row.state) }">
                  <Spinner v-if="row.state === 'running'" width="12px" height="12px" />
                  <span
                    v-else-if="row.key !== null"
                    class="w-[8px] h-[8px] shrink-0 rounded-full border border-current bg-current opacity-70"
                  ></span>
                  <span>{{ stateText(row.state) }}</span>
                </div>
              </template>
            </TableColumn>
            <TableColumn :title="$t('schedulerWorkflow.duration')" :min-width="70">
              <template #default="{ row }">{{ row.duration }}</template>
            </TableColumn>
            <TableColumn :width="32">
              <template #default="{ row }">
                <RightShape v-if="selectedStep === row.key" fill="#C4C6CC" />
              </template>
            </TableColumn>
          </Table>
        </div>
        <div class="flex flex-col flex-1 min-h-0 min-w-0 ml-[10px] bg-[#313238] text-[#c4c6cc] text-[12px]">
          <div class="flex items-center h-[50px] shrink-0 px-[16px] bg-[#202024]">
            {{ $t('schedulerWorkflow.executionLog') }}
          </div>
          <div class="flex-1 min-h-0 min-w-0 overflow-auto p-[16px]">
            <div v-if="!visibleLogs.length" class="text-center py-[24px]">{{ $t('schedulerWorkflow.noLogs') }}</div>
            <div
              v-for="(line, index) in visibleLogs"
              :key="index"
              class="whitespace-pre-wrap break-words mb-[8px] leading-[24px]"
              :class="{ 'text-[#ea3636]': line.level === 'ERROR', 'text-[#ff9c01]': line.level === 'WARN' }"
            >
              <span>[{{ logTime(line.time) }} {{ line.level }}] </span>{{ logText(line) }}
            </div>
          </div>
        </div>
      </div>
    </Loading>
  </div>
</template>

<script setup lang="ts">
import { Alert, Button, Loading } from 'bkui-vue';
import { RightShape, Spinner } from 'bkui-vue/lib/icon';
import { computed, onBeforeUnmount, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';

import { Table, TableColumn } from '@blueking/table';

import type { WorflowOperationInstanceData, WorkflowActionData, WorkflowActionMessageMessage } from '@/@types/workflow';
import { ScheduleWorkflowService } from '@/api/modules/schedule_workflow';
import { formatTimeByTimezone } from '@/common/util';
import { useMainStore } from '@/stores/main';

const props = defineProps<{ instance: WorflowOperationInstanceData }>();
const { t, te } = useI18n();
const mainStore = useMainStore();
const isZh = computed(() => mainStore.curLanguage === 'zh-CN');
const loading = ref(false);
const failed = ref(false);
type StepLog = Omit<WorkflowActionData, 'message'> & { message: { logs: WorkflowActionMessageMessage[] } };
interface InstanceLogs {
  oper_inst_logs: Record<string, StepLog>;
  extra_execution_logs: { logs: WorkflowActionMessageMessage[] };
}
const stepLogs = ref<InstanceLogs['oper_inst_logs']>({});
const extraLogs = ref<WorkflowActionMessageMessage[]>([]);
const selectedStep = ref<string | null>('');
const stepKeys = computed(() => [...new Set([
  ...(props.instance.action_names || []),
  ...Object.keys(stepLogs.value),
])]);
const visibleLogs = computed<WorkflowActionMessageMessage[]>(() => (selectedStep.value === null
  ? extraLogs.value : stepLogs.value[selectedStep.value]?.message?.logs || []));
const stateText = (state?: string) => (state && te(`schedulerWorkflow.states.${state}`)
  ? t(`schedulerWorkflow.states.${state}`) : state || '--');
const stepName = (key: string) => (isZh.value
  ? stepLogs.value[key]?.display_name_zh || key : stepLogs.value[key]?.display_name_en || key);
const logText = (line: WorkflowActionMessageMessage) => (isZh.value
  ? line.text_zh || line.text_en : line.text_en || line.text_zh);
const statusColor = (state?: string) => ({
  success: '#2caf5e',
  failed: '#ea3636',
  timeout: '#ff9c01',
  terminated: '#8e62d1',
  running: '#3a84ff',
  pending: '#c4c6cc',
}[state || ''] || '#979ba5');
const isValidTime = (time?: number): time is number => typeof time === 'number'
  && Number.isFinite(time) && time > 0 && Number.isFinite(new Date(time).getTime());
const stepDuration = (key: string) => {
  const { start_time: start, end_time: end } = stepLogs.value[key]?.life_cycle || {};
  if (!isValidTime(start) || !isValidTime(end) || end < start) return '--';
  return `${Math.floor((end - start) / 1000)}s`;
};
const logTime = (time?: number) => (isValidTime(time) ? formatTimeByTimezone(time) : '--');
const stepRows = computed(() => [
  ...stepKeys.value.map((key, index) => ({
    key,
    index: index + 1,
    name: stepName(key),
    state: stepLogs.value[key]?.life_cycle?.state,
    duration: stepDuration(key),
  })),
  ...(extraLogs.value.length ? [{
    key: null,
    index: 0,
    name: t('schedulerWorkflow.extraLogs'),
    state: undefined,
    duration: '--',
  }] : []),
]);
const selectStep = (_event: Event, row: { key: string | null }) => {
  selectedStep.value = row.key;
};
let requestId = 0;

const getLogs = async () => {
  requestId += 1;
  const currentRequest = requestId;
  loading.value = true;
  failed.value = false;
  try {
    const res = await ScheduleWorkflowService.ScheduleWorkflowOperationInstanceLogGet<unknown, InstanceLogs>({
      oper_inst_id: props.instance.oper_inst_id,
    });
    if (currentRequest !== requestId) return;
    stepLogs.value = res.oper_inst_logs || {};
    extraLogs.value = res.extra_execution_logs?.logs || [];
    if ((selectedStep.value === null && !extraLogs.value.length)
      || (selectedStep.value !== null && !stepKeys.value.includes(selectedStep.value))) {
      selectedStep.value = stepKeys.value[0] ?? null;
    }
  } catch {
    if (currentRequest !== requestId) return;
    stepLogs.value = {};
    extraLogs.value = [];
    failed.value = true;
  } finally {
    if (currentRequest === requestId) loading.value = false;
  }
};
watch(() => props.instance.oper_inst_id, () => {
  stepLogs.value = {};
  extraLogs.value = [];
  selectedStep.value = '';
  getLogs();
}, { immediate: true });
onBeforeUnmount(() => {
  requestId += 1;
});
</script>
