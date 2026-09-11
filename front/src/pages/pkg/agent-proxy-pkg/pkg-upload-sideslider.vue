<template>
  <Sideslider
    v-model:is-show="isShow"
    :width="960"
    :title="t('pkgUpload.title')"
    render-directive="if"
    :before-close="handleBeforeClose"
  >
    <template #default>
      <div class="px-[24px] pt-[28px] pb-[24px]">
        <template v-if="showUploadTypeTabs">
          <div class="upload-type-tabs" role="tablist" :aria-label="t('pkgUpload.title')">
            <button
              v-for="item in uploadTypeTabs"
              :key="item.id"
              type="button"
              role="tab"
              :aria-selected="activeUploadType === item.id"
              :tabindex="activeUploadType === item.id ? 0 : -1"
              :class="['upload-type-tab', { 'active': activeUploadType === item.id }]"
              v-bk-tooltips="{
                content: item.tipKey ? t(item.tipKey) : '',
                placement: 'top',
                disabled: !item.tipKey,
                theme: 'light',
              }"
              @click="handleUploadTypeChange(item.id)"
            >
              {{ item.name }}
            </button>
          </div>
          <div class="upload-type-panel">
            <pkg-upload
              :key="`${pluginUploadType}_${proxyUploadType}`"
              :plugin-type="pluginUploadType"
              :proxy-type="proxyUploadType"
              @upload="handleUpload"
              @cancel="handleCancel"
              @loading="handleLoading"
              class="mb-[24px]">
            </pkg-upload>
            <div
              v-if="showSharedSwitch"
              class="flex items-center my-[16px]">
              <span class="text-[14px] text-[#313238] mr-[8px]">{{ t('pkgUpload.isShared') }}</span>
              <Switcher v-model="isShared" theme="primary" />
              <span class="text-[12px] text-[#979BA5] ml-[8px]">{{ t('pkgUpload.isSharedTip') }}</span>
            </div>
            <upload-result-table
              :data="uploadData"
              :loading="parseLoading"
              :plugin-upload-type="pluginUploadType">
            </upload-result-table>
          </div>
        </template>
        <template v-else>
          <pkg-upload
            :key="`${pluginUploadType}_${proxyUploadType}`"
            :plugin-type="pluginUploadType"
            :proxy-type="proxyUploadType"
            @upload="handleUpload"
            @cancel="handleCancel"
            @loading="handleLoading"
            class="mb-[24px]">
          </pkg-upload>
          <div
            v-if="showSharedSwitch"
            class="flex items-center my-[16px]">
            <span class="text-[14px] text-[#313238] mr-[8px]">{{ t('pkgUpload.isShared') }}</span>
            <Switcher v-model="isShared" theme="primary" />
            <span class="text-[12px] text-[#979BA5] ml-[8px]">{{ t('pkgUpload.isSharedTip') }}</span>
          </div>
          <upload-result-table
            :data="uploadData"
            :loading="parseLoading"
            :plugin-upload-type="pluginUploadType">
          </upload-result-table>
        </template>
      </div>
    </template>
    <template #footer>
      <Button
        theme="primary"
        :disabled="!hasPkg"
        v-bk-tooltips="{
          content: t('pkgUpload.uploadTip'),
          disabled: hasPkg,
        }"
        class="mr-[8px]"
        @click="submit"
        :loading="loading"
      >{{ t('pkgUpload.submit') }}</Button
      >
      <Button @click="handleBeforeClose">{{ t('pkgUpload.cancel') }}</Button>
    </template>
  </Sideslider>
</template>
<script lang="ts" setup>
import {
  Button,
  InfoBox,
  Message,
  Sideslider,
  Switcher,
} from 'bkui-vue';
import { computed, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute } from 'vue-router';

import PkgUpload from './pkg-upload.vue';
import UploadResultTable from './upload-result-table.vue';

import type { PackageUploadOriginAgentRespData } from '@/@types/pkg';
import { PackageService } from '@/api/modules/pkg';
type UploadTab = {
  id: string;
  name: string;
  tipKey: string;
};

const isShow = defineModel('isShow', { type: Boolean });
const emit = defineEmits('confirm');
const { t } = useI18n();
const route = useRoute();
const hasPkg = computed(() => !!uploadData.value);
const uploadData = ref<PackageUploadOriginAgentRespData | null>(null);
// 全租户共享开关：多租户模式（BK_TENANT_MODE === 'multiple'）下仅 system 租户提供
// （cert 包除外，始终仅本租户），默认 true；单租户/非 system 租户不展示开关、请求不传 is_shared
const isShared = ref(true);
const isMultipleTenant = window.PROJECT_CONFIG.BK_TENANT_MODE === 'multiple';
const isSystemTenant = computed(() => window.PROJECT_CONFIG.BK_TENANT === 'system');
const isCertPackageRoute = computed(() => route.name === 'certPackageMng');
const showSharedSwitch = computed(() => hasPkg.value && isMultipleTenant && isSystemTenant.value && !isCertPackageRoute.value);

// 插件上传类型
const pluginUploadType = ref('v3/plugin');
// Proxy上传来源包类型
const proxyUploadType = ref<string>('origin_proxy');
const isPluginPackageRoute = computed(() => route.name === 'pluginPackageMng');
const isProxyPackageRoute = computed(() => route.name === 'proxyPackageMng');
const showUploadTypeTabs = computed(() => isPluginPackageRoute.value || isProxyPackageRoute.value);

const pluginUploadTypeList = computed<UploadTab[]>(() => [
  {
    id: 'v3/plugin',
    name: t('pkgUpload.officialPluginV3'),
    tipKey: '',
  },
  {
    id: 'v2/plugin',
    name: t('pkgUpload.officialPluginV2'),
    tipKey: 'pkgUpload.officialPluginV2Tip',
  },
  {
    id: 'v2/external_plugin',
    name: t('pkgUpload.externalPlugin'),
    tipKey: 'pkgUpload.externalPluginV2Tip',
  },
]);
const proxyUploadTypeList = computed<UploadTab[]>(() => [
  {
    id: 'origin_proxy',
    name: t('pkgUpload.proxyPkgTypeOriginProxy'),
    tipKey: 'pkgUpload.proxyPkgTypeOriginProxyTip',
  },
  {
    id: 'origin_server',
    name: t('pkgUpload.proxyPkgTypeOriginServer'),
    tipKey: 'pkgUpload.proxyPkgTypeOriginServerTip',
  },
]);
const uploadTypeTabs = computed<UploadTab[]>(() => {
  if (isPluginPackageRoute.value) return pluginUploadTypeList.value;
  if (isProxyPackageRoute.value) return proxyUploadTypeList.value;
  return [];
});
const activeUploadType = computed(() => (isPluginPackageRoute.value ? pluginUploadType.value : proxyUploadType.value));

const triggerHandler = (id: string) => {
  if (pluginUploadType.value === id) return;
  const doSwitch = () => {
    pluginUploadType.value = id;
    uploadData.value = null;
    parseLoading.value = false;
  };
  if (uploadData.value) {
    InfoBox({
      title: t('dialog.confirmSwitchType'),
      infoType: 'warning',
      onConfirm: () => doSwitch(),
    });
    return;
  }
  doSwitch();
};

const proxyTriggerHandler = (id: string) => {
  if (proxyUploadType.value === id) return;
  const doSwitch = () => {
    proxyUploadType.value = id;
    uploadData.value = null;
    parseLoading.value = false;
  };
  if (uploadData.value) {
    InfoBox({
      title: t('dialog.confirmSwitchType'),
      infoType: 'warning',
      onConfirm: () => doSwitch(),
    });
    return;
  }
  doSwitch();
};
const handleUploadTypeChange = (id: string) => {
  if (isPluginPackageRoute.value) {
    triggerHandler(id);
    return;
  }
  proxyTriggerHandler(id);
};

const handleBeforeClose = (): Promise<boolean> => new Promise((resolve, reject) => {
  // 未上传包：直接关闭，不弹确认（避免空操作骚扰）
  if (!hasPkg.value) {
    isShow.value = false;
    return resolve(true);
  }
  InfoBox({
    title: t('dialog.confirmClose'),
    infoType: 'warning',
    onConfirm: () => {
      resolve(true);
      isShow.value = false;
    },
    onCancel: () => reject(),
  });
});
const handleUpload = (data: PackageUploadOriginAgentRespData) => {
  uploadData.value = { ...data };
};
const handleCancel = () => {
  isShow.value = false;
};
// 加载到100后到解析完成的时间
const parseLoading = ref(false);
const handleLoading = (val: boolean) => {
  parseLoading.value = val;
};
const loading = ref(false);
const submit = async () => {
  try {
    loading.value = true;

    // 创建一个从路由名称到服务方法的映射
    const serviceMap: Record<string, (args: Record<string, string>) => Promise<void>> = {
      agentPackageMng: PackageService.PublishReleaseAgent,
      proxyPackageMng: PackageService.PublishReleaseProxy,
      certPackageMng: PackageService.PublishReleaseCert,
      bintoolPackageMng: PackageService.PublishReleaseBinTool,
      'pluginPackageMng_v2/plugin': PackageService.PublishReleasePluginV2,
      'pluginPackageMng_v2/external_plugin': PackageService.PublishReleaseExternalPluginV2,
      'pluginPackageMng_v3/plugin': PackageService.PublishReleasePluginV3,
      plugin_bintoolPackageMng: PackageService.PublishReleasePluginBinTool,
    };

    const key = route.name === 'pluginPackageMng' ? `pluginPackageMng_${pluginUploadType.value}` : route.name as string;
    // 获取映射中的服务方法
    const serviceMethod = route.name ? serviceMap[key] : undefined;

    // 如果有对应的服务方法，调用它
    if (serviceMethod && uploadData.value?.upload_id) {
      const params: Record<string, any> = { upload_id: uploadData.value.upload_id };
      // Proxy发布需要额外传 upload_origin_pkg_type
      if (route.name === 'proxyPackageMng') {
        params.upload_origin_pkg_type = proxyUploadType.value;
      }
      // 仅多租户模式下 system 租户（且非 cert 包）传 is_shared；其余情况不传，后端按默认处理
      if (isMultipleTenant && isSystemTenant.value && !isCertPackageRoute.value) {
        params.is_shared = isShared.value;
      }
      await serviceMethod(params);
    }
    Message({
      theme: 'success',
      message: t('pkgUpload.success'),
    });
    isShow.value = false;
    emit('confirm');
  } catch (error) {
    console.error('Failed to submit:', error);
  } finally {
    loading.value = false; // 确保在任何情况下都能执行
  }
};

watch(() => isShow.value, () => {
  if (isShow.value) {
    uploadData.value = null;
    proxyUploadType.value = 'origin_proxy';
    pluginUploadType.value = 'v3/plugin';
    isShared.value = true;
  }
}, { immediate: true });
</script>
<style lang="postcss" scoped>
.upload-type-tabs {
  display: flex;
  gap: 0;
  /* Tab 行与下方面板之间无间隙 */
  position: relative;
  z-index: 1;
}

.upload-type-tab {
  appearance: none;
  padding: 10px 20px;
  border: 1px solid #dcdee5;
  border-bottom: none;
  border-radius: 4px 4px 0 0;
  background: #f5f7fa;
  color: #63656e;
  cursor: pointer;
  font-size: 14px;
  line-height: 1;
  margin-right: 4px;
  transition: color 0.15s, background 0.15s;

  &:focus-visible {
    outline: 2px solid #3a84ff;
    outline-offset: 2px;
  }

  &:hover {
    color: #3a84ff;
    background: #fff;
  }

  &.active {
    background: #fff;
    color: #3a84ff;
    border-color: #dcdee5;
    /* 底边用白色盖住面板顶部边框，形成连通效果 */
    border-bottom: 1px solid #fff;
    margin-bottom: -1px;
  }
}

.upload-type-panel {
  border: 1px solid #dcdee5;
  border-radius: 0 4px 4px 4px;
  padding: 24px;
  background: #fff;
}
</style>
