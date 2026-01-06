<template>
  <div class="flex w-full h-full">
    <div class="w-[280px] h-full flex flex-col">
      <div class="h-[72px] p-[20px]">
        <Input v-model="searchValue" :placeholder="route.query.active === 'node' ? '请搜索ip' : '请搜索ip或插件名'"></Input>
      </div>
      <bk-loading title="数据加载中" :loading="operateLoading" class="flex-1 h-[calc(100%-72px)]">
        <div class="h-full overflow-y-auto">
          <div
            v-for="operate in filterIpOpearateList" :key="operate.operation_id"
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
    <div class="bg-[#fff] px-[24px] py-[20px] h-full flex-1 flex flex-col">
      <div class="flex items-center h-[30px]">
        <ArrowsLeft
          width="24"
          height="24"
          class="text-[#3a84ff] cursor-pointer mr-[5px]"
          @click="handleBackToHistoryDetail"
        />
        <span>{{ $t(title) }}</span>
      </div>
      <div class="h-[72px] my-[20px]">
        <Dropdown
          theme="light"
          trigger="click"
          ext-cls="dropdownCls"
          :popover-options="{
            clickContentAutoHide: true,
          }">
          <Button
            v-if="!['success'].includes(currentOperate?.state)"
            class="w-[86px]">
            <span>重试</span>
          </Button>
          <template #content>
            <Dropdown.DropdownMenu>
              <Dropdown.DropdownItem
                class="text-14px"
                v-for="item in reTryType"
                :key="item.id"
                v-bk-tooltips="{
                  content: item.tooltip
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
      <bk-loading title="数据加载中" :loading="instanceLoading" class="flex-1 flex h-[calc(100%-102px)]">
        <Table
          :data="tableData"
          :empty-text="'暂无数据'"
          :min-width="300"
          class="w-[40%] h-full mr-[10px] bg-[#f5f7fa]"
        >
          <TableColumn :title="'步骤'" min-width="200" fixed="left">
            <template #default="{ row }">
              <Button
                text
                :class="['cursor-pointer', { 'text-[#c6c4cc]': row.state === 'pending' }]"
                @click="handleClickStep(row)"
              >
                {{ row.index }}. {{ row.stepKey }}
              </Button>
            </template>
          </TableColumn>
          <TableColumn field="costTime" :title="'耗时'" min-width="60">
            <template #default="{ row }">
              <span :class="{ 'text-[#c6c4cc]': row.state === 'pending' }">
                {{ formatTimeToMS(row.costTime) }}
              </span>
            </template>
          </TableColumn>
          <TableColumn field="state" :title="'执行情况'" min-width="150">
            <template #default="{ row }">
              <div class="flex items-center">
                <div class="flex items-center w-[100px]">
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
                  <div>
                    <!-- eslint-disable-next-line max-len -->
                    <div v-if="currentOperate.latest_action_inst_brief_data?.tags.includes('need_manual_exec_install_script')
                      && last_oper_inst_step_key === row.stepKey && currentOperate.state === 'running'">
                      等待手动操作，查看
                      <Button class="ml-[2px]" text theme="primary" @click="handleOperateGuide">操作指引</Button>
                    </div>
                    <span v-else :class="['ml-[5px]', { 'text-[#c4c6cc]': row.state === 'pending' }]">
                      {{ statusMap[row.state]?.text }}
                    </span>
                  </div>
                </div>
                <Button
                  v-if="row.state === 'running'"
                  text
                  theme="primary"
                  @click="handleTerminate">
                  终止
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
        <div
          ref="contentRef"
          class="flex-1 flex flex-col"
          :class="[{ 'text-[12px]': !isFullscreen, 'text-[16px]': isFullscreen }]"
        >
          <div
            class="sticky top-0 z-10 h-[50px]
              flex flex-shrink-0 justify-between items-center px-[16px] bg-[#202024] text-[#C4C6CC]"
          >
            <div>执行日志</div>
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
              <Button text v-if="isFullscreen" @click="switchFullScreen">
                <i
                  class="nodeman-icon nc-icon-un-full-screen text-[16px] text-[#C4C6CC]"
                ></i>
              </Button>
              <Button text v-else @click="switchFullScreen">
                <i
                  class="nodeman-icon nc-icon-full-screen text-[16px] text-[#C4C6CC]"
                ></i>
              </Button>
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
                    'bg-[#422321] flex !mx-0':
                      item.level === 'ERROR' || item.text.includes('ERROR'),
                  }"
                >
                  <div class="flex justify-center items-baseline w-[26px] pt-[6px]">
                    <close
                      v-if="item.level === 'ERROR' || item.text.includes('ERROR')"
                      :fill="'#993D3D'"
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
                    <span class="line-height-[24px]">{{ item.text }}</span>
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
import { Button, Dropdown, Input, overflowTitle } from 'bkui-vue';
import { AngleUpFill, ArrowsLeft, Close, RightShape, Spinner } from 'bkui-vue/lib/icon';
import dayjs from 'dayjs';
import { debounce } from 'lodash';
import { computed, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';

import { Table, TableColumn } from '@blueking/table';

import guide from './guide.vue';

import { NodeWorkflowService } from '@/api/modules/node_workflow';
import { PluginWorkflowService } from '@/api/modules/plugin_workflow';
import useFullScreen from '@/composables/use-fullscreen';
import useInterval from '@/composables/use-interval';
import { useMainStore } from '@/stores/main';
import { useNodeManageStore } from '@/stores/node-manage';

const emit = defineEmits(['stop']);
const { t } = useI18n();
const route = useRoute();
const router = useRouter();

const mainStore = useMainStore();
const nodeManageStore = useNodeManageStore();
// 全屏
const { contentRef, isFullscreen, switchFullScreen } = useFullScreen();
const { start, stop } = useInterval(getLog, 1000); // 轮询

// 统一服务调用器
const serviceCaller = {
  // 根据路由参数获取当前服务类型
  getCurrentServiceType: () => (route.query.active === 'node' ? 'node' : 'plugin'),

  // 服务方法映射
  serviceMethods: {
    node: {
      retry: NodeWorkflowService.NodeWorkflowOperationRetry,
      terminate: NodeWorkflowService.NodeWorkflowOperationTerminate,
      operationList: NodeWorkflowService.NodeWorkflowOperationList,
      operationInstanceList: NodeWorkflowService.NodeWorkflowOperationInstanceList,
      operationInstanceLogGet: NodeWorkflowService.NodeWorkflowOperationInstanceLogGet,
    },
    plugin: {
      retry: PluginWorkflowService.PluginWorkflowOperationRetry,
      terminate: PluginWorkflowService.PluginWorkflowOperationTerminate,
      operationList: PluginWorkflowService.PluginWorkflowOperationList,
      operationInstanceList: PluginWorkflowService.PluginWorkflowOperationInstanceList,
      operationInstanceLogGet: PluginWorkflowService.PluginWorkflowOperationInstanceLogGet,
    },
  },

  // 统一调用方法
  async call(method: 'retry' | 'terminate' | 'operationList' | 'operationInstanceList' | 'operationInstanceLogGet', params: any) {
    const serviceType = this.getCurrentServiceType();
    const serviceMethod = this.serviceMethods[serviceType][method];
    return await serviceMethod(params);
  },
};
const activeKey = ref('');
const isNode = computed(() => route.query.active === 'node');
const operateList = ref<any[]>([]); // 子任务列表
// 搜索过滤
const filterIpOpearateList = computed(() => operateList.value.filter(item => !searchValue.value
    || item.bk_host_inner_list.includes(searchValue.value)
    || item.plugin_name?.includes(searchValue.value)
    || route.query.status === item.state));

const searchValue = ref();
const currentOperate = computed(() => operateList.value.find(item => isNode.value
  ? item.bk_host_id === Number(route.params.hostId)
  : route.params.hostId === (`${item.bk_host_id}_${item.plugin_name}`)));
const title = computed(() => `${currentOperate.value?.bk_host_inner_list ?? ''} ${typeMap.value[nodeManageStore.taskHistoryTableRowData.type] ?? ''} 的执行日志`);
const curOperInstId = ref('');
const curOperInstVal = ref('latest');
const curSortNames = ref<string[]>([]);
const logs = ref<{ text: string; level: string; time: string }[]>([]);
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
const statusMap = {
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
};
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

// 操作指引侧边栏
const isGuideShow = ref(false);
const guideData = ref<{workflow_id: string, operation_id: string, bk_host_innerip: string}>({
  workflow_id: '',
  operation_id: '',
  bk_host_innerip: '',
});
const handleOperateGuide = () => {
  isGuideShow.value = true;
  guideData.value = {
    workflow_id: route.params.taskId as string,
    operation_id: currentOperate.value.operation_id,
    bk_host_innerip: currentOperate.value?.bk_host_inner_list,
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
const handleTerminate = async () => {
  const res = await serviceCaller.call('terminate', {
    workflow_id: route.params.taskId,
    operation_ids: [currentOperate.value.operation_id],
  }).catch(() => false);
  if (res !== false) {
    await getOperateList();
    await getInstance();
    start();
    mainStore.updateLogTerminate(true);
  }
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
  operateLoading.value = true;
  const res = await serviceCaller.call('operationList', {
    page: { limit: 500, offset: 0 },
    workflow_id: route.params.taskId,
  }).catch(() => ({
    operations: [],
    total_count: 0,
  }));
  operateLoading.value = false;
  if (route.query.active === 'node') {
    operateList.value = res.operations.map(item => {
      const briefData = item.latest_oper_inst_brief_data;
      return {
        ...item.node_deployment_info,
        ...(briefData ? briefData.life_cycle : {}),
        latest_action_inst_brief_data: briefData ? briefData.latest_action_inst_brief_data : {},
        bk_host_inner_list: item.node_deployment_info.bk_host_inner_list?.join(',') || item.node_deployment_info.bk_host_innerip_list?.join(','),
        bk_host_innerip_v6_list: item.node_deployment_info.bk_host_innerip_v6_list?.join(',') || item.node_deployment_info.bk_host_innerip_v6_list?.join(','),
        operation_id: item.operation_id,
      }
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
      }
    });
  }
};

const instanceLoading = ref(false);
const getInstance = async () => {
  instanceLoading.value = true;
  const params = route.query.active === 'node'
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
  const res = await serviceCaller.call('operationInstanceLogGet', {
    oper_inst_id: curOperInstId.value,
  }).catch(() => ({
    total: 0,
    oper_inst_logs: {},
  }));
  const list: any[] = [];
  curSortNames.value.forEach((key: string, index: number) => {
    logData.value.oper_inst_logs[key] = res.oper_inst_logs[key];
    const { start_time, end_time, state } = res.oper_inst_logs[key].life_cycle || {};
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
      costTime,
      state: state || '未知',
    });
  });
  logData.value.total = Object.keys(res.oper_inst_logs).length;
  tableData.value = list;

  let currentKey;
  for (const [key, entry] of Object.entries(logData.value.oper_inst_logs)) {
    if (['failed', 'timeout', 'terminated'].includes(entry.life_cycle?.state)) {
      currentKey = key;
      hasErrorOrTimeout.value = true;
      isInterval.value = false;
      break;
    } else if (entry.life_cycle?.state === 'running') {
      currentKey = key;
      isInterval.value = true;
      break;
    }
    currentKey = key;
  }
  if (logData.value.oper_inst_logs[currentKey]?.life_cycle.state === 'success') {
    isInterval.value = false;
  }
  allLogs.value = curSortNames.value.map(key => ({
    key,
    logs: res.oper_inst_logs[key].message.logs,
  }));
  if (currentKey) {
    logs.value = { ...res.oper_inst_logs[currentKey].message.logs };
    activeKey.value = currentKey;
  }
};

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
