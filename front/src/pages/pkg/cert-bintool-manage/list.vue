<template>
  <NoPermission
    v-if="allPluginBinToolUnauthorized"
    type="action"
    :auth-items="pluginBinToolViewAuthItems"
    :resource-id="PLUGIN_BIN_TOOL_NAMES"
  />
  <div v-else class="p-[24px] h-[calc(100%_-_52px)] flex flex-col">
    <!-- 搜索栏 -->
    <div class="flex items-center w-full h-[32px] mb-[16px]">
      <Button
        theme="primary"
        :class="{ 'unAuthorized': !hasUploadAuth }"
        @click="hasUploadAuth ? handleUpload() : uploadAuthClick($event)"
        @mouseenter="uploadMouseEnter($event, hasUploadAuth)"
        @mousemove="uploadMouseMove($event, hasUploadAuth)"
        @mouseleave="uploadMouseLeave()"
        class="mr-[16px]"
      >{{ t('certBintool.upload') }}</Button>
      <SearchSelect
        :max-height="240"
        class="flex-1 bg-[#fff]"
        ref="searchSelect"
        :data="searchSelectData"
        v-model.trim="searchSelectValue"
        :unique-select="true"
        :placeholder="t('certBintool.searchPlaceholder')"
        @update:model-value="handleSearchSelectChange"
      >
      </SearchSelect>
    </div>
    <Loading
      :title="t('agentStrategy.loading')"
      :loading="loading"
      class="flex-1"
    >
      <Table
        class="w-full filterTable"
        :max-height="maxHeight"
        :data="packageList"
        :empty-text="t('table.empty')"
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
          :title="t('certBintool.fileName')"
          :min-width="320"
          fixed="left"
          show-overflow="tooltip"
        ></TableColumn>
        <TableColumn
          field="operator"
          :title="t('certBintool.uploader')"
          :min-width="120"
          :filter="filterOptionSource.operator"
        >
          <template #default="{ row }">
            <UserNameDisplay :name="row.operator" />
          </template>
        </TableColumn>
        <TableColumn
          field="updated_at"
          :title="t('certBintool.uploadTime')"
          :min-width="180"
          sort-type="number"
          sortable
        >
          <template #default="{ row }">
            {{ formatTimestamp(row.updated_at) }}
          </template>
        </TableColumn>
        <TableColumn
          v-if="isMultipleTenant"
          field="is_synced"
          :title="$t('pluginPackage.source')"
          :min-width="120"
        >
          <template #default="{ row }">
            <Tag v-if="row.is_synced" theme="info">{{ $t('pluginPackage.syncedFromSystem') }}</Tag>
            <span v-else>--</span>
          </template>
        </TableColumn>
        <TableColumn
          field="action"
          :title="t('certBintool.action')"
          fixed="right"
          :width="140"
        >
          <template #default="{ row }">
            <div class="flex items-center">
              <!-- 无权限时不渲染 PopConfirm，避免点击同时弹出确认气泡与权限申请弹窗 -->
              <PopConfirm
                v-if="hasRowManageAuth(row)"
                theme="light"
                trigger="click"
                :confirm-text="t('certBintool.delete')"
                @confirm="handleDelete(row)"
              >
                <Button
                  theme="primary"
                  text
                >{{ t('certBintool.delete') }}</Button>
                <template #content>
                  <div class="px-[4px] pt-[8px] pb-[16px]">
                    <div class="text-[16px] text-[#313238] mb-[6px]">
                      <!-- eslint-disable-next-line max-len -->
                      {{ $t('certBintool.confirmDelete', { type: route.name === "certPackageMng" ? t('certBintool.cert') : t('certBintool.tool') }) }}
                    </div>
                    <div class="text-[12px] text-[#4D4F56] w-full mb-[5px]">
                      {{ t('certBintool.deleteTarget', { name: row.file_name }) }}
                    </div>
                    <div class="text-[12px] text-[#4D4F56] w-full">{{ t('certBintool.deleteTip') }}</div>
                  </div>
                </template>
              </PopConfirm>
              <Button
                v-else
                theme="primary"
                text
                class="unAuthorized"
                @click="manageAuthClick($event, isPluginBinTool ? row.name : row.release_type)"
                @mouseenter="manageMouseEnter($event, hasRowManageAuth(row))"
                @mousemove="manageMouseMove($event, hasRowManageAuth(row))"
                @mouseleave="manageMouseLeave()"
              >{{ t('certBintool.delete') }}</Button>
            </div>
          </template>
        </TableColumn>
        <TableColumn
          field="download"
          :title="t('certBintool.download')"
          fixed="right"
          :width="downloadLabelWidth"
        >
          <template #default="{ row }">
            <download-pkg :data="row" :url="downloadUrl" :current-type="currentType">
              <i class="nodeman-icon nc-xiazai"></i>
            </download-pkg>
          </template>
        </TableColumn>
      </Table>
    </Loading>
  </div>
  <pkg-upload-sideslider v-model:is-show="isShow" @confirm="handleConfirm" />
</template>
<script lang="ts" setup>
// 排序类型
import { Button, Loading, PopConfirm, SearchSelect, Select, Tag, TagInput } from 'bkui-vue';
import { AngleDown, AngleRight, EditLine } from 'bkui-vue/lib/icon';
import { isArray } from 'lodash';
import { computed, onMounted, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute } from 'vue-router';
import type { VxeTablePropTypes } from 'vxe-table';

import { Table, TableColumn } from '@blueking/table';

import PkgUploadSideslider from '../agent-proxy-pkg/pkg-upload-sideslider.vue';

import useAuthLock from '@/composables/use-auth-lock';
import NoPermission from '@/components/no-permission.vue';

import type { Release } from '@/@types/common.d';
import { PackageService } from '@/api/modules/pkg';
import { PACKAGE_GENERATION } from '@/common/const';
import { formatTimestamp } from '@/common/util';
import { translateOperatorItems } from '@/common/user-display';
import usePage from '@/composables/use-page';
import useTableSetting from '@/composables/use-table-setting';
import { useMainStore } from '@/stores/main';
import { useAuthStore } from '@/stores/auth';
type filterProp = 'file_name' | 'operator';
interface IFilterOption {
  list: { value: string | boolean, text: string;  }[];
  checked: string[];
  filterScope: string;
  match?: string,
}

const { t } = useI18n();
const route = useRoute();
const mainStore = useMainStore();

// 多租户模式（BK_TENANT_MODE === 'multiple'）才展示 来源（is_synced）字段
const isMultipleTenant = window.PROJECT_CONFIG.BK_TENANT_MODE === 'multiple';

// Package permissions
const manageResourceId = ref<string>();
const { hasAuth: hasUploadAuth, handleMouseEnter: uploadMouseEnter, handleMouseMove: uploadMouseMove, handleMouseLeave: uploadMouseLeave, handleAuthClick: uploadAuthClick } = useAuthLock(
  'package_type_upload', () => currentType.value, { resourceType: 'package_type' },
);
const { handleMouseEnter: manageMouseEnter, handleMouseMove: manageMouseMove, handleMouseLeave: manageMouseLeave, handleAuthClick: _manageAuthClick } = useAuthLock(
  'package_manage', () => manageResourceId.value, { resourceType: 'package' },
);
const manageAuthClick = (e: MouseEvent, resourceId?: string) => {
  manageResourceId.value = resourceId;
  _manageAuthClick(e);
};
const downloadLabelWidth = computed(() => mainStore.curLanguage === 'zh-CN' ? 60 : 100);
const maxHeight = computed(() => mainStore.windowInnerHeight - 214 - (mainStore.noticeShow ? 40 : 0));
const currentType = computed(() => {
  const routeName = route.name?.toString() || '';
  const type = routeName.split('PackageMng')[0];
  return type;
});
const authStore = useAuthStore();
/** 插件工具管理：无 package_view 权限时展示申请占位页。
 *  列表接口对无权限用户按授权过滤后返回空数据，不能依赖列表内容判断占位；
 *  插件工具包固定只有 v2/v3 两个包名（与后端 pkg/types/constant.go 的 ReleaseNamePluginBinToolV2/V3 一致） */
const isPluginBinTool = computed(() => currentType.value === 'plugin_bintool');
const PLUGIN_BIN_TOOL_NAMES = ['plugin_bintool_v2', 'plugin_bintool_v3'];
const allPluginBinToolUnauthorized = computed(() => {
  if (!isPluginBinTool.value) return false;
  if (!authStore.authorizedMap['package_view']) return false; // 权限数据未加载完，先不判定
  return PLUGIN_BIN_TOOL_NAMES.every(name => !authStore.hasAuthorizedResource('package_view', name));
});
const pluginBinToolViewAuthItems = [{ id: 'package_view', action: 'package_view', resourceType: 'package', routes: [] }];
/** 行内操作权限：按当前行资源实例判断 package_manage 是否命中
 *  （插件工具包的 IAM package 实例是具体包名 plugin_bintool_v2/v3，其余类型实例就是固定 release_type） */
const hasRowManageAuth = (row: Release) => authStore.hasAuthorizedResource(
  'package_manage',
  isPluginBinTool.value ? row.name : row.release_type,
);
const isShow = ref(false);
const loading = ref(false);
const packageList = ref<Release[]>([]);
const originPackageList = ref<Release[]>([]);
const downloadUrl = computed(() => `${location.origin}/api/v3/package/release/${currentType.value}/download`);
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
// 搜索栏 operator 下拉（翻译后的显示名）
const translatedOperatorChildren = ref<{ id: string; name: string }[]>([]);
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
    name: t('certBintool.fileName'),
    children: getUniqueChildren('file_name'),
    multiple: true,
  },
  {
    id: 'operator',
    name: t('certBintool.uploader'),
    children: translatedOperatorChildren.value,
    multiple: true,
  },
]);

// 表格
const { isShowSetting, settings, handleSettingChange } = useTableSetting({
  checked: [
    'file_name',
    'operator',
    'updated_at',
    // 来源字段仅多租户模式存在
    ...(isMultipleTenant ? ['is_synced'] : []),
    'action',
    'download',
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
            name = item ? t('agentProxyPkg.enabled') : t('agentProxyPkg.disabled');
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

// 根据类型获取对应的列表服务方法
const getListServiceMethod = () => {
  const type = currentType.value;
  const serviceMap = {
    cert: PackageService.ListReleaseCert,
    bintool: PackageService.ListReleaseBinTool,
    plugin_bintool: PackageService.ListReleasePluginBinTool,
  };
  return serviceMap[type as keyof typeof serviceMap] || serviceMap.cert;
};

const getPackages = async () => {
  loading.value = true;

  try {
    const serviceMethod = getListServiceMethod();
    const res = await serviceMethod({ generation: PACKAGE_GENERATION });
    const list = res.items.map(item => ({
      ...item.release,
    }));
    originPackageList.value = list;
    packageList.value = list;
  } catch (error) {
    console.error('获取包列表失败:', error);
    originPackageList.value = [];
    packageList.value = [];
  } finally {
    loading.value = false;
  }
};
const getParams = (row: Release) => ({
  generation: row.generation,
  name: row.name,
});

// 根据类型获取对应的删除服务方法
const getDeleteServiceMethod = () => {
  const type = currentType.value;
  const serviceMap = {
    cert: PackageService.DeleteReleaseCert,
    bintool: PackageService.DeleteReleaseBinTool,
    plugin_bintool: PackageService.DeleteReleasePluginBinTool,
  };
  return serviceMap[type as keyof typeof serviceMap] || serviceMap.cert;
};

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

const handleDelete = async (row: Release) => {
  await handleOperation(row, getDeleteServiceMethod());
};
const handleConfirm = async () => {
  await getPackages();
};
watch(originPackageList, async () => {
  filterOptionSource.file_name.list = getUniqueChildren('file_name');
  filterOptionSource.operator.list = getUniqueChildren('operator');
  // 翻译 operator 筛选项显示名（复用 bk-user-display-name）
  await translateOperatorItems(filterOptionSource.operator.list);
  // 翻译搜索栏 operator 下拉显示名
  const operatorChildren = getUniqueChildren('operator');
  await translateOperatorItems(operatorChildren);
  translatedOperatorChildren.value = operatorChildren;
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
    await getPackages();
  },
  { immediate: true },
);
</script>
