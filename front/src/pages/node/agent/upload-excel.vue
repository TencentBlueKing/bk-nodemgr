<template>
  <div class="upload">
    <!-- 上传按钮区域（无文件时显示） -->
    <div v-if="!Object.keys(file).length" class="upload-wrapper relative">
      <Button class="upload-btn w-full">
        <span class="upload-btn-content flex items-center justify-center relative left-[-2px]">
          <i :class="icon" :style="{ 'font-size': `${iconSize}px` }"></i>
          <span class="ml-1">{{ title }}</span>
        </span>
      </Button>
      <input
        ref="uploadel"
        @change="handleChange"
        :accept="accept"
        :name="name"
        title=""
        type="file"
        class="upload-input absolute left-0 top-0 w-full h-full cursor-pointer opacity-0"
        :multiple="false"
      />
    </div>

    <!-- 文件信息区域（有文件时显示，支持插槽） -->
    <slot
      name="uploadInfo"
      v-else
      :file="file"
      :file-change="handleChange"
      :handle-retry="handleRetry"
    >
      <div
        class="upload-info flex items-center px-[2px] pr-1.5 h-8"
        :class="{ 'bg-gray-100 rounded': hoverInfo && !disableHoverCls }"
        @mouseenter="hoverInfo = true"
        @mouseleave="hoverInfo = false"
      >
        <!-- 文件图标 -->
        <div class="info-left text-gray-400 text-lg">
          <i :class="fileIcon"></i>
        </div>

        <!-- 文件信息（名称 + 进度条） -->
        <div class="info-right ml-1.5 flex-1 min-w-0">
          <!-- 文件名 + 删除按钮 -->
          <div class="info-name relative flex items-center h-4">
            <span
              class="file-name text-ellipsis overflow-hidden whitespace-nowrap text-sm"
              :title="file.name"
            >
              {{ file.name || '未命名文件' }}
            </span>
            <span class="file-extension ml-1 text-sm text-gray-500">
              {{ file.extension || '' }}
            </span>
            <!-- 删除按钮（悬浮显示） -->
            <i
              v-show="hoverInfo"
              class="file-abort nodeman-icon nc-delete absolute right-0 text-lg cursor-pointer text-gray-500 hover:text-red-500"
              @click="handleAbortUpload"
            ></i>
          </div>

          <!-- 上传进度条（非100%时显示） -->
          <div
            v-show="file.percentage !== '100%'"
            class="info-progress w-full h-0.5 mt-0.5 bg-gray-200 rounded"
          >
            <div
              class="progress-bar h-full rounded transition-all duration-300 ease-in-out"
              :class="{ 'bg-red-400': file.hasError, 'bg-blue-500': !file.hasError }"
              :style="{ width: file.percentage || '0%' }"
            ></div>
          </div>
        </div>
      </div>
    </slot>
  </div>
</template>

<script setup lang="ts">
import { Button, Message } from 'bkui-vue';
import { computed, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
interface IFileInfo {
  uid: number
  name: string
  value?: string
  extension: string // 文件类型
  size: number
  status: string
  percentage: string
  errorMsg: string
  hasError: boolean
  originFile: File
  type: string
}

// ---------------------- 1. 补全Props默认值，增强类型安全 ----------------------
const props = defineProps<{
  name?: string;
  accept?: string;
  acceptTips?: string;
  action?: string;
  maxSize?: number;
  unit?: 'KB' | 'MB';
  headers?: Record<string, string> | Array<{ name: string; value: string }>;
  withCredentials?: boolean;
  onUploadError?: (err: Error, file: File) => void;
  onUploadSuccess?: (res: any, file: File) => void;
  onUploadProgress?: (e: ProgressEvent & { percent?: number }, file: File) => void;
  icon?: string;
  fileIcon?: string;
  iconSize?: string | number;
  title?: string;
  parseText?: boolean;
  disableHoverCls?: boolean;
  fileInfo?: IFileInfo;
}>();

// ---------------------- 2. 定义Emits，明确事件类型 ----------------------
const emit = defineEmits<{
  (e: 'change', value: { value: string; fileInfo: IFileInfo }): void;
  (e: 'before-upload', file: IFileInfo): boolean | void;
}>();
// 补充Props默认值（避免undefined）
const defaultProps = {
  name: 'file_data',
  icon: 'bk-icon icon-plus',
  fileIcon: 'nodeman-icon nc-key',
  iconSize: '22',
  title: useI18n().t('上传文件'),
  maxSize: 500,
  unit: 'MB' as const,
  acceptTips: useI18n().t('文件类型不符'),
  parseText: false,
  disableHoverCls: false,
  withCredentials: false,
  headers: {},
  onUploadError: () => {},
  onUploadSuccess: () => {},
  onUploadProgress: () => {},
};
const getProp = <T extends keyof typeof defaultProps>(key: T): typeof defaultProps[T] => props[key] ?? defaultProps[key];

// ---------------------- 3. 响应式状态 ----------------------
const { t } = useI18n();
const uploadel = ref<HTMLInputElement | null>(null);
const file = ref<IFileInfo>({} as IFileInfo);
const reqsMap = ref<Record<string | number, XMLHttpRequest>>({});
const fileIndex = ref(1);
const hoverInfo = ref(false);

// ---------------------- 4. 计算属性：最大文件大小（字节） ----------------------
const maxFileSize = computed(() => {
  const base = getProp('maxSize');
  const unit = getProp('unit');
  return unit === 'KB' ? base * 1024 : base * 1024 * 1024;
});

// ---------------------- 5. Watch监听fileInfo，支持外部回显 ----------------------
watch(
  () => props.fileInfo,
  (newVal) => {
    if (newVal && Object.keys(newVal).length) {
      file.value = JSON.parse(JSON.stringify(newVal)); // 深拷贝避免引用污染
    }
  },
  { immediate: true, deep: true },
);

// ---------------------- 6. 补全缺失的onSuccess函数 ----------------------
/**
 * 上传成功：解析响应内容（处理JSON/纯文本）
 */
const onSuccess = (xhr: XMLHttpRequest): any => {
  const response = xhr.responseText || xhr.response;
  if (!response) return null;
  // 尝试解析JSON，失败则返回原始文本
  try {
    return JSON.parse(response);
  } catch (e) {
    return response;
  }
};

// ---------------------- 7. 补全HTTP进度/成功/失败处理函数 ----------------------
/**
 * 处理上传进度
 */
const handleHttpProgress = (e: ProgressEvent & { percent?: number }, rawFile: File) => {
  if (e.total > 0) {
    e.percent = Math.round((e.loaded / e.total) * 100);
    file.value.percentage = `${e.percent}%`;
  }
  file.value.status = 'uploading';
  getProp('onUploadProgress')(e, rawFile); // 触发外部进度回调
};

/**
 * 处理上传成功
 */
const handleHttpSuccess = (res: any, rawFile: File) => {
  file.value.status = 'success';
  file.value.percentage = '100%';
  file.value.hasError = false;
  // 触发外部成功回调和change事件
  getProp('onUploadSuccess')(res, rawFile);
  emit('change', { value: JSON.stringify(res), fileInfo: file.value });
};

/**
 * 处理上传失败
 */
const handleHttpError = (err: Error, rawFile: File) => {
  file.value.status = 'error';
  file.value.hasError = true;
  file.value.errorMsg = err.message;
  // 触发外部失败回调
  getProp('onUploadError')(err, rawFile);
  Message({ theme: 'error', message: err.message });
};

// ---------------------- 8. 核心业务逻辑（文件校验、解析、上传） ----------------------
/**
 * 文件校验（大小 + 类型）
 */
const validateFile = (rawFile: File): boolean => {
  if (!rawFile) return false;

  // 大小校验
  if (rawFile.size > maxFileSize.value) {
    const msg = t('文件不能超过 {size} {unit}', {
      size: getProp('maxSize'),
      unit: getProp('unit'),
    });
    Message({ theme: 'error', message: msg });
    return false;
  }

  // 类型校验（accept有值时）
  if (props.accept) {
    const acceptTypes = props.accept.split(',').map(t => t.trim());
    if (!acceptTypes.includes(rawFile.type)) {
      Message({ theme: 'error', message: getProp('acceptTips') });
      return false;
    }
  }

  return true;
};

/**
 * 组装文件信息（统一格式）
 */
const handleAssembleFile = (rawFile: File): IFileInfo => {
  const [fileName, ...extParts] = rawFile.name.split('.');
  const extension = extParts.length ? extParts.join('.') : ''; // 处理多后缀名（如.tar.gz）
  const uid = Date.now() + fileIndex.value++;

  return {
    name: fileName,
    extension,
    type: rawFile.type,
    size: rawFile.size,
    percentage: '0%',
    uid,
    originFile: rawFile,
    status: 'ready' as const,
    hasError: false,
    errorMsg: '',
  };
};

/**
 * 前端解析文件（仅文本格式）
 */
const handleParseText = (rawFile: File) => {
  file.value = handleAssembleFile(rawFile);
  const reader = new FileReader();

  reader.onload = (ev) => {
    try {
      const result = (ev.target as FileReader).result as string;
      file.value.status = 'success';
      file.value.percentage = '100%';
      emit('change', { value: result, fileInfo: file.value });
      getProp('onUploadSuccess')(result, rawFile);
    } catch (err) {
      const error = err as Error;
      file.value.status = 'error';
      file.value.hasError = true;
      file.value.errorMsg = error.message;
      getProp('onUploadError')(error, rawFile);
      Message({ theme: 'error', message: t('解析文件失败：{msg}', { msg: error.message }) });
    }
  };

  reader.onprogress = (e) => {
    handleHttpProgress(e as ProgressEvent & { percent?: number }, rawFile);
  };

  reader.readAsText(rawFile);
};

/**
 * 发起HTTP上传请求
 */
const handleHttpRequest = (option: {
  headers: Record<string, string> | Array<{ name: string; value: string }>;
  withCredentials: boolean;
  file: File;
  filename: string;
  action: string;
  onProgress: (e: ProgressEvent & { percent?: number }, file: File) => void;
  onSuccess: (res: any, file: File) => void;
  onError: (err: Error, file: File) => void;
}): XMLHttpRequest | undefined => {
  // 校验上传地址
  if (!option.action) {
    const error = new Error(t('上传地址不能为空'));
    option.onError(error, option.file);
    return undefined;
  }

  const xhr = new XMLHttpRequest();

  // 进度监听
  if (xhr.upload) {
    xhr.upload.onprogress = (e) => {
      const progressEv = e as ProgressEvent & { percent?: number };
      option.onProgress(progressEv, option.file);
    };
  }

  // 错误监听
  xhr.onerror = () => {
    option.onError(new Error(t('网络请求异常')), option.file);
  };

  // 响应处理
  xhr.onload = () => {
    if (xhr.status < 200 || xhr.status >= 300) {
      const error = new Error(t('请求失败：{status} {statusText}', {
        status: xhr.status,
        statusText: xhr.statusText,
      }));
      (error as any).status = xhr.status;
      option.onError(error, option.file);
      return;
    }

    // 解析响应并触发成功回调
    try {
      const res = onSuccess(xhr);
      option.onSuccess(res, option.file);
    } catch (err) {
      option.onError(err as Error, option.file);
    }
  };

  // 构建FormData
  const formData = new FormData();
  formData.append(option.filename, option.file, option.file.name);

  // 配置并发送请求
  xhr.open('POST', option.action, true);
  xhr.withCredentials = option.withCredentials;

  // 设置请求头（处理数组/对象两种格式）
  if (option.headers) {
    if (Array.isArray(option.headers)) {
      option.headers.forEach(({ name, value }) => xhr.setRequestHeader(name, value));
    } else {
      Object.entries(option.headers).forEach(([key, value]) => xhr.setRequestHeader(key, value));
    }
  }

  xhr.send(formData);
  return xhr;
};

/**
 * 处理文件选择变更
 */
const handleChange = (e: Event) => {
  const target = e.target as HTMLInputElement;
  const { files } = target;
  if (!files?.length) return;

  const rawFile = files[0];
  if (validateFile(rawFile)) {
    file.value = {} as IFileInfo; // 重置状态
    // 前端解析/后端上传二选一
    getProp('parseText') ? handleParseText(rawFile) : handleUploadFiles(rawFile);
  }
  target.value = ''; // 重置input，支持重复选择同一文件
};

/**
 * 发起文件上传（后端）
 */
const handleUploadFiles = (rawFile: File) => {
  file.value = handleAssembleFile(rawFile);
  const { uid, originFile } = file.value;

  // 触发before-upload，支持拦截上传（返回false则终止）
  const beforeResult = emit('before-upload', file.value);
  if (beforeResult === false) return;

  // 发起上传请求
  const xhr = handleHttpRequest({
    headers: getProp('headers'),
    withCredentials: getProp('withCredentials'),
    file: originFile,
    filename: getProp('name'),
    action: props.action || '',
    onProgress: handleHttpProgress,
    onSuccess: handleHttpSuccess,
    onError: handleHttpError,
  });

  // 存储请求实例，用于后续终止上传
  if (xhr && uid) {
    reqsMap.value[uid] = xhr;
  }
};

/**
 * 终止上传并重置状态
 */
const handleAbortUpload = () => {
  const { uid } = file.value;
  if (uid && reqsMap.value[uid]) {
    reqsMap.value[uid].abort();
    delete reqsMap.value[uid];
  }
  // 重置文件状态并触发change事件
  const emptyFile = {} as IFileInfo;
  file.value = emptyFile;
  hoverInfo.value = false;
  emit('change', { value: '', fileInfo: emptyFile });
};

/**
 * 重试上传（仅支持已选择过的文件）
 */
const handleRetry = () => {
  const { originFile } = file.value;
  if (originFile) {
    getProp('parseText') ? handleParseText(originFile) : handleUploadFiles(originFile);
  }
};

// ---------------------- 9. 暴露方法给父组件（如需外部调用） ----------------------
defineExpose({
  handleRetry,
  handleAbortUpload,
  handleChange,
});
</script>
<style lang="postcss" scoped>
  .upload-wrapper {
    position: relative;
    &:hover {
      button {
        border-color: #979ba5;
        color: #63656e;
      }
    }
    .upload-btn {
      width: 100%;
    }
    .upload-btn-content {
      position: relative;
      left: -2px;

      i {
        top: 0;
      }
    }
    .upload-input {
      position: absolute;
      left: 0;
      top: 0;
      width: 100%;
      height: 100%;
      cursor: pointer;
      opacity: 0;
    }
  }
  .upload-info {
    padding-left: 2px;
    padding-right: 5px;
    height: 32px;

    &.hover {
      background: #f0f1f5;
      border-radius: 2px;
    }
    .info-left {
      font-size: 18px;
      color: #c4c6cc;
    }
    .info-right {
      width: 0;
      flex: 1;
      .info-name {
        position: relative;
        line-height: 16px;

        .file-name {
          height: 16px;
          word-break: break-all;
          overflow: hidden;
          text-overflow: ellipsis;
          white-space: nowrap;
        }
        .file-extension {
          margin-right: 20px;
        }
        .file-abort {
          position: absolute;
          right: 0;
          font-size: 18px;
          cursor: pointer;
        }
      }
      .info-progress {
        width: 100%;
        height: 2px;
        background: #dcdee5;
        border-radius: 1px;
        .progress-bar {
          height: 2px;
          border-radius: 1px;
          background: #3a84ff;
          transition: width .3s ease-in-out;
        }
      }
    }
  }
</style>
