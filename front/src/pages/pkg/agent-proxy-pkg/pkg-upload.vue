<template>
  <Upload
    ref="uploader"
    type="formdata"
    :tip="t('pkgUpload.fileFormatTip')"
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
        <div class="ml-[10px] flex-1 min-w-0">
          <div class="text-[12px] break-all">{{ file.name }}</div>
          <div
            class="text-[12px] text-[#2DCB56] flex items-center"
            v-if="curFile.status === 'success'"
          >
            <i class="nodeman-icon nc-check-small text-[22px]"></i>
            <span>{{ t('pkgUpload.uploadSuccess') }}</span>
          </div>
          <div
            class="pkg-upload-error"
            v-else-if="curFile.status === 'failed'"
          >
            {{ curFile.message
              ? t('pkgUpload.uploadFailedDetail', { message: curFile.message })
              : t('pkgUpload.uploadFailed') }}
          </div>
          <div
            class="text-[12px] flex items-center cursor-pointer"
            v-else-if="curFile.status === 'existed'"
          >
            <PopConfirm
              width="320"
              theme="light"
              :title="t('pkgUpload.overwriteConfirm', { type: capitalizeFirstLetter(currentType) })"
              :confirm-text="t('pkgUpload.overwrite')"
              :cancel-text="t('pkgUpload.cancelUpload')"
              @confirm="handleOverwrite"
              @cancel="handleCancel"
            >
              <div class="w-full">
                <i class="nodeman-icon nc-remind-fill text-[#F8B64F]"></i>
                <span class="text-[#F59500] ml-[7px]">
                  {{ $t('pkgUpload.targetExists', { type: capitalizeFirstLetter(currentType) }) }}
                </span>
              </div>
              <template #content>
                <div class="text-[12px] text-[#4D4F56] w-full mb-[5px]">
                  {{ $t('pkgUpload.packageName', { name: curFile.data?.name }) }}
                </div>
                <div class="text-[12px] text-[#4D4F56] w-full mb-[5px]">
                  {{ $t('pkgUpload.md5', { md5: curFile.data?.md5 }) }}
                </div>
                <div class="text-[12px] text-[#4D4F56] w-full mb-[22px]">
                  {{ $t('pkgUpload.overwriteTip', { type: capitalizeFirstLetter(currentType) }) }}
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
  </Upload>
</template>
<script lang="ts" setup>
import { Button, Message, PopConfirm, Progress, Upload } from 'bkui-vue';
import { RightTurnLine } from 'bkui-vue/lib/icon';
import { computed, reactive, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute } from 'vue-router';

import { PACKAGE_GENERATION } from '@/common/const';
import { bytesToMegabytes, capitalizeFirstLetter } from '@/common/util';

const props = defineProps({
  pluginType: {
    type: String,
    default: '',
  },
  proxyType: {
    type: String,
    default: 'origin_proxy',
  },
});
const emit = defineEmits(['upload', 'cancel', 'loading']);

const { t } = useI18n();
const route = useRoute();
const uploader = ref(null);

const currentType = computed(() => {
  const routeName = route.name?.toString() || '';
  const type = routeName.split('PackageMng')[0];
  return type;
});
const url = computed(() => {
  let type;
  switch (currentType.value) {
    case 'proxy':
      type = props.proxyType === 'origin_proxy' ? 'proxy' : 'server';
      break;
    case 'plugin':
      type = props.pluginType;
      break;
    default:
      type = currentType.value;
      break;
  }
  return `${location.origin}/api/v3/package/upload/origin/${type}`;
});
const curFile = reactive({
  file: null as File | null,
  progress: 0,
  status: '',
  data: null,
  message: '',
});

type UploadRespBody = {
  code?: number;
  message?: string;
  data?: any;
  error?: { message?: string; details?: Array<{ message?: string }> };
};

/** parseUploadBody 解析上传响应体，非 JSON 响应（如网关错误页）返回空对象。 */
const parseUploadBody = (raw: string): UploadRespBody => {
  try {
    return JSON.parse(raw) as UploadRespBody;
  } catch {
    return {};
  }
};

/**
 * resolveUploadErrMsg 提取后端返回的具体失败原因：
 * 标准错误响应把真实原因放在 error.details，其次 error.message、message；都取不到时返回空串，由调用方兜底。
 */
const resolveUploadErrMsg = (body: UploadRespBody, statusText: string) => {
  const detailMsgs = (body.error?.details ?? [])
    .map(detail => detail.message)
    .filter((message): message is string => !!message);
  if (detailMsgs.length > 0) {
    return detailMsgs.join('; ');
  }
  return body.message || body.error?.message || statusText || '';
};

/** failUpload 标记当前文件上传失败；errMsg 为空时只提示通用文案，避免渲染出“上传失败：上传失败”。 */
const failUpload = (errMsg: string) => {
  const detail = errMsg.trim();
  curFile.message = detail;
  curFile.status = 'failed';
  Message({
    theme: 'error',
    message: detail ? t('pkgUpload.uploadFailedDetail', { message: detail }) : t('pkgUpload.uploadFailed'),
  });
  emit('upload', null);
};

const handleBeforeUpload = (file: File, fileList: File[]) => {
  const whiteList = ['tgz', 'tar', 'gz'];
  const AllFiles = fileList.filter(v => whiteList.includes(v.name.substring(file.name.lastIndexOf('.') + 1)));
  if (AllFiles.length !== fileList.length) {
    fileList.pop();
    Message({
      theme: 'warning',
      message: t('pkgUpload.fileFormatError'),
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
  const metadata = JSON.stringify({ generation: PACKAGE_GENERATION, overwrite: true });
  formData.append('metadata', metadata);
  curFile.status = '';
  curFile.progress = 0;
  curFile.data = null;
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
    const body = parseUploadBody(xhr.responseText);
    // 仅 HTTP 2xx 且业务 code 为 0 才算上传成功，其余情况都要把后端错误信息暴露出来
    if (xhr.status >= 200 && xhr.status < 300 && body.code === 0) {
      curFile.status = (body.data?.existed && !overwrite.value)
        ? 'existed'
        : 'success';
      curFile.data = body.data;
      emit('upload', body.data);
      return;
    }
    failUpload(resolveUploadErrMsg(body, xhr.statusText));
  };

  // 网络层错误（请求未完成，拿不到响应体）
  xhr.onerror = () => {
    failUpload('');
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
  // 重置当前文件状态并通知父组件清空上传结果：
  // 删除后没有可提交的包，父组件需据此隐藏「全租户共享」开关并禁用提交按钮
  curFile.file = null;
  curFile.progress = 0;
  curFile.status = '';
  curFile.data = null;
  curFile.message = '';
  emit('upload', null);
};
</script>

<style lang="postcss" scoped>
/* 文件条目默认 max-height: 60px，错误信息换行后会被裁切，这里放开高度让外框随内容增大 */
:deep(.bk-upload-list__item) {
  max-height: none;
}

/* 错误信息框：跟随文本换行自动增高，不与右侧重试/删除按钮重叠 */
.pkg-upload-error {
  margin: 6px 72px 0 0;
  padding: 6px 8px;
  border: 1px solid #ffd9d9;
  border-radius: 2px;
  background: #fff0f0;
  color: #ff5656;
  font-size: 12px;
  line-height: 18px;
  white-space: pre-wrap;
  word-break: break-all;
}
</style>
