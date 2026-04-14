<template>
  <Tab
    v-model:active="active"
    type="unborder-card"
    :label-height="41"
    class="text-[14px] bg-[#fff] h-[41px] absolute z-10 w-full"
  >
    <Tab.TabPanel
      v-for="item in panels"
      :key="item.name"
      :label="item.label"
      :name="item.name"
    >
    </Tab.TabPanel>
  </Tab>
  <div class="p-[24px] mt-[41px]">
    <section class="flex justify-between mb-[15px]">
      <div class="flex gap-[12px]">
        <Checkbox v-model="hideAutoTask">{{
          t("platform.nodeMan.taskHistory.button.hideAutoTask")
        }}</Checkbox>
        <DatePicker
          v-model="dateValue"
          :shortcut-selected-index="1"
          :shortcuts="shortcutsRange"
          format="yyyy-MM-dd HH:mm:ss"
          type="datetimerange"
          use-shortcut-text
          @change="pickSuccess"
        />
      </div>
      <div class="flex-1 ml-[8px]">
        <SearchSelect
          ref="searchSelect"
          class="bg-[#fff]"
          :data="searchSelectData"
          v-model.trim="searchSelectValue"
          :unique-select="true"
          :placeholder="
            t('platform.nodeMan.taskHistory.placeholder.listSearch')
          "
          @update:model-value="handleSearchSelectChange"
        >
        </SearchSelect>
      </div>
    </section>
    <bk-loading :title="t('table.loading')" :loading="loading">
      <Table
        class="filterTable"
        :data="tableData"
        :empty-text="$t('table.empty')"
        :pagination="pagination"
        :column-config="{ resizable: true }"
        show-overflow-tooltip
        :max-height="maxHeight"
        :show-settings="isShowSetting"
        :settings="settings"
        @setting-change="handleSettingChange"
        @column-filter="handleFilter"
        @page-limit-change="pageLimitChange"
        @page-value-change="pageValueChange"
      >
        <TableColumn
          field="workflow_id"
          :title="t('platform.nodeMan.taskHistory.label.taskID')"
          min-width="100"
          fixed="left"
        >
          <template #default="{ row }">
            <Button
              text
              theme="primary"
              @click="detailHandle(row)"
            >{{ "#" + row.workflow_id?.slice(-4) }}</Button
            >
          </template>
        </TableColumn>
        <TableColumn
          field="type"
          :title="t('platform.nodeMan.taskHistory.label.taskType')"
          :filter="filterOptionSource.type"
          min-width="150"
        >
          <template #default="{ row }">
            <span>{{ typeMap[row.type as taskType]?.text }}</span>
          </template>
        </TableColumn>
        <TableColumn
          v-if="['agent', 'proxy'].includes(active)"
          field="bk_biz_name"
          :title="t('platform.nodeMan.taskHistory.label.business')"
          min-width="150"
        ></TableColumn>
        <TableColumn
          field="operator"
          :title="t('platform.nodeMan.taskHistory.label.operator')"
          :filter="filterOptionSource.operator"
          min-width="150"
        ></TableColumn>
        <TableColumn
          field="operate_time"
          :title="t('platform.nodeMan.taskHistory.label.operateTime')"
          min-width="200"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            <span>{{ timeFormatter(row.operate_time) }}</span>
          </template>
        </TableColumn>
        <TableColumn
          field="cost_time"
          :title="t('platform.nodeMan.taskHistory.label.costTime')"
          min-width="100"
        >
          <template #default="{ row }">
            <span>{{ formatTimeToMS(row.cost_time) }}</span>
          </template>
        </TableColumn>
        <TableColumn
          field="status"
          :title="t('platform.nodeMan.taskHistory.label.status')"
          min-width="150"
          :filter="filterOptionSource.status"
        >
          <template #default="{ row }">
            <div
              class="flex items-center"
              v-if="row.status && statusMap[row.status]"
            >
              <Spinner v-if="row.status === 'running'" class="mr-[8px]" />
              <template v-else>
                <i
                  :class="`nodeman-icon nc-${
                    statusMap[row.status].icon
                  } status-icon`"
                ></i>
              </template>
              <span>{{ statusMap[row.status].text }}</span>
            </div>
            <div class="flex items-center" v-else>
              <span class="nodeman-icon nc-unknown status-icon"></span>
              <span>{{ row.status }}</span>
            </div>
          </template>
        </TableColumn>
        <TableColumn
          field="count"
          :title="t('platform.nodeMan.taskHistory.label.count')"
          min-width="150"
        >
          <template #default="{ row }">
            <template v-if="row.statistics">
              <span class="pr-[4px]">{{ row.statistics.total_count || 0 }}</span
              >/
              <a
                class="text-[#2dcb56] pr-[4px] cursor-pointer"
                @click.stop="detailHandle(row, 'success')"
              >{{ row.statistics.success_count || 0 }}</a
              >/
              <a
                class="text-[#ea3636] pr-[4px] cursor-pointer"
                @click.stop="detailHandle(row, 'failed')"
              >{{ row.statistics.failed_count || 0 }}</a
              >/
              <a
                class="text-[#ff9c01] pr-[4px] cursor-pointer"
                @click.stop="detailHandle(row, 'timeout')"
              >{{ row.statistics.timeout_count || 0 }}</a
              >
            </template>
            <span v-else>--</span>
          </template>
        </TableColumn>
      </Table>
    </bk-loading>
  </div>
</template>
<script setup lang="ts">
import {
  Button,
  Cascader,
  Checkbox,
  DatePicker,
  Dropdown,
  InfoBox,
  SearchSelect,
  Tab,
} from 'bkui-vue';
import { Spinner } from 'bkui-vue/lib/icon';
import dayjs from 'dayjs';
import { debounce } from 'lodash';
import { computed, onMounted, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';

import { Table, TableColumn } from '@blueking/table';

import type { NodeWorkflowDistinctRespData, NodeWorkflowInfo } from '@/@types/node_workflow';
import { NodeWorkflowService } from '@/api/modules/node_workflow';
import { PluginWorkflowService } from '@/api/modules/plugin_workflow';
import useTableSetting from '@/composables/use-table-setting';
import { useMainStore } from '@/stores/main';
import { useNodeManageStore } from '@/stores/node-manage';

interface FilterOption {
  list: { text: string; value: string }[];
  checked: string[];
  filterScope: string;
}
type taskType =
  | 'install_agent'
  | 'install_plugin'
  | 'upgrade_agent'
  | 'upgrade_plugin';
type filterProp = 'type' | 'operator' | 'status';

const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const mainStore = useMainStore();
const nodeManageStore = useNodeManageStore();
const tableData = ref<NodeWorkflowInfo[]>([]);
const maxHeight = computed(() => mainStore.windowInnerHeight - 214 - (mainStore.noticeShow ? 40 : 0));

// tab
const active = ref('agent');
const panels = ref([
  { name: 'agent', label: t('platform.nodeMan.taskHistory.tab.agentHistory') },
  { name: 'proxy', label: t('platform.nodeMan.taskHistory.tab.proxyHistory') },
  { name: 'plugin', label: t('platform.nodeMan.taskHistory.tab.pluginHistory') },
]);

// 分页
const pagination = reactive({ count: 0, limit: 50, current: 1, remote: true });

const pageLimitChange = async (limit: number) => {
  pagination.limit = limit;
  pagination.current = 1; // 页码重置为1
  await getTaskList(); // 分页变化不防抖，立即执行
};

const pageValueChange = async (current: number) => {
  pagination.current = current;
  await getTaskList(); // 分页变化不防抖，立即执行
};

// 跨页全选
const loading = ref(false);

// 日期选择
const dateValue = ref([
  new Date().setTime(new Date().getTime() - 3600 * 1000 * 24 * 7),
  new Date(),
]);
const shortcutsRange = reactive([
  {
    text: t('platform.nodeMan.taskHistory.shortcut.today'),
    value() {
      const end = new Date();
      const start = new Date(end.getFullYear(), end.getMonth(), end.getDate());
      return [start, end];
    },
  },
  {
    text: t('platform.nodeMan.taskHistory.shortcut.last7Days'),
    value() {
      const end = new Date();
      const start = new Date();
      start.setTime(start.getTime() - 3600 * 1000 * 24 * 7);
      return [start, end];
    },
  },
  {
    text: t('platform.nodeMan.taskHistory.shortcut.last15Days'),
    value() {
      const end = new Date();
      const start = new Date();
      start.setTime(start.getTime() - 3600 * 1000 * 24 * 15);
      return [start, end];
    },
  },
  {
    text: t('platform.nodeMan.taskHistory.shortcut.last30Days'),
    value() {
      const end = new Date();
      const start = new Date();
      start.setTime(start.getTime() - 3600 * 1000 * 24 * 30);
      return [start, end];
    },
  },
]);
const statusMap = computed(() => ({
  running: {
    text: t('platform.nodeMan.taskHistory.statusType.running'),
  },
  failed: {
    text: t('platform.nodeMan.taskHistory.statusType.failed'),
    icon: 'terminated',
  },
  success: {
    text: t('platform.nodeMan.taskHistory.statusType.success'),
    icon: 'running',
  },
  partial_failed: {
    text: t('platform.nodeMan.taskHistory.statusType.partial_failed'),
    icon: 'warning',
  },
}));
const typeMap = computed(() => ({
  install_agent: {
    text: t('platform.nodeMan.taskHistory.taskType.install_agent'),
  },
  upgrade_agent: {
    text: t('platform.nodeMan.taskHistory.taskType.upgrade_agent'),
  },
  reconfig_agent: {
    text: t('platform.nodeMan.taskHistory.taskType.reconfig_agent'),
  },
  restart_agent: {
    text: t('platform.nodeMan.taskHistory.taskType.restart_agent'),
  },
  uninstall_agent: {
    text: t('platform.nodeMan.taskHistory.taskType.uninstall_agent'),
  },
  install_proxy: {
    text: t('platform.nodeMan.taskHistory.taskType.install_proxy'),
  },
  upgrade_proxy: {
    text: t('platform.nodeMan.taskHistory.taskType.upgrade_proxy'),
  },
  reconfig_proxy: {
    text: t('platform.nodeMan.taskHistory.taskType.reconfig_proxy'),
  },
  restart_proxy: {
    text: t('platform.nodeMan.taskHistory.taskType.restart_proxy'),
  },
  uninstall_proxy: {
    text: t('platform.nodeMan.taskHistory.taskType.uninstall_proxy'),
  },
  install_plugin: {
    text: t('platform.nodeMan.taskHistory.taskType.install_plugin'),
  },
  upgrade_plugin: {
    text: t('platform.nodeMan.taskHistory.taskType.upgrade_plugin'),
  },
}));
const bussinessMap = computed(() => mainStore.businessList.map(item => ({
  id: item.bk_biz_id,
  name: item.bk_biz_name,
})));
const pickSuccess = async (val: string[]) => {
  await getTaskList();
};

// 隐藏自动部署任务
const hideAutoTask = ref(false);

const timeFormatter = (val: string, format = 'YYYY-MM-DD HH:mm:ss') => (val ? dayjs(val).format(format) : '--');

// 补零规则：非0且小于10时补零，0则直接显示0
const padIfNeeded = (num: number) => (num === 0 ? '0' : num < 10 ? `0${num}` : num.toString());
const formatTimeToMS = (duration: number) => {
  const minutes = Math.floor(duration / 60000);
  const seconds = padIfNeeded(Math.floor((duration % 60000) / 1000));
  return `${minutes}m ${seconds}s`;
};

// 表格
const { isShowSetting, settings, handleSettingChange } = useTableSetting({
  checked: [
    'workflow_id',
    'bk_biz_name',
    'type',
    'bk_policy_name',
    'operator',
    'operate_time',
    'cost_time',
    'status',
    'count',
  ],
  disabled: ['workflow_id'],
}, 'nodeMng-history');

const workflowDistinct = ref<NodeWorkflowDistinctRespData | null>();
function getUniqueChildrenFrom <K extends keyof NodeWorkflowDistinctRespData>(
  prop: K,
  keyMap?: Record<string, any>,
) {
  const uniqueValues = workflowDistinct.value?.[prop] || [];
  return uniqueValues
    .filter((item: any) => item !== '')
    .map((value: any) => ({
      id: value,
      name: keyMap?.[value].text || String(value),
    }));
}
const searchSelectData = computed(() => [
  { id: 'workflow_id', name: t('platform.nodeMan.taskHistory.label.taskID') },
  {
    id: 'bk_host_innerip',
    name: t('taskDetail.search.ipv4'),
    multiple: true,
  },
  {
    id: 'bk_host_innerip_v6',
    name: t('taskDetail.search.ipv6'),
    multiple: true,
  },
  {
    id: 'type',
    name: t('platform.nodeMan.taskHistory.label.taskType'),
    children: getUniqueChildrenFrom('type', typeMap.value),
  },
  ...(isNode.value ? [{
    id: 'bk_biz_id',
    name: t('platform.nodeMan.taskHistory.label.business'),
    children: bussinessMap.value,
    multiple: true,
  }] : []),
  {
    id: 'operator',
    name: t('platform.nodeMan.taskHistory.label.operator'),
    children: getUniqueChildrenFrom('operator'),
  },
  {
    id: 'status',
    name: t('platform.nodeMan.taskHistory.label.status'),
    children: getUniqueChildrenFrom('status', statusMap.value),
  },
]);
const filterOptionConfig = (prop: string, valMap?: Record<string, any>) => {
  const uniqueValues = Array.from(new Set(tableData.value.map((item: any) => item[prop]).filter((item: any) => item)));
  return uniqueValues.map(value => ({
    text:
      valMap && valMap.value && valMap.value[value as string] ? valMap.value[value as string].text : value,
    value,
  }));
};
// 搜索
const searchSelectValue = ref<{ id: string; name: string; values: any }[]>([]);
const handleSearchSelectChange = async (data: { id: string; name: string; values: { id: string; name: string }[] }[]) => {
  Object.keys(filterOptionSource).forEach((key) => {
    filterOptionSource[key].checked = [];
  });
  data.forEach((item) => {
    if (filterOptionSource[item.id as filterProp]) {
      filterOptionSource[item.id as filterProp].checked = item.values.map((item: any) => item.id) as string[];
    }
  });
};
// 筛选
const handleFilter = ({
  checked,
  field,
}: {
  checked: string[];
  field: string;
}) => {
  const index = searchSelectValue.value.findIndex((item: any) => item.id === field);
  index > -1 && searchSelectValue.value.splice(index, 1);
  if (checked.length) {
    searchSelectValue.value.push({
      id: field,
      name: t(field),
      values: checked.map((item: any) => {
        let name;
        switch (field) {
          case 'status':
            name = statusMap.value[item]?.text || item;
            break;
          case 'type':
            name = typeMap.value[item as taskType]?.text || item;
            break;
          default:
            name = item;
            break;
        }
        return {
          id: item,
          name,
        };
      }),
    });
  }
};
const filterOptionSource = reactive<Record<string, FilterOption>>({
  type: {
    list: [],
    checked: [],
    filterScope: 'all',
  },
  operator: {
    list: [],
    checked: [],
    filterScope: 'all',
  },
  status: {
    list: [],
    checked: [],
    filterScope: 'all',
  },
});

const getTimestampInSeconds = (originalDate: string) => {
  const timestampInMilliseconds = new Date(originalDate).getTime();
  const timestampInSeconds = Math.floor(timestampInMilliseconds / 1000);
  return timestampInSeconds;
};
const isNode = computed(() => ['agent', 'proxy'].includes(active.value));
const agentType = ['install_agent', 'upgrade_agent', 'reconfig_agent', 'restart_agent', 'uninstall_agent'];
const proxyType = ['install_proxy', 'upgrade_proxy', 'reconfig_proxy', 'restart_proxy', 'uninstall_proxy'];
const typeListMap = {
  agent: agentType,
  proxy: proxyType,
  plugin: [] as string[],
};
const getParams = () => {
  const params = {
    page: {
      limit: pagination.limit,
      offset: (pagination.current - 1) * pagination.limit,
    },
    exact_include_conditions: {
      bk_biz_id: mainStore.selectedBusinessId,
      type: typeListMap[active.value],
    },
    fuzzy_include_conditions: {} as Record<string, string[]>,
    operate_time_range: {
      start_timestamp_sec: getTimestampInSeconds(dateValue.value[0]),
      end_timestamp_sec: getTimestampInSeconds(dateValue.value[1]),
    },
  };
  searchSelectValue.value.forEach((item: any) => {
    const target = params.exact_include_conditions;
    target[item.id] = item.values?.map((value: any) => value.id);
  });
  return params;
};

// 获取搜索和筛选条件
const getWorkflowDistinct = async () => {
  const params = {
    exact_include_conditions: {
      bk_biz_id: mainStore.selectedBusinessId,
      type: typeListMap[active.value],
    },
  };
  let res: any;
  if (isNode.value) {
    res = await NodeWorkflowService.NodeWorkflowDistinct(params).catch((err: any) => {
      console.error('获取主机筛选条件唯一值失败:', err);
      return null;
    });
  } else {
    res = await PluginWorkflowService.PluginWorkflowDistinct(params).catch((err: any) => {
      console.error('获取主机筛选条件唯一值失败:', err);
      return null;
    });
  };
  if (res) {
    workflowDistinct.value = res;
    Object.keys(res).forEach((key: any) => {
      if (filterOptionSource[key]) {
        filterOptionSource[key].list = res[key]
          .filter((item: any) => item !== '')
          .map((value: string | number) => {
            let text = value;
            if (key === 'type') text = typeMap.value[value as string].text || value;
            if (key === 'status') text = statusMap.value[value as string].text || value;
            return { text, value };
          });
      }
    });
  }
};

const getTaskList = async () => {
  loading.value = true;
  let res;
  let statistics: any;
  if (isNode.value) {
    res = await NodeWorkflowService.NodeWorkflowList(getParams()).catch((err) => {
      console.log(err);
      return {
        total: 0,
        items: [],
      };
    });
    const workflowIds = res.items.map(item => item.workflow_id).filter(id => id);
    if (workflowIds.length > 0) {
      statistics = await NodeWorkflowService.NodeWorkflowStatistics({
        workflow_id: workflowIds,
      }).catch((err) => {
        console.log(err);
        return {
          items: [],
        };
      });
    } else {
      statistics = { items: [] };
    }
  } else {
    res = await PluginWorkflowService.PluginWorkflowList(getParams()).catch((err) => {
      console.log(err);
      return {
        total: 0,
        items: [],
      };
    });
    const workflowIds = res.items.map(item => item.workflow_id).filter(id => id);
    if (workflowIds.length > 0) {
      statistics = await PluginWorkflowService.PluginWorkflowStatistics({
        workflow_id: workflowIds,
      }).catch((err) => {
        console.log(err);
        return {};
      });
    } else {
      statistics = { items: [] };
    }
  };

  pagination.count = res.total;
  tableData.value = res.items.map((item) => {
    const statisticsItem = statistics.items?.find(statistic => statistic.workflow_id === item.workflow_id);
    return {
      statistics: statisticsItem,
      ...item,
      bk_biz_name: item.bk_biz_name?.filter(item => item),
      cost_time:
        item.finish_time > 0 ? item.finish_time - item.operate_time : 0,
    };
  });
  loading.value = false;
};
const debounceGetTaskList = debounce(() => {
  getTaskList();
}, 300);
// 跳转详情
const detailHandle = (row: NodeWorkflowInfo, status?: string) => {
  nodeManageStore.updateCurrentRowData(row);
  router.push({
    name: 'taskDetail',
    params: {
      taskId: row.workflow_id,
    },
    query: {
      active: active.value,
      status,
    },
  });
};
watch(
  () => route.query,
  () => {
    const newActive = route.query.active as string || 'agent';
    
    // 如果 URL 中没有 active 参数，主动设置一个（首次进入页面时）
    if (!route.query.active) {
      router.replace({
        query: {
          ...route.query,
          active: 'agent',
        },
      });
      return; // 等待下一次 watch 触发
    }
    
    // 仅当值真的不同时才更新，避免触发循环
    if (active.value !== newActive) {
      active.value = newActive;
    }
    
    // URL 变化时发起数据请求
    debounceGetTaskList.cancel();
    getTaskList();
    getWorkflowDistinct();
  },
  { immediate: true },
);

watch(
  [
    () => searchSelectValue,
    () => mainStore.selectedBusinessId,
  ],
  async () => {
    await debounceGetTaskList();
  },
  { immediate: true, deep: true },
);
// 监听 Tab 切换，只负责更新 URL
watch(
  () => active.value,
  (newActive) => {
    // 仅当 URL 不同时才更新，避免循环触发
    if (route.query.active !== newActive) {
      router.replace({
        query: {
          ...route.query,
          active: newActive,
        },
      });
    }
  },
);
watch(
  () => mainStore.selectedBusinessId,
  () => {
    getWorkflowDistinct();
  },
  { deep: true }, // 移除 immediate，避免首次加载时重复请求
);
</script>
<style lang="postcss" scoped>
:deep(.vxe-table--empty-content) {
  height: 200px;
  line-height: 200px;
}

.status-icon::before {
  content: "";
  display: inline-block;
  margin-right: 8px;
  width: 8px;
  height: 8px;
  border: 1px solid #f0f1f5;
  border-radius: 6.5px;
  background: #b2b5bd;
}
.nc-running {
  &::before {
    background: #cbf0da;
    border-color: #2caf5e;
  }
}
.nc-terminated {
  &::before {
    border-color: #ea3636;
    background: #ffdddd;
  }
}
.nc-warning {
  &::before {
    border-color: #f59500;
    background: #fce5c0;
  }
}
.nc-unknown {
  &::before {
    border-color: #b2b5bd;
    background: #f0f1f5;
  }
}
</style>
