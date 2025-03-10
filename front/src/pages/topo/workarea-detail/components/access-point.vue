<template>
  <div class="text-[12px]">
    <div class="mb-[6px] flex items-center">
      <div class="text-[#4D4F56] w-[72px]">上游接入点：</div>
      <template v-for="(item, index) in upstreamData" :key="index">
        <i v-if="index !== 0" class="nodeman-icon nc-arrows-right text-[#C4C6CC] mx-[5px] text-[24px]"></i>
        <span>{{ item }}</span>
      </template>
    </div>
    <div v-if="isExpand" class="w-[295px] h-[112px] bg-[#F5F7FA] p-[14px] text-[#4D4F56] ml-[72px]">
      <div class="flex items-center h-[20px]">
        <div class="w-[55px] text-right">Cluster：</div>
        <template v-for="(item, index) in upstreamData" :key="index">
          <i v-if="index !== 0" class="nodeman-icon nc-arrows-right text-[#C4C6CC] mx-[5px] text-[24px]"></i>
          <span>{{ item }}</span>
        </template>
      </div>
      <div class="mt-[12px] flex items-center h-[20px]">
        <div class="w-[55px] text-right">File：</div>
        <template v-for="(item, index) in upstreamData" :key="index">
          <i v-if="index !== 0" class="nodeman-icon nc-arrows-right text-[#C4C6CC] mx-[5px] text-[24px]"></i>
          <span>{{ item }}</span>
        </template>
      </div>
      <div class="mt-[12px] flex items-center h-[20px]">
        <div class="w-[55px] text-right">Data：</div>
        <template v-for="(item, index) in upstreamData" :key="index">
          <i v-if="index !== 0" class="nodeman-icon nc-arrows-right text-[#C4C6CC] mx-[5px] text-[24px]"></i>
          <span>{{ item }}</span>
        </template>
      </div>
    </div>

    <div class="flex mt-[14px]">
      <div class="text-[#4D4F56] w-[72px]">下游接入点：</div>
      <span v-if="!isExpand">{{ downstreamData?.map(item => item.name).join(' , ') }}</span>
      <div v-else>
        <div v-for="(item, index) in downstreamData" class="mb-[8px]" :key="index">
          <span>{{ item.name }}</span>
          <div class="w-[295px] h-[112px] bg-[#F5F7FA] p-[14px] text-[#4D4F56] mt-[6px]">
            <div class="flex items-center h-[20px]">
              <div class="w-[55px] text-right">Cluster：</div>
              <span>{{ item.Cluster }}</span>
            </div>
            <div class="flex items-center h-[20px] mt-[12px]">
              <div class="w-[55px] text-right">File</div>
              <span>{{ item.File }}</span>
            </div>
            <div class="flex items-center h-[20px] mt-[12px]">
              <div class="w-[55px] text-right">Data</div>
              <span>{{ item.Data }}</span>
            </div>
          </div>
        </div>
      </div>
      <div class="flex items-end">
        <Button text theme="primary" class="ml-[15px]" @click="toggleExpand">
          <template v-if="!isExpand">
            <span>展开</span>
            <i class="nodeman-icon nc-angle-double-down text-[24px]"></i>
          </template>
          <template v-else>
            <span>收起</span>
            <i class="nodeman-icon nc-double-up text-[24px]"></i>
          </template>
        </Button>
      </div>
    </div>

  </div>
</template>

<script lang="ts" setup>
import { Button } from 'bkui-vue';
import { ref } from 'vue';

defineProps({
  upstreamData: {
    type: Array,
  },
  downstreamData: {
    type: Array,
  },
});
const isExpand = ref(false);
const toggleExpand = () => {
  isExpand.value = !isExpand.value;
};
</script>
