<template>
  <PageHeader
    class="w-full absolute top-0 z-100"
    :title="$t('taskDetail.title')"
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
        <div>{{ item.name }}</div>
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
            <span>{{ $t('taskDetail.batchRetry') }}</span>
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
        <Button
          :disabled="!runningSelection.length"
          @click="handleBatchTerminate">
          {{ $t('taskDetail.batchTerminate') }}
        </Button>
        <copy-ip-dropdown
          type="agent"
          :list="list"
          :data="tableData"
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
        class="flex-1 bg-[#fff]"
        ref="searchSelect"
        :data="searchSelectData"
        v-model.trim="searchSelectValue"
        :unique-select="true"
        :placeholder="$t('platform.nodeMan.historySearchPlaceholder')"
        @update:model-value="handleSearchSelectChange"
        @paste.native="handleNativePaste"
      >
      </SearchSelect>
    </div>
    <div class="relative">
      <Table
        class="filterTable"
        :data="tableData"
        :empty-text="$t('table.empty')"
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
        <!-- <template #prepend>
          <div v-if="hasSelection" class="flex items-center justify-center h-[30px] bg-[#ebecf0] text-[12px]">
            <template v-if="isCrossPageSelection">
              <span>{{ $t('taskDetail.table.crossPageSelected') }}</span>
              <span class="font-bold mx-1"> {{ pagination.count - excludedIds.size }} </span>
              <span>{{ $t('taskDetail.table.items') }}</span>
              <Button text theme="primary" @click="handleClearSelection">
                {{ $t('taskDetail.table.cancelSelection') }}
              </Button>
            </template>
            <template v-else>
              <span>{{ $t('taskDetail.table.selected') }}</span>
              <span class="font-bold mx-1"> {{ selection.length }} </span>
              <span> {{ $t('taskDetail.table.items') }}</span>
              <Button
                text theme="primary" @click="handleSelectAllCrossPage">
                {{ $t('taskDetail.table.selectAllPages', { x: pagination.count }) }}
              </Button>
            </template>
          </div>
        </template> -->

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
                    <Dropdown.DropdownItem @click="handleSelectCurrentPage">
                      {{ $t('taskDetail.filter.currentPage') }}
                    </Dropdown.DropdownItem>
                    <!-- <Dropdown.DropdownItem @click="handleSelectAllCrossPage">
                      {{ $t('taskDetail.table.crossSelected') }}
                    </Dropdown.DropdownItem> -->
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
          :title="$t('taskDetail.table.pluginName')"
          width="150"
          fixed="left"
        ></TableColumn>
        <TableColumn
          field="bk_host_inner_list"
          :title="$t('taskDetail.table.ipv4')"
          width="150"
          fixed="left"
        ></TableColumn>
        <TableColumn
          field="bk_host_innerip_v6_list"
          :title="$t('taskDetail.table.ipv6')"
          width="150"
        ></TableColumn>
        <TableColumn
          field="bk_networkarea_id"
          :title="$t('taskDetail.table.workarea')"
          :filter="filterOptionSource.bk_networkarea_id"
          min-width="150"
        >
          <template #default="{ row }">
            {{ networkAreaListMap.get(row.bk_networkarea_id) }}
          </template>
        </TableColumn>
        <TableColumn
          field="bk_networkunit_id"
          :title="$t('taskDetail.table.workUnit')"
          :filter="filterOptionSource.bk_networkunit_id"
          min-width="150"
        >
          <template #default="{ row }">
            {{ networkUnitListMap.get(row.bk_networkunit_id) }}
          </template>
        </TableColumn>
        <TableColumn
          field="bk_biz_id"
          :title="$t('taskDetail.table.business')"
          :filter="filterOptionSource.bk_biz_id"
          min-width="150"
        >
          <template #default="{ row }">
            {{ bizListMap.get(row.bk_biz_id) || row.bk_biz_id }}
          </template>
        </TableColumn>
        <TableColumn
          v-if="route.query.active !== 'plugin'"
          field="node_version"
          :title="$t('taskDetail.table.targetVersion')"
          min-width="150"
        >
        </TableColumn>
        <TableColumn
          v-else
          field="plugin_version"
          :title="$t('taskDetail.table.targetVersion')"
          min-width="150"
        >
        </TableColumn>
        <TableColumn field="total_time_second" :title="$t('taskDetail.table.costTime')">
          <template #default="{ row }">
            <span>{{ formatCostTime(row.total_time_second) }}</span>
          </template>
        </TableColumn>
        <TableColumn
          field="state"
          :title="$t('taskDetail.table.status')"
          :filter="filterOptionSource.state"
          :min-width="stateMinWidth"
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
              <div>
                <!-- eslint-disable-next-line max-len -->
                <div v-if="isOfflineGuideStep(row.latest_action_inst_brief_data?.tags ?? []) && row.state === 'running'">
                  {{ $t('taskDetail.table.waitOffline') }}
                  <Button class="ml-[2px]" text theme="primary" @click="handleOfflineGuide(row)">
                    {{ $t('taskDetail.table.offlineGuide') }}
                  </Button>
                </div>
                <!-- eslint-disable-next-line max-len -->
                <div v-else-if="row.latest_action_inst_brief_data?.tags.includes('need_manual_exec_install_script') && row.state === 'running'">
                  {{ $t('taskDetail.table.waitManual') }}
                  <Button class="ml-[2px]" text theme="primary" @click="handleOperateGuide(row)">
                    {{ $t('taskDetail.table.guide') }}
                  </Button>
                </div>
                <span v-else>{{ statusMap[row.state].text }}</span>
              </div>
            </div>
            <div class="flex items-center" v-else>
              <span class="nodeman-icon nc-unknown status-icon"></span>
              <span>{{ row.state }}</span>
            </div>
          </template>
        </TableColumn>
        <TableColumn
          field="reTryCount"
          :title="$t('taskDetail.table.retryCount')"
          min-width="100"
        ></TableColumn>
        <TableColumn :title="$t('taskDetail.table.action')" fixed="right" width="200">
          <template #default="{ row }">
            <div class="flex items-center gap-[11px]">
              <Button text theme="primary" @click="handleViewLog(row)">{{ $t('taskDetail.table.viewLog') }}</Button>
              <Button text theme="primary" :disabled="row.state !== 'running'" @click="handleTerminate(row)">
                {{ $t('taskDetail.table.terminate') }}
              </Button>
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
                  <span>{{ $t('taskDetail.table.retry') }}</span>
                </Button>
                <template #content>
                  <Dropdown.DropdownMenu ext-cls="dropDown-menu">
                    <Dropdown.DropdownItem
                      class="text-14px"
                      v-for="item in reTryType"
                      :key="item.id"
                      v-bk-tooltips="{
                        content: item.tooltip,
                        placement: 'left',
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
  <guide v-model:is-show="isGuideShow" :data="guideData" />
</template>
<script setup lang="ts">
import {
  Button,
  Checkbox,
  Dropdown,
  InfoBox,
  Input,
  Message,
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
import { debounce } from 'lodash';
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

import guide from './guide.vue';
import { isOfflineGuideStep } from './offline-package';

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

interface IWorkflowStatisticsInfo {
  failed_count: number;
  init_count: number;
  launched_count: number;
  running_count: number;
  success_count: number;
  terminated_count: number;
  timeout_count: number;
  total_count: number;
  workflow_id: string;
}
type taskType =
  | 'install_agent'
  | 'upgrade_agent'
  | 'reconfig_agent'
  | 'restart_agent'
  | 'uninstall_agent'
  | 'install_proxy'
  | 'upgrade_proxy'
  | 'reconfig_proxy'
  | 'restart_proxy'
  | 'uninstall_proxy'
  | 'assign_proxy_unit'
  | 'install_plugin'
  | 'ensure_plugin_v2'
  | 'stop_plugin_v2'
  | 'upgrade_plugin'
  | 'uninstall_plugin'
  | 'uninstall_plugin_v2'
  | 'migrate_plugin_v2'
  | 'reconfig_plugin'
  | 'apply_plugin_subconfig'
  | 'restart_plugin'
  | 'stop_plugin';
type filterProp = 'state' | 'node_version';

const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const mainStore = useMainStore();
const nodeManageStore = useNodeManageStore();
const reTryType = computed(() => [
  {
    id: 'ALL',
    name: t('taskDetail.retryType.all'),
    tooltip: t('taskDetail.retryType.allTooltip'),
  },
  {
    id: 'PARTIAL',
    name: t('taskDetail.retryType.partial'),
    tooltip: t('taskDetail.retryType.partialTooltip'),
  },
]);
const maxHeight = computed(() => mainStore.windowInnerHeight - 264 - (mainStore.noticeShow ? 40 : 0));
const currentData = computed(() => nodeManageStore.taskHistoryTableRowData);
const stateMinWidth = computed(() => (tableData.value.some(item =>
  (item.latest_action_inst_brief_data?.tags?.includes('need_manual_exec_install_script')
    || isOfflineGuideStep(item.latest_action_inst_brief_data?.tags ?? []))
  && item.state === 'running') ? 240 : 120));

// 当前任务状态
const currentTaskStatus = computed(() => nodeManageStore.taskHistoryTableRowData.status);
const statusMap = computed(() => ({
  running: {
    text: t('taskDetail.status.running'),
    tagTheme: 'info',
  },
  failed: {
    text: t('taskDetail.status.failed'),
    icon: 'failed',
    tagTheme: 'danger',
  },
  success: {
    text: t('taskDetail.status.success'),
    icon: 'success',
    tagTheme: 'success',
  },
  partial_failed: {
    text: t('taskDetail.status.partial_failed'),
    icon: 'warning',
    tagTheme: 'warning',
  },
  ignored: {
    text: t('taskDetail.status.ignored'),
    icon: 'warning',
    tagTheme: '',
  },
  timeout: {
    text: t('taskDetail.status.timeout'),
    icon: 'timeout',
    tagTheme: '',
  },
  init: {
    text: t('taskDetail.status.init'),
    icon: 'incomplete',
    tagTheme: '',
  },
  terminated: {
    text: t('taskDetail.status.terminated'),
    icon: 'terminated',
  },
  launched: {
    text: t('taskDetail.status.launched'),
    icon: 'incomplete',
  },
  incomplete: {
    text: t('taskDetail.status.incomplete'),
    icon: 'incomplete',
  },
}));
const typeMap = computed(() => ({
  install_agent: t('platform.nodeMan.taskHistory.taskType.install_agent'),
  upgrade_agent: t('platform.nodeMan.taskHistory.taskType.upgrade_agent'),
  reconfig_agent: t('platform.nodeMan.taskHistory.taskType.reconfig_agent'),
  restart_agent: t('platform.nodeMan.taskHistory.taskType.restart_agent'),
  uninstall_agent: t('platform.nodeMan.taskHistory.taskType.uninstall_agent'),
  install_proxy: t('platform.nodeMan.taskHistory.taskType.install_proxy'),
  upgrade_proxy: t('platform.nodeMan.taskHistory.taskType.upgrade_proxy'),
  reconfig_proxy: t('platform.nodeMan.taskHistory.taskType.reconfig_proxy'),
  restart_proxy: t('platform.nodeMan.taskHistory.taskType.restart_proxy'),
  uninstall_proxy: t('platform.nodeMan.taskHistory.taskType.uninstall_proxy'),
  assign_proxy_unit: t('platform.nodeMan.taskHistory.taskType.assign_proxy_unit'),
  install_plugin: t('platform.nodeMan.taskHistory.taskType.install_plugin'),
  ensure_plugin_v2: t('platform.nodeMan.taskHistory.taskType.ensure_plugin_v2'),
  stop_plugin_v2: t('platform.nodeMan.taskHistory.taskType.stop_plugin_v2'),
  upgrade_plugin: t('platform.nodeMan.taskHistory.taskType.upgrade_plugin'),
  uninstall_plugin: t('platform.nodeMan.taskHistory.taskType.uninstall_plugin'),
  uninstall_plugin_v2: t('platform.nodeMan.taskHistory.taskType.uninstall_plugin_v2'),
  migrate_plugin_v2: t('platform.nodeMan.taskHistory.taskType.migrate_plugin_v2'),
  reconfig_plugin: t('platform.nodeMan.taskHistory.taskType.reconfig_plugin'),
  apply_plugin_subconfig: t('platform.nodeMan.taskHistory.taskType.apply_plugin_subconfig'),
  restart_plugin: t('platform.nodeMan.taskHistory.taskType.restart_plugin'),
  stop_plugin: t('platform.nodeMan.taskHistory.taskType.stop_plugin'),
}));

const networkAreaListMap = ref(new Map<number, string>([]));
const networkUnitListMap = ref(new Map<number, string>([[-1, t('taskDetail.table.unassigned')]]));

// 补零规则：非0且小于10时补零，0则直接显示0
const padIfNeeded = (num: number) => (num === 0 ? '0' : num < 10 ? `0${num}` : num.toString());
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
  const seconds = padIfNeeded(Math.floor((duration % 60000) / 1000));
  return `${minutes}m ${seconds}s`;
};

const timeFormatter = (
  val: number | string | undefined,
  format = 'YYYY-MM-DD HH:mm:ss',
) => (val ? dayjs(val).format(format) : '--');

// 精确返回到历史列表页面，保留筛选条件
const handleBackToHistory = () => {
  const filters = nodeManageStore.consumeHistoryFilters();
  if (filters) {
    // 从 store 恢复的筛选状态重建 URL query
    const query: Record<string, string> = { active: filters.active };
    filters.searchSelectValue.forEach((item) => {
      if (item.values?.length) {
        query[`f_${item.id}`] = JSON.stringify(item.values);
      }
    });
    if (filters.dateStart > 0 && filters.dateEnd > 0) {
      query.date_start = String(filters.dateStart);
      query.date_end = String(filters.dateEnd);
    }
    query.page = String(filters.page);
    query.limit = String(filters.limit);
    if (filters.hideAutoTask) {
      query.hide_auto = '1';
    }
    router.push({ name: 'history', query });
  } else {
    router.push({
      name: 'history',
      query: {
        active: route.query?.active as string || 'agent',
      },
    });
  }
};

const sliceWorkflowId = (val: string) => `#${val?.slice(-4)}`;
const taskInfoList = computed(() => [
  {
    prop: 'type',
    name: t('taskDetail.info.type'),
    value:
      typeMap.value[nodeManageStore.taskHistoryTableRowData?.type as taskType]
      || nodeManageStore.taskHistoryTableRowData?.type,
  },
  {
    prop: 'cost_time',
    name: t('taskDetail.info.costTime'),
    value: formatCostTime(nodeManageStore.taskHistoryTableRowData?.cost_time),
  },
  {
    prop: 'workflow_id',
    name: t('taskDetail.info.taskId'),
    value: sliceWorkflowId(nodeManageStore.taskHistoryTableRowData?.workflow_id),
  },
  {
    prop: 'operator',
    name: t('taskDetail.info.operator'),
    value: nodeManageStore.taskHistoryTableRowData?.operator,
  },
  {
    prop: 'operate_time',
    name: t('taskDetail.info.operateTime'),
    value: timeFormatter(nodeManageStore.taskHistoryTableRowData?.operate_time),
  },
]);

const tableData = ref<any[]>([]);
// 从接口获取的状态列表
const distinctStates = ref<string[]>([]);

const radioGroupValue = ref('all');
const radioGroup = computed(() => [
  {
    icon: '',
    label: t('taskDetail.status.all'),
    name: 'all',
    count: statistics.value.total_count,
  },
  {
    icon: 'nodeman-icon nc-incomplete status-icon',
    label: t('taskDetail.status.incomplete'),
    name: 'incomplete',
    count: statistics.value.init_count + statistics.value.launched_count + statistics.value.running_count,
  },
  {
    icon: 'nodeman-icon nc-success status-icon',
    label: t('taskDetail.status.success'),
    name: 'success',
    count: statistics.value.success_count,
  },
  {
    icon: 'nodeman-icon nc-failed status-icon',
    label: t('taskDetail.status.failed'),
    name: 'failed',
    count: statistics.value.failed_count,
  },
  {
    icon: 'nodeman-icon nc-timeout status-icon',
    label: t('taskDetail.status.timeout'),
    name: 'timeout',
    count: statistics.value.timeout_count,
  },
  {
    icon: 'nodeman-icon nc-terminated status-icon',
    label: t('taskDetail.status.terminated'),
    name: 'terminated',
    count: statistics.value.terminated_count,
  },
]);
const bussinessMap = computed(() => mainStore.businessList.map(item => ({
  id: item.bk_biz_id,
  name: item.bk_biz_name,
})));
// eslint-disable-next-line max-len
const bizListMap = computed(() => new Map<number, string>(mainStore.businessList.map((item: any) => [item.bk_biz_id, item.bk_biz_name])));

// 获取状态去重列表
const getDistinctStates = async () => {
  // 没有taskId，不请求distinctStates
  if (!route.params.taskId) return;

  try {
    const res = await serviceCaller.call('operationDistinct', {
      workflow_id: route.params.taskId,
    });
    if (res && res.state) {
      distinctStates.value = res.state;
    }
  } catch (error) {
    console.error('获取状态去重列表失败:', error);
    // 如果接口调用失败，回退到从当前列表获取
    distinctStates.value = Array.from(new Set(tableData.value
      .map((item: any) => item.state)
      .filter((item: any) => item !== null && item !== undefined && item !== '')));
  }
};

const getUniqueChildren = (prop: string) => {
  if (prop === 'state' && distinctStates.value.length > 0) {
    // 使用接口返回的状态列表
    return distinctStates.value.map((value) => {
      const name = statusMap.value[value as string]?.text || String(value);
      return {
        id: value,
        name,
      };
    });
  }

  // 其他属性保持原有逻辑
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
        name = networkAreaListMap.value.get(value as number) || String(value);
        break;
      case 'bk_networkunit_id':
        name = networkUnitListMap.value.get(value as number) || String(value);
        break;
      default:
        name = String(value);
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
const getFilterList = (prop: string) => {
  switch (prop) {
    case 'bk_biz_id':
      return bussinessMap.value.map((item: any) => ({ value: String(item.id), text: item.name }));
    case 'bk_networkarea_id':
      return Array.from(networkAreaListMap.value, ([id, name]) => ({ value: String(id), text: name }));
    case 'bk_networkunit_id':
      return Array.from(networkUnitListMap.value, ([id, name]) => ({ value: String(id), text: name }));
    default:
      return [];
  }
};
const filterOptionSource = reactive<Record<string, FilterOption>>({
  // node_version: {
  //   list: [],
  //   checked: [],
  //   filterScope: 'all',
  // },
  bk_biz_id: {
    list: [],
    checked: [],
    filterScope: 'all',
  },
  bk_networkarea_id: {
    list: [],
    checked: [],
    filterScope: 'all',
  },
  bk_networkunit_id: {
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
const IPV4_REG = /^((25[0-5]|2[0-4]\d|[01]?\d\d?)\.){3}(25[0-5]|2[0-4]\d|[01]?\d\d?)$/;
const IPV6_REG = /^(?:[A-F0-9]{1,4}:){7}[A-F0-9]{1,4}$/i;
const AREA_IP_REG = /^(\d+):(.+)$/; // 管控区域ID:IP 格式

const searchSelectData = computed(() => [
  { id: 'ip', name: 'IP', multiple: true },
  {
    id: 'area_ip',
    name: `${t('platform.nodeMan.bk_cloud_name')}ID:IP`,
    multiple: true,
  },
  {
    id: 'bk_networkarea_id',
    name: t('taskDetail.search.workarea'),
    children: Array.from(networkAreaListMap.value, ([id, name]) => ({ id: String(id), name })),
    multiple: true,
  },
  {
    id: 'bk_networkunit_id',
    name: t('taskDetail.search.workUnit'),
    children: Array.from(networkUnitListMap.value, ([id, name]) => ({ id: String(id), name })),
    multiple: true,
  },
  {
    id: 'bk_biz_id',
    name: t('taskDetail.search.business'),
    children: bussinessMap.value,
    multiple: true,
  },
  {
    id: 'node_version',
    name: t('taskDetail.search.targetVersion'),
    children: [],
    multiple: true,
  },
  {
    id: 'state',
    name: t('taskDetail.search.status'),
    children: getUniqueChildren('state'),
    multiple: true,
  },
]);
/**
 * 解析多分隔符输入，支持空格、换行、分号、逗号
 */
const parseMultiDelimiterInput = (text: string): string[] => text
  .split(/[\s\n;,|、]+/)
  .map(item => item.trim())
  .filter(item => item.length > 0);

/**
 * 智能识别输入类型
 */
const detectInputType = (text: string): { type: 'ip' | 'area_ip' | null; value: string } => {
  if (AREA_IP_REG.test(text)) return { type: 'area_ip', value: text };
  if (IPV4_REG.test(text)) return { type: 'ip', value: text };
  if (IPV6_REG.test(text)) return { type: 'ip', value: text };
  return { type: null, value: text };
};

/**
 * 拦截原生 paste 事件，将空格分隔符转换为组件能识别的逗号
 */
const handleNativePaste = (event: ClipboardEvent) => {
  const text = event.clipboardData?.getData('text');
  if (!text) return;
  if (text.includes(' ') && !/[|,、\r\n\n]/.test(text)) {
    event.preventDefault();
    const normalizedText = text.replace(/\s+/g, ',');
    const target = event.target as HTMLElement;
    if (target && target.isContentEditable) {
      document.execCommand('insertText', false, normalizedText);
    }
  }
};

/**
 * 处理粘贴/快速输入的逻辑
 */
const handleInputPaste = (data: { id: string; name: string; values: { id: string; name: string }[] }[]) => {
  if (data.length === 0) return;

  const lastItem = data[data.length - 1];

  if (['ip', 'area_ip'].includes(lastItem.id) && lastItem.values.length > 0) {
    const allParsedItems: string[] = [];
    lastItem.values.forEach((value: any) => {
      const parsedItems = parseMultiDelimiterInput(value.id);
      allParsedItems.push(...parsedItems);
    });
    const uniqueItems = Array.from(new Set(allParsedItems));
    if (uniqueItems.length > 0) {
      lastItem.values = uniqueItems.map(item => ({ id: item, name: item }));
      return;
    }
  }

  const searchFieldIds = new Set(searchSelectData.value.map(item => item.id));
  const tailRawInputIds: string[] = [];
  for (let i = data.length - 1; i >= 0; i--) {
    const current = data[i];
    if (searchFieldIds.has(current.id) || current.values?.length) break;
    tailRawInputIds.unshift(current.id);
  }

  const parsedItems = (tailRawInputIds.length > 0
    ? tailRawInputIds
    : [lastItem.id])
    .flatMap(text => parseMultiDelimiterInput(text));
  const uniqueParsedItems = Array.from(new Set(parsedItems));
  if (uniqueParsedItems.length === 0) return;

  const firstDetection = detectInputType(uniqueParsedItems[0]);
  if (!firstDetection.type) return;

  const allSameType = uniqueParsedItems.every(item => detectInputType(item).type === firstDetection.type);
  if (!allSameType) return;

  let targetId = '';
  let targetName = '';

  switch (firstDetection.type) {
    case 'ip':
      targetId = 'ip';
      targetName = 'IP';
      break;
    case 'area_ip':
      targetId = 'area_ip';
      targetName = `${t('platform.nodeMan.bk_cloud_name')}ID:IP`;
      break;
  }

  if (targetId) {
    searchSelectValue.value = searchSelectValue.value.filter((item: any) => {
      if (item.id === targetId) return false;
      return !tailRawInputIds.includes(item.id);
    });
    searchSelectValue.value.push({
      id: targetId,
      name: targetName,
      values: uniqueParsedItems.map(item => ({ id: item, name: item })),
    });
  }
};

// eslint-disable-next-line max-len
const handleSearchSelectChange = async (data: { id: string; name: string; values: { id: string; name: string }[] }[]) => {
  handleInputPaste(data);
  // 当搜素条件的执行状态变化时，都要触发radioGroup的变化
  const stateSearchItem = data.find((item) => item.id === 'state');
  if (stateSearchItem) {
    const state = stateSearchItem.values[0].id;
    radioGroupValue.value = ['init', 'launched', 'running'].includes(state) ? 'incomplete' : state;
  } else {
    radioGroupValue.value = 'all';
  }

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
const list = computed(() => [
  {
    id: 'all',
    name: t('taskDetail.copy.all'),
    children: [
      {
        id: 'ipv4',
        name: t('taskDetail.search.ipv4'),
      },
      {
        id: 'ipv6',
        name: t('taskDetail.search.ipv6'),
      },
    ],
  },
  {
    id: 'ignore',
    name: t('taskDetail.copy.ignored'),
    children: [
      {
        id: 'ipv4',
        name: t('taskDetail.search.ipv4'),
      },
      {
        id: 'ipv6',
        name: t('taskDetail.search.ipv6'),
      },
    ],
  },
  {
    id: 'failed',
    name: t('taskDetail.copy.failed'),
    children: [
      {
        id: 'ipv4',
        name: t('taskDetail.search.ipv4'),
      },
      {
        id: 'ipv6',
        name: t('taskDetail.search.ipv6'),
      },
    ],
  },
  {
    id: 'success',
    name: t('taskDetail.copy.success'),
    children: [
      {
        id: 'ipv4',
        name: t('taskDetail.search.ipv4'),
      },
      {
        id: 'ipv6',
        name: t('taskDetail.search.ipv6'),
      },
    ],
  },
]);

// 表格勾选
const selection = computed(() => tableData.value.filter((item: any) => item.checked));
const failedSelection = computed(() => selection.value.filter((item: any) => ['failed', 'timeout', 'terminated'].includes(item.state)));
const runningSelection = computed(() => selection.value.filter((item: any) => ['running'].includes(item.state)));

// --- 跨页全选核心状态 ---
const isCrossPageSelection = ref(false); // 是否开启跨页全选模式
const excludedIds = ref<Set<number>>(new Set()); // 全选模式下，用户手动"取消勾选"的 ID 集合

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
  if (!row) return;
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

const handleChangeRadio = (state: string) => {
  // 切换状态tab，重置选择状态
  isCrossPageSelection.value = false;
  excludedIds.value.clear();
  tableData.value.forEach(item => (item.checked = false));

  // 清除状态搜索条件
  const index = searchSelectValue.value.findIndex((item: any) => item.id === 'state');
  if (index > -1) searchSelectValue.value.splice(index, 1);

  // 清除筛选状态
  filterOptionSource.state.checked = [];

  if (state === 'all') return; // 全选不需要更新搜索条件
  const values = state !== 'incomplete'
    ? [{
      id: state,
      name: statusMap.value[state]?.text || state,
    }]
    : [
      { id: 'init', name: t('taskDetail.status.init') },
      { id: 'launched', name: t('taskDetail.status.launched') },
      { id: 'running', name: t('taskDetail.status.running') },
    ];
  // 更新状态搜索条件
  searchSelectValue.value.push({
    id: 'state',
    name: t('taskDetail.state'),
    values,
  });
  // 更新状态筛选
  filterOptionSource.state.checked = values.map((item: any) => item.id) as string[];
};

// 表格设置
const { isShowSetting, settings, handleSettingChange } = useTableSetting(
  {
    checked: [
      'bk_host_inner_list',
      'bk_host_innerip_v6_list',
      'bk_networkarea_id',
      'bk_networkunit_id',
      'bk_biz_id',
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
        let name;
        switch (field) {
          case 'bk_biz_id':
            name = bizListMap.value.get(Number(item)) || item;
            break;
          case 'bk_networkarea_id':
            name = networkAreaListMap.value.get(Number(item)) || item;
            break;
          case 'bk_networkunit_id':
            name = networkUnitListMap.value.get(Number(item)) || item;
            break;
          case 'state':
            name = statusMap.value[item]?.text || item;
            break;
          default:
            break;
        }
        const id = ['bk_biz_id', 'bk_networkarea_id', 'bk_networkunit_id'].includes(field) ? Number(item) : item;
        return {
          id,
          name,
        };
      }),
    });
  }
};

// 管控区域下拉列表获取
const getNetworkAreaList = async () => {
  const res = await TopoService.NetworkAreaList({
    page: {
      limit: 0,
    },
    exact_include_conditions: {
      bk_networkarea_id: [],
    },
  }).catch((err: any) => {
    console.log(err);
    return {
      total: 0,
      items: [],
    };
  });
  res.items.reverse().forEach((item) => {
    networkAreaListMap.value.set(item.bk_networkarea_id, item.bk_networkarea_name);
  });
  filterOptionSource.bk_networkarea_id.list = getFilterList('bk_networkarea_id');
};
// 管控单元下拉列表获取
const getNetworkUnitList = async () => {
  const res = await TopoService.NetworkUnitListBrief({
    exact_include_conditions: {
      bk_networkunit_id: [],
    },
  }).catch((err: any) => {
    console.log(err);
    return {
      total: 0,
      items: [],
    };
  });
  res.items.reverse().forEach((item) => {
    networkUnitListMap.value.set(item.bk_networkunit_id, item.bk_networkunit_name);
  });
  filterOptionSource.bk_networkunit_id.list = getFilterList('bk_networkunit_id');
};

// 统一服务调用器
const serviceCaller = {
  // 根据路由参数获取当前服务类型
  getCurrentServiceType: () => (route.query.active === 'plugin' ? 'plugin' : 'node'),

  // 服务方法映射
  serviceMethods: {
    node: {
      retry: NodeWorkflowService.NodeWorkflowOperationRetry,
      terminate: NodeWorkflowService.NodeWorkflowOperationTerminate,
      workflowList: NodeWorkflowService.NodeWorkflowList,
      operationList: NodeWorkflowService.NodeWorkflowOperationList,
      statistics: NodeWorkflowService.NodeWorkflowStatistics,
      operationDistinct: NodeWorkflowService.NodeWorkflowOperationDistinct,
    },
    plugin: {
      retry: PluginWorkflowService.PluginWorkflowOperationRetry,
      terminate: PluginWorkflowService.PluginWorkflowOperationTerminate,
      workflowList: PluginWorkflowService.PluginWorkflowList,
      operationList: PluginWorkflowService.PluginWorkflowOperationList,
      statistics: PluginWorkflowService.PluginWorkflowStatistics,
      operationDistinct: PluginWorkflowService.PluginWorkflowOperationDistinct,
    },
  },

  // 统一调用方法
  async call(method: 'retry' | 'terminate' | 'workflowList' | 'operationList' | 'statistics' | 'operationDistinct', params: any) {
    const serviceType = this.getCurrentServiceType();
    const serviceMethod = this.serviceMethods[serviceType][method];

    // 为workflowList请求添加不可取消配置，避免路由切换时被取消
    if (method === 'workflowList') {
      return await serviceMethod(params, { irrevocable: true });
    }

    return await serviceMethod(params);
  },
};

// 获取statistics
const statistics = ref<IWorkflowStatisticsInfo>({
  failed_count: 0,
  init_count: 0,
  launched_count: 0,
  running_count: 0,
  success_count: 0,
  terminated_count: 0,
  timeout_count: 0,
  total_count: 0,
  workflow_id: '',
});
const getStatistics = async () => {
  // 没有taskId，不请求statistics
  if (!route.params.taskId) return;
  const res = await serviceCaller.call('statistics', {
    workflow_id: [route.params.taskId],
  }).catch((err) => {
    console.log(err);
    return {
      items: [],
    };
  });
  const workflowStatisticsInfoItem = res.items.find(item => item.workflow_id === route.params.taskId);
  if (workflowStatisticsInfoItem) {
    statistics.value = workflowStatisticsInfoItem;
  };
};
const debouncedGetStatistics = debounce(getStatistics, 500);

// 重试
const handleRetry = async (row: any, type: string) => {
  if (!route.params.taskId) return;

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
  if (!route.params.taskId) return;

  const res = await serviceCaller.call('retry', {
    workflow_id: route.params.taskId,
    operation_ids: failedSelection.value.map(item => item.operation_id),
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

// 终止
const handleTerminate = (row: any) => {
  if (!route.params.taskId) return;
  InfoBox({
    title: t('taskDetail.table.terminateConfirmTitle'),
    subTitle: t('taskDetail.table.terminateConfirmSubTitle'),
    onConfirm: async () => {
      const res = await serviceCaller.call('terminate', {
        workflow_id: route.params.taskId,
        operation_ids: [row.operation_id],
      }).catch(() => {
        Message({ theme: 'error', message: t('taskDetail.table.terminateFailedMsg') });
        return false;
      });
      if (res !== false) {
        await getOperateList();
        if (currentTaskStatus.value === 'running' && needInterval.value) {
          start();
        }
        await updataCurrentTaskInfo();
      }
    },
  });
};
// 批量终止
const handleBatchTerminate = () => {
  if (!route.params.taskId) return;
  InfoBox({
    title: t('taskDetail.table.terminateConfirmTitle'),
    subTitle: t('taskDetail.table.terminateBatchConfirmSubTitle', { count: runningSelection.value.length }),
    onConfirm: async () => {
      const res = await serviceCaller.call('terminate', {
        workflow_id: route.params.taskId,
        operation_ids: runningSelection.value.map(item => item.operation_id),
      }).catch(() => {
        Message({ theme: 'error', message: t('taskDetail.table.terminateFailedMsg') });
        return false;
      });
      if (res !== false) {
        await getOperateList();
        if (currentTaskStatus.value === 'running' && needInterval.value) {
          start();
        }
        await updataCurrentTaskInfo();
      }
    },
  });
};

const updataCurrentTaskInfo = async () => {
  const res = await serviceCaller.call('workflowList', {
    page: {
      limit: pagination.limit,
      offset: (pagination.current - 1) * pagination.limit,
    },
    exact_include_conditions: {
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
    cost_time: item.finish_time > 0 ? item.finish_time - item.operate_time : new Date().getTime() - item.operate_time,
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
    exact_include_conditions: {} as Record<string, any>,
    workflow_id: route.params.taskId,
  };
  searchSelectValue.value.forEach((item: any) => {
    // IP 字段：自动识别 IPv4/IPv6 并分类
    if (item.id === 'ip' && item.values?.length) {
      const ipv4List: string[] = [];
      const ipv6List: string[] = [];
      item.values.forEach((value: any) => {
        if (IPV4_REG.test(value.id)) {
          ipv4List.push(value.id);
        } else if (IPV6_REG.test(value.id)) {
          ipv6List.push(value.id);
        }
      });
      if (ipv4List.length > 0) {
        params.exact_include_conditions.bk_host_innerip = ipv4List;
      }
      if (ipv6List.length > 0) {
        params.exact_include_conditions.bk_host_innerip_v6 = ipv6List;
      }
      return;
    }

    // 管控区域ID:IP：自动拆分为 bk_networkarea_id + IP 列表
    if (item.id === 'area_ip' && item.values?.length) {
      const areaIds = new Set<number>();
      const ipv4List: string[] = [];
      const ipv6List: string[] = [];
      item.values.forEach((value: any) => {
        const match = AREA_IP_REG.exec(value.id);
        if (match) {
          const areaId = Number(match[1]);
          const ip = match[2];
          areaIds.add(areaId);
          if (IPV4_REG.test(ip)) {
            ipv4List.push(ip);
          } else if (IPV6_REG.test(ip)) {
            ipv6List.push(ip);
          }
        }
      });
      if (areaIds.size > 0) {
        params.exact_include_conditions.bk_networkarea_id = Array.from(areaIds) as any;
      }
      if (ipv4List.length > 0) {
        params.exact_include_conditions.bk_host_innerip = ipv4List;
      }
      if (ipv6List.length > 0) {
        params.exact_include_conditions.bk_host_innerip_v6 = ipv6List;
      }
      return;
    }

    const target = params.exact_include_conditions;
    target[item.id] = item.values?.map((value: any) => value.id);
  });
  return params;
};
// 所有子任务状态
const subTasksStatus = ref<string[]>();
// 如果所有子任务状态包含运行中或者初始状态则需要轮询，即needInterval为true
const needInterval = computed(() => subTasksStatus.value?.includes('running')
    || subTasksStatus.value?.includes('empty_instance')
    || subTasksStatus.value?.includes('init'));
const getOperateList = async () => {
  // 没有taskId，不请求operationList
  if (!route.params.taskId) return;

  subTasksStatus.value = [];
  const searchParameters = getParams();
  const res = await serviceCaller.call('operationList', searchParameters).catch(() => ({
    operations: [],
    total: 0,
  }));
  pagination.count = res.total;
  const mapList = res.operations.map((item) => {
    const briefData = item.latest_oper_inst_brief_data;
    briefData && subTasksStatus.value?.push(briefData.life_cycle.state);
    const endTime = briefData ? briefData.life_cycle.end_time : new Date().getTime();
    if (route.query.active !== 'plugin') {
      return {
        ...item.node_deployment_info,
        ...(briefData ? briefData.life_cycle : {}),
        latest_action_inst_brief_data: briefData ? briefData.latest_action_inst_brief_data : {},
        bk_host_innerip: item.node_deployment_info.bk_host_inner_list?.join(',') || item.node_deployment_info.bk_host_innerip_list?.join(','),
        bk_host_innerip_v6: item.node_deployment_info.bk_host_innerip_v6_list?.join(',') || item.node_deployment_info.bk_host_innerip_v6_list?.join(','),
        bk_host_inner_list: item.node_deployment_info.bk_host_inner_list?.join(',') || item.node_deployment_info.bk_host_innerip_list?.join(','),
        bk_host_innerip_v6_list: item.node_deployment_info.bk_host_innerip_v6_list?.join(',') || item.node_deployment_info.bk_host_innerip_v6_list?.join(','),
        operation_id: item.operation_id,
        reTryCount: item.instance_ids?.length ? item.instance_ids?.length - 1 : 0,
        node_version: item.node_deployment_info.node_version,
        total_time_second: endTime > 0 ? endTime - item.create_time : new Date().getTime() - item.create_time,
      };
    }
    return {
      ...item.plugin_deployment_info,
      ...(briefData ? briefData.life_cycle : {}),
      latest_action_inst_brief_data: briefData ? briefData.latest_action_inst_brief_data : {},
      bk_host_innerip: item.plugin_deployment_info.bk_host_inner_list?.join(',') || item.plugin_deployment_info.bk_host_innerip_list?.join(','),
      bk_host_innerip_v6: item.plugin_deployment_info.bk_host_innerip_v6_list?.join(',') || item.plugin_deployment_info.bk_host_innerip_v6_list?.join(','),
      bk_host_inner_list: item.plugin_deployment_info.bk_host_inner_list?.join(',') || item.plugin_deployment_info.bk_host_innerip_list?.join(','),
      bk_host_innerip_v6_list: item.plugin_deployment_info.bk_host_innerip_v6_list?.join(',') || item.plugin_deployment_info.bk_host_innerip_v6_list?.join(','),
      operation_id: item.operation_id,
      reTryCount: item.instance_ids?.length ? item.instance_ids?.length - 1 : 0,
      node_version: item.plugin_deployment_info.node_version,
      total_time_second: endTime > 0 ? endTime - item.create_time : new Date().getTime() - item.create_time,
    };
  });
  const isEqual = tableData.value.length === mapList.length
    && tableData.value.every((item, index) => item.state === mapList[index]?.state
      && item.total_time_second === mapList[index]?.total_time_second);
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
  // 轮询期间实时更新任务头部耗时
  const taskRow = nodeManageStore.taskHistoryTableRowData;
  if (taskRow && taskRow.finish_time <= 0 && taskRow.operate_time > 0) {
    nodeManageStore.updateCurrentRowData({
      ...taskRow,
      cost_time: new Date().getTime() - taskRow.operate_time,
    });
  }
};
const handleViewLog = async (row: any) => {
  router.push({
    name: 'log',
    params: {
      hostId: route.query?.active !== 'plugin' ? row.bk_host_id : `${row.bk_host_id}_${row.plugin_name}`,
      taskId: route.params.taskId,
    },
    query: {
      active: route.query?.active,
      status: route.query?.status,
    },
  });
};
const { start, stop } = useInterval(getOperateList, 1000); // 轮询
// 日志中执行失败或者任务详情表中都失败则更新详情的信息状态
const handleStop = async () => {
  await updataCurrentTaskInfo();
  await getOperateList();
};

// 操作指引侧边栏
const isGuideShow = ref(false);
const guideData = ref<{
  workflow_id: string;
  operation_id: string;
  bk_host_innerip: string;
  bk_networkarea_id?: number;
  is_offline?: boolean;
}>({
  workflow_id: '',
  operation_id: '',
  bk_host_innerip: '',
  is_offline: false,
});
const handleOperateGuide = (row: any) => {
  isGuideShow.value = true;
  guideData.value = {
    workflow_id: route.params.taskId as string,
    operation_id: row.operation_id,
    bk_host_innerip: row.bk_host_innerip,
    is_offline: false,
  };
};
const handleOfflineGuide = (row: any) => {
  isGuideShow.value = true;
  guideData.value = {
    workflow_id: route.params.taskId as string,
    operation_id: row.operation_id,
    bk_host_innerip: row.bk_host_innerip,
    bk_networkarea_id: row.bk_networkarea_id,
    is_offline: true,
  };
};
watch(
  () => mainStore.businessList,
  () => {
    if (mainStore.businessList.length > 0) {
      filterOptionSource.bk_biz_id.list = getFilterList('bk_biz_id');
    }
  },
  { immediate: true }
);
watch(
  () => searchSelectValue,
  async () => {
    await getOperateList();
  },
  { deep: true },
);

// 移除这个监听器，避免覆盖接口返回的状态数据
// watch(
//   () => tableData,
//   () => {
//     filterOptionSource.state.list = filterOptionConfig('state', statusMap.value);
//   },
//   { deep: true, immediate: true },
// );

// 监听distinctStates变化，更新筛选选项
watch(distinctStates, (newStates) => {
  if (newStates.length > 0) {
    filterOptionSource.state.list = newStates.map(value => ({
      text: statusMap.value[value]?.text || value,
      value,
    }));
  }
}, { immediate: true });

// 监听tableData变化，自动更新统计信息
watch(() => tableData.value, async () => {
  // 防抖处理，避免频繁调用接口
  debouncedGetStatistics();
}, { deep: true });

watch(
  () => needInterval.value,
  async () => {
    if (!needInterval.value) {
      stop();
      await updataCurrentTaskInfo();
      // updataCurrentTaskInfo 会用父任务状态覆盖 subTasksStatus，
      // 若两接口数据不同步导致 needInterval 回弹，需重新启动轮询
      if (needInterval.value) {
        start();
      }
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
      name: t('taskDetail.search.status'),
      values: [{
        id: route.query.status,
        name: statusMap.value[route.query.status]?.text || route.query.status,
      }],
    });
  }
  await updataCurrentTaskInfo();
  await getOperateList();
  await Promise.all([getNetworkAreaList(), getNetworkUnitList()]);
  getStatistics();
  // 获取状态去重列表
  await getDistinctStates();
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
    border-color: #2caf5e;
    background: #cbf0da;
  }
}
.nc-terminated {
  &::before {
    border-color: #8E62D1;
    background: #e0d2f4;
  }
}
.nc-warning {
  &::before {
    border-color: #ff9c01;
    background: #fce5c0;
  }
}
.nc-unknown {
  &::before {
    border-color: #b2b5bd;
    background: #f0f1f5;
  }
}
.nc-incomplete {
  &::before {
    border-color: #90A4B2;
    background: #e2e7eb;
  }
}
.nc-success {
  &::before {
    border-color: #2DCC56;
    background: #cbf0da;
  }
}
.nc-failed {
  &::before {
    border-color: #EF5350;
    background: #f5cfcf;
  }
}
.nc-timeout {
  &::before {
    border-color: #ff9c01;
    background: #fce5c0;
  }
}
.dropdownCls {
  :deep(span) {
    display: inline-block;
    vertical-align: middle;
  }
}
</style>
