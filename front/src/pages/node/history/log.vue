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
        <div class="w-[258px] border-r border-[#0A0A0A] text-[#a8acb8]">
          <div
            v-for="(item, key, index) in logData.oper_inst_logs"
            :key="key"
            @click="handleToggleLogItem(key)"
            :class="[
              'h-[32px] flex items-center pl-[16.8px] cursor-pointer',
              { 'bg-[#242424]': activeKey === key },
            ]"
          >
            <success
              v-if="item.life_cycle?.state === 'success'"
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
            <Spinner
              v-else-if="item.life_cycle?.state === 'pending' && !hasError"
              width="12.25px"
              height="12.25px"
            />
            <span class="ml-[7px]">{{ index + 1 }}.</span>
            <span class="ml-[2px] mr-[4px]">{{ key }}</span>
            <span
              v-if="item.life_cycle?.end_time >= 0 && item.life_cycle?.start_time >= 0"
              >{{
                item.life_cycle?.end_time - item.life_cycle?.start_time
              }}s</span
            >
          </div>
        </div>
        <div class="flex-1 text-[#a8acb8]" v-if="logs">
          <div
            v-for="(item, index) in logs"
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
  </SlideDetail>
</template>
<script setup lang="ts">
import { ref, reactive, computed, onMounted, onBeforeUnmount, watch } from "vue";
import useFullScreen from "@/composables/use-fullscreen";
import useInterval from "@/composables/use-interval";
import { NodeWorkflowService } from "@/api/modules/node_workflow";
import dayjs from "dayjs";
import SlideDetail from "@/components/slide-detail.vue";
import { useI18n } from "vue-i18n";
import { RightTurnLine, Success, Close, AngleUpFill, Spinner } from 'bkui-vue/lib/icon';
import { Button, Dropdown } from 'bkui-vue';

interface IProps {
  data: any;
}
const props = defineProps<IProps>();
const { t } = useI18n();
// 全屏
const { contentRef, isFullscreen, switchFullScreen } = useFullScreen();
const { start, stop } = useInterval(getLog, 5000); // 轮询
const activeKey = ref("");
const curOperInstId = ref("");
const curOperInstVal = ref("LATEST");
const curOperationId = ref("");
const curSortNames = ref<string[]>([]);
const logs = ref<{ text: string; level: string; time: string }[]>([]);
const operInstList = ref<{ id: string; name: string; sort_names: string[] }[]>(
  []
);
const slideDetailRef = ref<InstanceType<typeof SlideDetail>>();
const logData = ref<{
  total: number;
  oper_inst_logs: Record<string, ActionMessage>;
}>({
  total: 0,
  oper_inst_logs: {},
});
const timeFormatter = (
  val: number | string | undefined,
  format = "YYYY-MM-DD HH:mm:ss"
) => {
  return val ? dayjs(val).format(format) : "--";
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
const getInstance = async () => {
  const res = await NodeWorkflowService.NodeWorkflowOperationInstanceList({
    operation_id: curOperationId.value,
  }).catch(() => ({
    oper_inst_data: [],
    total: 0,
  }));
  curOperInstId.value =
    res.total > 0 ? res.oper_inst_data[res.total - 1].oper_inst_id : "";
  curSortNames.value =
    res.total > 0 ? res.oper_inst_data[res.total - 1].action_names : [];
  operInstList.value = [];
  for (let i = 1; i <= res.total; i++) {
    operInstList.value.push({
      name: i < res.total ? `${i}nd` : "LATEST",
      id: res.oper_inst_data[i - 1].oper_inst_id,
      sort_names: res.oper_inst_data[i - 1].action_names,
    });
  }
};
const hasError = ref(false);
const isInterval = ref(false);
async function getLog() {
  const res = await NodeWorkflowService.NodeWorkflowOperationInstanceLogGet({
    oper_inst_id: curOperInstId.value,
  }).catch(() => ({
    total: 0,
    oper_inst_logs: {},
  }));

  curSortNames.value.forEach((key: string) => {
    logData.value.oper_inst_logs[key] = res.oper_inst_logs[key];
  });
  logData.value.total = res.total;

  let currentKey;
  for (const [key, entry] of Object.entries(logData.value.oper_inst_logs)) {
    if (entry.life_cycle?.state === "failed") {
      currentKey = key;
      hasError.value = true;
      isInterval.value = false;
      break;
    } else if (entry.life_cycle?.state === "pending") {
      currentKey = key;
      isInterval.value = true;
      break;
    }
    currentKey = key;
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
  }
})
onBeforeUnmount(() => {
  stop();
});
defineExpose({
  show,
});
</script>
