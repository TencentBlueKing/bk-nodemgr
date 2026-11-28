<template>
  <Dialog
    :is-show="isShow"
    :title="$t('schedulerWorkflow.editInterval')"
    :width="640"
    :close-icon="!saving"
    :esc-close="!saving"
    :quick-close="false"
    :before-close="canClose"
    render-directive="if"
    @closed="handleClose"
  >
    <Form ref="formRef" :model="form" :rules="rules" form-type="vertical">
      <Form.FormItem :label="$t('schedulerWorkflow.name')">
        <Input :model-value="workflow.workflow_name" disabled />
      </Form.FormItem>
      <Form.FormItem label="Workflow ID">
        <Input :model-value="workflow.workflow_id" disabled />
      </Form.FormItem>
      <Form.FormItem :label="$t('schedulerWorkflow.interval')" property="interval" required>
        <Input v-model="form.interval" :disabled="saving" />
        <div class="mt-[8px] text-[12px] text-[#63656E]">
          {{ $t('schedulerWorkflow.intervalHelp', { every: '@every 10m', cron: '0 */10 * * * *' }) }}
        </div>
      </Form.FormItem>
    </Form>
    <Alert theme="info" :title="$t('schedulerWorkflow.intervalEffect')" />
    <Alert v-if="failed" theme="danger" class="mt-[12px]" :title="$t('schedulerWorkflow.intervalUpdateFailed')" />
    <template #footer>
      <Button theme="primary" :loading="saving" :disabled="saving" class="mr-[8px]" @click="handleSave">
        {{ $t('action.save') }}
      </Button>
      <Button :disabled="saving" @click="handleClose">{{ $t('action.cancel') }}</Button>
    </template>
  </Dialog>
</template>

<script setup lang="ts">
import { Alert, Button, Dialog, Form, Input } from 'bkui-vue';
import { computed, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';

import type { ScheduleWorkflowListRespScheduleWorkflow } from '@/@types/schedule_workflow';
import { ScheduleWorkflowService } from '@/api/modules/schedule_workflow';

const isShow = defineModel<boolean>('isShow', { required: true });
const props = defineProps<{
  workflow: Pick<ScheduleWorkflowListRespScheduleWorkflow, 'workflow_id' | 'workflow_name' | 'interval'>;
}>();
const emit = defineEmits<{ updated: [] }>();
const { t } = useI18n();
const formRef = ref();
const form = reactive({ interval: '' });
const saving = ref(false);
const failed = ref(false);
const rules = computed(() => ({
  interval: [{
    required: true,
    validator: (value: string) => !!value.trim(),
    message: t('schedulerWorkflow.intervalRequired'),
    trigger: 'blur',
  }],
}));

const canClose = () => !saving.value;
const handleClose = () => {
  if (saving.value) return;
  isShow.value = false;
};
const handleSave = async () => {
  if (saving.value) return;
  saving.value = true;
  failed.value = false;
  try {
    form.interval = form.interval.trim();
    const valid = await formRef.value.validate().catch(() => false);
    if (!valid) return;
    await ScheduleWorkflowService.ScheduleWorkflowUpdateInterval({
      workflow_id: props.workflow.workflow_id,
      interval: form.interval,
    });
    isShow.value = false;
    emit('updated');
  } catch {
    failed.value = true;
  } finally {
    saving.value = false;
  }
};

watch(isShow, (visible) => {
  if (!visible) return;
  form.interval = props.workflow.interval;
  failed.value = false;
}, { immediate: true });
</script>
