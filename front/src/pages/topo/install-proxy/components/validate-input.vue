<template>
  <Input
    v-if="isValid === validStatus.failed"
    v-model="value"
    class="border-[#EA3636] cursor-pointer !hover:border-[#4487f5]">
    <template #suffix>
      <div class="flex items-center mr-[5px]">
        <i
          class="nodeman-icon nc-remind-fill text-[#ea3636]"
          v-bk-tooltips="{
            content: $t('validate.required'),
          }"></i>
      </div>
    </template>
  </Input>
  <Input v-model="value" behavior="simplicity" class="border-none cursor-pointer" v-else />
</template>

<script lang="ts" setup>
import { Input } from 'bkui-vue';
import { ref } from 'vue';

const value = defineModel<number | string>('value');

const props = defineProps({
  required: {
    type: Boolean,
    default: false,
  },
  validator: {
    type: Function,
    required: false,
  },
});

const enum validStatus {
  unValid,
  success,
  failed,
};

const isValid = ref(validStatus.unValid);

const clearValidate = () => isValid.value = validStatus.unValid;

const validateInput = () => {
  let valid = true;
  clearValidate();
  if (props.required) valid = value.value !== '';
  if (props.validator) {
    valid = props.validator(value.value);
  }
  if (!valid) isValid.value = validStatus.failed;
  return valid;
};

defineExpose({
  clearValidate,
  validateInput,
});

</script>
