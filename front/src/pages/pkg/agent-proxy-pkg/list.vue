<template>
  <div class="p-[24px] h-[calc(100%_-_52px)] flex flex-col">
    <!-- 搜索栏 -->
    <div class="flex items-center w-full h-[32px] mb-[16px]">
      <Button theme="primary" @click="handleUpload">包上传</Button>
      <SearchSelect
        class="ml-[16px] flex-1"
        ref="searchSelect"
        :data="searchSelectData"
        v-model="searchSelectValue"
        :unique-select="true"
        :placeholder="'版本号、操作系统、架构、标签、上传用户、状态、默认版本'"
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
            :class="{ 'bg-[#fff]': option.id === 'version' ? !state.versionExpand : !state.osExpand }"
            @click="handleExpand(option.id)">
            <div class="text-[12px]">{{ option.name }}</div>
            <angle-down v-show="option.id === 'version' ? state.versionExpand : state.osExpand" />
            <angle-right v-show="option.id === 'version' ? !state.versionExpand : !state.osExpand" />
          </div>
          <template v-if="option.id === 'version' ? state.versionExpand : state.osExpand">
            <div
              :class="[
                'h-[36px] mt-[8px] px-[16px] cursor-pointer',
                {
                  'bg-[#E1ECFF] text-[#3A84FF]':
                    option.id === 'version'
                      ? state.versionDimensionOptional.has('all')
                      : state.osDimensionOptional.has('all')
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
                  :checked="option.id === 'version'
                    ? state.versionDimensionOptional.has('all')
                    : state.osDimensionOptional.has('all')"
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
                    'bg-[#E1ECFF] text-[#3A84FF]':
                      option.id === 'version'
                        ? state.versionDimensionOptional.has(item.id)
                        : state.osDimensionOptional.has(item.id),
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
                  :checked="option.id === 'version'
                    ? state.versionDimensionOptional.has(item.id)
                    : state.osDimensionOptional.has(item.id)"
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
          class="w-full"
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
            field="version"
            :filter="filterOptionSource.version"
            :title="'版本号'"
            :min-width="130"
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
            field="labels"
            :title="'标签信息'"
            :min-width="180"
            :filter="filterOptionSource.labels"
          >
            <template #header>
              <span>标签信息</span>
              <PopConfirm
                width="320"
                theme="light"
                trigger="click"
                title="批量编辑标签"
                @confirm="batchUpdateTag"
              >
                <Button class="mx-[3px]" text :disabled="!packageList.length">
                  <i class="nodeman-icon nc-edit text-[18px]  cursor-pointer"></i>
                </Button>
                <template #content>
                  <div class="text-[12px] text-[#4D4F56] mb-[6px]">统一填充</div>
                  <Select
                    class="mb-[18px]"
                    v-model="selectTag"
                    :list="tagList"
                    :popover-options="{
                      boundary: 'parent'
                    }"
                    auto-focus
                    multiple
                    filterable>
                  </Select>
                </template>
              </PopConfirm>
            </template>
            <template #default="{ row }">
              <div v-show="!row.isShowTagInput" class="flex items-center group gap-[5px]">
                <create-tag :data="row" @blur="handleBlur"></create-tag>
              </div>
            </template>
          </TableColumn>
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
            field="host"
            :title="'已部署主机'"
            :min-width="120"
            sortable
          >
            <template #default="{ row }">
              <Button text theme="primary" @click="handleClickHost(row)">
                {{ row.host }}
              </Button>
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
            :min-width="120"
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
                <Popover
                  width="340"
                  theme="light"
                  trigger="click"
                  placement="top-start"
                  :is-show="row.isDisabledPopShow"
                >
                  <Button
                    class="mr-[8px]"
                    theme="primary"
                    text
                    v-show="row.enabled"
                    @click="row.isDisabledPopShow = true"
                  >停用</Button>
                  <template #content>
                    <div class="px-[4px] pt-[8px] pb-[4px]">
                      <div class="text-[16px] text-[#313238] mb-[6px]">
                        确认停用该 {{ currentType === 'agent' ? 'Agent' : 'Proxy' }} 包？
                      </div>
                      <div class="text-[12px] text-[#4D4F56] w-full mb-[5px]">停用目标：{{row.file_name}}</div>
                      <div class="text-[12px] text-[#262830] w-full">停用后，Agent 安装、重装、升级时，不可选择</div>
                      <div class="mt-[20px] flex gap-[8px] justify-end">
                        <Button theme="primary" @click="handleDisabled(row)">停用</Button>
                        <Button @click="row.isDisabledPopShow = false">取消</Button>
                      </div>
                    </div>
                  </template>
                </Popover>
                <Button
                  class="mr-[8px]"
                  theme="primary"
                  text
                  v-if="!row.enabled"
                  @click="handleEnabled(row)"
                >启用</Button>
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
                        确认删除该 {{ currentType === 'agent' ? 'Agent' : 'Proxy' }} 包？
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
  </div>
  <pkg-upload-sideslider v-model:is-show="isShow" @confirm="handleConfirm" />
</template>
<script lang="ts" setup>
import { Button, Loading, PopConfirm, Popover, SearchSelect, Select, Tag, TagInput } from 'bkui-vue';
import { AngleDown, AngleRight, EditLine, TextAll } from 'bkui-vue/lib/icon';
import { isArray } from 'lodash';
import { computed, onMounted, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';
import type { VxeTablePropTypes } from 'vxe-table';

import { Table, TableColumn } from '@blueking/table';

import PkgUploadSideslider from './pkg-upload-sideslider.vue';

import type { Release } from '@/@types/common.d';
import { PackageService } from '@/api/modules/pkg';
import { capitalizeFirstLetter, compareVersions, formatTimestamp } from '@/common/util';
import usePage from '@/composables/use-page';
import useTableSetting from '@/composables/use-table-setting';
import { useMainStore } from '@/stores/main';
import { usePackageStore } from '@/stores/package';
type PkgQuickType = 'os_cpu_arch' | 'version';
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
  list: { value: string | boolean, text: string;  }[];
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
const packageStore = usePackageStore();
const maxHeight = computed(() => mainStore.windowInnerHeight - 214);
const quickMaxHeight = computed(() => mainStore.windowInnerHeight - 314);
const currentType = computed(() => (route.name === 'agentPackageMng' ? 'agent' : 'proxy'));
const isShow = ref(false);
const loading = ref(false);
const packageList = ref<Release[]>([]);
const originPackageList = ref<Release[]>([]);
const state = reactive<{
  isLoading: boolean;
  panels: { name: PkgType; label: string }[];
  active: PkgType;
  dimension: PkgQuickType;
  versionDimensionOptional: Set<string>;
  osDimensionOptional: Set<string>;
  versionExpand: boolean;
  osExpand: boolean;
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
  versionDimensionOptional: new Set(['all']),
  osDimensionOptional: new Set(['all']),
  versionExpand: true,
  osExpand: true,
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
const dimensionList = computed(() => [
  {
    id: 'os_cpu_arch',
    name: '操作系统/架构',
    multiple: true,
    children: getUniqueChildren('os_cpu_arch'),
  },
  {
    id: 'version',
    name: '版本号',
    multiple: true,
    children: getUniqueChildren('version'),
  },
]);
const handleExpand = (id: string) => {
  if (id === 'version') {
    state.versionExpand = !state.versionExpand;
  } else {
    state.osExpand = !state.osExpand;
  }
};
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
  labels: {
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
    id: 'labels',
    name: '标签',
    children: getUniqueChildren('labels'),
    multiple: true,
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
    'version',
    'os_type',
    'cpu_arch',
    'labels',
    'operator',
    'updated_at',
    'host',
    'enabled',
    'as_default',
    'action',
  ],
  disabled: ['action'],
});
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
  if (field === 'version') {
    selectDimensionOptional('all', 'version');
  }
  if (field === 'os_type' || field === 'cpu_arch') {
    selectDimensionOptional('all', 'os_cpu_arch');
  }
};
// 搜索
const handleSearchSelectChange = async (data: {id: string, name: string, values: {id: string, name: string}[]}[]) => {
  if (data.findIndex(item => item.id === 'version') === -1) {
    selectDimensionOptional('all', 'version');
  }
  if (data.findIndex(item => item.id === 'os_type') === -1 || data.findIndex(item => item.id === 'cpu_arch') === -1) {
    selectDimensionOptional('all', 'os_cpu_arch');
  }
};
// 维度nav click 可以多选, 如果选择非all，则all取消选中状态，如果空了则选择all
const selectDimensionOptional = (id: string, dimension: PkgQuickType, type?: string) => {
  // 维度映射到状态属性
  const stateProperty = dimension === 'version' ? 'versionDimensionOptional' : 'osDimensionOptional';
  const targetSet = state[stateProperty]; // 提取目标集合，减少重复访问

  // 处理全选逻辑：清空并仅保留'all'
  if (id === 'all') {
    targetSet.clear();
    targetSet.add('all');
    return; // 全选后无需执行后续逻辑
  }

  // 若当前有全选状态，先取消全选（避免同时存在'all'和其他选项）
  if (targetSet.has('all')) {
    targetSet.delete('all');
  }

  // 切换当前选项的选中状态（存在则删除，不存在则添加）
  if (targetSet.has(id)) {
    targetSet.delete(id);
  } else {
    targetSet.add(id);
  }

  // 若所有选项都被取消，自动选中全选
  if (targetSet.size === 0) {
    targetSet.add('all');
  }
  if (type === 'click') {
    updateQuickOptToSearch(targetSet, dimension);
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
// 标签信息批量编辑
const selectTag = ref<string[]>([]);
const batchUpdateTag = async () => {
  await PackageService.SetReleaseLabelsMany({
    release_type: currentType.value,
    identify: packageList.value.map(item => ({
      generation: item.generation,
      platform: {
        os_type: item.os_type,
        cpu_arch: item.cpu_arch,
      },
      version: item.version,
    })),
    labels: [...selectTag.value],
  });
  await getPackages();
};
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
const handleClickHost = (row: Release) => {
  if (!row.host || row.release_type !== 'agent') return;
  router.push({
    name: 'agent',
    query: {
      os_type: row.os_type,
      node_version: row.version,
    },
  });
};
const handleBlur = async () => {
  await getPackages();
};
const handleUpload = () => {
  isShow.value = true;
};
const tagList = ref<{value: string, label: string}[]>([]);
const getPackages = async () => {
  loading.value = true;
  const res = await PackageService.ListRelease({
    release_type: currentType.value,
    generation: 2,
  });
  const allLabels = res.items.flatMap(item => item.labels || []);
  const list = Array.from(new Set(allLabels));
  tagList.value = list.map((tag: string) => ({
    value: tag,
    label: tag,
  }));
  packageStore.updateTagList(list);
  const hostList = await PackageService.DeployedHostCount({
    request_items: res.items.map(item => ({
      generation: item.generation,
      release_type: item.release_type,
      version: item.version,
      platform: {
        os_type: item.os_type,
        cpu_arch: item.cpu_arch,
      },
    })),
  });
  const items = res.items.map((item, index) => ({
    ...item,
    labels: item.labels || [],
    os_cpu_arch: `${item.os_type}_${item.cpu_arch}`,
    host: hostList.items[index],
    isDisabledPopShow: false,
    isDeletePopShow: false,
    isShowTagInput: false,
    createPopShow: false,
  }))
    .sort((a, b) => compareVersions(a.version, b.version));
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
const handleSetDefaultVersion = async (row: Release) => {
  await PackageService.SetAsDefaultRelease(getParams(row));
  await getPackages();
};
const handleDisabled = async (row: Release) => {
  await PackageService.DisableRelease(getParams(row));
  await getPackages();
};
const handleEnabled = async (row: Release) => {
  await PackageService.EnableRelease(getParams(row));
  await getPackages();
};
const handleDelete = async (row: Release) => {
  await PackageService.DeleteRelease(getParams(row));
  await getPackages();
};
const handleConfirm = async () => {
  await getPackages();
};
watch(originPackageList, () => {
  filterOptionSource.version.list = getUniqueChildren('version');
  filterOptionSource.os_type.list = getUniqueChildren('os_type');
  filterOptionSource.cpu_arch.list = getUniqueChildren('cpu_arch');
  filterOptionSource.labels.list = getUniqueChildren('labels');
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
    filterOptionSource.version.checked = [];
    filterOptionSource.labels.checked = [];
    filterOptionSource.operator.checked = [];
    filterOptionSource.enabled.checked = [];
    await getPackages();
  },
  { immediate: true },
);
onMounted(async () => {
});
</script>
