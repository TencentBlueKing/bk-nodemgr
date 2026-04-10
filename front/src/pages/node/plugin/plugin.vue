<template>
  <div class="p-[24px]">
    <Loading class="mb-[20px]" :loading="loading">
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
          min-width="150"
        >
          <template #default="{ row }">
            <!-- 编辑按钮：有权限正常点击，无权限置灰+hover带锁+点击申请权限 -->
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
    <!-- 弹窗 -->
    <editDialog
      v-model:is-show="isShowEditDialog"
      :plugin="currentPlugin"
      @confirm="handleEditInfoConfirm">
    </editDialog>
  </div>
</template>

<script setup lang="ts">
import { Button, Input, Loading, Sideslider } from 'bkui-vue';
import { computed, nextTick, onMounted, reactive, ref } from 'vue';
import { useRouter } from 'vue-router';

import { Table, TableColumn } from '@blueking/table';

import editDialog from './edit-dialog.vue';
import processSideslider from './process-sideslider.vue';

import { PluginAPIService } from '@/api/modules/plugin';
import useAuthLock from '@/composables/use-auth-lock';
import { ProcessAPIService } from '@/api/modules/process';
import useTableSetting from '@/composables/use-table-setting';
import { useMainStore } from '@/stores/main';

const router = useRouter();
const mainStore = useMainStore();

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
  pagination.current = 1; // 页码重置为1
  await loadPluginList(); // 分页变化不防抖，立即执行
};

const pageValueChange = async (current: number) => {
  pagination.current = current;
  await loadPluginList(); // 分页变化不防抖，立即执行
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

// 编辑备注
// const memoInput = ref<any>(null);
// const handleEdit = (row: any) => {
//   row.isEdit = true;
//   nextTick(() => {
//     memoInput.value?.focus();
//   });
// };
// const handleEditConfirm = async (row: any) => {
//   row.isEdit = false;
//   row.memo = row.editMemo.trim();
//   await PluginAPIService.SetPluginMemo({
//     plugin_name: row.name,
//     memo: row.memo,
//   });
//   await loadPluginList();
// };
// const handleEditCancel = (row: any) => {
//   row.isEdit = false;
//   row.editMemo = row.memo;
// };
// const handleBlur = (row: any) => {
//   // 使用setTimeout延迟处理，避免与handleEditConfirm事件冲突
//   setTimeout(() => {
//     row.isEdit = false;
//     row.editMemo = row.memo;
//   }, 150);
// };

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
    exact_include_conditions: {},
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
