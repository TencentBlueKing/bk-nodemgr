<template>
  <Sideslider
    v-model:is-show="isShow"
    render-directive="if"
    :title="title"
    width="1200"
    :before-close="handleBeforeClose"
  >
    <Loading
      :title="$t('table.loading')"
      :loading="loading"
      class="p-[24px] overflow-auto">
      <Table
        :data="processList"
        :empty-text="$t('table.empty')"
        :show-settings="isShowSetting"
        :pagination="pagination"
        :settings="settings"
        @setting-change="handleSettingChange"
        @page-limit-change="pageLimitChange"
        @page-value-change="pageValueChange"
      >
        <TableColumn title="Host ID" field="bk_host_id" fixed="left"></TableColumn>
        <TableColumn title="插件名" field="plugin_name"></TableColumn>
        <TableColumn title="插件包名" field="plugin_pkg_name"></TableColumn>
        <TableColumn title="插件组" field="plugin_group"></TableColumn>
        <TableColumn title="操作系统" field="os_type"></TableColumn>
        <TableColumn title="CPU 架构" field="cpu_arch"></TableColumn>
        <TableColumn title="Pid" field="pid"></TableColumn>
        <TableColumn title="版本" field="version"></TableColumn>
        <TableColumn title="Agent Id" field="agent_id"></TableColumn>
        <TableColumn title="进程名" field="name"></TableColumn>
        <TableColumn title="进程状态" field="status">
          <template #default="{ row }">
            <div class="flex items-center">
              <i :class="`nodeman-icon nc-${statusMap[row.status]?.icon} status-icon`"></i>
              <span>{{ statusMap[row.status]?.text || '--' }}</span>
            </div>
          </template>
        </TableColumn>
        <TableColumn title="安装路径" field="setup_path"></TableColumn>
        <TableColumn title="Pid 文件路径" field="pid_path"></TableColumn>
        <TableColumn title="配置文件路径" field="config_path"></TableColumn>
        <TableColumn title="日志文件夹路径" field="log_path"></TableColumn>
        <TableColumn title="进程所属系统账户" field="user"></TableColumn>
        <TableColumn title="启动命令" field="start_cmd"></TableColumn>
        <TableColumn title="停止命令" field="stop_cmd"></TableColumn>
        <TableColumn title="重启命令" field="restart_cmd"></TableColumn>
        <TableColumn title="Reload 命令" field="reload_cmd"></TableColumn>
        <TableColumn title="Kill命令" field="kill_cmd"></TableColumn>
        <TableColumn title="进程版本查询命令" field="version_cmd"></TableColumn>
        <TableColumn title="进程健康检查命令" field="health_cmd"></TableColumn>
        <TableColumn title="CPU 使用率上限百分比（总占比，非单核占比）" field="cpu_limit_percent"></TableColumn>
        <TableColumn title="Mem 使用率上限百分比" field="mem_limit_percent"></TableColumn>
        <TableColumn title="重启策略" field="auto_type"></TableColumn>
        <TableColumn title="启动后延迟检查的时间" field="start_check_seconds">
          <template #default="{ row }">
            <span>{{ formatTimestamp(row.start_check_seconds) }}</span>
          </template>
        </TableColumn>
        <TableColumn title="停止命令执行后开始检查进程存活的时间" field="stop_check_seconds">
          <template #default="{ row }">
            <span>{{ formatTimestamp(row.stop_check_seconds) }}</span>
          </template>
        </TableColumn>
        <TableColumn title="命令执行超时时间" field="operate_timeout_seconds">
          <template #default="{ row }">
            <span>{{ formatTimestamp(row.operate_timeout_seconds) }}</span>
          </template>
        </TableColumn>
      </Table>
    </Loading>
    <template #footer>
      <Button class="mr-[8px]" theme="primary" @click="handleBeforeClose">
        {{ $t('action.confirm') }}
      </Button>
      <Button @click="handleBeforeClose">
        {{ $t('action.cancel') }}
      </Button>
    </template>
  </Sideslider>
</template>
<script lang="ts" setup>
import {
  Button,
  InfoBox,
  Loading,
  Sideslider,
} from 'bkui-vue';
import { computed, reactive, ref, watch } from 'vue';

import { Table, TableColumn } from '@blueking/table';

import { ProcessAPIService } from '@/api/modules/process';
import { formatTimestamp } from '@/common/util';
import useTableSetting from '@/composables/use-table-setting';

// Sideslider显示状态
const isShow = defineModel('isShow', { type: Boolean });

const props = defineProps({
  type: {
    type: String,
    default: 'plugin',
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

const title = computed(() => '插件进程列表');

// 表格
const { isShowSetting, settings, handleSettingChange } = useTableSetting(
  {
    checked: [
      'bk_host_id',
      'plugin_name',
      'status',
      'pid',
      'version',
      'agent_id',
      ...(props.type === 'plugin' ? ['plugin_group'] : ['os_type', 'cpu_arch']),
    ],
    disabled: [''],
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
    text: '健康',
    icon: 'running',
  },
  stopped: {
    text: '停止',
    icon: 'terminated',
  },
  unregister: {
    text: '未注册',
    icon: 'unknown',
  },
  init: {
    text: '初始化',
    icon: 'unknown',
  },
  unknown: {
    text: '未知',
    icon: 'unknown',
  },
};

// 进程列表
const processList = ref<any[]>([]);

const loading = ref(false);

const getProcessList = async () => {
  loading.value = true;
  const res = await ProcessAPIService.ListProcesses({
    page: { limit: pagination.limit, offset: (pagination.current - 1) * pagination.limit },
    exact_include_conditions: props.type === 'plugin' ? {
      plugin_name: [props.plugin.name],
      plugin_pkg_name: [props.plugin.pkg_name],
    } : {
      bk_host_id: [props.node.bk_host_id],
    },
    fuzzy_include_conditions: {},
  }).catch((err) => {
    console.log(err);
    return {
      total: 0,
      items: [],
    };
  });
  pagination.count = res.total;
  loading.value = false;
  processList.value = res.items.map(item => ({
    ...item,
    ...item.platform,
    ...item.process_info,
    ...item.process_identity,
    ...item.process_controller,
    ...item.process_resource,
    ...item.process_monitor_policy,
  }));
};

const handleBeforeClose = (): Promise<boolean> => new Promise((resolve, reject) => {
  InfoBox({
    title: '确认关闭?',
    infoType: 'warning',
    onConfirm: () => {
      resolve(true);
      isShow.value = false;
    },
    onCancel: () => reject(),
  });
});

watch(
  () => isShow.value,
  async () => {
    if (isShow.value) {
      await getProcessList();
    }
  },
  { immediate: true },
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
</style>
