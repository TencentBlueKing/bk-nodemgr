<template>
  <div>
    <Dropdown trigger="manual" placement="right-start" :is-show="isShowDropdown">
      <Button
        text
        @click="isShowDropdown = true"
        @blur="isShowDropdown = false">
        <i class="nodeman-icon nc-more cursor"></i>
      </Button>
      <template #content>
        <Dropdown.DropdownMenu>
          <Dropdown.DropdownItem
            v-for="(item, index) in dropMenuList"
            :key="index"
            @mousedown="handleClickDropMenu(item.value)"
          >
            {{ item.label }}
          </Dropdown.DropdownItem>
        </Dropdown.DropdownMenu>
      </template>
    </Dropdown>
    <Dialog
      v-model:is-show="isShow"
      @closed="isShow = false"
      :width="curWidth"
    >
      <div>
        <!-- title -->
        <div class="text-[20px] text-[#313238] text-center font-medium mt-[40px]">
          {{ actionConfirmProps.title }}
        </div>
        <!-- content -->
        <div class="flex items-center mt-[16px]" :class="`justify-${actionConfirmProps.contentPosition}`">
          <span class="text-[#4D4F56] text-[12px] mr-[5px]">
            {{ $t('topoManager.workAreaDetail.dialogContent.ipv4') }} :
          </span>
          <span class="text-[#313238] text-[14px]">{{ actionConfirmProps.value }}</span>
        </div>
        <!-- tips -->
        <div
          v-if="actionConfirmProps.tips"
          class="w-[416px] h-[46px] bg-[#F5F6FA] rounded-[2px] text-[#4D4F56]
            text-[14px] mt-[16px] pl-[16px] leading-[46px]">
          {{ actionConfirmProps.tips }}
        </div>
      </div>
      <div class="flex items-center justify-center mt-[24px]">
        <Button
          :theme="actionConfirmProps.theme"
          class="mr-[9px] w-[87px] h-[32px]"
          @click="handleConfirm"
          :loading="loading">
          {{ actionConfirmProps.confirmText }}
        </Button>
        <Button class="w-[87px] h-[32px]" @click="isShow = false">
          {{ $t('action.cancel') }}
        </Button>
      </div>
      <template #footer>
      </template>
    </Dialog>
  </div>
</template>

<script lang="ts" setup>
/**
 * 1. 封装DropMenuList数据 + 对应操作dialog，简化table代码，专注操作逻辑
 * 2. Dropdown只能通过trigger="manual"控制isShow，table中有多个DropMenu
 *    拆分成组件能轻松隔绝click造成所有DropMenu打开的问题
 *    搭配focusout能轻松关闭上一个open的DropMenu，实在巧妙！
 */

import { Button, Dialog, Dropdown } from 'bkui-vue';
import { computed, ref } from 'vue';
import { useI18n } from 'vue-i18n';

interface DialogProps {
  title: string
  value: string
  tips?: string
  theme?: 'primary' | 'danger'
  contentPosition?: 'start' | 'center' | 'end'
  confirmText?: string
};

interface IProps {
  ipv4: string
}
const props = defineProps<IProps>();
const { t } = useI18n();
// dropMenuList
const dropMenuList = ref<{
  label: string
  value: keyof typeof confirmConfigMap
}[]>([
  {
    label: t('topoManager.workAreaDetail.dropdown.upgrade'),
    value: 'upgrade',
  },
  {
    label: t('topoManager.workAreaDetail.dropdown.unload'),
    value: 'unload',
  },
  {
    label: t('topoManager.workAreaDetail.dropdown.overload'),
    value: 'overload',
  },
  {
    label: t('topoManager.workAreaDetail.dropdown.restart'),
    value: 'restart',
  },
]);
// action dialog map
const confirmConfigMap = {
  upgrade: {
    title: t('topoManager.workAreaDetail.dialogTitle.upgrade'),
    tips: '',
    theme: 'primary',
    contentPosition: 'center',
    confirmText: t('action.confirm1'),
  },
  unload: {
    title: t('topoManager.workAreaDetail.dialogTitle.unload'),
    tips: t('topoManager.workAreaDetail.dialogTips.unload'),
    theme: 'danger',
    contentPosition: 'start',
    confirmText: t('action.unload'),
  },
  overload: {
    title: t('topoManager.workAreaDetail.dialogTitle.overload'),
    tips: '',
    theme: 'primary',
    contentPosition: 'center',
    confirmText: t('action.confirm1'),
  },
  restart: {
    title: t('topoManager.workAreaDetail.dialogTitle.restart'),
    tips: '',
    theme: 'primary',
    contentPosition: 'center',
    confirmText: t('action.confirm1'),
  },
} as const;

// show dropMenu
const isShowDropdown = ref(false);
// show action dialog
const isShow = ref(false);
// 根据 curAction 匹配对应的dialog props
const curAction = ref<keyof typeof confirmConfigMap>('upgrade');
const actionConfirmProps = computed((): DialogProps => ({
  ...confirmConfigMap[curAction.value],
  value: props.ipv4,
}));

// 选择dropMenuItem，打开对应的action dialog，关闭dropdown
const handleClickDropMenu = (action: keyof typeof confirmConfigMap) => {
  curAction.value = action;
  // 打开dialog
  isShow.value = true;
  // 隐藏dropdown
  isShowDropdown.value = false;
};

// dialog action loading
const loading = ref(false);
// dialog width
const curWidth = computed(() => (actionConfirmProps.value.theme === 'primary' ? 400 : 480));

// // 升级
// const upgradeVersion = async () => {

// };
// // 卸载
// const unloadProxy = async () => {

// };
// // 重载配置
// const overloadConfig = async () => {

// };
// // 重启
// const reloadProxy = async () => {

// };

const handleConfirm = async () => {
  loading.value = true;
  try {
  } catch (err) {
    console.error(err);
  } finally {
    loading.value = false;
    isShow.value = false;
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
