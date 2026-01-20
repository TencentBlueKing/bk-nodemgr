<template>
  <div class="p-[24px] h-[calc(100%_-_52px)] flex flex-col">
    <!-- 搜索栏 -->
    <div class="flex items-center w-full h-[32px] mb-[16px]">
      <Button theme="primary" @click="handleUpload">包上传</Button>
      <SearchSelect
        class="ml-[16px] flex-1 bg-[#fff]"
        ref="searchSelect"
        :data="searchSelectData"
        v-model.trim="searchSelectValue"
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
        <div class="px-[16px] py-[10px] text-[14px]">快捷筛选</div>
        <div v-for="option in dimensionList" :key="option.id">
          <div
            class="flex justify-between items-center bg-[#F0F1F5] h-[32px] px-[16px] cursor-pointer"
            :class="{ 'bg-[#fff]': option.id === 'version' ? !state.versionExpand : !state.osExpand }"
            @click="handleExpand(option.id)">
            <div class="text-[14px]">{{ option.name }}</div>
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
                    id-key="value"
                    display-key="label"
                    :popover-options="{
                      boundary: 'parent'
                    }"
                    auto-focus
                    allow-create
                    multiple
                    multiple-mode="tag"
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
            :width="180"
          >
            <template #default="{ row }">
              <div class="flex">
                <Button
                  theme="primary"
                  class="mr-[8px]"
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
                    theme="primary"
                    class="mr-[8px]"
                    text
                    v-show="row.enabled"
                  >停用</Button>
                  <template #content>
                    <div class="px-[4px] pt-[8px] pb-[16px]">
                      <div class="text-[16px] text-[#313238] mb-[6px]">
                        确认停用该 {{ currentType === 'agent' ? 'Agent' : 'Proxy' }} 包？
                      </div>
                      <div class="text-[12px] text-[#4D4F56] w-full mb-[5px]">停用目标：{{row.file_name}}</div>
                      <div class="text-[12px] text-[#262830] w-full">
                        停用后，{{ currentType === 'agent' ? 'Agent' : 'Proxy' }} 安装、重装、升级时，不可选择
                      </div>
                    </div>
                  </template>
                </PopConfirm>
                <Button
                  theme="primary"
                  class="mr-[8px]"
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
                        确认删除该 {{ currentType === 'agent' ? 'Agent' : 'Proxy' }} 包？
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
import { Button, Loading, PopConfirm, SearchSelect, Select, Tag, TagInput } from 'bkui-vue';
import { AngleDown, AngleRight, EditLine, TextAll } from 'bkui-vue/lib/icon';
import { debounce, isArray } from 'lodash';
import { computed, onMounted, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';
import type { VxeTablePropTypes } from 'vxe-table';

import { Table, TableColumn } from '@blueking/table';

import PkgUploadSideslider from './pkg-upload-sideslider.vue';

import type { Release } from '@/@types/common.d';
import { PackageService } from '@/api/modules/pkg';
import { compareVersions, formatTimestamp  } from '@/common/util';
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
const downloadUrl = computed(() => `${location.origin}/api/v3/package/release/${currentType.value}/download`);
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
  pageConf,
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
  const uniqueValues = prop === 'labels' ? [...new Set(res.flat())] : res;
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
    'download',
  ],
  disabled: ['action', 'download'],
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
  // 当表头筛选版本和系统架构时，快捷筛选选择全部
  if (field === 'version') {
    selectDimensionOptional(null, 'version');
  }
  if (field === 'os_type' || field === 'cpu_arch') {
    selectDimensionOptional(null, 'os_cpu_arch');
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
const selectDimensionOptional = (id: string | null, dimension: PkgQuickType, type?: string) => {
  // 维度映射到状态属性
  const stateProperty = dimension === 'version' ? 'versionDimensionOptional' : 'osDimensionOptional';
  const targetSet = state[stateProperty]; // 提取目标集合，减少重复访问

  // 处理筛选逻辑：清空并失去焦点
  if (id === null) {
    targetSet.clear();
    return;
  }

  // 处理全选逻辑：清空并仅保留'all'
  if (id === 'all') {
    targetSet.clear();
    targetSet.add('all');
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
    filterDatas.forEach((item) => {
      searchSelectValue.value.push({
        id: item.id,
        name: item.name,
        values: item.children.filter((child) => {
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

// 根据类型获取对应的服务方法
const getServiceMethod = (baseName: string) => {
  const suffix = currentType.value === 'agent' ? 'Agent' : 'Proxy';
  return PackageService[`${baseName}${suffix}` as keyof typeof PackageService];
};

// 标签信息批量编辑
const selectTag = ref<string[]>([]);
const batchUpdateTag = async () => {
  const serviceMethod = currentType.value === 'agent' ? PackageService.SetReleaseAgentLabelsMany : PackageService.SetReleaseProxyLabelsMany;
  await serviceMethod({
    exact_include_conditions: {
      platform: packageList.value.map(item => item.platform),
      version: packageList.value.map(item => item.version),
    },
    generation: 2,
    labels: [...Array.from(new Set(selectTag.value))],
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
  if (!row.host) return;
  router.push({
    name: row.release_type,
    query: {
      os_type: row.os_type,
      cpu_arch: row.cpu_arch,
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

const fetchCurrentPageCounts = async () => {
  // 1. 根据当前分页计算索引范围
  const { current, limit } = pageConf;
  const start = (current - 1) * limit;
  const end = start + limit;

  // 2. 切片获取当前页显示的那些行对象
  // 注意：slice 返回的是对象的引用，修改 currentViewItems 中的 item 会直接更新 packageList
  const currentViewItems = packageList.value.slice(start, end);

  if (!currentViewItems.length) return;

  // 3. 准备 API 参数
  const countApi = currentType.value === 'agent'
    ? PackageService.CountDeployedReleasedAgent
    : PackageService.CountDeployedReleasedProxy;

  // 构造请求体，只包含当前页的包信息
  const requestItems = currentViewItems.map(item => ({
    generation: item.generation,
    version: item.version,
    platform: item.platform,
  }));

  try {
    // 4. 调用接口获取数量
    const countData = await countApi({ items: requestItems });

    // ============================================
    // 【关键点】：在这里给列表数据赋值 host 数量
    // ============================================
    currentViewItems.forEach((item, index) => {
      // countData.counts 的顺序与 requestItems 一致
      // 直接修改属性，Vue 的响应式系统会自动更新表格 DOM
      item.host = countData.counts[index] || 0;
    });
  } catch (error) {
    console.error('获取部署数量失败', error);
  }
};

const getPackages = async () => {
  loading.value = true;

  try {
    // 并行执行API调用
    const [listApi, countApi] = currentType.value === 'agent'
      ? [PackageService.ListReleaseAgent, PackageService.CountDeployedReleasedAgent]
      : [PackageService.ListReleaseProxy, PackageService.CountDeployedReleasedProxy];

    // 先获取列表数据
    const listData = await listApi({
      page: { limit: 500, offset: 0 },
      generation: 2,
      exact_include_conditions: {},
    }).catch(() => ({ total: 0, items: [] }));

    // 处理标签数据
    const allLabels = listData.items.flatMap(item => item.release.labels || []);
    const uniqueLabels = Array.from(new Set(allLabels));

    tagList.value = uniqueLabels.map((tag: string) => ({
      value: tag,
      label: tag,
    }));

    packageStore.updateTagList(uniqueLabels);

    // 处理包列表数据
    const processedItems = listData.items.map((item, index) => ({
      ...item.release,
      platform: {
        os_type: item.release.os_type,
        cpu_arch: item.release.cpu_arch,
      },
      labels: item.release.labels || [],
      os_cpu_arch: `${item.release.os_type}_${item.release.cpu_arch}`,
      host: 0,
      isShowTagInput: false,
      createPopShow: false,
    })).sort((a, b) => compareVersions(a.version, b.version));

    originPackageList.value = processedItems;
    packageList.value = processedItems;
    fetchCurrentPageCounts();
  } catch (error) {
    console.error('获取包列表失败:', error);
    // 确保在错误情况下也清空数据
    originPackageList.value = [];
    packageList.value = [];
    tagList.value = [];
  } finally {
    loading.value = false;
  }
};
const getParams = (row: Release) => ({
  generation: row.generation,
  platform: {
    os_type: row.os_type,
    cpu_arch: row.cpu_arch,
  },
  version: row.version,
});

// 通用操作处理函数
const handleOperation = async (row: Release, serviceMethod: (params: any) => Promise<any>) => {
  try {
    await serviceMethod(getParams(row));
    await getPackages();
  } catch (error) {
    console.error('操作失败:', error);
    // 可以在这里添加错误提示或重试逻辑
  }
};

// 简化的操作函数
const handleSetDefaultVersion = async (row: Release) => {
  await handleOperation(row, getServiceMethod('SetAsDefaultRelease'));
};

const handleCancelAsDefaultVersion = async (row: Release) => {
  await handleOperation(row, getServiceMethod('CancelAsDefaultRelease'));
};

const handleDisabled = async (row: Release) => {
  await handleOperation(row, getServiceMethod('DisableRelease'));
};

const handleEnabled = async (row: Release) => {
  await handleOperation(row, getServiceMethod('EnableRelease'));
};

const handleDelete = async (row: Release) => {
  await handleOperation(row, getServiceMethod('DeleteRelease'));
};
const handleConfirm = async () => {
  await getPackages();
};

const debounceFetchCurrentPageCounts = debounce(fetchCurrentPageCounts, 300);
watch(
  [
    () => pageConf.current,  // 页码变了
    () => pageConf.limit,    // 每页条数变了
  ],
  () => {
    // 只要视图变化，就重新获取当前视图内的 host 数量
    debounceFetchCurrentPageCounts();
  },
);
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
    packageList.value = originPackageList.value.filter((row: Release) => searchSelectValue.value.every((searchItem: any) => {
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
