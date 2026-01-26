<template>
  <div class="p-[24px] h-[calc(100%_-_52px)] flex flex-col">
    <!-- 搜索栏 -->
    <div class="flex items-center w-full h-[32px] mb-[16px]">
      <Button theme="primary" @click="handleUpload" class="mr-[16px]">{{ t('certBintool.upload') }}</Button>
      <SearchSelect
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
        ></TableColumn>
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
          field="action"
          :title="t('certBintool.action')"
          fixed="right"
          :min-width="120"
        >
          <template #default="{ row }">
            <div class="flex items-center">
              <PopConfirm
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
            </div>
          </template>
        </TableColumn>
        <TableColumn
          field="download"
          :title="t('certBintool.download')"
          fixed="right"
          :min-width="60"
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
const currentType = computed(() => {
  const routeName = route.name?.toString() || '';
  const type = routeName.split('PackageMng')[0];
  return type;
});
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
    const res = await serviceMethod({ generation: 2 });
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
