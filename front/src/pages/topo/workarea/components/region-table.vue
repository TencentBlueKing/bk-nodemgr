<template>
  <Loading :loading="loading">
    <Table
      class="mt-[16px] w-full"
      ref="tableRef"
      :data="tableData"
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
        :label="$t('topoManager.workArea.table.workareaName')"
        field="bk_networkarea_name"
        show-overflow="tooltip"
        :min-width="280">
        <template #default="{ row }">
          <Button theme="primary" class="!text-[12px]" text @click="handleToWorkareaDetail(row.bk_networkarea_id)">
            {{ row.bk_networkarea_name }}
          </Button>
        </template>
      </TableColumn>
      <TableColumn
        :label="$t('topoManager.workArea.table.workareaId')"
        field="bk_networkarea_id"
        show-overflow="tooltip"
        :min-width="240">
        <template #default="{ row }">
          <span class="!text-[12px]">
            {{ row.bk_networkarea_id ? `#${row.bk_networkarea_id}` : '--' }}
          </span>
        </template>
      </TableColumn>
      <TableColumn
        :label="$t('topoManager.workArea.table.vendor')"
        field="cloud_vendor"
        show-overflow="tooltip"
        :filter="filterOption"
        :min-width="240">
        <template #default="{ row }">
          <span class="!text-[12px]">
            {{ props.vendorList[row.cloud_vendor] ?
              $t(
                vendorMap[props.vendorList[row.cloud_vendor]]?.label
              ) : '--'
            }}
          </span>
        </template>
      </TableColumn>
      <TableColumn
        :label="$t('topoManager.workArea.table.workUnitsCount')"
        field="networkunit_count"
        show-overflow="tooltip"
        sortable
        :min-width="240">
        <template #default="{ row }">
          <span class="!text-[12px]">
            {{ !isNaN(row.networkunit_count) ? row.networkunit_count : '--' }}
          </span>
        </template>
      </TableColumn>
      <TableColumn
        :label="$t('topoManager.workArea.table.proxyCount')"
        field="proxy_count"
        show-overflow="tooltip"
        sortable
        :min-width="180">
        <template #default="{ row }">
          <span class="!text-[12px]">
            {{ !isNaN(row.proxy_count) ? row.proxy_count : '--' }}
          </span>
        </template>
      </TableColumn>
      <TableColumn
        :label="$t('topoManager.workArea.table.agentCount')"
        field="agent_count"
        show-overflow="tooltip"
        sortable
        :min-width="180">
        <template #default="{ row }">
          <span class="!text-[12px]">
            {{ !isNaN(row.agent_count) ? row.agent_count : '--' }}
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
            <Button theme="primary" text class="mr-[12px]" @click="handleEditWorkarea(row)">
              {{ $t('action.edit') }}
            </Button>
            <Button theme="primary" text @click="handleDeleteWorkarea(row.bk_networkarea_id)">
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
import { reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';

import { Table, TableColumn } from '@blueking/table';

import { vendorMap } from '../vendorMap';

import useDynamicsHeight from '@/composables/use-table-height';
import useTableSetting from '@/composables/use-table-setting';
import type { INetWorkArea } from '@/stores/workarea';
import { useWorkareaStore } from '@/stores/workarea';

interface IProps {
  list: INetWorkArea[],
  vendorList: string[]
}
const props = defineProps<IProps>();

const emit = defineEmits(['edit']);

const tableData = ref(props.list.sort((a: INetWorkArea,b: INetWorkArea) => {
  return a.bk_networkarea_id - b.bk_networkarea_id
}));
const { t } = useI18n();
const router = useRouter();
const workareaStore = useWorkareaStore();
const sortConfig = ref({ multiple: true });

// table setting逻辑
const { isShowSetting, settings, handleSettingChange } = useTableSetting({
  checked: [
    'bk_networkarea_name',
    'bk_networkarea_id',
    'cloud_vendor',
    'networkunit_count',
    'proxy_count',
    'agent_count',
    'action',
  ],
  disabled: ['action'],
});

// table filter逻辑
const filterOption = reactive<{
  list: {
    value: any
    text: string
  }[],
  checked: string[]
}>({
  list: [],
  checked: [],
});
const loading = ref(true);
const handleColumnFilter = ({ checked }: { checked: string[] }) => {
  filterOption.checked = checked;
  handleFilter(checked);
};

// 改变includeConditions 重新请求table data
const handleFilter = (currentChecked: string[]) => {
  if (currentChecked.length === 0) {
    tableData.value = props.list;
    return;
  }
  tableData.value = tableData.value.filter(item => currentChecked.includes(parseInt(item.cloud_vendor)));
};

// table height逻辑
// 52(顶部导航)+52(二级导航)+24*2(content-padding)+32(flex-row)+16(table-margin-top)
const tableOffset = 200;
const { maxHeight } = useDynamicsHeight(tableOffset);

const handleToWorkareaDetail = (bk_networkarea_id: number) => {
  router.push({
    name: 'workareaDetail',
    params: {
      workarea: bk_networkarea_id,
    },
  });
};

const handleEditWorkarea = (workareaData: NetworkArea) => {
  emit('edit', workareaData);
};

// todo
// 可能需要补充交互(message/重置筛选/重置pagination)
const handleDeleteWorkarea = (bk_networkarea_id: number) => {
  InfoBox({
    title: t('topoManager.workArea.delete.title'),
    cancelText: t('action.cancel'),
    onConfirm() {
      workareaStore.handleDeleteWorkarea(bk_networkarea_id);
      workareaStore.handleFetchWorkareaList();
    },
  });
};

watch(() => props.vendorList, () => {
  filterOption.list = props.vendorList.map((item, index) => ({
    value: index,
    text: t(vendorMap[item]?.label || ''),
  }));
});

watch(() => props.list, () => {
  loading.value = true;
  tableData.value = props.list;
  loading.value = false;
}, { immediate: true });

const tableRef = ref();
defineExpose({
  tableRef,
});

</script>
