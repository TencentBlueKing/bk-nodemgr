<template>
  <div v-if="title" class="flex items-center h-[52px] shadow px-[24px] bg-[#fff]">
    <ArrowsLeft width="24" height="24" class="text-[#3a84ff] cursor-pointer" v-if="back" @click="handleBack" />
    <span>{{ title }}</span>
    <slot name="after-title"></slot>
    <div v-if="routeSubTitle.subTitle" class="flex items-center">
      <div class="w-[1px] h-[14px] bg-[#DCDEE5] mx-[16px]"></div>
      <span class="text-[12px] text-[#979BA5]">{{ routeSubTitle.subTitle }}</span>
    </div>
    <slot></slot>
  </div>
</template>
<script lang="ts" setup>
import { ArrowsLeft } from 'bkui-vue/lib/icon';
import { useRouter } from 'vue-router';

import { useRouteSubTitle } from '@/stores/route-sub-title';
const props = withDefaults(defineProps<Props>(), {
  back: true,
  onBack: undefined,
});

const routeSubTitle = useRouteSubTitle();

interface Props {
  title?: string
  subTitle?: string
  back?: boolean
  onBack?: () => void
}

const router = useRouter();
function handleBack() {
  if (props.onBack) {
    props.onBack();
  } else {
    router.back();
  }
}
</script>
