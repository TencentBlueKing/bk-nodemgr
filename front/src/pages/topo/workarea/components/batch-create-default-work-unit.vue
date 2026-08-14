<template>
  <Sideslider
    v-model:is-show="isShow"
    :title="$t('topoManager.workArea.batchCreate.title')"
    :before-close="handleBeforeClose"
    width="640"
  >
    <Loading :loading="loading">
      <template v-if="!isResultMode">
        <div class="mt-[24px]">
          <div class="mb-[8px] pl-[16px] text-[14px] text-[#313238]">
            {{ $t('topoManager.workArea.batchCreate.selectedAreas', { count: selectedAreas.length }) }}
          </div>
          <Table
            :data="selectedAreas"
            :pagination="pagination"
            :empty-text="$t('table.empty')"
            @page-limit-change="pageLimitChange"
            @page-value-change="pageValueChange"
          >
            <TableColumn
              :label="$t('topoManager.workArea.table.workareaId')"
              field="bk_networkarea_id"
              width="140"
            />
            <TableColumn
              :label="$t('topoManager.workArea.table.workareaName')"
              field="bk_networkarea_name"
            />
          </Table>
        </div>
        <Form
          ref="formRef"
          class="pt-[28px]"
          :model="form"
          :rules="rules"
        >
          <Form.FormItem
            :label="$t('topoManager.workUnit.form.workUnitName')"
            property="bk_networkunit_name"
            label-width="130"
            required
          >
            <Input
              v-model="form.bk_networkunit_name"
              class="w-[488px]"
              clearable
            />
          </Form.FormItem>
          <Form.FormItem
            :label="$t('topoManager.workUnit.form.upstream')"
            property="upstream"
            label-width="130"
            required
          >
            <SelectGroup
              v-model:link="form.upstream"
              :is-create="true"
              :work-unit-id="-1"
            />
          </Form.FormItem>
          <div class="ml-[130px] w-[488px] text-[12px] text-[#979BA5]">
            {{ $t('topoManager.workArea.batchCreate.description') }}
          </div>
        </Form>
      </template>
      <template v-else>
        <div class="mt-[24px] text-[14px] text-[#313238]">
          {{ $t('topoManager.workArea.batchCreate.resultSummary', {
            success: result?.success_count || 0,
            failed: result?.failed_count || 0,
          }) }}
        </div>
        <div class="mt-[16px] max-h-[420px] overflow-auto">
          <div
            v-for="item in resultItems"
            :key="item.bk_networkarea_id"
            class="mb-[8px] rounded-[2px] border border-[#DCDEE5] px-[16px] py-[12px]"
          >
            <div class="flex items-center justify-between">
              <span class="text-[13px] text-[#313238]">
                #{{ item.bk_networkarea_id }}
              </span>
              <Tag
                :theme="item.success ? 'success' : 'danger'"
                size="small"
              >
                {{
                  item.success
                    ? $t('topoManager.workArea.batchCreate.success')
                    : $t('topoManager.workArea.batchCreate.failed')
                }}
              </Tag>
            </div>
            <div
              v-if="!item.success"
              class="mt-[6px] text-[12px] text-[#EA3636]"
            >
              {{ getResultMessage(item) }}
            </div>
            <div
              v-else
              class="mt-[6px] text-[12px] text-[#63656E]"
            >
              {{ $t('topoManager.workArea.batchCreate.createdUnit', {
                id: item.bk_networkunit_id,
              }) }}
            </div>
          </div>
        </div>
      </template>
    </Loading>
    <template #footer>
      <div class="flex">
        <Button
          v-if="isResultMode"
          theme="primary"
          class="mr-[8px] w-[120px]"
          :loading="submitLoading"
          :disabled="selectedAreas.length === 0"
          @click="handleConfirm"
        >
          {{ $t('topoManager.workArea.batchCreate.retry') }}
        </Button>
        <Button
          v-else
          theme="primary"
          class="mr-[8px] w-[88px]"
          :loading="submitLoading"
          @click="handleConfirm"
        >
          {{ $t('action.save') }}
        </Button>
        <Button class="w-[88px]" @click="handleBeforeClose">
          {{ $t('action.cancel') }}
        </Button>
      </div>
    </template>
  </Sideslider>
</template>

<script setup lang="ts">
import { Button, Form, InfoBox, Input, Loading, Message, Sideslider, Tag } from 'bkui-vue';
import { cloneDeep, isEqual } from 'lodash';
import { computed, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';

import { Table, TableColumn } from '@blueking/table';

import SelectGroup from '../../upsert-workunit/components/select-group.vue';

import type {
  TopoNetworkUnitCreateDefaultMultiReq,
  TopoNetworkUnitCreateDefaultMultiResp,
  TopoNetworkUnitCreateDefaultMultiRespResult,
} from '@/@types/topo';
import { TopoService } from '@/api/modules/topo';
import usePage from '@/composables/use-page';
import type { INetWorkArea } from '@/stores/workarea';
import { useWorkareaStore } from '@/stores/workarea';

const isShow = defineModel<boolean>('isShow', { required: true });

const props = defineProps<{
  selectedAreas: INetWorkArea[];
}>();

const emit = defineEmits<{
  submitted: [result: TopoNetworkUnitCreateDefaultMultiResp['data']];
}>();

const { t } = useI18n();
const workareaStore = useWorkareaStore();
const loading = ref(false);
const submitLoading = ref(false);
const formRef = ref();
const result = ref<TopoNetworkUnitCreateDefaultMultiResp['data']>();
const resultItems = ref<TopoNetworkUnitCreateDefaultMultiRespResult[]>([]);

const form = reactive<{
  bk_networkunit_name: string;
  upstream: Partial<Link>;
}>({
  bk_networkunit_name: '',
  upstream: {
    bk_networkarea_id: undefined,
    bk_networkunit_id: undefined,
    accesspoint_id: undefined,
  },
});

const originForm = ref(cloneDeep(form));
const isResultMode = computed(() => resultItems.value.length > 0);

const selectedAreaList = computed(() => props.selectedAreas);
const { pagination, pageConf } = usePage(selectedAreaList);

const pageLimitChange = (limit: number) => {
  pageConf.limit = limit;
  pageConf.current = 1;
};

const pageValueChange = (current: number) => {
  pageConf.current = current;
};

const validateUpstream = () => {
  const { bk_networkarea_id, bk_networkunit_id, accesspoint_id } = form.upstream;
  return (
    Number(bk_networkarea_id) >= 0
    && Number(bk_networkunit_id) >= 0
    && Number(accesspoint_id) >= 0
  );
};

const rules = reactive({
  bk_networkunit_name: [{
    required: true,
    trigger: 'blur',
  }],
  upstream: [{
    required: true,
    trigger: 'blur',
    validator: validateUpstream,
  }],
});

const resetForm = () => {
  form.bk_networkunit_name = '';
  form.upstream = {
    bk_networkarea_id: undefined,
    bk_networkunit_id: undefined,
    accesspoint_id: undefined,
  };
  result.value = undefined;
  resultItems.value = [];
  originForm.value = cloneDeep(form);
};

const fetchUpstreamOptions = async () => {
  loading.value = true;
  try {
    await Promise.all([
      workareaStore.handleFetchAllWorkarea(),
      workareaStore.handleFetchAllWorkUnit(),
    ]);
  } finally {
    loading.value = false;
  }
};

const getResultMessage = (item: TopoNetworkUnitCreateDefaultMultiRespResult) => (
  item.error_code
    ? t(`topoManager.workArea.batchCreate.errors.${item.error_code}`)
    : item.message
);

const handleConfirm = async () => {
  const valid = await formRef.value?.validate().catch(() => false);
  if (!valid && !isResultMode.value) return;
  if (props.selectedAreas.length === 0) return;

  submitLoading.value = true;
  try {
    const params: TopoNetworkUnitCreateDefaultMultiReq = {
      bk_networkarea_id: props.selectedAreas.map(item => item.bk_networkarea_id),
      bk_networkunit_name: form.bk_networkunit_name,
      upstream: {
        bk_networkarea_id: form.upstream.bk_networkarea_id as number,
        bk_networkunit_id: form.upstream.bk_networkunit_id as number,
        accesspoint_id: form.upstream.accesspoint_id as number,
      },
    };
    const response = await TopoService.NetworkUnitCreateDefaultMulti(params);
    result.value = response;
    resultItems.value = response.items || [];
    emit('submitted', response);

    if (response.failed_count === 0) {
      Message({
        theme: 'success',
        message: t('topoManager.workArea.batchCreate.allSuccess'),
      });
      isShow.value = false;
    }
  } catch {
    // 错误提示由全局响应拦截器统一弹出，这里不再重复弹窗。
  } finally {
    submitLoading.value = false;
  }
};

const handleBeforeClose = (): Promise<boolean> => new Promise((resolve) => {
  if (isResultMode.value || isEqual(form, originForm.value)) {
    resolve(true);
    isShow.value = false;
    return;
  }

  InfoBox({
    title: t('dialog.confirmClose'),
    infoType: 'warning',
    onConfirm: () => {
      resolve(true);
      isShow.value = false;
    },
    onCancel: () => resolve(false),
    onClose: () => resolve(false),
  });
});

watch(isShow, async (visible) => {
  if (!visible) {
    formRef.value?.clearValidate();
    resetForm();
    return;
  }

  resetForm();
  await fetchUpstreamOptions();
});
</script>
