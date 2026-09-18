<template>
  <div class="p-[24px]">
    <!-- 操作栏：安装按钮 -->
    <div class="flex items-center gap-[8px] mb-[16px]">
      <Button
        v-if="enabledOperations.has('install')"
        class="w-[130px]"
        theme="primary"
        @click="handlePluginOperate('install', [], false)"
      >
        {{ $t('pluginManagement.plugin.operate.install') }}
      </Button>
    </div>

    <Loading :loading="loading">
      <Table
        :data="pluginList"
        :max-height="maxHeight"
        :empty-text="$t('table.empty')"
        :show-settings="isShowSetting"
        :pagination="pagination"
        :settings="settings"
        @setting-change="handleSettingChange"
        @page-limit-change="pageLimitChange"
        @page-value-change="pageValueChange"
      >
        <TableColumn
          :title="$t('pluginManagement.plugin.table.pluginName')"
          field="name"
          min-width="150"
        ></TableColumn>
        <TableColumn
          :title="$t('pluginManagement.plugin.table.packageName')"
          field="pkg_name"
          min-width="150"
        >
          <template #default="{ row }">
            <Button text theme="primary" @click="handleGoToPluginPkgMng(row)">
              {{ row.pkg_name }}
            </Button>
          </template>
        </TableColumn>
        <TableColumn
          :title="$t('pluginManagement.plugin.table.pluginGroup')"
          field="group"
          min-width="120"
        ></TableColumn>
        <TableColumn
          :title="$t('pluginManagement.plugin.table.memo')"
          field="memo"
          min-width="380"
        ></TableColumn>
        <TableColumn
          :title="$t('pluginManagement.plugin.table.nodeCount')"
          field="node_num"
          min-width="120"
        >
          <template #default="{ row }">
            <Button text theme="primary" @click="openSidebar(row)">
              {{ row.node_num || 0 }}
              <i class="nodeman-icon nc-cloud-machine ml-[5px]"></i>
            </Button>
          </template>
        </TableColumn>
        <TableColumn
          :title="$t('pluginManagement.plugin.table.operation')"
          field="operation"
          min-width="200"
          fixed="right"
        >
          <template #default="{ row }">
            <!-- 重装按钮（提到下拉外） -->
            <Button
              v-if="enabledOperations.has('reinstall')"
              text
              theme="primary"
              @click="handlePluginOperate('reinstall', [row])"
            >
              {{ $t('pluginManagement.plugin.operate.reinstall') }}
            </Button>
            <!-- 更多操作下拉（升级/重载/卸载/启动/重启/停止） -->
            <Dropdown
              v-if="getSingleOperateList(row).length > 0"
              class="ml-[10px]"
              theme="light"
              trigger="click"
              :popover-options="{ clickContentAutoHide: true }"
            >
              <Button text>
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

    <!-- 侧边栏 -->
    <processSideslider
      v-model:is-show="isShowSideslider"
      type="plugin"
      :plugin="currentPlugin"
    ></processSideslider>
    <!-- 操作确认弹窗（启动/重启/卸载/停止） -->
    <operate-dialog
      v-model:is-show="operateDialogIsShow"
      :title="operateDialogData.title"
      :type="operateDialogData.type"
      :sub-title="operateDialogData.subTitle"
      :data="pendingOperateData"
      :columns="operateDialogColumns"
      :status-map="processStatusTextMap"
      :hide-restart-options="true"
      :selection-confirm-formatter="selectionConfirmFormatter"
      @confirm="handleOperateConfirm"
      ></operate-dialog>
  </div>
</template>

<script setup lang="ts">
import { Button, Dropdown, Loading, Message } from 'bkui-vue';
import { computed, onMounted, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';

import { Table, TableColumn } from '@blueking/table';

import processSideslider from './process-sideslider.vue';

import { PluginAPIService } from '@/api/modules/plugin';
import { ProcessAPIService } from '@/api/modules/process';
import OperateDialog from '@/components/operate-dialog.vue';
import useTableSetting from '@/composables/use-table-setting';
import { useAuthStore } from '@/stores/auth';
import { usePermissionStore } from '@/stores/permission';
import { useMainStore } from '@/stores/main';

const router = useRouter();
const mainStore = useMainStore();
const { t } = useI18n();

const authStore = useAuthStore();
const permissionStore = usePermissionStore();
const maxHeight = computed(() => mainStore.windowInnerHeight - 214 - (mainStore.noticeShow ? 40 : 0));
// 插件列表数据
const pluginList = ref<any[]>([]);

// 表格
const { isShowSetting, settings, handleSettingChange } = useTableSetting(
  {
    checked: [
      'name',
      'pkg_nama',
      'group',
      'memo',
      'node_num',
      'operation',
    ],
    disabled: ['workflow_id'],
  },
  'nodeMng-plugin-status',
);

// 分页
const pagination = reactive({ count: 0, limit: 50, current: 1, remote: true });

const pageLimitChange = async (limit: number) => {
  pagination.limit = limit;
  pagination.current = 1;
  await loadPluginList();
};

const pageValueChange = async (current: number) => {
  pagination.current = current;
  await loadPluginList();
};

// ---------- 操作相关 ----------
// 需要跳转到单独页面配置自定义参数的操作: 安装、重装、升级、重载
// 无需跳转的操作: 卸载、启动、重启、停止
const needPageOperations = ['install', 'reinstall', 'upgrade', 'reload'];

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
      // 接口返回 reconfigv 对应前端 reload 操作
      if (p === 'reconfigv') ops.add('reload');
      // 接口返回 install 权限时，同时启用 reinstall（重装）操作
      if (p === 'install') ops.add('reinstall');
    });
  });
  enabledOperations.value = ops;
};

// 所有操作列表（单行操作使用） — 只显示有权限的操作
const allOperations = [
  { id: 'reinstall', nameKey: 'pluginManagement.plugin.operate.reinstall' },
  { id: 'upgrade', nameKey: 'pluginManagement.plugin.operate.upgrade' },
  { id: 'reload', nameKey: 'pluginManagement.plugin.operate.reload' },
  { id: 'uninstall', nameKey: 'pluginManagement.plugin.operate.uninstall' },
  { id: 'start', nameKey: 'pluginManagement.plugin.operate.start' },
  { id: 'restart', nameKey: 'pluginManagement.plugin.operate.restart' },
  { id: 'stop', nameKey: 'pluginManagement.plugin.operate.stop' },
];

// 单行操作列表，根据 plugin group 区分，只显示有权限的操作
const getSingleOperateList = (row: any) => {
  const ops = row.group === 'strategy'
    ? [{ id: 'restart', nameKey: 'pluginManagement.plugin.operate.restart' }, { id: 'stop', nameKey: 'pluginManagement.plugin.operate.stop' }]
    : allOperations;
  return ops
    .filter(op => op.id !== 'reinstall') // 重装已提到下拉外
    .filter(op => enabledOperations.value.has(op.id))
    .map(op => ({ id: op.id, name: t(op.nameKey) }));
};

// 操作确认弹窗
const operateDialogIsShow = ref(false);
const isBatchOperate = ref(false);
const operateDialogData = reactive({ type: '', title: '', subTitle: '' });
const pendingOperateData = ref<any[]>([]);
const pendingOperateType = ref('');

// 操作弹窗表格列配置（展示受影响的插件进程列表，与进程侧边栏一致）
const operateDialogColumns = [
  { field: 'plugin_name', label: t('pluginManagement.plugin.table.pluginName'), minWidth: 250 },
  { field: 'bk_host_id', label: 'Host ID', minWidth: 200 },
  { field: 'status', label: t('pluginManagement.plugin.process.processStatus'), minWidth: 200 },
];

// 进程状态映射（用于 operate-dialog 的 status 列渲染）
const processStatusTextMap = computed(() => ({
  running: t('pluginManagement.plugin.process.status.running'),
  stopped: t('pluginManagement.plugin.process.status.stopped'),
  unregister: t('pluginManagement.plugin.process.status.unregister'),
  init: t('pluginManagement.plugin.process.status.init'),
  unknown: t('pluginManagement.plugin.process.status.unknown'),
}));

// 勾选确认提示格式化：按插件分组展示 host_id
const selectionConfirmFormatter = (selectedRows: any[]) => {
  if (selectedRows.length === 0) return '';
  // 按 plugin_name 分组，拼装 "插件A的host1, host2、插件B的host3"
  const groupMap = new Map<string, string[]>();
  selectedRows.forEach((row: any) => {
    const name = row.plugin_name || '';
    const hostId = String(row.bk_host_id || '');
    if (!groupMap.has(name)) groupMap.set(name, []);
    groupMap.get(name)!.push(hostId);
  });
  const detail = [...groupMap.entries()]
    .map(([name, ids]) => `${name}(${ids.join(', ')})`)
    .join('、');
  const count = selectedRows.length;

  // 根据操作类型选择对应的完整模板
  const keyMap: Record<string, string> = {
    uninstall: 'pluginManagement.plugin.operate.selectionConfirmUninstall',
    start: 'pluginManagement.plugin.operate.selectionConfirmStart',
    restart: 'pluginManagement.plugin.operate.selectionConfirmRestart',
    stop: 'pluginManagement.plugin.operate.selectionConfirmStop',
  };
  const key = keyMap[pendingOperateType.value] || keyMap.restart;
  return t(key, { detail, count });
};

// 处理插件操作
const handlePluginOperate = async (operateType: string, data: any[], batch = false) => {
  if (needPageOperations.includes(operateType)) {
    // 安装/重装/升级/重载 → 跳转到配置页面
    const pluginName = data[0]?.name || '';

    // 重装/升级单行操作：从 process/list 获取主机和版本信息回填
    let prefillHosts = '';
    let prefillVersions = '';
    if ((operateType === 'reinstall' || operateType === 'upgrade') && !batch && pluginName) {
      const processRes = await ProcessAPIService.ListProcesses({
        page: { limit: 500, offset: 0 },
        exact_include_conditions: {
          plugin_name: [pluginName],
          bk_biz_id: mainStore.selectedBusinessId || [],
        } as any,
      }).catch(() => ({ total: 0, items: [] }));

      const processItems = (processRes.items || []).map((item: any) => ({
        ...item,
        ...item.platform,
        ...item.process_info,
        ...item.process_identity,
        ...item.process_controller,
      }));

      // 提取去重的主机ID（IP选择器通过 fetchHostDetails 自动回填完整信息）
      const hostIdSet = new Set<number>();
      processItems.forEach((item: any) => {
        if (item.bk_host_id) hostIdSet.add(item.bk_host_id);
      });
      prefillHosts = JSON.stringify([...hostIdSet]);

      // 提取版本映射（os_type_cpu_arch → version）
      const versionMap = new Map<string, string>();
      processItems.forEach((item: any) => {
        const osType = item.os_type || '';
        const cpuArch = item.cpu_arch || '';
        const version = item.version || '';
        if (osType && cpuArch && version) {
          const key = `${osType}_${cpuArch}`;
          // 只取第一个匹配的版本（同平台可能有多个版本，取进程当前版本）
          if (!versionMap.has(key)) {
            versionMap.set(key, version);
          }
        }
      });
      prefillVersions = JSON.stringify(Object.fromEntries(versionMap));
    }

    router.push({
      name: 'pluginOperate',
      query: {
        operationType: operateType,
        pluginName,
        ...(prefillHosts ? { prefillHosts } : {}),
        ...(prefillVersions ? { prefillVersions } : {}),
      },
    });
  } else if (operateType === 'restart' || operateType === 'uninstall' || operateType === 'start' || operateType === 'stop') {
    // 重启/卸载/启动/停止 → 先获取进程数据，再展示 operate-dialog
    pendingOperateType.value = operateType;
    isBatchOperate.value = batch;
    operateDialogData.type = operateType;
    // 根据插件名获取进程列表
    const pluginNameList = [...new Set(data.map((item: any) => item.name))];
    const processRes = await ProcessAPIService.ListProcesses({
      page: { limit: 500, offset: 0 },
      exact_include_conditions: {
        plugin_name: pluginNameList,
        bk_biz_id: mainStore.selectedBusinessId,
      },
    }).catch(() => ({ total: 0, items: [] }));

    pendingOperateData.value = (processRes.items || []).map((item: any) => ({
      ...item,
      ...item.platform,
      ...item.process_info,
      ...item.process_identity,
      ...item.process_controller,
    }));

    if (operateType === 'restart') {
      operateDialogData.title = batch
        ? t('pluginManagement.plugin.operate.batchRestartTitle')
        : t('pluginManagement.plugin.operate.restartTitle');
      operateDialogData.subTitle = t('pluginManagement.plugin.operate.selectProcessHint', { action: t('pluginManagement.plugin.operate.actionRestart') });
    } else if (operateType === 'uninstall') {
      operateDialogData.title = batch
        ? t('pluginManagement.plugin.operate.batchUninstallTitle')
        : t('pluginManagement.plugin.operate.uninstallTitle');
      operateDialogData.subTitle = t('pluginManagement.plugin.operate.selectProcessHint', { action: t('pluginManagement.plugin.operate.actionUninstall') });
    } else if (operateType === 'stop') {
      operateDialogData.title = batch
        ? t('pluginManagement.plugin.operate.batchStopTitle')
        : t('pluginManagement.plugin.operate.stopTitle');
      operateDialogData.subTitle = t('pluginManagement.plugin.operate.selectProcessHint', { action: t('pluginManagement.plugin.operate.actionStop') });
    } else {
      operateDialogData.title = batch
        ? t('pluginManagement.plugin.operate.batchStartTitle')
        : t('pluginManagement.plugin.operate.startTitle');
      operateDialogData.subTitle = t('pluginManagement.plugin.operate.selectProcessHint', { action: t('pluginManagement.plugin.operate.actionStart') });
    }
    operateDialogIsShow.value = true;
  }
};

// 操作弹窗确认（启动/重启/卸载/停止）
const handleOperateConfirm = async (extraData: any = {}) => {
  // 使用弹窗中勾选的行，若无勾选则使用全部
  const data = extraData.selection?.length ? extraData.selection : pendingOperateData.value;
  const operateType = pendingOperateType.value;

  // 按弹窗中实际选择的进程所归属业务校验 plugin_operate 权限：任一业务无权限则弹申请并拦截
  const bizIds = [...new Set(data.map((item: any) => item.bk_biz_id).filter((id: any) => id != null && id !== '' && id !== 0))];
  if (bizIds.length) {
    await authStore.batchVerify(
      [{ id: 'plugin_operate', action: 'plugin_operate', resourceType: 'biz', routes: [] }],
      bizIds,
    );
    if (!authStore.hasPermission('plugin_operate', bizIds)) {
      const detail = authStore.permissionDetail;
      if (detail) permissionStore.showDialog(detail);
      return;
    }
  }

  let res: any;

  const pluginParams = data.map((item: any) => ({ bk_host_id: item.bk_host_id, plugin_name: item.plugin_name }));

  if (operateType === 'uninstall') {
    res = await PluginAPIService.UninstallPlugin({ plugin: pluginParams });
  } else if (operateType === 'start') {
    res = await PluginAPIService.StartPlugin({ plugin: pluginParams });
  } else if (operateType === 'stop') {
    res = await PluginAPIService.StopPlugin({ plugin: pluginParams });
  } else {
    // 默认重启
    res = await PluginAPIService.RestartPlugin({ plugin: pluginParams });
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

// 跳转插件包管理
const handleGoToPluginPkgMng = (row: any) => {
  router.push({
    name: 'pluginPackageMng',
    query: {
      name: row.pkg_name,
    },
  });
};

// 侧边栏相关状态
const currentPlugin = ref<any>(null);

const loading = ref(false);
// 加载插件列表
const getParams = () => {
  const params = {
    page: { limit: pagination.limit, offset: (pagination.current - 1) * pagination.limit },
    exact_include_conditions: {
      visible_biz_ids: mainStore.selectedBusinessId,
    } as any,
    fuzzy_include_conditions: {} as Record<string, string[]>,
  };
  return params;
};
const loadPluginList = async () => {
  loading.value = true;
  try {
    const response = await PluginAPIService.ListPlugins(getParams()).catch((err) => {
      console.error('获取插件列表失败:', err);
      return {
        total: 0,
        items: [],
      };
    });

    const nodeNumMap = await ProcessAPIService.GetProcessDistributionByPluginName({
      exact_include_conditions: {
        plugin_name: response.items.map((item: any) => item.name),
        bk_biz_id: mainStore.selectedBusinessId,
      },
    }).catch((err: any) => {
      console.error('获取插件数量失败:', err);
      return {} as Record<number, number>;
    });

    pagination.count = response.total;
    pluginList.value = response.items.map((plugin: any) => ({
      ...plugin,
      editMemo: plugin.memo,
      isEdit: false,
      node_num: nodeNumMap[plugin.name] || 0,
    }));
  } catch (error) {
    console.error('加载插件列表失败:', error);
  } finally {
    loading.value = false;
  }
};

const isShowSideslider = ref(false);
// 打开侧边栏并加载进程列表
const openSidebar = async (plugin: any) => {
  isShowSideslider.value = true;
  currentPlugin.value = plugin;
};

// 组件挂载时加载数据
onMounted(() => {
  loadPermittedOperations();
  loadPluginList();
});

// 切换业务时重置分页并重新查询
watch(() => mainStore.selectedBusinessId, () => {
  pagination.current = 1;
  loadPluginList();
});
</script>

<style lang="postcss" scoped>
</style>
