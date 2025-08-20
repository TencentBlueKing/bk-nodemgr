<template>
  <Sideslider
    v-model:isShow="isShow"
    :width="960"
    :title="'包上传'"
    render-directive="if"
    :before-close="handleBeforeClose"
  >
    <template #default>
      <div class="px-[24px] pt-[28px]">
        <pkg-upload @upload="handleUpload" class="mb-[24px]"></pkg-upload>
        <upload-result-table :data="uploadData"></upload-result-table>
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
import { watch, ref, onMounted, computed } from "vue";
import {
  Sideslider,
  InfoBox,
  Button,
} from "bkui-vue";
import PkgUpload from "./pkg-upload.vue";
import UploadResultTable from "./upload-result-table.vue";
import type { PackageUploadOriginAgentRespData } from "@/@types/pkg";
import { usePackageStore } from '@/stores/package';
import { PackageService } from "@/api/modules/pkg";
import { useRoute } from "vue-router";

const route = useRoute();
const isShow = defineModel("isShow", { type: Boolean });
const emit = defineEmits('confirm')
const hasPkg = computed(() => !!uploadData.value);
const uploadData = ref<PackageUploadOriginAgentRespData | null>(null);
const packageStore = usePackageStore();

const handleBeforeClose = () =>
  new Promise((resolve, reject) => {
    InfoBox({
      title: "确认关闭?",
      infoType: "warning",
      onConfirm: () => {
        resolve(true);
        isShow.value = false;
      },
      onCancel: () => reject(),
    });
  });
const handleUpload = (data: PackageUploadOriginAgentRespData) => {
  uploadData.value = {...data};
};
const loading = ref(false);
const submit = async () => {
  loading.value = true;
  if (route.name === "agentPackageMng") {
    await PackageService.PublishReleaseAgent({
      upload_id: uploadData.value?.upload_id
    });
  } else {
    await PackageService.PublishReleaseProxy({
      upload_id: uploadData.value?.upload_id
    });
  }
  loading.value = false;
  isShow.value = false;
  emit('confirm');
}
watch(
  () => isShow.value,
  () => {
    if (!isShow.value) {
    }
  },
  { immediate: true }
);
onMounted(async () => {
  await packageStore.getPackages();
});
watch(() => isShow.value, () => {
  if(isShow.value) {
    uploadData.value = null;
  }
},{immediate: true})
</script>
