<template>
  <Dialog
    :is-show="isShow"
    :width="dialogWidth"
    :title="title"
    @closed="isShow = false"
    @confirm="handleConfirm"
    @cancel="handleCancel"
  >
    <div class="w-full">
      <p class="text-[16px] mb-[20px]">{{ subTitle }}</p>

      <!-- 操作目标数据表格（简化模式下不展示） -->
      <Table
        v-if="showTable && isShow && data.length > 0 && columns.length > 0"
        :data="data"
        :max-height="300"
        class="mt-[16px]"
        @checkbox-change="handleSelectChange"
        @checkbox-all="handleSelectAllChange"
      >
        <TableColumn type="checkbox" width="60" fixed="left" />
        <TableColumn
          v-for="col in columns"
          :key="col.field"
          :title="col.label"
          :field="col.field"
          :min-width="col.minWidth"
        >
          <template #default="{ row }">
            <template v-if="col.field === 'status'">
              {{ statusMap[row[col.field]] || row[col.field] || '--' }}
            </template>
            <template v-else>
              {{ row[col.field] || '--' }}
            </template>
          </template>
        </TableColumn>
      </Table>
      <!-- 勾选后的确认提示 -->
      <div v-if="selectionConfirmText" class="selection-confirm-box mt-[12px]">
        <p class="text-[14px] text-[#63656E]">{{ selectionConfirmText }}</p>
      </div>
    </div>
  </Dialog>
</template>
<script lang="ts" setup>
import { Dialog } from 'bkui-vue';
import { computed, type PropType, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { Table, TableColumn } from '@blueking/table';

export interface ColumnConfig {
  field: string;
  label: string;
  width?: number;
  minWidth?: number;
}

const isShow = defineModel('isShow', { type: Boolean, default: false });
const props = defineProps({
  type: {
    type: String,
    default: '',
  },
  title: {
    type: String,
    default: '',
  },
  subTitle: {
    type: String,
    default: '',
  },
  /** 操作目标数据 */
  data: {
    type: Array as () => any[],
    default: () => [],
  },
  /** 表格列配置 */
  columns: {
    type: Array as () => ColumnConfig[],
    default: () => [],
  },
  /** 进程状态映射（用于 status 列渲染） */
  statusMap: {
    type: Object as () => Record<string, string>,
    default: () => ({}),
  },
  /** 勾选确认提示的格式化函数，接收 selection 数组，返回提示文案 */
  selectionConfirmFormatter: {
    type: Function as PropType<(selection: any[]) => string>,
    default: null,
  },
  /** 是否展示表格（简化模式下为 false，仅显示确认提示） */
  showTable: {
    type: Boolean,
    default: true,
  },
});
const emit = defineEmits(['confirm', 'cancel', 'selection-change']);
const { t } = useI18n();

/** 勾选的行数据 */
const selection = ref<any[]>([]);

const dialogWidth = computed(() => {
  // 简化模式（无表格）使用较窄的弹窗
  if (!props.showTable) return 700;
  // 有表格时弹窗更宽
  return props.data.length > 0 && props.columns.length > 0 ? 900 : 700;
});

/** 勾选后表格下方的确认提示文案 */
const selectionConfirmText = computed(() => {
  if (!props.selectionConfirmFormatter) return '';
  // 简化模式下使用全部数据，表格模式下使用勾选数据
  const items = props.showTable ? selection.value : props.data;
  if (items.length === 0) return '';
  return props.selectionConfirmFormatter(items);
});

function handleConfirm() {
  // 简化模式下使用全部数据作为 selection，表格模式下使用勾选数据
  const finalSelection = props.showTable ? selection.value : props.data;
  emit('confirm', {
    selection: finalSelection,
  });
  isShow.value = false;
}

function handleCancel() {
  isShow.value = false;
}

function handleSelectChange({ checked, row }: { checked: boolean; row: any }) {
  if (row) row.checked = checked;
  syncSelection();
}

function handleSelectAllChange({ checked }: { checked: boolean }) {
  props.data.forEach((row: any) => {
    if (row) row.checked = checked;
  });
  syncSelection();
}

function syncSelection() {
  selection.value = props.data.filter((row: any) => row.checked);
  emit('selection-change', selection.value);
}

watch(
  () => isShow,
  async () => {
    if (isShow.value) {
      // 重置勾选状态
      props.data.forEach((row: any) => {
        if (row) row.checked = false;
      });
      selection.value = [];
    }
  },
  { immediate: true, deep: true },
);
</script>
<style lang="postcss" scoped>
.selection-confirm-box {
  padding: 10px 16px;
  border: 1px solid #dcdee5;
  border-radius: 2px;
  background: #f5f7fa;
}
</style>
