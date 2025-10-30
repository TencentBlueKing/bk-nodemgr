<template>
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
          :data="searchSelectData"
          v-model.trim="searchSelectValue"
          :unique-select="true"
          :placeholder="'搜索、配置ID、配置名称、版本、配置类型、操作类型、操作人'"
          @update:model-value="handleSearchSelectChange"
        >
        </SearchSelect>
      </div>
    </section>
    <bk-loading title="数据加载中" :loading="loading">
      <Table
        :data="tableData"
        :empty-text="'暂无数据'"
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
          field="configpolicy_id"
          :title="t('配置ID')"
          min-width="130"
          :filter="filterOptionSource.configpolicy_id"
        ></TableColumn>
        <TableColumn
          field="configpolicy_name"
          :title="t('配置名称')"
          min-width="130"
          :filter="filterOptionSource.configpolicy_name"
        ></TableColumn>
        <TableColumn
          field="version"
          :title="t('版本')"
          min-width="130"
          fixed="left"
          sortable
          :filter="filterOptionSource.version"
        >
          <template #default="{ row }">
            {{ `V${row.version}` }}
          </template>
        </TableColumn>
        <TableColumn
          field="configpolicy_type"
          :title="t('配置类型')"
          min-width="130"
          :filter="filterOptionSource.configpolicy_type"
        >
          <template #default="{ row }">
            {{ configMap[row.configpolicy_type]}}
          </template>
        </TableColumn>
        <TableColumn
          field="type"
          :title="t('操作类型')"
          min-width="150"
          :filter="filterOptionSource.type"
        >
          <template #default="{ row }">
            {{ operateMap[row.type]}}
          </template>
        </TableColumn>
        <TableColumn
          field="operator"
          :title="t('操作人')"
          min-width="150"
          :filter="filterOptionSource.operator"
        ></TableColumn>
        <TableColumn
          field="operate_time"
          :title="t('操作时间')"
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
} from 'bkui-vue';
import dayjs from 'dayjs';
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
import useTableSetting from '@/composables/use-table-setting';
import { useMainStore } from '@/stores/main';
import { useNodeManageStore } from '@/stores/node-manage';

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
const tableData = ref<ConfigPolicyEvent[]>([]);

const maxHeight = computed(() => mainStore.windowInnerHeight - 214);
const loading = ref(false);

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
    text: '今天',
    value() {
      const end = new Date();
      const start = new Date(end.getFullYear(), end.getMonth(), end.getDate());
      return [start, end];
    },
  },
  {
    text: '近7天',
    value() {
      const end = new Date();
      const start = new Date();
      start.setTime(start.getTime() - 3600 * 1000 * 24 * 7);
      return [start, end];
    },
  },
  {
    text: '近15天',
    value() {
      const end = new Date();
      const start = new Date();
      start.setTime(start.getTime() - 3600 * 1000 * 24 * 15);
      return [start, end];
    },
  },
  {
    text: '近30天',
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

const timeFormatter = (val: string, format = 'YYYY-MM-DD HH:mm:ss') => (val ? dayjs(val).format(format) : '--');


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
const operateMap = {
  create: '新增',
  enable: '开启',
  disable: '禁用',
  update: '更新',
  delete: '删除',
};
// 配置类型映射
const configMap = {
  config_policy_agent: 'Agent 策略',
  config_policy_proxy: 'Proxy 策略',
};

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
    name: '配置ID',
    children: getUniqueChildrenFrom('configpolicy_id'),
    multiple: true,
  },
  {
    id: 'configpolicy_name',
    name: '配置名称',
    children: getUniqueChildrenFrom('configpolicy_name'),
    multiple: true,
  },
  {
    id: 'version',
    name: '版本',
    children: getUniqueChildrenFrom('version'),
    multiple: true,
  },
  {
    id: 'configpolicy_type',
    name: '配置类型',
    children: getUniqueChildrenFrom('configpolicy_type', configMap),
    multiple: true,
  },
  {
    id: 'type',
    name: '操作类型',
    children: getUniqueChildrenFrom('type', operateMap),
    multiple: true,
  },
  {
    id: 'operator',
    name: '操作人',
    children: getUniqueChildrenFrom('operator'),
    multiple: true,
  },
]);
// 筛选
const hostDistinct = ref<ConfigPolicyEventDistinctRespData | null>();
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
  const res = await ConfigPolicyAPIService.ConfigPolicyEventDistinct({}).catch(() => null);
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
                text = operateMap[value] || value;
                break;
              case 'configpolicy_type':
                text = configMap[value] || value;
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
    searchSelectValue.value.push({
      id: field,
      name: t(field),
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
    exact_include_conditions: {},
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
onMounted(async () => {
  await getTaskList();
  await getHostDistinct();
});
</script>
