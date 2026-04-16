<template>
  <div ref="contentRef">
    <VxeTable
      ref="xTableRef"
      :data="tableData"
      :size="'medium'"
      :border="true"
      :max-height="maxHeight"
      :scroll-y="{ enabled: true, gt: 20 }"
      :row-config="{ isHover: true, useKey: true }"
      round
    >
      <!-- 主机 IP -->
      <VxeColgroup align="center">
        <template #header>
          <span>{{ $t('components.installTable.hostIp') }}</span>
        </template>
        <VxeColumn
          field="bk_host_innerip"
          :title="$t('components.installTable.innerIPv4')"
          :min-width="150"
        >
          <template #default="{ row }">
            <Input v-model.trim="row.bk_host_innerip" :disabled="true" />
          </template>
        </VxeColumn>
        <VxeColumn
          field="bk_host_innerip_v6"
          :title="$t('components.installTable.innerIPv6')"
          :min-width="150"
        >
          <template #default="{ row }">
            <Input v-model.trim="row.bk_host_innerip_v6" :disabled="true" />
          </template>
        </VxeColumn>
      </VxeColgroup>

      <!-- 业务属性 -->
      <VxeColgroup :title="$t('components.installTable.bizProperty')" align="center">
        <VxeColumn
          field="bk_biz_id"
          :title="$t('components.installTable.bkBizId')"
          :min-width="150"
        >
          <template #default="{ row }">
            <Select
              v-model="row.bk_biz_id"
              :disabled="true"
              filterable
            >
              <Select.Option
                v-for="item in businessList"
                :key="item.bk_biz_id"
                :name="item.bk_biz_name"
                :id="item.bk_biz_id"
              >
                [{{ item.bk_biz_id }}] {{ item.bk_biz_name }}
              </Select.Option>
            </Select>
          </template>
        </VxeColumn>
      </VxeColgroup>

      <!-- 拓扑属性 -->
      <VxeColgroup :title="$t('components.installTable.topoProperty')" align="center">
        <VxeColumn
          field="bk_networkarea_name"
          :title="$t('components.installTable.networkArea')"
          :min-width="150"
        >
          <template #default="{ row }">
            <Input v-model.trim="row.bk_networkarea_name" :disabled="true" />
          </template>
        </VxeColumn>
        <VxeColumn
          field="bk_networkunit_id"
          :title="$t('components.installTable.networkUnit')"
          :min-width="150"
        >
          <template #header>
            <span class="mr-[5px]">{{ $t('components.installTable.networkUnit') }}</span>
            <span class="mx-[3px] text-[#FF5656]">*</span>
            <BatchEdit
              :title="$t('components.installTable.batchEditNetworkUnit')"
              type="select"
              :options="networkUnitBatchOptions"
              :disabled="!isSameNetworkArea || networkUnitLoading"
              :disabled-tip="$t('components.installTable.batchEditNetworkUnitDisabledTip')"
              @confirm="(value: string) => handleBatchEdit(value)"
            />
            <i
              v-if="!networkUnitLoading && !autoAssignLoading"
              class="nodeman-icon nc-manual text-[18px] cursor-pointer ml-[5px]"
              v-bk-tooltips="$t('components.installTable.autoAssign')"
              @click="handleAutoAssign"
            ></i>
            <i
              v-else
              class="nodeman-icon nc-manual text-[18px] cursor-not-allowed text-[#C4C6CC] ml-[5px]"
              v-bk-tooltips="$t('components.installTable.autoAssign')"
            ></i>
          </template>
          <template #default="{ row, rowIndex }">
            <ValidateCell :error="getError(rowIndex, 'bk_networkunit_id')">
              <Select
                v-if="!networkUnitLoading"
                v-model="row.bk_networkunit_id"
                auto-focus
                filterable
                @change="
                  (val: string) => {
                    clearError(rowIndex, 'bk_networkunit_id');
                    handleNetworkUnitChange(val, row);
                  }
                "
                @toggle="
                  (val: boolean) =>
                    !val &&
                    handleFieldBlur(rowIndex, 'bk_networkunit_id', row.bk_networkunit_id)
                "
              >
                <Select.Option
                  v-for="option in getNetworkUnitsByAreaId(row.bk_networkarea_id)"
                  :key="option.bk_networkunit_id"
                  :id="String(option.bk_networkunit_id)"
                  :name="option.bk_networkunit_name"
                >
                  [{{ option.bk_networkunit_id }}] {{ option.bk_networkunit_name }}
                </Select.Option>
              </Select>
              <div v-else class="h-[32px] w-full rounded-[2px] bg-[#F5F7FA]"></div>
            </ValidateCell>
          </template>
        </VxeColumn>
      </VxeColgroup>

      <!-- 操作 -->
      <VxeColumn :min-width="60" field="action" :title="$t('platform.nodeMan.operate')">
        <template #default="{ rowIndex }">
          <Button
            text
            :disabled="tableData?.length <= 1"
            @click="handleDelRow(rowIndex)"
          >
            <i class="nodeman-icon nc-minus"></i>
          </Button>
        </template>
      </VxeColumn>

      <template #empty>
        <slot></slot>
      </template>
    </VxeTable>
  </div>
</template>

<script lang="ts" setup>
import { Button, InfoBox, Input, Select } from 'bkui-vue';
import { groupBy } from 'lodash';
import { computed, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';

import { VxeColgroup, VxeColumn, VxeTable } from '@blueking/vxe-table';

import ValidateCell from './validateCell.vue';

import { TopoService } from '@/api/modules/topo';
import BatchEdit from '@/components/batch-edit.vue';
import useFullScreen from '@/composables/use-fullscreen';
import useTableErrors from '@/composables/use-table-errors';
import { useMainStore } from '@/stores/main';

const tableData = defineModel<any[]>('data');

defineProps({
  maxHeight: { type: Number, default: 640 },
});
const { t } = useI18n();
const { contentRef } = useFullScreen();
const xTableRef = ref();
const mainStore = useMainStore();
const businessList = computed(() => mainStore.businessList);

const { getError, setError, clearError, clearAllErrors, shiftErrors } = useTableErrors();

const doDelRow = (index: number) => {
  if (!Array.isArray(tableData.value)) return;
  if (tableData.value.length === 1) return;
  tableData.value.splice(index, 1);
  shiftErrors(index, -1);
};

const handleDelRow = (index: number) => {
  const row = tableData.value?.[index];
  const ip = row?.bk_host_innerip || row?.bk_host_innerip_v6 || '';
  InfoBox({
    title: t('platform.nodeMan.agentStatus.assignUnitConfirmDeleteRow'),
    subTitle: t('platform.nodeMan.agentStatus.assignUnitConfirmDeleteRowSub', { ip }),
    onConfirm: () => doDelRow(index),
  });
};

const handleFieldBlur = (rowIndex: number, field: string, value: any) => {
  if (field === 'bk_networkunit_id' && !value && value !== 0) {
    setError(rowIndex, field, t('validate.required'));
    return;
  }
  clearError(rowIndex, field);
};

const networkUnitList = ref<any[]>([]);
const networkUnitGroupMap = ref<Record<number, any[]>>({});
const networkUnitLoading = ref(false);

const getNetworkUnitList = async () => {
  networkUnitLoading.value = true;
  try {
    const res = await TopoService.NetworkUnitListBrief({
      exact_include_conditions: {
        bk_networkarea_id: tableData.value?.map((item: any) => Number(item.bk_networkarea_id)) || [],
      },
    }).catch(() => ({ total: 0, items: [] }));
    networkUnitList.value = res.items;
    networkUnitGroupMap.value = groupBy(res.items, 'bk_networkarea_id');
  } finally {
    networkUnitLoading.value = false;
  }
};

const getNetworkUnitsByAreaId = (bkNetworkAreaId: number) => networkUnitGroupMap.value[bkNetworkAreaId] || [];

const isSameNetworkArea = computed(() => {
  if (!tableData.value?.length) return false;
  const firstAreaId = Number(tableData.value[0].bk_networkarea_id);
  return tableData.value.every((item: any) => Number(item.bk_networkarea_id) === firstAreaId);
});

const networkUnitBatchOptions = computed(() => {
  if (!isSameNetworkArea.value || !tableData.value?.length) return [];
  const areaId = Number(tableData.value[0].bk_networkarea_id);
  return getNetworkUnitsByAreaId(areaId).map((unit: any) => ({
    id: String(unit.bk_networkunit_id),
    name: `[${unit.bk_networkunit_id}] ${unit.bk_networkunit_name}`,
  }));
});

const handleBatchEdit = (value: string) => {
  tableData.value?.forEach((item: any, index: number) => {
    item.bk_networkunit_id = value;
    const networkUnit = networkUnitList.value.find((unit: any) => String(unit.bk_networkunit_id) === value);
    if (networkUnit) {
      item.bk_networkunit_name = networkUnit.bk_networkunit_name;
    }
    clearError(index, 'bk_networkunit_id');
  });
};

const autoAssignLoading = ref(false);

const handleAutoAssign = async () => {
  if (!tableData.value?.length) return;
  
  autoAssignLoading.value = true;
  try {
    const items = tableData.value.map((row: any) => ({
      bk_networkarea_id: Number(row.bk_networkarea_id),
      ip: row.bk_host_innerip || row.bk_host_innerip_v6 || '',
    }));
    
    const res = await TopoService.NetworkUnitRecommendByNetworkSegment({ items });
    
    if (res?.items) {
      res.items.forEach((result: any, index: number) => {
        if (index >= tableData.value!.length) return;
        
        const row = tableData.value![index];
        if (result.bk_networkunit_id === -1) {
          // 清空该行的选择
          row.bk_networkunit_id = undefined;
          row.bk_networkunit_name = undefined;
        } else {
          // 回填推荐结果
          row.bk_networkunit_id = String(result.bk_networkunit_id);
          const networkUnit = networkUnitList.value.find(
            (unit: any) => unit.bk_networkunit_id === result.bk_networkunit_id
          );
          if (networkUnit) {
            row.bk_networkunit_name = networkUnit.bk_networkunit_name;
          }
        }
        clearError(index, 'bk_networkunit_id');
      });
    }
  } catch (error) {
    console.error('Auto assign failed:', error);
  } finally {
    autoAssignLoading.value = false;
  }
};

const handleNetworkUnitChange = (val: string, row: any) => {
  if (!val) return;
  const networkUnit = networkUnitList.value.find((unit: any) => String(unit.bk_networkunit_id) === val);
  if (networkUnit) {
    row.bk_networkunit_name = networkUnit.bk_networkunit_name;
  }
};

const tableValidate = async () => {
  const data = tableData.value;
  if (!Array.isArray(data) || !data.length) return true;

  clearAllErrors();

  let isValid = true;
  let firstErrorRowIndex = -1;

  for (let i = 0; i < data.length; i++) {
    if (!data[i].bk_networkunit_id) {
      setError(i, 'bk_networkunit_id', t('validate.required'));
      isValid = false;
      if (firstErrorRowIndex === -1) firstErrorRowIndex = i;
    }
  }

  if (!isValid && firstErrorRowIndex !== -1 && xTableRef.value) {
    await xTableRef.value.scrollToRow(data[firstErrorRowIndex]);
  }

  return isValid;
};

watch(
  () => tableData.value?.map((item: any) => Number(item.bk_networkarea_id)).join(',') || '',
  async (val: string) => {
    if (!val) {
      networkUnitList.value = [];
      networkUnitGroupMap.value = {};
      return;
    }
    await getNetworkUnitList();
  },
  { immediate: true },
);

defineExpose({ tableValidate });
</script>

<style lang="postcss" scoped>
::v-deep(.vxe-header--column) {
  font-weight: normal !important;
  font-size: 12px !important;
}
::v-deep(.vxe-body--column) {
  height: 56px !important;
  font-size: 12px !important;
}
</style>
