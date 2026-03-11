<template>
  <Sideslider
    v-model:is-show="isShow"
    :width="960"
    :title="t('pkgUpload.title')"
    render-directive="if"
    :before-close="handleBeforeClose"
  >
    <template #header>
      <div class="flex items-center justify-between w-full">
        <span>{{ t('pkgUpload.title') }}<span v-if="subTitle" class="text-[14px] ml-[10px]">{{ subTitle }}</span></span>
        <template v-if="route.name === 'pluginPackageMng'">
          <Dropdown
            theme="light"
            trigger="click"
            placement="bottom-start"
            :popover-options="{
              clickContentAutoHide: true,
            }"
          >
            <Button
              class="mr-[24px]"
              text
            >
              <i class="nodeman-icon nc-setting"></i>
            </Button>
            <template #content>
              <Dropdown.DropdownMenu ext-cls="dropDown-menu">
                <Dropdown.DropdownItem
                  :class="['text-14px', { 'active': pluginUploadType === item.id }]"
                  v-for="item in pluginUploadTypeList"
                  :key="item.id"
                  v-bk-tooltips="{ content: t(item.tipKey), placement: 'right' }"
                  @click="triggerHandler(item.id)"
                >
                  {{ item.name }}
                </Dropdown.DropdownItem>
              </Dropdown.DropdownMenu>
            </template>
          </Dropdown>
        </template>
        <template v-if="route.name === 'proxyPackageMng'">
          <Dropdown
            theme="light"
            trigger="click"
            placement="bottom-start"
            :popover-options="{
              clickContentAutoHide: true,
            }"
          >
            <Button
              class="mr-[24px]"
              text
            >
              <i class="nodeman-icon nc-setting"></i>
            </Button>
            <template #content>
              <Dropdown.DropdownMenu ext-cls="dropDown-menu">
                <Dropdown.DropdownItem
                  :class="['text-14px', { 'active': proxyUploadType === item.id }]"
                  v-for="item in proxyUploadTypeList"
                  :key="item.id"
                  @click="proxyTriggerHandler(item.id)"
                >
                  {{ item.name }}
                </Dropdown.DropdownItem>
              </Dropdown.DropdownMenu>
            </template>
          </Dropdown>
        </template>
      </div>
    </template>
    <template #default>
      <div class="px-[24px] pt-[28px]">
        <pkg-upload
          :key="`${pluginUploadType}_${proxyUploadType}`"
          :plugin-type="pluginUploadType"
          :proxy-type="proxyUploadType"
          @upload="handleUpload"
          @cancel="handleCancel"
          @loading="handleLoading"
          class="mb-[24px]">
        </pkg-upload>
        <upload-result-table
          :data="uploadData"
          :loading="parseLoading">
        </upload-result-table>
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
  Dropdown,
  InfoBox,
  Message,
  Sideslider,
} from 'bkui-vue';
import { computed, onMounted, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute } from 'vue-router';

import PkgUpload from './pkg-upload.vue';
import UploadResultTable from './upload-result-table.vue';

import type { PackageUploadOriginAgentRespData } from '@/@types/pkg';
import { PackageService } from '@/api/modules/pkg';
import { usePackageStore } from '@/stores/package';

const { t } = useI18n();
const isShow = defineModel('isShow', { type: Boolean });
const emit = defineEmits('confirm');
const route = useRoute();
const hasPkg = computed(() => !!uploadData.value);
const uploadData = ref<PackageUploadOriginAgentRespData | null>(null);
const packageStore = usePackageStore();

// 插件上传类型
const pluginUploadType = ref('v3/plugin');
// Proxy上传来源包类型
const proxyUploadType = ref<string>('origin_proxy');
const subTitle = ref('');
const pluginUploadTypeList = computed(() => [
  {
    id: 'v2/plugin',
    name: t('pkgUpload.officialPlugin'),
    tipKey: 'pkgUpload.officialPluginTip',
  },
  {
    id: 'v2/external_plugin',
    name: t('pkgUpload.externalPlugin'),
    tipKey: 'pkgUpload.externalPluginTip',
  },
]);
const proxyUploadTypeList = computed(() => [
  {
    id: 'origin_proxy',
    name: t('pkgUpload.proxyPkgTypeOriginProxy'),
  },
  {
    id: 'origin_server',
    name: t('pkgUpload.proxyPkgTypeOriginServer'),
  },
]);
const triggerHandler = (id: string) => {
  const doSwitch = () => {
    if (pluginUploadType.value === id) {
      pluginUploadType.value = 'v3/plugin';
      subTitle.value = '';
    } else {
      pluginUploadType.value = id;
      subTitle.value = pluginUploadTypeList.value.find(item => item.id === id)?.name || '';
    }
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
  const doSwitch = () => {
    proxyUploadType.value = id;
    subTitle.value = proxyUploadTypeList.value.find(item => item.id === id)?.name || '';
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

const handleBeforeClose = () => new Promise((resolve, reject) => {
  // 没有上传数据，直接关闭
  if (!uploadData.value) {
    resolve(true);
    isShow.value = false;
    return;
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
      const params: Record<string, string> = { upload_id: uploadData.value.upload_id };
      // Proxy发布需要额外传 upload_origin_pkg_type
      if (route.name === 'proxyPackageMng') {
        params.upload_origin_pkg_type = proxyUploadType.value;
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

watch(() => isShow.value, async () => {
  if (isShow.value) {
    // await packageStore.getPackages();
    uploadData.value = null;
    proxyUploadType.value = 'origin_proxy';
    subTitle.value = route.name === 'proxyPackageMng' ? t('pkgUpload.proxyPkgTypeOriginProxy') : '';
  }
}, { immediate: true });
</script>
<style lang="postcss" scoped>
.active {
  color: #3a84ff !important;
  background-color: #eaf3ff;
}
</style>
