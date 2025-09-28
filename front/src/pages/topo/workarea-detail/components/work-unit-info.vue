<template>
  <!-- 基本信息 -->
  <div class="h-[44px] flex mb-[24px]">
    <!-- 管控单元数量 -->
    <div class="mr-[53px]">
      <div class="text-[#979BA5] text-[12px]">
        {{ $t('topoManager.workAreaDetail.workUnitCount') }}
      </div>
      <div class="text-[12px] mt-[4px]">
        {{ workUnitCount }}
      </div>
    </div>
    <!-- Proxy 数量 -->
    <div class="mr-[53px]">
      <div class="text-[#979BA5] text-[12px]">
        {{ $t('topoManager.workAreaDetail.proxyQuantity') }}
      </div>
      <div class="text-[12px] mt-[4px]">
        {{ info?.proxy_count || 0 }}
      </div>
    </div>
    <!-- Agent 数量 -->
    <div class="mr-[53px]">
      <div class="text-[#979BA5] text-[12px]">
        {{ $t('topoManager.workAreaDetail.agentQuantity') }}
      </div>
      <div class="text-[12px] mt-[4px]">
        {{ info?.agent_count || 0 }}
      </div>
    </div>
    <!-- 更新人 -->
    <div class="mr-[53px]">
      <div class="text-[#979BA5] text-[12px]">
        {{ $t('topoManager.workAreaDetail.updatePerson') }}
      </div>
      <div class="text-[12px] mt-[4px]">
        {{ info?.last_operator || '--'}}
      </div>
    </div>
    <!-- 更新时间 -->
    <div class="mr-[53px]">
      <div class="text-[#979BA5] text-[12px]">
        {{ $t('topoManager.workAreaDetail.updateTime') }}
      </div>
      <div class="text-[12px] mt-[4px]">
        {{ info?.last_operate_time > 0 ? formatTimestamp(info.last_operate_time) : '--'}}
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { computed } from 'vue';
import { useRoute } from 'vue-router';
import { useWorkareaStore } from '@/stores/workarea';
import { formatTimestamp } from '@/common/util';

defineProps({
  workUnitCount: {
    type: Number,
    default: 0,
  },
});

const workareaStore = useWorkareaStore();
const route = useRoute();
const info = computed(() => workareaStore.workareaList.find(item =>
  item.bk_networkarea_id === Number(route.params.workarea)));
</script>
