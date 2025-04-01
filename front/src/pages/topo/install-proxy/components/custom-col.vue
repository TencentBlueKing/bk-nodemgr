<template>
  <VxeColumn
    :field="field"
    :width="width || defaultSubTitleWidth"
  >
    <template #header>
      <span class="mr-[5px]">{{ title }}</span>
      <span v-if="required" class="mx-[3px] text-[#FF5656]">*</span>
      <slot name="edit" :field="field"></slot>
    </template>

    <!-- 传递required逻辑、Vxe Slot数据 -->
    <template #default="{ row, $rowIndex, $columnIndex, rowid }">
      <slot
        :row="row"
        :rowid="rowid"
        :field="field"
        :row-index="$rowIndex"
        :column-index="$columnIndex"
        :required="required">
      </slot>
    </template>
  </VxeColumn>
</template>

<script lang="ts" setup>
import { VxeColumn } from '@blueking/vxe-table';
// 二级表头
export interface IColumn {
  title: string
  field: string
  width?: number // 默认宽度 defaultSubTitleWidth
  required?: boolean // 是否必填(展示 '*' 标志)，使用 Slot 自行配置校验逻辑
}

defineProps<IColumn>();
const defaultSubTitleWidth = 100;
</script>
