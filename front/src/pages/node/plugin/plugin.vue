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
        <TableColumn title="插件名" field="name"></TableColumn>
        <TableColumn title="插件包名" field="pkg_name">
          <template #default="{ row }">
            <Button text theme="primary" @click="handleGoToPluginPkgMng(row)">
              {{ row.pkg_name }}
            </Button>
          </template>
        </TableColumn>
        <TableColumn title="插件组" field="group"></TableColumn>
        <TableColumn title="节点数" field="node_num">
          <template #default="{ row }">
            <Button text theme="primary" @click="openSidebar(row)">
              {{ row.node_num || 0 }}
              <i class="nodeman-icon nc-cloud-machine ml-[5px]"></i>
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
  </div>
</template>

<script setup lang="ts">
import { Button, Loading, Sideslider } from 'bkui-vue';
import { onMounted, reactive, ref } from 'vue';
import { useRouter } from 'vue-router';

import { Table, TableColumn } from '@blueking/table';

import processSideslider from './process-sideslider.vue';

import type { PluginListRespPlugin } from '@/@types/plugin';
import { PluginAPIService } from '@/api/modules/plugin';
import { ProcessAPIService } from '@/api/modules/process';
import useTableSetting from '@/composables/use-table-setting';

const router = useRouter();
// 插件列表数据
const pluginList = ref<Array<PluginListRespPlugin & { node_num: number }>>([]);

// 表格
const { isShowSetting, settings, handleSettingChange } = useTableSetting(
  {
    checked: [
      'name',
      'pkg_nama',
      'group',
      'node_num',
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
const handleGoToPluginPkgMng = (row: PluginListRespPlugin) => {
  router.push({
    name: 'pluginPackageMng',
    query: {
      name: row.pkg_name,
    },
  });
};

// 侧边栏相关状态
const currentPlugin = ref<PluginListRespPlugin | null>(null);

const loading = ref(false);
// 加载插件列表
const getParams = () => {
  const params = {
    page: {
      limit: 0,
      offset: 0,
    },
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
    pluginList.value = response.items.map((plugin: PluginListRespPlugin) => ({
      ...plugin,
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
const openSidebar = async (plugin: PluginListRespPlugin) => {
  isShowSideslider.value = true;
  currentPlugin.value = plugin;
};

// 组件挂载时加载数据
onMounted(() => {
  loadPluginList();
});
</script>
