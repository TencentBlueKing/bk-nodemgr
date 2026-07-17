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
      :edit-config="{ trigger: 'click', mode: 'row', showIcon: false }"
      round
    >
      <!-- 业务属性 -->
      <VxeColgroup :title="$t('components.installTable.bizProperty')" align="center">
        <VxeColumn
          field="bk_biz_id"
          :title="$t('components.installTable.bkBizId')"
          :min-width="150"
        >
          <template #default="{ row }">
            <!-- 归属业务始终不可编辑 -->
            <div class="cell-disabled">
              {{ row.bk_biz_id ? `[${row.bk_biz_id}] ${businessList.find((b: any) => b.bk_biz_id === row.bk_biz_id)?.bk_biz_name || row.bk_biz_id}` : '' }}
            </div>
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
            <!-- 管控区域始终不可编辑 -->
            <div class="cell-disabled">{{ row.bk_networkarea_name }}</div>
          </template>
        </VxeColumn>
        <VxeColumn
          v-if="!hideNetworkUnit"
          field="bk_networkunit_id"
          :title="$t('components.installTable.networkUnit')"
          :width="220"
          :edit-render="{ name: 'VxeInput' }"
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
              v-bk-tooltips="$t('components.installTable.autoAssignTooltip')"
              @click="handleAutoAssign"
            ></i>
            <i
              v-else
              class="nodeman-icon nc-manual text-[18px] cursor-not-allowed text-[#C4C6CC] ml-[5px]"
              v-bk-tooltips="$t('components.installTable.autoAssignTooltip')"
            ></i>
          </template>
          <template #default="{ row, rowIndex }">
            <ValidateCell :error="getError(rowIndex, 'bk_networkunit_id')">
              <div :class="{ 'cell-disabled--error': getError(rowIndex, 'bk_networkunit_id') }">
                <span v-if="getNetworkUnitName(row.bk_networkunit_id)">{{ getNetworkUnitName(row.bk_networkunit_id) }}</span>
                <span v-else class="cell-placeholder">{{ $t('platform.nodeMan.agentNodeStatus.assignUnitSelectPlaceholder') }}</span>
              </div>
            </ValidateCell>
          </template>
          <template #edit="{ row, rowIndex }">
            <ValidateCell :error="getError(rowIndex, 'bk_networkunit_id')">
              <Select
                v-if="!networkUnitLoading"
                class="w-[220px]"
                v-model="row.bk_networkunit_id"
                auto-focus
                filterable
                transfer
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
                  :disabled="disableDirect && option.is_direct"
                  v-bk-tooltips="{
                    content: $t('topoManager.installProxy.form.tip'),
                    disabled: !(disableDirect && option.is_direct),
                    boundary: 'parent',
                    placement: 'left',
                  }"
                >
                  <div
                    class="w-[220px] h-[32px] flex items-center -mx-[12px] px-[12px]"
                    :class="{ 'unauthorized-unit-row': !isUnitAuthorized(option.bk_networkunit_id) }"
                    @click="handleUnitOptionClick($event, option.bk_networkunit_id)"
                    @mouseenter="handleUnitOptionMouseEnter($event, option.bk_networkunit_id)"
                    @mousemove="handleUnitOptionMouseMove($event, option.bk_networkunit_id)"
                    @mouseleave="handleUnitOptionMouseLeave()"
                  >
                    [{{ option.bk_networkunit_id }}] {{ option.bk_networkunit_name }}
                  </div>
                </Select.Option>
              </Select>
              <div v-else class="h-[32px] w-full rounded-[2px] bg-[#F5F7FA]"></div>
            </ValidateCell>
          </template>
        </VxeColumn>
      </VxeColgroup>

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
            <!-- 内网IPv4不可编辑 -->
            <div class="cell-disabled">{{ row.bk_host_innerip }}</div>
          </template>
        </VxeColumn>
        <VxeColumn
          field="bk_host_innerip_v6"
          :title="$t('components.installTable.innerIPv6')"
          :min-width="150"
        >
          <template #default="{ row }">
            <!-- 内网IPv6不可编辑 -->
            <div class="cell-disabled">{{ row.bk_host_innerip_v6 }}</div>
          </template>
        </VxeColumn>
      </VxeColgroup>

      <!-- 主机属性 -->
      <VxeColgroup :title="$t('components.installTable.hostAttr')" align="center">
        <VxeColumn
          field="os_type"
          :title="$t('components.installTable.osType')"
          :min-width="120"
        >
          <template #default="{ row }">
            <!-- 操作系统不可编辑 -->
            <div class="cell-disabled">{{ row.os_type }}</div>
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
import { Button, InfoBox, Select } from 'bkui-vue';
import { groupBy } from 'lodash';
import { computed, onMounted, onBeforeUnmount, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';

import { VxeColgroup, VxeColumn, VxeTable } from '@blueking/vxe-table';

import ValidateCell from './validateCell.vue';

import { TopoService } from '@/api/modules/topo';
import { getFirstIp } from '@/common/util';
import BatchEdit from '@/components/batch-edit.vue';
import useFullScreen from '@/composables/use-fullscreen';
import useTableErrors from '@/composables/use-table-errors';
import useUnitAuth from '@/composables/use-unit-auth';
import { useMainStore } from '@/stores/main';

const tableData = defineModel<any[]>('data');

const props = defineProps({
  maxHeight: { type: Number, default: 640 },
  hideNetworkUnit: { type: Boolean, default: false },
  disableDirect: { type: Boolean, default: true },
});
const { t } = useI18n();
const { contentRef } = useFullScreen();
const xTableRef = ref();

// ===== 阻止 vxe-table 因 Select popover 点击而退出编辑（与 install-table 一致）=====
// transfer 将下拉弹窗挂到 body，vxe-table 视为"行外点击" → 提前退出编辑 → @change 丢失
// 在捕获阶段拦截 mousedown：若目标在 .bk-select-dropdown 内则阻止 vxe-table 退出编辑
const handleCaptureMouseDown = (e: MouseEvent) => {
  const target = e.target as HTMLElement;
  if (target?.closest?.('.bk-select-dropdown')) {
    e.stopImmediatePropagation();
  }
};

onMounted(() => {
  document.addEventListener('mousedown', handleCaptureMouseDown, true);
});

onBeforeUnmount(() => {
  document.removeEventListener('mousedown', handleCaptureMouseDown, true);
});
const mainStore = useMainStore();
const {
  isUnitAuthorized,
  handleOptionMouseEnter: handleUnitOptionMouseEnter,
  handleOptionMouseMove: handleUnitOptionMouseMove,
  handleOptionMouseLeave: handleUnitOptionMouseLeave,
  handleOptionClick: handleUnitOptionClick,
} = useUnitAuth('networkunit_use_for_agent');
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
    title: t('platform.nodeMan.agentNodeStatus.assignUnitConfirmDeleteRow'),
    subTitle: t('platform.nodeMan.agentNodeStatus.assignUnitConfirmDeleteRowSub', { ip }),
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
const autoAssignLoading = ref(false);

const handleAutoAssign = async () => {
  if (!tableData.value?.length) return;
  autoAssignLoading.value = true;
  try {
    const items = tableData.value.map((row: any) => ({
      bk_networkarea_id: Number(row.bk_networkarea_id),
      ip: getFirstIp(row.bk_host_innerip) || getFirstIp(row.bk_host_innerip_v6),
    }));
    const res = await TopoService.RecommendNetworkUnitByNetworkSegment({ items });
    if (res?.items) {
      res.items.forEach((result: any, index: number) => {
        if (index >= tableData.value!.length) return;
        const row = tableData.value![index];
        row.bk_networkunit_id = result.bk_networkunit_id === -1 ? '' : String(result.bk_networkunit_id);
        clearError(index, 'bk_networkunit_id');
      });
    }
  } catch (error) {
    console.error('Auto assign failed:', error);
  } finally {
    autoAssignLoading.value = false;
  }
};

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

const getNetworkUnitsByAreaId = (bkNetworkAreaId: number | string) => networkUnitGroupMap.value[Number(bkNetworkAreaId)] || [];
const getNetworkUnitName = (networkUnitId: string | number) => networkUnitList.value.find((u: any) => String(u.bk_networkunit_id) === String(networkUnitId))?.bk_networkunit_name || '';

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
    disabled: props.disableDirect && unit.is_direct,
    disabledTip: props.disableDirect ? t('topoManager.installProxy.form.tip') : '',
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
.cell-placeholder {
  color: #c4c6cc;
}
/* 不可编辑单元格公共样式，与 install-table 保持一致 */
.cell-disabled,
.cell-disabled--error {
  display: flex;
  align-items: center;
  width: 100%;
  height: 32px;
  padding: 0 10px;
  border: 1px solid transparent;
  border-radius: 2px;
  font-size: 12px;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}
/* 禁用态（无错误） */
.cell-disabled {
  background-color: #f5f7fa;
  color: #c4c6cc;
  cursor: not-allowed;
}
/* 校验失败态 */
.cell-disabled--error {
  background-color: #fff0f0;
  border-color: #ea3636;
  color: #ea3636;
}
</style>
