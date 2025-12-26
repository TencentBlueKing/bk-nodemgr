<template>
  <PageHeader
    class="w-full absolute top-0 z-100"
    :title="$t('platform.taskHistory.taskDetail.title')"
    :back="true"
    :on-back="handleBackToHistory"
  >
    <span class="mx-[6px] text-[#979BA5] text-[14px]">-</span>
    <span class="text-[#979BA5] text-[14px] mr-[14px]" v-if="currentData">{{
      currentData?.workflow_id
    }}</span>
    <Tag
      :theme="statusMap[currentTaskStatus]?.tagTheme"
      type="filled"
      v-if="currentTaskStatus"
    >
      {{ statusMap[currentTaskStatus]?.text || "" }}
    </Tag>
  </PageHeader>
  <div class="p-[24px] mt-[52px]">
    <div class="flex">
      <div
        v-for="item in taskInfoList"
        :key="item.name"
        class="leading-[30px] mr-[52px] text-[12px]"
      >
        <div class="w-[50px]">{{ item.name }}</div>
        <div>{{ item.value }}</div>
      </div>
    </div>
    <div class="mt-[24px] mb-[14px] flex justify-between">
      <div class="flex gap-[12px]">
        <Dropdown
          theme="light"
          trigger="click"
          :popover-options="{
            clickContentAutoHide: true,
          }"
        >
          <Button :disabled="!failedSelection.length">
            <span>批量重试</span>
            <i
              class="nodeman-icon nc-arrow-down ml-[5px] text-[18px] text-[#979BA5]"
            ></i>
          </Button>
          <template #content>
            <Dropdown.DropdownMenu ext-cls="dropDown-menu">
              <Dropdown.DropdownItem
                class="text-14px"
                v-for="item in reTryType"
                :key="item.id"
                v-bk-tooltips="{
                  content: item.tooltip,
                }"
                @click="handleFullRetry(item.id)"
              >
                <Button
                  text
                  :disabled="!!failedSelection.find(el => el.state === 'terminated') && item.id === 'PARTIAL'"
                >{{ item.name }}</Button>
              </Dropdown.DropdownItem>
            </Dropdown.DropdownMenu>
          </template>
        </Dropdown>
        <Button :disabled="!runningSelection.length" @click="handleBatchTerminate">批量终止</Button>
        <copy-ip-dropdown
          type="agent"
          :list="list"
          :data="filterTableData"
          filter-prop="state"
          :disabled="!hasSelection"
        ></copy-ip-dropdown>
        <div
          class="h-[32px] bg-[#EAEBF0] rounded-[2px] flex items-center text-[12px] mr-[12px]"
        >
          <Radio.Group v-model="radioGroupValue" type="capsule" @change="handleChangeRadio">
            <Radio.Button
              v-for="item in radioGroup"
              :label="item.name"
              :key="item.name"
            >
              <i v-if="item.icon" :class="item.icon"></i>
              <span>{{ item.label }} ({{ item.count }})</span>
            </Radio.Button>
          </Radio.Group>
        </div>
      </div>
      <SearchSelect
        class="flex-1"
        ref="searchSelect"
        :data="searchSelectData"
        v-model.trim="searchSelectValue"
        :unique-select="true"
        :placeholder="'请输入IP、管控区域、业务、目标版本、执行状态 搜索'"
        @update:model-value="handleSearchSelectChange"
      >
      </SearchSelect>
    </div>
    <div class="relative">
      <Table
        class="filterTable"
        :data="filterTableData"
        :empty-text="'暂无数据'"
        :pagination="pagination"
        show-overflow-tooltip
        :max-height="maxHeight"
        :show-settings="isShowSetting"
        :settings="settings"
        @setting-change="handleSettingChange"
        @column-filter="handleFilter"
        @page-limit-change="pageLimitChange"
        @page-value-change="pageValueChange"
      >
        <template #prepend>
          <div v-if="hasSelection" class="flex items-center justify-center h-[30px] bg-[#ebecf0] text-[12px]">
            <template v-if="isCrossPageSelection">
              已跨页全选 <span class="font-bold mx-1">{{ pagination.count - excludedIds.size }}</span> 条，
              <Button text theme="primary" @click="handleClearSelection">取消选择</Button>
            </template>
            <template v-else>
              已选择 <span class="font-bold mx-1">{{ selection.length }}</span> 条，
              <Button
                text theme="primary" @click="handleSelectAllCrossPage">
                选择所有页共 {{ pagination.count }} 条
              </Button>
            </template>
          </div>
        </template>

        <TableColumn width="80" fixed="left">
          <template #header>
            <div class="flex items-center justify-start">
              <Checkbox
                :model-value="isCurrentPageAllChecked"
                :indeterminate="isIndeterminate"
                @change="handleHeaderClick" />
              <Dropdown trigger="click" placement="bottom-start">
                <i class="nodeman-icon nc-arrow-down ml-1 text-[18px]"></i>
                <template #content>
                  <Dropdown.DropdownMenu>
                    <Dropdown.DropdownItem @click="handleSelectCurrentPage">本页全选</Dropdown.DropdownItem>
                    <Dropdown.DropdownItem @click="handleSelectAllCrossPage">跨页全选</Dropdown.DropdownItem>
                  </Dropdown.DropdownMenu>
                </template>
              </Dropdown>
            </div>
          </template>
          <template #default="{ row }">
            <Checkbox class="mt-[6px]" :model-value="row.checked" @change="(val) => handleRowCheck(val, row)" />
          </template>
        </TableColumn>
        <TableColumn
          v-if="route.query.active === 'plugin'"
          field="plugin_name"
          :title="'插件名'"
          width="150"
          fixed="left"
        ></TableColumn>
        <TableColumn
          field="bk_host_inner_list"
          :title="'IPv4'"
          width="150"
          fixed="left"
        ></TableColumn>
        <TableColumn
          field="bk_host_innerip_v6_list"
          :title="'IPv6'"
          width="150"
        ></TableColumn>
        <TableColumn
          field="bk_networkarea_id"
          :title="'管控区域'"
          min-width="150"
        >
          <template #default="{ row }">
            {{ networkAreaListMap.get(row.bk_networkarea_id) }}
          </template>
        </TableColumn>
        <TableColumn
          field="bk_networkunit_id"
          :title="'管控单元'"
          min-width="150"
        >
          <template #default="{ row }">
            {{ networkUnitListMap.get(row.bk_networkunit_id) }}
          </template>
        </TableColumn>
        <TableColumn
          field="bk_biz_name"
          :title="'业务'"
          min-width="150"
        ></TableColumn>
        <TableColumn
          field="node_version"
          :title="'目标版本'"
          min-width="150"
          :filter="filterOptionSource.node_version"
        >
        </TableColumn>
        <TableColumn field="total_time_second" :title="'耗时'">
          <template #default="{ row }">
            <span>{{ formatTimeToMS(row.total_time_second) }}</span>
          </template>
        </TableColumn>
        <TableColumn
          field="state"
          :title="'执行状态'"
          :filter="filterOptionSource.state"
          min-width="120"
        >
          <template #default="{ row }">
            <div
              class="flex items-center"
              v-if="row.state && statusMap[row.state]"
            >
              <Spinner v-if="row.state === 'running'" class="mr-[8px]" />
              <template v-else>
                <i
                  :class="`nodeman-icon nc-${
                    statusMap[row.state].icon
                  } status-icon`"
                ></i>
              </template>
              <span>{{ statusMap[row.state].text }}</span>
            </div>
            <div class="flex items-center" v-else>
              <span class="nodeman-icon nc-unknown status-icon"></span>
              <span>{{ row.state }}</span>
            </div>
          </template>
        </TableColumn>
        <TableColumn
          field="reTryCount"
          :title="'重试次数'"
          min-width="100"
        ></TableColumn>
        <TableColumn :title="'操作'" fixed="right" width="200">
          <template #default="{ row }">
            <div class="flex items-center gap-[11px]">
              <Button text theme="primary" @click="handleViewLog(row)">查看日志</Button>
              <Button text theme="primary" :disabled="row.state !== 'running'" @click="handleTerminate(row)">终止</Button>
              <Dropdown
                theme="light"
                trigger="click"
                ext-cls="dropdownCls"
                :popover-options="{
                  clickContentAutoHide: true,
                }"
              >
                <Button
                  text
                  theme="primary"
                  v-if="!['success', 'running'].includes(row.state)"
                  class="flex items-stretch"
                >
                  <right-turn-line fill="#3A84FF" />
                  <span>重试</span>
                </Button>
                <template #content>
                  <Dropdown.DropdownMenu ext-cls="dropDown-menu">
                    <Dropdown.DropdownItem
                      class="text-14px"
                      v-for="item in reTryType"
                      :key="item.id"
                      v-bk-tooltips="{
                        content: item.tooltip,
                      }"
                      @click="handleRetry(row, item.id)"
                    >
                      <Button
                        text
                        :disabled="
                          row.state === 'terminated' && item.id === 'PARTIAL'
                        "
                      >{{ item.name }}</Button
                      >
                    </Dropdown.DropdownItem>
                  </Dropdown.DropdownMenu>
                </template>
              </Dropdown>
            </div>
          </template>
        </TableColumn>
      </Table>
    </div>
  </div>
</template>
<script setup lang="ts">
import {
  Button,
  Checkbox,
  Dropdown,
  Input,
  Radio,
  ResizeLayout,
  SearchSelect,
  Tag,
} from 'bkui-vue';
import {
  AngleUpFill,
  Close,
  RightTurnLine,
  Spinner,
  Success,
} from 'bkui-vue/lib/icon';
import dayjs from 'dayjs';
import {
  computed,
  onBeforeUnmount,
  onMounted,
  reactive,
  ref,
  watch,
} from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';

import { Table, TableColumn } from '@blueking/table';

import { NodeWorkflowService } from '@/api/modules/node_workflow';
import { PluginWorkflowService } from '@/api/modules/plugin_workflow';
import { TopoService } from '@/api/modules/topo';
import useInterval from '@/composables/use-interval';
import useTableSetting from '@/composables/use-table-setting';
import { useMainStore } from '@/stores/main';
import { useNodeManageStore } from '@/stores/node-manage';

interface FilterOption {
  list: { text: string; value: string }[];
  checked: string[];
  filterScope: string;
}
type taskType =
  | 'install_agent'
  | 'install_plugin'
  | 'upgrade_agent'
  | 'upgrade_plugin';
type filterProp = 'state' | 'node_version';

const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const mainStore = useMainStore();
const nodeManageStore = useNodeManageStore();
const reTryType = [
  {
    id: 'ALL',
    name: '重新开始执行',
    tooltip: '重新开始执行完整的任务',
  },
  {
    id: 'PARTIAL',
    name: '最近失败重试',
    tooltip: '从最近失败的步骤开始重试',
  },
];
const maxHeight = computed(() => mainStore.windowInnerHeight - 214);
const currentData = computed(() => nodeManageStore.taskHistoryTableRowData);

// 当前任务状态
const currentTaskStatus = computed(() => nodeManageStore.taskHistoryTableRowData.status);
const statusMap = computed(() => ({
  running: {
    text: '执行中',
    tagTheme: 'info',
  },
  failed: {
    text: '失败',
    icon: 'terminated',
    tagTheme: 'danger',
  },
  success: {
    text: '成功',
    icon: 'running',
    tagTheme: 'success',
  },
  partial_failed: {
    text: '部分失败',
    icon: 'warning',
    tagTheme: 'warning',
  },
  ignored: {
    text: '已忽略（没有需要变更的实例）',
    icon: 'warning',
    tagTheme: '',
  },
  timeout: {
    text: '超时',
    icon: 'terminated',
    tagTheme: '',
  },
  init: {
    text: '初始化',
    icon: 'unknown',
    tagTheme: '',
  },
  terminated: {
    text: '终止',
    icon: 'terminated',
  },
  launched: {
    text: '等待执行',
    icon: 'unknown',
  },
}));
const typeMap = computed(() => ({
  install_agent: t('platform.nodeMan.taskHistory.taskType.install_agent'),
  install_plugin: t('platform.nodeMan.taskHistory.taskType.install_plugin'),
  upgrade_agent: t('platform.nodeMan.taskHistory.taskType.upgrade_agent'),
  upgrade_plugin: t('platform.nodeMan.taskHistory.taskType.upgrade_plugin'),
  reconfig_agent: t('platform.nodeMan.taskHistory.taskType.reconfig_agent'),
  restart_agent: t('platform.nodeMan.taskHistory.taskType.restart_agent'),
  uninstall_agent: t('platform.nodeMan.taskHistory.taskType.uninstall_agent'),
  install_proxy: t('platform.nodeMan.taskHistory.taskType.install_proxy'),
  upgrade_proxy: t('platform.nodeMan.taskHistory.taskType.upgrade_proxy'),
  reconfig_proxy: t('platform.nodeMan.taskHistory.taskType.reconfig_proxy'),
  restart_proxy: t('platform.nodeMan.taskHistory.taskType.restart_proxy'),
  uninstall_proxy: t('platform.nodeMan.taskHistory.taskType.uninstall_proxy'),
}));

const formatTimeToMS = (duration = 0) => {
  // 处理非数字或负数情况
  if (typeof duration !== 'number' || duration < 0) {
    return '0m 0s';
  }

  // 智能判断单位：
  // 1. 大于等于100000的整数视为毫秒（时间戳差值通常较大）
  // 2. 小数视为毫秒（如1234.5毫秒）
  // 3. 较小的整数视为秒（如3600秒 = 1小时）
  const isMs = duration >= 100000 || !Number.isInteger(duration);
  const ms = isMs ? duration : duration * 1000;

  // 计算各时间单位
  const totalSeconds = Math.floor(ms / 1000);
  const hours = Math.floor(totalSeconds / 3600);
  const remainingSecondsAfterHours = totalSeconds % 3600;
  const minutes = Math.floor(remainingSecondsAfterHours / 60);
  const seconds = remainingSecondsAfterHours % 60;

  // 补零规则：非0且小于10时补零，0则直接显示0
  const padIfNeeded = (num: number) => (num === 0 ? '0' : num < 10 ? `0${num}` : num.toString());

  // 构建结果
  const parts = [];
  if (hours > 0) {
    parts.push(`${hours}h`);
    parts.push(`${padIfNeeded(minutes)}m`);
  } else if (minutes > 0) {
    parts.push(`${minutes}m`);
  }
  parts.push(`${padIfNeeded(seconds)}s`);

  return parts.join(' ');
};
const formatCostTime = (duration: number) => {
  const minutes = Math.floor(duration / 60000);
  const seconds = Math.floor((duration % 60000) / 1000);
  return `${minutes}m ${seconds}s`;
};

const timeFormatter = (
  val: number | string | undefined,
  format = 'YYYY-MM-DD HH:mm:ss',
) => (val ? dayjs(val).format(format) : '--');

// 精确返回到历史详情页面
const handleBackToHistory = () => {
  router.push({
    name: 'history',
    query: {
      active: route.query?.active,
    },
  });
};

const sliceWorkflowId = (val: string) => `#${val?.slice(-4)}`;
const taskInfoList = computed(() => [
  {
    prop: 'type',
    name: '任务类型',
    value:
      typeMap.value[nodeManageStore.taskHistoryTableRowData?.type as taskType]
      || nodeManageStore.taskHistoryTableRowData?.type,
  },
  {
    prop: 'cost_time',
    name: '总耗时',
    value: formatCostTime(nodeManageStore.taskHistoryTableRowData?.cost_time),
  },
  {
    prop: 'workflow_id',
    name: '任务ID',
    value: sliceWorkflowId(nodeManageStore.taskHistoryTableRowData?.workflow_id),
  },
  {
    prop: 'operator',
    name: '执行人',
    value: nodeManageStore.taskHistoryTableRowData?.operator,
  },
  {
    prop: 'operate_time',
    name: '执行时间',
    value: timeFormatter(nodeManageStore.taskHistoryTableRowData?.operate_time),
  },
]);

const tableData = ref<any[]>([]);
const filterTableData = computed(() => tableData.value.filter((item: any) => radioGroupValue.value === 'all' || item.state === radioGroupValue.value));
const radioGroupValue = ref('all');
const radioGroup = computed(() => [
  {
    icon: '',
    label: '全部',
    name: 'all',
    count: tableData.value.length,
  },
  {
    icon: 'nodeman-icon nc-running status-icon',
    label: '成功',
    name: 'success',
    count: tableData.value.filter((item: { state: string }) => item.state === 'success').length,
  },
  {
    icon: 'nodeman-icon nc-terminated status-icon',
    label: '失败',
    name: 'failed',
    count: tableData.value.filter((item: { state: string }) => item.state === 'failed').length,
  },
]);
const bussinessMap = computed(() => mainStore.businessList.map(item => ({
  id: item.bk_biz_id,
  name: item.bk_biz_name,
})));
const getUniqueChildren = (prop: string) => {
  const uniqueValues = Array.from(new Set(tableData.value
    .map((item: any) => item[prop])
    .filter((item: any) => item !== null && item !== undefined && item !== '')));
  return uniqueValues.map((value) => {
    let name;
    switch (prop) {
      case 'state':
        name = statusMap.value[value as string]?.text || String(value);
        break;
      case 'bk_networkarea_id':
        name = networkAreaListMap.get(value as number) || String(value);
        break;
      case 'bk_networkunit_id':
        name = networkUnitListMap.get(value as number) || String(value);
        break;
      default:
        String(value);
        break;
    }
    return {
      id: value,
      name,
    };
  });
};
const filterOptionConfig = (prop: string, textMap?: Record<string, any>) => {
  const uniqueValues = Array.from(new Set(tableData.value
    .map((item: any) => item[prop])
    ?.filter((item: any) => item)));
  return uniqueValues.map(value => ({
    text:
      textMap && textMap[value as string]
        ? textMap[value as string].text
        : value,
    value,
  }));
};
const filterOptionSource = reactive<Record<string, FilterOption>>({
  node_version: {
    list: [],
    checked: [],
    filterScope: 'all',
  },
  state: {
    list: [],
    checked: [],
    filterScope: 'all',
  },
});
const loading = ref(false);
// 搜索
const searchSelectValue = ref<{ id: string; name: string; values: any[] }[]>([]);
const searchSelectData = computed(() => [
  { id: 'bk_host_innerip', name: 'IPv4', multiple: true },
  { id: 'bk_host_innerip_v6', name: 'IPv6', multiple: true },
  {
    id: 'bk_networkarea_id',
    name: '管控区域',
    children: getUniqueChildren('bk_networkarea_id'),
    multiple: true,
  },
  {
    id: 'bk_networkunit_id',
    name: '管控单元',
    children: getUniqueChildren('bk_networkunit_id'),
    multiple: true,
  },
  {
    id: 'bk_biz_id',
    name: '业务',
    children: bussinessMap.value,
    multiple: true,
  },
  {
    id: 'node_version',
    name: '目标版本',
    children: getUniqueChildren('node_version'),
    multiple: true,
  },
  {
    id: 'state',
    name: '执行状态',
    children: getUniqueChildren('state'),
    multiple: true,
  },
]);
// eslint-disable-next-line max-len
const handleSearchSelectChange = async (data: { id: string; name: string; values: { id: string; name: string }[] }[]) => {
  Object.keys(filterOptionSource).forEach((key) => {
    filterOptionSource[key].checked = [];
  });
  data.forEach((item) => {
    if (filterOptionSource[item.id as filterProp]) {
      filterOptionSource[item.id as filterProp].checked = item.values.map((item: any) => item.id) as string[];
    }
  });
};

// 分页
const pagination = reactive({ count: 0, limit: 50, current: 1, remote: true });
const pageLimitChange = async (limit: number) => {
  pagination.limit = limit;
  pagination.current = 1; // 页码重置为1
  await getOperateList(); // 分页变化不防抖，立即执行
};

const pageValueChange = async (current: number) => {
  pagination.current = current;
  await getOperateList(); // 分页变化不防抖，立即执行
};

// 复制
const list = [
  {
    id: 'all',
    name: '所有IP',
    children: [
      {
        id: 'ipv4',
        name: 'IPv4',
      },
      {
        id: 'ipv6',
        name: 'IPv6',
      },
    ],
  },
  {
    id: 'ignore',
    name: '被忽略IP',
    children: [
      {
        id: 'ipv4',
        name: 'IPv4',
      },
      {
        id: 'ipv6',
        name: 'IPv6',
      },
    ],
  },
  {
    id: 'failed',
    name: '失败IP',
    children: [
      {
        id: 'ipv4',
        name: 'IPv4',
      },
      {
        id: 'ipv6',
        name: 'IPv6',
      },
    ],
  },
  {
    id: 'success',
    name: '成功IP',
    children: [
      {
        id: 'ipv4',
        name: 'IPv4',
      },
      {
        id: 'ipv6',
        name: 'IPv6',
      },
    ],
  },
];

// 表格勾选
const selection = computed(() => filterTableData.value.filter((item: any) => item.checked));
const failedSelection = computed(() => selection.value.filter((item: any) => ['failed', 'timeout', 'terminated'].includes(item.state)));
const runningSelection = computed(() => selection.value.filter((item: any) => ['running'].includes(item.state)));

// --- 跨页全选核心状态 ---
const isCrossPageSelection = ref(false); // 是否开启跨页全选模式
const excludedIds = ref<Set<number>>(new Set()); // 全选模式下，用户手动“取消勾选”的 ID 集合

// 计算属性：是否有任何选中（用于禁用批量按钮）
const hasSelection = computed(() => filterTableData.value.some(item => item.checked) || isCrossPageSelection.value);

// 计算属性：当前页是否全选（用于表头 Checkbox 状态）
// eslint-disable-next-line max-len
const isCurrentPageAllChecked = computed(() => filterTableData.value.length > 0 && filterTableData.value.every(item => item.checked));
const isIndeterminate = computed(() => {
  const selectedCount = filterTableData.value.filter(item => item.checked).length;
  return selectedCount > 0 && selectedCount < filterTableData.value.length;
});

// 1. 处理单行勾选
const handleRowCheck = (checked: boolean, row: any) => {
  row.checked = checked;
  if (isCrossPageSelection.value) {
    if (!checked) excludedIds.value.add(row.bk_host_id);
    else excludedIds.value.delete(row.bk_host_id);
  }
};

// 2. 跨页全选
const handleSelectAllCrossPage = () => {
  isCrossPageSelection.value = true;
  excludedIds.value.clear();
  filterTableData.value.forEach(item => (item.checked = true));
};

// 3. 取消选择
const handleClearSelection = () => {
  isCrossPageSelection.value = false;
  excludedIds.value.clear();
  filterTableData.value.forEach(item => (item.checked = false));
};

// 4. 本页全选
const handleSelectCurrentPage = () => {
  isCrossPageSelection.value = false;
  filterTableData.value.forEach(item => (item.checked = true));
};

// 5. 表头 Checkbox 快速切换
const handleHeaderClick = () => {
  isCurrentPageAllChecked.value ? handleClearSelection() : handleSelectCurrentPage();
};

const handleChangeRadio = (value: string) => {
  isCrossPageSelection.value = false;
  excludedIds.value.clear();
  filterTableData.value.forEach(item => (item.checked = false));
};

// 表格设置
const { isShowSetting, settings, handleSettingChange } = useTableSetting(
  {
    checked: [
      'bk_host_inner_list',
      'bk_host_innerip_v6_list',
      'bk_networkarea_id',
      'bk_networkunit_id',
      'bk_biz_name',
      'node_version',
      'total_time_second',
      'state',
      'reTryCount',
      'plugin_name',
    ],
    disabled: [],
  },
  'nodeMng-task-detail',
);

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
      name: field,
      values: checked.map((item: any) => {
        const name = field === 'state' ? statusMap.value[item]?.text || item : item;
        return {
          id: item,
          name,
        };
      }),
    });
  }
};

const networkAreaListMap = new Map<number, string | number>([[-1, -1]]);
// 管控区域下拉列表获取
const getNetworkAreaList = async (data: {bk_networkarea_id: number}[]) => {
  const res = await TopoService.NetworkAreaList({
    page: {
      limit: 0,
    },
    exact_include_conditions: {
      bk_networkarea_id: data.map(item => item.bk_networkarea_id),
    },
  }).catch((err: any) => {
    console.log(err);
    return {
      total: 0,
      items: [],
    };
  });
  res.items.forEach((item) => {
    networkAreaListMap.set(item.bk_networkarea_id, item.bk_networkarea_name);
  });
};
// 管控单元下拉列表获取
const networkUnitListMap = new Map<number, string | number>([[-1, -1]]);
const getNetworkUnitList = async (data: {bk_networkunit_id: number}[]) => {
  const res = await TopoService.NetworkUnitList({
    exact_include_conditions: {
      bk_networkunit_id: data.map(item => item.bk_networkunit_id),
    },
  }).catch((err: any) => {
    console.log(err);
    return {
      total: 0,
      items: [],
    };
  });
  res.items.forEach((item) => {
    networkUnitListMap.set(item.bk_networkunit_id, item.bk_networkunit_name);
  });
};

// 统一服务调用器
const serviceCaller = {
  // 根据路由参数获取当前服务类型
  getCurrentServiceType: () => (route.query.active === 'node' ? 'node' : 'plugin'),

  // 服务方法映射
  serviceMethods: {
    node: {
      retry: NodeWorkflowService.NodeWorkflowOperationRetry,
      terminate: NodeWorkflowService.NodeWorkflowOperationTerminate,
      workflowList: NodeWorkflowService.NodeWorkflowList,
      operationList: NodeWorkflowService.NodeWorkflowOperationList,
    },
    plugin: {
      retry: PluginWorkflowService.PluginWorkflowOperationRetry,
      terminate: PluginWorkflowService.PluginWorkflowOperationTerminate,
      workflowList: PluginWorkflowService.PluginWorkflowList,
      operationList: PluginWorkflowService.PluginWorkflowOperationList,
    },
  },

  // 统一调用方法
  async call(method: 'retry' | 'terminate' | 'workflowList' | 'operationList', params: any) {
    const serviceType = this.getCurrentServiceType();
    const serviceMethod = this.serviceMethods[serviceType][method];
    return await serviceMethod(params);
  },
};
// 重试
const handleRetry = async (row: any, type: string) => {
  const res = await serviceCaller.call('retry', {
    workflow_id: route.params.taskId,
    operation_ids: [row.operation_id],
    retry_mod: type,
  }).catch(() => false);
  if (res !== false) {
    await getOperateList();
    if (currentTaskStatus.value === 'running' && needInterval.value) {
      start();
    }
    await updataCurrentTaskInfo();
  }
};
const handleFullRetry = async (type: string) => {
  const res = await serviceCaller.call('retry', {
    workflow_id: route.params.taskId,
    operation_ids: failedSelection.value.map(item => item.operation_id),
    retry_mod: type,
  }).catch(() => false);
  if (res) {
    await getOperateList();
    if (currentTaskStatus.value === 'running' && needInterval.value) {
      start();
    }
    await updataCurrentTaskInfo();
  }
};

// 终止
const handleTerminate = async (row: any) => {
  const res = await serviceCaller.call('terminate', {
    workflow_id: route.params.taskId,
    operation_ids: [row.operation_id],
  }).catch(() => false);
  if (res !== false) {
    await getOperateList();
    if (currentTaskStatus.value === 'running' && needInterval.value) {
      start();
    }
    await updataCurrentTaskInfo();
  }
};
// 批量终止
const handleBatchTerminate = async () => {
  const res = await serviceCaller.call('terminate', {
    workflow_id: route.params.taskId,
    operation_ids: failedSelection.value.map(item => item.operation_id),
  }).catch(() => false);
  if (res !== false) {
    await getOperateList();
    if (currentTaskStatus.value === 'running' && needInterval.value) {
      start();
    }
    await updataCurrentTaskInfo();
  }
};

const updataCurrentTaskInfo = async () => {
  const res = await serviceCaller.call('workflowList', {
    exact_include_conditions: {
      bk_biz_id: mainStore.selectedBusinessId,
      workflow_id: [route.params.taskId],
    },
  }).catch((err) => {
    console.log(err);
    return {
      total: 0,
      items: [],
    };
  });
  const list = res.items.map(item => ({
    ...item,
    bk_biz_name: item.bk_biz_name?.filter(item => item),
    cost_time: item.finish_time > 0 ? item.finish_time - item.operate_time : 0,
  }));
  const findItem = list.find((item: any) => item.workflow_id === route.params.taskId);
  if (findItem) {
    // 在任务list中找到当前任务，并更新任务信息和任务状态
    nodeManageStore.updateCurrentRowData(findItem);
    // 任务状态优先于子任务状态
    subTasksStatus.value = [findItem.status];
  }
};
const getParams = () => {
  const params = {
    page: {
      limit: pagination.limit,
      offset: (pagination.current - 1) * pagination.limit,
    },
    exact_include_conditions: {} as Record<string, string[] | string>,
    fuzzy_include_conditions: {} as Record<string, string[]>,
  };
  searchSelectValue.value.forEach((item: any) => {
    const target = params.exact_include_conditions;
    target[item.id] = item.values?.map((value: any) => value.id);
  });
  params.exact_include_conditions['workflow_id'] = route.params.taskId;
  return params;
};
// 所有子任务状态
const subTasksStatus = ref<string[]>();
// 如果所有子任务状态包含运行中或者初始状态则需要轮询，即needInterval为true
const needInterval = computed(() => subTasksStatus.value?.includes('running')
    || subTasksStatus.value?.includes('empty_instance')
    || subTasksStatus.value?.includes('init'));
const getOperateList = async () => {
  subTasksStatus.value = [];
  const searchParameters = getParams();
  const res = await serviceCaller.call('operationList', searchParameters).catch(() => ({
    operations: [],
    total: 0,
  }));
  pagination.count = res.total;
  const mapList = res.operations.map((item) => {
    subTasksStatus.value?.push(item.status.state);
    return {
      ...item.param,
      ...item.status,
      bk_host_innerip: item.param.bk_host_inner_list?.join(',') || item.param.bk_host_innerip_list?.join(','),
      bk_host_innerip_v6: item.param.bk_host_innerip_v6_list?.join(',') || item.param.bk_host_innerip_v6_list?.join(','),
      bk_host_inner_list: item.param.bk_host_inner_list?.join(',') || item.param.bk_host_innerip_list?.join(','),
      bk_host_innerip_v6_list: item.param.bk_host_innerip_v6_list?.join(',') || item.param.bk_host_innerip_v6_list?.join(','),
      bk_biz_name: mainStore.businessList.find(biz => biz.bk_biz_id === item.param.bk_biz_id)?.bk_biz_name
         || item.param.bk_biz_id,
      operation_id: item.operation_id,
      reTryCount: item.instance_ids?.length ? item.instance_ids?.length - 1 : 0,
      node_version: route.query.active === 'node' ? item.param.node_version : item.param.plugin_version,
    };
  });
  const isEqual = tableData.value.length === mapList.length
    && tableData.value.every((item, index) => item.state === mapList[index]?.state);
  const isLengthEqual = tableData.value.length === mapList.length;
  if (!isLengthEqual && mapList.length > 0) {
    await Promise.all([
      getNetworkAreaList(mapList),
      getNetworkUnitList(mapList),
    ]);
  }
  if (!isEqual) {
    tableData.value = mapList;
    tableData.value.forEach((item) => {
      let isChecked = false;
      if (isCrossPageSelection.value) {
        // 如果是跨页模式，只要不在排除名单里就是选中
        isChecked = !excludedIds.value.has(item.bk_host_id);
      }
      item.checked = isChecked;
    });
  }
};
const handleViewLog = async (row: any) => {
  router.push({
    name: 'log',
    params: {
      hostId: row.bk_host_id,
      taskId: route.params.taskId,
    },
    query: {
      active: route.query?.active,
    },
  });
};
const { start, stop } = useInterval(getOperateList, 1000); // 轮询
// 日志中执行失败或者任务详情表中都失败则更新详情的信息状态
const handleStop = async () => {
  await updataCurrentTaskInfo();
  await getOperateList();
};
watch(
  () => searchSelectValue,
  async () => {
    await getOperateList();
  },
  { deep: true },
);
watch(
  () => tableData,
  () => {
    filterOptionSource.node_version.list = filterOptionConfig('node_version');
    filterOptionSource.state.list = filterOptionConfig('state', statusMap.value);
  },
  { deep: true, immediate: true },
);
watch(
  () => needInterval.value,
  async () => {
    if (!needInterval.value) {
      stop();
      await updataCurrentTaskInfo();
    } else {
      start();
    }
  },
);

// 日志页面点击重试触发此页面的list的数据轮询
watch(
  () => mainStore.isLogRetry,
  (val: Boolean) => {
    if (val) {
      start();
      mainStore.updateLogRetry(false);
    }
  },
);
// 日志页面点击终止触发此页面的list的数据轮询
watch(
  () => mainStore.isLogTerminate,
  (val: Boolean) => {
    if (val) {
      start();
      mainStore.updateLogRetry(false);
    }
  },
);

onMounted(async () => {
  if (route.query.status) {
    searchSelectValue.value.push({
      id: 'state',
      name: '执行状态',
      values: [{
        id: route.query.status,
        name: statusMap.value[route.query.status]?.text || route.query.status,
      }],
    });
  }
  await updataCurrentTaskInfo();
  await getOperateList();
  if (currentTaskStatus.value === 'running' && needInterval.value) {
    start();
  }
});
onBeforeUnmount(() => {
  stop();
});
</script>
<style lang="postcss" scoped>
.status-icon::before {
  content: "";
  display: inline-block;
  margin-right: 8px;
  width: 8px;
  height: 8px;
  border: 1px solid #f0f1f5;
  border-radius: 6.5px;
  background: #b2b5bd;
}
.nc-running {
  &::before {
    background: #cbf0da;
    border-color: #2caf5e;
  }
}
.nc-terminated {
  &::before {
    border-color: #ea3636;
    background: #ffdddd;
  }
}
.dropdownCls {
  :deep(span) {
    display: inline-block;
    vertical-align: middle;
  }
}
</style>
