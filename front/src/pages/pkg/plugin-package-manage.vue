<template>
  <div class="p-[24px] h-[calc(100%_-_52px)] flex flex-col">
    <!-- 搜索栏 -->
    <div class="flex items-center w-full h-[32px] mb-[16px]">
      <Button theme="primary" @click="handleUpload" :disabled="true">插件上传</Button>
      <SearchSelect
        class="ml-[16px] flex-1"
        ref="searchSelect"
        :data="searchSelectData"
        v-model="searchSelectValue"
        :uniqueSelect="true"
        :placeholder="t('请输入 插件名称、插件别名、状态 搜索')"
        @update:modelValue="handleSearchSelectChange"
      >
      </SearchSelect>
    </div>
    <div class="flex flex-1 w-full">
      <Loading
        title="数据加载中"
        :loading="loading"
        class="flex-1 overflow-auto"
      >
        <Table
          class="w-full"
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
            :title="t('插件名称')"
            :min-width="320"
            fixed="left"
            show-overflow="tooltip"
          ></TableColumn>
          <TableColumn
            field="version"
            :title="t('插件别名')"
            :min-width="130"
            show-overflow="tooltip"
          ></TableColumn>
          <TableColumn
            field="os_cpu_arch"
            :title="t('插件介绍')"
            :min-width="150"
          ></TableColumn>
          <TableColumn
            field="host"
            :title="t('已部署主机')"
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
            :title="t('状态')"
            :min-width="120"
          >
            <template #default="{ row }">
              <Tag v-if="row.enabled" theme="success">{{ t("启用") }}</Tag>
              <Tag v-else>{{ t("禁用") }}</Tag>
            </template>
          </TableColumn>
          <TableColumn
            field="action"
            :title="t('操作')"
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
                      <div class="text-[16px] text-[#313238] mb-[6px]">确认停用该 Agent 包？</div>
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
                      <div class="text-[16px] text-[#313238] mb-[6px]">确认删除该 Agent 包？</div>
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
type PkgQuickType = "os_cpu_arch" | "version";
type PkgType = "gse_agent" | "gse_proxy";
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
}
// 排序类型
type PkgOrderType = "version" | "-version";
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
import PkgUploadSideslider from "./agent-proxy-pkg/pkg-upload-sideslider.vue";
import { EditLine } from "bkui-vue/lib/icon";

const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const loading = ref(false);
const packageList = ref<Release[]>([]);
const originPackageList = ref<Release[]>([]);
const isShow = ref(false);
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
      if (field === 'version') {
        return compareVersions(a[field], b[field]);
      } else {
        return a[field] - b[field];
      }
    }

    const sortedList = data.sort((a: Release, b: Release) => {
      const comparison = sortData(a, b, field);
      return order === 'desc' ? -comparison : comparison;
    });

    return sortedList;
  }
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
    if (["os_cpu_arch", "version"].includes(prop)) {
      count = countByProp(originPackageList.value, prop)[value as string];
      if (prop === "os_cpu_arch") {
        name = capitalizeFirstLetter(value);
      }
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
    id: "version",
    name: t("版本号"),
    children: getUniqueChildren("version"),
  },
  {
    id: "os_cpu_arch",
    name: t("操作系统/架构"),
    children: getUniqueChildren("os_cpu_arch"),
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
    "version",
    "os_cpu_arch",
    "host",
    "enabled",
    "action",
  ],
  disabled: ["action"],
});

// 搜索
const handleSearchSelectChange = async (data: {id: string, name: string, values: {id: string,name: string}[]}[]) => {
  
}
const handleClickHost = (row: Release) => {
  if(!row.host || row.release_type !== 'agent') return;
  router.push({
    name: 'agent',
    query: {
      os_type: row.os_type,
      node_version: row.version
    }
  });
}
const handleChangeTag = () => {

}
const handleUpload = () => {
  isShow.value = true;
}
const handleConfirm = async () => {
  await getPackages();
}
const getPackages = async () => {
  loading.value = true;
  const res = await PackageService.ListRelease({
    release_type: "agent",
    generation: 2
  });
  const hostList = await PackageService.DeployedHostCount({
    request_items: res.items.map(item => ({
      generation: item.generation,
      release_type: item.release_type,
      version: item.version,
      platform: {
        os_type: item.os_type,
        cpu_arch: item.cpu_arch
      }
    }))
  });
  const items = res.items.map((item, index) => ({
    ...item,
    os_cpu_arch: item.os_type + "_" + item.cpu_arch,
    host: hostList.items[index],
    isDisabledPopShow: false,
    isDeletePopShow: false,
    isShowTagInput: false,
    createPopShow: false
  }))
  .sort((a, b) => compareVersions(a.version,b.version));
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
// 前端过滤数据
watch(
  searchSelectValue,
  () => {
    packageList.value = originPackageList.value.filter((row: Release) => {
      return searchSelectValue.value.every((searchItem: any) => {
        const { id: searchField, values } = searchItem;
        const searchIds = values.map((value: {id: string}) => value.id);
        if (searchField === 'enabled') {
          return searchIds.includes(row[searchField]);
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
    await getPackages();
  },
  { immediate: true }
);
onMounted(async () => {
});
</script>
