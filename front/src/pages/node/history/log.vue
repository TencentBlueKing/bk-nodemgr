<template>
  <div class="flex w-full h-full">
    <div class="w-[240px] h-full flex flex-col">
      <div class="h-[72px] p-[20px]">
        <Input v-model="searchValue" :placeholder="'请搜索ip'"></Input>
      </div>
      <bk-loading title="数据加载中" :loading="operateLoading" class="flex-1 h-[calc(100%-72px)]">
        <div class="h-full overflow-y-auto">
          <div
            v-for="operate in filterIpOpearateList" :key="operate.operation_id"
            class="cursor-pointer w-full px-[20px] h-[40px] leading-[40px] flex items-center"
            :class="{ 'bg-[#e1ecff]': route.params.ip === operate.bk_host_inner_list }"
            @click="handleChangeIp(operate.bk_host_inner_list)"
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
            <bk-overflow-title type="tips" class="text-[#63656e] w-[179px]">
              {{ operate.bk_host_inner_list }}
            </bk-overflow-title>
          </div>
        </div>
      </bk-loading>
    </div>
    <div class="bg-[#fff] px-[24px] py-[20px] h-full flex-1 flex flex-col">
      <div class="flex items-center h-[30px]">
        <ArrowsLeft width="24" height="24" class="text-[#3a84ff] cursor-pointer mr-[5px]" @click="router.back()" />
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
                  :disabled="currentOperate?.state === 'terminate' && item.id === 'PARTIAL'"
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
                {{ row.costTime > 0 ? row.costTime : 0 }}s
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
                  <span :class="['ml-[5px]', { 'text-[#c4c6cc]': row.state === 'pending' }]">
                    {{ statusMap[row.state]?.text }}
                  </span>
                </div>
                <Button text theme="primary" v-if="row.state === 'running'">终止</Button>
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
          class="flex-1 flex flex-col overflow-y-auto"
          :class="[{ 'text-[12px]': !isFullscreen, 'text-[16px]': isFullscreen }]"
        >
          <div
            class="sticky top-0 z-10 h-[50px]
              flex flex-shrink-0 justify-between items-center px-[16px] bg-[#202024] text-[#C4C6CC]"
          >
            <div>{{ t("执行日志") }}</div>
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
          <div class="bg-[#313238] flex-1 flex">
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
</template>
<script setup lang="ts">
import { Button, Dropdown, Input, overflowTitle } from 'bkui-vue';
import { AngleUpFill, ArrowsLeft, Close, RightShape, Spinner } from 'bkui-vue/lib/icon';
import dayjs from 'dayjs';
import { computed, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';

import { Table, TableColumn } from '@blueking/table';

import { NodeWorkflowService } from '@/api/modules/node_workflow';
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
const activeKey = ref('');

const operateList = ref<any[]>([]); // 子任务列表
// 搜索过滤
// eslint-disable-next-line max-len
const filterIpOpearateList = computed(() => operateList.value.filter(item => !searchValue.value || item.bk_host_inner_list.includes(searchValue.value)));

const searchValue = ref();
const title = computed(() => `${route.params.ip} ${typeMap[nodeManageStore.taskHistoryTableRowData.type] ?? ''} 的执行日志`);
const currentOperate = computed(() => operateList.value.find(item => item.bk_host_inner_list === route.params.ip));
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
const statusMap = {
  running: {
    text: '执行中',
  },
  failed: {
    text: '执行失败',
    icon: 'terminated',
  },
  timeout: {
    text: '超时',
    icon: 'terminated',
  },
  success: {
    text: '执行成功',
    icon: 'running',
  },
  skipped: {
    text: '执行成功',
    icon: 'running',
  },
  pending: {
    text: '等待执行',
    icon: 'unknown',
  },
  terminate: {
    text: '终止',
    icon: 'terminated',
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

const timeFormatter = (
  val: number | string | undefined,
  format = 'YYYY-MM-DD HH:mm:ss',
) => {
  if (typeof val === 'number') {
    // 使用 dayjs.unix() 直接解析秒级时间戳
    return dayjs.unix(val).format(format);
  }
  return val ? dayjs(val).format(format) : '--';
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
  const res = await NodeWorkflowService.NodeWorkflowOperationRetry({
    workflow_id: route.params.taskId,
    operation_id: [row.operation_id],
    retry_mod: type,
  }).catch(() => false);
  if (res !== false) {
    isInterval.value = true;
    await getOperateList();
    await getInstance();
    start();
    mainStore.updateLogRetry(true);
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
};
const activeStepKey = ref();
const handleClickStep = (row: any) => {
  activeStepKey.value = row.stepKey;
  scrollToFirstErrorByClassNames(`.${row.stepKey}`);
};

// 搜索Ip
const handleChangeIp = async (ip: string) => {
  router.replace({
    name: 'log',
    params: {
      ip,
      taskId: route.params.taskId,
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
  const res = await NodeWorkflowService.NodeWorkflowOperationList({
    exact_include_conditions: {
      workflow_id: route.params.taskId,
    },
  }).catch(() => ({
    operations: [],
    total_count: 0,
  }));
  operateLoading.value = false;
  operateList.value = res.operations.map(item => ({
    ...item.param,
    ...item.status,
    bk_host_inner_list: item.param.bk_host_inner_list?.join(','),
    bk_host_innerip_v6_list: item.param.bk_host_innerip_v6_list?.join(','),
    operation_id: item.operation_id,
  }));
};

const instanceLoading = ref(false);
const getInstance = async () => {
  instanceLoading.value = true;
  const res = await NodeWorkflowService.NodeWorkflowOperationInstanceList({
    operation_id: currentOperate.value.operation_id,
  }).catch(() => ({
    oper_inst_data: [],
    total: 0,
  }));
  instanceLoading.value = false;
  curOperInstId.value = res.total > 0 ? res.oper_inst_data[res.total - 1].oper_inst_id : '';
  curSortNames.value = res.total > 0 ? res.oper_inst_data[res.total - 1].action_names : [];
  operInstList.value = [];
  for (let i = 1; i <= res.total; i++) {
    operInstList.value.push({
      name: i < res.total ? getOrdinalSuffix(i) : 'latest',
      id: res.oper_inst_data[i - 1].oper_inst_id,
      sort_names: res.oper_inst_data[i - 1].action_names,
    });
  }
};
const hasErrorOrTimeout = ref(false);
const isInterval = ref(false);
async function getLog() {
  const res = await NodeWorkflowService.NodeWorkflowOperationInstanceLogGet({
    oper_inst_id: curOperInstId.value,
  }).catch(() => ({
    total: 0,
    oper_inst_logs: {},
  }));
  const list: any[] = [];
  curSortNames.value.forEach((key: string, index: number) => {
    logData.value.oper_inst_logs[key] = res.oper_inst_logs[key];
    const { start_time, end_time, state } = res.oper_inst_logs[key].life_cycle || {};
    list.push({
      index: index + 1,
      stepKey: key,
      costTime: start_time && end_time ? end_time - start_time : 0,
      state: state || '未知',
    });
  });
  logData.value.total = res.total;
  tableData.value = list;

  let currentKey;
  for (const [key, entry] of Object.entries(logData.value.oper_inst_logs)) {
    if (['failed', 'timeout'].includes(entry.life_cycle?.state)) {
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
.nc-warning {
  &::before {
    border-color: #f59500;
    background: #fce5c0;
  }
}
.nc-unknown {
  &::before {
    border-color: #b2b5bd;
    background: #f0f1f5;
  }
}
</style>
