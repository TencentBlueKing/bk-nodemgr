<template>
  <Upload
    ref="uploader"
    type="formdata"
    :tip="'支持 tgz、tar、gz 扩展名格式文件'"
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
            <span>上传成功</span>
          </div>
          <div
            class="text-[#FF5656] text-[12px]"
            v-else-if="curFile.status === 'failed'"
          >
            {{ curFile.message }}
          </div>
          <div
            class="text-[12px] flex items-center cursor-pointer"
            v-else-if="curFile.status === 'existed'"
          >
            <PopConfirm
              width="320"
              theme="light"
              :title="`存在同名${ capitalizeFirstLetter(currentType) }包，是否覆盖上传？`"
              confirmText="覆盖上传"
              cancelText="取消上传"
              @confirm="handleOverwrite"
              @cancel="handleCancel"
            >
              <div class="w-full">
                <i class="nodeman-icon nc-remind-fill text-[#F8B64F]"></i>
                <span class="text-[#F59500] ml-[7px]">存在同名 {{ capitalizeFirstLetter(currentType) }} 包</span>
              </div>
              <template #content>
                <div class="text-[12px] text-[#4D4F56] w-full mb-[5px]">
                  包名：{{ curFile.data?.name }}
                </div>
                <div class="text-[12px] text-[#4D4F56] w-full mb-[5px]">
                  MD5：{{ curFile.data?.md5 }}
                </div>
                <div class="text-[12px] text-[#4D4F56] w-full mb-[22px]">
                  继续上传，将会覆盖当前平台同名的 {{ capitalizeFirstLetter(currentType) }} 包
                </div>
              </template>
            </PopConfirm>
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
          >{{ bytesToMegabytes(file.size) }}M</span
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
  </Upload>
</template>
<script lang="ts" setup>
import { Button, Message, PopConfirm, Progress, Upload } from 'bkui-vue';
import { RightTurnLine } from 'bkui-vue/lib/icon';
import { computed, reactive, ref } from 'vue';
import { useRoute } from 'vue-router';

import { bytesToMegabytes, capitalizeFirstLetter } from '@/common/util';


const emit = defineEmits(['upload', 'cancel', 'loading']);
const route = useRoute();
const uploader = ref(null);
const currentType = computed(() => {
  const routeName = route.name?.toString() || '';
  const type = routeName.split('PackageMng')[0];
  return type;
});
const url = computed(() => {
  const type = currentType.value === 'proxy' ? 'server' : currentType.value;
  return `${location.origin}/api/v3/package/upload/origin/${type}`;
});
const curFile = reactive({
  file: null as File | null,
  progress: 0,
  status: '',
  data: null,
  message: '',
});

const handleBeforeUpload = (file: File, fileList: File[]) => {
  const whiteList = ['tgz', 'tar', 'gz'];
  const AllFiles = fileList.filter(v => whiteList.includes(v.name.substring(file.name.lastIndexOf('.') + 1)));
  if (AllFiles.length !== fileList.length) {
    fileList.pop();
    Message({
      theme: 'warning',
      message: '只允许上传tgz、tar、gz的文件',
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
  const metadata = JSON.stringify({ generation: 2, overwrite: true });
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
      curFile.status = response.code === 0
        ? (response.data.existed && !overwrite.value)
        ? 'existed'
        : 'success'
        : 'failed';
      curFile.data = response.data;
      Message({
        theme: 'success',
        message: '文件上传成功',
      });
      emit('upload', response.data);
    } else {
      curFile.message = xhr.statusText;
      curFile.status = 'failed';
      Message({
        theme: 'error',
        message: xhr.statusText,
      });
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
  uploader.value?.handleRemove(curFile);
};
</script>
