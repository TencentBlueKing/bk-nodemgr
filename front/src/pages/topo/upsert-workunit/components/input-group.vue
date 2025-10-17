<template>
  <div class="flex items-center mb-[8px]" v-for="(item, index) in values" :key="index">
    <Input v-model.trim="values[index]" class="w-[283px] text-[14px] mr-[13px]" :placeholder="placeholder" clearable />
    <i class="nodeman-icon nc-plus mr-[10px] text-[#C4C6CC] cursor-pointer" @click="handleAddInput"></i>
    <i class="nodeman-icon nc-minus text-[#DCDEE5] cursor-not-allowed" v-if="values.length === 1"></i>
    <i class="nodeman-icon nc-minus text-[#C4C6CC] cursor-pointer" v-else @click="handleMinusInput(index)"></i>
  </div>
</template>

<script lang="ts" setup>
import { Input } from 'bkui-vue';
const values = defineModel<string[]>('values', { required: true });

defineProps({
  placeholder: {
    type: String,
    default: '',
  },
});

const handleAddInput = () => {
  values.value.push('');
};

const handleMinusInput = (index: number) => {
  values.value.splice(index, 1);
};

</script>
