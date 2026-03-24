<template>
  <Sideslider
    :is-show="isShow"
    :width="1100"
    :title="sidesliderTitle"
    @update:is-show="handleClose"
    @closed="handleClosed"
  >
    <Loading :loading="isLoading" class="min-h-[400px] bg-[#f5f7fa] !p-0">

      <!-- offline install guide -->
      <div v-if="data.is_offline" class="p-8">
        <div class="relative ml-4">
          <div
            v-for="(step, index) in offlineGuideSteps"
            :key="step.key"
            class="relative pl-10 pb-10 last:pb-0"
          >
            <div
              v-if="index !== offlineGuideSteps.length - 1"
              class="absolute left-[11px] top-[26.8px] w-[2px] h-[calc(100%-24px)] bg-[#3a84ff] opacity-20"
            ></div>

            <!-- eslint-disable-next-line max-len -->
            <div class="absolute left-0 top-[3px] w-[24px] h-[24px] rounded-full bg-[#3a84ff] text-white flex items-center justify-center text-sm font-bold z-10 shadow-sm">
              {{ index + 1 }}
            </div>

            <div class="flex flex-col gap-4">
              <div class="text-sm font-bold text-[#313238] leading-8">
                {{ locale === 'zh-CN' ? step.name_zh : step.name_en }}
              </div>

              <!-- step description -->
              <div
                v-if="step.desc_zh || step.desc_en"
                class="text-sm text-[#63656e]"
              >
                {{ locale === 'zh-CN' ? step.desc_zh : step.desc_en }}
              </div>

              <!-- download button for "download" step -->
              <div v-if="step.key === 'download'">
                <Button
                  theme="primary"
                  :loading="isDownloading"
                  @click="handleDownloadPackage"
                >
                  {{ $t('platform.nodeMan.guide.offline.downloadPackage') }}
                </Button>
              </div>

              <!-- extract command block for "transfer" step -->
              <!-- eslint-disable max-len -->
              <div
                v-if="step.key === 'transfer'"
                class="group relative bg-[#2f323b] rounded-md p-5 border border-transparent hover:border-[#3a84ff] transition-all shadow-lg"
              >
                <div class="flex items-start">
                  <i class="bk-icon icon-file text-[#979ba5] mt-1 mr-4 text-lg"></i>
                  <pre class="flex-1 text-[#dcdee5] font-mono text-xs whitespace-pre-wrap break-all leading-6">{{ offlineExtractCmd }}</pre>
                  <Button
                    theme="primary"
                    size="small"
                    class="ml-4 px-6 h-8 flex-shrink-0"
                    @click="handleCopy(offlineExtractCmd)"
                  >
                    {{ $t('action.copy') }}
                  </Button>
                </div>
                <div class="absolute left-0 top-0 bottom-0 w-1 bg-[#3a84ff] rounded-l-md opacity-0 group-hover:opacity-100 transition-opacity"></div>
              </div>
              <!-- eslint-enable max-len -->

              <!-- execute command block for "execute" step -->
              <!-- eslint-disable max-len -->
              <div
                v-if="step.key === 'execute'"
                class="group relative bg-[#2f323b] rounded-md p-5 border border-transparent hover:border-[#3a84ff] transition-all shadow-lg"
              >
                <div class="flex items-start">
                  <i class="bk-icon icon-file text-[#979ba5] mt-1 mr-4 text-lg"></i>
                  <pre class="flex-1 text-[#dcdee5] font-mono text-xs whitespace-pre-wrap break-all leading-6">{{ offlineCdAndRunCmd }}</pre>
                  <Button
                    theme="primary"
                    size="small"
                    class="ml-4 px-6 h-8 flex-shrink-0"
                    @click="handleCopy(offlineCdAndRunCmd)"
                  >
                    {{ $t('action.copy') }}
                  </Button>
                </div>
                <div class="absolute left-0 top-0 bottom-0 w-1 bg-[#3a84ff] rounded-l-md opacity-0 group-hover:opacity-100 transition-opacity"></div>
              </div>
              <!-- eslint-enable max-len -->

              <!-- result submit area for the last step -->
              <div v-if="step.has_result_input" class="flex flex-col gap-3">
                <Input
                  v-model="resultData"
                  type="textarea"
                  :rows="8"
                  :placeholder="$t('platform.nodeMan.guide.offline.resultPlaceholder')"
                  class="font-mono text-xs"
                />
                <div class="flex items-center gap-3">
                  <Button
                    theme="primary"
                    :loading="isSubmitting"
                    :disabled="!resultData.trim()"
                    @click="handleSubmitResult"
                  >
                    {{ $t('platform.nodeMan.guide.offline.submitResult') }}
                  </Button>
                  <span v-if="submitSuccess" class="text-sm text-[#2dcb56]">
                    {{ $t('platform.nodeMan.guide.offline.submitSuccess') }}
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- manual install guide (existing) -->
      <div v-else-if="installSolutions.length > 0" class="p-8">

        <div class="mb-8 p-6 bg-white rounded shadow-sm border border-[#dcdee5]">
          <div class="text-[#63656e] mb-3 text-sm font-bold">{{ $t('platform.nodeMan.installAgentPage.type') }}:</div>
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
            <span class="text-xs text-[#979ba5]">
              {{ $t('platform.nodeMan.guide.installWith', { type: activeSolutionType }) }}
            </span>
          </div>
        </div>

        <div v-if="currentSolution" class="mt-4">
          <div class="mb-8 text-sm font-bold text-[#313238]">
            {{ $t('platform.nodeMan.guide.onHost') }}
            <span class="text-[#3a84ff]"> {{ activeSolutionType }} </span>
            {{ $t('platform.nodeMan.guide.install') }}:
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
                      {{ $t('action.copy') }}
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
        v-else-if="!isLoading && !data.is_offline"
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
  Input,
  Loading,
  Message,
  Radio,
  Sideslider } from 'bkui-vue';
import { computed, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';

import { offlineGuideSteps } from './offline-guide';
import { OFFLINE_PACKAGE_DOWNLOAD_PATH } from './offline-package';

import type { Config } from '@/api/interceptors';
import { fetch as httpFetch } from '@/api/interceptors';
import { NodeWorkflowService } from '@/api/modules/node_workflow';

// types for manual install solution.
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
  bk_networkarea_id?: number;
  // is_offline indicates this is an offline install guide (show offline steps instead of API steps).
  is_offline?: boolean;
}

const props = defineProps<{
  isShow: boolean;
  data: IData;
}>();

const emit = defineEmits(['update:isShow']);
const { t, locale } = useI18n();
const isLoading = ref(false);
const installSolutions = ref<ManualSolution[]>([]);
const activeSolutionType = ref('');

// offline state
const isDownloading = ref(false);
const resultData = ref('');
const isSubmitting = ref(false);
const submitSuccess = ref(false);

// offline package stem derived from networkAreaID + host IP (matches backend buildOfflinePackageStem):
// bk-nodemgr-proxy-offline-{networkAreaID}-{ipSlug}; first inner IP only when multiple are comma-separated.
const offlinePkgStem = computed(() => {
  const raw = (props.data.bk_host_innerip || '').trim();
  const firstIp = raw.split(',')
    .map(s => s.trim())
    .find(Boolean) || '';
  const ipSlug = firstIp.replace(/[.:]/g, '-') || 'unknown';
  const areaId = props.data.bk_networkarea_id;
  const areaPart = typeof areaId === 'number' ? String(areaId) : 'unknown';
  return `bk-nodemgr-proxy-offline-${areaPart}-${ipSlug}`;
});
const offlineExtractCmd = computed(() => `tar xzf ${offlinePkgStem.value}.tar.gz`);
const offlineCdAndRunCmd = computed(() => `cd ${offlinePkgStem.value}\nbash install.sh`);

const currentSolution = computed(() => installSolutions.value.find(s => s.type === activeSolutionType.value));

const sidesliderTitle = computed(() => {
  const titleKey = props.data.is_offline
    ? 'platform.nodeMan.guide.offlineGuideTitle'
    : 'platform.nodeMan.installAgentPage.guideTitle';
  return `${t(titleKey)} ${props.data.bk_host_innerip || ''}`;
});

const handleClose = (val: boolean) => {
  emit('update:isShow', val);
};

watch(() => props.isShow, async (val) => {
  if (val && props.data.workflow_id) {
    if (!props.data.is_offline) {
      fetchManualSolutions();
    }
  }
}, { immediate: true });

const fetchManualSolutions = async () => {
  if (!props.data.workflow_id || !props.data.operation_id) return;

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
  resultData.value = '';
  submitSuccess.value = false;
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
    Message({ theme: 'success', message: t('platform.nodeMan.guide.copySuccess'), delay: 1500 });
  } finally {
    document.body.removeChild(textarea);
  }
};

const defaultOfflineTarName = 'offline-package.tar.gz';

// parseFilenameFromContentDisposition extracts a filename from Content-Disposition (RFC 5987 optional).
const parseFilenameFromContentDisposition = (header: string | null): string => {
  if (!header) return defaultOfflineTarName;
  const utf8 = header.match(/filename\*\s*=\s*UTF-8''([^;\s]+)/i);
  if (utf8?.[1]) {
    try {
      return decodeURIComponent(utf8[1].replace(/["']/g, ''));
    } catch {
      return utf8[1].replace(/["']/g, '');
    }
  }
  const quoted = header.match(/filename\s*=\s*"([^"]+)"/i);
  if (quoted?.[1]) return quoted[1];
  const plain = header.match(/filename\s*=\s*([^;\s]+)/i);
  if (plain?.[1]) return plain[1].replace(/^["']|["']$/g, '');
  return defaultOfflineTarName;
};

const handleDownloadPackage = async () => {
  if (!props.data.operation_id) return;
  isDownloading.value = true;
  try {
    const prefix = import.meta.env.BK_API_PREFIX ?? '';
    const url = `${prefix}${OFFLINE_PACKAGE_DOWNLOAD_PATH}`;
    const response = (await httpFetch(url, {
      method: 'POST',
      credentials: 'include',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ operation_id: String(props.data.operation_id) }),
      originalResponse: true,
      validateCode: false,
      interceptorErr: false,
      responseType: 'blod',
    } as Config)) as Response | undefined;

    if (!response?.ok) {
      Message({ theme: 'error', message: t('platform.nodeMan.guide.offline.downloadFailed'), delay: 3000 });
      return;
    }

    const blob = await response.blob();
    const filename = parseFilenameFromContentDisposition(response.headers.get('Content-Disposition'));
    const objectUrl = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = objectUrl;
    link.download = filename;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    URL.revokeObjectURL(objectUrl);
  } catch (error) {
    console.error('Download offline package error:', error);
    Message({ theme: 'error', message: t('platform.nodeMan.guide.offline.downloadFailed'), delay: 3000 });
  } finally {
    isDownloading.value = false;
  }
};

const handleSubmitResult = async () => {
  const raw = resultData.value.trim();
  if (!raw) return;

  isSubmitting.value = true;
  submitSuccess.value = false;
  try {
    await NodeWorkflowService.NodeWorkflowOperationOfflineInstallResultSubmit({
      operation_id: String(props.data.operation_id),
      result_data: raw,
    });
    submitSuccess.value = true;
    Message({ theme: 'success', message: t('platform.nodeMan.guide.offline.submitSuccess'), delay: 2000 });
  } catch (error) {
    console.error('Submit offline install result error:', error);
    Message({ theme: 'error', message: t('platform.nodeMan.guide.offline.submitFailed'), delay: 3000 });
  } finally {
    isSubmitting.value = false;
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
