<template>
  <PageHeader class="w-full sticky top-0 z-1" :title="title" :back="false">
    <!-- <Tag radius="14px" class="ml-[20px]">当前版本：{{ curAgentVersion }}</Tag> -->
  </PageHeader>
  <div class="p-[24px] h-[calc(100%_-_52px)] flex flex-col">
    <!-- 搜索栏 -->
    <div class="flex items-center w-full h-[32px] mb-[16px]">
      <Button theme="primary" @click="handleUpload">包上传</Button>
      <SearchSelect
        class="ml-[16px] flex-1"
        ref="searchSelect"
        :data="searchSelectData"
        v-model="searchSelectValue"
        :uniqueSelect="true"
        :placeholder="t('版本号、操作系统/架构、标签、上传用户、状态')"
        @update:modelValue="handleSearchSelectChange"
      >
      </SearchSelect>
    </div>
    <div class="flex flex-1 w-full">
      <div
        class="w-[240px] flex-shrink-0 bg-[#fff] rounded-[2px] shadow-[0_2px_4px_#1919290d] h-full mr-[17px]"
      >
        <div class="px-[16px] py-[10px] text-[12px]">{{ t("快捷筛选") }}</div>
        <div class="flex h-[32px] px-[16px]">
          <div
            class="bg-[#fafbfd] w-[40px] flex items-center justify-center border border-r-none border-[#C4C6CC] text-[12px]"
          >
            {{ $t("维度") }}
          </div>
          <Select
            class="flex-1"
            v-model="state.dimension"
            :clearable="false"
            @change="val => updateOptionalList(val, true)"
          >
            <Select.Option
              v-for="opt in dimensionList"
              :key="opt.id"
              :id="opt.id"
              :name="$t(opt.name)"
            ></Select.Option>
          </Select>
        </div>
        <div
          :class="[
            'border-b flex justify-between items-center h-[36px] mt-[8px] px-[16px] cursor-pointer',
            {
              'bg-[#E1ECFF] text-[#3A84FF]': state.dimensionOptional === 'all',
            },
          ]"
          @click="selectDimensionOptional('all', 'click')"
        >
          <div class="text-[13px]">
            <span class="mr-[5px]">All</span>
            <span>全部</span>
          </div>
          <Tag
            size="large"
            checkable
            :checked="state.dimensionOptional === 'all'"
            >{{ originPackageList.length }}</Tag
          >
        </div>
        <div
          v-for="item in dimensionOptionalList"
          :key="item.id"
          :class="[
            'flex justify-between items-center h-[36px] cursor-pointer px-[16px]',
            {
              'bg-[#E1ECFF] text-[#3A84FF]':
                state.dimensionOptional === item.id,
            },
          ]"
          @click="selectDimensionOptional(item.id, 'click')"
        >
          <div>
            <i :class="`nodeman-icon nc-${item.id.split('_')[0]} mr-[5px]`"></i>
            <span class="text-[12px]">{{ `${item.name}` }}</span>
          </div>
          <Tag
            v-if="item.count"
            size="large"
            checkable
            :checked="state.dimensionOptional === item.id"
          >
            {{ item.count }}
          </Tag>
        </div>
      </div>
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
            :title="t('包名称')"
            :min-width="320"
            fixed="left"
            show-overflow="tooltip"
          ></TableColumn>
          <TableColumn
            field="version"
            :title="t('版本号')"
            :min-width="130"
            sortable
            show-overflow="tooltip"
          ></TableColumn>
          <TableColumn
            field="os_cpu_arch"
            :title="t('操作系统/架构')"
            :min-width="150"
          ></TableColumn>
          <TableColumn
            field="labels"
            :title="t('标签信息')"
            :min-width="180"
            :filter="filterOptionSource.labels"
          >
            <template #default="{ row }">
              <div v-show="!row.isShowTagInput" class="flex items-center group gap-[5px]">
                <Tag
                  v-for="tag in row.labels?.slice(0, 2)"
                  :key="tag"
                  >{{ tag }}</Tag
                >
                <Tag
                  v-if="row.labels?.length > 2"
                  v-bk-tooltips="row.labels?.join(', ')"
                  >+{{ row.labels?.length - 2 }}</Tag
                >
                <!-- <edit-line class="!hidden !group-hover:block" @click="row.isShowTagInput = true" /> -->
              </div>
              <TagInput
                v-show="row.isShowTagInput"
                :model-value="row.labels"
                :list="getUniqueChildren('labels')"
                trigger="focus"
                collapse-tags
                @change="handleChangeTag"
              />
            </template>
          </TableColumn>
          <TableColumn
            field="operator"
            :title="t('上传用户')"
            :min-width="120"
            :filter="filterOptionSource.operator"
          ></TableColumn>
          <TableColumn
            field="updated_at"
            :title="t('上传时间')"
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
            :filter="filterOptionSource.enabled"
          >
            <template #default="{ row }">
              <Tag v-if="row.enabled" theme="success">{{ t("启用") }}</Tag>
              <Tag v-else>{{ t("禁用") }}</Tag>
            </template>
          </TableColumn>
          <TableColumn
            field="as_default"
            :title="t('默认版本')"
            :min-width="120"
          >
            <template #default="{ row }">
              <Tag v-if="row.enabled" theme="success">{{ t("是") }}</Tag>
              <Tag v-else>{{ t("否") }}</Tag>
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
</template>
<script lang="ts" setup>
type PkgQuickType = "os_cpu_arch" | "version";
type PkgType = "gse_agent" | "gse_proxy";
type filterProp = 'labels' | 'operator' | 'enabled';
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
import { EditLine } from "bkui-vue/lib/icon";

const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const title = computed(() =>
  route.name === "agentPackageMng" ? t("Agent 包管理") : t("Proxy 包管理")
);
const curAgentVersion = ref("v2.2.6-beta.30");
const loading = ref(false);
const packageList = ref<Release[]>([]);
const originPackageList = ref<Release[]>([]);
const state = reactive<{
  isLoading: boolean;
  panels: { name: PkgType; label: string }[];
  active: PkgType;
  dimension: PkgQuickType;
  dimensionOptional: string;
  uploadShow: boolean;
  ordering: PkgOrderType | "";
}>({
  isLoading: true,
  panels: [
    { name: "gse_agent", label: "Agent" },
    { name: "gse_proxy", label: "Proxy" },
  ],
  active: "gse_agent",
  // 维度
  dimension: "os_cpu_arch",
  dimensionOptional: "all",
  uploadShow: false,
  ordering: "",
});
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
const dimensionList = computed(() =>
  searchSelectData.value.filter((item: { id: string }) =>
    ["version", "os_cpu_arch"].includes(item.id)
  )
);
const dimensionOptionalList = computed(
  () =>
    searchSelectData.value.find(
      (item: { id: string }) => item.id === state.dimension
    )?.children
);
const filterOptionSource = reactive<Record<string, IFilterOption>>({
  labels: {
    list: getUniqueChildren("labels"),
    checked: [],
    filterScope: "all",
  },
  operator: {
    list: getUniqueChildren("operator"),
    checked: [],
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
    "version",
    "os_cpu_arch",
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
  updateOptionalList('os_cpu_arch');
  selectDimensionOptional('all');
  data.forEach(item => {
    if (filterOptionSource[item.id as filterProp]) {
      filterOptionSource[item.id as filterProp].checked = item.values.map((item: any) => item.id) as string[];
    }
    if (["version", "os_cpu_arch"].includes(item.id)) {
      updateOptionalList(item.id as PkgQuickType);
      selectDimensionOptional(item.values[0].id);
    }
  });
}
// 更新快捷筛选 维度
const updateOptionalList = (
  id: PkgQuickType = "os_cpu_arch",
  isTabChange: boolean = false
) => {
  state.dimension = id;
  if (!isTabChange) return;
  // 如果搜索栏没这个维度搜索值，切换的时候默认选择全部, 如果有值则切换后勾选搜索值
  const findItem = searchSelectValue.value.find((item: {id: string}) => item.id === id);
  if(!findItem) {
    selectDimensionOptional("all");
  } else {
    selectDimensionOptional(findItem.values[0].id);
  }
};
// 维度nav click
const selectDimensionOptional = (id: string, type?: string) => {
  if (id === state.dimensionOptional) return;
  state.dimensionOptional = id;
  if (type === 'click') {
    updateQuickOptToSearch(id);
  }
};
const updateQuickOptToSearch = (id: string) => {
  const index = searchSelectValue.value.findIndex(
    (item: any) => item.id === state.dimension
  );
  index > -1 && searchSelectValue.value.splice(index, 1);
  if (id === 'all') return;
  const findData = searchSelectData.value.find((item: {id: string}) => item.id === state.dimension);
  searchSelectValue.value.push({
    id: state.dimension,
    name: findData.name,
    values: [findData.children.find((item: {id: string, name: string}) => item.id === id)]
  });
};
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
  
}
const getPackages = async () => {
  loading.value = true;
  const currentType = route.name === "agentPackageMng" ? "agent" : "proxy";
  const res = await PackageService.ListRelease({
    exact_include_conditions: {
      release_type: [currentType],
    },
  });
  const hostList = await PackageService.DeployedHostCount({
    request_items: res.items
  });
  const items = res.items.map((item, index) => ({
    ...item,
    labels: item.labels || [],
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
