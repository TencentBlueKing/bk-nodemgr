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

      <!-- 重载配置勾选框（原始设计：放在内容区，默认勾选） -->
      <div v-if="showRestartOptions" class="mb-[16px]">
        <Checkbox v-model="isReconfig">{{ $t('components.operateDialog.reloadConfig') }}</Checkbox>
      </div>

      <!-- 操作目标数据表格（简化模式下不展示） -->
      <Table
        v-if="showTable && isShow && data.length > 0 && columns.length > 0"
        :data="data"
        :max-height="500"
        class="mt-[16px]"
        @checkbox-change="handleSelectChange"
        @checkbox-all="handleSelectAllChange"
      >
        <TableColumn v-if="!hideCheckbox" type="checkbox" width="60" fixed="left" />
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
    <template #footer>
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-[8px]" v-if="showRestartOptions">
          <Radio.Group v-model="isForce">
            <Radio.Button :label="true">
              {{ $t('components.operateDialog.forceRestart') }}
            </Radio.Button>
            <Radio.Button :label="false">
              {{ $t('components.operateDialog.gracefulRestart') }}
            </Radio.Button>
          </Radio.Group>
          <div class="flex items-center gap-[3px] ml-[20px]" v-show="!isForce">
            <span>{{ $t('components.operateDialog.gracefulTime') }}</span>
            <Input type="number" v-model="gracefulTimeout" class="w-[80px] mx-[3px]"></Input>
            <span>{{ $t('components.operateDialog.seconds') }}</span>
          </div>
        </div>
        <div class="ml-auto">
          <Button theme="primary" @click="handleConfirm">{{ $t('action.confirm') }}</Button>
          <Button class="ml-[8px]" @click="handleCancel">{{ $t('action.cancel') }}</Button>
        </div>
      </div>
    </template>
  </Dialog>
</template>
<script lang="ts" setup>
import { Button, Checkbox, Dialog, Input, Radio } from 'bkui-vue';
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
  /** 隐藏重启选项（插件重启等场景不需要强制/无损切换） */
  hideRestartOptions: {
    type: Boolean,
    default: false,
  },
  /** 隐藏勾选列（插件操作等场景不需要勾选） */
  hideCheckbox: {
    type: Boolean,
    default: false,
  },
});
const emit = defineEmits(['confirm', 'cancel', 'selection-change']);
const { t } = useI18n();

/** 勾选的行数据 */
const selection = ref<any[]>([]);

/** 是否强制重启 */
const isForce = ref(false);
/** 无损重启超时时间（秒） */
const gracefulTimeout = ref(120);
/** 是否重载配置 */
const isReconfig = ref(true);

/** 是否展示重启选项（强制/无损 + 超时时间） */
const showRestartOptions = computed(() => props.type === 'restart' && !props.hideRestartOptions);

const dialogWidth = computed(() => {
  // 重启选项需要更宽的弹窗以容纳 footer 中的强制/无损选项
  if (showRestartOptions.value) return 800;
  // 简化模式（无表格）使用较窄的弹窗
  if (!props.showTable) return 700;
  // 有表格时弹窗更宽
  return props.data.length > 0 && props.columns.length > 0 ? 900 : 700;
});

/** 勾选后表格下方的确认提示文案 */
const selectionConfirmText = computed(() => {
  if (!props.selectionConfirmFormatter) return '';
  // 无勾选列或简化模式下使用全部数据，表格模式下使用勾选数据
  const items = (props.hideCheckbox || !props.showTable) ? props.data : selection.value;
  if (items.length === 0) return '';
  return props.selectionConfirmFormatter(items);
});

function handleConfirm() {
  // 无勾选列或简化模式下使用全部数据，表格模式下使用勾选数据
  const finalSelection = (props.hideCheckbox || !props.showTable) ? props.data : selection.value;
  emit('confirm', {
    selection: finalSelection,
    isForce: isForce.value,
    time: gracefulTimeout.value,
    isReconfig: isReconfig.value,
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
      // 重置强制重启选项
      isForce.value = false;
      gracefulTimeout.value = 120;
      isReconfig.value = true;
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
