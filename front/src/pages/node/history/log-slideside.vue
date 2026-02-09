<template>
  <SlideDetail class="bg-[#fff]" ref="slideDetailRef">
    <div
      ref="contentRef"
      class="h-[500px] flex flex-col overflow-y-auto"
      :class="{ 'text-[12px]': !isFullscreen, 'text-[16px]': isFullscreen }"
    >
      <div
        class="h-[50px] flex flex-shrink-0 justify-between items-center px-[16px] bg-[#2E2E2E] text-[#C4C6CC]"
      >
        <div>{{ t("platform.nodeMan.log.executionLog") }}</div>
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
              <ul class="w-[80px] py-[4px]">
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
      <div class="bg-[#1A1A1A] flex-1 flex">
        <div class="w-[258px] border-r border-[#0A0A0A] text-[#a8acb8]" :class="{ 'w-[360px]': isFullscreen }">
          <div
            v-for="(item, key, index) in logData.oper_inst_logs"
            :key="`${key}${Math.random()}`"
            @click="handleToggleLogItem(key)"
            :class="[
              'h-[32px] flex items-center pl-[16.8px] cursor-pointer',
              { 'bg-[#242424]': activeKey === key },
            ]"
          >
            <success
              v-if="['success', 'skipped'].includes(item.life_cycle?.state)"
              width="12.25px"
              height="12.25px"
              :fill="activeKey === key ? '#24954f' : '#4D4F56'"
            />
            <close
              v-else-if="item.life_cycle?.state === 'failed'"
              width="12.25px"
              height="12.25px"
              :fill="activeKey === key ? '#993D3D' : '#4D4F56'"
            />
            <i class="nodeman-icon nc-history" v-else-if="item.life_cycle?.state === 'timeout'"></i>
            <Spinner
              v-else-if="['pending', 'running'].includes(item.life_cycle?.state) && !hasErrorOrTimeout"
              width="12.25px"
              height="12.25px"
            />
            <span class="ml-[7px]">{{ index + 1 }}.</span>
            <bk-overflow-title class="overflow-ellipsis w-[200px]" :class="{ 'w-[300px]': isFullscreen }">
              <span class="ml-[2px] mr-[4px]">{{ getDisplayName(key) }}</span>
              <span
                v-if="item.life_cycle?.end_time >= 0 && item.life_cycle?.start_time >= 0"
              >{{
                item.life_cycle?.end_time - item.life_cycle?.start_time
              }}s</span
              >
            </bk-overflow-title>
          </div>
        </div>
        <div class="flex-1 text-[#a8acb8]" v-if="logs">
          <div
            v-for="(item, index) in logs"
            :key="index"
            class="mx-[30px] my-[8px]"
            :class="{
              'bg-[#422321] flex !mx-0':
                item.level === 'ERROR' || getLogText(item).includes('ERROR'),
            }"
          >
            <div class="flex justify-center items-baseline w-[26px] pt-[6px]">
              <close
                v-if="item.level === 'ERROR' || getLogText(item).includes('ERROR')"
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
              <span class="line-height-[24px]">{{ getLogText(item) }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </SlideDetail>
</template>
<script setup lang="ts">
import { Button, Dropdown, overflowTitle } from 'bkui-vue';
import { AngleUpFill, Close, RightTurnLine, Spinner, Success } from 'bkui-vue/lib/icon';
import dayjs from 'dayjs';
import { computed, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';

import { NodeWorkflowService } from '@/api/modules/node_workflow';
import SlideDetail from '@/components/slide-detail.vue';
import useFullScreen from '@/composables/use-fullscreen';
import useInterval from '@/composables/use-interval';
import { useMainStore } from '@/stores/main';

interface IProps {
  data: any;
}
const props = defineProps<IProps>();
const emit = defineEmits(['stop']);
const { t } = useI18n();

const mainStore = useMainStore();

// 全屏
const { contentRef, isFullscreen, switchFullScreen } = useFullScreen();
const { start, stop } = useInterval(getLog, 1000); // 轮询
const activeKey = ref('');
const curOperInstId = ref('');
const curOperInstVal = ref('latest');
const curOperationId = ref('');
const curSortNames = ref<string[]>([]);
const logs = ref<{ text: string; text_zh?: string; text_en?: string; level: string; time: string }[]>([]);
const operInstList = ref<{ id: string; name: string; sort_names: string[] }[]>([]);
const slideDetailRef = ref<InstanceType<typeof SlideDetail>>();
const logData = ref<{
  total: number;
  oper_inst_logs: Record<string, ActionMessage>;
}>({
  total: 0,
  oper_inst_logs: {},
});

// 语言判断
const isZh = computed(() => mainStore.curLanguage === 'zh-CN');

// 根据 key 获取 action 显示名称
const getDisplayName = (key: string) => {
  const actionData = logData.value.oper_inst_logs[key];
  if (!actionData) return key;
  return isZh.value
    ? (actionData.display_name_zh || key)
    : (actionData.display_name_en || key);
};

// 获取日志文本
const getLogText = (item: any) => {
  return isZh.value
    ? (item.text_zh || '')
    : (item.text_en || '');
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
const getInstance = async () => {
  const res = await NodeWorkflowService.NodeWorkflowOperationInstanceList({
    operation_id: curOperationId.value,
  }).catch(() => ({
    oper_inst_data: [],
    total: 0,
  }));
  const total = res.oper_inst_data.length;
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
  const res = await NodeWorkflowService.NodeWorkflowOperationInstanceLogGet({
    oper_inst_id: curOperInstId.value,
  }).catch(() => ({
    total: 0,
    oper_inst_logs: {},
  }));
  const total = Object.keys(res.oper_inst_logs).length;
  curSortNames.value.forEach((key: string) => {
    logData.value.oper_inst_logs[key] = res.oper_inst_logs[key];
  });
  logData.value.total = total;

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
  if (currentKey) {
    logs.value = { ...res.oper_inst_logs[currentKey].message.logs };
    activeKey.value = currentKey;
  }
};

const handleToggleLogItem = (key: string) => {
  logs.value = logData.value.oper_inst_logs[key].message.logs;
  activeKey.value = key;
};
// 显示面板
const show = async () => {
  slideDetailRef.value?.show();
  curOperationId.value = props.data.operation_id;
  hasErrorOrTimeout.value = false;
  await getInstance();
  await getLog();
  if (isInterval.value) {
    start();
  } else {
    stop();
  }
};
watch(() => isInterval.value, (val: boolean) => {
  if (!val) {
    stop();
    emit('stop');
  }
});
onBeforeUnmount(() => {
  stop();
});
defineExpose({
  show,
});
</script>
