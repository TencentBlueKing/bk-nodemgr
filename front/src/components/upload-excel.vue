<template>
  <Upload
    class="p-[24px] min-h-[220px]"
    ref="uploader"
    type="formdata"
    :url="url"
    :size="1000"
    :multiple="false"
    :limit="1"
    with-credentials
    :before-upload="handleBeforeUpload"
    :custom-request="handleUpload"
  >
    <template #file="{ file }">
      <div class="relative flex w-full">
        <div class="w-[32px] h-full">
          <i class="nodeman-icon nc-package-2 text-[32px]"></i>
        </div>
        <div class="ml-[10px] flex-1">
          <div class="text-[12px]">{{ file.name }}</div>
          <div
            class="text-[12px] text-[#2DCB56] flex items-center"
            v-if="curFile.status === 'success'"
          >
            <i class="nodeman-icon nc-check-small text-[22px]"></i>
            <span>{{ t('pkgUpload.uploadSuccess') }}</span>
          </div>
          <div
            class="text-[#FF5656] text-[12px]"
            v-else-if="curFile.status === 'failed'"
          >
            {{ curFile.message }}
          </div>
          <div v-else>
            <Progress :percent="curFile.progress" :size="'small'" />
          </div>
        </div>
        <div
          class="absolute right-[12px] h-full flex items-center"
          v-if="curFile.status"
        >
          <span
            class="text-[12px]"
            v-if="['success', 'existed'].includes(curFile.status)"
          >{{ bytesToMegabytes(file.size) }}</span
          >
          <template v-if="curFile.status === 'failed'">
            <Button @click="handleOverwrite" text>
              <right-turn-line fill="#979BA5" />
            </Button>
            <Button @click="handleDelete" text>
              <i class="nodeman-icon nc-delete-3 text-[#979BA5] ml-[15px]"></i>
            </Button>
          </template>
        </div>
      </div>
    </template>
    <template #tip>
      <div class="flex items-center gap-[3px]">
        <span>{{ $t('components.uploadExcel.tip') }}</span>
        <a :href="downloadUrl" download="bk_nodemgr_info.xlsx">
          <Button text theme="primary">
            {{ $t('components.uploadExcel.templateFile') }}
          </Button>
        </a>
      </div>
    </template>
  </Upload>
</template>
<script lang="ts" setup>
import { Button, Message, PopConfirm, Progress, Upload } from 'bkui-vue';
import { RightTurnLine } from 'bkui-vue/lib/icon';
import { computed, reactive, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute } from 'vue-router';
import { bytesToMegabytes } from '@/common/util';

const { t } = useI18n();
const emit = defineEmits(['upload', 'cancel', 'loading']);
const props = withDefaults(defineProps<{
  type?: 'agent' | 'proxy';
}>(), {
  type: 'agent',
});
const uploader = ref(null);
const url = computed(() => `${window.location.origin}/api/v3/node/${props.type}/install_template/upload`);
const downloadUrl = computed(() => `${window.location.origin}/api/v3/node/${props.type}/install_template/download`);
const curFile = reactive({
  file: null as File | null,
  progress: 0,
  status: '',
  data: null,
  message: '',
});

const handleBeforeUpload = (file: File, fileList: File[]) => {
  const whiteList = ['xlsx'];
  const AllFiles = fileList.filter(v => whiteList.includes(v.name.substring(file.name.lastIndexOf('.') + 1)));
  if (AllFiles.length !== fileList.length) {
    fileList.pop();
    Message({
      theme: 'warning',
      message: t('components.uploadExcel.tip'),
    });
    return false;
  }
  curFile.file = file;
  return true;
};
const overwrite = ref(false);
const handleUpload = () => {
  const formData = new FormData();

  // 添加元数据
  const metadata = JSON.stringify({ key: '', value: '' });
  formData.append('metadata', metadata);
  if (overwrite.value) {
    curFile.status = '';
    curFile.progress = 0;
    curFile.data = null;
  }
  formData.append('file', curFile.file as File);
  formData.append('filename', curFile.file?.name as string);

  const xhr = new XMLHttpRequest();
  xhr.open('POST', url.value, true);
  xhr.setRequestHeader('X-Bk-Tenant-Id', 'single');

  // 上传进度事件监听器
  xhr.upload.onprogress = (event) => {
    if (event.lengthComputable) {
      const progress = (event.loaded / event.total) * 100;
      curFile.progress = progress;
      if (progress === 100) {
        emit('loading', true);
      }
    }
  };

  xhr.onload = () => {
    emit('loading', false);
    if (xhr.status >= 200 && xhr.status < 300) {
      const response = JSON.parse(xhr.responseText);
      if (response.code === 0) {
        curFile.status = 'success';
        curFile.data = response.data;
        emit('upload', response.data);
      } else {
        curFile.status = 'failed';
        const showMessageData = response.message || response.error?.message || '';
        curFile.message = showMessageData;
        Message({
          theme: 'error',
          message: {
            code: response.code ?? response.status,
            overview: showMessageData,
            suggestion: '',
            type: 'key-value',
          },
        });
        emit('upload', null);
      }
    } else {
      curFile.message = xhr.statusText;
      curFile.status = 'failed';
      const parsed = (() => { try { return JSON.parse(xhr.responseText); } catch { return {}; } })();
      const showMessageData = parsed?.message || parsed?.error?.message || xhr.statusText || '';
      Message({
        theme: 'error',
        message: {
          code: xhr.status,
          overview: showMessageData,
          suggestion: '',
          type: 'key-value',
          details: {
            code: parsed?.code,
            message: parsed?.error?.details?.[0]?.message ?? showMessageData,
            url: xhr.responseURL,
          },
        },
      });
      emit('upload', null);
    }
  };

  // 错误处理
  xhr.onerror = () => {
    curFile.message = t('components.uploadExcel.error');
    Message({
      theme: 'error',
      message: {
        code: 0,
        overview: t('components.uploadExcel.error'),
        suggestion: '',
        type: 'key-value',
      },
    });
    emit('upload', null);
  };

  xhr.send(formData);
};
const handleOverwrite = () => {
  overwrite.value = true;
  handleUpload();
};
const handleCancel = () => {
  emit('cancel');
};
const handleDelete = () => {
  curFile.data = null;
  uploader.value?.handleRemove(curFile);
};

defineExpose({
  curFile,
  handleDelete,
});
</script>
