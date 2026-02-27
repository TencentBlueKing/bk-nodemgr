<template>
  <div
    class="mx-[24px] flex min-h-[56px] bg-[#F0F8FF] border border-[#C5DAFF] rounded-[2px] py-[6px] px-[9px] gap-[9px]"
  >
    <i class="nodeman-icon nc-tips pt-[2px] text-[#3A84FF]"></i>
    <div class="text-[#4d4f56] text-[12px] leading-[20px] text-left whitespace-nowrap">
      <i18n-t keypath="platform.nodeMan.installAgentPage.tip1" tag="p">
        <span class="text-[#313238] font-bold">{{
          $t("platform.nodeMan.installAgentPage.tip1FirstSlotText")
        }}</span>
        <span class="text-[#313238] font-bold">{{
          $t("platform.nodeMan.installAgentPage.tip1SecondSlotText")
        }}</span>
      </i18n-t>

    </div>
  </div>
</template>
<script setup lang="ts">
import { Button } from 'bkui-vue';
import { ref } from 'vue';

defineProps({
  showInstallRequirement: {
    type: Boolean,
    default: false,
  },
});
const emit = defineEmits(['showSetting']);
const showRightPanel = ref(false);
// 显示侧边栏安装策略
const handleShowPanel = () => {
  showRightPanel.value = true;
};
// 显示表格设置
const handleShowSetting = () => {
  emit('showSetting');
};
</script>
