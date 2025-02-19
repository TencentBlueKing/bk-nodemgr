<template>
  <div v-if="title" class="flex items-center h-[52px] shadow px-[24px] bg-[#fff]">
    <ArrowsLeft width="24" height="24" class="text-[#3a84ff] cursor-pointer" v-if="back" @click="handleBack" />
    <span>{{ $t(title) }}</span>
    <slot></slot>
  </div>
</template>
<script lang="ts" setup>
import { ArrowsLeft } from 'bkui-vue/lib/icon';
import { useRouter } from 'vue-router';

interface Props {
  title?: string
  back?: boolean
}

withDefaults(defineProps<Props>(), {
  back: true,
});

const router = useRouter();
function handleBack() {
  router.back();
}
</script>
