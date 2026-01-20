<template>
  <div class="form-item-content">
    <div
      v-for="item in installTypeList"
      :key="item.type"
      :class="['type', { 'active': activeType === item.type }]"
      @click="handleClick(item.type)"
    >
      <div class="prefix">
        <i :class="['text-[18px]', 'nodeman-icon', item.icon]"></i>
      </div>
      <div
        class="text-[12px] text-[#313238] leading-[19px] ml-[12px] content"
        v-bk-tooltips="{
          content: item.desc,
        }">
        {{ item.name }}
      </div>
      <div class="checked" v-show="activeType === item.type">
        <i class="nodeman-icon nc-check-small"></i>
      </div>
    </div>
  </div>
</template>
<script lang="ts" setup>
import { computed, onMounted } from 'vue';

import { useMainStore } from '@/stores/main';

const props = defineProps({
  needTypeList: {
    type: Array,
    default: () => ['setup', 'import', 'manual'],
  },
  currentNodeType: {
    type: String,
    default: 'agent',
  },
});
const emit = defineEmits(['change']);
const mainStore = useMainStore();
const installTypeConfig = [
  {
    type: 'setup',
    name: '远程安装',
    icon: 'nc-monitor',
    desc: '线上表单填写，需要提供登录信息',
  },
  // {
  //   type: 'import',
  //   name: 'Excel 导入远程安装',
  //   icon: 'nc-excel',
  //   desc: 'Excel 导入填写， 需要提供登录信息',
  // },
  {
    type: 'manual',
    name: '手动安装',
    icon: 'nc-manual',
    desc: '无需提供登录信息，自行在服务器上执行给定命令完成安装',
  },
];
const installTypeList = computed(() => installTypeConfig.filter(item => props.needTypeList.includes(item.type)));
const activeType = computed(() => props.currentNodeType === 'agent' ? mainStore.agentSetupType : mainStore.proxySetupType);
const handleClick = (type: string) => {
  emit('change', type);
  props.currentNodeType === 'agent' ? mainStore.updateAgentSetupType(type) : mainStore.updateProxySetupType(type);
};
onMounted(() => {
  mainStore.updateAgentSetupType('setup');
  mainStore.updateProxySetupType('setup');
});
</script>
<style lang="postcss" scoped>
.form-item-content {
    display: flex;
    align-items: center;
    gap: 8px;
    width: 568px;
    .active {
      border-color: #3A84FF !important;
      &.type {
        border-color: none;
        box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06);;
        .prefix {
          background: #E1ECFF;
          color: #3A84FF;
        }
      }
    }
    .checked {
      position: absolute;
      top: 0;
      right: 0;
      width: 0;
      height: 0;
      border-style: solid;
      border-width: 0 32px 32px 0;
      border-color: transparent #3A84FF transparent transparent;
      .nc-check-small {
          position: absolute;
          font-size: 20px;
          color: #fff;
          top: 0;
          right: -32px;
      }
    }
    .type {
      position: relative;
      width: 50%;
      height: 32px;
      background: #FFFFFF;
      border: 1px solid #C4C6CC;
      border-radius: 2px;
      display: flex;
      align-items: center;
      cursor: pointer;
      &:hover {
        border-color: #3A84FF;
      }
      .prefix {
        min-width: 40px;
        height: 100%;
        border-right: 1px solid #C4C6CC;
        background: #F5F7FA;
        display: flex;
        align-items: center;
        justify-content: center;
        font-size: 21px;
        color: #979BA5;
      }
      .content {
        border-bottom: 1px dashed #c4c6cc;
      }
    }
}
</style>
