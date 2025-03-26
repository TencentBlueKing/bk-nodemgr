<template>
  <div @click="toggleCheck">
    <div class="flex cursor-pointer">
      <!-- 图标 -->
      <div
        class="w-[48px] h-[56px] flex justify-center items-center
         border-1 border-solid"
        :class="iconContainerStyle">
        <i
          class="text-[20px]"
          :class="[icon, iconStyle]"></i>
      </div>
      <!-- 内容 -->
      <div
        class="h-[56px] px-[8px] py-[6px] bg-[#FFFFFF] border-1 border-solid border-l-0 select-none"
        :class="contentStyle">
        <div class="text-[14px] text-[#313238] h-[22px] leading-[22px]">{{ title }}</div>
        <div class="text-[12px] text-[#4D4F56] h-[20px] leading-[20px] mt-[2px]">{{ content }}</div>
      </div>
      <!-- 勾选 -->
      <div class="relative" v-if="checked">
        <div
          class="absolute top-0 w-0 h-0 right-0
          border-[16px] border-transparent border-t-[#3A84FF] border-r-[#3A84FF]">
        </div>
        <i class="nodeman-icon nc-check-small text-[#ffffff] text-[20px] absolute right-0"></i>
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { computed } from 'vue';

const props = defineProps({
  value: {
    type: [String, Number],
  },
  icon: {
    type: String,
  },
  title: {
    type: String,
  },
  content: {
    type: String,
  },
  checked: {
    type: Boolean,
    default: false,
  },
});

const emit = defineEmits(['change']);

const iconContainerStyle = computed(() => (props.checked ? 'bg-[#E1ECFF] border-[#3A84FF]' : 'bg-[#F5F7FA] border-[#C4C6CC]'));
const iconStyle = computed(() => (props.checked ? 'text-[#3A84FF]' : 'text-[#979BA5]'));
const contentStyle = computed(() => (props.checked ? 'border-[#3A84FF]' : 'border-[#C4C6CC]'));

const toggleCheck = () => {
  const curState = !props.checked;
  emit('change', curState, props.value);
};

</script>
