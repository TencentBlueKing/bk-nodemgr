<template>
  <Loading :loading="workareaStore.loading">
    <Table
      class="mt-[16px] w-full"
      ref="tableRef"
      :data="workareaStore.list"
      :empty-text="$t('table.empty')"
      :pagination="workareaStore.pagination"
      :sort-config="sortConfig"
      :show-settings="isShowSetting"
      :settings="settings"
      :max-height="maxHeight"
      @setting-change="handleSettingChange"
      @column-filter="handleColumnFilter">
      <TableColumn type="checkbox" :width="60" :resizable="false" />
      <TableColumn
        :label="$t('topoManager.region.workarea.table.workareaName')"
        field="bk_networkarea_name"
        show-overflow="tooltip"
        :min-width="280">
        <template #default="{ row }">
          <Button theme="primary" class="!text-[12px]" text>{{ row.bk_networkarea_name }}</Button>
        </template>
      </TableColumn>
      <TableColumn
        :label="$t('topoManager.region.workarea.table.workareaId')"
        field="bk_networkarea_id"
        show-overflow="tooltip"
        :min-width="280">
        <template #default="{ row }">
          <span class="!text-[12px]">
            {{ row.bk_networkarea_id ? `#${row.bk_networkarea_id}` : '--' }}
          </span>
        </template>
      </TableColumn>
      <TableColumn
        :label="$t('topoManager.region.workarea.table.vendor')"
        field="bk_cloud_vendor"
        show-overflow="tooltip"
        :filter="filterOption"
        :min-width="280">
        <template #default="{ row }">
          <span class="!text-[12px]">
            {{ row.bk_cloud_vendor || '--' }}
          </span>
        </template>
      </TableColumn>
      <TableColumn
        :label="$t('topoManager.region.workarea.table.workUnitsCount')"
        field="areaCount"
        show-overflow="tooltip"
        sortable
        :min-width="240">
        <template #default="{ row }">
          <span class="!text-[12px]">
            {{ row.areaCount || '--' }}
          </span>
        </template>
      </TableColumn>
      <TableColumn
        :label="$t('topoManager.region.workarea.table.nodesCount')"
        field="unitCount"
        show-overflow="tooltip"
        sortable
        :min-width="130">
        <template #default="{ row }">
          <span class="!text-[12px]">
            {{ row.unitCount || '--' }}
          </span>
        </template>
      </TableColumn>
      <TableColumn
        :label="$t('table.action')"
        field="action"
        show-overflow="tooltip"
        :min-width="140">
        <template #default="{ row }">
          <div class="flex">
            <Button theme="primary" text class="mr-[12px]" @click="handleEditWorkarea">
              {{ $t('action.edit') }}
            </Button>
            <Button theme="primary" text @click="handlehandleDeleteWorkarea(row.bk_networkarea_id)">
              {{ $t('action.delete') }}
            </Button>
          </div>
        </template>
      </TableColumn>
    </Table>
  </Loading>
</template>

<script lang="ts" setup>
import { Button, InfoBox, Loading } from 'bkui-vue';
import { onMounted, reactive, ref } from 'vue';
import { useI18n } from 'vue-i18n';

import { Table, TableColumn } from '@blueking/table';

import useDynamicsHeight from '@/composables/use-table-height';
import useTableSetting from '@/composables/use-table-setting';
import { useWorkareaStore } from '@/stores/workarea';

const { t } = useI18n();
const workareaStore = useWorkareaStore();
const sortConfig = ref({ multiple: true });

// table setting逻辑
const { isShowSetting, settings, handleSettingChange } = useTableSetting({
  checked: [
    'bk_networkarea_name',
    'bk_networkarea_id',
    'bk_cloud_vendor',
    'areaCount',
    'unitCount',
    'action',
  ],
  disabled: ['action'],
});

// table filter逻辑
const vendorMap = {
  tencent: t('topoManager.region.workarea.vendor.tencent'),
  google: t('topoManager.region.workarea.vendor.google'),
  huawei: t('topoManager.region.workarea.vendor.huawei'),
  microsoft: t('topoManager.region.workarea.vendor.microsoft'),
  aws: 'AWS',
  ali: t('topoManager.region.workarea.vendor.ali'),
};
const filterOption = reactive({
  list: Object.entries(vendorMap).map(item => ({
    value: item[0],
    text: item[1],
  })),
  checked: [] as string[],
});

const handleColumnFilter = ({ checked }: { checked: string[] }) => {
  filterOption.checked = checked;
  handleFilter(checked.map(item => parseInt(item)));
};

// 改变includeConditions 重新请求table data
const handleFilter = (currentChecked: number[]) => {
  workareaStore.includeConditions.bk_cloud_vendor = currentChecked;
  workareaStore.handleFetchWorkareaList();
};

// table height逻辑
// 52(顶部导航)+52(二级导航)+24*2(content-padding)+32(flex-row)+16(table-margin-top)
const tableOffset = 200;
const { maxHeight } = useDynamicsHeight(tableOffset);

const handleEditWorkarea = () => {
  // todo 打开管控区域详情侧栏
};

// todo
// 可能需要补充交互(message/重置筛选/重置pagination)
const handlehandleDeleteWorkarea = (bk_networkarea_id: number) => {
  InfoBox({
    title: t('topoManager.region.workarea.delete.title'),
    cancelText: t('action.cancel'),
    onConfirm() {
      workareaStore.handleDeleteWorkarea(bk_networkarea_id);
      workareaStore.handleFetchWorkareaList();
    },
  });
};

onMounted(() => {
  workareaStore.handleFetchWorkareaList();
});

const tableRef = ref();
defineExpose({
  tableRef,
});

</script>
