<template>
  <Button
    text
    theme="primary"
    :loading="downloadLoading"
    @click="handleDownload"
  >
    <slot></slot>
  </Button>
</template>
<script setup lang="ts">
import { Button } from 'bkui-vue';
import { debounce } from 'lodash';
import { computed, ref } from 'vue';

const props = defineProps({
  url: {
    type: String,
    required: true,
  },
  data: {
    type: Object,
    required: true,
  },
  currentType: {
    type: String,
    default: '',
  },
});


const downloadLoading = ref(false);

// 定义下载参数类型
interface DownloadParams {
  generation: number;
  name?: string;
  version?: string;
  platform?: {
    os_type: string;
    cpu_arch: string;
  };
}

// 下载
const handleDownload = async () => {
  try {
    downloadLoading.value = true;

    // 构建下载参数
    const params: DownloadParams = { generation: 2 };

    // 根据类型设置不同参数
    switch (props.currentType) {
      case 'plugin_bintool':
        if (props.data.name) params.name = props.data.name;
        break;
      case 'cert':
      case 'bintool':
        // 这两个类型不需要额外参数
        break;
      default:
        if (props.data.name) params.name = props.data.name;
        params.version = props.data.version;
        params.platform = {
          os_type: props.data.os_type,
          cpu_arch: props.data.cpu_arch,
        };
    }

    await debounceFetchDownloadFile(props.url, params, props.data.file_name);
  } catch (error) {
    console.error('下载失败:', error);
    // 可以在这里添加用户提示，比如使用 Message 组件
  } finally {
    downloadLoading.value = false;
  }
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
// 防抖
const debounceFetchDownloadFile = debounce(fetchDownloadFile, 300);
</script>
