<template>
  <page-header :title="'pluginPackage.title'" :back="!!route.query?.name"></page-header>
  <div class="p-[24px] h-[calc(100%_-_52px)] flex flex-col">
    <!-- 搜索栏 -->
    <div class="flex items-center w-full h-[32px] mb-[16px]">
      <Button
        theme="primary"
        :class="{ 'unAuthorized': !hasUploadAuth }"
        @click="hasUploadAuth ? handleUpload() : uploadAuthClick($event)"
        @mouseenter="uploadMouseEnter($event, hasUploadAuth)"
        @mousemove="uploadMouseMove($event, hasUploadAuth)"
        @mouseleave="uploadMouseLeave()"
      >
        <span>{{ $t('pluginPackage.upload') }}</span>
      </Button>
      <SearchSelect
        :max-height="240"
        class="ml-[16px] flex-1 bg-[#fff]"
        ref="searchSelect"
        :data="searchSelectData"
        v-model.trim="searchSelectValue"
        :unique-select="true"
        :placeholder="$t('pluginPackage.searchPlaceholder')"
        @update:model-value="handleSearchSelectChange"
      >
      </SearchSelect>
    </div>
    <!-- 未设置默认版本提醒 -->
    <Alert
      v-if="missingDefaultCombinations.length > 0"
      class="mb-[16px]"
      theme="warning"
      :title="$t('pluginPackage.noDefaultWarning')"
    >
      <template v-for="(combo, i) in missingDefaultCombinations" :key="combo">
        <Tag class="mr-[8px] mb-[4px]">{{ combo }}</Tag>
        <span v-if="i < missingDefaultCombinations.length - 1" class="mr-[4px]">、</span>
      </template>
    </Alert>
    <div class="flex flex-1 w-full">
      <div
        class="w-[240px] flex-shrink-0 bg-[#fff] rounded-[2px] shadow-[0_2px_4px_#1919290d] h-full mr-[17px]"
      >
        <div class="px-[16px] py-[10px] text-[14px]">{{ $t('pluginPackage.quickFilter') }}</div>
        <div v-for="option in dimensionList" :key="option.id">
          <div
            class="flex justify-between items-center bg-[#F0F1F5] h-[32px] px-[16px] cursor-pointer"
            :class="{ 'bg-[#fff]': !option.expand }"
            @click="handleExpand(option)">
            <div class="text-[14px]">{{ option.name }}</div>
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
                  <span>{{ $t('pluginPackage.all') }}</span>
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
                    'unAuthorized': option.id === 'name' && !isNameAuthorized(item.id),
                  },
                ]"
                @click="option.id === 'name' && !isNameAuthorized(item.id)
                  ? sidebarNameClick($event, item.id)
                  : selectDimensionOptional(item.id, option.id as PkgQuickType, 'click')"
                @mouseenter="(e) => option.id === 'name' && sidebarNameMouseEnter(e, item.id)"
                @mousemove="(e) => option.id === 'name' && sidebarNameMouseMove(e, item.id)"
                @mouseleave="() => option.id === 'name' && sidebarNameMouseLeave()"
              >
                <div>
                  <i v-if="item.id.includes('darwin')" class="nodeman-icon nc-macos mr-[5px]"></i>
                  <i v-else-if="option.id === 'name'" class="nodeman-icon nc-plug-circle mr-[5px]"></i>
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
      <NoPermission
        v-if="allPluginsUnauthorized"
        type="plugin"
        class="flex-1"
      />
      <Loading
        v-else
        :title="$t('agentStrategy.loading')"
        :loading="loading"
        class="flex-1 overflow-auto"
      >
        <Table
          class="w-full filterTable"
          :max-height="maxHeight"
          :data="packageList"
          :empty-text="$t('table.empty')"
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
            :title="$t('pluginPackage.packageName')"
            :min-width="150"
            fixed="left"
            show-overflow="tooltip"
          ></TableColumn>
          <TableColumn
            field="file_name"
            :title="$t('pluginPackage.fileName')"
            :min-width="320"
            fixed="left"
            show-overflow="tooltip"
          ></TableColumn>
          <TableColumn
            field="version"
            :filter="filterOptionSource.version"
            :title="$t('pluginPackage.version')"
            :min-width="180"
            sortable
            show-overflow="tooltip"
          ></TableColumn>
          <TableColumn
            field="os_type"
            :title="$t('pluginPackage.os')"
            :filter="filterOptionSource.os_type"
            :min-width="110"
          ></TableColumn>
          <TableColumn
            field="cpu_arch"
            :title="$t('pluginPackage.arch')"
            :filter="filterOptionSource.cpu_arch"
            :min-width="80"
          ></TableColumn>
          <TableColumn
            field="operator"
            :title="$t('pluginPackage.uploader')"
            :min-width="120"
            :filter="filterOptionSource.operator"
          >
            <template #default="{ row }">
              <UserNameDisplay :name="row.operator" />
            </template>
          </TableColumn>
          <TableColumn
            field="updated_at"
            :title="$t('pluginPackage.uploadTime')"
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
            :title="$t('pluginPackage.status')"
            :min-width="80"
            :filter="filterOptionSource.enabled"
          >
            <template #default="{ row }">
              <Tag v-if="row.enabled" theme="success">{{ $t('pluginPackage.enabled') }}</Tag>
              <Tag v-else>{{ $t('pluginPackage.disabled') }}</Tag>
            </template>
          </TableColumn>
          <TableColumn
            field="is_hidden"
            :title="$t('pluginPackage.hiddenStatus')"
            :min-width="80"
            :filter="filterOptionSource.is_hidden"
          >
            <template #default="{ row }">
              <Tag v-if="row.is_hidden" theme="warning">{{ $t('pluginPackage.hidden') }}</Tag>
              <Tag v-else theme="success">{{ $t('pluginPackage.visible') }}</Tag>
            </template>
          </TableColumn>
          <TableColumn
            v-if="isMultipleTenant"
            field="is_synced"
            :title="$t('pluginPackage.source')"
            :min-width="150"
          >
            <template #default="{ row }">
              <Tag v-if="row.is_synced" theme="info">{{ $t('pluginPackage.syncedFromSystem') }}</Tag>
              <span v-else>--</span>
            </template>
          </TableColumn>
          <TableColumn
            field="as_default"
            :title="$t('pluginPackage.defaultVersion')"
            :min-width="120"
            :filter="filterOptionSource.as_default"
          >
            <template #default="{ row }">
              <Tag v-if="row.as_default" theme="success">{{ $t('pluginPackage.yes') }}</Tag>
              <Tag v-else>{{ $t('pluginPackage.no') }}</Tag>
            </template>
          </TableColumn>
          <TableColumn
            field="action"
            :title="$t('pluginPackage.action')"
            fixed="right"
            :width="200"
          >
            <template #default="{ row }">
              <div class="flex items-center">
                <Button
                  class="mr-[8px]"
                  theme="primary"
                  text
                  v-if="row.enabled && !row.as_default"
                  :class="{ 'unAuthorized': !hasManageAuth }"
                  @click="hasManageAuth ? handleSetDefaultVersion(row) : manageAuthClick($event, row.name)"
                  @mouseenter="manageMouseEnter($event, hasManageAuth)"
                  @mousemove="manageMouseMove($event, hasManageAuth)"
                  @mouseleave="manageMouseLeave()"
                >
                  {{ $t('pluginPackage.setDefault') }}
                </Button>
                <Button
                  theme="primary"
                  class="mr-[8px]"
                  text
                  v-if="row.enabled && row.as_default"
                  :class="{ 'unAuthorized': !hasManageAuth }"
                  @click="hasManageAuth ? handleCancelAsDefaultVersion(row) : manageAuthClick($event, row.name)"
                  @mouseenter="manageMouseEnter($event, hasManageAuth)"
                  @mousemove="manageMouseMove($event, hasManageAuth)"
                  @mouseleave="manageMouseLeave()"
                >
                  {{ $t('pluginPackage.cancelDefault') }}
                </Button>
                <PopConfirm
                  theme="light"
                  trigger="click"
                  :confirm-text="$t('pluginPackage.disable')"
                  @confirm="handleDisabled(row)"
                >
                  <Button
                    class="mr-[8px]"
                    theme="primary"
                    text
                    v-show="row.enabled && !row.is_synced"
                    :class="{ 'unAuthorized': !hasManageAuth }"
                    @click="!hasManageAuth && manageAuthClick($event, row.name)"
                    @mouseenter="manageMouseEnter($event, hasManageAuth)"
                    @mousemove="manageMouseMove($event, hasManageAuth)"
                    @mouseleave="manageMouseLeave()"
                  >{{ $t('pluginPackage.disable') }}</Button>
                  <template #content>
                    <div class="px-[4px] pt-[8px] pb-[16px]">
                      <div class="text-[16px] text-[#313238] mb-[6px]">
                        {{ $t('pluginPackage.confirmDisable') }}
                      </div>
                      <div class="text-[12px] text-[#4D4F56] w-full mb-[5px]">
                        {{ $t('pluginPackage.disableTarget', { name: row.file_name }) }}
                      </div>
                      <!-- <div class="text-[12px] text-[#262830] w-full">停用后，Agent 安装、重装、升级时，不可选择</div> -->
                    </div>
                  </template>
                </PopConfirm>
                <Button
                  class="mr-[8px]"
                  theme="primary"
                  text
                  v-if="!row.enabled && !row.is_synced"
                  :class="{ 'unAuthorized': !hasManageAuth }"
                  @click="hasManageAuth ? handleEnable(row) : manageAuthClick($event, row.name)"
                  @mouseenter="manageMouseEnter($event, hasManageAuth)"
                  @mousemove="manageMouseMove($event, hasManageAuth)"
                  @mouseleave="manageMouseLeave()"
                >{{ $t('pluginPackage.enable') }}</Button>
                <PopConfirm
                  theme="light"
                  trigger="click"
                  :confirm-text="$t('pluginPackage.delete')"
                  @confirm="handleDelete(row)"
                >
                  <Button
                    theme="primary"
                    text
                    v-show="!row.enabled && !row.is_synced"
                    :class="{ 'unAuthorized': !hasManageAuth }"
                    @click="!hasManageAuth && manageAuthClick($event, row.name)"
                    @mouseenter="manageMouseEnter($event, hasManageAuth)"
                    @mousemove="manageMouseMove($event, hasManageAuth)"
                    @mouseleave="manageMouseLeave()"
                  >{{ $t('pluginPackage.delete') }}</Button>
                  <template #content>
                    <div class="px-[4px] pt-[8px] pb-[16px]">
                      <div class="text-[16px] text-[#313238] mb-[6px]">
                        {{ $t('pluginPackage.confirmDelete') }}
                      </div>
                      <div class="text-[12px] text-[#4D4F56] w-full mb-[5px]">
                        {{ $t('pluginPackage.deleteTarget', { name: row.file_name }) }}
                      </div>
                      <div class="text-[12px] text-[#4D4F56] w-full">{{ $t('pluginPackage.deleteTip') }}</div>
                    </div>
                  </template>
                </PopConfirm>
              </div>
            </template>
          </TableColumn>
          <TableColumn
            field="download"
            :title="$t('pluginPackage.download')"
            fixed="right"
            :width="downloadLabelWidth"
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
import { Alert, Button, Loading, PopConfirm, SearchSelect, Tag } from 'bkui-vue';
import { AngleDown, AngleRight, TextAll } from 'bkui-vue/lib/icon';
import { debounce, isArray } from 'lodash';
import type { ComputedRef } from 'vue';
import { computed, onMounted, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute } from 'vue-router';
import type { VxeTablePropTypes } from 'vxe-table';

import { Table, TableColumn } from '@blueking/table';

import NoPermission from '@/components/no-permission.vue';
import PkgUploadSideslider from './agent-proxy-pkg/pkg-upload-sideslider.vue';

import type { Release } from '@/@types/common.d';
import useAuthLock from '@/composables/use-auth-lock';
import { PackageService } from '@/api/modules/pkg';
import { PACKAGE_GENERATION } from '@/common/const';
import { getPageAuthorizedItems } from '@/constants/auth';
import { compareVersions, formatTimestamp } from '@/common/util';
import { translateOperatorItems } from '@/common/user-display';
import usePage from '@/composables/use-page';
import useTableSetting from '@/composables/use-table-setting';
import { useAuthStore } from '@/stores/auth';
import { useMainStore } from '@/stores/main';
type PkgQuickType = 'os_cpu_arch' | 'name';
type filterProp = 'version' | 'labels' | 'operator' | 'enabled' | 'is_hidden';
interface IFilterOption {
  list: ComputedRef<{ value: string | boolean, text: string;  }[]> | { value: string | boolean, text: string;  }[];
  checked: string[];
  filterScope: string;
  match?: string,
}

const { t } = useI18n();
const route = useRoute();
const mainStore = useMainStore();
const authStore = useAuthStore();

// 多租户模式（BK_TENANT_MODE === 'multiple'）才展示 来源（is_synced）字段
const isMultipleTenant = window.PROJECT_CONFIG.BK_TENANT_MODE === 'multiple';

// 侧边栏快捷筛选：插件名（来自 DistinctReleasePlugin）+ 当前 hover 名称
const distinctPluginNames = ref<string[]>([]);
const hoveredPluginName = ref<string>();

// 判断当前用户对某个插件包名是否有 package_view 查看权限（决定侧边栏可选/置灰）；
// 管理能力（启用/禁用/删除/设默认）由行内按钮的 package_manage 单独控制
const isNameAuthorized = (name: string) => authStore.hasAuthorizedResource('package_view', name);
// 获取当前用户有权限的插件包名（在 distinctPluginNames 中过滤）
const getAuthorizedPluginNames = () => distinctPluginNames.value.filter(name => isNameAuthorized(name));
// 是否全部插件都无权限（用于切换 NoPermission 占位页）
const allPluginsUnauthorized = computed(() => distinctPluginNames.value.length > 0
  && getAuthorizedPluginNames().length === 0);

// Package permissions
const manageResourceId = ref<string>();
const { hasAuth: hasUploadAuth, handleMouseEnter: uploadMouseEnter, handleMouseMove: uploadMouseMove, handleMouseLeave: uploadMouseLeave, handleAuthClick: uploadAuthClick } = useAuthLock(
  'package_type_upload', () => 'plugin', { resourceType: 'package_type' },
);
const { hasAuth: hasManageAuth, handleMouseEnter: manageMouseEnter, handleMouseMove: manageMouseMove, handleMouseLeave: manageMouseLeave, handleAuthClick: _manageAuthClick } = useAuthLock(
  'package_manage', () => manageResourceId.value, { resourceType: 'package' },
);
const manageAuthClick = (e: MouseEvent, releaseName?: string) => {
  manageResourceId.value = releaseName;
  _manageAuthClick(e);
};

// 侧边栏用独立的 useAuthLock 实例（package_view），与表格行内 package_manage 区分开
// 理由：侧边栏点击是为"查看"该插件包列表申请权限，行内按钮才是"管理"操作
const sidebarViewResourceId = ref<string>();
const { handleMouseEnter: viewMouseEnter, handleMouseMove: viewMouseMove, handleMouseLeave: viewMouseLeave, handleAuthClick: _viewAuthClick } = useAuthLock(
  'package_view', () => sidebarViewResourceId.value, { resourceType: 'package' },
);
const sidebarViewAuthClick = (e: MouseEvent, name: string) => {
  sidebarViewResourceId.value = name;
  _viewAuthClick(e);
};

// 侧边栏无权限插件名：hover 锁图标 + 点击申请权限
const sidebarNameMouseEnter = (e: MouseEvent, name: string) => {
  hoveredPluginName.value = name;
  viewMouseEnter(e, isNameAuthorized(name));
};
const sidebarNameMouseMove = (e: MouseEvent, name: string) => {
  hoveredPluginName.value = name;
  viewMouseMove(e, isNameAuthorized(name));
};
const sidebarNameMouseLeave = () => {
  hoveredPluginName.value = undefined;
  viewMouseLeave();
};
const sidebarNameClick = (e: MouseEvent, name: string) => {
  hoveredPluginName.value = name;
  sidebarViewAuthClick(e, name);
};
const downloadLabelWidth = computed(() => mainStore.curLanguage === 'zh-CN' ? 60 : 100);
const maxHeight = computed(() => mainStore.windowInnerHeight - 214 - (mainStore.noticeShow ? 40 : 0) - (missingDefaultCombinations.value.length > 0 ? 56 : 0));
const quickMaxHeight = computed(() => mainStore.windowInnerHeight - 314 - (mainStore.noticeShow ? 40 : 0) - (missingDefaultCombinations.value.length > 0 ? 56 : 0));
const downloadUrl = computed(() => `${location.origin}/api/v3/package/release/plugin/download`);
// 检查哪些 name+os_type+cpu_arch 组合（有启用包）尚未设置默认版本
const missingDefaultCombinations = computed(() => {
  const enabledItems = originPackageList.value.filter((item: Release) => item.enabled);
  if (enabledItems.length === 0) return [];
  const allCombos = [...new Set(enabledItems.map((item: Release) => `${item.name}_${item.os_type}_${item.cpu_arch}`))];
  const combosWithDefault = new Set(
    enabledItems.filter((item: Release) => item.as_default).map((item: Release) => `${item.name}_${item.os_type}_${item.cpu_arch}`),
  );
  return allCombos.filter(combo => !combosWithDefault.has(combo));
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
    id: 'name',
    name: t('pluginPackage.packageName'),
    multiple: true,
    expand: true,
    optionalSet: new Set(['all']),
    children: computed(() => getUniqueChildren('name')),
  },
  {
    id: 'os_cpu_arch',
    name: t('pluginPackage.osArch'),
    multiple: true,
    expand: true,
    optionalSet: new Set(['all']),
    children: computed(() => getUniqueChildren('os_cpu_arch')),
  },
]);

const handleExpand = (option: any) => {
  option.expand = !option.osExpand;
};
// operator 筛选项：先取原始用户名，再用 bk-user-display-name 异步翻译为显示名
const operatorList = ref<{ id: string; name: string; value: string; text: string; count?: number }[]>([]);
watch(originPackageList, async () => {
  const raw = Array.from(new Set(originPackageList.value
    .map((item: any) => item.operator)
    .filter((item: any) => item)));
  const items = raw.map((id) => ({ id, name: id, value: id, text: id }));
  await translateOperatorItems(items);
  operatorList.value = items;
}, { immediate: true });

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
    list: computed(() => [
      { value: true, text: t('pluginPackage.enabled') },
      { value: false, text: t('pluginPackage.disabled') },
    ]),
    checked: [],
    filterScope: 'all',
  },
  is_hidden: {
    list: computed(() => [
      { value: false, text: t('pluginPackage.visible') },
      { value: true, text: t('pluginPackage.hidden') },
    ]),
    checked: [],
    filterScope: 'all',
  },
  as_default: {
    list: computed(() => [
      { value: true, text: t('pluginPackage.yes') },
      { value: false, text: t('pluginPackage.no') },
    ]),
    checked: [],
    filterScope: 'all',
  },
});
// 搜索
const searchSelectValue = ref<{ id: string; name: string; values: any[] }[]>([]);
const booleanFilterFields = new Set(['enabled', 'is_hidden', 'as_default']);
const normalizeSearchValue = (field: string, value: unknown) => {
  if (booleanFilterFields.has(field) && typeof value === 'string') {
    return value === 'true';
  }
  return value;
};
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
  if (prop === 'operator') return operatorList.value;
  // 插件名维度优先使用 DistinctReleasePlugin 接口返回的全量列表（而非仅 originPackageList），
  // 保证即便后端只返回当前页数据（limit=500）也不会漏掉侧边栏的快捷筛选项
  if (prop === 'name') {
    return [...distinctPluginNames.value]
      .sort((a, b) => a.localeCompare(b))
      .map((value: string) => {
        const count = countByProp(originPackageList.value, 'name')[value] || 0;
        return {
          id: value,
          name: value,
          value,
          text: value,
          count,
        };
      });
  }
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
    name: t('pluginPackage.packageName'),
    multiple: true,
    children: getUniqueChildren('name'),
  },
  {
    id: 'version',
    name: t('pluginPackage.version'),
    multiple: true,
    children: getUniqueChildren('version'),
  },
  {
    id: 'os_type',
    name: t('pluginPackage.os'),
    multiple: true,
    children: getUniqueChildren('os_type'),
  },
  {
    id: 'cpu_arch',
    name: t('pluginPackage.arch'),
    multiple: true,
    children: getUniqueChildren('cpu_arch'),
  },
  {
    id: 'operator',
    name: t('pluginPackage.uploader'),
    children: getUniqueChildren('operator'),
    multiple: true,
  },
  {
    id: 'enabled',
    name: t('pluginPackage.status'),
    children: [
      { id: 'true', name: t('pluginPackage.enabled') },
      { id: 'false', name: t('pluginPackage.disabled') },
    ],
  },
  {
    id: 'is_hidden',
    name: t('pluginPackage.hiddenStatus'),
    children: [
      { id: 'false', name: t('pluginPackage.visible') },
      { id: 'true', name: t('pluginPackage.hidden') },
    ],
  },
  {
    id: 'as_default',
    name: t('pluginPackage.defaultVersion'),
    children: [
      { id: 'true', name: t('pluginPackage.yes') },
      { id: 'false', name: t('pluginPackage.no') },
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
    'is_hidden',
    // 来源字段仅多租户模式存在
    ...(isMultipleTenant ? ['is_synced'] : []),
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
    const filterName = searchSelectData.value.find((item: {id: string}) => item.id === field)?.name ?? t(field);
    searchSelectValue.value.push({
      id: field,
      name: filterName,
      values: checked.map((item: any) => {
        const normalizedItem = normalizeSearchValue(field, item);
        let name;
        switch (field) {
          case 'enabled':
            name = normalizedItem ? t('pluginPackage.enabled') : t('pluginPackage.disabled');
            break;
          case 'is_hidden':
            name = normalizedItem ? t('pluginPackage.hidden') : t('pluginPackage.visible');
            break;
          case 'as_default':
            name = normalizedItem ? t('pluginPackage.yes') : t('pluginPackage.no');
            break;
          default:
            name = item;
            break;
        }
        return {
          id: booleanFilterFields.has(field) ? String(normalizedItem) : item,
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

const getDistinctPluginNames = async () => {
  const res = await PackageService.DistinctReleasePlugin({
    generation: PACKAGE_GENERATION,
    distinct_field: { name: true },
  }).catch((err: any) => {
    console.error('DistinctReleasePlugin error:', err);
    return null;
  });
  // 接口返回结构是 {name: [], os_type: [], cpu_arch: [], version: []}，字段直接在顶层
  distinctPluginNames.value = (res as any)?.name || [];
};

const getPackages = async () => {
  // 等待 package_view 授权加载完成再请求：未加载时拿不到有权限的插件名，
  // 查询条件为空既无意义、也会把无权限插件的短暂全量数据拉回前端
  if (!authStore.authorizedMap['package_view']) return;
  loading.value = true;
  // 仅查询有权限的插件包名对应的数据
  const authorizedNames = getAuthorizedPluginNames();
  const res = await PackageService.ListReleasePlugin({
    page: { limit: 500, offset: 0 },
    generation: PACKAGE_GENERATION,
    exact_include_conditions: { name: authorizedNames },
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
const handleEnable = async (row: Release) => {
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
        const searchIds = values?.map((value: {id: string}) => normalizeSearchValue(searchField, value.id));
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
    await getDistinctPluginNames();
    debounceGetPackages();
  },
  { immediate: true },
);

// 进入页面/权限就绪两个触发点走防抖：竞态时（authorized 恰好在前一次请求期间完成）
// 合并为一次请求，避免重复的列表请求
const debounceGetPackages = debounce(getPackages, 200);

// 权限加载完成 / 授权项变化时重新拉取数据，确保只查有权限的插件名
watch(
  () => authStore.authorizedMap['package_view']?.resourceIds,
  () => {
    if (distinctPluginNames.value.length > 0) {
      debounceGetPackages();
    }
  },
);
watch(() => route.query, () => {
  if (route.query.name) {
    searchSelectValue.value.push({
      id: 'name',
      name: t('pluginPackage.packageName'),
      values: [{
        id: route.query.name as string,
        name: route.query.name as string,
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
      filterOptionSource[item.id as filterProp].checked = item.values
        .map((value: any) => normalizeSearchValue(item.id, value.id)) as string[];
    }
  });
}, { immediate: true, deep: true });
onMounted(async () => {
  // 加载行内管理按钮 / 上传按钮所需的授权数据（package_manage / package_type_upload）；
  // package_view / package_history_view 已由路由守卫的 pkgManager 模块加载
  const items = getPageAuthorizedItems('pluginPackageMng');
  await authStore.fetchAuthorized(items, 'pluginPackageMng').catch(() => {});
});
</script>
