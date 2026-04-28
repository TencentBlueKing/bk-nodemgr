<template>
  <div class="p-[24px]">
    <!-- 操作栏：安装按钮 + 批量操作 -->
    <div class="flex items-center gap-[8px] mb-[16px]">
      <Button v-if="enabledOperations.has('install')" class="w-[130px]" theme="primary" @click="handlePluginOperate('install', [], false)">
        {{ $t('pluginManagement.plugin.operate.install') }}
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
        @checkbox-change="handleSelectChange"
        @checkbox-all="handleSelectAllChange"
      >
        <TableColumn type="checkbox" width="60" fixed="left" />
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
            <!-- 编辑按钮 -->
            <Button
              v-if="hasPluginOperateAuth"
              text
              theme="primary"
              @click="handleEditInfo(row)"
            >
              {{ $t('action.edit') }}
            </Button>
            <span
              v-else
              class="inline-flex items-center auth-lock-wrapper"
              @click="handleAuthClick"
              @mouseenter="authLockMouseEnter($event, false)"
              @mousemove="authLockMouseMove($event, false)"
              @mouseleave="authLockMouseLeave()"
            >
              <Button theme="primary" text class="auth-disabled-text-btn">
                {{ $t('action.edit') }}
              </Button>
            </span>
            <!-- 更多操作下拉（启停重启等） -->
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
    <!-- 编辑弹窗 -->
    <editDialog
      v-model:is-show="isShowEditDialog"
      :plugin="currentPlugin"
      @confirm="handleEditInfoConfirm">
    </editDialog>
    <!-- 操作确认弹窗（重启） -->
    <operate-dialog
      v-model:is-show="operateDialogIsShow"
      :title="operateDialogData.title"
      :type="operateDialogData.type"
      :sub-title="operateDialogData.subTitle"
      @confirm="handleOperateConfirm"
    ></operate-dialog>
  </div>
</template>

<script setup lang="ts">
import { Button, Dropdown, InfoBox, Loading, Message } from 'bkui-vue';
import { computed, onMounted, reactive, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';

import { Table, TableColumn } from '@blueking/table';

import editDialog from './edit-dialog.vue';
import processSideslider from './process-sideslider.vue';

import { PluginAPIService } from '@/api/modules/plugin';
import useAuthLock from '@/composables/use-auth-lock';
import { ProcessAPIService } from '@/api/modules/process';
import OperateDialog from '@/components/operate-dialog.vue';
import useTableSetting from '@/composables/use-table-setting';
import { useMainStore } from '@/stores/main';

const router = useRouter();
const mainStore = useMainStore();
const { t } = useI18n();

// ===== plugin_operate 权限控制（编辑操作） =====
const {
  hasAuth: hasPluginOperateAuth,
  handleMouseEnter: authLockMouseEnter,
  handleMouseMove: authLockMouseMove,
  handleMouseLeave: authLockMouseLeave,
  handleAuthClick,
} = useAuthLock('plugin_operate', () => mainStore.selectedBusinessId);
const maxHeight = computed(() => mainStore.windowInnerHeight - 214 - (mainStore.noticeShow ? 40 : 0));
// 插件列表数据
const pluginList = ref<any[]>([]);

// 表格选中
const selection = computed(() => pluginList.value.filter((item: any) => item.checked));
const handleSelectChange = ({ checked, row }: { checked: boolean; row: any }) => {
  row.checked = checked;
};
const handleSelectAllChange = ({ checked }: { checked: boolean }) => {
  pluginList.value.forEach((item: any) => (item.checked = checked));
};

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
// 需要跳转到单独页面配置自定义参数的操作: 安装、升级、重载
// 无需跳转的操作: 卸载、重启、停止
const needPageOperations = ['install', 'upgrade', 'reload'];

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
    });
  });
  enabledOperations.value = ops;
};

// 批量操作列表 (default组) — 只显示有权限的操作
const allOperations = [
  { id: 'install', nameKey: 'pluginManagement.plugin.operate.install' },
  { id: 'upgrade', nameKey: 'pluginManagement.plugin.operate.upgrade' },
  { id: 'reload', nameKey: 'pluginManagement.plugin.operate.reload' },
  { id: 'uninstall', nameKey: 'pluginManagement.plugin.operate.uninstall' },
  { id: 'restart', nameKey: 'pluginManagement.plugin.operate.restart' },
  { id: 'stop', nameKey: 'pluginManagement.plugin.operate.stop' },
];

const batchOperateList = computed(() =>
  allOperations
    .filter(op => enabledOperations.value.has(op.id))
    .map(op => ({ id: op.id, name: t(op.nameKey) })),
);

// 单行操作列表，根据 plugin group 区分，只显示有权限的操作
const getSingleOperateList = (row: any) => {
  const ops = row.group === 'strategy'
    ? [{ id: 'restart', nameKey: 'pluginManagement.plugin.operate.restart' }, { id: 'stop', nameKey: 'pluginManagement.plugin.operate.stop' }]
    : allOperations;
  return ops
    .filter(op => enabledOperations.value.has(op.id))
    .map(op => ({ id: op.id, name: t(op.nameKey) }));
};

// 操作确认弹窗
const operateDialogIsShow = ref(false);
const operateDialogData = reactive({ type: '', title: '', subTitle: '' });
const pendingOperateData = ref<any[]>([]);
const pendingOperateType = ref('');

// 处理插件操作
const handlePluginOperate = (operateType: string, data: any[], batch = false) => {
  const pluginNames = [...new Set(data.map((item: any) => item.name))].join(', ');
  const count = data.length;

  if (needPageOperations.includes(operateType)) {
    // 安装/升级/重载 → 跳转到配置页面
    router.push({
      name: 'pluginOperate',
      query: {
        operationType: operateType,
        pluginName: data[0]?.name || '',
      },
    });
  } else if (operateType === 'restart') {
    // 重启 → 使用 operate-dialog
    pendingOperateData.value = data;
    pendingOperateType.value = operateType;
    operateDialogIsShow.value = true;
    operateDialogData.type = 'restart';
    operateDialogData.title = batch
      ? t('pluginManagement.plugin.operate.batchRestartTitle')
      : t('pluginManagement.plugin.operate.restartTitle');
    operateDialogData.subTitle = batch
      ? t('pluginManagement.plugin.operate.batchRestartSubTitle', { pluginNames, count })
      : t('pluginManagement.plugin.operate.restartSubTitle', { pluginName: pluginNames });
  } else if (operateType === 'uninstall') {
    // 卸载 → 使用 InfoBox 确认
    InfoBox({
      title: batch
        ? t('pluginManagement.plugin.operate.batchUninstallTitle')
        : t('pluginManagement.plugin.operate.uninstallTitle'),
      subTitle: batch
        ? t('pluginManagement.plugin.operate.batchUninstallSubTitle', { pluginNames, count })
        : t('pluginManagement.plugin.operate.uninstallSubTitle', { pluginName: pluginNames }),
      onConfirm: () => {
        // TODO: 调用卸载接口
        Message({ theme: 'success', message: t('pluginManagement.plugin.operate.operateSuccess') });
      },
    });
  } else if (operateType === 'stop') {
    // 停止 → 使用 InfoBox 确认
    InfoBox({
      title: batch
        ? t('pluginManagement.plugin.operate.batchStopTitle')
        : t('pluginManagement.plugin.operate.stopTitle'),
      subTitle: batch
        ? t('pluginManagement.plugin.operate.batchStopSubTitle', { pluginNames, count })
        : t('pluginManagement.plugin.operate.stopSubTitle', { pluginName: pluginNames }),
      onConfirm: () => {
        // TODO: 调用停止接口
        Message({ theme: 'success', message: t('pluginManagement.plugin.operate.operateSuccess') });
      },
    });
  }
};

// 操作弹窗确认 (重启)
const handleOperateConfirm = (extraData: any = {}) => {
  // TODO: 调用重启接口
  Message({ theme: 'success', message: t('pluginManagement.plugin.operate.operateSuccess') });
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

// 编辑操作
const isShowEditDialog = ref(false);
const handleEditInfo = (row: any) => {
  isShowEditDialog.value = true;
  currentPlugin.value = row;
};
const handleEditInfoConfirm = async () => {
  isShowEditDialog.value = false;
  await loadPluginList();
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
</script>

<style lang="postcss" scoped>
/* 无权限按钮：模拟 disabled 视觉效果但保持鼠标事件可响应 */
.auth-lock-wrapper {
  cursor: pointer;

  .auth-disabled-text-btn {
    color: #c4c6cc !important;
    cursor: pointer !important;
    pointer-events: auto !important;

    &:hover {
      color: #c4c6cc !important;
      background: transparent !important;
    }
  }
}
</style>
