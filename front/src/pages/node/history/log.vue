<template>
  <page-header :title="title" :on-back="handleBackToHistoryDetail" class="border-b border-[#e4e7ef]"></page-header>
  <div class="flex w-full h-[calc(100%-52px)]">
    <div class="w-[280px] h-full flex flex-col">
      <div class="h-[72px] p-[20px]">
        <SearchSelect
          class="flex-1 bg-[#fff]"
          ref="searchSelect"
          :data="searchSelectData"
          v-model.trim="searchSelectValue"
          :unique-select="true"
          :max-height="240"
          :placeholder="isNode
            ? $t('platform.nodeMan.historySearchPlaceholder')
            : $t('platform.nodeMan.log.searchPlugin')"
          @update:model-value="handleSearchSelectChange"
          @paste.native="handleNativePaste"
        >
        </SearchSelect>
      </div>
      <bk-loading :title="$t('table.loading')" :loading="operateLoading" class="flex-1 h-[calc(100%-72px)] flex flex-col">
        <div class="flex-1 overflow-y-auto">
          <div
            v-for="operate in operateList" :key="operate.bk_host_id"
            class="cursor-pointer w-full px-[20px] h-[40px] leading-[40px] flex items-center"
            :class="{ 'bg-[#e1ecff] log-active-ip': isNode
              ? Number(route.params.hostId) === operate.bk_host_id
              : route.params.hostId === (`${operate.bk_host_id}_${operate.plugin_name}`) }"
            @click="handleChangeIp(operate.bk_host_id, operate.plugin_name)"
          >
            <span class="mr-[5px] leading-none">
              <i
                v-if="statusMap[operate.state]?.icon"
                :class="[`nodeman-icon nc-${
                  statusMap[operate.state].icon
                } status-icon`, 'align-middle', 'text-[8px]']"
              ></i>
              <Spinner
                v-else-if="operate.state === 'running'"
                width="12px"
                height="12px"
              ></Spinner>
              <i class="nodeman-icon nc-unknown status-icon align-middle text-[8px]" v-else></i>
            </span>
            <bk-overflow-title type="tips" class="text-[#63656e] w-[219px]">
              {{ operate.bk_host_inner_list }} <span v-if="operate.plugin_name">({{ operate.plugin_name }})</span>
            </bk-overflow-title>
          </div>
        </div>
        <!-- 左侧 IP 列表分页（后端分页） -->
        <div v-if="totalPages > 1" class="flex items-center justify-center gap-[12px] py-[6px] border-t border-[#dcdee5] text-[12px] text-[#63656e]">
          <Button
            size="small"
            theme="primary"
            text
            :disabled="pagination.current <= 1"
            @click="goIpPage(pagination.current - 1)"
          >
            <i class="nodeman-icon nc-arrow-left text-[14px]"></i>
          </Button>
          <span>{{ pagination.current }} / {{ totalPages }}</span>
          <Button
            size="small"
            theme="primary"
            text
            :disabled="pagination.current >= totalPages"
            @click="goIpPage(pagination.current + 1)"
          >
            <i class="nodeman-icon nc-arrow-right text-[14px]"></i>
          </Button>
        </div>
      </bk-loading>
    </div>
    <div class="bg-[#fff] px-[24px] pb-[20px] h-full flex-1 flex flex-col">
      <div class="h-[32px] mt-[20px]" v-if="currentOperate && ['failed', 'timeout', 'terminated'].includes(currentOperate.state)">
        <Dropdown
          theme="light"
          trigger="click"
          ext-cls="dropdownCls"
          :popover-options="{
            clickContentAutoHide: true,
          }">
          <Button
            class="w-[86px]">
            <span>{{ $t('taskDetail.table.retry') }}</span>
          </Button>
          <template #content>
            <Dropdown.DropdownMenu>
              <Dropdown.DropdownItem
                class="text-14px"
                v-for="item in reTryType"
                :key="item.id"
                v-bk-tooltips="{
                  content: item.tooltip,
                  placement: 'right',
                }"
                @click="handleRetry(currentOperate, item.id)">
                <Button
                  text
                  :disabled="currentOperate?.state === 'terminated' && item.id === 'PARTIAL'"
                >{{ item.name }}</Button>
              </Dropdown.DropdownItem>
            </Dropdown.DropdownMenu>
          </template>
        </Dropdown>
      </div>
      <bk-loading
        :title="$t('table.loading')"
        :loading="instanceLoading"
        class="flex-1 flex h-[calc(100%-82px)] mt-[20px]">
        <Table
          :data="tableData"
          :empty-text="$t('table.empty')"
          :min-width="300"
          :auto-resize="false"
          class="w-[40%] h-full mr-[10px] bg-[#f5f7fa]"
        >
          <TableColumn :title="$t('platform.nodeMan.log.step')" min-width="200" fixed="left">
            <template #default="{ row }">
              <Button
                text
                :class="['cursor-pointer', { 'text-[#c6c4cc]': row.state === 'pending' }]"
                @click="handleClickStep(row)"
              >
                <span v-if="row.stepKey !== 'extra_execution_logs'">{{ row.index }}. </span>
                {{ getDisplayName(row) }}
              </Button>
            </template>
          </TableColumn>
          <TableColumn field="costTime" :title="$t('platform.nodeMan.log.costTime')" min-width="60">
            <template #default="{ row }">
              <span :class="{ 'text-[#c6c4cc]': row.state === 'pending' }">
                {{ formatTimeToMS(row.costTime) }}
              </span>
            </template>
          </TableColumn>
          <TableColumn
            field="state"
            :title="$t('platform.nodeMan.log.executionStatus')"
            :min-width="isManual || isOffline ? 260 : 180">
            <template #default="{ row }">
              <div class="flex items-center gap-[5px]">
                <div class="flex items-center flex-1 min-w-0">
                  <i
                    v-if="statusMap[row.state]?.icon"
                    :class="`nodeman-icon nc-${
                      statusMap[row.state].icon
                    } status-icon`"
                  ></i>
                  <Spinner
                    v-else-if="row.state === 'running'"
                    width="12.25px"
                    height="12.25px"
                  />
                  <span class="nodeman-icon nc-unknown status-icon" v-else></span>
                  <div :class="['ml-[5px]', 'min-w-0']">
                    <div v-if="showOfflineGuideButton(row)">
                      {{ $t('platform.nodeMan.log.waitOfflineOperation') }}
                      <Button class="ml-[2px]" text theme="primary" @click="handleOfflineGuide">
                        {{ $t('platform.nodeMan.log.offlineGuide') }}
                      </Button>
                    </div>
                    <!-- eslint-disable-next-line max-len -->
                    <div v-else-if="isManual && last_oper_inst_step_key === row.stepKey && currentOperate?.state === 'running'">
                      {{ $t('platform.nodeMan.log.waitManualOperation') }}
                      <Button class="ml-[2px]" text theme="primary" @click="handleOperateGuide">
                        {{ $t('platform.nodeMan.log.operationGuide') }}
                      </Button>
                    </div>
                    <span v-else :class="[{ 'text-[#c4c6cc]': row.state === 'pending' }]">
                      {{ statusMap[row.state]?.text }}
                    </span>
                  </div>
                  <Button
                    v-if="row.sub_workflow_refs?.length === 1"
                    text
                    theme="primary"
                    class="ml-[6px] !px-0 min-w-0 leading-none"
                    data-test="sub-workflow-entry"
                    v-bk-tooltips="{
                      content: $t('platform.nodeMan.log.viewSubWorkflow'),
                      placement: 'top',
                    }"
                    @click="openSubWorkflowDetail(row.sub_workflow_refs[0])"
                  >
                    <img :src="JumpLink" alt="" class="w-[11px] h-[11px]" />
                  </Button>
                  <Dropdown
                    v-else-if="row.sub_workflow_refs && row.sub_workflow_refs.length > 1"
                    :popover-options="{
                      clickContentAutoHide: true,
                      boundary: 'body',
                      trigger: 'click',
                    }"
                  >
                    <Button
                      text
                      theme="primary"
                      class="ml-[6px] !px-0 min-w-0 leading-none"
                      data-test="sub-workflow-entry"
                      v-bk-tooltips="{
                        content: $t('platform.nodeMan.log.viewSubWorkflow'),
                        placement: 'top',
                      }"
                    >
                      <img :src="JumpLink" alt="" class="w-[11px] h-[11px]" />
                    </Button>
                    <template #content>
                      <ul class="py-[4px]">
                        <li
                          v-for="(subRef, idx) in row.sub_workflow_refs"
                          :key="idx"
                          class="px-[16px] py-[6px] cursor-pointer hover:bg-[#f0f1f5] text-[12px] whitespace-nowrap"
                          @click="openSubWorkflowDetail(subRef)"
                        >
                          {{ $t('platform.nodeMan.log.subTaskN', { n: Number(idx) + 1 }) }}
                        </li>
                      </ul>
                    </template>
                  </Dropdown>
                </div>
                <Button
                  v-if="row.state === 'running'"
                  text
                  theme="primary"
                  @click="handleTerminate">
                  {{ $t('platform.nodeMan.log.terminate') }}
                </Button>
              </div>
            </template>
          </TableColumn>
          <TableColumn fixed="right" min-width="32">
            <template #default="{ row }">
              <div class="flex items-center justify-end">
                <right-shape fill="#C4C6CC" v-if="row.stepKey === activeStepKey" />
              </div>
            </template>
          </TableColumn>
        </Table>
        <div class="flex-1 flex flex-col text-[12px]">
          <div
            class="sticky top-0 z-10 h-[50px]
              flex flex-shrink-0 justify-between items-center px-[16px] bg-[#202024] text-[#C4C6CC]"
          >
            <div class="flex items-center">
              <span>{{ $t('platform.nodeMan.log.executionLog') }}</span>
              <Dropdown trigger="click" :distance="4" class="ml-[20px]">
                <div
                  class="flex items-center rounded-[2px] border border-[#3a3a3e] px-[16px] h-[32px] cursor-pointer text-[#E1ECFF] text-[14px] select-none"
                  style="background: #2c2c30;"
                >
                  <span>{{ $t('platform.nodeMan.log.level.label') }} ({{ selectedLogLevels.length }}/{{ logLevelOptions.length }})</span>
                  <i class="nodeman-icon nc-arrow-down ml-[6px] text-[10px]"></i>
                </div>
                <template #content>
                  <div
                    class="rounded-[2px] border border-[#3a3a3e] py-[6px] px-[14px] text-[14px] log-level-dropdown"
                    style="background: #2c2c30; min-width: 120px;"
                  >
                    <Checkbox.Group v-model="selectedLogLevels">
                      <div
                        v-for="opt in logLevelOptions"
                        :key="opt.value"
                        class="py-[5px] mr-[20px]"
                      >
                        <Checkbox
                          :label="opt.value"
                          :class="{
                            'warn-level': opt.value === 'WARN',
                            'error-level': opt.value === 'ERROR',
                            'debug-level': opt.value === 'DEBUG',
                          }"
                        >{{ opt.label }}</Checkbox>
                      </div>
                    </Checkbox.Group>
                  </div>
                </template>
              </Dropdown>
            </div>
            <div class="flex">
              <Dropdown
                :popover-options="{
                  clickContentAutoHide: true,
                  boundary: 'body',
                  trigger: 'click',
                }"
              >
                <Button text class="mr-[16px] flex items-center">
                  <span class="text-[#C4C6CC] mr-[6px]">{{ curOperInstVal }}</span>
                  <angle-up-fill class="text-[16px] text-[#C4C6CC]" />
                </Button>
                <template #content>
                  <ul class="w-[80px] py-[4px] max-h-[300px] overflow-auto">
                    <li
                      v-for="item in operInstList"
                      :key="item.name"
                      @click="handleClick(item)"
                      :class="[
                        'w-full h-[32px] flex items-center px-[12px] cursor-pointer',
                        {
                          'bg-[#E1ECFF] text-[#3A84FF]':
                            curOperInstVal === item.name,
                        },
                      ]"
                    >
                      {{ item.name }}
                    </li>
                  </ul>
                </template>
              </Dropdown>
            </div>
          </div>
          <div class="bg-[#313238] flex-1 flex  overflow-y-auto log-content">
            <!-- 当前筛选下无日志时展示空状态 -->
            <div v-if="!visibleLogs || visibleLogs.length === 0" class="flex-1 flex items-center justify-center text-[#979BA5] text-[14px]">
              {{ $t('platform.nodeMan.log.noLogForFilter') }}
            </div>
            <!-- 筛选后的日志列表 -->
            <div class="flex-1 text-[#a8acb8]" v-if="visibleLogs && visibleLogs.length > 0">
              <div v-for="logItem in visibleLogs" :key="logItem.key" :class="logItem.key">
                <div
                  v-for="(item, index) in logItem.logs"
                  :key="index"
                  class="mx-[30px] my-[8px]"
                  :class="{
                    'bg-[#422321] flex !mx-0': isExecutionLogError(item),
                    'bg-[#3d3220] flex !mx-0': isExecutionLogWarn(item),
                  }"
                >
                  <div class="flex justify-center items-baseline w-[26px] pt-[6px]">
                    <close
                      v-if="isExecutionLogError(item)"
                      :fill="'#993D3D'"
                      width="12.25px"
                      height="12.25px"
                    />
                    <exclamation-circle-shape
                      v-else-if="isExecutionLogWarn(item)"
                      :fill="'#FF9C01'"
                      width="12.25px"
                      height="12.25px"
                    />
                  </div>
                  <div class="flex-1">
                    <span class="mr-[8px]"> [{{ timeFormatter(item.time) }}
                      <span class="ml-[5px]">
                        {{ item.level }}
                      </span>
                      ]
                    </span>
                    <span class="line-height-[24px]">{{ getLogText(item) }}</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </bk-loading>
    </div>
  </div>
  <guide v-model:is-show="isGuideShow" :data="guideData" />
</template>
<script setup lang="ts">
import { Button, Checkbox, Dropdown, InfoBox, Message, SearchSelect } from 'bkui-vue';
import {
  AngleUpFill,
  Close,
  ExclamationCircleShape,
  RightShape,
  Spinner,
} from 'bkui-vue/lib/icon';
import { debounce } from 'lodash';
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';

import { Table, TableColumn } from '@blueking/table';

import JumpLink from '../../../../public/images/jump-link.svg';

import guide from './guide.vue';
import { isOfflineGuideStep, STEP_KEY_WAIT_OFFLINE_MANUAL_INSTALL } from './offline-package';

import { NodeWorkflowService } from '@/api/modules/node_workflow';
import { PluginWorkflowService } from '@/api/modules/plugin_workflow';
import { formatTimeByTimezone } from '@/common/util';
import useInterval from '@/composables/use-interval';
import { useMainStore } from '@/stores/main';
import { useNodeManageStore } from '@/stores/node-manage';

const { t } = useI18n();
const route = useRoute();
const router = useRouter();

const mainStore = useMainStore();
const nodeManageStore = useNodeManageStore();

// 语言判断
const isZh = computed(() => mainStore.curLanguage === 'zh-CN');

// 获取 action 显示名称
const getDisplayName = (row: any) => (isZh.value
  ? (row.display_name_zh || row.stepKey)
  : (row.display_name_en || row.stepKey));

// 获取日志文本
const getLogText = (item: any) => (isZh.value
  ? (item.text_zh || '')
  : (item.text_en || ''));

const isExecutionLogError = (item: any) => item.level === 'ERROR' || getLogText(item).includes('ERROR');
const isExecutionLogWarn = (item: any) => !isExecutionLogError(item) && (item.level === 'WARN' || getLogText(item).includes('WARN'));

const IPV4_REG = /^((25[0-5]|2[0-4]\d|[01]?\d\d?)\.){3}(25[0-5]|2[0-4]\d|[01]?\d\d?)$/;
const IPV6_REG = /^(?:[A-F0-9]{1,4}:){7}[A-F0-9]{1,4}$/i;

const { start, stop } = useInterval(getLog, 1000); // 轮询

// 统一服务调用器
const serviceCaller = {
  // 根据路由参数获取当前服务类型
  getCurrentServiceType: () => (route.query.active === 'plugin' ? 'plugin' : 'node'),

  // 服务方法映射
  serviceMethods: {
    node: {
      retry: NodeWorkflowService.NodeWorkflowOperationRetry,
      terminate: NodeWorkflowService.NodeWorkflowOperationTerminate,
      operationList: NodeWorkflowService.NodeWorkflowOperationList,
      operationInstanceList: NodeWorkflowService.NodeWorkflowOperationInstanceList,
      operationInstanceLogGet: NodeWorkflowService.NodeWorkflowOperationInstanceLogGet,
      operationDistinct: NodeWorkflowService.NodeWorkflowOperationDistinct,
    },
    plugin: {
      retry: PluginWorkflowService.PluginWorkflowOperationRetry,
      terminate: PluginWorkflowService.PluginWorkflowOperationTerminate,
      operationList: PluginWorkflowService.PluginWorkflowOperationList,
      operationInstanceList: PluginWorkflowService.PluginWorkflowOperationInstanceList,
      operationInstanceLogGet: PluginWorkflowService.PluginWorkflowOperationInstanceLogGet,
      operationDistinct: PluginWorkflowService.PluginWorkflowOperationDistinct,
    },
  },

  // 统一调用方法
  async call(method: 'retry' | 'terminate' | 'operationList' | 'operationInstanceList' | 'operationInstanceLogGet' | 'operationDistinct', params: any) {
    const serviceType = this.getCurrentServiceType();
    const serviceMethod = this.serviceMethods[serviceType][method];
    return await serviceMethod(params);
  },
};
const activeKey = ref('');
const isNode = computed(() => route.query.active !== 'plugin');
const operateList = ref<any[]>([]); // 子任务列表（后端分页，仅当前页数据）
// 左侧 IP 列表分页（后端分页）
const pagination = reactive({ count: 0, limit: 20, current: 1 });
const totalPages = computed(() => Math.max(1, Math.ceil(pagination.count / pagination.limit)));
const goIpPage = (page: number) => {
  if (page < 1 || page > totalPages.value) return;
  pagination.current = page;
  getOperateList();
};

// 当前选中的子任务（从列表当前页查找；跳转进入时通过路由带入的 IP 搜索条件保证目标在结果中）
const currentOperate = ref<any>(null);
// 定位当前选中子任务：hostId 在当前页内（IP 搜索过滤后结果极少，目标必在列表）
const locateCurrentOperate = () => {
  const hostIdParam = String(route.params.hostId ?? '');
  if (!hostIdParam) {
    currentOperate.value = null;
    return;
  }
  const matches = (item: any): boolean => (isNode.value
    ? Number(item.bk_host_id) === Number(hostIdParam)
    : `${item.bk_host_id}_${item.plugin_name}` === hostIdParam);
  currentOperate.value = operateList.value.find(matches) ?? null;
};
// 轻量刷新选中子任务的 state（launched/init 期间防抖调用，只请求该 operation 的实例）
const refreshCurrentOperateState = async () => {
  const operationId = currentOperate.value?.operation_id;
  if (!operationId) return;
  const params = isNode.value
    ? { operation_id: operationId }
    : { operation_id: [operationId] };
  const res = await serviceCaller.call('operationInstanceList', params).catch(() => ({ oper_inst_data: [] }));
  const instances = res?.oper_inst_data ?? [];
  const latest = instances[instances.length - 1];
  if (latest?.life_cycle?.state) {
    currentOperate.value = { ...currentOperate.value, state: latest.life_cycle.state };
  }
};
const isManual = computed(() => !!currentOperate.value?.latest_action_inst_brief_data?.tags?.includes('need_manual_exec_install_script'));
const isOffline = computed(() => isOfflineGuideStep(currentOperate.value?.latest_action_inst_brief_data?.tags ?? []));

const openSubWorkflowDetail = (ref: { workflow_id: string; workflow_domain: 'node' | 'plugin' }) => {
  const loc = router.resolve({
    name: 'taskDetail',
    params: { taskId: ref.workflow_id },
    query: ref.workflow_domain === 'plugin' ? { active: 'plugin' } : {},
  });
  window.open(loc.href, '_blank', 'noopener,noreferrer');
};

// Pin to wait_offline_manual_install row; last_oper_inst_step_key moves to the next step after submit.
const showOfflineGuideButton = (row: { stepKey: string; state: string }): boolean => (
  currentOperate.value?.state === 'running'
    && row.stepKey === STEP_KEY_WAIT_OFFLINE_MANUAL_INSTALL
    && row.state === 'running'
);
const title = computed(() => t('platform.nodeMan.log.executionLogOf', { inner: currentOperate.value?.bk_host_inner_list, type: typeMap.value[nodeManageStore.taskHistoryTableRowData.type] }));
const curOperInstId = ref('');
const curOperInstVal = ref('latest');
const curSortNames = ref<string[]>([]);
const logs = ref<{ text: string; text_zh?: string; text_en?: string; level: string; time: string }[]>([]);
const allLogs = ref<any[]>([]);
const selectedLogLevels = ref<string[]>(['INFO', 'WARN', 'ERROR']);
const logLevelOptions = computed(() => [
  { value: 'INFO', label: t('platform.nodeMan.log.level.info') },
  { value: 'WARN', label: t('platform.nodeMan.log.level.warn') },
  { value: 'ERROR', label: t('platform.nodeMan.log.level.error') },
  { value: 'DEBUG', label: t('platform.nodeMan.log.level.debug') },
]);
const visibleLogs = computed(() => allLogs.value
  .map(logItem => ({
    ...logItem,
    logs: logItem.logs.filter((item: any) => selectedLogLevels.value.includes(String(item.level || '').toUpperCase())),
  }))
  .filter(logItem => logItem.logs.length > 0));
const operInstList = ref<{ id: string; name: string; sort_names: string[] }[]>([]);
const logData = ref<{
  total: number;
  oper_inst_logs: Record<string, ActionMessage>;
}>({
  total: 0,
  oper_inst_logs: {},
});
const tableData = ref<any[]>([]);
const last_oper_inst_step_key = computed(() => {
  for (let i = 0; i < tableData.value.length; i++) {
    if (tableData.value[i].state === 'pending') {
      return tableData.value[i - 1].stepKey;
    }
  }
  return '';
});
const statusMap = computed(() => ({
  running: {
    text: '执行中',
  },
  failed: {
    text: '执行失败',
    icon: 'failed',
  },
  timeout: {
    text: '超时',
    icon: 'timeout',
  },
  success: {
    text: '执行成功',
    icon: 'success',
  },
  skipped: {
    text: '跳过',
    icon: 'unknown',
  },
  pending: {
    text: '等待执行',
    icon: 'unknown',
  },
  terminated: {
    text: '终止',
    icon: 'terminated',
  },
  launched: {
    text: '等待执行',
    icon: 'incomplete',
  },
  init: {
    text: '初始化',
    icon: 'incomplete',
  },
}));
const typeMap = computed(() => ({
  install_agent: t('platform.nodeMan.taskHistory.taskType.install_agent'),
  install_plugin: t('platform.nodeMan.taskHistory.taskType.install_plugin'),
  debug_plugin: t('platform.nodeMan.taskHistory.taskType.debug_plugin'),
  ensure_plugin_v2: t('platform.nodeMan.taskHistory.taskType.ensure_plugin_v2'),
  stop_plugin_v2: t('platform.nodeMan.taskHistory.taskType.stop_plugin_v2'),
  upgrade_agent: t('platform.nodeMan.taskHistory.taskType.upgrade_agent'),
  upgrade_plugin: t('platform.nodeMan.taskHistory.taskType.upgrade_plugin'),
  uninstall_plugin_v2: t('platform.nodeMan.taskHistory.taskType.uninstall_plugin_v2'),
  migrate_plugin_v2: t('platform.nodeMan.taskHistory.taskType.migrate_plugin_v2'),
  reconfig_agent: t('platform.nodeMan.taskHistory.taskType.reconfig_agent'),
  restart_agent: t('platform.nodeMan.taskHistory.taskType.restart_agent'),
  uninstall_agent: t('platform.nodeMan.taskHistory.taskType.uninstall_agent'),
  install_proxy: t('platform.nodeMan.taskHistory.taskType.install_proxy'),
  upgrade_proxy: t('platform.nodeMan.taskHistory.taskType.upgrade_proxy'),
  reconfig_proxy: t('platform.nodeMan.taskHistory.taskType.reconfig_proxy'),
  restart_proxy: t('platform.nodeMan.taskHistory.taskType.restart_proxy'),
  uninstall_proxy: t('platform.nodeMan.taskHistory.taskType.uninstall_proxy'),
  assign_proxy_unit: t('platform.nodeMan.taskHistory.taskType.assign_proxy_unit'),
  uninstall_plugin: t('platform.nodeMan.taskHistory.taskType.uninstall_plugin'),
  reconfig_plugin: t('platform.nodeMan.taskHistory.taskType.reconfig_plugin'),
  apply_plugin_subconfig: t('platform.nodeMan.taskHistory.taskType.apply_plugin_subconfig'),
  start_plugin: t('platform.nodeMan.taskHistory.taskType.start_plugin'),
  restart_plugin: t('platform.nodeMan.taskHistory.taskType.restart_plugin'),
  stop_plugin: t('platform.nodeMan.taskHistory.taskType.stop_plugin'),
}));

const timeFormatter = (
  val: number | string | undefined,
  format = 'YYYY-MM-DD HH:mm:ss',
) => (val ? formatTimeByTimezone(val, format) : '--');

const formatTimeToMS = (duration: number) => {
  const seconds = Math.floor(duration / 1000);
  return `${seconds}s`;
};

// 获取状态去重列表
const distinctStates = ref<string[]>([]);
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
    distinctStates.value = Array.from(new Set(operateList.value
      .map((item: any) => item.state)
      .filter((item: any) => item !== null && item !== undefined && item !== '')));
  }
};
// 搜索
const searchSelectValue = ref<{ id: string; name: string; values: any[] }[]>([]);
// 搜索条件 → 后端 exact_include_conditions（ip 按 IPv4/IPv6 分类）
const buildExactConditions = (): Record<string, any> | undefined => {
  const cond: Record<string, any> = {};
  searchSelectValue.value.forEach((item: any) => {
    if (item.id === 'ip' && item.values?.length) {
      const ipv4List: string[] = [];
      const ipv6List: string[] = [];
      item.values.forEach((v: any) => {
        if (IPV4_REG.test(v.id)) ipv4List.push(v.id);
        else if (IPV6_REG.test(v.id)) ipv6List.push(v.id);
      });
      if (ipv4List.length) cond.bk_host_innerip = ipv4List;
      if (ipv6List.length) cond.bk_host_innerip_v6 = ipv6List;
      return;
    }
    if (item.id === 'state' && item.values?.length) {
      cond.state = item.values.map((v: any) => v.id);
      return;
    }
    if (item.id === 'plugin_name' && item.values?.length) {
      cond.plugin_name = item.values.map((v: any) => v.id);
    }
  });
  return Object.keys(cond).length > 0 ? cond : undefined;
};
// 从任务详情跳转进入时，把路由带入的目标主机条件初始化到搜索栏，
// 复用「搜索→后端过滤」链路直接命中当前机器；用户清空搜索条件即恢复全量
const initSearchFromRoute = () => {
  if (isNode.value) {
    // node：协议支持 bk_host_innerip / bk_host_innerip_v6 过滤，搜索栏带 IP
    const ipStr = String(route.query.ip ?? '');
    const ipv6Str = String(route.query.ipv6 ?? '');
    const ips = [...parseMultiDelimiterInput(ipStr), ...parseMultiDelimiterInput(ipv6Str)]
      .filter(ip => !!ip);
    if (ips.length === 0) return;
    searchSelectValue.value = [{
      id: 'ip',
      name: 'IP',
      values: Array.from(new Set(ips)).map(ip => ({ id: ip, name: ip })),
    }];
    return;
  }
  // plugin：协议无 IP 过滤条件，改带 plugin_name（hostId 格式为 `${bk_host_id}_${plugin_name}`）。
  // 列表会显示该任务下同插件的所有主机，locateCurrentOperate 按 hostId 字符串精确匹配定位目标
  const hostIdParam = String(route.params.hostId ?? '');
  const sepIndex = hostIdParam.indexOf('_');
  if (sepIndex < 0) return;
  const pluginName = hostIdParam.slice(sepIndex + 1);
  searchSelectValue.value = [{
    id: 'plugin_name',
    name: '插件名',
    values: [{ id: pluginName, name: pluginName }],
  }];
};
// 搜索条件变化：重置到第一页并触发后端查询
watch(searchSelectValue, () => {
  pagination.current = 1;
  getOperateList();
}, { deep: true });
const searchSelectData = computed(() => [
  { id: 'ip', name: 'IP', multiple: true },
  ...(isNode.value ? [] : [{ id: 'plugin_name', name: '插件名' }]),
  {
    id: 'state',
    name: '执行状态',
    // 状态选项来自 operationDistinct 接口（全量去重，与后端分页解耦）
    children: distinctStates.value.map((value: string) => ({
      id: value,
      name: statusMap.value[value]?.text || value,
    })),
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
const detectInputType = (text: string): { type: 'ip' | null; value: string } => {
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

  if (['ip'].includes(lastItem.id) && lastItem.values.length > 0) {
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

  if (firstDetection.type === 'ip') {
    targetId = 'ip';
    targetName = 'IP';
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

const handleSearchSelectChange = async (data: { id: string; name: string; values: { id: string; name: string }[] }[]) => {
  handleInputPaste(data);
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
const getCurrentOperationId = () => currentOperate.value?.operation_id;

const handleOperateGuide = () => {
  const operationId = getCurrentOperationId();
  if (!operationId) return;

  isGuideShow.value = true;
  guideData.value = {
    workflow_id: route.params.taskId as string,
    operation_id: operationId,
    bk_host_innerip: currentOperate.value?.bk_host_inner_list,
    is_offline: false,
  };
};
const handleOfflineGuide = () => {
  const operationId = getCurrentOperationId();
  if (!operationId) return;

  isGuideShow.value = true;
  guideData.value = {
    workflow_id: route.params.taskId as string,
    operation_id: operationId,
    bk_host_innerip: currentOperate.value?.bk_host_inner_list,
    bk_networkarea_id: currentOperate.value?.bk_networkarea_id,
    is_offline: true,
  };
};

// 重试
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

const handleRetry = async (row: { operation_id?: string } | undefined, type: string) => {
  if (!route.params.taskId) return;

  const operationId = row?.operation_id;
  if (!operationId) return;

  const res = await serviceCaller.call('retry', {
    workflow_id: route.params.taskId,
    operation_ids: [operationId],
    retry_mod: type,
  }).catch(() => false);
  if (res !== false) {
    setTimeout(async () => { // 重试后数据有0.5s到1s的延迟更新
      isInterval.value = true;
      await getOperateList();
      await getInstance();
      start();
    }, 1000);
    mainStore.updateLogRetry(true);
  }
};

// 终止
const handleTerminate = () => {
  if (!route.params.taskId) return;

  const operationId = getCurrentOperationId();
  if (!operationId) return;

  InfoBox({
    title: t('platform.nodeMan.log.terminateConfirmTitle'),
    subTitle: t('platform.nodeMan.log.terminateConfirmSubTitle'),
    onConfirm: async () => {
      const res = await serviceCaller.call('terminate', {
        workflow_id: route.params.taskId,
        operation_ids: [operationId],
      }).catch(() => {
        Message({ theme: 'error', message: t('platform.nodeMan.log.terminateFailedMsg') });
        return false;
      });
      if (res !== false) {
        await getOperateList();
        await getInstance();
        start();
        mainStore.updateLogTerminate(true);
      }
    },
  });
};

const scrollToFirstErrorByClassNames = (classNames: string | undefined) => {
  // 1. 校验参数：类名不存在直接返回，避免无效查询
  if (!classNames || typeof classNames !== 'string') return;

  // 2. 查询首个匹配类名的元素（classNames直接传入选择器，如".error-log"）
  const targetEl = document.querySelector(classNames);
  if (!targetEl) return; // 3. 元素不存在时终止，避免报错

  // 4. 平滑滚动到元素顶部
  targetEl.scrollIntoView({
    behavior: 'smooth',
    block: 'start', // 对齐元素顶部
  });
  // 添加高亮动画效果
  setTimeout(() => {
    // 添加高亮样式类
    targetEl.classList.add('scroll-highlight-animation');

    // 3秒后移除高亮样式
    setTimeout(() => {
      targetEl.classList.remove('scroll-highlight-animation');
    }, 3000);
  }, 500); // 滚动完成后开始高亮动画
};
const activeStepKey = ref();
const handleClickStep = (row: any) => {
  activeStepKey.value = row.stepKey;
  scrollToFirstErrorByClassNames(`.${row.stepKey}`);
};

// 搜索Ip
const handleChangeIp = async (hostId: number, plugin_name = '') => {
  router.replace({
    name: 'log',
    params: {
      hostId: isNode.value ? hostId : `${hostId}_${plugin_name}`,
      taskId: route.params.taskId,
    },
    query: {
      active: route.query?.active,
    },
  });
};


// 精确返回到历史详情页面
const handleBackToHistoryDetail = () => {
  router.push({
    name: 'taskDetail',
    params: {
      taskId: route.params.taskId,
    },
    query: {
      active: route.query?.active,
    },
  });
};

const handleClick = async (item: {
  name: string;
  id: string;
  sort_names: string[];
}) => {
  curOperInstId.value = item.id;
  curOperInstVal.value = item.name;
  curSortNames.value = item.sort_names;
  await getLog();
  if (isInterval.value) {
    start();
  } else {
    stop();
  }
};
function getOrdinalSuffix(number: number) {
  // 处理 11, 12, 13 的特殊情况
  if (number % 100 >= 11 && number % 100 <= 13) {
    return `${number}th`;
  }

  // 根据最后一位数字决定后缀
  const lastDigit = number % 10;
  let suffix;
  switch (lastDigit) {
    case 1:
      suffix = 'st';
      break;
    case 2:
      suffix = 'nd';
      break;
    case 3:
      suffix = 'rd';
      break;
    default:
      suffix = 'th';
      break;
  }
  return `${number}${suffix}`;
}

// 获取任务列表
const operateLoading = ref(false);
const getOperateList = async () => {
  // 没有taskId，不请求operationList
  if (!route.params.taskId) return;

  operateLoading.value = true;
  // 后端分页：单页请求，搜索条件透传给后端过滤
  const params: Record<string, any> = {
    page: { limit: pagination.limit, offset: (pagination.current - 1) * pagination.limit },
    workflow_id: route.params.taskId,
  };
  const exact = buildExactConditions();
  if (exact) params.exact_include_conditions = exact;
  const res = await serviceCaller.call('operationList', params).catch(() => ({
    operations: [],
    total: 0,
  }));
  operateLoading.value = false;
  pagination.count = res?.total ?? 0;
  operateList.value = (res?.operations ?? []).map((item) => {
    const briefData = item.latest_oper_inst_brief_data;
    const deployInfo = route.query.active !== 'plugin' ? item.node_deployment_info : item.plugin_deployment_info;
    return {
      ...deployInfo,
      ...(briefData ? briefData.life_cycle : {}),
      latest_action_inst_brief_data: briefData ? briefData.latest_action_inst_brief_data : {},
      bk_host_inner_list: deployInfo?.bk_host_inner_list?.join(',') || deployInfo?.bk_host_innerip_list?.join(','),
      bk_host_innerip_v6_list: deployInfo?.bk_host_innerip_v6_list?.join(',') || deployInfo?.bk_host_innerip_v6_list?.join(','),
      operation_id: item.operation_id,
    };
  });
};

const instanceLoading = ref(false);
// keepLoading=true 时不在函数内关闭 loading，由调用方在后续请求（如 getLog）完成后统一关闭
const getInstance = async (keepLoading = false) => {
  instanceLoading.value = true;

  const operationId = getCurrentOperationId();
  if (!operationId) {
    if (!keepLoading) instanceLoading.value = false;
    isInterval.value = false;
    stop();
    hasErrorOrTimeout.value = false;
    activeKey.value = '';
    curOperInstId.value = '';
    curOperInstVal.value = 'latest';
    curSortNames.value = [];
    operInstList.value = [];
    tableData.value = [];
    allLogs.value = [];
    logs.value = [];
    logData.value = { total: 0, oper_inst_logs: {} };
    return;
  }

  const params = route.query.active !== 'plugin'
    ? {
      operation_id: operationId,
    }
    : {
      operation_id: [operationId],
    };
  const res = await serviceCaller.call('operationInstanceList', params).catch(() => ({
    oper_inst_data: [],
  }));

  const total = res.oper_inst_data.length;
  curOperInstId.value = total > 0 ? res.oper_inst_data[total - 1].oper_inst_id : '';
  curOperInstVal.value = 'latest';
  curSortNames.value = total > 0 ? res.oper_inst_data[total - 1].action_names : [];
  operInstList.value = [];
  for (let i = 1; i <= total; i++) {
    operInstList.value.push({
      name: i < total ? getOrdinalSuffix(i) : 'latest',
      id: res.oper_inst_data[i - 1].oper_inst_id,
      sort_names: res.oper_inst_data[i - 1].action_names,
    });
  }
  if (!keepLoading) instanceLoading.value = false;
};
const hasErrorOrTimeout = ref(false);
const isInterval = ref(false);
const needRefreshOperateList = ref(false);
// 记录上一次轮询的 hasRunning 状态，仅在 running→非running 边界触发一次
// operationList 刷新，避免任务已失败/超时/终止后仍每轮重复请求。
let prevHasRunning = false;
async function getLog() {
  if (!curOperInstId.value) return;
  hasErrorOrTimeout.value = false;
  needRefreshOperateList.value = false;
  logData.value.oper_inst_logs = {};
  const res: any = await serviceCaller.call('operationInstanceLogGet', {
    oper_inst_id: curOperInstId.value,
  }).catch(() => ({
    total: 0,
    oper_inst_logs: {},
    extra_execution_logs: { logs: [] },
  }));

  const operInstLogs = res.oper_inst_logs || {};
  const list: any[] = [];
  curSortNames.value.forEach((key: string, index: number) => {
    const actionData = operInstLogs[key];
    if (!actionData) return; // 跳过不存在的 key，防止报错

    logData.value.oper_inst_logs[key] = actionData;
    const { start_time, end_time, state } = actionData.life_cycle || {};
    let costTime = 0;
    if (start_time > 0) {
      if (end_time > 0) {
        costTime = end_time - start_time;
      } else if (state !== 'pending') {
        costTime = new Date().getTime() - start_time;
      }
    }
    list.push({
      index: index + 1,
      stepKey: key,
      display_name_zh: actionData.display_name_zh || key,
      display_name_en: actionData.display_name_en || key,
      costTime,
      state: state || '未知',
      sub_workflow_refs: actionData.sub_workflow_refs || [],
    });
  });
  logData.value.total = Object.keys(operInstLogs).length;

  // 额外执行日志和步骤日志统一保留，日志等级筛选只负责展示层过滤
  tableData.value = list;

  let currentKey;
  let hasRunning = false;
  for (const [key, entry] of Object.entries(logData.value.oper_inst_logs)) {
    const lifeState = (entry as any).life_cycle?.state;
    if (['failed', 'timeout', 'terminated'].includes(lifeState)) {
      currentKey = key;
      hasErrorOrTimeout.value = true;
    } else if (lifeState === 'running') {
      currentKey = key;
      hasRunning = true;
    } else {
      currentKey = key;
    }
  }
  // 当步骤状态从 running 变为非 running（含 failed/timeout/terminated/success）时，
  // 刷新操作列表获取最新整体状态，确保左侧 IP 列表和重试按钮及时更新。
  // 通过 prevHasRunning 记录上一次状态，避免任务已失败/超时/终止后仍每轮重复请求
  // operationList 接口。
  if (hasRunning) {
    needRefreshOperateList.value = false;
  } else if (prevHasRunning) {
    needRefreshOperateList.value = true;
  }
  prevHasRunning = hasRunning;
  if (needRefreshOperateList.value) {
    await getOperateList();
    needRefreshOperateList.value = false;
  }
  const operateRunning = ['running', 'launched', 'init'].includes(currentOperate.value?.state);
  isInterval.value = hasRunning || operateRunning;

  // 构建 allLogs，等级过滤由 visibleLogs 计算属性负责
  const extraLogs = res.extra_execution_logs?.logs?.length
    ? [{ key: 'extra_execution_logs', logs: res.extra_execution_logs.logs }]
    : [];
  allLogs.value = [
    ...extraLogs,
    ...curSortNames.value
      .filter(key => operInstLogs[key]?.message)
      .map(key => ({
        key,
        logs: operInstLogs[key].message.logs,
      })),
  ];

  if (currentKey && operInstLogs[currentKey]?.message) {
    logs.value = { ...operInstLogs[currentKey].message.logs };
    activeKey.value = currentKey;
  }
}

watch(() => isInterval.value, async (val: boolean) => {
  if (!val) {
    stop();
    await getOperateList();
    await getInstance();
    // emit('stop');
  }
});
watch(() => route.path, async () => {
  locateCurrentOperate();
  await getInstance(true);
  await getLog();
  // getInstance + getLog 全部完成后关闭 loading，避免切换 IP 时 loading 关闭后数据才刷新
  instanceLoading.value = false;
  // 路由变化后，将左侧 IP 列表滚动到当前选中的 IP 项（置顶）
  scrollActiveIpIntoView();
  if (isInterval.value) {
    start();
  } else {
    stop();
  }
});

// 将左侧选中的 IP 滚动到列表顶部（用于从任务详情跳转后定位）
const scrollActiveIpIntoView = async () => {
  await nextTick();
  const el = document.querySelector('.log-active-ip') as HTMLElement | null;
  if (el && typeof el.scrollIntoView === 'function') {
    el.scrollIntoView({ behavior: 'auto', block: 'start' });
  }
};

// launched/init 期间只轮询刷新"当前选中那一台"的 state（单请求），不再整表刷新
const debouncedRefreshCurrentOperateState = debounce(() => {
  refreshCurrentOperateState();
}, 1000);
watch(() => currentOperate.value?.state, async (state) => {
  if (['launched', 'init'].includes(state)) {
    await debouncedRefreshCurrentOperateState();
  }
});

onBeforeUnmount(() => {
  stop();
});
onMounted(async () => {
  // 跳转进入：把路由带入的 IP 初始化到搜索栏（触发 watch 拉取，此处再拉一次兜底时序）
  initSearchFromRoute();
  await getOperateList();
  locateCurrentOperate();
  // getDistinctStates 成功路径不依赖 getInstance/getLog，与后续链并行
  await Promise.all([
    (async () => { await getInstance(true); await getLog(); })(),
    getDistinctStates(),
  ]);
  instanceLoading.value = false;
  if (isInterval.value) {
    start();
  } else {
    stop();
  }
  // 首屏加载完成后，将左侧选中的 IP 置顶（从任务详情跳转进入）
  scrollActiveIpIntoView();
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
/* 滚动高亮动画样式 */
.scroll-highlight-animation {
  animation: highlightPulse 1.5s ease-in-out 3;
  border-radius: 4px;
}

@keyframes highlightPulse {
  0% {
    box-shadow: 0 0 0 0 rgba(58, 132, 255, 0.7);
    background-color: rgba(58, 132, 255, 0.1);
  }
  50% {
    box-shadow: 0 0 0 8px rgba(58, 132, 255, 0.3);
    background-color: rgba(58, 132, 255, 0.2);
  }
  100% {
    box-shadow: 0 0 0 0 rgba(58, 132, 255, 0);
    background-color: rgba(58, 132, 255, 0.05);
  }
}
.log-content {
  &::-webkit-scrollbar {
    width: 14px;
    height: 16px;
  }
  &::-webkit-scrollbar-thumb {
    border-radius: 0;
    border: 1px solid #63656e;
    background: #3b3c42;
    box-shadow: none;
  }
}

/* 日志等级下拉：与日志中实际颜色一致 */
.log-level-dropdown :deep(.bk-checkbox-label) {
  color: #fff; /* INFO 白色 */
}
.warn-level :deep(.bk-checkbox-label) {
  color: #FF9C01; /* WARN 橙色 */
}
.error-level :deep(.bk-checkbox-label) {
  color: #EA3636; /* ERROR 红色 */
}
.debug-level :deep(.bk-checkbox-label) {
  color: #808A94; /* DEBUG 灰色 */
}
</style>
