<template>
  <page-header :title="title" @click="handleBackToHistoryDetail" class="border-b border-[#e4e7ef]"></page-header>
  <div class="flex w-full h-[calc(100%-52px)]">
    <div class="w-[280px] h-full flex flex-col">
      <div class="h-[72px] p-[20px]">
        <SearchSelect
          class="flex-1 bg-[#fff]"
          ref="searchSelect"
          :data="searchSelectData"
          v-model.trim="searchSelectValue"
          :unique-select="true"
          :placeholder="isNode
            ? $t('platform.nodeMan.log.searchNode')
            : $t('platform.nodeMan.log.searchPlugin')"
          @update:model-value="handleSearchSelectChange"
        >
        </SearchSelect>
      </div>
      <bk-loading :title="$t('table.loading')" :loading="operateLoading" class="flex-1 h-[calc(100%-72px)]">
        <div class="h-full overflow-y-auto">
          <div
            v-for="(operate, index) in filterOperateList" :key="index"
            class="cursor-pointer w-full px-[20px] h-[40px] leading-[40px] flex items-center"
            :class="{ 'bg-[#e1ecff]': isNode
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
      </bk-loading>
    </div>
    <div class="bg-[#fff] px-[24px] pb-[20px] h-full flex-1 flex flex-col">
      <div class="h-[32px] mt-[20px]" v-if="!['success'].includes(currentOperate?.state)">
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
            :min-width="isManual || isOffline ? 230 : 150">
            <template #default="{ row }">
              <div class="flex items-center gap-[5px]">
                <div class="flex items-center flex-1">
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
                  <div :class="['ml-[5px]']">
                    <div v-if="showOfflineGuideButton(row)">
                      {{ $t('platform.nodeMan.log.waitOfflineOperation') }}
                      <Button class="ml-[2px]" text theme="primary" @click="handleOfflineGuide">
                        {{ $t('platform.nodeMan.log.offlineGuide') }}
                      </Button>
                    </div>
                    <!-- eslint-disable-next-line max-len -->
                    <div v-else-if="isManual && last_oper_inst_step_key === row.stepKey && currentOperate.state === 'running'">
                      {{ $t('platform.nodeMan.log.waitManualOperation') }}
                      <Button class="ml-[2px]" text theme="primary" @click="handleOperateGuide">
                        {{ $t('platform.nodeMan.log.operationGuide') }}
                      </Button>
                    </div>
                    <span v-else :class="[{ 'text-[#c4c6cc]': row.state === 'pending' }]">
                      {{ statusMap[row.state]?.text }}
                    </span>
                  </div>
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
          <TableColumn fixed="right" min-width="20">
            <template #default="{ row }">
              <div class="text-right">
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
            <div>{{ $t('platform.nodeMan.log.executionLog') }}</div>
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
            <div class="flex-1 text-[#a8acb8]" v-if="allLogs">
              <div v-for="logItem in allLogs" :key="logItem.key" :class="logItem.key">
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
import { Button, Dropdown, InfoBox, Input, Message, overflowTitle, SearchSelect } from 'bkui-vue';
import {
  AngleUpFill,
  ArrowsLeft,
  Close,
  ExclamationCircleShape,
  RightShape,
  Spinner,
} from 'bkui-vue/lib/icon';
import dayjs from 'dayjs';
import { debounce } from 'lodash';
import { computed, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';

import { Table, TableColumn } from '@blueking/table';

import guide from './guide.vue';
import { isOfflineGuideStep, STEP_KEY_WAIT_OFFLINE_MANUAL_INSTALL } from './offline-package';

import { NodeWorkflowService } from '@/api/modules/node_workflow';
import { PluginWorkflowService } from '@/api/modules/plugin_workflow';
import useInterval from '@/composables/use-interval';
import { useMainStore } from '@/stores/main';
import { useNodeManageStore } from '@/stores/node-manage';

const emit = defineEmits(['stop']);
const { t } = useI18n();
const route = useRoute();
const router = useRouter();

const mainStore = useMainStore();
const nodeManageStore = useNodeManageStore();

// 语言判断
const isZh = computed(() => mainStore.curLanguage === 'zh-CN');

// 获取 action 显示名称
const getDisplayName = (row: any) => {
  return isZh.value
    ? (row.display_name_zh || row.stepKey)
    : (row.display_name_en || row.stepKey);
};

// 获取日志文本
const getLogText = (item: any) => {
  return isZh.value
    ? (item.text_zh || '')
    : (item.text_en || '');
};

const isExecutionLogError = (item: any) => item.level === 'ERROR' || getLogText(item).includes('ERROR');
const isExecutionLogWarn = (item: any) => !isExecutionLogError(item) && (item.level === 'WARN' || getLogText(item).includes('WARN'));

// 正则表达式
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
const operateList = ref<any[]>([]); // 子任务列表
const filterOperateList = computed(() => operateList.value.filter((row: any) =>
  searchSelectValue.value.every((searchItem: any) => {
    const { id: searchField, values } = searchItem;
    const searchIds = values?.map((value: { id: string }) => value.id);
    return searchIds.includes(row[searchField]);
  })
));
// 搜索过滤
const currentOperate = computed(() => operateList.value.find(item => isNode.value
  ? item.bk_host_id === Number(route.params.hostId)
  : route.params.hostId === (`${item.bk_host_id}_${item.plugin_name}`)));
const isManual = computed(() => !!currentOperate.value?.latest_action_inst_brief_data?.tags?.includes('need_manual_exec_install_script'));
const isOffline = computed(() => isOfflineGuideStep(currentOperate.value?.latest_action_inst_brief_data?.tags ?? []));

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

const timeFormatter = (
  val: number | string | undefined,
  format = 'YYYY-MM-DD HH:mm:ss',
) => {
  if (typeof val === 'number') {
    // 判断时间戳位数：10位为秒级，13位为毫秒级
    const timestampStr = val.toString();
    if (timestampStr.length === 10) {
      // 秒级时间戳，使用 dayjs.unix()
      return dayjs.unix(val).format(format);
    } else if (timestampStr.length === 13) {
      // 毫秒级时间戳，使用 dayjs()
      return dayjs(val).format(format);
    } else {
      // 其他长度的数字，默认按毫秒处理
      return dayjs(val).format(format);
    }
  }
  return val ? dayjs(val).format(format) : '--';
};

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
const searchSelectData = computed(() => [
  { id: 'bk_host_inner_list', name: 'IPv4', multiple: true },
  { id: 'bk_host_innerip_v6_list', name: 'IPv6', multiple: true },
  ...(isNode.value ? [] : [{ id: 'plugin_name', name: '插件名' }]),
  {
    id: 'state',
    name: '执行状态',
    children: distinctStates.value.map(value => ({
      id: value,
      name: statusMap.value[value]?.text || value,
    })),
    multiple: true,
  },
]);
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
      targetId = 'bk_host_inner_list';
      targetName = 'IPv4';
    } else if (IPV6_REG.test(text)) {
      targetId = 'bk_host_innerip_v6_list';
      targetName = 'ipV6';
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
// eslint-disable-next-line max-len
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
const handleOperateGuide = () => {
  isGuideShow.value = true;
  guideData.value = {
    workflow_id: route.params.taskId as string,
    operation_id: currentOperate.value.operation_id,
    bk_host_innerip: currentOperate.value?.bk_host_inner_list,
    is_offline: false,
  };
};
const handleOfflineGuide = () => {
  isGuideShow.value = true;
  guideData.value = {
    workflow_id: route.params.taskId as string,
    operation_id: currentOperate.value.operation_id,
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

const handleRetry = async (row: any, type: string) => {
  if (!route.params.taskId) return;
  
  const res = await serviceCaller.call('retry', {
    workflow_id: route.params.taskId,
    operation_ids: [row.operation_id],
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
  InfoBox({
    title: t('platform.nodeMan.log.terminateConfirmTitle'),
    subTitle: t('platform.nodeMan.log.terminateConfirmSubTitle'),
    onConfirm: async () => {
      const res = await serviceCaller.call('terminate', {
        workflow_id: route.params.taskId,
        operation_ids: [currentOperate.value.operation_id],
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
const handleChangeIp = async (hostId: number, plugin_name: string = '') => {
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
  const res = await serviceCaller.call('operationList', {
    page: { limit: 500, offset: 0 },
    workflow_id: route.params.taskId,
    exact_include_conditions: {
      state: route.query.status ? [route.query.status] : [],
    },
  }).catch(() => ({
    operations: [],
    total_count: 0,
  }));
  operateLoading.value = false;
  if (route.query.active !== 'plugin') {
    operateList.value = res.operations.map(item => {
      const briefData = item.latest_oper_inst_brief_data;
      return {
        ...item.node_deployment_info,
        ...(briefData ? briefData.life_cycle : {}),
        latest_action_inst_brief_data: briefData ? briefData.latest_action_inst_brief_data : {},
        bk_host_inner_list: item.node_deployment_info.bk_host_inner_list?.join(',') || item.node_deployment_info.bk_host_innerip_list?.join(','),
        bk_host_innerip_v6_list: item.node_deployment_info.bk_host_innerip_v6_list?.join(',') || item.node_deployment_info.bk_host_innerip_v6_list?.join(','),
        operation_id: item.operation_id,
      };
    });
  } else {
    operateList.value = res.operations.map(item => {
      const briefData = item.latest_oper_inst_brief_data;
      return {
        ...item.plugin_deployment_info,
        ...(briefData ? briefData.life_cycle : {}),
        latest_action_inst_brief_data: briefData ? briefData.latest_action_inst_brief_data : {},
        bk_host_inner_list: item.plugin_deployment_info.bk_host_inner_list?.join(',') || item.plugin_deployment_info.bk_host_innerip_list?.join(','),
        bk_host_innerip_v6_list: item.plugin_deployment_info.bk_host_innerip_v6_list?.join(',') || item.plugin_deployment_info.bk_host_innerip_v6_list?.join(','),
        operation_id: item.operation_id,
      };
    });
  }
};

const instanceLoading = ref(false);
const getInstance = async () => {
  instanceLoading.value = true;
  const params = route.query.active !== 'plugin'
    ? {
      operation_id: currentOperate.value.operation_id,
    }
    : {
      operation_id: [currentOperate.value.operation_id],
    };
  const res = await serviceCaller.call('operationInstanceList', params).catch(() => ({
    oper_inst_data: [],
  }));

  const total = res.oper_inst_data.length;
  instanceLoading.value = false;
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
};
const hasErrorOrTimeout = ref(false);
const isInterval = ref(false);
async function getLog() {
  if (!curOperInstId.value) return;
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
    if (start_time && end_time) {
      if (end_time <= 0 && state !== 'pending') {
        costTime = new Date().getTime() - start_time;
        if (start_time < 0) {
          costTime = 0;
        }
      } else if (end_time > 0) {
        costTime = end_time - start_time;
      }
    }
    list.push({
      index: index + 1,
      stepKey: key,
      display_name_zh: actionData.display_name_zh || key,
      display_name_en: actionData.display_name_en || key,
      costTime,
      state: state || '未知',
    });
  });
  logData.value.total = Object.keys(operInstLogs).length;

  // 额外执行日志：仅判断是否存在且含有 ERROR（用于日志区域展示），不再插入步骤标签
  const hasExtraLogs = res.extra_execution_logs?.logs?.length > 0;
  const hasExtraError = hasExtraLogs
    && res.extra_execution_logs.logs.some((log: any) => log.level === 'ERROR');

  tableData.value = list;

  let currentKey;
  for (const [key, entry] of Object.entries(logData.value.oper_inst_logs)) {
    if (['failed', 'timeout', 'terminated'].includes((entry as any).life_cycle?.state)) {
      currentKey = key;
      hasErrorOrTimeout.value = true;
      isInterval.value = false;
      break;
    } else if ((entry as any).life_cycle?.state === 'running') {
      currentKey = key;
      isInterval.value = true;
      break;
    }
    currentKey = key;
  }
  if (currentKey && logData.value.oper_inst_logs[currentKey]?.life_cycle.state === 'success') {
    isInterval.value = false;
  }

  // 构建 allLogs：仅在额外日志包含 ERROR 时才加入日志区域展示
  const extraLogs = hasExtraError
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

const handleToggleLogItem = (key: string) => {
  logs.value = logData.value.oper_inst_logs[key].message.logs;
  activeKey.value = key;
};

watch(() => isInterval.value, async (val: boolean) => {
  if (!val) {
    stop();
    await getOperateList();
    await getInstance();
    // emit('stop');
  }
});
watch(() => route.path, async () => {
  await getInstance();
  await getLog();
  if (isInterval.value) {
    start();
  } else {
    stop();
  }
});

const debouncedGetOperateList = debounce(() => {
  getOperateList();
}, 1000);
watch(() => currentOperate.value, async () => {
  if (['launched', 'init'].includes(currentOperate.value?.state)) {
    await debouncedGetOperateList();
  }
});

onBeforeUnmount(() => {
  stop();
});
onMounted(async () => {
  await getOperateList();
  await getInstance();
  await getLog();
  await getDistinctStates();
  if (isInterval.value) {
    start();
  } else {
    stop();
  }
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
</style>
