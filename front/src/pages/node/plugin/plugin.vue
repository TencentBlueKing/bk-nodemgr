<template>
  <div class="p-[24px]">
    <Loading class="mb-[20px]" :loading="loading">
      <Table
        :data="pluginList"
        :empty-text="$t('table.empty')"
        :show-settings="isShowSetting"
        :pagination="pagination"
        :settings="settings"
        @setting-change="handleSettingChange"
        @page-limit-change="pageLimitChange"
        @page-value-change="pageValueChange"
      >
        <TableColumn title="插件名" field="name" min-width="150"></TableColumn>
        <TableColumn title="插件包名" field="pkg_name" min-width="150">
          <template #default="{ row }">
            <Button text theme="primary" @click="handleGoToPluginPkgMng(row)">
              {{ row.pkg_name }}
            </Button>
          </template>
        </TableColumn>
        <TableColumn title="插件组" field="group" min-width="120"></TableColumn>
        <TableColumn title="备注" field="memo" min-width="380"></TableColumn>
        <TableColumn title="节点数" field="node_num" min-width="120">
          <template #default="{ row }">
            <Button text theme="primary" @click="openSidebar(row)">
              {{ row.node_num || 0 }}
              <i class="nodeman-icon nc-cloud-machine ml-[5px]"></i>
            </Button>
          </template>
        </TableColumn>
        <TableColumn title="操作" field="operation" min-width="150">
          <template #default="{ row }">
            <Button text theme="primary" @click="handleEditInfo(row)">
              {{ $t('action.edit') }}
            </Button>
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
import { nextTick, onMounted, reactive, ref } from 'vue';
import { useRouter } from 'vue-router';

import { Table, TableColumn } from '@blueking/table';

import editDialog from './edit-dialog.vue';
import processSideslider from './process-sideslider.vue';

import { PluginAPIService } from '@/api/modules/plugin';
import { ProcessAPIService } from '@/api/modules/process';
import useTableSetting from '@/composables/use-table-setting';

const router = useRouter();
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
      console.log(err);
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
