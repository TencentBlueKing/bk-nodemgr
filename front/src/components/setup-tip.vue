<template>
  <div
    class="mx-[24px] flex min-h-[56px] bg-[#F0F8FF] border border-[#C5DAFF] rounded-[2px] py-[6px] px-[9px] gap-[9px]"
  >
    <i class="nodeman-icon nc-tips pt-[2px] text-[#3A84FF]"></i>
    <div class="text-[#4d4f56] text-[12px] leading-[20px] text-left">
      <i18n-t keypath="platform.nodeMan.installAgentPage.tip1" tag="p">
        <span class="text-[#313238] font-bold">{{
          $t("platform.nodeMan.installAgentPage.tip1FirstSlotText")
        }}</span>
        <span class="text-[#313238] font-bold">{{
          $t("platform.nodeMan.installAgentPage.tip1SecondSlotText")
        }}</span>
      </i18n-t>
      <i18n-t keypath="platform.nodeMan.installAgentPage.tip2" tag="p">
        <span class="text-[#313238] font-bold">{{
          $t("platform.nodeMan.installAgentPage.tip2FirstSlotText")
        }}</span>
        <Button text theme="primary" @click="handleShowPanel">{{
          $t("platform.nodeMan.installAgentPage.tip2SecondSlotText")
        }}</Button>
        <Button text theme="primary" @click="handleShowSetting">{{
          $t("platform.nodeMan.installAgentPage.tip2ThirdSlotText")
        }}</Button>
      </i18n-t>
    </div>
  </div>
</template>
<script setup lang="ts">
import { ref } from 'vue';
import { Button } from 'bkui-vue';

const showRightPanel = ref(false);
// 显示侧边栏安装策略
const handleShowPanel = () => {
  showRightPanel.value = true;
};
// 显示表格设置
const handleShowSetting = () => {};
</script>
