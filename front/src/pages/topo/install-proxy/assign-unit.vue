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
          :label="$t('platform.nodeMan.agentStatus.assignUnitHostInfo')"
          label-width="90"
          required
        >
          <Loading :loading="loading">
            <assign-unit-table
              ref="assignUnitTableRef"
              v-model:data="form.info"
              :max-height="520"
              :hide-network-unit="true"
            ></assign-unit-table>
          </Loading>
        </Form.FormItem>
        <Form.FormItem
          :label="$t('components.installTable.networkArea')"
          label-width="90"
        >
          <Input :model-value="networkAreaName" class="w-[488px]" disabled />
        </Form.FormItem>
        <Form.FormItem
          :label="$t('components.installTable.networkUnit')"
          label-width="90"
          required
        >
          <Select
            v-model="form.networkUnitId"
            class="w-[488px]"
            filterable
            :loading="networkUnitLoading"
            :clearable="true"
            @change="handleNetworkUnitChange"
          >
            <Select.Option
              v-for="option in networkUnitOptions"
              :key="option.bk_networkunit_id"
              :id="String(option.bk_networkunit_id)"
              :name="`[${option.bk_networkunit_id}] ${option.bk_networkunit_name}`"
              :disabled="option.is_direct"
              :class="{ 'unauthorized-unit-row': !isUnitAuthorized(option.bk_networkunit_id) }"
              @click="handleUnitOptionClick($event, option.bk_networkunit_id)"
              @mouseenter="handleUnitOptionMouseEnter($event, option.bk_networkunit_id)"
              @mousemove="handleUnitOptionMouseMove($event, option.bk_networkunit_id)"
              @mouseleave="handleUnitOptionMouseLeave()"
              v-bk-tooltips="{
                content: $t('topoManager.installProxy.form.tip'),
                disabled: !option.is_direct,
                boundary: 'parent',
                placement: 'left',
              }"
            />
          </Select>
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
import { Button, Form, InfoBox, Input, Loading, Message, Select, Sideslider } from 'bkui-vue';
import { computed, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';

import type { Host } from '@/@types/common';
import { NodeProxyService } from '@/api/modules/node_proxy';
import { TopoService } from '@/api/modules/topo';
import AssignUnitTable from '@/components/assign-unit-table.vue';
import useUnitAuth from '@/composables/use-unit-auth';

const { t } = useI18n();
const router = useRouter();

const {
  isUnitAuthorized,
  handleOptionMouseEnter: handleUnitOptionMouseEnter,
  handleOptionMouseMove: handleUnitOptionMouseMove,
  handleOptionMouseLeave: handleUnitOptionMouseLeave,
  handleOptionClick: handleUnitOptionClick,
} = useUnitAuth();

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

const isShow = computed({
  get: () => props.isShow,
  set: (val) => emit('update:isShow', val),
});

const loading = ref(false);
const submitting = ref(false);
const formRef = ref(null);
const assignUnitTableRef = ref<InstanceType<typeof AssignUnitTable> | null>(null);

const form = reactive({
  info: [] as any[],
  networkUnitId: '' as string,
});

// Network unit options for the form-level selector
const networkUnitList = ref<any[]>([]);
const networkUnitLoading = ref(false);

const networkUnitOptions = computed(() => {
  const areaIds = new Set(
    form.info.map((item: any) => Number(item.bk_networkarea_id)).filter((id: number) => !isNaN(id)),
  );
  return networkUnitList.value.filter((unit: any) => areaIds.has(unit.bk_networkarea_id));
});

const loadNetworkUnits = async () => {
  const areaIds = form.info.map((item: any) => Number(item.bk_networkarea_id)).filter((id: number) => !isNaN(id));
  if (areaIds.length === 0) {
    networkUnitList.value = [];
    return;
  }
  networkUnitLoading.value = true;
  try {
    const res = await TopoService.NetworkUnitListBrief({
      exact_include_conditions: { bk_networkarea_id: areaIds },
    }).catch(() => ({ total: 0, items: [] }));
    networkUnitList.value = res.items || [];
  } finally {
    networkUnitLoading.value = false;
  }
};

const networkAreaName = computed(() => {
  if (form.info.length === 0) return '';
  return form.info[0].bk_networkarea_name || '';
});

const handleNetworkUnitChange = () => {
  // Clear validation errors when user selects a unit
};

// Load table data
watch(() => props.isShow, async (show) => {
  if (!show) return;

  form.networkUnitId = '';
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

    await loadNetworkUnits();
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
  if (!form.networkUnitId) {
    Message({ theme: 'warning', message: t('installProxy.pleaseSelectNetworkUnit') });
    return;
  }

  const hostIds = form.info.map((row: any) => row.bk_host_id);
  if (hostIds.length === 0) return;

  const unitId = Number(form.networkUnitId);

  InfoBox({
    title: t('installProxy.confirmAssignUnit'),
    subTitle: t('installProxy.confirmAssignUnitSubTitle'),
    onConfirm: async () => {
      await executeAssign(hostIds, unitId);
    },
  });
};

// Execute assign operation
const executeAssign = async (hostIds: number[], networkUnitID: number) => {
  submitting.value = true;

  try {
    const result = await NodeProxyService.NodeProxyAssignUnit({
      bk_host_id: hostIds,
      bk_networkunit_id: networkUnitID,
    });

    const successCount = result.success_count || 0;
    const failedCount = result.failed_count || 0;

    if (successCount > 0) {
      isShow.value = false;
      window.dispatchEvent(new Event('proxy-assign-unit-success'));

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
      } else {
        Message({
          theme: 'warning',
          message: t('installProxy.assignUnitPartialSuccess', {
            success: successCount,
            failed: failedCount,
          }),
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
