<template>
  <Sideslider
    v-model:is-show="isShow"
    :width="960"
    :title="'包上传'"
    render-directive="if"
    :before-close="handleBeforeClose"
  >
    <template #header>
      <div class="flex items-center justify-between w-full">
        <span>包上传</span>
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
                  @click="triggerHandler(item.id)"
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
          :plugin-type="pluginUploadType"
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
          content: '请先上传包文件',
          disabled: hasPkg,
        }"
        class="mr-[8px]"
        @click="submit"
        :loading="loading"
      >提交</Button
      >
      <Button @click="handleBeforeClose">取消</Button>
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
import { useRoute } from 'vue-router';

import PkgUpload from './pkg-upload.vue';
import UploadResultTable from './upload-result-table.vue';

import type { PackageUploadOriginAgentRespData } from '@/@types/pkg';
import { PackageService } from '@/api/modules/pkg';
import { usePackageStore } from '@/stores/package';

const isShow = defineModel('isShow', { type: Boolean });
const emit = defineEmits('confirm');
const route = useRoute();
const hasPkg = computed(() => !!uploadData.value);
const uploadData = ref<PackageUploadOriginAgentRespData | null>(null);
const packageStore = usePackageStore();

// 插件上传类型
const pluginUploadType = ref('v3/plugin');
const pluginUploadTypeList = ref([
  {
    id: 'v2/plugin',
    name: '2.0 官方插件包',
  },
  {
    id: 'v2/external_plugin',
    name: '2.0 业务插件包',
  },
]);
const triggerHandler = (id: string) => {
  isShow.value = true;
  pluginUploadType.value = id;
};

const handleBeforeClose = () => new Promise((resolve, reject) => {
  InfoBox({
    title: '确认关闭?',
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
    const serviceMap: Record<string, (args: { upload_id: string }) => Promise<void>> = {
      agentPackageMng: PackageService.PublishReleaseAgent,
      proxyPackageMng: PackageService.PublishReleaseProxy,
      certPackageMng: PackageService.PublishReleaseCert,
      bintoolPackageMng: PackageService.PublishReleaseBinTool,
      'pluginPackageMng_v2/plugin': PackageService.PublishReleasePluginV2,
      'pluginPackageMng_v2/external_plugin': PackageService.PublishReleaseExternalPluginV2,
      'pluginPackageMng_v3/plugin': PackageService.PublishReleasePluginV3,
    };

    const key = route.name === 'pluginPackageMng' ? `pluginPackageMng_${pluginUploadType.value}` : route.name;
    // 获取映射中的服务方法
    const serviceMethod = route.name ? serviceMap[key] : undefined;

    // 如果有对应的服务方法，调用它
    if (serviceMethod && uploadData.value?.upload_id) {
      await serviceMethod({ upload_id: uploadData.value.upload_id });
    }
    Message({
      theme: 'success',
      message: '发布成功',
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
  }
}, { immediate: true });
</script>
<style lang="postcss" scoped>
.active {
  color: #3a84ff !important;
  background-color: #eaf3ff;
}
</style>
