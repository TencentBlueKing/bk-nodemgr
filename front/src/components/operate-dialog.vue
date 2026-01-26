<template>
  <Dialog
    :is-show="isShow"
    :width="700"
    :title="title"
    @closed="isShow = false"
    @confirm="handleConfirm"
    @cancel="handleCancel"
  >
    <div class="w-full">
      <p class="text-[16px] mb-[20px]">{{ subTitle }}</p>
      <template v-if="type === 'restart'">
        <Checkbox v-model="isReconfig" class="mb-[10px]">
          {{ $t('components.operateDialog.reloadConfig') }}
        </Checkbox>
        <div class="flex items-center gap-[6px] h-[32px]">
          <Radio.Group v-model="isForce" class="w-[180px]">
            <Radio.Button :label="true" :key="true">
              {{ $t('components.operateDialog.forceRestart') }}
            </Radio.Button>
            <Radio.Button :label="false" :key="false">
              {{ $t('components.operateDialog.gracefulRestart') }}
            </Radio.Button>
          </Radio.Group>
          <div class="flex items-center gap-[3px]" v-show="!isForce">
            <span>{{ $t('components.operateDialog.gracefulTime') }}</span>
            <Input type="number" v-model="time" class="w-[80px]"></Input>
            <span>{{ $t('components.operateDialog.seconds') }}</span>
          </div>
        </div>
      </template>
    </div>
  </Dialog>
</template>
<script lang="ts" setup>
import { Checkbox, Dialog, Input, Radio } from 'bkui-vue';
import { computed, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';

const isShow = defineModel('isShow', { type: Boolean, default: false });
const props = defineProps({
  type: {
    type: String,
    default: '',
  },
  title: {
    type: String,
    default: '',
  },
  subTitle: {
    type: String,
    default: '',
  },
});
const emit = defineEmits(['confirm', 'cancel']);
const { t } = useI18n();

const isReconfig = ref(true);
const isForce = ref(false);
const time = ref(120);

function handleConfirm() {
  emit('confirm', { isReconfig: isReconfig.value, isForce: isForce.value, time: time.value });
  isShow.value = false;
}

function handleCancel() {
  isShow.value = false;
}

watch(
  () => isShow,
  async () => {
    if (isShow.value) {

    }
  },
  { immediate: true, deep: true },
);
</script>
