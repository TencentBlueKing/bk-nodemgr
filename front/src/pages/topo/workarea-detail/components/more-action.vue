<template>
  <div>
    <Dropdown
      trigger="click"
      :placement="placement"
      :popover-options="{
        clickContentAutoHide: true,
      }">
      <Button
        text
        :loading="crossPageSelectLoading">
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
import { TopoService } from '@/api/modules/topo';

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
  isCrossPageSelection: {
    type: Boolean,
    default: false,
  },
  crossPageQueryParams: {
    type: Object,
    default: () => ({}),
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

// 获取跨页全选的host_id数据
const crossPageSelectLoading = ref(false);
const crossPageHostIdData = ref<number[]>([]);
const getCorssPageHostIds = async () => {
  try {
    crossPageSelectLoading.value = true;
    const res = await TopoService.HostSelectHostID(props.crossPageQueryParams);
    crossPageHostIdData.value = res.items;
  } catch (error) {
    console.error('获取跨页全选数据失败:', error);
  } finally {
    crossPageSelectLoading.value = false;
  }
};

// 选择dropMenuItem，打开对应的action dialog，关闭dropdown
const handleClickDropMenu = async (action: keyof typeof confirmConfigMap) => {
  if (action === 'reinstall') {
    emit('reinstall');
  } else {
    let operateData = props.data;
    let batch = props.batch;
    if (props.isCrossPageSelection) {
      await getCorssPageHostIds();
      operateData = crossPageHostIdData.value;
      batch = true; // 强制设置为批量模式
    }
    const titleObj = {
      firstIp: props.data[0].bk_host_innerip,
      num: operateData.length,
    };
    if (action === 'upgrade') {
      chooseVersionData.title = 'Proxy 升级/回退';
      chooseVersionData.isShow = true;
      chooseVersionData.data = operateData;
      chooseVersionData.batch = batch;
    } else if (action === 'restart') {
      operateDialogIsShow.value = true;
      operateDialogData.type = action;
      operateDialogData.title = batch ? '请确认是否批量重启' : '请确认是否重启';
      operateDialogData.subTitle = batch
        ? `重启 ${titleObj.firstIp} 等${titleObj.num}个IP的Proxy`
        : `重启 ${titleObj.firstIp} 的Proxy`;
    } else if (action === 'unload') {
      InfoBox({
        title: batch ? '请确认是否批量卸载' : '请确认是否卸载',
        subTitle: batch
          ? `卸载 ${titleObj.firstIp} 等${titleObj.num}个IP的Agent`
          : `卸载 ${titleObj.firstIp} 的Agent`,
        onConfirm: () => {
          handleUninstall();
        },
      });
    }
  }
};

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
  let operateData = props.data;
  if (props.isCrossPageSelection) {
    operateData = crossPageHostIdData.value;
  }
  const result = await NodeProxyService.NodeProxyUninstall({
    host: operateData?.map((item: any) => ({
      bk_host_id: item.bk_host_id,
    })),
  }).catch(() => ({
    workflow_id: '',
  }));
  if (result.workflow_id) {
    router.push({
      name: 'taskDetail',
      params: { taskId: result.workflow_id, routerBackName: 'taskList' },
      query: {
        active: 'node',
      },
    });
  }
};
const operateJob = async (extraData: any = {}) => {
  let operateData = props.data;
  if (props.isCrossPageSelection) {
    operateData = crossPageHostIdData.value;
  }
  const params = {
    host: operateData?.map((item: any) => ({
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
  if (result.workflow_id) {
    router.push({
      name: 'taskDetail',
      params: { taskId: result.workflow_id, routerBackName: 'taskList' },
      query: {
        active: 'node',
      },
    });
  }
};
// 升级回退
const handleUpgrade = async (versionList: any[]) => {
  let operateData = props.data;
  if (props.isCrossPageSelection) {
    operateData = crossPageHostIdData.value;
  }
  const params = {
    host: operateData?.map(item => ({
      bk_host_id: item.bk_host_id,
      force: false,
      graceful_restart_timeout_sec: 0,
    })),
    target_version: versionList,
  };
  const result = await NodeProxyService.NodeProxyUpgrade(params).catch(() => ({
    workflow_id: '',
  }));
  if (result.workflow_id) {
    router.push({
      name: 'taskDetail',
      params: { taskId: result.workflow_id, routerBackName: 'taskList' },
      query: {
        active: 'node',
      },
    });
  }
};
</script>
