<template>
  <div class="p-[24px] h-[calc(100%_-_52px)] flex flex-col">
    <!-- 搜索栏 -->
    <div class="flex items-center w-full h-[32px] mb-[16px]">
      <!-- <Button theme="primary" @click="handleUpload">包上传</Button> -->
      <SearchSelect
        class="flex-1"
        ref="searchSelect"
        :data="searchSelectData"
        v-model="searchSelectValue"
        :uniqueSelect="true"
        :placeholder="t('版本号、操作系统/架构、标签、上传用户、状态')"
        @update:modelValue="handleSearchSelectChange"
      >
      </SearchSelect>
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
          :title="'包名称'"
          :min-width="320"
          fixed="left"
          show-overflow="tooltip"
        ></TableColumn>
        <TableColumn
          field="labels"
          :title="'标签信息'"
          :min-width="180"
          :filter="filterOptionSource.labels"
        >
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
                    <div class="text-[16px] text-[#313238] mb-[6px]">确认删除该{{ route.name === "certMng" ? "证书" : "工具" }}？</div>
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
type PkgType = "gse_agent" | "gse_proxy";
type filterProp = 'labels' | 'operator' | 'enabled';
interface IFilterOption {
  list: { value: string | boolean, text: string;  }[];
  checked: string[];
  filterScope: string;
  match?: string,
}
// 排序类型
import { ref, reactive, computed, onMounted, watch } from "vue";
import { useI18n } from "vue-i18n";
import { Button, SearchSelect, Tag, Select, Loading, Popover, TagInput } from "bkui-vue";
import useTableSetting from "@/composables/use-table-setting";
import { PackageService } from "@/api/modules/pkg";
import { Table, TableColumn } from "@blueking/table";
import { formatTimestamp, capitalizeFirstLetter, compareVersions } from "@/common/util";
import type { Release } from "@/@types/common.d";
import { isArray } from "lodash";
import { useRoute, useRouter } from "vue-router";
import type { VxeTablePropTypes } from 'vxe-table';
import usePage from '@/composables/use-page';
import { EditLine, AngleDown, AngleRight } from "bkui-vue/lib/icon";
import PkgUploadSideslider from "../agent-proxy-pkg/pkg-upload-sideslider.vue";
import { useMainStore } from "@/stores/main";

const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const mainStore = useMainStore();
const maxHeight = computed(() => mainStore.windowInnerHeight - 214);
const title = computed(() =>
  route.name === "agentPackageMng" ? t("Agent 包管理") : t("Proxy 包管理")
);
const isShow = ref(false);
const loading = ref(false);
const packageList = ref<Release[]>([]);
const originPackageList = ref<Release[]>([]);
// 分页
const {
  pagination
} = usePage(packageList);
const sortConfig = ref<VxeTablePropTypes.SortConfig>({
  sortMethod ({ data, sortList }) {
    const sortItem = sortList[0]
    // 取出第一个排序的列
    const { field, order } = sortItem
    // 通用排序函数
    function sortData(a: Release, b: Release, field: string) {
      return a[field] - b[field];
    }

    const sortedList = data.sort((a: Release, b: Release) => {
      const comparison = sortData(a, b, field);
      return order === 'desc' ? -comparison : comparison;
    });

    return sortedList;
  }
});

const filterOptionSource = reactive<Record<string, IFilterOption>>({
  labels: {
    list: [],
    checked: [],
    match: 'fuzzy',
    filterScope: "all",
  },
  operator: {
    list: [],
    checked: [],
    match: 'fuzzy',
    filterScope: "all",
  },
  enabled: {
    list: [
      { value: true, text: t("启用") },
      { value: false, text: t("禁用") },
    ],
    checked: [],
    filterScope: "all",
  },
});
// 搜索
const searchSelectValue = ref<{ id: string; name: string; values: any[] }[]>([]);
function getUniqueChildren(prop: string) {
  const res = Array.from(
    new Set(
      originPackageList.value
        .map((item: any) => item[prop])
        .filter((item: any) => String(item))
    )
  );
  const uniqueValues = isArray(res[0]) ? [...res[0]] : res;
  return uniqueValues.map((value: any) => {
    let name = String(value);
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
    id: "labels",
    name: t("标签"),
    children: getUniqueChildren("labels"),
    multiple: true,
  },
  {
    id: "operator",
    name: t("上传用户"),
    children: getUniqueChildren("operator"),
    multiple: true,
  },
  {
    id: "enabled",
    name: t("状态"),
    children: [
      { id: true, name: t("启用") },
      { id: false, name: t("禁用") },
    ],
  },
]);

// 表格
const { isShowSetting, settings, handleSettingChange } = useTableSetting({
  checked: [
    "file_name",
    "labels",
    "operator",
    "updated_at",
    "host",
    "enabled",
    "as_default",
    "action",
  ],
  disabled: ["action"],
});
// 筛选
const handleFilter = ({
  checked,
  field,
}: {
  checked: string[];
  field: string;
}) => {
  const index = searchSelectValue.value.findIndex(
    (item: any) => item.id === field
  );
  index > -1 && searchSelectValue.value.splice(index, 1);
  if (checked.length) {
    searchSelectValue.value.push({
      id: field,
      name: t(field),
      values: checked.map((item: any) => {
        let name;
        switch (field) {
          case 'enabled':
            name = item ? t("启用") : t("禁用");
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
const handleSearchSelectChange = async (data: {id: string, name: string, values: {id: string,name: string}[]}[]) => {
  Object.keys(filterOptionSource).forEach(key => {
    filterOptionSource[key].checked = [];
  });
  data.forEach(item => {
    if (filterOptionSource[item.id as filterProp]) {
      filterOptionSource[item.id as filterProp].checked = item.values.map((item: any) => item.id) as string[];
    }
  });
}

const handleBlur = async () => {
  await getPackages();
}
const handleUpload = () => {
  isShow.value = true;
}
const getPackages = async () => {
  loading.value = true;
  const currentType = route.name === "certMng" ? "cert" : "bintool";
  const res = await PackageService.ListRelease({
    release_type: currentType,
    generation: [2]
  });
  const items = res.items.map((item, index) => ({
    ...item,
    labels: item.labels || [],
    isDisabledPopShow: false,
    isDeletePopShow: false,
    isShowTagInput: false,
    createPopShow: false
  }))
  originPackageList.value = items;
  packageList.value = items;
  loading.value = false;
};
const getParams = (row: Release) => {
  return {
    generation: row.generation,
    release_type: row.release_type,
    platform: {
      os_type: row.os_type,
      cpu_arch: row.cpu_arch,
    },
    version: row.version,
  };
};
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
}
watch(originPackageList, () => {
  filterOptionSource.labels.list = getUniqueChildren("labels");
  filterOptionSource.operator.list = getUniqueChildren("operator");
},{ immediate: true, deep: true })
// 前端过滤数据
watch(
  [
    searchSelectValue,
    originPackageList
  ],
  () => {
    packageList.value = originPackageList.value.filter((row: Release) => {
      return searchSelectValue.value.every((searchItem: any) => {
        const { id: searchField, values } = searchItem;
        const searchIds = values?.map((value: {id: string}) => value.id);
        if (isArray(row[searchField])) {
          return !!row[searchField].find((el: string) => searchIds.includes(el));
        }
        return searchIds.includes(row[searchField]);
      });
    });
  },
  { immediate: true, deep: true }
);
watch(
  () => route.name,
  async () => {
    searchSelectValue.value = [];
    await getPackages();
  },
  { immediate: true }
);
onMounted(async () => {
});
</script>
