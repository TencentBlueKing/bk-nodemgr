<template>
  <div class="absolute bottom-[50px] right-[50px]">
    <div id="minimap-container" class="mb-[4px]" v-show="mapShow"></div>
    <div class="flex items-center p-[4px] text-[#000] bg-[#FFFFFF]">
      <div
        class="nodemgr-icon nodemgr-icon-minimap p-[7px] mr-[4px] hover:bg-[#F0F1F5] cursor-pointer"
        :class="mapShow ? 'bg-[#E1ECFF]': ''"
        v-bk-tooltips="{
          content: '地图'
        }"
        @click="handleClickTool('map')"></div>
      <div
        class="nodemgr-icon nodemgr-icon-return-center p-[7px] hover:bg-[#F0F1F5] cursor-pointer"
        v-bk-tooltips="{
          content: '回中'
        }"
        @click="handleClickTool('center')"></div>
      <div class="mx-[4px] bg-[#EAEBF0] h-[24px] w-[1px] border-l"></div>
      <div
        class="nodemgr-icon nodemgr-icon-proportion p-[7px] mx-[4px] hover:bg-[#F0F1F5] cursor-pointer"
        v-bk-tooltips="{
          content: '原始大小'
        }"
        @click="handleClickTool('init')"></div>
      <!-- 地图缩放 -->
      <div class="flex">
        <div
          class="nodemgr-icon nodemgr-icon-minus-line p-[7px] hover:bg-[#F0F1F5] cursor-pointer"
          @click="handleClickTool('zoomIn')"></div>
        <Slider v-model="mapSize" class="w-[64px] mx-[13px]" :max-value="100" :min-value="0"></Slider>
        <div
          class="nodemgr-icon nodemgr-icon-plus-line p-[7px] hover:bg-[#F0F1F5] cursor-pointer"
          @click="handleClickTool('zoomOut')"></div>
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { Slider } from 'bkui-vue';
import { ref, watch, PropType } from 'vue';
import { throttle } from 'lodash';

const emit = defineEmits(['triggerTools']);
type graphZoomRange = [number, number];
const props = defineProps({
  zoomRange: {
    type: Array as unknown as PropType<graphZoomRange>,
    default: [0.2, 1.5] // graph 缩放区间默认为[0.2, 1.5]
  },
  zoomSpeed: {
    type: Number,
    default: 1.01 // 可配置的缩放速度
  },
  mapSize: { // 初始缩放大小
    type: Number,
    default: 50,
  },
})

const mapSize = ref(props.mapSize);
const mapShow = ref(false);
const handleClickTool = (code: string) => {
  switch (code) {
    // 基于mapSize自动计算ratio
    case 'zoomIn':
      mapSize.value = Math.min(100, mapSize.value - 10);
      break;
    case 'zoomOut':
      mapSize.value = Math.min(100, mapSize.value + 10);
      break;
    case 'init':
      mapSize.value = props.mapSize;
      break;
    case 'map':
      mapShow.value = !mapShow.value;
      break;
    case 'center':
      emit('triggerTools', code);
      break;
  };
};

watch(mapSize, throttle((newVal, oldVal) => {
  // 限制值的范围在0-100之间
  if (newVal > 100 || newVal < 0) return;
  const boundedNewVal = Math.min(100, Math.max(0, newVal));
  
  // 直接将滑动条的值(0-100)映射到缩放范围
  const minZoom = props.zoomRange[0];
  const maxZoom = props.zoomRange[1];
  const ratio = minZoom + (boundedNewVal / 100) * (maxZoom - minZoom);
  emit('triggerTools', 'mapSize', ratio);
}, 20, { leading: true, trailing: true }));
</script>