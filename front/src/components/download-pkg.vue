<template>
  <Button
    text
    theme="primary"
    :loading="downloadLoading"
    @click="handleDownload"
  >下载</Button>
</template>
<script setup lang="ts">
import { Button } from 'bkui-vue';
import { ref } from 'vue';

const props = defineProps({
  url: {
    type: String,
    required: true,
  },
  data: {
    type: Object,
    required: true,
  },
});


const downloadLoading = ref(false);
// 下载
const handleDownload = async () => {
  downloadLoading.value = true;
  const params = {
    generation: 2,
    version: props.data.version,
    platform: {
      os_type: props.data.os_type,
      cpu_arch: props.data.cpu_arch,
    },
  };
  await fetchDownloadFile(props.url, params, props.data.file_name);
  downloadLoading.value = false;
};

/**
 * 使用 fetch 下载文件
 * @param url 请求URL
 * @param params 请求参数
 * @param filename 下载的文件名
 */
async function fetchDownloadFile(reqUrl: string, params: any, filename: string) {
  try {
    const response = await fetch(reqUrl, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(params),
    });

    if (!response.ok) {
      throw new Error(`下载失败: ${response.status}`);
    }

    const blob = await response.blob();
    const url = window.URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.setAttribute('download', filename);
    document.body.appendChild(link);
    link.click();

    // 清理
    document.body.removeChild(link);
    window.URL.revokeObjectURL(url);
  } catch (error) {
    console.error('下载失败:', error);
    throw error;
  }
}
</script>
