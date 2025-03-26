<template>
  <Dialog :is-show="isShow" :width="480" @closed="isShow = false">
    <div class="pt-[11px]">
      <!-- icon -->
      <div class="flex justify-center">
        <div class="w-[42px] h-[42px] rounded-[50px] bg-[#E5F6EA] flex items-center justify-center">
          <i class="nodeman-icon nc-check-small text-[#3FC06D] text-[36px]"></i>
        </div>
      </div>
      <!-- result text -->
      <div class="text-[20px] text-[#313238] text-center font-medium mt-[19px]">
        {{ $t('topoManager.workArea.form.success') }}
      </div>
      <!-- tips -->
      <div
        class="w-[416px] h-[46px] bg-[#F5F6FA] rounded-[2px] text-[#4D4F56]
          text-[14px] mt-[16px] mb-[26px] pl-[16px] leading-[46px]">
        {{ $t('topoManager.workArea.form.tips') }}
      </div>
      <!-- footer -->
      <div class="flex justify-center">
        <Button theme="primary" class="mr-[8px]" @click="handleInstallProxy">{{ $t('action.installProxy') }}</Button>
        <Button class="text-[#63656E]" @click="handleInstallLater">{{ $t('action.installLater') }}</Button>
      </div>
    </div>
    <template #footer></template>
  </Dialog>
</template>

<script lang="ts" setup>
import { Button, Dialog } from 'bkui-vue';
const isShow = defineModel('isShow', { type: Boolean, default: false });

const emit = defineEmits(['confirm', 'cancel']);

const handleInstallProxy = () => {
  emit('confirm');
  isShow.value = false;
};

const handleInstallLater = () => {
  emit('cancel');
  isShow.value = false;
};

</script>

<style lang="less" scoped>
:deep(.bk-dialog-header) {
  display: none;
}
:deep(.bk-dialog-footer) {
  display: none;
}
</style>
