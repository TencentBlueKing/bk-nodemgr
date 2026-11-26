<template>
  <div>
    <div class="flex items-center justify-between mb-[12px] text-[12px]">
      <span>{{ $t('schedulerWorkflow.executionStatus') }}: {{ stateText(instance.life_cycle?.state) }}</span>
      <Button :loading="loading" @click="getLogs">{{ $t('schedulerWorkflow.refreshLogs') }}</Button>
    </div>
    <Alert v-if="failed" theme="danger" class="mb-[12px]" :title="$t('schedulerWorkflow.loadFailed')" />
    <Loading :loading="loading" :title="$t('table.loading')">
      <div class="flex border border-[#dcdee5] min-h-[280px]">
        <div class="w-[300px] shrink-0 max-h-[480px] overflow-y-auto bg-[#f5f7fa]">
          <button
            v-for="(key, index) in stepKeys"
            :key="key"
            type="button"
            class="block w-full text-left p-[12px] cursor-pointer border-0 bg-transparent"
            :class="{ '!bg-[#e1ecff] text-[#3a84ff]': selectedStep === key }"
            @click="selectedStep = key"
          >
            <div class="break-words">{{ index + 1 }}. {{ stepName(key) }}</div>
            <div class="mt-[4px] text-[12px] text-[#63656e]">
              {{ stateText(stepLogs[key]?.life_cycle?.state) }}
            </div>
          </button>
          <button
            v-if="extraLogs.length"
            type="button"
            class="block w-full text-left p-[12px] cursor-pointer border-0 bg-transparent"
            :class="{ '!bg-[#e1ecff] text-[#3a84ff]': selectedStep === null }"
            @click="selectedStep = null"
          >{{ $t('schedulerWorkflow.extraLogs') }}</button>
        </div>
        <div class="flex-1 min-w-0 bg-[#313238] text-[#c4c6cc] p-[16px] max-h-[480px] overflow-auto">
          <div v-if="!visibleLogs.length" class="text-center py-[24px]">{{ $t('schedulerWorkflow.noLogs') }}</div>
          <div
            v-for="(line, index) in visibleLogs"
            :key="index"
            class="text-[12px] whitespace-pre-wrap break-words mb-[8px]"
            :class="{ 'text-[#ea3636]': line.level === 'ERROR', 'text-[#ff9c01]': line.level === 'WARN' }"
          >
            <span>[{{ line.time ? formatTimeByTimezone(line.time) : '--' }} {{ line.level }}] </span>{{ logText(line) }}
          </div>
        </div>
      </div>
    </Loading>
  </div>
</template>

<script setup lang="ts">
import { Alert, Button, Loading } from 'bkui-vue';
import { computed, onBeforeUnmount, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';

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
    if (selectedStep.value !== null && !stepKeys.value.includes(selectedStep.value)) {
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
