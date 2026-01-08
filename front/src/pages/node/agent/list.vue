<template>
  <page-header :title="'Agent状态'" :back="Object.keys(route.query).length > 0"></page-header>
  <div class="p-[24px]">
    <!-- agnet操作及搜索 -->
    <section class="flex justify-between mb-[15px]">
      <div class="flex gap-[8px]">
        <!-- <Dropdown
          theme="light"
          trigger="click"
          placement="bottom-start"
          :popover-options="{
            clickContentAutoHide: true,
          }"
        >
          <Button class="w-[130px]" theme="primary" @click="handleInstall">{{
            $t("platform.nodeMan.installAgent")
          }}</Button>
          <template #content>
            <Dropdown.DropdownMenu ext-cls="dropDown-menu">
              <Dropdown.DropdownItem
                class="text-14px"
                v-for="item in agentInstallType"
                :key="item.id"
                @click="triggerHandler('setup', item.id)"
              >
                {{ item.name }}
              </Dropdown.DropdownItem>
            </Dropdown.DropdownMenu>
          </template>
        </Dropdown> -->
        <Button class="w-[130px]" theme="primary" @click="triggerHandler('setup')">{{
          $t("platform.nodeMan.installAgent")
        }}</Button>
        <Dropdown
          theme="light"
          trigger="click"
          :popover-options="{
            clickContentAutoHide: true,
          }">
          <Button :disabled="!selection.length" :loading="crossPageSelectLoading">
            <span>{{ $t("platform.nodeMan.batchOperate") }}</span>
            <i
              class="nodeman-icon nc-arrow-down ml-[5px] text-[18px] text-[#979BA5]"
            ></i>
          </Button>
          <template #content>
            <Dropdown.DropdownMenu>
              <Dropdown.DropdownItem
                v-for="item in operate"
                :key="item.id"
                @click="handleOperate(item.id, selection, true)"
              >
                {{ item.name }}
              </Dropdown.DropdownItem>
            </Dropdown.DropdownMenu>
          </template>
        </Dropdown>
        <copy-ip-dropdown
          :type="'agent'"
          :disabled="!hasSelection"
          :data="tableData"
          :list="[]"
          :is-cross-page-selection="isCrossPageSelection"
          :cross-page-query-params="crossPageQueryParams"
        ></copy-ip-dropdown>
      </div>
      <div class="flex gap-[8px]">
        <!-- <Cascader
          class="w-[250px]"
          is-remote
          clearable
          v-model="topo"
          :list="topoBizFilterList"
          id-key="bk_biz_id"
          name-key="bk_biz_name"
          :remote-method="topoRemotehandler"
          ref="topoSelect"
          :placeholder="$t('platform.nodeMan.bussinessTopology')"
        /> -->
        <SearchSelect
          class="w-[480px] z-99"
          ref="searchSelect"
          :data="searchSelectData"
          v-model.trim="searchSelectValue"
          :unique-select="true"
          :placeholder="$t('platform.nodeMan.agentSearchPlaceholder')"
          @update:model-value="handleSearchSelectChange"
        >
        </SearchSelect>
      </div>
    </section>
    <Loading
      :title="$t('table.loading')"
      :loading="loading"
      class="w-full overflow-auto"
    >
      <Table
        class="filterTable"
        :data="tableData"
        :empty-text="$t('table.empty')"
        :pagination="pagination"
        :column-config="{ resizable: true }"
        :max-height="maxHeight"
        :show-settings="isShowSetting"
        :settings="settings"
        @setting-change="handleSettingChange"
        @checkbox-change="handleSelectChange"
        @checkbox-all="handleSelectAllChange"
        @column-filter="handleFilter"
        @page-limit-change="pageLimitChange"
        @page-value-change="pageValueChange"
      >
        <template #prepend>
          <div v-if="hasSelection" class="flex items-center justify-center h-[30px] bg-[#ebecf0] text-[12px]">
            <template v-if="isCrossPageSelection">
              已跨页全选 <span class="font-bold mx-1">{{ total - excludedIds.size }}</span> 条，
              <Button text theme="primary" @click="handleClearSelection">取消选择</Button>
            </template>
            <template v-else>
              已选择 <span class="font-bold mx-1">{{ selection.length }}</span> 条，
              <Button
                v-if="total > pagination.limit"
                text theme="primary" @click="handleSelectAllCrossPage">
                选择所有页共 {{ total }} 条
              </Button>
              <Button v-else text theme="primary" @click="handleClearSelection">取消选择</Button>
            </template>
          </div>
        </template>

        <TableColumn width="80" fixed="left">
          <template #header>
            <Button text class="flex items-center justify-start">
              <Checkbox
                :model-value="isCurrentPageAllChecked"
                :indeterminate="isIndeterminate"
                @change="handleHeaderClick" />
              <Dropdown trigger="click" placement="bottom-start">
                <i class="nodeman-icon nc-arrow-down ml-1 text-[18px]"></i>
                <template #content>
                  <Dropdown.DropdownMenu>
                    <Dropdown.DropdownItem @click="handleSelectCurrentPage">本页全选</Dropdown.DropdownItem>
                    <Dropdown.DropdownItem @click="handleSelectAllCrossPage">
                      <Button text :disabled="total <= pagination.limit">跨页全选</Button>
                    </Dropdown.DropdownItem>
                  </Dropdown.DropdownMenu>
                </template>
              </Dropdown>
            </Button>
          </template>
          <template #default="{ row }">
            <Checkbox class="mt-[6px]" :model-value="row.checked" @change="(val) => handleRowCheck(val, row)" />
          </template>
        </TableColumn>
        <TableColumn
          field="bk_host_id"
          title="Host ID"
          :min-width="100"
          fixed="left"
        ></TableColumn>
        <TableColumn
          field="bk_host_innerip"
          :title="t('platform.nodeMan.inner_ip')"
          :min-width="150"
          fixed="left"
        ></TableColumn>
        <TableColumn
          field="bk_host_innerip_v6"
          :title="t('platform.nodeMan.inner_ipv6')"
          :min-width="150"
        ></TableColumn>
        <TableColumn
          field="bk_agent_id"
          :title="t('platform.nodeMan.agentId')"
          :min-width="320"
        ></TableColumn>
        <TableColumn
          field="bk_networkarea_id"
          :title="t('platform.nodeMan.bk_cloud_name')"
          :filter="filterOptionSource.bk_networkarea_id"
          :min-width="120"
          show-overflow
        >
          <template #default="{ row }">
            {{ row.bk_networkarea_name }}
          </template>
        </TableColumn>
        <TableColumn
          show-overflow
          field="bk_networkunit_id"
          :title="t('platform.nodeMan.bk_cloud_unit')"
          :filter="filterOptionSource.bk_networkunit_id"
          :min-width="120"
        >
          <template #default="{ row }">
            {{ row.bk_networkunit_name }}
          </template>
        </TableColumn>
        <TableColumn
          show-overflow
          field="os_type"
          :title="t('platform.nodeMan.os_type')"
          :filter="filterOptionSource.os_type"
          :min-width="120"
        ></TableColumn>
        <TableColumn
          show-overflow
          field="node_version"
          :title="t('platform.nodeMan.agent_version')"
          :filter="filterOptionSource.node_version"
          :min-width="120"
        ></TableColumn>
        <TableColumn
          field="node_status"
          :title="t('platform.nodeMan.status')"
          :min-width="150"
          :filter="filterOptionSource.node_status"
        >
          <template #default="{ row }">
            <div class="flex items-center" v-if="row.node_status">
              <span
                :class="`nodeman-icon nc-${row.node_status.toLowerCase()} status-icon`"
              ></span>
              <span>{{ statusMap.get(row.node_status) || row.node_status }}</span>
            </div>
            <div class="flex items-center" v-else>
              <span>--</span>
            </div>
          </template>
        </TableColumn>
        <TableColumn
          title="插件数"
          field="pluginNum"
          :min-width="122"
        >
          <template #default="{ row }">
            <Button text theme="primary" @click="openSidebar(row)">
              {{ row.pluginNum || 0 }}
              <i class="nodeman-icon nc-plug-in ml-[5px]"></i>
            </Button>
          </template>
        </TableColumn>
        <TableColumn
          field="action"
          :title="t('platform.nodeMan.operate')"
          :min-width="100"
          fixed="right"
        >
          <template #default="{ row }">
            <Button
              theme="primary"
              text
              ext-cls="reinstall"
              @click="handleOperate('reinstall', [row])"
            >
              {{ $t("platform.nodeMan.agentStatus.button.reinstall") }}
            </Button>

            <Dropdown
              theme="light"
              trigger="click"
              :popover-options="{
                clickContentAutoHide: true,
              }">
              <Button class="ml-[15px]" text>
                <span class="nodeman-icon nc-more"></span>
              </Button>
              <template #content>
                <Dropdown.DropdownMenu>
                  <Dropdown.DropdownItem
                    v-for="item in operate"
                    :key="item.id"
                    v-show="getOperateShow(row, item)"
                    @click="handleOperate(item.id, [row])"
                  >
                    {{ item.name }}
                  </Dropdown.DropdownItem>
                </Dropdown.DropdownMenu>
              </template>
            </Dropdown>
          </template>
        </TableColumn>
      </Table>
    </Loading>

    <choose-version-dialog
      v-model:is-show="chooseVersionData.isShow"
      :title="chooseVersionData.title"
      :data="chooseVersionData.data"
      :batch="chooseVersionData.batch"
      :is-cross-page-selection="isCrossPageSelection"
      type="upgrade"
      @confirm="handleUpgrade"
    >
    </choose-version-dialog>

    <operate-dialog
      v-model:is-show="operateDialogIsShow"
      :title="operateDialogData.title"
      :type="operateDialogData.type"
      :sub-title="operateDialogData.subTitle"
      @confirm="operateJob"
    ></operate-dialog>

    <!-- 侧边栏 -->
    <processSideslider
      v-model:is-show="isShowSideslider"
      type="node"
      :node="currentAgent"
    ></processSideslider>
  </div>
</template>
<script setup lang="ts">
import { Button, Checkbox, Dropdown, InfoBox, Loading, SearchSelect } from 'bkui-vue';
// 引入lodash的debounce来处理防抖，解决重复请求问题
import { debounce } from 'lodash';
import { computed, onBeforeMount, onMounted, onUnmounted, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';

import { Table, TableColumn } from '@blueking/table';

import processSideslider from '../plugin/process-sideslider.vue';

import type { TopoHostDistinctRespData } from '@/@types/topo';
import type {
  TopoHostExactConditions,
  TopoHostFuzzyConditions,
} from '@/@types/topo.d';
import { NodeAgentService } from '@/api/modules/node_agent';
import { ProcessAPIService } from '@/api/modules/process';
import { TopoService } from '@/api/modules/topo';
import useTableSetting from '@/composables/use-table-setting';
import { useMainStore } from '@/stores/main';
import { useNodeManageStore } from '@/stores/node-manage';

interface FilterOption {
  list: { text: string; value: string }[];
  checked: string[];
  filterScope: string;
}

const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const mainStore = useMainStore();
const nodeManageStore = useNodeManageStore();
// 正则表达式
const IPV4_REG = /^((25[0-5]|2[0-4]\d|[01]?\d\d?)\.){3}(25[0-5]|2[0-4]\d|[01]?\d\d?)$/;
const IPV6_REG = /^(?:[A-F0-9]{1,4}:){7}[A-F0-9]{1,4}$/i;

// ---------- 响应式数据 ----------
const tableData = ref<Host[]>([]);
const agentList = ref<Host[]>([]);
const loading = ref(false);
// 【优化点1】标记基础数据（区域、单元列表）是否加载完成
const isInitialDataLoaded = ref(false);

// 分页
const pagination = reactive({ count: 0, limit: 50, current: 1, remote: true });

// 搜索和筛选
const searchSelectValue = ref<{ id: string; name: string; values: any[] }[]>([]);
const filterOptionSource: Record<string, FilterOption> = reactive({
  bk_networkarea_id: { list: [], checked: [], filterScope: 'all' },
  bk_networkunit_id: { list: [], checked: [], filterScope: 'all' },
  os_type: { list: [], checked: [], filterScope: 'all' },
  node_version: { list: [], checked: [], filterScope: 'all' },
  node_status: { list: [], checked: [], filterScope: 'all' },
});

// 其他UI相关响应式数据
const chooseVersionData = reactive({ title: '', isShow: false, data: [], batch: false });
const operateDialogIsShow = ref(false);
const operateDialogData = { type: '', title: '', subTitle: '' };
const dropdownShow = ref(false);
const operateData = ref<Host[]>([]);
const topo = ref([]);

// ---------- 常量定义 ----------
const topoBizFilterList = computed(() => mainStore.businessList);
const operate = [
  { id: 'reinstall', name: '重装', disabled: false, show: true },
  { id: 'upgrade', name: '升级/回退', disabled: false, show: true },
  { id: 'restart', name: '重启', disabled: false, show: true },
  { id: 'uninstall', name: '卸载', disabled: false, show: true },
];
const agentInstallType = [
  { id: 'setup', name: '普通远程安装' },
  { id: 'import', name: 'Excel 导入远程安装' },
  { id: 'manual', name: '手动安装' },
];

const statusMap = ref(new Map<string, string>([
  ['init', '初始化'],
  ['running', '正常'],
  ['damaged', '异常'],
  ['unknown', '未安装'],
]));

const fuzzyKeys = new Set(['bk_host_innerip', 'bk_host_innerip_v6', 'bk_host_name', 'dept_name']);

// ---------- 计算属性 ----------
const maxHeight = computed(() => mainStore.windowInnerHeight - 214);
const selection = computed(() => tableData.value.filter((item: any) => item.checked));
const total = computed(() => pagination.count);
const networkAreaListMap = ref(new Map<number, string>([[-1, '未分配']]));
const networkUnitListMap = ref(new Map<number, string>([[-1, '未分配']]));
const hostDistinct = ref<TopoHostDistinctRespData | null>();

const searchSelectData = computed(() => [
  { id: 'bk_host_innerip', name: t('platform.nodeMan.inner_ip'), multiple: true },
  { id: 'bk_host_innerip_v6', name: t('platform.nodeMan.inner_ipv6'), multiple: true },
  { id: 'bk_agent_id', name: 'Agent ID', multiple: true },
  {
    id: 'bk_networkarea_id',
    name: '管控区域',
    children: getUniqueChildrenFrom('bk_networkarea_id', networkAreaListMap.value),
    multiple: true,
  },
  {
    id: 'bk_networkunit_id',
    name: '管控单元',
    children: getUniqueChildrenFrom('bk_networkunit_id', networkUnitListMap.value),
    multiple: true,
  },
  {
    id: 'os_type',
    name: '操作系统',
    children: getUniqueChildrenFrom('os_type'),
    multiple: true,
  },
  {
    id: 'node_version',
    name: 'Agent 版本',
    children: getUniqueChildrenFrom('node_version'),
    multiple: true,
  },
  {
    id: 'node_status',
    name: 'Agent 状态',
    children: getUniqueChildrenFrom('node_status', statusMap.value),
    multiple: true,
  },
]);

// 表格设置
const { isShowSetting, settings, handleSettingChange } = useTableSetting({
  checked: [
    'bk_host_innerip',
    'bk_host_innerip_v6',
    'bk_agent_id',
    'bk_networkarea_id',
    'bk_networkunit_id',
    'os_type',
    'node_version',
    'node_status',
    'pluginNum',
    'action',
  ],
  disabled: ['action'],
}, 'nodeMng-agent');

// --- 跨页全选核心状态 ---
const isCrossPageSelection = ref(false); // 是否开启跨页全选模式
const excludedIds = ref<Set<number>>(new Set()); // 全选模式下，用户手动“取消勾选”的 ID 集合
const crossPageQueryParams = computed(() => ({
  exact_include_conditions: getParams().exact_include_conditions,
  fuzzy_include_conditions: getParams().fuzzy_include_conditions,
  exact_exclude_conditions: {
    bk_host_id: [...excludedIds.value],
  },
}));

// 计算属性：是否有任何选中（用于禁用批量按钮）
const hasSelection = computed(() => tableData.value.some(item => item.checked) || isCrossPageSelection.value);

// 计算属性：当前页是否全选（用于表头 Checkbox 状态）
// eslint-disable-next-line max-len
const isCurrentPageAllChecked = computed(() => tableData.value.length > 0 && tableData.value.every(item => item.checked));
const isIndeterminate = computed(() => {
  const selectedCount = tableData.value.filter(item => item.checked).length;
  return selectedCount > 0 && selectedCount < tableData.value.length;
});

// 1. 处理单行勾选
const handleRowCheck = (checked: boolean, row: any) => {
  row.checked = checked;
  if (isCrossPageSelection.value) {
    if (!checked) {
      excludedIds.value.add(row.bk_host_id);
    } else {
      excludedIds.value.delete(row.bk_host_id);
    }
  }
};

// 2. 跨页全选
const handleSelectAllCrossPage = async () => {
  isCrossPageSelection.value = true;
  excludedIds.value.clear();

  // 更新当前页面的选中状态
  tableData.value.forEach(item => (item.checked = true));
};

// 3. 取消选择
const handleClearSelection = () => {
  isCrossPageSelection.value = false;
  excludedIds.value.clear();
  tableData.value.forEach(item => (item.checked = false));
};

// 4. 本页全选
const handleSelectCurrentPage = () => {
  isCrossPageSelection.value = false;
  tableData.value.forEach(item => (item.checked = true));
};

// 5. 表头 Checkbox 快速切换
const handleHeaderClick = () => {
  isCurrentPageAllChecked.value ? handleClearSelection() : handleSelectCurrentPage();
};

// 获取跨页全选的host_id数据
const crossPageSelectLoading = ref(false);
const crossPageHostIdData = ref<number[]>([]);
const getCorssPageHostIds = async () => {
  try {
    crossPageSelectLoading.value = true;
    const res = await TopoService.HostSelectHostID({
      exact_include_conditions: getParams().exact_include_conditions,
      fuzzy_include_conditions: getParams().fuzzy_include_conditions,
      exact_exclude_conditions: {
        bk_host_id: [...excludedIds.value],
      },
    });
    crossPageHostIdData.value = res.items;
  } catch (error) {
    console.error('获取跨页全选数据失败:', error);
  } finally {
    crossPageSelectLoading.value = false;
  }
};

// ---------- 辅助函数 ----------
function getUniqueChildrenFrom <K extends keyof TopoHostDistinctRespData>(
  prop: K,
  keyMap?: Map<number | string, string | number>,
) {
  const uniqueValues = hostDistinct.value?.[prop] || [];
  return uniqueValues
    .filter((item: any) => item !== '')
    .map((value: any) => ({
      id: value,
      name: keyMap?.get(value) || String(value),
    }));
}

const getParams = () => {
  const params = {
    page: { limit: pagination.limit, offset: (pagination.current - 1) * pagination.limit },
    exact_include_conditions: { bk_biz_id: mainStore.selectedBusinessId, node_role: ['agent', 'blank'] } as TopoHostExactConditions,
    fuzzy_include_conditions: {} as TopoHostFuzzyConditions,
  };
  searchSelectValue.value.forEach((item: any) => {
    const target = fuzzyKeys.has(item.id)
      ? params.fuzzy_include_conditions
      : params.exact_include_conditions;
    target[item.id] = item.values?.map((value: any) => value.id);
  });
  return params;
};

// ---------- API 请求函数 ----------

/**
 * 获取管控区域列表
 */
const getNetworkAreaList = async (data: {bk_networkarea_id: number[]} | null) => {
  const res = await TopoService.NetworkAreaList({
    page: { limit: 0 },
    exact_include_conditions: { bk_networkarea_id: data?.bk_networkarea_id || [] },
  }).catch((err: any) => {
    console.error('获取管控区域列表失败:', err);
    return { total: 0, items: [] };
  });
  res.items.forEach((item) => {
    networkAreaListMap.value.set(item.bk_networkarea_id, item.bk_networkarea_name);
  });
};

/**
 * 获取管控单元列表
 */
const getNetworkUnitList = async (data: {bk_networkunit_id: number[]} | null) => {
  const res = await TopoService.NetworkUnitList({
    exact_include_conditions: { bk_networkunit_id: data?.bk_networkunit_id || [] },
  }).catch((err: any) => {
    console.error('获取管控单元列表失败:', err);
    return { total: 0, items: [] };
  });
  res.items.forEach((item) => {
    networkUnitListMap.value.set(item.bk_networkunit_id, item.bk_networkunit_name);
  });
};

/**
 * 获取主机筛选条件的唯一值
 */
const getHostDistinct = async () => {
  const params = {
    exact_include_conditions: {
      node_role: ['agent', 'blank'],
      bk_biz_id: mainStore.selectedBusinessId,
    },
  };
  const res = await TopoService.HostDistinct(params).catch((err: any) => {
    console.error('获取主机筛选条件唯一值失败:', err);
    return null;
  });
  await Promise.all([
    getNetworkAreaList(res),
    getNetworkUnitList(res),
  ]);
  if (res) {
    hostDistinct.value = res;
    Object.keys(res).forEach((key: any) => {
      if (filterOptionSource[key]) {
        filterOptionSource[key].list = res[key]
          .filter((item: any) => item !== '')
          .map((value: string | number) => {
            let text = value;
            if (key === 'bk_networkarea_id') text = networkAreaListMap.value.get(Number(value)) || value;
            if (key === 'bk_networkunit_id') text = networkUnitListMap.value.get(Number(value)) || value;
            if (key === 'node_status') text = statusMap.value.get(value as string) || value;
            return { text, value };
          });
      }
    });
  }
};

// ---------- 侧边栏 ----------
const isShowSideslider = ref(false);
const currentAgent = ref();
// 打开侧边栏并加载进程列表
const openSidebar = async (agent) => {
  isShowSideslider.value = true;
  currentAgent.value = agent;
};


/**
 * 获取Agent列表
 */
const getAgentList = async () => {
  // 【修复点】如果基础数据未加载完成，延迟执行而不是直接返回
  if (!isInitialDataLoaded.value) {
    console.log('基础数据尚未加载完成，延迟执行getAgentList...');
    // 延迟100ms后重试，确保基础数据已加载
    setTimeout(() => {
      if (isInitialDataLoaded.value) {
        getAgentList();
      }
    }, 100);
    return;
  }

  loading.value = true;
  try {
    const res = await TopoService.HostList(getParams()).catch((err: any) => {
      console.error('获取Agent列表失败:', err);
      return { total: 0, items: [] };
    });
    pagination.count = res.total;

    const pluginNumMap = await ProcessAPIService.GetProcessDistributionByHostID({
      exact_include_conditions: {
        bk_host_id: res.items.map((item: any) => item.bk_host_id),
      },
    }).catch((err: any) => {
      console.error('获取插件数量失败:', err);
      return {} as Record<number, number>;
    });
    tableData.value = res.items.map((item: any) => {
      let isChecked = false;
      if (isCrossPageSelection.value) {
        // 如果是跨页模式，只要不在排除名单里就是选中
        isChecked = !excludedIds.value.has(item.bk_host_id);
      }
      return {
        ...item.state,
        ...item.info,
        ...item,
        bk_host_innerip: item.info.bk_host_innerip_list?.join(','),
        bk_host_innerip_v6: item.info.bk_host_innerip_v6_list?.join(','),
        pluginNum: pluginNumMap[item.bk_host_id] || 0,
        checked: isChecked,
      };
    });
    agentList.value = tableData.value;
  } catch (err) {
    console.error('获取Agent列表失败:', err);
    tableData.value = [];
    pagination.count = 0;
  } finally {
    loading.value = false;
  }
};

// 【优化点2】使用debounce包装getAgentList，延迟300ms执行，防止重复请求
const debouncedGetAgentList = debounce(getAgentList, 300);

/**
 * 加载所有初始化数据（区域、单元、筛选条件）
 */
const loadInitialData = async () => {
  if (!mainStore.selectedBusinessId || isInitialDataLoaded.value) {
    return;
  }

  try {
    // 【修复点】并行执行三个基础请求，但只执行一次
    await getHostDistinct();
    isInitialDataLoaded.value = true; // 标记基础数据已加载完成
  } catch (error) {
    console.error('加载初始化数据失败:', error);
    // 即使API失败，也标记为已加载完成，避免无限重试
    isInitialDataLoaded.value = true;
  }
};

// ---------- 事件处理函数 ----------
/**
 * 处理粘贴/快速输入的逻辑
 */
const handleInputPaste = (data: { id: string; name: string; values: { id: string; name: string }[] }[]) => {
  const text = data.length > 0 ? data[data.length - 1].id : '';
  if (text) {
    let targetId = '';
    let targetName = '';

    // 1. 自动识别 IP 类型
    if (IPV4_REG.test(text)) {
      targetId = 'bk_host_innerip';
      targetName = t('platform.nodeMan.inner_ip');
    } else if (IPV6_REG.test(text)) {
      targetId = 'bk_host_innerip_v6';
      targetName = t('platform.nodeMan.inner_ipv6');
    }

    // 2. 如果匹配成功，直接构造并推入 searchSelectValue
    if (targetId) {
      const index = searchSelectValue.value.findIndex((item: any) => item.id === targetId);
      if (index > -1) searchSelectValue.value.splice(index, 1);

      searchSelectValue.value.push({
        id: targetId,
        name: targetName,
        values: [{ id: text, name: text }],
      });
      // 3. 移除粘贴的文本
      searchSelectValue.value.splice(data.length - 2, 1);
      return;
    }
  }
};
const handleSearchSelectChange = (data: { id: string; name: string; values: { id: string; name: string }[] }[]) => {
  handleInputPaste(data);
  Object.keys(filterOptionSource).forEach((key) => {
    filterOptionSource[key].checked = [];
  });
  data.forEach((item) => {
    if (filterOptionSource[item.id]) {
      filterOptionSource[item.id].checked = item.values.map((value: any) => value.id);
    }
  });
  // 【优化点2】使用防抖后的函数
  debouncedGetAgentList();
};

const handleFilter = ({ checked, field }: { checked: string[]; field: string }) => {
  const index = searchSelectValue.value.findIndex((item: any) => item.id === field);
  if (index > -1) searchSelectValue.value.splice(index, 1);
  if (checked.length) {
    searchSelectValue.value.push({
      id: field,
      name: field,
      values: checked.map((item: any) => {
        let name = item;
        if (field === 'bk_networkarea_id') name = networkAreaListMap.value.get(Number(item)) || item;
        if (field === 'bk_networkunit_id') name = networkUnitListMap.value.get(Number(item)) || item;
        if (field === 'node_status') name = statusMap.value.get(item) || item;
        return { id: item, name };
      }),
    });
  }
};

const pageLimitChange = async (limit: number) => {
  pagination.limit = limit;
  pagination.current = 1; // 页码重置为1
  await getAgentList(); // 分页变化不防抖，立即执行
};

const pageValueChange = async (current: number) => {
  pagination.current = current;
  await getAgentList(); // 分页变化不防抖，立即执行
};

// const handleInstall = () => {
//   if (selection.value.length) {
//     dropdownShow.value = false;
//     triggerHandler('reinstall');
//   } else {
//     dropdownShow.value = !dropdownShow.value;
//   }
// };

const triggerHandler = (type: string, setupType = 'setup') => {
  switch (type) {
    case 'restart':
    case 'reinstall':
    case 'uninstall':
    case 'upgrade':
      handleOperate(type, selection.value, true);
      break;
    case 'setup':
      router.push({ name: 'agentSetup' });
      mainStore.updateAgentSetupType(setupType);
      break;
  }
};

const getOperateShow = (row: Host, config: any) => {
  if (config.id === 'reinstall') {
    return false;
  }
  return config.show;
};

const handleOperate = async (type: string, data: Host[], batch = false) => {
  // 如果是跨页全选模式，获取所有数据
  let operateData = data;
  if (isCrossPageSelection.value && type !== 'reinstall') {
    await getCorssPageHostIds();
    operateData = crossPageHostIdData.value.map((item: any) => ({ bk_host_id: item }));
    batch = true; // 强制设置为批量模式
  }

  switch (type) {
    case 'restart':
      handleOperatetHost(operateData, batch, 'restart');
      break;
    case 'uninstall':
      handleOperatetHost(operateData, batch, 'uninstall');
      break;
    case 'upgrade':
      handleOperatetHost(operateData, batch, 'upgrade');
      break;
  }
  if (type !== 'reinstall') return;
  const params = {
    tableData: operateData.map((item: any) => ({ ...item })),
    type: 'reinstall',
    isCrossPageSelection: isCrossPageSelection.value,
    queryParams: crossPageQueryParams.value,
  };
  nodeManageStore.updateAgentEditRowData(params);
  router.push({ name: 'agentEdit' });
};
const handleSelectChange = ({ checked, row }: { checked: boolean; row: any }) => {
  row.checked = checked;
};

const handleSelectAllChange = ({ checked }: { checked: boolean }) => {
  tableData.value.forEach((item: any) => (item.checked = checked));
};

const operateJob = async (extraData: any = {}) => {
  loading.value = true;
  const params = {
    host: operateData.value?.map((item: any) => ({
      bk_host_id: item.bk_host_id,
      force: extraData.isForce,
      graceful_restart_timeout_sec: extraData.time,
    })),
  };
  let result;
  if (extraData.isReconfig) {
    result = await NodeAgentService.NodeAgentReconfig(params).catch(() => ({ workflow_id: '' }));
  } else {
    result = await NodeAgentService.NodeAgentRestart(params).catch(() => ({ workflow_id: '' }));
  }
  loading.value = false;
  if (result.workflow_id) {
    router.push({
      name: 'taskDetail',
      params: { taskId: result.workflow_id, routerBackName: 'taskList' },
      query: {
        active: 'node',
      },
    });
  }
};

const handleUpgrade = async (osVersion: any[], upgradeData: {force: boolean, graceful_restart_timeout_sec: number}) => {
  loading.value = true;
  const params = {
    host: operateData.value?.map((item: any) => ({
      bk_host_id: item.bk_host_id,
      target_version: osVersion[0].version,
      force: upgradeData.force,
      graceful_restart_timeout_sec: upgradeData.graceful_restart_timeout_sec,
    })),
  };
  const result = await NodeAgentService.NodeAgentUpgrade(params).catch(() => ({ workflow_id: '' }));
  loading.value = false;
  if (result.workflow_id) {
    router.push({
      name: 'taskDetail',
      params: { taskId: result.workflow_id, routerBackName: 'taskList' },
      query: {
        active: 'node',
      },
    });
  }
};

const handleUninstall = async () => {
  loading.value = true;
  const result = await NodeAgentService.NodeAgentUninstall({
    host: operateData.value?.map((item: any) => ({ bk_host_id: item.bk_host_id })),
  }).catch(() => ({ workflow_id: '' }));
  loading.value = false;
  if (result.workflow_id) {
    router.push({
      name: 'taskDetail',
      params: { taskId: result.workflow_id, routerBackName: 'taskList' },
      query: {
        active: 'node',
      },
    });
  }
};

const handleOperatetHost = async (data: Host[], batch: boolean, operateType: string) => {
  const titleObj = {
    firstIp: isCrossPageSelection.value ? selection.value[0].bk_host_innerip : data[0].bk_host_innerip,
    num: data.length,
  };
  let type = '';
  switch (operateType) {
    case 'restart': type = '重启'; break;
    case 'upgrade': type = '升级/回退'; break;
    case 'uninstall': type = '卸载'; break;
  }
  operateData.value = data;
  if (operateType === 'upgrade') {
    chooseVersionData.title = 'Agent 升级/回退';
    chooseVersionData.isShow = true;
    chooseVersionData.data = data;
    chooseVersionData.batch = batch;
  } else if (operateType === 'restart') {
    operateDialogIsShow.value = true;
    operateDialogData.type = operateType;
    operateDialogData.title = batch ? `请确认是否批量${type}` : `请确认是否${type}`;
    operateDialogData.subTitle = batch
      ? `${type} ${titleObj.firstIp} 等${titleObj.num}个IP的Agent`
      : `${type} ${titleObj.firstIp} 的Agent`;
  } else if (operateType === 'uninstall') {
    InfoBox({
      title: batch ? `请确认是否批量${type}` : `请确认是否${type}`,
      subTitle: batch
        ? `${type} ${titleObj.firstIp} 等${titleObj.num}个IP的Agent`
        : `${type} ${titleObj.firstIp} 的Agent`,
      onConfirm: () => {
        handleUninstall();
      },
    });
  }
};


// ---------- 监听与生命周期 ----------

// 【修复点1】添加统一的加载状态控制
const isLoading = ref(false);

// 【修复点2】优化路由参数处理，避免重复触发
watch(() => route.query, (newQuery, oldQuery) => {
  // 只有当路由参数实际发生变化时才处理
  if (JSON.stringify(newQuery) === JSON.stringify(oldQuery)) return;

  const { os_type, cpu_arch, node_version, bk_networkarea_id, bk_networkunit_id } = newQuery;

  // 处理操作系统、架构、版本的筛选
  if (os_type && cpu_arch && node_version) {
    searchSelectValue.value = [
      ...searchSelectValue.value.filter(item => !['os_type', 'cpu_arch', 'node_version'].includes(item.id)),
      { id: 'os_type', name: '操作系统', values: [{ id: os_type, name: os_type }] },
      { id: 'cpu_arch', name: '架构', values: [{ id: cpu_arch, name: cpu_arch }] },
      { id: 'node_version', name: 'Agent版本', values: [{ id: node_version, name: node_version }] },
    ];
  } else if (bk_networkarea_id !== undefined && bk_networkunit_id !== undefined) {
    const areaId = Number(bk_networkarea_id);
    const unitId = Number(bk_networkunit_id);
    searchSelectValue.value = [
      ...searchSelectValue.value.filter(item => !['bk_networkarea_id', 'bk_networkunit_id'].includes(item.id)),
      {
        id: 'bk_networkarea_id',
        name: t('platform.nodeMan.bk_cloud_name'),
        values: [{ id: areaId, name: networkAreaListMap.value.get(areaId) || areaId }],
      },
      {
        id: 'bk_networkunit_id',
        name: t('platform.nodeMan.bk_cloud_unit'),
        values: [{ id: unitId, name: networkUnitListMap.value.get(unitId) || unitId }],
      },
    ];
  } else {
    searchSelectValue.value = [];
  }
}, { immediate: true });

// 【修复点3】统一初始化逻辑，避免重复调用
const initializePage = async () => {
  if (isLoading.value) return;

  isLoading.value = true;
  try {
    // 重置状态
    isInitialDataLoaded.value = false;
    tableData.value = [];
    pagination.count = 0;
    pagination.current = 1;

    // 重新加载基础数据
    await loadInitialData();

    // 基础数据加载完成后，加载Agent列表
    if (isInitialDataLoaded.value) {
      await getAgentList();
    }
  } finally {
    isLoading.value = false;
  }
};

// 【修复点4】优化业务ID监听，使用防抖避免重复初始化
const debouncedInitialize = debounce(initializePage, 100);

watch(() => mainStore.selectedBusinessId, (newId, oldId) => {
  if (newId && newId !== oldId) {
    debouncedInitialize();
  }
}, { immediate: true });

// 【修复点5】优化搜索条件监听，避免与初始化冲突
watch(
  searchSelectValue,
  () => {
    if (isInitialDataLoaded.value) {
      pagination.current = 1;
      debouncedGetAgentList();
    }
  },
  { deep: true },
);

// 清理防抖函数
onUnmounted(() => {
  debouncedGetAgentList.cancel();
  debouncedInitialize.cancel();
});

</script>
<style lang="postcss" scoped>
.dropDown-menu {
  .bk-dropdown-item {
    font-size: 14px;
  }
}
.bk-cascader-wrapper {
  width: 250px;

  .bk-cascader-wrapper:not(:last-of-type) {
    margin-bottom: 20px;
  }
}

.status-icon::before {
  content: "";
  display: inline-block;
  margin-right: 8px;
  width: 14px;
  height: 14px;
  border: 3px solid #f0f1f5;
  border-radius: 6.5px;
  background: #b2b5bd;
  flex-shrink: 0;
  vertical-align: middle;
}
.nc-running {
  &::before {
    background: #3fc06d;
    border-color: #e5f6ea;
  }
}
.nc-terminated {
  &::before {
    border-color: #ffe6e6;
    background: #ea3636;
  }
}
.nc-unknown {
  &::before {
    border-color: #f0f1f5;
    background: #b2b5bd;
  }
}
</style>
