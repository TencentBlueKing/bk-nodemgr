<template>
  <PageHeader
    class="w-full absolute top-0 z-100"
    :title="'任务详情'"
    :back="true"
  >
    <span class="mx-[6px] text-[#979BA5] text-[14px]">-</span>
    <span class="text-[#979BA5] text-[14px] mr-[14px]" v-if="currentData">{{ currentData?.workflow_id }}</span>
    <Tag :theme="statusMap[currentTaskStatus]?.tagTheme" type="filled" v-if="currentTaskStatus">
      {{ statusMap[currentTaskStatus]?.text || '' }}
    </Tag>
  </PageHeader>
  <div class="p-[24px] mt-[52px]">
    <div class="flex">
      <div v-for="item in taskInfoList" :key="item.name" class="leading-[30px] mr-[52px] text-[12px]">
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
          }">
          <Button :disabled="!failedSelection.length">全部失败重试</Button>
          <template #content>
            <Dropdown.DropdownMenu ext-cls="dropDown-menu">
              <Dropdown.DropdownItem
                class="text-14px"
                v-for="item in reTryType"
                :key="item.id"
                @click="handleFullRetry(item.id)">
                {{ item.name }}
              </Dropdown.DropdownItem>
            </Dropdown.DropdownMenu>
          </template>
        </Dropdown>
        <copy-ip-dropdown
          type="agent"
          :list="list"
          :data="tableData"
          filter-prop="state"
          :disabled="!selection.length"
        ></copy-ip-dropdown>
        <div class="h-[32px] bg-[#EAEBF0] rounded-[2px] flex items-center text-[12px] mr-[12px]">
          <Radio.Group
            v-model="radioGroupValue"
            type="capsule"
          >
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
        @update:model-value="handleSearchSelectChange">
      </SearchSelect>
    </div>
    <bk-loading title="数据加载中" :loading="loading">
      <div class="relative">
        <Table
          class="filterTable"
          :data="filterTableData"
          :empty-text="'暂无数据'"
          :pagination="pagination"
          show-overflow-tooltip
          :max-height="maxHeight"
          @checkbox-change="handleSelectChange"
          @checkbox-all="handleSelectAllChange"
          :show-settings="isShowSetting"
          :settings="settings"
          @setting-change="handleSettingChange"
          @column-filter="handleFilter"
        >
          <TableColumn type="checkbox" width="80" fixed="left"></TableColumn>
          <TableColumn field="bk_host_inner" :title="'IPv4'" width="150" fixed="left"></TableColumn>
          <TableColumn field="bk_host_innerip_v6" :title="'IPv6'" width="150"></TableColumn>
          <TableColumn field="bk_networkarea_id" :title="'管控区域'" min-width="150"></TableColumn>
          <TableColumn field="bk_biz_name" :title="'业务'" min-width="150"></TableColumn>
          <TableColumn
            field="node_version"
            :title="'目标版本'"
            min-width="150"
            :filter="filterOptionSource.node_version">
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
              <div class="flex items-center" v-if="row.state && statusMap[row.state]">
                <Spinner v-if="row.state === 'running'" class="mr-[8px]" />
                <template v-else>
                  <i :class="`nodeman-icon nc-${statusMap[row.state].icon} status-icon`"></i>
                </template>
                <span>{{ statusMap[row.state].text }}</span>
              </div>
              <div class="flex items-center" v-else>
                <span class="nodeman-icon nc-unknown status-icon"></span>
                <span>{{ row.state }}</span>
              </div>
            </template>
          </TableColumn>
          <TableColumn field="reTryCount" :title="'重试次数'" min-width="100"></TableColumn>
          <TableColumn
            :title="'操作'"
            fixed="right"
            width="150"
          >
            <template #default="{ row }">
              <div class="flex items-center">
                <Button text theme="primary" class="mr-[11px]" @click="handleViewLog(row)">查看日志</Button>
                <Dropdown
                  theme="light"
                  trigger="click"
                  ext-cls="dropdownCls"
                  :popover-options="{
                    clickContentAutoHide: true,
                  }">
                  <Button text theme="primary" v-if="!['success', 'running'].includes(row.state)">
                    <right-turn-line fill="#3A84FF" />
                    <span>重试</span>
                  </Button>
                  <template #content>
                    <Dropdown.DropdownMenu ext-cls="dropDown-menu">
                      <Dropdown.DropdownItem
                        class="text-14px"
                        v-for="item in reTryType"
                        :key="item.id"
                        @click="handleRetry(row, item.id)">
                        {{ item.name }}
                      </Dropdown.DropdownItem>
                    </Dropdown.DropdownMenu>
                  </template>
                </Dropdown>
              </div>
            </template>
          </TableColumn>
        </Table>
        <Log :data="curRow" v-if="curRow" ref="logRef" @stop="handleStop"></Log>
      </div>
    </bk-loading>
  </div>
</template>
<script setup lang="ts">
import { Button, Dropdown, Input, Radio, ResizeLayout, SearchSelect, Tag } from 'bkui-vue';
import { AngleUpFill, Close, RightTurnLine, Spinner, Success } from 'bkui-vue/lib/icon';
import dayjs from 'dayjs';
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute } from 'vue-router';

import { Table, TableColumn } from '@blueking/table';

import Log from './log.vue';

import { NodeWorkflowService } from '@/api/modules/node_workflow';
import useInterval from '@/composables/use-interval';
import usePage from '@/composables/use-page';
import useTableSetting from '@/composables/use-table-setting';
import { useMainStore } from '@/stores/main';
import { useNodeManageStore } from '@/stores/node-manage';

interface FilterOption {
  list: { text: string, value: string }[];
  checked: string[];
  filterScope: string;
}
type taskType = 'install_agent' | 'install_plugin' | 'upgrade_agent' | 'upgrade_plugin';
type filterProp = 'state' | 'node_version';

const { t } = useI18n();
const route = useRoute();
const mainStore = useMainStore();
const nodeManageStore = useNodeManageStore();
const reTryType = [
  {
    id: 'full_node_instance_retry',
    name: '全部重试',
  },
  {
    id: 'partial_node_instance_retry',
    name: '部分重试',
  },
];
const maxHeight = computed(() => mainStore.windowInnerHeight - 214);
const currentData = computed(() => nodeManageStore.taskHistoryTableRowData);

// 当前任务状态
const currentTaskStatus = computed(() => nodeManageStore.taskHistoryTableRowData.status);
const statusMap = {
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
    icon: 'unknown',
    tagTheme: '',
  },
  init: {
    text: '初始化',
    icon: 'unknown',
    tagTheme: '',
  },
};
const typeMap = {
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
};

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
  const padIfNeeded = (num: number) => num === 0 ? '0' : num < 10 ? `0${num}` : num.toString();

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

const timeFormatter = (val: number | string | undefined, format = 'YYYY-MM-DD HH:mm:ss') => (val ? dayjs(val).format(format) : '--');

const sliceWorkflowId = (val: string) => `#${val?.slice(-4)}`;
const taskInfoList = computed(() => ([
  { prop: 'type', name: '任务类型', value: typeMap[nodeManageStore.taskHistoryTableRowData?.type as taskType] || nodeManageStore.taskHistoryTableRowData?.type },
  { prop: 'cost_time', name: '总耗时', value: formatCostTime(nodeManageStore.taskHistoryTableRowData?.cost_time) },
  { prop: 'workflow_id', name: '任务ID', value: sliceWorkflowId(nodeManageStore.taskHistoryTableRowData?.workflow_id) },
  { prop: 'operator', name: '执行人', value: nodeManageStore.taskHistoryTableRowData?.operator },
  { prop: 'operate_time', name: '执行时间', value: timeFormatter(nodeManageStore.taskHistoryTableRowData?.operate_time) },
]));

const tableData = ref<any[]>([]);
const filterTableData = computed(() => tableData.value.filter((item: any) => radioGroupValue.value === 'all'
        || item.state === radioGroupValue.value));
const radioGroupValue = ref('all');
const curOperationId = ref('');
const radioGroup = computed(() => ([
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
    count: tableData.value.filter((item: {state: string}) => item.state === 'success').length,
  },
  {
    icon: 'nodeman-icon nc-terminated status-icon',
    label: '失败',
    name: 'failed',
    count: tableData.value.filter((item: {state: string}) => item.state === 'failed').length,
  },
]));
const bussinessMap = computed(() => mainStore.businessList.map(item => ({
  id: item.bk_biz_id,
  name: item.bk_biz_name,
})));
const getUniqueChildren = (prop: string) => {
  const uniqueValues = Array.from(new Set(tableData.value.map((item: any) => item[prop]).filter((item: any) => item)));
  return uniqueValues.map(value => ({
    id: value,
    name: prop === 'state' ? statusMap[value as string]?.text : String(value),
  }));
};
const filterOptionConfig = (prop: string, textMap?: Record<string, any>) => {
  const uniqueValues = Array.from(new Set(tableData.value.map((item: any) => item[prop])?.filter((item: any) => item)));
  return uniqueValues.map(value => ({
    text: textMap && textMap[value as string] ? textMap[value as string].text : value,
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
const searchSelectValue = ref<{id: string, name: string, values: any[]}[]>([]);
const searchSelectData = computed(() => [
  { id: 'bk_host_innerip', name: 'IP', multiple: true },
  { id: 'bk_networkarea_id', name: '管控区域', children: getUniqueChildren('bk_networkarea_id'), multiple: true },
  { id: 'bk_biz_id', name: '业务', children: bussinessMap.value, multiple: true },
  { id: 'node_version', name: '目标版本', children: getUniqueChildren('node_version'), multiple: true },
  { id: 'state', name: '执行状态', children: getUniqueChildren('state'), multiple: true },
]);
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

// 分页
const {
  pagination,
} = usePage(tableData);
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
const selection = computed(() => tableData.value.filter((item: any) => item.checked));
const failedSelection = computed(() => selection.value.filter((item: any) => ['failed', 'timeout'].includes(item.state)));
const handleSelectChange = ({ checked, row }: {checked: boolean, row: any}) => {
  row.checked = checked;
};

// 表格全选
const handleSelectAllChange = ({ checked }: { checked: boolean}) => {
  tableData.value.forEach((item: any) => item.checked = checked);
};
// 表格设置
const { isShowSetting, settings, handleSettingChange } = useTableSetting({
  checked: [
    'bk_host_inner',
    'bk_host_innerip_v6',
    'bk_networkarea_id',
    'bk_biz_name',
    'node_version',
    'total_time_second',
    'state',
    'reTryCount',
  ],
  disabled: [],
}, 'nodeMng-task-detail');

// 筛选
const handleFilter = ({ checked, field }: {checked: string[], field: string}) => {
  const index = searchSelectValue.value.findIndex((item: any) => item.id === field);
  index > -1 && searchSelectValue.value.splice(index, 1);
  if (checked.length) {
    searchSelectValue.value.push({ id: field, name: field, values: checked.map((item: any) => {
      const name = field === 'state' ? statusMap[item].text : item;
      return {
        id: item,
        name,
      };
    }) });
  }
};
// 重试
const handleRetry = async (row: any, type: string) => {
  const res = await NodeWorkflowService.NodeWorkflowOperationRetry({
    workflow_id: route.params.taskId,
    operation_id: [row.operation_id],
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
const handleFullRetry = async (type: string) => {
  const res = await NodeWorkflowService.NodeWorkflowOperationRetry({
    workflow_id: route.params.taskId,
    operation_id: failedSelection.value.map(item => item.operation_id),
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
const updataCurrentTaskInfo = async () => {
  const res = await NodeWorkflowService.NodeWorkflowList({
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
    bk_biz_name: item.bk_biz_name.filter(item => item),
    cost_time: item.finish_time > 0 ? (item.finish_time - item.operate_time) : 0,
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
      limit: 0,
      offset: 0,
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
const needInterval = computed(() => subTasksStatus.value?.includes('running') || subTasksStatus.value?.includes('empty_instance') || subTasksStatus.value?.includes('init'));
const getOperateList = async () => {
  subTasksStatus.value = [];
  const currentRowData = nodeManageStore.taskHistoryTableRowData;
  const searchParameters = getParams();
  const res = await NodeWorkflowService.NodeWorkflowOperationList(searchParameters).catch(() => ({
    operations: [],
    total_count: 0,
  }));
  const mapList = res.operations.map((item) => {
    subTasksStatus.value?.push(item.status.state);
    return {
      ...item.param,
      ...item.status,
      bk_biz_name: currentRowData?.bk_biz_name || item.bk_biz_id,
      operation_id: item.operation_id,
      reTryCount: item.instance_ids.length - 1,
    };
  });
  const isEqual =
    tableData.value.length === mapList.length &&
    tableData.value.every((item, index) => item.state === mapList[index]?.state);

  if (!isEqual) {
    tableData.value = mapList;
  }

};
const logRef = ref<InstanceType<typeof Log>>();
const curRow = ref(null);
const handleViewLog = async (row: any) => {
  curOperationId.value = row.operation_id;
  curRow.value = row;
  nextTick(() => {
    logRef.value?.show();
  });
};
const { start, stop } = useInterval(getOperateList, 1000); // 轮询
// 日志中执行失败或者任务详情表中都失败则更新详情的信息状态
const handleStop = async () => {
  await updataCurrentTaskInfo();
  await getOperateList();
};
watch(() => searchSelectValue, async () => {
  await getOperateList();
}, { deep: true });
watch(() => tableData, () => {
  filterOptionSource.node_version.list = filterOptionConfig('node_version', typeMap);
  filterOptionSource.state.list = filterOptionConfig('state', statusMap);
}, { deep: true, immediate: true });
watch(() => needInterval.value, async () => {
  if (!needInterval.value) {
    stop();
    await updataCurrentTaskInfo();
  } else {
    start();
  }
});
onMounted(async () => {
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
  content: '';
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
    background: #CBF0DA;
    border-color: #2CAF5E;
  }
}
.nc-terminated {
  &::before {
    border-color: #EA3636;
    background: #FFDDDD;
  }
}
.dropdownCls {
    :deep(span) {
        display: inline-block;
        vertical-align: middle;
    }
}
</style>
