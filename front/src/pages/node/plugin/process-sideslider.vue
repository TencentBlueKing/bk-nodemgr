<template>
  <Sideslider
    v-model:is-show="isShow"
    render-directive="if"
    :title="title"
    width="1200"
  >
    <div class="p-[24px]">
      <copy-ip-dropdown
        v-if="type === 'plugin'"
        class="mb-[20px]"
        :type="'agent'"
        :disabled="!selection.length"
        :data="processList"
        :list="[]"
      ></copy-ip-dropdown>
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
        >
          <TableColumn v-if="type === 'plugin'" type="checkbox" width="60" fixed="left"></TableColumn>
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
            title="插件名"
            field="plugin_name"
            min-width="150"
            :filter="filterOptionSource.plugin_name"
          ></TableColumn>
          <TableColumn
            title="插件包名"
            field="plugin_pkg_name"
            min-width="150"
            :filter="filterOptionSource.plugin_pkg_name"
          ></TableColumn>
          <TableColumn
            title="插件组"
            field="plugin_group"
            min-width="120"
            :filter="filterOptionSource.plugin_group"
          ></TableColumn>
          <TableColumn
            title="操作系统"
            field="os_type"
            min-width="120"
            :filter="filterOptionSource.os_type"
          ></TableColumn>
          <TableColumn
            title="CPU 架构"
            field="cpu_arch"
            min-width="120"
            :filter="filterOptionSource.cpu_arch"
          ></TableColumn>
          <TableColumn title="Pid" field="pid" min-width="100"></TableColumn>
          <TableColumn
            title="版本"
            field="version"
            min-width="120"
            :filter="filterOptionSource.version"
          ></TableColumn>
          <TableColumn
            title="Agent Id"
            field="agent_id"
            min-width="120"
          ></TableColumn>
          <TableColumn
            title="进程名"
            field="name"
            min-width="120"
          ></TableColumn>
          <TableColumn
            title="进程状态"
            field="status"
            min-width="120"
            :filter="filterOptionSource.status"
          >
            <template #default="{ row }">
              <div class="flex items-center">
                <i
                  :class="`nodeman-icon nc-${
                    statusMap[row.status]?.icon
                  } status-icon`"
                ></i>
                <span>{{ statusMap[row.status]?.text || "--" }}</span>
              </div>
            </template>
          </TableColumn>
          <TableColumn
            title="安装路径"
            field="setup_path"
            min-width="120"
          ></TableColumn>
          <TableColumn
            title="Pid 文件路径"
            field="pid_path"
            min-width="120"
          ></TableColumn>
          <TableColumn
            title="配置文件路径"
            field="config_path"
            min-width="120"
          ></TableColumn>
          <TableColumn
            title="日志文件夹路径"
            field="log_path"
            min-width="130"
          ></TableColumn>
          <TableColumn
            title="进程所属系统账户"
            field="user"
            min-width="150"
          ></TableColumn>
          <TableColumn
            title="启动命令"
            field="start_cmd"
            min-width="180"
          ></TableColumn>
          <TableColumn
            title="停止命令"
            field="stop_cmd"
            min-width="180"
          ></TableColumn>
          <TableColumn
            title="重启命令"
            field="restart_cmd"
            min-width="180"
          ></TableColumn>
          <TableColumn
            title="Reload 命令"
            field="reload_cmd"
            min-width="180"
          ></TableColumn>
          <TableColumn
            title="Kill命令"
            field="kill_cmd"
            min-width="180"
          ></TableColumn>
          <TableColumn
            title="进程版本查询命令"
            field="version_cmd"
            min-width="180"
          ></TableColumn>
          <TableColumn
            title="进程健康检查命令"
            field="health_cmd"
            min-width="180"
          ></TableColumn>
          <TableColumn
            title="CPU 使用率上限百分比（总占比，非单核占比）"
            field="cpu_limit_percent"
            min-width="120"
          ></TableColumn>
          <TableColumn
            title="Mem 使用率上限百分比"
            field="mem_limit_percent"
            min-width="120"
          ></TableColumn>
          <TableColumn
            title="重启策略"
            field="auto_type"
            min-width="120"
          ></TableColumn>
          <TableColumn
            title="启动后延迟检查的时间"
            field="start_check_seconds"
            min-width="120"
          >
            <template #default="{ row }">
              <span>{{ row.start_check_seconds }}s</span>
            </template>
          </TableColumn>
          <TableColumn
            title="停止命令执行后开始检查进程存活的时间"
            field="stop_check_seconds"
            min-width="120"
          >
            <template #default="{ row }">
              <span>{{ row.stop_check_seconds }}s</span>
            </template>
          </TableColumn>
          <TableColumn
            title="命令执行超时时间"
            field="operate_timeout_seconds"
            min-width="120"
          >
            <template #default="{ row }">
              <span>{{ row.operate_timeout_seconds }}s</span>
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
import { TopoService } from '@/api/modules/topo';

import useTableSetting from '@/composables/use-table-setting';

import type { DistinctProcessRespData } from '@/@types/process';


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
      'plugin_name',
      'status',
      'pid',
      'version',
      'agent_id',
      ...(props.type === 'plugin' ? ['bk_host_id', 'plugin_group', 'bk_host_innerip', 'bk_host_innerip_v6'] : ['os_type', 'cpu_arch']),
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

// 表格勾选
const selection = computed(() => processList.value.filter((item: any) => item.checked));

const handleSelectChange = ({ checked, row }: { checked: boolean; row: any }) => {
  row.checked = checked;
};

const handleSelectAllChange = ({ checked }: { checked: boolean }) => {
  processList.value.forEach((item: any) => (item.checked = checked));
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

// distinct
const distinct = ref<DistinctProcessRespData>();
const getDistinct = async () => {
  const res = await ProcessAPIService.DistinctProcess({
    exact_include_conditions: {
      bk_host_id: props.type === 'node' ? [props.node.bk_host_id] : processList.value.map((item: any) => item.bk_host_id),
    },
  });
  if (res) {
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

  let hostListMap = new Map();
  if (props.type === 'plugin') {
    const hostList = await TopoService.HostList({
      exact_include_conditions: {
        bk_host_id: res.items.map(item => item.bk_host_id),
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
  getDistinct();
};

// 暂时没有编辑数据，不需要离开前确认
// const handleBeforeClose = (): Promise<boolean> => new Promise((resolve, reject) => {
//   InfoBox({
//     title: '确认关闭?',
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
<style lang="postcss">
.vxe-table--tooltip-wrapper {
  &.process-table {
    z-index: 2004 !important;
  }
}
</style>

