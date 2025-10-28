<template>
  <div>
    <Dropdown trigger="click" :placement="placement" :is-show="isShowDropdown">
      <Button
        text
        @click="isShowDropdown = true"
        @blur="isShowDropdown = false">
        <slot></slot>
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
    <choose-version-dialog
      v-model:is-show="chooseVersionData.isShow"
      :title="chooseVersionData.title"
      :data="chooseVersionData.data"
      :batch="chooseVersionData.batch"
      release-type="proxy"
      @confirm="handleUpgrade"
    >
    </choose-version-dialog>
    <operate-dialog
      v-model:is-show="operateDialogIsShow"
      :title="operateDialogData.title"
      :type="operateDialogData.type"
      :sub-title="operateDialogData.subTitle"
      @confirm="operateJob"
    ></operate-dialog>
  </div>
</template>

<script lang="ts" setup>
/**
 * 1. 封装DropMenuList数据 + 对应操作dialog，简化table代码，专注操作逻辑
 * 2. Dropdown只能通过trigger="manual"控制isShow，table中有多个DropMenu
 *    拆分成组件能轻松隔绝click造成所有DropMenu打开的问题
 *    搭配focusout能轻松关闭上一个open的DropMenu，实在巧妙！
 */

import { Button, Dropdown, InfoBox } from 'bkui-vue';
import type { PropType } from 'vue';
import { computed, reactive, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';

import { NodeProxyService } from '@/api/modules/node_proxy';

interface DialogProps {
  title: string
  value: string
  tips?: string
  theme?: 'primary' | 'danger'
  contentPosition?: 'start' | 'center' | 'end'
  confirmText?: string
};

const props = defineProps({
  placement: {
    type: String,
    default: 'right-start',
  },
  ipv4: {
    type: String,
    default: '',
  },
  data: {
    type: Array as PropType<Host[]>,
    default: [],
  },
  batch: {
    type: Boolean,
    default: false,
  },
});
const emit = defineEmits(['reinstall']);
const { t } = useI18n();
const router = useRouter();
// dropMenuList
const dropMenuList = ref<{
  label: string
  value: keyof typeof confirmConfigMap
}[]>([
  {
    label: t('topoManager.workAreaDetail.table.Reassembly'),
    value: 'reinstall',
  },
  {
    label: t('topoManager.workAreaDetail.dropdown.upgrade'),
    value: 'upgrade',
  },
  {
    label: t('topoManager.workAreaDetail.dropdown.restart'),
    value: 'restart',
  },
  {
    label: t('topoManager.workAreaDetail.dropdown.unload'),
    value: 'unload',
  },
]);
// action dialog map
const confirmConfigMap = {
  reinstall: {
    title: t('topoManager.workAreaDetail.table.Reassembly'),
    tips: t('topoManager.workAreaDetail.table.Reassembly'),
    theme: 'primary',
    contentPosition: 'center',
  },
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

// 选择dropMenuItem，打开对应的action dialog，关闭dropdown
const handleClickDropMenu = (action: keyof typeof confirmConfigMap) => {
  if (action === 'reinstall') {
    emit('reinstall');
  } else {
    const titleObj = {
      firstIp: props.data[0].info.bk_host_innerip,
      num: props.data.length,
    };
    if (action === 'upgrade') {
      chooseVersionData.title = 'Proxy 升级/回退';
      chooseVersionData.isShow = true;
      chooseVersionData.data = props.data;
      chooseVersionData.batch = props.batch;
    } else if (action === 'restart') {
      operateDialogIsShow.value = true;
      operateDialogData.type = action;
      operateDialogData.title = props.batch ? '请确认是否批量重启' : '请确认是否重启';
      operateDialogData.subTitle = props.batch
        ? `重启 ${titleObj.firstIp} 等${titleObj.num}个IP的Proxy`
        : `重启 ${titleObj.firstIp} 的Proxy`;
    } else if (action === 'unload') {
      InfoBox({
        title: props.batch ? '请确认是否批量卸载' : '请确认是否卸载',
        subTitle: props.batch
          ? `卸载 ${titleObj.firstIp} 等${titleObj.num}个IP的Agent`
          : `卸载 ${titleObj.firstIp} 的Agent`,
        onConfirm: () => {
          handleUninstall();
        },
      });
    }
  }
  // 隐藏dropdown
  isShowDropdown.value = false;
};

// dialog action loading
const loading = ref(false);
// dialog width
const curWidth = computed(() => (actionConfirmProps.value.theme === 'primary' ? 400 : 480));
const chooseVersionData = reactive({
  title: '',
  isShow: false,
  data: null,
  batch: false,
});
const operateDialogIsShow = ref(false);
const operateDialogData = {
  type: '',
  title: '',
  subTitle: '',
};
// 卸载
const handleUninstall = async () => {
  loading.value = true;
  const result = await NodeProxyService.NodeProxyUninstall({
    host: props.data?.map((item: any) => ({
      bk_host_id: item.bk_host_id,
    })),
  }).catch(() => ({
    workflow_id: '',
  }));
  loading.value = false;
  if (result.workflow_id) {
    router.push({
      name: 'taskDetail',
      params: { taskId: result.workflow_id, routerBackName: 'taskList' },
    });
  }
};
const operateJob = async (extraData: any = {}) => {
  loading.value = true;
  const params = {
    host: props.data?.map((item: any) => ({
      bk_host_id: item.bk_host_id,
      force: extraData.isForce,
      graceful_restart_timeout_sec: extraData.time,
    })),
  };
  let result;
  if (extraData.isReconfig) {
    // 重载配置
    result = await NodeProxyService.NodeProxyReconfig(params).catch(() => ({
      workflow_id: '',
    }));
  } else {
    // 重启
    result = await NodeProxyService.NodeProxyRestart(params).catch(() => ({
      workflow_id: '',
    }));
  }
  loading.value = false;
  if (result.workflow_id) {
    router.push({
      name: 'taskDetail',
      params: { taskId: result.workflow_id, routerBackName: 'taskList' },
    });
  }
};
// 升级回退
const handleUpgrade = async (versionObj: any) => {
  loading.value = true;
  const params = {
    host: props.data.map(item => ({
      bk_host_id: item.bk_host_id,
      force: false,
      graceful_restart_timeout_sec: 0,
    })),
    target_version: [
      {
        version: versionObj.version === 'auto' ? '' : versionObj.version,
        cpu_arch: versionObj.cpu_arch,
        os_type: versionObj.os_type,
      },
    ],
  };
  const result = await NodeProxyService.NodeProxyUpgrade(params).catch(() => ({
    workflow_id: '',
  }));
  loading.value = false;
  if (result.workflow_id) {
    router.push({
      name: 'taskDetail',
      params: { taskId: result.workflow_id, routerBackName: 'taskList' },
    });
  }
};
</script>
