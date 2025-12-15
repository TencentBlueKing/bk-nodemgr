<template>
  <page-header :title="'插件包管理'" :back="!!route.query?.name"></page-header>
  <div class="p-[24px] h-[calc(100%_-_52px)] flex flex-col">
    <!-- 搜索栏 -->
    <div class="flex items-center w-full h-[32px] mb-[16px]">
      <Button theme="primary" @click="handleUpload">
        <span>包上传</span>
      </Button>
      <SearchSelect
        class="ml-[16px] flex-1"
        ref="searchSelect"
        :data="searchSelectData"
        v-model.trim="searchSelectValue"
        :unique-select="true"
        :placeholder="'请输入 插件包名、版本号、操作系统、架构、上传用户、状态、默认版本 搜索'"
        @update:model-value="handleSearchSelectChange"
      >
      </SearchSelect>
    </div>
    <div class="flex flex-1 w-full">
      <div
        class="w-[240px] flex-shrink-0 bg-[#fff] rounded-[2px] shadow-[0_2px_4px_#1919290d] h-full mr-[17px]"
      >
        <div class="px-[16px] py-[10px] text-[12px]">快捷筛选</div>
        <div v-for="option in dimensionList" :key="option.id">
          <div
            class="flex justify-between items-center bg-[#F0F1F5] h-[32px] px-[16px] cursor-pointer"
            :class="{ 'bg-[#fff]': !option.expand }"
            @click="handleExpand(option)">
            <div class="text-[12px]">{{ option.name }}</div>
            <angle-down v-show="option.expand" />
            <angle-right v-show="!option.expand" />
          </div>
          <template v-if="option.expand">
            <div
              :class="[
                'h-[36px] mt-[8px] px-[16px] cursor-pointer',
                {
                  'bg-[#E1ECFF] text-[#3A84FF]': option.optionalSet.has('all')
                },
              ]"
              @click="selectDimensionOptional('all', option.id as PkgQuickType, 'click')"
            >
              <div class="border-b flex justify-between items-center h-[36px]">
                <div class="flex items-center text-[13px]">
                  <text-all class="mr-[5px]" />
                  <span>全部</span>
                </div>
                <Tag
                  size="large"
                  checkable
                  :checked="option.optionalSet.has('all')"
                >{{ originPackageList.length }}</Tag
                >
              </div>
            </div>
            <div class="overflow-auto" :style="{ maxHeight: quickMaxHeight / 2 + 'px' }">
              <div
                v-for="item in option.children"
                :key="item.id"
                :class="[
                  'flex justify-between items-center h-[36px] cursor-pointer px-[16px]',
                  {
                    'bg-[#E1ECFF] text-[#3A84FF]': option.optionalSet.has(item.id),
                  },
                ]"
                @click="selectDimensionOptional(item.id, option.id as PkgQuickType, 'click')"
              >
                <div>
                  <i v-if="item.id.includes('darwin')" class="nodeman-icon nc-macos mr-[5px]"></i>
                  <i v-else :class="`nodeman-icon nc-${item.id.split('_')[0]} mr-[5px]`"></i>
                  <span class="text-[12px]">{{ `${item.name}` }}</span>
                </div>
                <Tag
                  v-if="item.count"
                  size="large"
                  checkable
                  :checked="option.optionalSet.has(item.id)"
                >
                  {{ item.count }}
                </Tag>
              </div>
            </div>
          </template>
        </div>
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
            field="name"
            :title="'插件包名'"
            :min-width="150"
            fixed="left"
            show-overflow="tooltip"
          ></TableColumn>
          <TableColumn
            field="file_name"
            :title="'包文件名'"
            :min-width="320"
            fixed="left"
            show-overflow="tooltip"
          ></TableColumn>
          <TableColumn
            field="version"
            :filter="filterOptionSource.version"
            :title="'版本号'"
            :min-width="180"
            sortable
            show-overflow="tooltip"
          ></TableColumn>
          <TableColumn
            field="os_type"
            :title="'操作系统'"
            :filter="filterOptionSource.os_type"
            :min-width="110"
          ></TableColumn>
          <TableColumn
            field="cpu_arch"
            :title="'架构'"
            :filter="filterOptionSource.cpu_arch"
            :min-width="80"
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
            field="enabled"
            :title="'状态'"
            :min-width="120"
            :filter="filterOptionSource.enabled"
          >
            <template #default="{ row }">
              <Tag v-if="row.enabled" theme="success">启用</Tag>
              <Tag v-else>禁用</Tag>
            </template>
          </TableColumn>
          <TableColumn
            field="as_default"
            :title="'默认版本'"
            :min-width="120"
            :filter="filterOptionSource.as_default"
          >
            <template #default="{ row }">
              <Tag v-if="row.as_default" theme="success">是</Tag>
              <Tag v-else>否</Tag>
            </template>
          </TableColumn>
          <TableColumn
            field="action"
            :title="'操作'"
            fixed="right"
            :min-width="180"
          >
            <template #default="{ row }">
              <div class="flex items-center">
                <Button
                  class="mr-[8px]"
                  theme="primary"
                  text
                  v-if="row.enabled && !row.as_default"
                  @click="handleSetDefaultVersion(row)"
                >
                  设为默认版本
                </Button>
                <Button
                  theme="primary"
                  class="mr-[8px]"
                  text
                  v-if="row.enabled && row.as_default"
                  @click="handleCancelAsDefaultVersion(row)"
                >
                  取消设置默认版本
                </Button>
                <PopConfirm
                  theme="light"
                  trigger="click"
                  confirm-text="停用"
                  @confirm="handleDisabled(row)"
                >
                  <Button
                    class="mr-[8px]"
                    theme="primary"
                    text
                    v-show="row.enabled"
                  >停用</Button>
                  <template #content>
                    <div class="px-[4px] pt-[8px] pb-[16px]">
                      <div class="text-[16px] text-[#313238] mb-[6px]">
                        确认停用该插件包？
                      </div>
                      <div class="text-[12px] text-[#4D4F56] w-full mb-[5px]">停用目标：{{row.file_name}}</div>
                      <!-- <div class="text-[12px] text-[#262830] w-full">停用后，Agent 安装、重装、升级时，不可选择</div> -->
                    </div>
                  </template>
                </PopConfirm>
                <Button
                  class="mr-[8px]"
                  theme="primary"
                  text
                  v-if="!row.enabled"
                  @click="handleEnabled(row)"
                >启用</Button>
                <PopConfirm
                  theme="light"
                  trigger="click"
                  confirm-text="删除"
                  @confirm="handleDelete(row)"
                >
                  <Button
                    theme="primary"
                    text
                    v-show="!row.enabled"
                  >删除</Button>
                  <template #content>
                    <div class="px-[4px] pt-[8px] pb-[16px]">
                      <div class="text-[16px] text-[#313238] mb-[6px]">
                        确认删除该插件包？
                      </div>
                      <div class="text-[12px] text-[#4D4F56] w-full mb-[5px]">删除目标：{{row.file_name}}</div>
                      <div class="text-[12px] text-[#4D4F56] w-full">删除后不可恢复，请谨慎操作！</div>
                    </div>
                  </template>
                </PopConfirm>
              </div>
            </template>
          </TableColumn>
          <TableColumn
            field="download"
            title="下载"
            fixed="right"
            :width="60"
          >
            <template #default="{ row }">
              <download-pkg :data="row" :url="downloadUrl">
                <i class="nodeman-icon nc-xiazai"></i>
              </download-pkg>
            </template>
          </TableColumn>
        </Table>
      </Loading>
    </div>
  </div>
  <pkg-upload-sideslider v-model:is-show="isShow" @confirm="handleConfirm" />
</template>
<script lang="ts" setup>
import { Button, Dropdown, Loading, PopConfirm, SearchSelect, Select, Tag, TagInput } from 'bkui-vue';
import { AngleDownLine, AngleRight, EditLine, TextAll } from 'bkui-vue/lib/icon';
import { isArray } from 'lodash';
import type { ComputedRef } from 'vue';
import { computed, onMounted, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';
import type { VxeTablePropTypes } from 'vxe-table';

import { Table, TableColumn } from '@blueking/table';

import PkgUploadSideslider from './agent-proxy-pkg/pkg-upload-sideslider.vue';

import type { Release } from '@/@types/common.d';
import { PackageService } from '@/api/modules/pkg';
import { compareVersions, formatTimestamp } from '@/common/util';
import usePage from '@/composables/use-page';
import useTableSetting from '@/composables/use-table-setting';
import { useMainStore } from '@/stores/main';
type PkgQuickType = 'os_cpu_arch' | 'name';
type PkgType = 'gse_agent' | 'gse_proxy';
type filterProp = 'version' | 'labels' | 'operator' | 'enabled';
interface ISearchSelect {
  id: string;
  name: string;
  children: {
    id: string;
    name: string;
    count: number;
    icon?: string;
    tips?: boolean;
    isAll?: boolean;
  }[];
}
interface IFilterOption {
  list: ComputedRef<{ value: string | boolean, text: string;  }[]> | { value: string | boolean, text: string;  }[];
  checked: string[];
  filterScope: string;
  match?: string,
}
// 排序类型
type PkgOrderType = 'version' | '-version';

const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const mainStore = useMainStore();
const maxHeight = computed(() => mainStore.windowInnerHeight - 214);
const quickMaxHeight = computed(() => mainStore.windowInnerHeight - 314);
const downloadUrl = computed(() => `${location.origin}/api/v3/package/release/plugin/download`);
const isShow = ref(false);
const loading = ref(false);
const packageList = ref<Release[]>([]);
const originPackageList = ref<Release[]>([]);
const state = reactive<{
  isLoading: boolean;
  panels: { name: PkgType; label: string }[];
  active: PkgType;
  dimension: PkgQuickType;
  uploadShow: boolean;
  ordering: PkgOrderType | '';
}>({
  isLoading: true,
  panels: [
    { name: 'gse_agent', label: 'Agent' },
    { name: 'gse_proxy', label: 'Proxy' },
  ],
  active: 'gse_agent',
  // 维度
  dimension: 'os_cpu_arch',
  uploadShow: false,
  ordering: '',
});
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
const dimensionList = ref([
  {
    id: 'os_cpu_arch',
    name: '操作系统/架构',
    multiple: true,
    expand: true,
    optionalSet: new Set(['all']),
    children: computed(() => getUniqueChildren('os_cpu_arch')),
  },
  {
    id: 'name',
    name: '插件包名',
    multiple: true,
    expand: true,
    optionalSet: new Set(['all']),
    children: computed(() => getUniqueChildren('name')),
  },
]);

// 展示包上传类型下拉菜单
const isUploadTypeShow = ref(false);
const handleShowUploadType = () => {
  isUploadTypeShow.value = !isUploadTypeShow.value;
};

const handleExpand = (option: any) => {
  option.expand = !option.osExpand;
};
const filterOptionSource = reactive<Record<string, IFilterOption>>({
  name: {
    list: computed(() => getUniqueChildren('name')),
    checked: [],
    match: 'fuzzy',
    filterScope: 'all',
  },
  version: {
    list: computed(() => getUniqueChildren('version')),
    checked: [],
    match: 'fuzzy',
    filterScope: 'all',
  },
  os_type: {
    list: computed(() => getUniqueChildren('os_type')),
    checked: [],
    match: 'fuzzy',
    filterScope: 'all',
  },
  cpu_arch: {
    list: computed(() => getUniqueChildren('cpu_arch')),
    checked: [],
    match: 'fuzzy',
    filterScope: 'all',
  },
  operator: {
    list: computed(() => getUniqueChildren('operator')),
    checked: [],
    match: 'fuzzy',
    filterScope: 'all',
  },
  enabled: {
    list: [
      { value: true, text: '启用' },
      { value: false, text: '禁用' },
    ],
    checked: [],
    filterScope: 'all',
  },
  as_default: {
    list: [
      { value: true, text: '是' },
      { value: false, text: '否' },
    ],
    checked: [],
    filterScope: 'all',
  },
});
// 搜索
const searchSelectValue = ref<{ id: string; name: string; values: any[] }[]>([]);
function countByProp(data: Release[], prop: string) {
  return data.reduce((acc, item) => {
    const key = item[prop];
    if (acc[key]) {
      acc[key]++;
    } else {
      acc[key] = 1;
    }
    return acc;
  }, {});
}
function getUniqueChildren(prop: string) {
  const res = Array.from(new Set(originPackageList.value
    .map((item: any) => item[prop])
    .filter((item: any) => String(item))));
  const uniqueValues = isArray(res[0]) ? [...res[0]] : res;
  return uniqueValues.map((value: any) => {
    const name = String(value);
    let count;
    if (['os_cpu_arch', 'version'].includes(prop)) {
      count = countByProp(originPackageList.value, prop)[value as string];
    }

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
    id: 'name',
    name: '插件包名',
    multiple: true,
    children: getUniqueChildren('name'),
  },
  {
    id: 'version',
    name: '版本号',
    multiple: true,
    children: getUniqueChildren('version'),
  },
  {
    id: 'os_type',
    name: '操作系统',
    multiple: true,
    children: getUniqueChildren('os_type'),
  },
  {
    id: 'cpu_arch',
    name: '架构',
    multiple: true,
    children: getUniqueChildren('cpu_arch'),
  },
  {
    id: 'operator',
    name: '上传用户',
    children: getUniqueChildren('operator'),
    multiple: true,
  },
  {
    id: 'enabled',
    name: '状态',
    children: [
      { id: true, name: '启用' },
      { id: false, name: '禁用' },
    ],
  },
  {
    id: 'as_default',
    name: '默认版本',
    children: [
      { id: true, name: '是' },
      { id: false, name: '否' },
    ],
  },
]);

// 表格
const { isShowSetting, settings, handleSettingChange } = useTableSetting({
  checked: [
    'name',
    'version',
    'os_type',
    'cpu_arch',
    'operator',
    'updated_at',
    'enabled',
    'as_default',
    'action',
    'download',
  ],
  disabled: ['action', 'download'],
}, 'pkgMng-plugin');
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
  // 当表头筛选版本和系统架构时，快捷筛选选择全部
  if (field === 'os_type' || field === 'cpu_arch') {
    selectDimensionOptional(null, 'os_cpu_arch');
  }
};
// 搜索
const handleSearchSelectChange = async (data: {id: string, name: string, values: {id: string, name: string}[]}[]) => {
  if (data.findIndex(item => item.id === 'os_type') === -1 || data.findIndex(item => item.id === 'cpu_arch') === -1) {
    selectDimensionOptional('all', 'os_cpu_arch');
  }
};
// 维度nav click 可以多选, 如果选择非all，则all取消选中状态，如果空了则选择all
const selectDimensionOptional = (id: string | null, dimension: PkgQuickType, type?: string) => {
  const targetSet = dimensionList.value.find(item => item.id === dimension)?.optionalSet;

  // 处理筛选逻辑：清空并失去焦点
  if (id === null) {
    targetSet?.clear();
    return;
  }
  // 处理全选逻辑：清空并仅保留'all'
  if (id === 'all') {
    targetSet?.clear();
    targetSet?.add('all');
  }

  // 若当前有全选状态，先取消全选（避免同时存在'all'和其他选项）
  if (targetSet?.has('all')) {
    targetSet?.delete('all');
  }

  // 切换当前选项的选中状态（存在则删除，不存在则添加）
  if (targetSet?.has(id)) {
    targetSet?.delete(id);
  } else {
    targetSet?.add(id);
  }

  // 若所有选项都被取消，自动选中全选
  if (targetSet?.size === 0) {
    targetSet?.add('all');
  }
  if (type === 'click') {
    updateQuickOptToSearch(targetSet as Set<string>, dimension);
  }
};
const updateQuickOptToSearch = (ids: Set<string>, dimension: PkgQuickType) => {
  if (dimension === 'os_cpu_arch') {
    const index = searchSelectValue.value.findIndex((item: any) => item.id === 'os_type');
    index > -1 && searchSelectValue.value.splice(index, 1);
    const index2 = searchSelectValue.value.findIndex((item: any) => item.id === 'cpu_arch');
    index2 > -1 && searchSelectValue.value.splice(index2, 1);
  } else {
    const index = searchSelectValue.value.findIndex((item: any) => item.id === dimension);
    index > -1 && searchSelectValue.value.splice(index, 1);
  }

  if (ids.has('all')) return;
  if (dimension === 'os_cpu_arch') {
    const filterDatas = searchSelectData.value.filter((item: {id: string}) => ['os_type', 'cpu_arch'].includes(item.id));
    filterDatas.forEach(item => {
      searchSelectValue.value.push({
        id: item.id,
        name: item.name,
        values: item.children.filter(child => {
          for (const item of ids) {
            if (item.includes(child.id)) {
              return item.includes(child.id);
            }
          }
          return false;
        }),
      });
    });
  } else {
    const findData = searchSelectData.value.find((item: {id: string}) => item.id === dimension);
    findData && searchSelectValue.value.push({
      id: dimension,
      name: findData.name,
      values: findData.children.filter((item: any) => ids.has(item.id)),
    });
  }
};

// 包上传
const handleUpload = () => {
  isShow.value = true;
};

const getPackages = async () => {
  loading.value = true;
  const res = await PackageService.ListReleasePlugin({
    generation: 2,
  }).catch(() => ({
    total: 0,
    items: [],
  }));
  const items = res.items.map(item => ({
    ...item.release,
    os_cpu_arch: `${item.release.os_type}_${item.release.cpu_arch}`,
  })).sort((a, b) => compareVersions(a.version, b.version));
  originPackageList.value = items;
  packageList.value = items;
  loading.value = false;
};
const getParams = (row: Release) => ({
  generation: row.generation,
  name: row.name,
  platform: {
    os_type: row.os_type,
    cpu_arch: row.cpu_arch,
  },
  version: row.version,
});
const handleSetDefaultVersion = async (row: Release) => {
  await PackageService.SetAsDefaultReleasePlugin(getParams(row));
  await getPackages();
};
const handleCancelAsDefaultVersion = async (row: Release) => {
  await PackageService.CancelAsDefaultReleasePlugin(getParams(row));
  await getPackages();
};
const handleDisabled = async (row: Release) => {
  await PackageService.DisableReleasePlugin(getParams(row));
  await getPackages();
};
const handleEnabled = async (row: Release) => {
  await PackageService.EnableReleasePlugin(getParams(row));
  await getPackages();
};
const handleDelete = async (row: Release) => {
  await PackageService.DeleteReleasePlugin(getParams(row));
  await getPackages();
};
const handleConfirm = async () => {
  await getPackages();
};
// 前端过滤数据
watch(
  [
    searchSelectValue,
    originPackageList,
  ],
  () => {
    packageList.value = originPackageList.value
      .filter((row: Release) => searchSelectValue.value.every((searchItem: any) => {
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
    filterOptionSource.version.checked = [];
    filterOptionSource.operator.checked = [];
    filterOptionSource.enabled.checked = [];
    await getPackages();
  },
  { immediate: true },
);
watch(() => route.query, () => {
  if (route.query.name) {
    searchSelectValue.value.push({
      id: 'name',
      name: '插件包名',
      values: [{
        id: route.query.name,
        name: route.query.name,
      }],
    });
  }
}, { immediate: true });
watch(() => searchSelectValue.value, (data) => {
  Object.keys(filterOptionSource).forEach((key) => {
    filterOptionSource[key].checked = [];
  });
  data.forEach((item) => {
    if (filterOptionSource[item.id as filterProp]) {
      filterOptionSource[item.id as filterProp].checked = item.values.map((item: any) => item.id) as string[];
    }
  });
}, { immediate: true, deep: true });
onMounted(async () => {
});
</script>
