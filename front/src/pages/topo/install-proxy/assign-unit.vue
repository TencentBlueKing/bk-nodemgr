<template>
  <Sideslider
    v-model:is-show="isShow"
    :title="$t('installProxy.assignUnit')"
    width="1200"
    render-directive="if"
    :before-close="handleBeforeClose"
  >
    <div class="py-[20px] px-[40px]">
      <Form ref="formRef" :model="form" class="mt-[24px]">
        <Form.FormItem
          :label="$t('platform.nodeMan.agentNodeStatus.assignUnitHostInfo')"
          label-width="90"
          required
        >
          <Loading :loading="loading">
            <assign-unit-table
              ref="assignUnitTableRef"
              v-model:data="form.info"
              :max-height="520"
              :auth-action="'networkunit_use_for_proxy'"
            ></assign-unit-table>
          </Loading>
        </Form.FormItem>
      </Form>
      <div class="flex mt-[32px] ml-[90px]">
        <Button
          theme="primary"
          class="mr-[8px] w-[120px]"
          :loading="submitting"
          @click="handleConfirm"
        >
          <span>{{ $t('installProxy.goAssign') }}</span>
          <span
            class="ml-[8px] px-[6px] bg-[#e1ecff] rounded-[8px] text-[#3a84ff] text-[12px] h-[16px] leading-[16px]"
          >
            {{ form.info.length }}
          </span>
        </Button>
        <Button @click="handleBeforeClose">
          {{ $t('action.cancel') }}
        </Button>
      </div>
    </div>
  </Sideslider>
</template>

<script setup lang="ts">
import { Button, Form, InfoBox, Loading, Message, Sideslider } from 'bkui-vue';
import { computed, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';

import { NodeProxyService } from '@/api/modules/node_proxy';
import { TopoService } from '@/api/modules/topo';
import AssignUnitTable from '@/components/assign-unit-table.vue';

interface Props {
  isShow: boolean;
  data: Host[];
  isCrossPageSelection?: boolean;
  params?: any;
}

const props = withDefaults(defineProps<Props>(), {
  isShow: false,
  data: () => [],
  isCrossPageSelection: false,
  params: () => ({}),
});
const emit = defineEmits(['update:isShow']);
const { t } = useI18n();
const router = useRouter();

const isShow = computed({
  get: () => props.isShow,
  set: val => emit('update:isShow', val),
});

const loading = ref(false);
const submitting = ref(false);
const formRef = ref(null);
const assignUnitTableRef = ref<InstanceType<typeof AssignUnitTable> | null>(null);

const form = reactive({
  info: [] as any[],
});

// Load table data
watch(() => props.isShow, async (show) => {
  if (!show) return;

  loading.value = true;
  try {
    if (props.isCrossPageSelection) {
      // Cross-page selection: fetch all hosts via HostList
      const allHosts: any[] = [];
      let offset = 0;
      const pageSize = 1000;
      let hasMore = true;

      while (hasMore) {
        const hostListData = await TopoService.HostList({
          page: { offset, limit: pageSize },
          only_count: false,
          ...props.params,
        }).catch(() => ({ total: 0, items: [] }));

        if (hostListData.items && hostListData.items.length > 0) {
          allHosts.push(...hostListData.items);
          offset += pageSize;
          hasMore = hostListData.items.length === pageSize;
        } else {
          hasMore = false;
        }
      }

      form.info = allHosts.map((host: any) => ({
        ...host.state,
        ...host.info,
        ...host,
        bk_host_innerip: host.info?.bk_host_innerip_list?.join(',') || '',
        bk_host_innerip_v6: host.info?.bk_host_innerip_v6_list?.join(',') || '',
      }));
    } else {
      // Direct selection: props.data is already flattened by DetailTable
      form.info = props.data.map(host => ({
        ...host,
      }));
    }
  } catch (error) {
    console.error('Load proxy data error:', error);
    Message({
      theme: 'error',
      message: t('common.loadDataFailed'),
    });
  } finally {
    loading.value = false;
  }
});

// Handle before close
const handleBeforeClose = (): Promise<boolean> => new Promise((resolve, reject) => {
  if (submitting.value) {
    reject();
    return;
  }
  if (form.info.length === 0) {
    isShow.value = false;
    resolve(true);
    return;
  }
  InfoBox({
    title: t('installProxy.confirmClose'),
    infoType: 'warning',
    onConfirm: () => {
      isShow.value = false;
      resolve(true);
    },
    onCancel: () => {
      reject();
    },
  });
});

// Handle confirm
const handleConfirm = async () => {
  const valid = await assignUnitTableRef.value?.tableValidate();
  if (!valid) return;

  const grouped = new Map<number, number[]>();
  for (const row of form.info) {
    const networkUnitID = Number(row.bk_networkunit_id);
    if (!Number.isInteger(networkUnitID) || networkUnitID < 0) continue;
    if (!grouped.has(networkUnitID)) grouped.set(networkUnitID, []);
    grouped.get(networkUnitID)!.push(row.bk_host_id);
  }
  if (grouped.size === 0) return;

  const items = Array.from(grouped, ([networkUnitID, hostIDs]) => ({
    bk_host_id: hostIDs,
    bk_networkunit_id: networkUnitID,
  }));

  InfoBox({
    title: t('installProxy.confirmAssignUnit'),
    subTitle: t('installProxy.confirmAssignUnitSubTitle'),
    onConfirm: async () => {
      await executeAssign(items);
    },
  });
};

// Execute assign operation
const executeAssign = async (items: { bk_host_id: number[]; bk_networkunit_id: number }[]) => {
  submitting.value = true;

  try {
    const result = await NodeProxyService.NodeProxyAssignUnitMulti({ items });

    const successCount = result.success_count || 0;
    const failedCount = result.failed_count || 0;

    if (successCount > 0) {
      isShow.value = false;
      window.dispatchEvent(new Event('proxy-assign-unit-success'));

      if (failedCount > 0) {
        Message({
          theme: 'warning',
          message: result.failed_reasons?.join('; ') || t('installProxy.assignUnitPartialSuccess', {
            success: successCount,
            failed: failedCount,
          }),
        });
      }

      if (result.workflow_id) {
        router.push({
          name: 'taskDetail',
          params: { taskId: result.workflow_id, routerBackName: 'taskList' },
          query: { active: 'node' },
        });
      } else if (failedCount === 0) {
        Message({
          theme: 'success',
          message: t('installProxy.assignUnitSuccess', { count: successCount }),
        });
      }
    } else {
      Message({
        theme: 'error',
        message: t('installProxy.assignUnitFailed'),
      });
      if (result.failed_reasons?.length) {
        console.error('Failed reasons:', result.failed_reasons);
      }
    }
  } catch (error) {
    Message({
      theme: 'error',
      message: t('installProxy.assignUnitFailed'),
    });
    console.error('Assign unit error:', error);
  } finally {
    submitting.value = false;
  }
};
</script>
