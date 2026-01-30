<template>
  <Dialog
    v-model:is-show="isShow"
    @closed="isShow = false"
  >
    <div>
      <!-- icon -->
      <div class="flex justify-center mt-[27px]">
        <div class="w-[42px] h-[42px] rounded-[50px] bg-[#FCE5C0] text-center flex items-center justify-center">
          <i class="nodeman-icon nc-alert text-[#F59500] text-[22px]"></i>
        </div>
      </div>
      <!-- title -->
      <div class="text-[20px] text-[#313238] text-center font-medium mt-[19px]">
        {{ $t('topoManager.workUnit.delete.title') }}
      </div>
      <!-- content -->
      <div class="flex items-center mt-[16px]">
        <span class="text-[#4D4F56] text-[12px] mr-[5px]">
          {{ $t('topoManager.workUnit.delete.content') }} :
        </span>
        <span class="text-[#313238] text-[14px]">{{ workUnitName }}</span>
      </div>
      <!-- tips -->
      <p
        class="min-h-[46px] bg-[#F5F6FA] rounded-[2px] text-[#4D4F56]
          text-[14px] mt-[16px] mb-[17px] px-[16px] leading-[30px] flex items-center">
        {{ isDirect ? $t('topoManager.workUnit.delete.directUnitTip') : $t('topoManager.workUnit.delete.tips') }}
      </p>
    </div>
    <div class="flex items-center justify-center">
      <Button
        theme="danger"
        class="mr-[9px] w-[87px] h-[32px]"
        @click="handleDeleteWorkUnit"
        :loading="loading">
        {{ $t('action.delete') }}
      </Button>
      <Button class="w-[87px] h-[32px]" @click="isShow = false">
        {{ $t('action.cancel') }}
      </Button>
    </div>
    <template #footer>
    </template>
  </Dialog>
</template>

<script lang="ts" setup>
import { Button, Dialog, Message } from 'bkui-vue';
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';

import { TopoService } from '@/api/modules/topo';

const isShow = defineModel<boolean>('isShow', { default: false });

const props = defineProps({
  workUnitName: {
    type: String,
  },
  workUnitId: {
    type: Number,
  },
  isDirect: {
    type: Boolean,
    default: false,
  },
});
const emit = defineEmits(['delete']);
const { t } = useI18n();

const loading = ref(false);

const handleDeleteWorkUnit = async () => {
  loading.value = true;
  try {
    await TopoService.NetworkUnitDelete({
      bk_networkunit_id: props.workUnitId,
    });
    Message({
      theme: 'success',
      message: t('message.success.delete'),
    });
    emit('delete');
    isShow.value = false;
  } catch (err) {
    console.error(err);
  } finally {
    loading.value = false;
  }
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
