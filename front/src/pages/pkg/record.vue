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
          :placeholder="'搜索版本号、操作系统、架构、操作类型、操作人、包类型'"
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
        :sort-config="sortConfig"
      >
        <TableColumn
          field="version"
          :title="t('版本号')"
          min-width="130"
          fixed="left"
          sortable
          :filter="filterOptionSource.version"
        ></TableColumn>
        <TableColumn
          field="release_type"
          :title="t('包类型')"
          min-width="130"
          :filter="filterOptionSource.release_type"
        ></TableColumn>
        <TableColumn
          field="os_type"
          :title="t('操作系统')"
          min-width="130"
          :filter="filterOptionSource.os_type"
        ></TableColumn>
        <TableColumn
          field="cpu_arch"
          :title="t('架构')"
          min-width="130"
          :filter="filterOptionSource.cpu_arch"
        ></TableColumn>
        <TableColumn
          field="event_type"
          :title="t('操作类型')"
          min-width="150"
          :filter="filterOptionSource.event_type"
        >
          <template #default="{ row }">
            {{ eventMap[row.event_type]}}
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

import type { PackageEventDistinctRespData } from '@/@types/pkg';
import { PackageService } from '@/api/modules/pkg';
import { compareVersions  } from '@/common/util';
import usePage from '@/composables/use-page';
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
const tableData = ref<PackageEvent[]>([]);
// 分页
const { pagination } = usePage(tableData);

const maxHeight = computed(() => mainStore.windowInnerHeight - 214);
const loading = ref(false);

const sortConfig = ref<VxeTablePropTypes.SortConfig>({
  sortMethod({ data, sortList }) {
    const sortItem = sortList[0];
    // 取出第一个排序的列
    const { field, order } = sortItem;
    // 通用排序函数
    function sortData(a: Release, b: Release, field: string) {
      if (field === 'version') {
        return compareVersions(a[field], b[field]);
      }
      return a[field] - b[field];
    }

    const sortedList = data.sort((a: Release, b: Release) => {
      const comparison = sortData(a, b, field);
      return order === 'desc' ? -comparison : comparison;
    });

    return sortedList;
  },
});

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
    'version',
    'release_type',
    'os_type',
    'cpu_arch',
    'event_type',
    'operator',
    'operate_time',
  ],
  disabled: [],
}, 'pkgMng-record');

// 操作类型中文映射
const eventMap = {
  publish: '发布',
  enable: '启用',
  disable: '禁用',
  delete: '删除',
  set_as_default: '设置为默认版本',
  unset_as_default: '取消默认版本',
};

const getUniqueChildrenFrom = <K extends keyof PackageEventDistinctRespData>(
  prop: K,
  keyMap?: Record<string, any>,
) => {
  const uniqueValues = hostDistinct.value?.[prop] || [];
  return uniqueValues
    .filter((item: any) => item !== '')
    .map((value: any) => ({
      id: value,
      name: keyMap && keyMap[value] ? keyMap[value] : String(value),
    }));
};
const searchSelectData = computed(() => [
  {
    id: 'version',
    name: '版本号',
    children: getUniqueChildrenFrom('version'),
    multiple: true,
  },
  {
    id: 'release_type',
    name: '包类型',
    children: getUniqueChildrenFrom('release_type'),
    multiple: true,
  },
  {
    id: 'os_type',
    name: '操作系统',
    children: getUniqueChildrenFrom('os_type'),
    multiple: true,
  },
  {
    id: 'cpu_arch',
    name: '架构',
    children: getUniqueChildrenFrom('cpu_arch'),
    multiple: true,
  },
  {
    id: 'event_type',
    name: '操作类型',
    children: getUniqueChildrenFrom('event_type', eventMap),
    multiple: true,
  },
  {
    id: 'operator',
    name: '操作人',
    children: getUniqueChildrenFrom('operator'),
    multiple: true,
  },
]);

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
      limit: 0,
      offset: 0,
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
  const res = await PackageService.PackageEventList(getParams()).catch(() => ({
    total: 0,
    items: [],
  }));
  tableData.value = res.items;
  loading.value = false;
};

// 筛选
const hostDistinct = ref<PackageEventDistinctRespData | null>();
const filterOptionSource = reactive<Record<string, IFilterOption>>({
  version: {
    list: [],
    checked: [],
    match: 'fuzzy',
    filterScope: 'all',
  },
  os_type: {
    list: [],
    checked: [],
    match: 'fuzzy',
    filterScope: 'all',
  },
  cpu_arch: {
    list: [],
    checked: [],
    match: 'fuzzy',
    filterScope: 'all',
  },
  release_type: {
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
  event_type: {
    list: [],
    checked: [],
    match: 'fuzzy',
    filterScope: 'all',
  },
});
const getHostDistinct = async () => {
  const res = await PackageService.PackageEventDistinct({}).catch(() => null);
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
              case 'event_type':
                text = eventMap[value] || value;
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
