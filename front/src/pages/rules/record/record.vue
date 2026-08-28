<template>
  <!-- <Tab
    v-model:active="active"
    type="unborder-card"
    :label-height="41"
    class="text-[14px] bg-[#fff] h-[41px] absolute z-10 w-full"
  >
    <Tab.TabPanel
      v-for="item in panels"
      :key="item.name"
      :label="item.label"
      :name="item.name"
    >
    </Tab.TabPanel>
  </Tab> -->
  <div class="p-[24px]">
    <section class="flex justify-between mb-[15px]">
      <div class="flex gap-[12px]">
        <DatePicker
          v-model="dateValue"
          :shortcut-selected-index="1"
          :shortcuts="shortcutsRange"
          format="yyyy-MM-dd HH:mm:ss"
          type="datetimerange"
          use-shortcut-text
          @change="pickSuccess"
        />
      </div>
      <div class="flex-1 ml-[8px]">
        <SearchSelect
          ref="searchSelect"
          class="bg-[#fff]"
          :data="searchSelectData"
          v-model.trim="searchSelectValue"
          :unique-select="true"
          :placeholder="$t('rulesRecord.searchPlaceholder')"
          @update:model-value="handleSearchSelectChange"
        >
        </SearchSelect>
      </div>
    </section>
    <bk-loading :title="$t('table.loading')" :loading="loading">
      <Table
        class="filterTable"
        :data="tableData"
        :empty-text="$t('table.empty')"
        :pagination="pagination"
        :column-config="{ resizable: true }"
        show-overflow-tooltip
        :max-height="maxHeight"
        :show-settings="isShowSetting"
        :settings="settings"
        @setting-change="handleSettingChange"
        @column-filter="handleFilter"
        @page-limit-change="pageLimitChange"
        @page-value-change="pageValueChange"
        :sort-config="sortConfig"
      >
        <TableColumn
          field="configpolicy_name"
          :title="$t('rulesRecord.configName')"
          fixed="left"
          min-width="130"
          :filter="filterOptionSource.configpolicy_name"
        ></TableColumn>
        <TableColumn
          field="configpolicy_id"
          :title="$t('rulesRecord.configID')"
          min-width="130"
          :filter="filterOptionSource.configpolicy_id"
        ></TableColumn>
        <TableColumn
          field="version"
          :title="$t('rulesRecord.version')"
          min-width="130"
          :filter="filterOptionSource.version"
        >
          <template #default="{ row }">
            {{ `V${row.version}` }}
          </template>
        </TableColumn>
        <TableColumn
          field="configpolicy_type"
          :title="$t('rulesRecord.configType')"
          min-width="130"
          :filter="filterOptionSource.configpolicy_type"
        >
          <template #default="{ row }">
            {{ configMap[row.configpolicy_type]}}
          </template>
        </TableColumn>
        <TableColumn
          field="type"
          :title="$t('rulesRecord.operateType')"
          min-width="150"
          :filter="filterOptionSource.type"
        >
          <template #default="{ row }">
            {{ operateMap[row.type]}}
          </template>
        </TableColumn>
        <TableColumn
          field="operator"
          :title="$t('rulesRecord.operator')"
          min-width="150"
          :filter="filterOptionSource.operator"
        >
          <template #default="{ row }">
            <UserNameDisplay :name="row.operator" />
          </template>
        </TableColumn>
        <TableColumn
          field="operate_time"
          :title="$t('rulesRecord.operateTime')"
          sort-type="number"
          sortable
          min-width="200"
        >
          <template #default="{ row }">
            <span>{{ timeFormatter(row.operate_time) }}</span>
          </template>
        </TableColumn>
      </Table>
    </bk-loading>
  </div>
</template>
<script setup lang="ts">
import {
  DatePicker,
  SearchSelect,
  Tab,
} from 'bkui-vue';
import { computed, onMounted, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';
import type { VxeTablePropTypes } from 'vxe-table';

import { Table, TableColumn } from '@blueking/table';

import type {
  ConfigPolicyEventDistinctRespData,
  ConfigPolicyEventExactConditions,
  ConfigPolicyEventFuzzyConditions,
} from '@/@types/configpolicy';
import { ConfigPolicyAPIService } from '@/api/modules/configpolicy';
import { formatTimeByTimezone } from '@/common/util';
import useTableSetting from '@/composables/use-table-setting';
import { useMainStore } from '@/stores/main';
import { useNodeManageStore } from '@/stores/node-manage';
import { translateOperatorItems } from '@/common/user-display';

interface IFilterOption {
  list: { value: string | boolean, text: string;  }[];
  checked: string[];
  filterScope: string;
  match?: string,
}
interface ISearch {
  id: string;
  name: string;
  values: {
    id: string;
    name: string
  }[];
}

const { t } = useI18n();
const router = useRouter();
const mainStore = useMainStore();
const nodeManageStore = useNodeManageStore();
// 从 store 获取策略专用的单选业务 ID
const strategyBizId = computed(() => mainStore.strategyBizId);
const tableData = ref<ConfigPolicyEvent[]>([]);

const maxHeight = computed(() => mainStore.windowInnerHeight - 214 - (mainStore.noticeShow ? 40 : 0));
const loading = ref(false);

const active = ref('');
const panels = computed(() => [
  { name: 'configStrategy', label: t('rulesRecord.configStrategy') },
  // { name: 'deploymentStrategy', label: '部署策略' },
]);

const sortConfig = ref<VxeTablePropTypes.SortConfig>({
  sortMethod({ data, sortList }) {
    const sortItem = sortList[0];
    // 取出第一个排序的列
    const { field, order } = sortItem;
    // 通用排序函数
    function sortData(a: Release, b: Release, field: string) {
      return a[field] - b[field];
    }

    const sortedList = data.sort((a: Release, b: Release) => {
      const comparison = sortData(a, b, field);
      return order === 'desc' ? -comparison : comparison;
    });

    return sortedList;
  },
});
// 后端分页
const pagination = reactive({ count: 0, limit: 50, current: 1, remote: true });
// 日期选择
const dateValue = ref([
  new Date().setTime(new Date().getTime() - 3600 * 1000 * 24 * 7),
  new Date(),
]);
const shortcutsRange = reactive([
  {
    text: t('rulesRecord.today'),
    value() {
      const end = new Date();
      const start = new Date(end.getFullYear(), end.getMonth(), end.getDate());
      return [start, end];
    },
  },
  {
    text: t('rulesRecord.last7Days'),
    value() {
      const end = new Date();
      const start = new Date();
      start.setTime(start.getTime() - 3600 * 1000 * 24 * 7);
      return [start, end];
    },
  },
  {
    text: t('rulesRecord.last15Days'),
    value() {
      const end = new Date();
      const start = new Date();
      start.setTime(start.getTime() - 3600 * 1000 * 24 * 15);
      return [start, end];
    },
  },
  {
    text: t('rulesRecord.last30Days'),
    value() {
      const end = new Date();
      const start = new Date();
      start.setTime(start.getTime() - 3600 * 1000 * 24 * 30);
      return [start, end];
    },
  },
]);
const pickSuccess = async () => {
  await getTaskList();
};

const timeFormatter = (val: string | number, format = 'YYYY-MM-DD HH:mm:ss') => (val ? formatTimeByTimezone(val, format) : '--');


// 表格
const { isShowSetting, settings, handleSettingChange } = useTableSetting({
  checked: [
    'configpolicy_id',
    'configpolicy_name',
    'configpolicy_type',
    'type',
    'version',
    'operator',
    'operate_time',
  ],
  disabled: [],
}, 'rulesMng-record');

// 操作类型中文映射
const operateMap = computed(() => ({
  create: t('rulesRecord.action.create'),
  enable: t('rulesRecord.action.enable'),
  disable: t('rulesRecord.action.disable'),
  update: t('rulesRecord.action.update'),
  delete: t('rulesRecord.action.delete'),
  reorder_priorities: t('rulesRecord.action.reorderPriorities'),
}));
// 配置类型映射
const configMap = computed(() => ({
  config_policy_agent: t('rulesRecord.configTypeMap.agent'),
  config_policy_proxy: t('rulesRecord.configTypeMap.proxy'),
}));

// eslint-disable-next-line max-len
const getUniqueChildrenFrom = <K extends keyof ConfigPolicyEventExactConditions | keyof ConfigPolicyEventFuzzyConditions>(
  prop: K,
  keyMap?: Record<string, any>,
) => {
  const uniqueValues = hostDistinct.value?.[prop] || [];
  return uniqueValues
    .filter((item: any) => item !== '')
    .map((value: string | number) => {
      let text;
      switch (prop) {
        case 'version':
          text = `V${value}`;
          break;
        case 'type':
        case 'configpolicy_type':
          text = keyMap?.[value] || String(value);
          break;
        default:
          text = String(value);
          break;
      }
      return {
        id: value,
        name: text,
      };
    });
};
const searchSelectData = computed(() => [
  {
    id: 'configpolicy_id',
    name: t('rulesRecord.configID'),
    children: getUniqueChildrenFrom('configpolicy_id'),
    multiple: true,
  },
  {
    id: 'configpolicy_name',
    name: t('rulesRecord.configName'),
    children: getUniqueChildrenFrom('configpolicy_name'),
    multiple: true,
  },
  {
    id: 'version',
    name: t('rulesRecord.version'),
    children: getUniqueChildrenFrom('version'),
    multiple: true,
  },
  {
    id: 'configpolicy_type',
    name: t('rulesRecord.configType'),
    children: getUniqueChildrenFrom('configpolicy_type', configMap.value),
    multiple: true,
  },
  {
    id: 'type',
    name: t('rulesRecord.operateType'),
    children: getUniqueChildrenFrom('type', operateMap.value),
    multiple: true,
  },
  {
    id: 'operator',
    name: t('rulesRecord.operator'),
    children: translatedOperatorChildren.value,
    multiple: true,
  },
]);
// 筛选
const hostDistinct = ref<ConfigPolicyEventDistinctRespData | null>();
// 搜索栏 operator 下拉（翻译后的显示名）
const translatedOperatorChildren = ref<{ id: string; name: string }[]>([]);
const filterOptionSource = reactive<Record<string, IFilterOption>>({
  version: {
    list: [],
    checked: [],
    match: 'fuzzy',
    filterScope: 'all',
  },
  configpolicy_type: {
    list: [],
    checked: [],
    match: 'fuzzy',
    filterScope: 'all',
  },
  type: {
    list: [],
    checked: [],
    match: 'fuzzy',
    filterScope: 'all',
  },
  operator: {
    list: [],
    checked: [],
    match: 'fuzzy',
    filterScope: 'all',
  },
});
const getHostDistinct = async () => {
  const res = await ConfigPolicyAPIService.ConfigPolicyEventDistinct({
    exact_include_conditions: {
      bk_biz_id: strategyBizId.value ? [strategyBizId.value] : [],
    },
  }).catch(() => null);
  if (res) {
    hostDistinct.value = res;
    Object.keys(res).forEach((key: any) => {
      const curUniqueValues = res[key] || [];
      if (filterOptionSource[key]) {
        filterOptionSource[key].list = curUniqueValues
          .filter((item: any) => item !== '')
          .map((value: string | number) => {
            let text;
            switch (key) {
              case 'version':
                text = `V${value}`;
                break;
              case 'type':
                text = operateMap.value[value] || value;
                break;
              case 'configpolicy_type':
                text = configMap.value[value] || value;
                break;
              default:
                text = value;
                break;
            }
            return {
              text,
              value,
            };
          });
      }
    });
    // 翻译 operator 筛选项显示名
    await translateOperatorItems(filterOptionSource.operator.list);
    // 翻译搜索栏 operator 下拉显示名
    const operatorChildren = getUniqueChildrenFrom('operator');
    await translateOperatorItems(operatorChildren);
    translatedOperatorChildren.value = operatorChildren;
  }
};
const handleFilter = ({
  checked,
  field,
}: {
  checked: string[];
  field: string;
}) => {
  const index = searchSelectValue.value.findIndex((item: any) => item.id === field);
  index > -1 && searchSelectValue.value.splice(index, 1);
  if (checked.length) {
    const item = searchSelectData.value.find(item => item.id === field);
    searchSelectValue.value.push({
      id: field,
      name: item ? item.name : field,
      values: checked.map((item: any) => ({
        id: item,
        name: item,
      })),
    });
  }
};
// 分页操作
const pageLimitChange = async (limit: number) => {
  pagination.limit = limit;
  await getTaskList();
};
const pageValueChange = async (current: number) => {
  pagination.current = current;
  await getTaskList();
};

// 搜索
const searchSelectValue = ref<{ id: string; name: string; values: any }[]>([]);
const handleSearchSelectChange = async (data: ISearch[]) => {
  // 给筛选器添加选中值
  Object.keys(filterOptionSource).forEach((key) => {
    filterOptionSource[key].checked = [];
  });
  data.forEach((item) => {
    if (filterOptionSource[item.id]) {
      filterOptionSource[item.id].checked = item.values.map((item: any) => item.id);
    }
  });
};

const getTimestampInSeconds = (originalDate: number | Date) => {
  const timestampInMilliseconds = new Date(originalDate).getTime();
  const timestampInSeconds = Math.floor(timestampInMilliseconds / 1000);
  return timestampInSeconds;
};
const getParams = () => {
  const params = {
    page: {
      limit: pagination.limit,
      offset: (pagination.current - 1) * pagination.limit,
    },
    exact_include_conditions: {
      bk_biz_id: strategyBizId.value ? [strategyBizId.value] : [],
    } as Record<string, any>,
    fuzzy_include_conditions: {} as Record<string, string[]>,
    operate_time_range: {
      start_timestamp_sec: getTimestampInSeconds(dateValue.value[0]),
      end_timestamp_sec: getTimestampInSeconds(dateValue.value[1]),
    },
  };
  searchSelectValue.value.forEach((item: any) => {
    const target = params.exact_include_conditions;
    target[item.id] = item.values?.map((value: any) => value.id);
  });
  return params;
};
const getTaskList = async () => {
  loading.value = true;
  const res = await ConfigPolicyAPIService.ConfigPolicyEventList(getParams()).catch(() => ({
    total: 0,
    items: [],
  }));
  pagination.count = res.total;
  tableData.value = res.items;
  loading.value = false;
};

watch(
  () => searchSelectValue,
  async () => {
    await getTaskList();
  },
  { deep: true },
);
// 监听策略业务切换，重新加载操作记录
watch(
  strategyBizId,
  async () => {
    if (strategyBizId.value) {
      // 主列表与 distinct 筛选数据无依赖，并行加载
      await Promise.all([
        getTaskList(),
        getHostDistinct(),
      ]);
    }
  },
);
onMounted(async () => {
  // 主列表与 distinct 筛选数据无依赖，并行加载
  await Promise.all([
    getTaskList(),
    getHostDistinct(),
  ]);
});
</script>
