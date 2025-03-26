<template>
  <div class="p-[24px]">
    <FlexRow>
      <template #left></template>
      <template #right>
        <div class="flex items-center">
          <DatePicker
            type="daterange"
            :placeholder="$t('topoManager.record.searchPlaceholder.date')"
            class="mr-[8px]"
          >
          </DatePicker>
          <SearchSelect
            class="w-[480px]"
            unique-select
            :placeholder="$t('topoManager.record.searchPlaceholder.searchSelect')"
            v-model.trim="searchSelectValue"
            :data="searchSelectData">
          </SearchSelect>
        </div>
      </template>
    </FlexRow>
    <Loading :loading="loading">
      <Table
        class="mt-[16px] w-full"
        ref="tableRef"
        :data="list"
        :empty-text="$t('table.empty')"
        :pagination="pagination"
        :sort-config="sortConfig"
        :show-settings="isShowSetting"
        :settings="settings"
        :max-height="maxHeight"
        @setting-change="handleSettingChange">
        <TableColumn
          :label="$t('topoManager.record.table.workareaName')"
          field="bk_networkarea_name"
          show-overflow="tooltip">
          <template #default="{ row }">
            {{ row.bk_networkarea_name || '--' }}
          </template>
        </TableColumn>
        <TableColumn
          :label="$t('topoManager.record.table.workareaId')"
          field="bk_networkarea_id"
          show-overflow="tooltip">
          <template #default="{ row }">
            {{ `#${row.bk_networkarea_id}` || '--' }}
          </template>
        </TableColumn>
        <TableColumn
          :label="$t('topoManager.record.table.workUnit')"
          field="bk_networkunit_name"
          show-overflow="tooltip">
          <template #default="{ row }">
            {{ row.bk_networkunit_name || '--' }}
          </template>
        </TableColumn>
        <TableColumn
          :label="$t('topoManager.record.table.type')"
          field="type"
          show-overflow="tooltip">
          <template #default="{ row }">
            {{ row.type || '--' }}
          </template>
        </TableColumn>
        <TableColumn
          :label="$t('topoManager.record.table.operator')"
          field="operator"
          show-overflow="tooltip">
          <template #default="{ row }">
            {{ row.operator || '--' }}
          </template>
        </TableColumn>
        <TableColumn
          :label="$t('topoManager.record.table.operateTime')"
          field="operate_time"
          show-overflow="tooltip"
          sortable>
          <template #default="{ row }">
            {{ filterTimeFormat(row.operate_time) || '--' }}
          </template>
        </TableColumn>
      </Table>
    </Loading>
  </div>
</template>

<script lang="ts" setup>
import { DatePicker, Loading, SearchSelect } from 'bkui-vue';
import { computed, onMounted, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';

import { Table, TableColumn } from '@blueking/table';
import { useDebounce } from '@vueuse/core';

import type { TopoEventListReqExactConditions } from '@/@types/topo';
import { filterTimeFormat } from '@/common/util';
import useDynamicsHeight from '@/composables/use-table-height';
import useTableSetting from '@/composables/use-table-setting';
import { useWorkareaStore } from '@/stores/workarea';

const { t } = useI18n();
const {
  handleFetchRecordList,
  handleFetchAllWorkarea,
  handleFetchAllWorkUnit,
} = useWorkareaStore();
const workareaStore = useWorkareaStore();
const searchSelectValue = ref([]);
const debounceSearch = useDebounce(searchSelectValue, 300);
const loading = ref(false);

const allWorkareaList = computed(() => Array.from(workareaStore.allWorkareaList.values()));
const allWorkUnitList = computed(() => Array.from(workareaStore.allWorkUnitList.values()).flat());

const searchSelectData = ref([]);

// 待优化 各影响table最大高度的元素的高度
const tableOffset = 200;
const { maxHeight } = useDynamicsHeight(tableOffset);

const pagination = reactive({ count: 0, limit: 50, current: 1 });
const sortConfig = ref({ multiple: true });

const { isShowSetting, settings, handleSettingChange } = useTableSetting({
  checked: [
    'bk_networkarea_name',
    'bk_networkarea_id',
    'bk_networkunit_name',
    'type',
    'operator',
    'operate_time',
  ],
});

const list = ref<TopoEvent[]>([]);

const extraData = computed(() => {
  const exact_include_conditions: Omit<TopoEventListReqExactConditions, 'accesspoint_id'> = {
    bk_networkarea_id: [],
    bk_networkunit_id: [],
    type: [],
    operator: [],
  };
  for (const item of Object.values(debounceSearch.value)) {
    switch (item.id) {
      case 'workareaName':
        break;
      case 'workareaId':
        exact_include_conditions.bk_networkarea_id = item.values.map(item => Number(item.id));
        break;
      case 'workUnit':
        break;
      case 'type':
        break;
      case 'actionPerson':
        break;
    }
  }
  return exact_include_conditions as TopoEventListReqExactConditions;
});

const fetchRecordList = async () => {
  try {
    loading.value = true;
    const params = {
      page: {
        offset: pagination.current - 1,
        limit: pagination.limit,
      },
      exact_include_conditions: extraData.value,
    };
    const res = await handleFetchRecordList(params);
    pagination.count = res.total || 0;
    list.value = res.items || [];
  } catch (err) {
    console.error(err);
  } finally {
    loading.value = false;
  }
};

const initSearchList = async () => {
  await Promise.all([
    handleFetchAllWorkarea(),
    handleFetchAllWorkUnit(),
  ]);

  searchSelectData.value = [
    {
      name: t('topoManager.record.searchItems.workareaName'),
      id: 'workareaName',
      multiple: true,
      children: allWorkareaList.value.map(item => ({
        id: item.bk_networkarea_name,
        name: item.bk_networkarea_name,
      })),
    },
    {
      name: t('topoManager.record.searchItems.workareaId'),
      id: 'workareaId',
      multiple: true,
      children: allWorkareaList.value.map(item => ({
        id: String(item.bk_networkarea_id),
        name: String(item.bk_networkarea_id),
      })),
    },
    {
      name: t('topoManager.record.searchItems.workUnit'),
      id: 'workUnit',
      multiple: true,
      children: allWorkUnitList.value.map(item => ({
        id: item.bk_networkunit_name,
        name: item.bk_networkunit_name,
      })),
    },
    {
      name: t('topoManager.record.searchItems.type'),
      id: 'type',
      multiple: true,
    },
    {
      name: t('topoManager.record.searchItems.operator'),
      id: 'actionPerson',
      multiple: true,
    },
  ];
};

watch(extraData, fetchRecordList);

onMounted(() => {
  fetchRecordList();
  initSearchList();
});

</script>
