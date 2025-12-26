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
    <template #tip>
      <div class="flex items-center gap-[3px]">
        <span>{{ '仅支持 .xlsx 类型文件，下载'}}</span>
        <a :href="downloadUrl" download="bk_nodeman_info.xlsx">
          <Button text theme="primary">
            {{ '模板文件' }}
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
import { useRoute } from 'vue-router';


const emit = defineEmits(['upload', 'cancel', 'loading']);
const uploader = ref(null);
const url = `${window.location.origin}/api/v3/node/agent/upload_template`;
const downloadUrl = `${window.location.origin}/api/v3/node/agent/download_template`;
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
      message: '仅支持 .xlsx 类型文件',
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
  xhr.open('POST', url, true);
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
      curFile.status = response.code === 0 ? 'success' : 'failed';
      curFile.data = response.data;
      emit('upload', response.data);
    } else {
      curFile.message = xhr.statusText;
      curFile.status = 'failed';
      emit('upload', null);
    }
  };

  // 错误处理
  xhr.onerror = () => {
    curFile.message = '上传失败';
    Message({
      theme: 'error',
      message: '上传失败',
    });
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
