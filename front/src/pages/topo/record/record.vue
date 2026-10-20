<template>
  <NoPermission
    v-if="!hasViewAuth"
    :auth-items="viewAuthItems"
    :resource-id="networkareaHistoryResourceIds"
  />
  <div v-else class="p-[24px]">
    <div class="flex items-center">
      <DatePicker
        type="daterange"
        v-model="operateTime"
        :placeholder="$t('topoManager.record.searchPlaceholder.date')"
        class="mr-[8px]"
      >
      </DatePicker>
      <SearchSelect
        :max-height="240"
        class="flex-1 bg-[#fff]"
        unique-select
        :placeholder="$t('topoManager.record.searchPlaceholder.searchSelect')"
        v-model.trim="searchSelectValue"
        @update:model-value="handleSearchSelectChange"
        :data="searchSelectData">
      </SearchSelect>
    </div>
    <Loading :loading="loading">
      <Table
        class="mt-[16px] w-full filterTable"
        ref="tableRef"
        :data="list"
        :empty-text="$t('table.empty')"
        :pagination="pagination"
        :sort-config="sortConfig"
        :show-settings="isShowSetting"
        :settings="settings"
        :max-height="maxHeight"
        @column-filter="handleFilter"
        @setting-change="handleSettingChange">
        <TableColumn
          :label="$t('topoManager.record.table.workareaName')"
          field="bk_networkarea_name"
          :filter="filterOptionSource.workareaName"
          show-overflow="tooltip">
          <template #default="{ row }">
            {{ row.bk_networkarea_name || '--' }}
          </template>
        </TableColumn>
        <TableColumn
          :label="$t('topoManager.record.table.workareaId')"
          field="bk_networkarea_id"
          :filter="filterOptionSource.workareaId"
          show-overflow="tooltip">
          <template #default="{ row }">
            {{ `#${row.bk_networkarea_id}` || '--' }}
          </template>
        </TableColumn>
        <TableColumn
          :label="$t('topoManager.record.table.workUnit')"
          field="bk_networkunit_name"
          :filter="filterOptionSource.workUnit"
          show-overflow="tooltip">
          <template #default="{ row }">
            {{ row.bk_networkunit_name || '--' }}
          </template>
        </TableColumn>
        <TableColumn
          :label="$t('topoManager.record.table.type')"
          field="type"
          :filter="filterOptionSource.type"
          show-overflow="tooltip">
          <template #default="{ row }">
            {{ typeMap[row.type] || row.type || '--' }}
          </template>
        </TableColumn>
        <TableColumn
          :label="$t('topoManager.record.table.operator')"
          field="operator"
          :filter="filterOptionSource.operator"
          show-overflow="tooltip">
          <template #default="{ row }">
            <UserNameDisplay :name="row.operator" />
          </template>
        </TableColumn>
        <TableColumn
          :label="$t('topoManager.record.table.operateTime')"
          field="operate_time"
          show-overflow="tooltip"
          sortable>
          <template #default="{ row }">
            {{ formatTimestamp(row.operate_time) || '--' }}
          </template>
        </TableColumn>
      </Table>
    </Loading>
  </div>
</template>

<script lang="ts" setup>
import { DatePicker, Loading, SearchSelect } from 'bkui-vue';
import type { ISearchItem, ISearchValue } from 'bkui-vue/lib/search-select/utils';
import { computed, onMounted, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';

import { Table, TableColumn } from '@blueking/table';

import type { TopoEventExactConditions, TopoEventFuzzyConditions } from '@/@types/topo';
import { TopoService } from '@/api/modules/topo';
import { translateOperatorItems } from '@/common/user-display';
import { formatTimestamp, getTimeStamp } from '@/common/util';
import NoPermission from '@/components/no-permission.vue';
import useDynamicsHeight from '@/composables/use-table-height';
import useTableSetting from '@/composables/use-table-setting';
import type { PageAuthItem } from '@/constants/auth';
import { getPageAuthorizedItems } from '@/constants/auth';
import { useAuthStore } from '@/stores/auth';
import { useWorkareaStore } from '@/stores/workarea';

const { t } = useI18n();
const {
  handleFetchRecordList,
  handleFetchAllWorkarea,
  handleFetchAllWorkUnit,
} = useWorkareaStore();

// ===== 权限判断：networkarea_history_view 无权限时展示占位页 =====
const authStore = useAuthStore();
const viewAuthItems: PageAuthItem[] = [
  { id: 'networkarea_history_view', action: 'networkarea_history_view', resourceType: 'networkarea', routes: [] },
];
const hasViewAuth = computed(() => {
  // authorized 尚未加载 → 先放行，避免闪烁
  if (!authStore.authorizedMap['networkarea_history_view']) return true;
  return authStore.hasAuthorizedResource('networkarea_history_view');
});

// 申请查看权限时携带的资源实例：全部管控区域 ID（verify 逐区域鉴权，
// 无权限的区域进入申请信息，用户可在权限中心选择具体区域申请）。
// 过滤负数 ID（异常数据）；0 为直连区域，保留
const networkareaHistoryResourceIds = computed<number[]>(() => Array.from(workareaStore.allWorkareaList.keys())
  .filter(id => id >= 0));

// 操作类型映射
const typeMap = ref({
  'networkarea-create': t('topoManager.record.table.createNetworkarea'),
  'networkarea-update': t('topoManager.record.table.updateNetworkarea'),
  'networkarea-delete': t('topoManager.record.table.deleteNetworkarea'),
  'networkunit-create': t('topoManager.record.table.createNetworkunit'),
  'networkunit-update': t('topoManager.record.table.updateNetworkunit'),
  'networkunit-delete': t('topoManager.record.table.deleteNetworkunit'),
  'accesspoint-create': t('topoManager.record.table.createAccesspoint'),
  'accesspoint-update': t('topoManager.record.table.updateAccesspoint'),
  'accesspoint-delete': t('topoManager.record.table.deleteAccesspoint'),
});
const workareaStore = useWorkareaStore();
const searchSelectValue = ref<ISearchValue[]>([]);
const loading = ref(false);

const allWorkareaList = computed(() => Array.from(workareaStore.allWorkareaList.values()));
const allWorkUnitList = computed(() => Array.from(workareaStore.allWorkUnitList.values()).flat());
const uniqueChildren = (values: (string | number)[]) => [...new Set(values.filter(value => value !== '' && value !== null && value !== undefined))]
  .map(value => ({ id: String(value), name: String(value) }));

const operatorChildren = ref<{ id: string; name: string }[]>([]);
const searchSelectData = computed<ISearchItem[]>(() => [
  {
    name: t('topoManager.record.searchItems.workareaName'),
    id: 'workareaName',
    multiple: true,
    children: uniqueChildren(allWorkareaList.value.map(item => item.bk_networkarea_name)),
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
    children: uniqueChildren(allWorkUnitList.value.map(item => item.bk_networkunit_name)),
  },
  {
    name: t('topoManager.record.searchItems.type'),
    id: 'type',
    multiple: true,
    children: Object.entries(typeMap.value).map(([id, name]) => ({ id, name })),
  },
  {
    name: t('topoManager.record.searchItems.operator'),
    id: 'actionPerson',
    multiple: true,
    children: operatorChildren.value,
  },
]);

const filterOptionSource = reactive({
  workareaName: { list: uniqueChildren(allWorkareaList.value.map(item => item.bk_networkarea_name)).map(({ id, name }) => ({ value: id, text: name })), checked: [] as string[], filterScope: 'all' },
  workareaId: { list: uniqueChildren(allWorkareaList.value.map(item => item.bk_networkarea_id)).map(({ id, name }) => ({ value: id, text: name })), checked: [] as string[], filterScope: 'all' },
  workUnit: { list: uniqueChildren(allWorkUnitList.value.map(item => item.bk_networkunit_name)).map(({ id, name }) => ({ value: id, text: name })), checked: [] as string[], filterScope: 'all' },
  type: {
    list: Object.entries(typeMap.value).map(([value, text]) => ({ value, text })),
    checked: [] as string[],
    filterScope: 'all',
  },
  operator: { list: [] as { value: string; text: string }[], checked: [] as string[], filterScope: 'all' },
});

watch([allWorkareaList, allWorkUnitList], () => {
  filterOptionSource.workareaName.list = uniqueChildren(
    allWorkareaList.value.map(item => item.bk_networkarea_name),
  ).map(({ id, name }) => ({ value: id, text: name }));
  filterOptionSource.workareaId.list = uniqueChildren(
    allWorkareaList.value.map(item => item.bk_networkarea_id),
  ).map(({ id, name }) => ({ value: id, text: name }));
  filterOptionSource.workUnit.list = uniqueChildren(
    allWorkUnitList.value.map(item => item.bk_networkunit_name),
  ).map(({ id, name }) => ({ value: id, text: name }));
});

const handleSearchSelectChange = (data: ISearchValue[]) => {
  Object.keys(filterOptionSource).forEach((key) => {
    filterOptionSource[key as keyof typeof filterOptionSource].checked = [];
  });
  data.forEach((item) => {
    const key = (item.id === 'actionPerson' ? 'operator' : item.id) as keyof typeof filterOptionSource;
    if (filterOptionSource[key]) {
      filterOptionSource[key].checked = item.values?.map(value => value.id) || [];
    }
  });
};

const handleFilter = ({ checked, field }: { checked: string[]; field: string }) => {
  const searchId = ({ bk_networkarea_name: 'workareaName', bk_networkarea_id: 'workareaId', bk_networkunit_name: 'workUnit', operator: 'actionPerson' } as Record<string, string>)[field] || field;
  const index = searchSelectValue.value.findIndex(item => item.id === searchId);
  if (index > -1) searchSelectValue.value.splice(index, 1);
  if (checked.length) {
    const filter = searchSelectData.value.find(item => item.id === searchId);
    searchSelectValue.value.push({
      id: searchId,
      name: filter?.name || field,
      values: checked.map(id => ({
        id,
        name: filter?.children?.find(item => item.id === id)?.name || id,
      })),
    });
  }
  handleSearchSelectChange(searchSelectValue.value);
};

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
}, 'topoMng-record');

const list = ref<TopoEvent[]>([]);
const operateTime = ref([]);

const exactData = computed(() => {
  const exact_include_conditions: Omit<TopoEventExactConditions, 'accesspoint_id'> = {
    bk_networkarea_id: [],
    bk_networkunit_id: [],
    type: [],
    operator: [],
  };
  for (const item of searchSelectValue.value) {
    switch (item.id) {
      case 'workareaId':
        exact_include_conditions.bk_networkarea_id = item.values?.map(item => Number(item.id)) || [];
        break;
      case 'type':
        exact_include_conditions.type = item.values?.map(item => item.id) as string[];
        break;
      case 'actionPerson':
        exact_include_conditions.operator = item.values?.map(item => item.id) as string[];
        break;
    }
  }
  return exact_include_conditions as TopoEventExactConditions;
});

const fuzzyData = computed(() => {
  const fuzzy_include_conditions: TopoEventFuzzyConditions = {
    bk_networkarea_name: [],
    bk_networkunit_name: [],
  };
  for (const item of searchSelectValue.value) {
    switch (item.id) {
      case 'workareaName':
        fuzzy_include_conditions.bk_networkarea_name = item.values?.map(value => value.name) || [];
        break;
      case 'workUnit':
        fuzzy_include_conditions.bk_networkunit_name = item.values?.map(value => value.name) || [];
        break;
    }
  }
  return fuzzy_include_conditions;
});

const operateTimeRange = computed(() => {
  if (operateTime.value.length !== 0 && operateTime.value.every(item => item !== '')) {
    const operate_time_range: TimeRange = {
      start_timestamp_sec: getTimeStamp(operateTime.value[0]),
      end_timestamp_sec: getTimeStamp(operateTime.value[1]),
    };
    return operate_time_range;
  }
  return null as unknown as TimeRange;
});

const fetchRecordList = async () => {
  try {
    loading.value = true;
    const params = {
      page: {
        offset: (pagination.current - 1) * pagination.limit,
        limit: pagination.limit,
      },
      exact_include_conditions: exactData.value,
      fuzzy_include_conditions: fuzzyData.value,
      operate_time_range: operateTimeRange.value,
    };
    const res = await handleFetchRecordList(params).catch(() => ({ total: 0, items: [] }));
    pagination.count = res.total;
    list.value = res.items || [];
  } catch (err) {
    console.error(err);
  } finally {
    loading.value = false;
  }
};

const fetchOperatorDistinct = async () => {
  const res = await TopoService.EventDistinct({
    exact_include_conditions: exactData.value,
    fuzzy_include_conditions: fuzzyData.value,
    operate_time_range: operateTimeRange.value,
  }).catch(() => null);
  if (!res) return;

  const operator = searchSelectData.value.find(item => item.id === 'actionPerson');
  if (operator) {
    const children = (res.operator || []).filter(Boolean).map(name => ({ id: name, name }));
    await translateOperatorItems(children);
    operatorChildren.value = children;
    filterOptionSource.operator.list = children.map(item => ({ value: item.id, text: item.name }));
  }
};

// const initSearchList = async () => {
//   await Promise.all([
//     handleFetchAllWorkarea(),
//     handleFetchAllWorkUnit(),
//   ]);

//   searchSelectData.value = [
//     {
//       name: t('topoManager.record.searchItems.workareaName'),
//       id: 'workareaName',
//       multiple: false,
//     },
//     {
//       name: t('topoManager.record.searchItems.workareaId'),
//       id: 'workareaId',
//       multiple: true,
//       children: allWorkareaList.value.map(item => ({
//         id: String(item.bk_networkarea_id),
//         name: String(item.bk_networkarea_id),
//       })),
//     },
//     {
//       name: t('topoManager.record.searchItems.workUnit'),
//       id: 'workUnit',
//       multiple: false,
//     },
//     {
//       name: t('topoManager.record.searchItems.type'),
//       id: 'type',
//       multiple: true,
//     },
//     {
//       name: t('topoManager.record.searchItems.operator'),
//       id: 'actionPerson',
//       multiple: true,
//     },
//   ];
// };

watch([exactData, fuzzyData, operateTime], () => {
  pagination.current = 1;
  if (hasViewAuth.value) fetchRecordList();
});

onMounted(async () => {
  // networkarea_history_view 按页面 items 兜底加载
  const pageItems = getPageAuthorizedItems('record');
  await authStore.fetchAuthorized(pageItems, 'page:record').catch(() => {});
  // 拉取全部管控区域：搜索下拉数据源 + 无权限时申请查看权限所需的资源实例列表
  await handleFetchAllWorkarea().catch(() => {});
  await handleFetchAllWorkUnit().catch(() => {});
  if (hasViewAuth.value) {
    fetchRecordList();
    fetchOperatorDistinct();
  }
});

</script>
