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
        {{ proxyCount }}
      </div>
    </div>
    <!-- Agent 数量 -->
    <div class="mr-[53px]">
      <div class="text-[#979BA5] text-[12px]">
        {{ $t('topoManager.workAreaDetail.agentQuantity') }}
      </div>
      <div class="text-[12px] mt-[4px]">
        {{ agentCount }}
      </div>
    </div>
    <!-- 更新人 -->
    <div class="mr-[53px]">
      <div class="text-[#979BA5] text-[12px]">
        {{ $t('topoManager.workAreaDetail.updatePerson') }}
      </div>
      <div class="text-[12px] mt-[4px]">
        {{ updatePerson }}
      </div>
    </div>
    <!-- 更新时间 -->
    <div class="mr-[53px]">
      <div class="text-[#979BA5] text-[12px]">
        {{ $t('topoManager.workAreaDetail.updateTime') }}
      </div>
      <div class="text-[12px] mt-[4px]">
        {{ updateTime }}
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
defineProps({
  workUnitCount: {
    type: Number,
    default: 0,
  },
  proxyCount: {
    type: Number,
    default: 0,
  },
  agentCount: {
    type: Number,
    default: 0,
  },
  updatePerson: {
    type: String,
    default: '--',
  },
  updateTime: {
    type: String,
    default: '--',
  },
});
</script>
