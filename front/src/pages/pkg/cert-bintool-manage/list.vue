<template>
  <div class="p-[24px] h-[calc(100%_-_52px)] flex flex-col">
    <!-- 搜索栏 -->
    <div class="flex items-center w-full h-[32px] mb-[16px]">
      <Button theme="primary" @click="handleUpload" class="mr-[16px]">包上传</Button>
      <SearchSelect
        class="flex-1"
        ref="searchSelect"
        :data="searchSelectData"
        v-model.trim="searchSelectValue"
        :unique-select="true"
        :placeholder="t('包文件名、上传用户')"
        @update:model-value="handleSearchSelectChange"
      >
      </SearchSelect>
    </div>
    <Loading
      title="数据加载中"
      :loading="loading"
      class="flex-1 overflow-auto"
    >
      <Table
        class="w-full filterTable"
        :max-height="maxHeight"
        :data="packageList"
        :empty-text="'暂无数据'"
        :pagination="pagination"
        show-overflow-tooltip
        :show-settings="isShowSetting"
        :settings="settings"
        @setting-change="handleSettingChange"
        @column-filter="handleFilter"
        :sort-config="sortConfig"
      >
        <TableColumn
          field="file_name"
          :title="'包文件名'"
          :min-width="320"
          fixed="left"
          show-overflow="tooltip"
        ></TableColumn>
        <TableColumn
          field="operator"
          :title="'上传用户'"
          :min-width="120"
          :filter="filterOptionSource.operator"
        ></TableColumn>
        <TableColumn
          field="updated_at"
          :title="'上传时间'"
          :min-width="180"
          sort-type="number"
          sortable
        >
          <template #default="{ row }">
            {{ formatTimestamp(row.updated_at) }}
          </template>
        </TableColumn>
        <TableColumn
          field="action"
          :title="'操作'"
          fixed="right"
          :min-width="120"
        >
          <template #default="{ row }">
            <div class="flex items-center">
              <Popover
                width="360"
                theme="light"
                trigger="click"
                placement="top-start"
                :is-show="row.isDeletePopShow"
              >
                <Button
                  theme="primary"
                  text
                  v-show="!row.enabled"
                  @click="row.isDeletePopShow = true"
                >删除</Button>
                <template #content>
                  <div class="px-[4px] pt-[8px] pb-[4px]">
                    <div class="text-[16px] text-[#313238] mb-[6px]">
                      确认删除该{{ route.name === "certMng" ? "证书" : "工具" }}？
                    </div>
                    <div class="text-[12px] text-[#4D4F56] w-full mb-[5px]">删除目标：{{row.file_name}}</div>
                    <div class="text-[12px] text-[#4D4F56] w-full">删除后不可恢复，请谨慎操作！</div>
                    <div class="mt-[20px] flex gap-[8px] justify-end">
                      <Button theme="primary" @click="handleDelete(row)">删除</Button>
                      <Button @click="row.isDeletePopShow = false">取消</Button>
                    </div>
                  </div>
                </template>
              </Popover>
            </div>
          </template>
        </TableColumn>
      </Table>
    </Loading>
  </div>
  <pkg-upload-sideslider v-model:is-show="isShow" @confirm="handleConfirm" />
</template>
<script lang="ts" setup>
// 排序类型
import { Button, Loading, Popover, SearchSelect, Select, Tag, TagInput } from 'bkui-vue';
import { AngleDown, AngleRight,EditLine } from 'bkui-vue/lib/icon';
import { isArray } from 'lodash';
import { computed, onMounted, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';
import type { VxeTablePropTypes } from 'vxe-table';

import { Table, TableColumn } from '@blueking/table';

import PkgUploadSideslider from '../agent-proxy-pkg/pkg-upload-sideslider.vue';

import type { Release } from '@/@types/common.d';
import { PackageService } from '@/api/modules/pkg';
import { formatTimestamp } from '@/common/util';
import usePage from '@/composables/use-page';
import useTableSetting from '@/composables/use-table-setting';
import { useMainStore } from '@/stores/main';
type PkgType = 'gse_agent' | 'gse_proxy';
type filterProp = 'file_name' | 'operator';
interface IFilterOption {
  list: { value: string | boolean, text: string;  }[];
  checked: string[];
  filterScope: string;
  match?: string,
}

const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const mainStore = useMainStore();
const maxHeight = computed(() => mainStore.windowInnerHeight - 214);
const title = computed(() => (route.name === 'agentPackageMng' ? t('Agent 包管理') : t('Proxy 包管理')));
const currentType = computed(() => {
  const routeName = route.name?.toString() || '';
  const type = routeName.split('PackageMng')[0];
  return type;
});
const isShow = ref(false);
const loading = ref(false);
const packageList = ref<Release[]>([]);
const originPackageList = ref<Release[]>([]);
// 分页
const {
  pagination,
} = usePage(packageList);
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

const filterOptionSource = reactive<Record<string, IFilterOption>>({
  file_name: {
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
// 搜索
const searchSelectValue = ref<{ id: string; name: string; values: any[] }[]>([]);
function getUniqueChildren(prop: string) {
  const res = Array.from(new Set(originPackageList.value
    .map((item: any) => item[prop])
    .filter((item: any) => String(item))));
  const uniqueValues = isArray(res[0]) ? [...res[0]] : res;
  return uniqueValues.map((value: any) => {
    const name = String(value);
    let count;

    return {
      id: value,
      name,
      value,
      text: name,
      count,
    };
  });
}

const searchSelectData = computed(() => [
  {
    id: 'file_name',
    name: t('包文件名'),
    children: getUniqueChildren('file_name'),
    multiple: true,
  },
  {
    id: 'operator',
    name: t('上传用户'),
    children: getUniqueChildren('operator'),
    multiple: true,
  },
]);

// 表格
const { isShowSetting, settings, handleSettingChange } = useTableSetting({
  checked: [
    'file_name',
    'operator',
    'updated_at',
    'action',
  ],
  disabled: ['action'],
}, `pkgMng-${currentType.value}`);
// 筛选
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
      values: checked.map((item: any) => {
        let name;
        switch (field) {
          case 'enabled':
            name = item ? t('启用') : t('禁用');
            break;
          default:
            name = item;
            break;
        }
        return {
          id: item,
          name,
        };
      }),
    });
  }
};
// 搜索
const handleSearchSelectChange = async (data: {id: string, name: string, values: {id: string, name: string}[]}[]) => {
  Object.keys(filterOptionSource).forEach((key) => {
    filterOptionSource[key].checked = [];
  });
  data.forEach((item) => {
    if (filterOptionSource[item.id as filterProp]) {
      filterOptionSource[item.id as filterProp].checked = item.values.map((item: any) => item.id) as string[];
    }
  });
};

const handleBlur = async () => {
  await getPackages();
};
const handleUpload = () => {
  isShow.value = true;
};

const getPackages = async () => {
  loading.value = true;
  const res = await PackageService.ListRelease({
    release_type: currentType.value === 'plugin_bintool' ? 'plugin_bintool_v2' : currentType.value,
    generation: 2,
  }).catch(() => ({
    items: [],
  }));
  const items = res.items.map((item, index) => ({
    ...item,
    labels: item.labels || [],
    isDisabledPopShow: false,
    isDeletePopShow: false,
    isShowTagInput: false,
    createPopShow: false,
  }));
  originPackageList.value = items;
  packageList.value = items;
  loading.value = false;
};
const getParams = (row: Release) => ({
  generation: row.generation,
  release_type: row.release_type,
  platform: {
    os_type: row.os_type,
    cpu_arch: row.cpu_arch,
  },
  version: row.version,
});
const handleDelete = async (row: Release) => {
  await PackageService.DeleteRelease(getParams(row));
  await getPackages();
};
const handleConfirm = async () => {
  await getPackages();
};
watch(originPackageList, () => {
  filterOptionSource.file_name.list = getUniqueChildren('file_name');
  filterOptionSource.operator.list = getUniqueChildren('operator');
}, { immediate: true, deep: true });
// 前端过滤数据
watch(
  [
    searchSelectValue,
    originPackageList,
  ],
  () => {
    packageList.value = originPackageList.value.filter((row: Release) =>
      searchSelectValue.value.every((searchItem: any) => {
        const { id: searchField, values } = searchItem;
        const searchIds = values?.map((value: {id: string}) => value.id);
        if (isArray(row[searchField])) {
          return !!row[searchField].find((el: string) => searchIds.includes(el));
        }
        return searchIds.includes(row[searchField]);
      }));
  },
  { immediate: true, deep: true },
);
watch(
  () => route.name,
  async () => {
    searchSelectValue.value = [];
    await getPackages();
  },
  { immediate: true },
);
</script>
