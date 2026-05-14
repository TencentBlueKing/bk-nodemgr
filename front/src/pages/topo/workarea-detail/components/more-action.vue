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
        <Dropdown.DropdownMenu ext-cls="proxy-action-dropdown-menu">
          <Dropdown.DropdownItem
            v-for="(item, index) in dropMenuList"
            :key="index"
            :disabled="getItemDisabled(item).disabled"
            :class="{ 'auth-lock-dropdown-item': !hasAuth, 'operate-item-disabled': getItemDisabled(item).disabled }"
            v-bk-tooltips="{
              content: getItemDisabled(item).tooltip,
              disabled: !getItemDisabled(item).disabled,
            }"
            @mousedown="!getItemDisabled(item).disabled && handleClickDropMenu(item.value)"
            @mouseenter="!hasAuth && emit('authLockEnter', $event)"
            @mousemove="!hasAuth && emit('authLockMove', $event)"
            @mouseleave="!hasAuth && emit('authLockLeave')"
          >
            {{ item.label }}
          </Dropdown.DropdownItem>
        </Dropdown.DropdownMenu>
      </template>
    </Dropdown>
    <upgrade-sideslider
      v-model:is-show="upgradePreviewData.isShow"
      :hosts="upgradePreviewData.hosts"
      release-type="proxy"
    />
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

import { isNetworkUnitAssigned } from '@/common/const';
import { NodeProxyService } from '@/api/modules/node_proxy';
import { TopoService } from '@/api/modules/topo';
import UpgradeSideslider from '@/pages/node/agent/upgrade-sideslider.vue';

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
  hasAuth: {
    type: Boolean,
    default: true,
  },
});
const emit = defineEmits(['reinstall', 'assignUnit', 'authClick', 'authLockEnter', 'authLockMove', 'authLockLeave']);
const { t } = useI18n();
const router = useRouter();

// ===== 权限控制：锁图标由父组件通过 props.authLockHandlers 传入 =====
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
  {
    label: t('topoManager.workAreaDetail.dropdown.assignUnit'),
    value: 'assignUnit',
  },
]);

// Item disabled logic based on network unit assignment
const getItemDisabled = (item: { value: string }): { disabled: boolean; tooltip: string } => {
  const data = props.data;
  if (data.length === 0) return { disabled: false, tooltip: '' };

  if (props.batch) {
    // Batch mode: check across all selected hosts
    if (item.value === 'assignUnit') {
      const hasAssigned = data.some((h: any) => isNetworkUnitAssigned(h.bk_networkunit_id));
      if (hasAssigned) {
        return {
          disabled: true,
          tooltip: t('platform.nodeMan.proxyStatus.assignUnitDisabledAssigned'),
        };
      }
      const areaIds = new Set(data.map((h: any) => h.bk_networkarea_id));
      if (areaIds.size > 1) {
        return {
          disabled: true,
          tooltip: t('platform.nodeMan.proxyStatus.assignUnitDisabledMultiArea'),
        };
      }
    }

    if (item.value === 'upgrade' || item.value === 'restart' || item.value === 'unload') {
      const hasNotRunning = data.some((h: any) => h.node_status !== 'running');
      if (hasNotRunning) {
        return {
          disabled: true,
          tooltip: t('platform.nodeMan.proxyStatus.operateDisabledNotRunning'),
        };
      }
      const hasUnassigned = data.some((h: any) => !isNetworkUnitAssigned(h.bk_networkunit_id));
      if (hasUnassigned) {
        return {
          disabled: true,
          tooltip: t('platform.nodeMan.proxyStatus.operateDisabledUnassigned'),
        };
      }
    }
  } else {
    // Row mode: check single host
    if (item.value === 'assignUnit') {
      const row = data[0] as any;
      if (isNetworkUnitAssigned(row?.bk_networkunit_id)) {
        return {
          disabled: true,
          tooltip: t('platform.nodeMan.proxyStatus.assignUnitDisabledRowAssigned'),
        };
      }
    }

    if (item.value === 'upgrade' || item.value === 'restart' || item.value === 'unload') {
      const row = data[0] as any;
      if (row?.node_status !== 'running') {
        return {
          disabled: true,
          tooltip: t('platform.nodeMan.proxyStatus.operateDisabledRowNotRunning'),
        };
      }
      if (!isNetworkUnitAssigned(row?.bk_networkunit_id)) {
        return {
          disabled: true,
          tooltip: t('platform.nodeMan.proxyStatus.operateDisabledRowUnassigned'),
        };
      }
    }
  }

  return { disabled: false, tooltip: '' };
};

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
  if (!props.hasAuth) {
    emit('authClick');
    return;
  }
  if (action === 'reinstall') {
    emit('reinstall');
  } else if (action === 'assignUnit') {
    emit('assignUnit');
  } else {
    let operateData = props.data;
    let batch = props.batch;
    if (props.isCrossPageSelection) {
      await getCorssPageHostIds();
      operateData = crossPageHostIdData.value.map(item => ({ bk_host_id: item })) as any;
      batch = true; // 强制设置为批量模式
    }
    const titleObj = {
      firstIp: props.data[0].bk_host_innerip,
      num: operateData.length,
    };
    if (action === 'upgrade') {
      upgradePreviewData.hosts = operateData as Host[];
      upgradePreviewData.isShow = true;
    } else if (action === 'restart') {
      operateDialogIsShow.value = true;
      operateDialogData.type = action;
      operateDialogData.title = batch
        ? t('topoManager.workAreaDetail.batchRestartTitle') : t('topoManager.workAreaDetail.restartTitle');
      operateDialogData.subTitle = batch
        ? t('topoManager.workAreaDetail.restartSubTitle', { firstIp: titleObj.firstIp, num: titleObj.num })
        : t('topoManager.workAreaDetail.restartSubTitle', { firstIp: titleObj.firstIp });
    } else if (action === 'unload') {
      InfoBox({
        title: batch ? t('topoManager.workAreaDetail.batchUnloadTitle') : t('topoManager.workAreaDetail.unloadTitle'),
        subTitle: batch
          ? t('topoManager.workAreaDetail.batchUnloadSubTitle', { firstIp: titleObj.firstIp, num: titleObj.num })
          : t('topoManager.workAreaDetail.unloadSubTitle', { firstIp: titleObj.firstIp }),
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
  type: '',
});
const operateDialogIsShow = ref(false);
const operateDialogData = {
  type: '',
  title: '',
  subTitle: '',
};
const upgradePreviewData = reactive({ isShow: false, hosts: [] as Host[] });
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
  // 按 os_type_cpu_arch 建立版本映射
  const versionMap = new Map<string, string>();
  versionList.forEach((v: any) => {
    versionMap.set(`${v.os_type}_${v.cpu_arch}`, v.version);
  });
  const params = {
    host: operateData?.map((item: any) => {
      const host = typeof item === 'number' ? { os_type: '', cpu_arch: '' } : (item.info ?? item);
      const key = `${host.os_type}_${host.cpu_arch}`;
      return {
        bk_host_id: typeof item === 'number' ? item : item.bk_host_id,
        target_version: versionMap.get(key) || '',
        force: false,
        graceful_restart_timeout_sec: 0,
      };
    }),
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

<style lang="postcss">
.proxy-action-dropdown-menu {
  min-width: 100px !important;
}
/* 无权限菜单项：灰色文字 + cursor: pointer 保持事件可响应 */
.auth-lock-dropdown-item {
  color: #c4c6cc !important;
  cursor: pointer;

  &:hover {
    background-color: #f0f1f5 !important;
  }
}
/* 因业务规则禁用的菜单项 */
.operate-item-disabled {
  color: #c4c6cc !important;
  cursor: not-allowed !important;

  &:hover {
    background-color: transparent !important;
  }
}
</style>
