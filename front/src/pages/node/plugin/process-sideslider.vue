<template>
  <Sideslider
    v-model:is-show="isShow"
    render-directive="if"
    :title="$t('pluginManagement.plugin.process.processList')"
    width="1200"
  >
    <div class="p-[24px]">
      <!-- 操作栏：安装按钮 + 批量操作 + 复制IP -->
      <div class="flex items-center gap-[8px] mb-[16px]">
        <Button v-if="enabledOperations.has('install')" theme="primary" :class="{ 'btn-reinstall': props.type === 'plugin' && hasSelection }" @click="handlePrimaryButtonClick">
          {{ primaryButtonLabel }}
        </Button>
        <Dropdown
          theme="light"
          trigger="click"
          :popover-options="{ clickContentAutoHide: true }"
        >
          <Button :disabled="!selection.length">
            <span>{{ $t('pluginManagement.plugin.batchOperate') }}</span>
            <i class="nodeman-icon nc-arrow-down ml-[5px] text-[18px] text-[#979BA5]"></i>
          </Button>
          <template #content>
            <Dropdown.DropdownMenu>
              <Dropdown.DropdownItem
                v-for="item in batchOperateList"
                :key="item.id"
                @click="handlePluginOperate(item.id, selection, true)"
              >
                {{ item.name }}
              </Dropdown.DropdownItem>
            </Dropdown.DropdownMenu>
          </template>
        </Dropdown>
      <copy-ip-dropdown
        :type="'agent'"
        :disabled="!selection.length"
        :data="processList"
        :list="[]"
      ></copy-ip-dropdown>
    </div>
    <!-- 非默认版本提示 -->
    <Alert
      v-if="hasOutdatedVersion"
      class="mb-[16px]"
      theme="warning"
      :title="$t('pluginManagement.plugin.process.notDefaultVersionTitle', { names: outdatedPluginNames.join('、') })"
    />
      <Loading
        :title="$t('table.loading')"
        :loading="loading"
        class="overflow-auto">
        <Table
          class="filterTable"
          :data="processList"
          :empty-text="$t('table.empty')"
          :show-settings="isShowSetting"
          :pagination="pagination"
          :settings="settings"
          show-overflow-tooltip
          :tooltip-config="{
            popupClassName: 'process-table',
          }"
          @setting-change="handleSettingChange"
          @page-limit-change="pageLimitChange"
          @page-value-change="pageValueChange"
          @checkbox-change="handleSelectChange"
          @checkbox-all="handleSelectAllChange"
          @column-filter="handleFilter"
        >
          
          <TableColumn type="checkbox" width="60" fixed="left"></TableColumn>
          <TableColumn
            v-if="type === 'plugin'"
            title="Host ID"
            field="bk_host_id"
            fixed="left"
            min-width="120"
          ></TableColumn>
          <TableColumn
            v-if="type === 'plugin'"
            :title="$t('platform.nodeMan.inner_ip')"
            field="bk_host_innerip"
            min-width="150"
            fixed="left"
          ></TableColumn>
          <TableColumn
            v-if="type === 'plugin'"
            :title="$t('platform.nodeMan.inner_ipv6')"
            field="bk_host_innerip_v6"
            min-width="120"
          >
          </TableColumn>
          <TableColumn
            :title="$t('pluginManagement.plugin.table.pluginName')"
            field="plugin_name"
            min-width="150"
            :filter="filterOptionSource.plugin_name"
          ></TableColumn>
          <TableColumn
            :title="$t('pluginManagement.plugin.table.packageName')"
            field="plugin_pkg_name"
            min-width="150"
            :filter="filterOptionSource.plugin_pkg_name"
          ></TableColumn>
          <TableColumn
            :title="$t('pluginManagement.plugin.table.pluginGroup')"
            field="plugin_group"
            min-width="120"
            :filter="filterOptionSource.plugin_group"
          ></TableColumn>
          <TableColumn
            :title="$t('pluginManagement.plugin.process.os')"
            field="os_type"
            min-width="120"
            :filter="filterOptionSource.os_type"
          ></TableColumn>
          <TableColumn
            :title="$t('pluginManagement.plugin.process.cpuArch')"
            field="cpu_arch"
            min-width="120"
            :filter="filterOptionSource.cpu_arch"
          ></TableColumn>
          <TableColumn
            :title="$t('pluginManagement.plugin.process.version')"
            field="version"
            min-width="160"
            :filter="filterOptionSource.version"
          >
            <template #default="{ row }">
              <span class="flex items-center w-full gap-[4px]">
                <span
                  class="flex-1 min-w-0 truncate"
                  v-bk-tooltips="{ content: row.version, disabled: !row.version || row.version.length <= 20 }"
                >{{ row.version }}</span>
                <!-- relay 非默认：Popover + 建议升级（仅 running 状态提示） -->
                <Popover
                  v-if="row.status === 'running' && !isProcessDefaultVersion(row) && row.plugin_name === 'bk-nodemgr-relay'"
                  theme="light"
                  trigger="hover"
                  placement="top"
                  :arrow="true"
                >
                  <i class="nodeman-icon nc-tips text-[#FF9C01] text-[16px] flex-shrink-0 cursor-pointer"></i>
                  <template #content>
                    <div class="flex items-center gap-[8px] p-[4px]">
                      <span>{{ $t('pluginManagement.plugin.process.notDefaultVersionTooltip') }}</span>
                      <Button
                        text
                        theme="primary"
                        @click="handlePluginOperate('upgrade', [row])"
                      >
                        {{ $t('pluginManagement.plugin.process.suggestUpgrade') }}
                      </Button>
                    </div>
                  </template>
                </Popover>
                <!-- 非 relay 非默认：简单 tooltip（仅 running 状态提示） -->
                <i
                  v-if="row.status === 'running' && !isProcessDefaultVersion(row) && row.plugin_name !== 'bk-nodemgr-relay'"
                  v-bk-tooltips="$t('pluginManagement.plugin.process.notDefaultVersionSubTip', { version: getDefaultVersion(row) || '--' })"
                  class="nodeman-icon nc-tips text-[#FF9C01] text-[16px] flex-shrink-0 cursor-pointer"
                ></i>
              </span>
            </template>
          </TableColumn>
          <TableColumn
            title="Agent ID"
            field="agent_id"
            min-width="220"
          ></TableColumn>
          <TableColumn
            :title="$t('pluginManagement.plugin.process.processName')"
            field="name"
            min-width="120"
          ></TableColumn>
          <TableColumn
            :title="$t('pluginManagement.plugin.process.processStatus')"
            field="status"
            min-width="120"
            :filter="filterOptionSource.status"
          >
            <template #default="{ row }">
              <div class="flex items-center">
                <i
                  :class="`nodeman-icon nc-${statusMap[row.status]?.icon} status-icon`"
                ></i>
                <span>{{ statusMap[row.status]?.text || "--" }}</span>
              </div>
            </template>
          </TableColumn>
          <TableColumn
            :title="$t('pluginManagement.plugin.process.installPath')"
            field="setup_path"
            min-width="120"
          ></TableColumn>
          <TableColumn
            :title="$t('pluginManagement.plugin.process.pidPath')"
            field="pid_path"
            min-width="120"
          ></TableColumn>
          <TableColumn
            :title="$t('pluginManagement.plugin.process.configPath')"
            field="config_path"
            min-width="120"
          ></TableColumn>
          <TableColumn
            :title="$t('pluginManagement.plugin.process.logPath')"
            field="log_path"
            min-width="130"
          ></TableColumn>
          <TableColumn
            :title="$t('pluginManagement.plugin.process.systemAccount')"
            field="user"
            min-width="150"
          ></TableColumn>
          <TableColumn
            :title="$t('pluginManagement.plugin.process.startCmd')"
            field="start_cmd"
            min-width="180"
          ></TableColumn>
          <TableColumn
            :title="$t('pluginManagement.plugin.process.stopCmd')"
            field="stop_cmd"
            min-width="180"
          ></TableColumn>
          <TableColumn
            :title="$t('pluginManagement.plugin.process.restartCmd')"
            field="restart_cmd"
            min-width="180"
          ></TableColumn>
          <TableColumn
            :title="$t('pluginManagement.plugin.process.reloadCmd')"
            field="reload_cmd"
            min-width="180"
          ></TableColumn>
          <TableColumn
            :title="$t('pluginManagement.plugin.process.killCmd')"
            field="kill_cmd"
            min-width="180"
          ></TableColumn>
          <TableColumn
            :title="$t('pluginManagement.plugin.process.versionQueryCmd')"
            field="version_cmd"
            min-width="180"
          ></TableColumn>
          <TableColumn
            :title="$t('pluginManagement.plugin.process.healthCheckCmd')"
            field="health_cmd"
            min-width="180"
          ></TableColumn>
          <TableColumn
            :title="$t('pluginManagement.plugin.process.cpuLimitPercent')"
            field="cpu_limit_percent"
            min-width="120"
          ></TableColumn>
          <TableColumn
            :title="$t('pluginManagement.plugin.process.memLimitPercent')"
            field="mem_limit_percent"
            min-width="120"
          ></TableColumn>
          <TableColumn
            :title="$t('pluginManagement.plugin.process.restartStrategy')"
            field="restart_type"
            min-width="120"
          >
            <template #default="{ row }">
              {{ row.restart_type === 'auto' && row.auto_start
                ? $t('pluginManagement.plugin.process.autoStart')
                : $t('pluginManagement.plugin.process.manualStart') }}
            </template>
          </TableColumn>
          <TableColumn
            :title="$t('pluginManagement.plugin.process.startCheckDelay')"
            field="start_check_seconds"
            min-width="120"
          >
            <template #default="{ row }">
              <span>{{ row.start_check_seconds }}s</span>
            </template>
          </TableColumn>
          <TableColumn
            :title="$t('pluginManagement.plugin.process.stopCheckDelay')"
            field="stop_check_seconds"
            min-width="120"
          >
            <template #default="{ row }">
              <span>{{ row.stop_check_seconds }}s</span>
            </template>
          </TableColumn>
          <TableColumn
            :title="$t('pluginManagement.plugin.process.commandTimeout')"
            field="operate_timeout_seconds"
            min-width="120"
          >
            <template #default="{ row }">
              <span>{{ row.operate_timeout_seconds }}s</span>
            </template>
          </TableColumn>
          <!-- 操作列 -->
          <TableColumn
            :title="$t('pluginManagement.plugin.table.operation')"
            field="action"
            min-width="100"
            fixed="right"
          >
            <template #default="{ row }">
              <Button
                v-if="enabledOperations.has('install')"
                theme="primary"
                text
                class="reinstall"
                @click="handlePluginOperate('reinstall', [row])"
              >
                {{ $t('pluginManagement.plugin.operate.reinstall') }}
              </Button>
              <Dropdown
                theme="light"
                trigger="click"
                :popover-options="{ clickContentAutoHide: true }"
              >
                <Button class="ml-[15px]" text>
                  <span class="nodeman-icon nc-more"></span>
                </Button>
                <template #content>
                  <Dropdown.DropdownMenu>
                    <Dropdown.DropdownItem
                      v-for="item in getSingleOperateList(row)"
                      :key="item.id"
                      @click="handlePluginOperate(item.id, [row])"
                    >
                      {{ item.name }}
                    </Dropdown.DropdownItem>
                  </Dropdown.DropdownMenu>
                </template>
              </Dropdown>
            </template>
          </TableColumn>
        </Table>
      </Loading>
    </div>
    <!-- <template #footer>
      <Button class="mr-[8px]" theme="primary" @click="handleBeforeClose">
        {{ $t('action.confirm') }}
      </Button>
      <Button @click="handleBeforeClose">
        {{ $t('action.cancel') }}
      </Button>
    </template> -->
  </Sideslider>
  <!-- 操作确认弹窗（重启/卸载/停止） -->
  <operate-dialog
    v-model:is-show="operateDialogIsShow"
    :title="operateDialogData.title"
    :type="operateDialogData.type"
    :sub-title="operateDialogData.subTitle"
    :data="pendingOperateData"
    :columns="operateDialogColumns"
    :status-map="processStatusTextMap"
    :show-table="operateDialogShowTable"
    :hide-restart-options="true"
    :hide-checkbox="!isBatchOperate"
    :selection-confirm-formatter="operateDialogShowTable ? undefined : selectionConfirmFormatter"
    @confirm="handleOperateConfirm"
  ></operate-dialog>
</template>
<script lang="ts" setup>
import {
  Alert,
  Button,
  Dropdown,
  Loading,
  Message,
  Popover,
  Sideslider,
} from 'bkui-vue';
import { computed, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';

import { Table, TableColumn } from '@blueking/table';

import type { DistinctProcessRespData } from '@/@types/process';
import { PluginAPIService } from '@/api/modules/plugin';
import { PackageService } from '@/api/modules/pkg';
import { ProcessAPIService } from '@/api/modules/process';
import { TopoService } from '@/api/modules/topo';
import OperateDialog from '@/components/operate-dialog.vue';
import useTableSetting from '@/composables/use-table-setting';
import { PACKAGE_GENERATION } from '@/common/const';
import { useMainStore } from '@/stores/main';
import { useAuthStore } from '@/stores/auth';


interface FilterOption {
  list: { text: string; value: string }[];
  checked: string[];
  filterScope: string;
}

// Sideslider显示状态
const isShow = defineModel('isShow', { type: Boolean });

const props = defineProps({
  type: {
    type: String,
    default: 'plugin',
  },
  /** 节点类型，用于 IP 选择器过滤：'agent' | 'proxy' | 不传则不过滤 */
  nodeType: {
    type: String,
    default: '',
  },
  plugin: {
    type: Object,
    default: () => ({}),
  },
  node: {
    type: Object,
    default: () => ({}),
  },
});

const { t } = useI18n();
const router = useRouter();
const mainStore = useMainStore();
const authStore = useAuthStore();

/**
 * 根据当前用户权限解析 node_role 过滤条件
 * 与 ip-selector.ts 中 resolveNodeRoleFilter 逻辑一致：
 * - agent_view 有权 + proxy_view 无权 → ['agent', 'blank']
 * - proxy_view 有权 + agent_view 无权 → ['proxy']
 * - 都有权 → [] 不过滤
 * - 都无权或权限未加载 → ['agent', 'blank']（保守过滤）
 */
const resolveNodeRoleFilter = (): string[] => {
  const bizIds = mainStore.selectedBusinessId;
  if (!bizIds || bizIds.length === 0) return [];

  // 权限数据尚未加载完成 → 保守过滤：只显示 agent/blank
  if (!authStore.authorizedLoaded) {
    return ['agent', 'blank'];
  }

  // 检查所有选中的业务是否有 agent_view / proxy_view 权限
  let hasAgentView = false;
  let hasProxyView = false;
  for (const bizId of bizIds) {
    if (!hasAgentView && authStore.hasAuthorizedResource('agent_view', bizId)) hasAgentView = true;
    if (!hasProxyView && authStore.hasAuthorizedResource('proxy_view', bizId)) hasProxyView = true;
    if (hasAgentView && hasProxyView) break;
  }

  if (hasAgentView && !hasProxyView) return ['agent', 'blank'];
  if (hasProxyView && !hasAgentView) return ['proxy'];
  // 都有权 → [] 不过滤；都无权 → 保守过滤
  if (!hasAgentView && !hasProxyView) return ['agent', 'blank'];
  return [];
};

// ---------- 操作相关 ----------
// 需要跳转到单独页面配置的操作: 安装、升级、重载
const needPageOperations = ['install', 'upgrade', 'reload', 'reinstall'];

// 已启用的操作 — 从接口动态获取
const enabledOperations = ref(new Set<string>());

const loadPermittedOperations = async () => {
  const res = await PluginAPIService.ListPluginPermittedOperation({
    page: { limit: 100, offset: 0 },
  }).catch(() => ({ operations: [] }));
  const ops = new Set<string>();
  (res?.operations || []).forEach((op: any) => {
    (op.permissions || op.permission || []).forEach((p: string) => {
      ops.add(p);
      if (p === 'reconfigv') ops.add('reload');
      // 接口返回 install 权限时，同时启用 reinstall（重装）操作
      if (p === 'install') ops.add('reinstall');
    });
  });
  enabledOperations.value = ops;
};

// 所有操作列表（单行操作使用）
const allOperations = [
  { id: 'reinstall', nameKey: 'pluginManagement.plugin.operate.reinstall' },
  { id: 'upgrade', nameKey: 'pluginManagement.plugin.operate.upgrade' },
  { id: 'reload', nameKey: 'pluginManagement.plugin.operate.reload' },
  { id: 'uninstall', nameKey: 'pluginManagement.plugin.operate.uninstall' },
  { id: 'restart', nameKey: 'pluginManagement.plugin.operate.restart' },
  { id: 'stop', nameKey: 'pluginManagement.plugin.operate.stop' },
];

// 批量操作列表 — 根据来源区分：
// type=plugin（从插件列表进入）：包含升级，排除重装和重载
// type=node（从 Agent/Proxy 列表进入）：只有卸载、重启、停止
const batchOperateList = computed(() => {
  const excludedOps = props.type === 'plugin'
    ? ['reinstall', 'reload']
    : ['upgrade', 'reinstall', 'reload'];
  return allOperations
    .filter(op => enabledOperations.value.has(op.id) && !excludedOps.includes(op.id))
    .map(op => ({ id: op.id, name: t(op.nameKey) }));
});

// 单行操作列表 — 只显示有权限的（重装已作为独立按钮，故排除）
const getSingleOperateList = (row: any) => {
  const ops = row.plugin_group === 'strategy'
    ? [{ id: 'restart', nameKey: 'pluginManagement.plugin.operate.restart' }, { id: 'stop', nameKey: 'pluginManagement.plugin.operate.stop' }]
    : allOperations;
  return ops
    .filter(op => op.id !== 'reinstall' && enabledOperations.value.has(op.id))
    .map(op => ({ id: op.id, name: t(op.nameKey) }));
};

// 操作确认弹窗
const operateDialogIsShow = ref(false);
const isBatchOperate = ref(false);
const operateDialogData = reactive({ type: '', title: '', subTitle: '' });
const pendingOperateData = ref<any[]>([]);
const pendingOperateType = ref('');
const operateDialogShowTable = ref(true);

// 操作弹窗表格列配置（展示受影响的主机列表）
const operateDialogColumns = [
  { field: 'bk_host_innerip', label: t('platform.nodeMan.inner_ip'), minWidth: 150 },
  { field: 'plugin_name', label: t('pluginManagement.plugin.table.pluginName'), minWidth: 150 },
  { field: 'status', label: t('pluginManagement.plugin.process.processStatus'), minWidth: 120 },
];

// 批量操作确认提示格式化（统一按插件名分组，显示主机 IP）
const selectionConfirmFormatter = (selectedRows: any[]) => {
  if (selectedRows.length === 0) return '';
  const count = selectedRows.length;

  // 按插件名分组，每个插件下列出主机 IP
  const pluginMap = new Map<string, Set<string>>();
  selectedRows.forEach((row: any) => {
    const name = row.plugin_name || '';
    const ip = row.bk_host_innerip || String(row.bk_host_id) || '';
    if (name) {
      if (!pluginMap.has(name)) pluginMap.set(name, new Set());
      if (ip) pluginMap.get(name)!.add(ip);
    }
  });
  const parts: string[] = [];
  pluginMap.forEach((ips, name) => {
    parts.push(ips.size > 0 ? `${name}(${[...ips].join(', ')})` : name);
  });
  const detail = parts.join('、');

  const keyMap: Record<string, string> = {
    uninstall: 'pluginManagement.plugin.operate.selectionConfirmUninstall',
    restart: 'pluginManagement.plugin.operate.selectionConfirmRestart',
    stop: 'pluginManagement.plugin.operate.selectionConfirmStop',
  };
  const key = keyMap[pendingOperateType.value] || keyMap.restart;
  return t(key, { detail, count });
};

// 处理插件操作
const handlePluginOperate = (operateType: string, data: any[], batch = false) => {
  // 从节点状态页进入时 props.plugin 为空，插件名需从行数据取；从插件列表进入时 props.plugin.name 有效
  const pluginName = data[0]?.plugin_name || props.plugin?.name || '';
  const count = data.length;

  if (needPageOperations.includes(operateType)) {
    // 安装/重装/升级/重载 → 关闭侧边栏，跳转到配置页面
    isShow.value = false;

    // 重装/升级/重载：从当前勾选数据构造 prefillHosts 和 prefillVersions（不再调 process/list）
    let prefillHosts = '';
    let prefillVersions = '';
    if (data.length > 0 && operateType !== 'install') {
      // 提取去重的主机ID
      const hostIdSet = new Set<number>();
      data.forEach((item: any) => {
        if (item.bk_host_id) hostIdSet.add(item.bk_host_id);
      });
      prefillHosts = JSON.stringify([...hostIdSet]);

      // 提取版本映射（os_type_cpu_arch → version），version 字段可能包含 BuildTime/GitHash 等信息
      // 只截取第一行纯 Version 部分（如 "v3.0.1-alpha-7-27-gdce6b11f-26-03.18"）
      const versionMap = new Map<string, string>();
      data.forEach((item: any) => {
        const osType = item.os_type || '';
        const cpuArch = item.cpu_arch || '';
        const rawVersion = item.version || '';
        // version 字段格式为 "Version: xxx\nBuildTime: ...\nGitHash: ..."，只取 Version 行的值
        const versionMatch = rawVersion.match(/^Version:\s*(.+?)(?:\r?\n|$)/m);
        const version = versionMatch ? versionMatch[1].trim() : rawVersion;
        if (osType && cpuArch && version) {
          const key = `${osType}_${cpuArch}`;
          if (!versionMap.has(key)) versionMap.set(key, version);
        }
      });
      prefillVersions = JSON.stringify(Object.fromEntries(versionMap));
    }

    router.push({
      name: 'pluginOperate',
      query: {
        operationType: operateType,
        pluginName,
        ...(props.nodeType ? { type: props.nodeType } : {}),
        ...(prefillHosts ? { prefillHosts } : {}),
        ...(prefillVersions ? { prefillVersions } : {}),
      },
    });
  } else if (operateType === 'restart' || operateType === 'uninstall' || operateType === 'stop') {
    // 重启/卸载/停止 → operate-dialog 确认
    // 浅拷贝数据，避免弹窗内部 checkbox 事件修改主表格的 checked 状态
    pendingOperateData.value = data.map((item: any) => ({ ...item }));
    pendingOperateType.value = operateType;
    isBatchOperate.value = batch;
    // 批量操作使用简化弹窗（无表格，直接展示确认提示）
    operateDialogShowTable.value = !batch;
    operateDialogData.type = operateType;

    // 提取插件名列表（用于批量 subTitle）
    const pluginNames = [...new Set(data.map((item: any) => item.plugin_name).filter(Boolean))].join('、')
      || props.plugin?.name || '';

    if (operateType === 'restart') {
      operateDialogData.title = batch
        ? t('pluginManagement.plugin.operate.batchRestartTitle')
        : t('pluginManagement.plugin.operate.restartTitle');
      operateDialogData.subTitle = batch
        ? t('pluginManagement.plugin.operate.batchRestartSubTitle', { pluginNames, count })
        : t('pluginManagement.plugin.operate.restartSubTitle', { pluginName });
    } else if (operateType === 'uninstall') {
      operateDialogData.title = batch
        ? t('pluginManagement.plugin.operate.batchUninstallTitle')
        : t('pluginManagement.plugin.operate.uninstallTitle');
      operateDialogData.subTitle = batch
        ? t('pluginManagement.plugin.operate.batchUninstallSubTitle', { pluginNames, count })
        : t('pluginManagement.plugin.operate.uninstallSubTitle', { pluginName });
    } else {
      operateDialogData.title = batch
        ? t('pluginManagement.plugin.operate.batchStopTitle')
        : t('pluginManagement.plugin.operate.stopTitle');
      operateDialogData.subTitle = batch
        ? t('pluginManagement.plugin.operate.batchStopSubTitle', { pluginNames, count })
        : t('pluginManagement.plugin.operate.stopSubTitle', { pluginName });
    }

    operateDialogIsShow.value = true;
  }
};

// 操作弹窗确认（重启/卸载/停止）
const handleOperateConfirm = async (extraData: any = {}) => {
  // 优先使用弹窗中勾选的数据，若无则使用全部
  const data = extraData.selection?.length ? extraData.selection : pendingOperateData.value;
  const operateType = pendingOperateType.value;
  let res: any;

  if (operateType === 'uninstall') {
    res = await PluginAPIService.UninstallPlugin({
      plugin: data.map((item: any) => ({ bk_host_id: item.bk_host_id, plugin_name: item.plugin_name })),
    });
  } else if (operateType === 'stop') {
    res = await PluginAPIService.StopPlugin({
      plugin: data.map((item: any) => ({ bk_host_id: item.bk_host_id, plugin_name: item.plugin_name })),
    });
  } else {
    // 默认重启
    res = await PluginAPIService.RestartPlugin({
      plugin: data.map((item: any) => ({ bk_host_id: item.bk_host_id, plugin_name: item.plugin_name })),
    });
  }

  if (res?.workflow_id) {
    Message({ theme: 'success', message: t('pluginManagement.plugin.operate.operateSuccess') });
    router.push({
      name: 'taskDetail',
      params: { taskId: res.workflow_id },
      query: { active: 'plugin' },
    });
  }
};

// 表格
const { isShowSetting, settings, handleSettingChange } = useTableSetting(
  {
    checked: [
      'plugin_name',
      'status',
      'pid',
      'version',
      ...(props.type === 'plugin' ? ['bk_host_id', 'bk_host_innerip', 'bk_host_innerip_v6'] : ['os_type', 'cpu_arch']),
      'action',
    ],
    disabled: ['action'],
  },
  `nodeMng-${props.type}-status-sideslider`,
);

// 分页
const pagination = reactive({ count: 0, limit: 50, current: 1, remote: true });

const pageLimitChange = async (limit: number) => {
  pagination.limit = limit;
  pagination.current = 1; // 页码重置为1
  await getProcessList(); // 分页变化不防抖，立即执行
};

const pageValueChange = async (current: number) => {
  pagination.current = current;
  await getProcessList(); // 分页变化不防抖，立即执行
};

// 进程状态映射
const statusMap = {
  running: {
    text: t('pluginManagement.plugin.process.status.running'),
    icon: 'running',
  },
  stopped: {
    text: t('pluginManagement.plugin.process.status.stopped'),
    icon: 'terminated',
  },
  unregister: {
    text: t('pluginManagement.plugin.process.status.unregister'),
    icon: 'unknown',
  },
  init: {
    text: t('pluginManagement.plugin.process.status.init'),
    icon: 'unknown',
  },
  unknown: {
    text: t('pluginManagement.plugin.process.status.unknown'),
    icon: 'unknown',
  },
};

// 操作弹窗中 status 列的纯文本映射
const processStatusTextMap = computed(() => {
  const map: Record<string, string> = {};
  for (const [key, val] of Object.entries(statusMap)) {
    map[key] = val.text;
  }
  return map;
});

// 进程列表
const processList = ref<any[]>([]);

// 默认版本映射：key = `${plugin_name}_${os_type}_${cpu_arch}`，value = 默认版本号
const defaultVersionMap = ref<Map<string, string>>(new Map());

// 判断进程版本是否与默认版本一致（支持 version 字段为多行字符串，取 Version: 行）
const extractVersion = (rawVersion: string): string => {
  if (!rawVersion) return '';
  const match = rawVersion.match(/^Version:\s*(.+?)(?:\r?\n|$)/m);
  return match ? match[1].trim() : rawVersion.trim();
};

const getDefaultVersionKey = (row: any) => `${row.plugin_name}_${row.os_type}_${row.cpu_arch}`;
const getDefaultVersion = (row: any): string | undefined => defaultVersionMap.value.get(getDefaultVersionKey(row));

const isProcessDefaultVersion = (row: any): boolean => {
  const defaultVersion = getDefaultVersion(row);
  if (!defaultVersion) return true;
  return extractVersion(row.version) === extractVersion(defaultVersion);
};

const NEED_DEFAULT_CHECK_PLUGIN = 'bk-nodemgr-relay';

// 非默认版本的插件名集合（用于顶部 Alert，仅针对 relay）
const outdatedPluginNames = computed(() => {
  const names = new Set<string>();
  processList.value.forEach((row: any) => {
    if (row.plugin_name === NEED_DEFAULT_CHECK_PLUGIN && !isProcessDefaultVersion(row)) {
      names.add(row.plugin_name);
    }
  });
  return [...names];
});

const hasOutdatedVersion = computed(() => outdatedPluginNames.value.length > 0);

// 表格勾选
const selection = computed(() => processList.value.filter((item: any) => item.checked));
const hasSelection = computed(() => selection.value.length > 0);

// 主按钮文案：从插件列表进入且选中主机时显示"重装"，否则显示"安装"
// 从节点（Agent/Proxy）状态页进入时不提供"变重装"逻辑，始终显示"安装"
const primaryButtonLabel = computed(() =>
  props.type === 'plugin' && hasSelection.value
    ? t('pluginManagement.plugin.operate.reinstall')
    : t('pluginManagement.plugin.operate.install'),
);

// 主按钮点击：从插件列表进入且选中主机时走重装流程，否则走安装流程
// 从节点状态页进入时始终走安装流程
const handlePrimaryButtonClick = () => {
  if (props.type === 'plugin' && hasSelection.value) {
    handlePluginOperate('reinstall', selection.value, true);
  } else {
    handlePluginOperate('install', [], false);
  }
};

const handleSelectChange = ({ checked, row }: { checked: boolean; row: any }) => {
  if (row) row.checked = checked;
};

const handleSelectAllChange = ({ checked }: { checked: boolean }) => {
  processList.value.forEach((item: any) => { if (item) item.checked = checked; });
};

// 筛选
const filterOptionSource: Record<string, FilterOption> = reactive({
  os_type: { list: [], checked: [], filterScope: 'all' },
  cpu_arch: { list: [], checked: [], filterScope: 'all' },
  version: { list: [], checked: [], filterScope: 'all' },
  plugin_group: { list: [], checked: [], filterScope: 'all' },
  plugin_name: { list: [], checked: [], filterScope: 'all' },
  plugin_pkg_name: { list: [], checked: [], filterScope: 'all' },
  status: { list: [], checked: [], filterScope: 'all' },
});

const searchSelectValue = ref<any[]>([]);
const handleFilter = ({ checked, field }: { checked: string[]; field: string }) => {
  let newId = field;
  switch (field) {
    case 'os_type':
      newId = 'platform_os';
      break;
    case 'cpu_arch':
      newId = 'platform_arch';
      break;
  }
  const index = searchSelectValue.value.findIndex((item: any) => item.id === newId);
  if (index > -1) searchSelectValue.value.splice(index, 1);
  if (checked.length) {
    searchSelectValue.value.push({
      id: newId,
      name: newId,
      values: checked.map((item: any) => {
        let name = item;
        switch (field) {
          case 'status':
            name = statusMap[item]?.text || item;
            break;
        }
        return { id: item, name };
      }),
    });
  }
};

// distinct
const distinct = ref<DistinctProcessRespData>();
const getDistinct = async () => {
  const res = await ProcessAPIService.DistinctProcess({
    exact_include_conditions: props.type === 'plugin' ? {
      plugin_name: [props.plugin.name],
      bk_biz_id: mainStore.selectedBusinessId,
    } : {
      bk_host_id: [props.node.bk_host_id],
      bk_biz_id: mainStore.selectedBusinessId,
    },
  });
  if (res && Object.keys(res).length > 0) {
    distinct.value = res;
    Object.keys(res).forEach((key: any) => {
      if (filterOptionSource[key]) {
        filterOptionSource[key].list = res[key]
          .filter((item: any) => item !== '')
          .map((value: string | number) => {
            let text = value;
            if (key === 'status') text = statusMap[value]?.text || value;
            return { text, value };
          });
      }
    });
  }
};

// 加载默认插件版本信息
const loadDefaultPluginVersions = async () => {
  const pluginNames = [...new Set(processList.value.map((item: any) => item.plugin_name).filter(Boolean))];
  if (pluginNames.length === 0) { defaultVersionMap.value = new Map(); return; }

  const res = await PackageService.ListReleasePlugin({
    page: { limit: 500, offset: 0 },
    generation: PACKAGE_GENERATION,
    only_count: false,
    exact_include_conditions: { name: pluginNames },
  }).catch(() => ({ total: 0, items: [] }));

  const map = new Map<string, string>();
  (res.items || []).forEach((item: any) => {
    const release = item.release || item;
    if (release.as_default && release.name && release.os_type && release.cpu_arch) {
      map.set(`${release.name}_${release.os_type}_${release.cpu_arch}`, release.version);
    }
  });
  defaultVersionMap.value = map;
};

const loading = ref(false);
const fuzzyKeys = new Set(['name', 'plugin_pkg_name']);
const getParams = () => {
  const params = {
    page: { limit: pagination.limit, offset: (pagination.current - 1) * pagination.limit },
    exact_include_conditions: props.type === 'plugin' ? {
      plugin_name: [props.plugin.name],
      bk_biz_id: mainStore.selectedBusinessId,
    } : {
      bk_host_id: [props.node.bk_host_id],
      bk_biz_id: mainStore.selectedBusinessId,
    },
    fuzzy_include_conditions: {},
  };
  searchSelectValue.value.forEach((item: any) => {
    const target = fuzzyKeys.has(item.id)
      ? params.fuzzy_include_conditions
      : params.exact_include_conditions;
    target[item.id] = item.values?.map((value: any) => value.id);
  });
  return params;
};
const getProcessList = async () => {
  loading.value = true;
  const nodeRoleFilter = resolveNodeRoleFilter();
  const res = await ProcessAPIService.ListProcesses(getParams()).catch((err) => {
    console.log(err);
    return {
      total: 0,
      items: [],
    };
  });
  pagination.count = res.total;
  loading.value = false;

  let hostListMap = new Map();
  if (props.type === 'plugin') {
    const hostList = await TopoService.HostList({
      page: {
        limit: pagination.limit,
        offset: (pagination.current - 1) * pagination.limit,
      },
      exact_include_conditions: {
        bk_host_id: res.items.map(item => item.bk_host_id),
        bk_biz_id: mainStore.selectedBusinessId,
        ...(nodeRoleFilter.length > 0 ? { node_role: nodeRoleFilter } : {}),
      },
    }).catch((err) => {
      console.log(err);
      return {
        items: [],
      };
    });
    hostListMap = new Map(hostList.items.map(item => [item.bk_host_id, {
      bk_host_innerip: item.info.bk_host_innerip_list?.join(','),
      bk_host_innerip_v6: item.info.bk_host_innerip_v6_list?.join(','),
    }]));
  };

  processList.value = res.items.map(item => ({
    ...item,
    ...item.platform,
    ...item.process_info,
    ...item.process_identity,
    ...item.process_controller,
    ...item.process_resource,
    ...item.process_monitor_policy,
    bk_host_innerip: hostListMap.get(item.bk_host_id)?.bk_host_innerip || '',
    bk_host_innerip_v6: hostListMap.get(item.bk_host_id)?.bk_host_innerip_v6 || '',
  }));

  // 加载当前进程涉及的插件默认版本，用于版本一致性提示
  await loadDefaultPluginVersions();
};

// 暂时没有编辑数据，不需要离开前确认
// const handleBeforeClose = (): Promise<boolean> => new Promise((resolve, reject) => {
//   InfoBox({
//     title: t('dialog.confirmClose'),
//     infoType: 'warning',
//     onConfirm: () => {
//       resolve(true);
//       isShow.value = false;
//     },
//     onCancel: () => reject(),
//   });
// });

watch(
  () => isShow.value,
  async () => {
    if (isShow.value) {
      // 先加载权限数据，确保 resolveNodeRoleFilter 能正确判断 node_role
      if (!authStore.authorizedLoaded) {
        await authStore.fetchAuthorized(
          [
            { action: 'agent_view', resource_type: 'biz' },
            { action: 'proxy_view', resource_type: 'biz' },
          ],
          'plugin-process',
        );
      }
      loadPermittedOperations();
      await getProcessList();
      if (processList.value.length > 0) {
        getDistinct();
      }
    } else {
      Object.keys(filterOptionSource).forEach((key: any) => {
        if (filterOptionSource[key]) {
          filterOptionSource[key].list = [];
        }
      });
      processList.value = [];
    }
  },
  { immediate: true },
);
watch(
  searchSelectValue,
  async () => {
    await getProcessList();
  },
  { deep: true },
);
</script>
<style lang="postcss" scoped>
.status-icon::before {
  content: "";
  display: inline-block;
  margin-right: 8px;
  width: 14px;
  height: 14px;
  border: 3px solid #f0f1f5;
  border-radius: 6.5px;
  background: #b2b5bd;
  flex-shrink: 0;
  vertical-align: middle;
}
.nc-running {
  &::before {
    background: #3fc06d;
    border-color: #e5f6ea;
  }
}
.nc-terminated {
  &::before {
    border-color: #ffe6e6;
    background: #ea3636;
  }
}
.nc-unknown {
  &::before {
    border-color: #f0f1f5;
    background: #b2b5bd;
  }
}
.operate-item-disabled {
  color: #c4c6cc !important;
  cursor: not-allowed !important;
  pointer-events: auto !important;

  &:hover {
    background-color: transparent !important;
  }
}

</style>
<style lang="postcss">
.vxe-table--tooltip-wrapper {
  &.process-table {
    z-index: 2004 !important;
  }
}

/* 插件进程表格 — 内置设置面板样式覆盖 */
.tippy-box[data-theme~='bk-vxe-table-setting-column-theme'] {
  min-width: 540px !important;
  .field-list-wrapper {
    /* flex + 弹性宽度：短项两列并排，长项自动撑满换行独占一行 */
    .bk-checkbox-group {
      display: flex !important;
      flex-wrap: wrap !important;
      column-gap: 20px;
      row-gap: 4px;
    }
    .field-list-item {
      /* flex-grow:1 让短项平分剩余空间(两列)；max-width 限制最多占一半；
         min-width 比一半略大：当文本撑超过一半时，下一项放不下被挤换行，长项独占一行 */
      flex: 1 1 auto;
      min-width: calc(50% - 10px);
      max-width: 100%;
      width: auto !important;
      height: auto !important;
      min-height: 32px;
      padding: 6px 8px;
      margin: 0 !important;
      box-sizing: border-box;
      align-items: flex-start !important;
      border-radius: 4px;
      transition: background-color 0.15s ease;

      &:hover {
        background-color: #f5f7fa;
      }

      .bk-checkbox {
        align-items: flex-start;
        /* checkbox 方框与多行文字的顶部对齐 */
        .bk-checkbox-input {
          margin-top: 2px;
          flex-shrink: 0;
        }
      }
      .bk-checkbox-label {
        white-space: normal !important;
        word-break: break-word;
        line-height: 1.5;
        font-size: 13px;
        color: #63656e;
      }
    }
  }
}
</style>

