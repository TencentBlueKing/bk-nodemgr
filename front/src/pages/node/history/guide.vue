<template>
  <Sideslider
    :is-show="isShow"
    :width="1100"
    :title="`手动安装 ${data.bk_host_innerip || ''}`"
    @update:is-show="handleClose"
    @closed="handleClosed"
  >
    <Loading :loading="isLoading" class="min-h-[400px] bg-[#f5f7fa] !p-0">
      <div v-if="installSolutions.length > 0" class="p-8">

        <div class="mb-8 p-6 bg-white rounded shadow-sm border border-[#dcdee5]">
          <div class="text-[#63656e] mb-3 text-sm font-bold">安装方式:</div>
          <div class="flex items-center gap-4">
            <Radio.Group v-model="activeSolutionType" type="capsule">
              <Radio.Button
                v-for="item in installSolutions"
                :key="item.type"
                :label="item.type"
              >
                {{ item.type }}
              </Radio.Button>
            </Radio.Group>
            <span class="text-xs text-[#979ba5]">通过 {{ activeSolutionType }} 进行安装</span>
          </div>
        </div>

        <div v-if="currentSolution" class="mt-4">
          <div class="mb-8 text-sm font-bold text-[#313238]">
            在目标主机通过 <span class="text-[#3a84ff]">{{ activeSolutionType }}</span> 安装:
          </div>

          <div class="relative ml-4">
            <div
              v-for="(step, index) in currentSolution.steps"
              :key="index"
              class="relative pl-10 pb-10 last:pb-0"
            >
              <div
                v-if="index !== currentSolution.steps.length - 1"
                class="absolute left-[11px] top-[26.8px] w-[2px] h-[calc(100%-24px)] bg-[#3a84ff] opacity-20"
              ></div>

              <!-- eslint-disable-next-line max-len -->
              <div class="absolute left-0 top-[3px] w-[24px] h-[24px] rounded-full bg-[#3a84ff] text-white flex items-center justify-center text-sm font-bold z-10 shadow-sm">
                {{ index + 1 }}
              </div>

              <div class="flex flex-col gap-4">
                <div class="text-sm font-bold text-[#313238] leading-8">
                  {{ step.name_zh }}
                </div>
                <!-- eslint-disable-next-line max-len -->
                <div class="group relative bg-[#2f323b] rounded-md p-5 border border-transparent hover:border-[#3a84ff] transition-all shadow-lg">
                  <div class="flex items-start">
                    <i class="bk-icon icon-file text-[#979ba5] mt-1 mr-4 text-lg"></i>

                    <!-- eslint-disable-next-line max-len -->
                    <pre class="flex-1 text-[#dcdee5] font-mono text-xs whitespace-pre-wrap break-all leading-6">{{ step.content_zh }}</pre>

                    <Button
                      theme="primary"
                      size="small"
                      class="ml-4 px-6 h-8 flex-shrink-0"
                      @click="handleCopy(step.content_zh)"
                    >
                      复制
                    </Button>
                  </div>

                  <!-- eslint-disable-next-line max-len -->
                  <div class="absolute left-0 top-0 bottom-0 w-1 bg-[#3a84ff] rounded-l-md opacity-0 group-hover:opacity-100 transition-opacity"></div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <Exception
        v-else-if="!isLoading"
        type="empty"
        scene="part"
        class="mt-20"
      />
    </Loading>
  </Sideslider>
</template>

<script setup lang="ts">
import {
  Button,
  Exception,
  Loading,
  Message,
  Radio,
  Sideslider } from 'bkui-vue';
import { computed, ref, watch } from 'vue';

import { NodeWorkflowService } from '@/api/modules/node_workflow';

// 类型定义
interface IStep {
  name_zh: string;
  content_zh: string;
}

interface ManualSolution {
  type: string;
  description_zh: string;
  steps: IStep[];
}

interface IData {
  workflow_id: string | number;
  operation_id: string | number;
  bk_host_innerip: string;
}

const props = defineProps<{
  isShow: boolean;
  data: IData;
}>();

const emit = defineEmits(['update:isShow']);

const isLoading = ref(false);
const installSolutions = ref<ManualSolution[]>([]);
const activeSolutionType = ref('');

// 计算当前选中的方案
const currentSolution = computed(() => installSolutions.value.find(s => s.type === activeSolutionType.value));

const handleClose = (val: boolean) => {
  emit('update:isShow', val);
};

// 监听开启
watch(() => props.isShow, async (val) => {
  if (val && props.data.workflow_id) {
    fetchData();
  }
}, { immediate: true });

const fetchData = async () => {
  isLoading.value = true;
  try {
    const res = await NodeWorkflowService.NodeWorkflowOperationManualSolutionGet({
      workflow_id: props.data.workflow_id,
      operation_id: props.data.operation_id,
    });
    installSolutions.value = res;

    if (res.length > 0) {
      activeSolutionType.value = res[0].type;
    }
  } catch (error) {
    console.error('Fetch error:', error);
  } finally {
    isLoading.value = false;
  }
};

const handleClosed = () => {
  installSolutions.value = [];
  activeSolutionType.value = '';
};

const handleCopy = (content: string) => {
  if (!content) return;
  const textarea = document.createElement('textarea');
  textarea.value = content;
  textarea.style.position = 'fixed';
  textarea.style.opacity = '0';
  document.body.appendChild(textarea);
  textarea.select();
  try {
    document.execCommand('copy');
    Message({ theme: 'success', message: '已成功复制命令', delay: 1500 });
  } finally {
    document.body.removeChild(textarea);
  }
};
</script>

<style scoped>
/* 确保侧边栏内容无内边距 */
:deep(.bk-sideslider-content) {
  padding: 0 !important;
}

/* 优化代码块滚动条 */
pre::-webkit-scrollbar {
  width: 4px;
}
pre::-webkit-scrollbar-thumb {
  background: #63656e;
  border-radius: 2px;
}
</style>
