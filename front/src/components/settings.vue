<template>
  <Popover
    ref="settingRef"
    placement="bottom"
    width="390"
    theme="light"
    trigger="click"
    @after-hidden="afterHidden"
  >
    <Button
      text
    >
      <i class="nodeman-icon nc-setting"></i>
    </Button>
    <template #content>
      <div class="setting-panel">
        <div class="setting-title">表格设置</div>
        <div class="setting-field-section">
          <div class="setting-field-header">
            <span class="setting-field-label">字段显示设置</span>
            <Checkbox
              v-model="isSelectAll"
              :indeterminate="isIndeterminate"
              @change="handleSelectAll"
              class="setting-select-all"
            >
              全选
            </Checkbox>
          </div>
          <Checkbox.Group
            v-model="checkboxGroupValue"
            @change="handleChange"
            class="setting-checkbox-group"
          >
            <Checkbox
              v-for="item in settings.fields"
              :key="item.field"
              :label="item.field"
              :disabled="settings.disabled.includes(item.field)"
              class="setting-checkbox-item"
            >
              {{ item.title }}
            </Checkbox>
          </Checkbox.Group>
        </div>
      </div>
    </template>
  </Popover>
</template>
<script lang="ts" setup>
import { Button, Checkbox, Popover } from 'bkui-vue';
import { computed, ref } from 'vue';

const props = defineProps({
  settings: {
    type: Object,
    default: () => [{
      fields: [],
      checked: [] as string[],
      disabled: [] as string[],
      size: 'medium',
    }],
  },
});
const emit = defineEmits(['setting-change']);

const checkboxGroupValue = ref(props.settings.checked);
const size = ref(props.settings.size);

// 全选相关逻辑
const enabledFields = computed(() =>
  props.settings.fields
    .filter((item: any) => !props.settings.disabled.includes(item.field))
    .map((item: any) => item.field),
);
const isSelectAll = computed(() =>
  enabledFields.value.every((field: string) => checkboxGroupValue.value.includes(field)),
);
const isIndeterminate = computed(() => {
  const checkedCount = enabledFields.value.filter((field: string) => checkboxGroupValue.value.includes(field)).length;
  return checkedCount > 0 && checkedCount < enabledFields.value.length;
});

const handleSelectAll = (val: boolean) => {
  if (val) {
    // 全选：合并 disabled 中已选 + 所有 enabled
    const disabledChecked = checkboxGroupValue.value.filter((f: string) => props.settings.disabled.includes(f));
    checkboxGroupValue.value = [...new Set([...disabledChecked, ...enabledFields.value])];
  } else {
    // 取消全选：只保留 disabled 中已选
    checkboxGroupValue.value = checkboxGroupValue.value.filter((f: string) => props.settings.disabled.includes(f));
  }
};

const handleChange = (value: string[]) => {
  checkboxGroupValue.value = value;
};

const afterHidden = () => {
  emit('setting-change', { checked: checkboxGroupValue.value, size: size.value });
};
const settingRef = ref();
const showSetting = () => {
  if (settingRef.value) {
    settingRef.value.show();
  }
};

defineExpose({
  showSetting,
});
</script>
<style lang="postcss" scoped>
.setting-panel {
  margin: -12px;
}

.setting-title {
  font-size: 18px;
  font-weight: 700;
  color: #313238;
  line-height: 24px;
  padding: 25px 24px 0;
}

.setting-field-section {
  padding: 0 24px;
  margin-top: 16px;
  padding-bottom: 8px;
}

.setting-field-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;
}

.setting-field-label {
  font-size: 14px;
  font-weight: normal;
  color: #63656e;
  line-height: 20px;
}

.setting-select-all {
  font-size: 14px;
}

.setting-checkbox-group {
  display: flex !important;
  flex-wrap: wrap;
  width: 100%;
}

.setting-checkbox-item {
  height: 36px;
  box-sizing: border-box;
  &:nth-child(odd) {
    width: 55%;
  }
  &:nth-child(even) {
    width: 45%;
  }
}

.bk-checkbox ~ .bk-checkbox {
  margin-left: 0;
}

.nodeman-icon nc-setting {
  color: #c4c6cc;
}
</style>
