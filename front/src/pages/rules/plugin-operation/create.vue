<template>
  <div class="create-plugin-operation">
    <!-- 步骤条 -->
    <div class="steps-wrapper px-[24px] py-[16px] bg-[#fff]">
      <div class="steps-bar flex items-center justify-center">
        <div
          v-for="(step, index) in stepList"
          :key="step.key"
          class="step-item flex items-center cursor-pointer"
          :class="{ 'mr-[0px]': index === stepList.length - 1 }"
          @click="handleStepClick(index)"
        >
          <div
            class="step-circle w-[24px] h-[24px] rounded-full flex items-center justify-center text-[12px] shrink-0"
            :class="getStepClass(index)"
          >
            <i v-if="index < currentStep" class="nodeman-icon nc-check-small text-[14px]"></i>
            <span v-else>{{ index + 1 }}</span>
          </div>
          <span
            class="step-label text-[14px] ml-[8px] whitespace-nowrap"
            :class="{
              'text-[#3A84FF] font-bold': index === currentStep,
              'text-[#313238]': index < currentStep,
              'text-[#979BA5]': index > currentStep,
            }"
          >{{ step.label }}</span>
          <div
            v-if="index < stepList.length - 1"
            class="step-line w-[80px] h-[1px] mx-[16px]"
            :class="index < currentStep ? 'bg-[#3A84FF]' : 'bg-[#DCDEE5]'"
          ></div>
        </div>
      </div>
    </div>

    <!-- 步骤内容 -->
    <div class="step-content px-[24px] pt-[16px] pb-[20px] overflow-auto" :style="{ height: contentHeight }">
      <!-- 步骤1: 部署目标 -->
      <StepDeployTarget
        v-show="currentStep === 0"
        ref="deployTargetRef"
        v-model:formData="formData"
        :initial-plugin-name="currentPluginName"
      />

      <!-- 步骤2: 参数配置 -->
      <StepParamConfig
        v-show="currentStep === 1"
        ref="paramConfigRef"
        v-model:formValues="formData.paramConfig"
        :plugin-name="formData.pluginName || ''"
        :platform-versions="deployTargetRef?.systemData || []"
      />

      <!-- 步骤3: 执行预览 -->
      <StepExecPreview
        v-show="currentStep === 2"
        :form-data="formData"
      />
    </div>

    <!-- 底部操作按钮 -->
    <div class="footer fixed bottom-0 left-0 right-0 h-[48px] bg-[#fff] border-t border-[#DCDEE5] flex items-center pl-[84px] pr-[24px] z-[100]">
      <Button
        v-if="currentStep === stepList.length - 1"
        class="mr-[8px] w-[88px]"
        theme="primary"
        @click="handleSubmit"
      >
        {{ $t('pluginOperation.preview.submitStrategy') }}
      </Button>
      <Button
        v-if="currentStep < stepList.length - 1"
        class="mr-[8px] w-[88px]"
        theme="primary"
        @click="handleNext"
      >
        {{ $t('pluginOperation.action.nextStep') }}
      </Button>
      <Button
        v-if="currentStep > 0"
        class="mr-[8px] w-[88px]"
        @click="handlePrev"
      >
        {{ $t('pluginOperation.action.prevStep') }}
      </Button>
      <Button class="w-[88px]" @click="handleBack">
        {{ $t('pluginOperation.action.cancel') }}
      </Button>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { Button, Message } from 'bkui-vue';
import { computed, onBeforeUnmount, onMounted, reactive, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';

import { PluginAPIService } from '@/api/modules/plugin';

import StepDeployTarget from './steps/step-deploy-target.vue';
import StepExecPreview from './steps/step-exec-preview.vue';
import StepParamConfig from './steps/step-param-config.vue';

import { useMainStore } from '@/stores/main';

const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const mainStore = useMainStore();

const currentStep = ref(0);
// 从路由 query 获取操作类型和插件名称
const operationType = ref((route.query.operationType as string) || '');
const currentPluginName = ref((route.query.pluginName as string) || '');

// 操作类型映射显示名称
const operationTypeLabel = computed(() => {
  const map: Record<string, string> = {
    install: t('pluginManagement.plugin.operate.install'),
    upgrade: t('pluginManagement.plugin.operate.upgrade'),
    reload: t('pluginManagement.plugin.operate.reload'),
  };
  return map[operationType.value] || t('pluginOperation.createStrategy');
});

// 头部标题：如 "Plugin Install" / "安装插件"
const headerTitle = computed(() => {
  if (operationType.value) {
    const pluginName = formData.pluginName || currentPluginName.value;
    return pluginName ? `${operationTypeLabel.value} ${pluginName}` : operationTypeLabel.value;
  }
  return t('pluginOperation.createStrategy');
});
const deployTargetRef = ref<InstanceType<typeof StepDeployTarget> | null>(null);
const paramConfigRef = ref<InstanceType<typeof StepParamConfig> | null>(null);

const contentHeight = computed(() => `${mainStore.windowInnerHeight - 160 - 48 - (mainStore.noticeShow ? 40 : 0)}px`);

const stepList = computed(() => [
  { key: 'deployTarget', label: t('pluginOperation.steps.deployTarget') },
  { key: 'paramConfig', label: t('pluginOperation.steps.paramConfig') },
  { key: 'execPreview', label: t('pluginOperation.steps.execPreview') },
]);

const formData = reactive({
  strategyName: '', // 保留字段兼容 model 类型
  pluginName: currentPluginName.value,
  selectedHosts: [] as any[],
  selectedVersion: '',
  platform: '', // 用户选择的操作系统平台，如 'linux_x86_64'
  paramConfig: {} as Record<string, any>,
});

const getStepClass = (index: number) => {
  if (index < currentStep.value) {
    return 'bg-[#3A84FF] text-[#fff]';
  }
  if (index === currentStep.value) {
    return 'bg-[#3A84FF] text-[#fff]';
  }
  return 'bg-[#F0F1F5] text-[#979BA5]';
};

const handleStepClick = (index: number) => {
  // 只允许点击已完成的步骤或当前步骤的下一步
  if (index <= currentStep.value) {
    currentStep.value = index;
  }
};

const handleNext = async () => {
  if (currentStep.value === 0) {
    // 校验第一步
    const valid = await deployTargetRef.value?.validate();
    if (!valid) return;
  }
  if (currentStep.value < stepList.value.length - 1) {
    currentStep.value += 1;
    // 进入参数配置步骤时，加载各平台的配置变量
    if (currentStep.value === 1) {
      paramConfigRef.value?.load();
    }
  }
};

const handlePrev = () => {
  if (currentStep.value > 0) {
    currentStep.value -= 1;
  }
};

const handleSubmit = async () => {
  try {
    const hosts = formData.selectedHosts;
    if (!formData.pluginName?.trim()) {
      Message({ theme: 'warning', message: t('platform.nodeMan.pluginOperation.form.pluginNameRequired') });
      return;
    }
    if (!hosts.length) {
      Message({ theme: 'warning', message: t('common.selectedCount', { count: 0 }) });
      return;
    }

    // 构造插件参数
    const pluginName = formData.pluginName?.trim() || currentPluginName.value;
    const pluginPayload = hosts.map((host: any) => ({
      bk_host_id: host.bk_host_id || host.host_id,
      plugin_name: pluginName,
      version: formData.selectedVersion || '',
      config_name: [pluginName],
      custom_config_context: formData.paramConfig || {},
    }));

    let res: { workflow_id: string };
    // 根据操作类型调用不同 API
    if (operationType.value === 'upgrade') {
      res = await PluginAPIService.UpgradePlugin({ plugin: pluginPayload });
    } else {
      res = await PluginAPIService.InstallPlugin({ plugin: pluginPayload });
    }

    // 安装/升级成功 → 跳转到任务详情
    if (res?.workflow_id) {
      Message({ theme: 'success', message: t('pluginManagement.plugin.operate.operateSuccess') });
      router.push({
        name: 'taskDetail',
        params: { taskId: res.workflow_id },
        query: { active: 'plugin' },
      });
    }
  } catch (e) {
    console.error('Plugin operation failed:', e);
  }
};

const handleBack = () => {
  if (operationType.value) {
    // 从插件列表页跳转来的，返回插件列表
    router.push({ name: 'plugin' });
  } else {
    router.push({ name: 'pluginStrategy' });
  }
};

// 动态修改路由 meta.title，让 PageHeader 显示正确的操作标题
const originalTitle = route.meta.title;
onMounted(() => {
  if (operationType.value) {
    // 直接设置为翻译后的文本，$t 找不到 key 时会 fallback 到原文
    route.meta.title = headerTitle.value;
  }
});
onBeforeUnmount(() => {
  // 离开页面时恢复原始 title
  route.meta.title = originalTitle;
});
</script>

<style lang="postcss" scoped>
.create-plugin-operation {
  background: #f5f7fa;
}
.step-content {
  background: #f5f7fa;
}
</style>
