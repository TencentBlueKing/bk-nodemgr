<template>
  <Popover
    ref="settingRef"
    placement="bottom"
    width="240"
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
      <div class="m-[-12px]">
        <div class="flex h-[42px] text-[14px] text-[#63656e] bg-[#f0f1f5]">
          <div
            :class="[
              'flex items-center justify-center w-[120px] cursor-pointer transition-all duration-100',
              { 'text-[#3a84ff] bg-[#fff]': activeSetting === 'field' },
            ]"
            @click="activeSetting = 'field'"
          >
            字段设置
          </div>
          <!-- <div
            :class="[
              'flex items-center justify-center w-[120px] cursor-pointer transition-all duration-100',
              { 'text-[#3a84ff] bg-[#fff]': activeSetting === 'advance' },
            ]"
            @click="activeSetting = 'advance'"
          >
            高级设置
          </div> -->
        </div>
        <div
          class="bg-[#fff] px-[16px] my-[8px] overflow-y-auto"
          v-show="activeSetting === 'field'"
        >
          <Checkbox.Group
            v-model="checkboxGroupValue"
            @change="handleChange"
            class="flex flex-col justify-center"
          >
            <Checkbox
              v-for="item in settings.fields"
              :key="item.field"
              :label="item.field"
              :disabled="settings.disabled.includes(item.field)"
              class="text-[14px] h-[32px]"
            >
              {{ item.title }}
            </Checkbox>
          </Checkbox.Group>
        </div>
        <div
          class="bg-[#fff] px-[16px] overflow-y-auto my-[16px]"
          v-show="activeSetting === 'advance'"
        >
          <div class="text-[14px] mb-[8px]">表格行高</div>
          <Radio.Group
            v-model="size"
            class="flex justify-center"
            type="card"
            @change="handleSizeChange"
          >
            <Radio.Button label="small">小</Radio.Button>
            <Radio.Button label="medium">中</Radio.Button>
            <Radio.Button label="large">大</Radio.Button>
          </Radio.Group>
        </div>
      </div>
    </template>
  </Popover>
</template>
<script lang="ts" setup>
import { Button, Checkbox, Popover, Radio } from 'bkui-vue';
import { ref } from 'vue';

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

const activeSetting = ref('field');
const checkboxGroupValue = ref(props.settings.checked);
const size = ref(props.settings.size);
const handleChange = (value: string[]) => {
  checkboxGroupValue.value = value;
};
const handleSizeChange = (value: string[]) => {
  size.value = value;
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
.bk-checkbox ~ .bk-checkbox {
  margin-left: 0;
}
</style>
