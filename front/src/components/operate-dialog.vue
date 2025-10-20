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
        <Checkbox v-model="isReconfig">同时重载配置</Checkbox>
        <div class="flex items-center gap-[6px] h-[32px]">
          <Checkbox v-model="isForce" class="w-[100px]">强制重启</Checkbox>
          <div class="flex items-center gap-[3px]" v-show="!isForce">
            <span>时间</span>
            <Input type="number" v-model="time" class="w-[80px]"></Input>
            <span>秒</span>
          </div>
        </div>
      </template>
    </div>
  </Dialog>
</template>
<script lang="ts" setup>
import { Checkbox, Dialog, Input } from 'bkui-vue';
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
