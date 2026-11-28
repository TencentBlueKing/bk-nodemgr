<template>
  <div class="h-full flex flex-col min-h-0">
    <PageHeader :title="'schedulerWorkflow.executionDetail'" :on-back="backToHistory" />
    <section class="flex-1 min-h-0 flex flex-col p-[24px] gap-[16px]">
      <div class="flex flex-wrap items-center gap-x-[32px] gap-y-[8px] text-[12px] text-[#63656e]">
        <span class="break-all">Workflow ID: {{ workflowId }}</span>
        <span class="break-all">Operation ID: {{ operationId }}</span>
      </div>
      <div class="flex items-center gap-[16px]">
        <span class="text-[14px] shrink-0">{{ $t('schedulerWorkflow.instance') }}</span>
        <Select
          v-model="selectedInstanceId"
          class="flex-1 min-w-0"
          :clearable="false"
          :loading="loading"
          :placeholder="$t('table.empty')"
        >
          <Select.Option
            v-for="instance in instances"
            :key="instance.oper_inst_id"
            :value="instance.oper_inst_id"
            :label="`${instance.oper_inst_id} (${stateText(instance.life_cycle?.state)})`"
          />
        </Select>
        <Button :loading="loading" @click="getInstances">{{ $t('schedulerWorkflow.refresh') }}</Button>
      </div>
      <Alert v-if="failed" theme="danger" :title="$t('schedulerWorkflow.loadFailed')" />
      <Loading :loading="loading" :title="$t('table.loading')" class="flex-1 min-h-0">
        <SchedulerLog
          v-if="selectedInstance"
          :key="selectedInstance.oper_inst_id"
          :instance="selectedInstance"
          class="h-full"
        />
        <div v-else class="py-[48px] text-center text-[#979ba5]">
          {{ failed ? $t('schedulerWorkflow.loadFailed') : $t('table.empty') }}
        </div>
      </Loading>
    </section>
  </div>
</template>

<script setup lang="ts">
import { Alert, Button, Loading, Select } from 'bkui-vue';
import { computed, onBeforeUnmount, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';

import SchedulerLog from './scheduler-log.vue';

import type { WorflowOperationInstanceData } from '@/@types/workflow';
import { ScheduleWorkflowService } from '@/api/modules/schedule_workflow';
import PageHeader from '@/components/page-header.vue';

const route = useRoute();
const router = useRouter();
const { t, te } = useI18n();
const workflowId = computed(() => String(route.params.workflowId || ''));
const operationId = computed(() => String(route.params.operationId || ''));
const instances = ref<WorflowOperationInstanceData[]>([]);
const selectedInstanceId = ref('');
const selectedInstance = computed(() => instances.value.find(item => item.oper_inst_id === selectedInstanceId.value));
const loading = ref(false);
const failed = ref(false);
const stateText = (state?: string) => (state && te(`schedulerWorkflow.states.${state}`)
  ? t(`schedulerWorkflow.states.${state}`) : state || '--');
let requestId = 0;

const getInstances = async () => {
  requestId += 1;
  const currentRequest = requestId;
  loading.value = true;
  failed.value = false;
  try {
    const res = await ScheduleWorkflowService.ScheduleWorkflowOperationInstanceList<unknown, {
      total: number; oper_inst_data: WorflowOperationInstanceData[];
    }>({ operation_id: [operationId.value] });
    if (currentRequest !== requestId) return;
    instances.value = res.oper_inst_data || [];
    if (!instances.value.some(item => item.oper_inst_id === selectedInstanceId.value)) {
      selectedInstanceId.value = instances.value[instances.value.length - 1]?.oper_inst_id || '';
    }
  } catch {
    if (currentRequest !== requestId) return;
    instances.value = [];
    selectedInstanceId.value = '';
    failed.value = true;
  } finally {
    if (currentRequest === requestId) loading.value = false;
  }
};
const backToHistory = () => router.push({
  name: 'schedulerWorkflow',
  query: { workflow_id: workflowId.value },
});
watch([workflowId, operationId], () => {
  instances.value = [];
  selectedInstanceId.value = '';
  getInstances();
}, { immediate: true });
onBeforeUnmount(() => {
  requestId += 1;
});
</script>
