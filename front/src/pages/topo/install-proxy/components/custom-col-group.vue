<template>
  <VxeColgroup align="center">
    <template #header>
      <span>{{ title }}</span>
      <span v-if="required" class="mx-[3px] text-[#FF5656]">*</span>
    </template>
    <slot :required="required"></slot>
  </VxeColgroup>
</template>

<script lang="ts" setup>
import { VxeColgroup } from '@blueking/vxe-table';
// 一级表头
export interface EditColGroupProps {
  title: string
  required?: boolean // 所属的所有的二级表头是否必填(展示 '*' 标志)，使用 Slot 自行配置校验逻辑
}

defineProps<EditColGroupProps>();
</script>
