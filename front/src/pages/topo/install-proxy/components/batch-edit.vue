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
      <Select v-if="type === 'select'" v-model="batchValue">
        <Select.Option
          v-for="item in options"
          :key="item.id"
          :id="item.id"
          :name="item.value">
        </Select.Option>
      </Select>
    </template>
  </PopConfirm>
</template>

<script lang="ts" setup>
import { Input, PopConfirm, Select, Switcher } from 'bkui-vue';
import { ref, watch } from 'vue';

type batchEditType = 'input' | 'switcher' | 'select';

interface IOptions {
  id: string | number
  value: string | number
}

interface IProps {
  type: batchEditType
  options?: IOptions[]
  title: string
}

defineProps<IProps>();

const emit = defineEmits(['confirm']);

const batchValue = ref();
const handleBatchEdit = () => {
  emit('confirm', batchValue.value);
};

const batchEditRef = ref();

watch(() => batchEditRef.value?.popoverRef.localIsShow, (curShow) => {
  if (!curShow) {
    batchValue.value = '';
  }
});
</script>
