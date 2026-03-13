<template>
  <PopConfirm
    ref="batchEditRef"
    width="280"
    trigger="click"
    placement="bottom"
    :title="title"
    @confirm="handleBatchEdit"
  >
    <i class="nodeman-icon nc-edit text-[18px] cursor-pointer"></i>
    <template #content>
      <Input v-if="type === 'input'" v-model="batchValue" class="mb-[15px]" />
      <Switcher v-if="type === 'switcher'" v-model="batchValue" class="mb-[15px]" />
      <Select
        v-if="type === 'select'"
        class="mb-[15px]"
        v-model="batchValue"
        :popover-options="{
          boundary: 'parent'
        }"
      >
        <Select.Option
          v-for="item in options"
          :key="item.id"
          :id="item.id"
          :name="item.name">
        </Select.Option>
      </Select>
      <div v-if="type === 'credit'">
        <Input type="password" v-model="batchValue" class="mb-[15px]" />
        <p class="text-[12px] text-[#979ba5] mt-[-6px]">{{ $t('components.batchEdit.passwordTip') }}</p>
        <p class="mt-[14px] mb-[10px] text-[14px]">{{ $t('components.batchEdit.batchEditKey') }}</p>
        <Upload
          ref="uploader"
          type="formdata"
          :url="url"
          :size="100"
          :multiple="false"
          :limit="1"
          theme="button"
          :before-upload="handleBeforeUpload"
          :custom-request="() => {}"
        ></Upload>
        <p class="text-[12px] text-[#979ba5] mt-[6px]">{{ $t('components.batchEdit.keyTip') }}</p>
      </div>
    </template>
  </PopConfirm>
</template>

<script lang="ts" setup>
import { Input, PopConfirm, Select, Switcher, Upload } from 'bkui-vue';
import { ref, watch } from 'vue';

type batchEditType = 'input' | 'credit' | 'switcher' | 'select';

interface IOptions {
  id: string | number
  name: string | number
}

interface IProps {
  type: batchEditType
  options?: IOptions[]
  title: string
}

const props = defineProps<IProps>();

const emit = defineEmits(['confirm']);

const batchValue = ref();
const keyBase64 = ref();
const handleBatchEdit = () => {
  if (props.type === 'credit') {
    emit('confirm', {
      password: batchValue.value,
      key: keyBase64.value,
    });
    return;
  }
  emit('confirm', batchValue.value);
};

const batchEditRef = ref();
// 上传密钥
const url = location.href;
const handleBeforeUpload = (file: File) => {
  const reader = new FileReader();
  reader.onload = (event) => {
    const result = event.target?.result as string;
    if (result) {
      // 检查是否是Data URL格式
      if (result.startsWith('data:')) {
        keyBase64.value = result.split(',')[1];
      } else {
        keyBase64.value = result;
      }
      
    }
  };
  reader.readAsDataURL(file);
  return true;
};
watch(() => batchEditRef.value?.popoverRef.localIsShow, (curShow: boolean) => {
  if (!curShow) {
    batchValue.value = '';
  }
});
</script>
