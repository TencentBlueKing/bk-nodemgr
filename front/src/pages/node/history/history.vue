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
            t('platform.nodeMan.historySearchPlaceholder')
          "
          @update:model-value="handleSearchSelectChange"
          @paste.native="handleNativePaste"
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
import { computed, nextTick, onMounted, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { type LocationQuery, useRoute, useRouter } from 'vue-router';

import { Table, TableColumn } from '@blueking/table';

import type { NodeWorkflowDistinctRespData, NodeWorkflowInfo } from '@/@types/node_workflow';
import { NodeWorkflowService } from '@/api/modules/node_workflow';
import { PluginWorkflowService } from '@/api/modules/plugin_workflow';
import useTableSetting from '@/composables/use-table-setting';
import { useMainStore } from '@/stores/main';
import { useNodeManageStore } from '@/stores/node-manage';
import { useRouteSubTitle } from '@/stores/route-sub-title';

interface FilterOption {
  list: { text: string; value: string }[];
  checked: string[];
  filterScope: string;
}
type taskType =
  | 'install_agent'
  | 'install_plugin'
  | 'ensure_plugin_v2'
  | 'stop_plugin_v2'
  | 'upgrade_agent'
  | 'upgrade_plugin'
  | 'uninstall_plugin_v2'
  | 'migrate_plugin_v2'
  | 'assign_proxy_unit';
type filterProp = 'type' | 'operator' | 'status';
type HistoryTab = 'agent' | 'proxy' | 'plugin';

const historyTabs: HistoryTab[] = ['agent', 'proxy', 'plugin'];
const normalizeHistoryTab = (tab: unknown): HistoryTab => (
  historyTabs.includes(tab as HistoryTab) ? tab as HistoryTab : 'agent'
);

const historySubtitleMap: Record<HistoryTab, string> = {
  agent: 'platform.nodeMan.taskHistory.subtitle.agent',
  proxy: 'platform.nodeMan.taskHistory.subtitle.proxy',
  plugin: 'platform.nodeMan.taskHistory.subtitle.plugin',
};

const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const mainStore = useMainStore();
const nodeManageStore = useNodeManageStore();
const routeSubTitle = useRouteSubTitle();
const tableData = ref<NodeWorkflowInfo[]>([]);
const maxHeight = computed(() => mainStore.windowInnerHeight - 264 - (mainStore.noticeShow ? 40 : 0));

// URL query 与筛选状态同步

/** 将当前筛选状态序列化为 URL query */
const serializeFiltersToQuery = (): LocationQuery => {
  const query: LocationQuery = { active: active.value };

  // searchSelectValue → 每个条件序列化为 query[key] = JSON 字符串
  searchSelectValue.value.forEach((item) => {
    if (item.values?.length) {
      query[`f_${item.id}`] = JSON.stringify(item.values.map((v: any) => ({ id: v.id, name: v.name })));
    }
  });

  // dateValue
  if (dateValue.value?.length === 2) {
    query.date_start = String(Math.floor(new Date(dateValue.value[0]).getTime() / 1000));
    query.date_end = String(Math.floor(new Date(dateValue.value[1]).getTime() / 1000));
  }

  // pagination
  query.page = String(pagination.current);
  query.limit = String(pagination.limit);

  // hideAutoTask
  if (hideAutoTask.value) {
    query.hide_auto = '1';
  }

  return query;
};

/** 从 URL query 恢复筛选状态 */
const restoreFiltersFromQuery = (query: LocationQuery) => {
  isRestoringFromUrl.value = true;

  // active tab
  active.value = normalizeHistoryTab(query.active);

  // searchSelectValue
  const filterItems: { id: string; name: string; values: any }[] = [];
  Object.keys(query).forEach((key) => {
    if (key.startsWith('f_')) {
      const filterId = key.slice(2);
      try {
        const values = JSON.parse(query[key] as string);
        if (Array.isArray(values) && values.length > 0) {
          const searchItem = searchSelectData.value.find(d => d.id === filterId);
          filterItems.push({
            id: filterId,
            name: searchItem?.name || filterId,
            values,
          });
        }
      } catch { /* ignore invalid JSON */ }
    }
  });
  if (filterItems.length > 0) {
    searchSelectValue.value = filterItems;
    // 同步 filterOptionSource
    Object.keys(filterOptionSource).forEach((key) => {
      filterOptionSource[key].checked = [];
    });
    filterItems.forEach((item) => {
      if (filterOptionSource[item.id as filterProp]) {
        filterOptionSource[item.id as filterProp].checked = item.values.map((v: any) => v.id) as string[];
      }
    });
  }

  // dateValue
  if (query.date_start && query.date_end) {
    const start = Number(query.date_start) * 1000;
    const end = Number(query.date_end) * 1000;
    if (start > 0 && end > 0) {
      dateValue.value = [start, end];
    }
  }

  // pagination
  if (query.page) {
    pagination.current = Number(query.page) || 1;
  }
  if (query.limit) {
    pagination.limit = Number(query.limit) || 50;
  }

  // hideAutoTask
  hideAutoTask.value = query.hide_auto === '1';

  // 下一 tick 解除标记
  nextTick(() => {
    isRestoringFromUrl.value = false;
  });
};

// tab
const active = ref('agent');
const updateRouteSubTitle = () => {
  routeSubTitle.subTitle = historySubtitleMap[normalizeHistoryTab(active.value)];
};
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
  assign_proxy_unit: {
    text: t('platform.nodeMan.taskHistory.taskType.assign_proxy_unit'),
  },
  install_plugin: {
    text: t('platform.nodeMan.taskHistory.taskType.install_plugin'),
  },
  ensure_plugin_v2: {
    text: t('platform.nodeMan.taskHistory.taskType.ensure_plugin_v2'),
  },
  stop_plugin_v2: {
    text: t('platform.nodeMan.taskHistory.taskType.stop_plugin_v2'),
  },
  upgrade_plugin: {
    text: t('platform.nodeMan.taskHistory.taskType.upgrade_plugin'),
  },
  uninstall_plugin_v2: {
    text: t('platform.nodeMan.taskHistory.taskType.uninstall_plugin_v2'),
  },
  migrate_plugin_v2: {
    text: t('platform.nodeMan.taskHistory.taskType.migrate_plugin_v2'),
  },
  uninstall_plugin: {
    text: t('platform.nodeMan.taskHistory.taskType.uninstall_plugin'),
  },
  reconfig_plugin: {
    text: t('platform.nodeMan.taskHistory.taskType.reconfig_plugin'),
  },
  apply_plugin_subconfig: {
    text: t('platform.nodeMan.taskHistory.taskType.apply_plugin_subconfig'),
  },
  restart_plugin: {
    text: t('platform.nodeMan.taskHistory.taskType.restart_plugin'),
  },
  stop_plugin: {
    text: t('platform.nodeMan.taskHistory.taskType.stop_plugin'),
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
      name: keyMap?.[value]?.text || String(value),
    }));
}
const IPV4_REG = /^((25[0-5]|2[0-4]\d|[01]?\d\d?)\.){3}(25[0-5]|2[0-4]\d|[01]?\d\d?)$/;
const IPV6_REG = /^(?:[A-F0-9]{1,4}:){7}[A-F0-9]{1,4}$/i;
const AREA_IP_REG = /^(\d+):(.+)$/; // 管控区域ID:IP 格式

const searchSelectData = computed(() => [
  { id: 'workflow_id', name: t('platform.nodeMan.taskHistory.label.taskID') },
  { id: 'ip', name: 'IP', multiple: true },
  {
    id: 'area_ip',
    name: `${t('platform.nodeMan.bk_cloud_name')}ID:IP`,
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
/**
 * 解析多分隔符输入，支持空格、换行、分号、逗号
 */
const parseMultiDelimiterInput = (text: string): string[] => text
  .split(/[\s\n;,|、]+/)
  .map(item => item.trim())
  .filter(item => item.length > 0);

/**
 * 智能识别输入类型
 */
const detectInputType = (text: string): { type: 'ip' | 'area_ip' | null; value: string } => {
  if (AREA_IP_REG.test(text)) return { type: 'area_ip', value: text };
  if (IPV4_REG.test(text)) return { type: 'ip', value: text };
  if (IPV6_REG.test(text)) return { type: 'ip', value: text };
  return { type: null, value: text };
};

/**
 * 拦截原生 paste 事件，将空格分隔符转换为组件能识别的逗号
 */
const handleNativePaste = (event: ClipboardEvent) => {
  const text = event.clipboardData?.getData('text');
  if (!text) return;
  if (text.includes(' ') && !/[|,、\r\n\n]/.test(text)) {
    event.preventDefault();
    const normalizedText = text.replace(/\s+/g, ',');
    const target = event.target as HTMLElement;
    if (target && target.isContentEditable) {
      document.execCommand('insertText', false, normalizedText);
    }
  }
};

/**
 * 处理粘贴/快速输入的逻辑
 */
const handleInputPaste = (data: { id: string; name: string; values: { id: string; name: string }[] }[]) => {
  if (data.length === 0) return;

  const lastItem = data[data.length - 1];

  if (['ip', 'area_ip'].includes(lastItem.id) && lastItem.values.length > 0) {
    const allParsedItems: string[] = [];
    lastItem.values.forEach((value: any) => {
      const parsedItems = parseMultiDelimiterInput(value.id);
      allParsedItems.push(...parsedItems);
    });
    const uniqueItems = Array.from(new Set(allParsedItems));
    if (uniqueItems.length > 0) {
      lastItem.values = uniqueItems.map(item => ({ id: item, name: item }));
      return;
    }
  }

  const searchFieldIds = new Set(searchSelectData.value.map(item => item.id));
  const tailRawInputIds: string[] = [];
  for (let i = data.length - 1; i >= 0; i--) {
    const current = data[i];
    if (searchFieldIds.has(current.id) || current.values?.length) break;
    tailRawInputIds.unshift(current.id);
  }

  const parsedItems = (tailRawInputIds.length > 0
    ? tailRawInputIds
    : [lastItem.id])
    .flatMap(text => parseMultiDelimiterInput(text));
  const uniqueParsedItems = Array.from(new Set(parsedItems));
  if (uniqueParsedItems.length === 0) return;

  const firstDetection = detectInputType(uniqueParsedItems[0]);
  if (!firstDetection.type) return;

  const allSameType = uniqueParsedItems.every(item => detectInputType(item).type === firstDetection.type);
  if (!allSameType) return;

  let targetId = '';
  let targetName = '';

  switch (firstDetection.type) {
    case 'ip':
      targetId = 'ip';
      targetName = 'IP';
      break;
    case 'area_ip':
      targetId = 'area_ip';
      targetName = `${t('platform.nodeMan.bk_cloud_name')}ID:IP`;
      break;
  }

  if (targetId) {
    searchSelectValue.value = searchSelectValue.value.filter((item: any) => {
      if (item.id === targetId) return false;
      return !tailRawInputIds.includes(item.id);
    });
    searchSelectValue.value.push({
      id: targetId,
      name: targetName,
      values: uniqueParsedItems.map(item => ({ id: item, name: item })),
    });
  }
};

const handleSearchSelectChange = async (data: { id: string; name: string; values: { id: string; name: string }[] }[]) => {
  handleInputPaste(data);
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
const proxyType = ['install_proxy', 'upgrade_proxy', 'reconfig_proxy', 'restart_proxy', 'uninstall_proxy', 'assign_proxy_unit'];
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
      // 按 Tab 维度过滤节点角色：agent Tab 只看 agent 任务，proxy Tab 只看 proxy 任务
      // plugin Tab 走 PluginWorkflow 系列接口，不受此字段影响
      ...(isNode.value ? { node_role: [active.value] } : {}),
    },
    operate_time_range: {
      start_timestamp_sec: getTimestampInSeconds(dateValue.value[0]),
      end_timestamp_sec: getTimestampInSeconds(dateValue.value[1]),
    },
  };
  searchSelectValue.value.forEach((item: any) => {
    // IP 字段：自动识别 IPv4/IPv6 并分类
    if (item.id === 'ip' && item.values?.length) {
      const ipv4List: string[] = [];
      const ipv6List: string[] = [];
      item.values.forEach((value: any) => {
        if (IPV4_REG.test(value.id)) {
          ipv4List.push(value.id);
        } else if (IPV6_REG.test(value.id)) {
          ipv6List.push(value.id);
        }
      });
      if (ipv4List.length > 0) {
        params.exact_include_conditions.bk_host_innerip = ipv4List;
      }
      if (ipv6List.length > 0) {
        params.exact_include_conditions.bk_host_innerip_v6 = ipv6List;
      }
      return;
    }

    // 管控区域ID:IP：自动拆分为 bk_networkarea_id + IP 列表
    if (item.id === 'area_ip' && item.values?.length) {
      const areaIds = new Set<number>();
      const ipv4List: string[] = [];
      const ipv6List: string[] = [];
      item.values.forEach((value: any) => {
        const match = AREA_IP_REG.exec(value.id);
        if (match) {
          const areaId = Number(match[1]);
          const ip = match[2];
          areaIds.add(areaId);
          if (IPV4_REG.test(ip)) {
            ipv4List.push(ip);
          } else if (IPV6_REG.test(ip)) {
            ipv6List.push(ip);
          }
        }
      });
      if (areaIds.size > 0) {
        params.exact_include_conditions.bk_networkarea_id = Array.from(areaIds);
      }
      if (ipv4List.length > 0) {
        params.exact_include_conditions.bk_host_innerip = ipv4List;
      }
      if (ipv6List.length > 0) {
        params.exact_include_conditions.bk_host_innerip_v6 = ipv6List;
      }
      return;
    }

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
      ...(isNode.value ? { node_role: [active.value] } : {}),
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
            if (key === 'type') text = typeMap.value[value as string]?.text || value;
            if (key === 'status') text = statusMap.value[value as string]?.text || value;
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
        item.finish_time > 0 ? item.finish_time - item.operate_time : Date.now() - item.operate_time,
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
  // 保存当前筛选状态到 store，返回时恢复
  nodeManageStore.saveHistoryFilters({
    active: active.value,
    searchSelectValue: searchSelectValue.value.map(item => ({
      id: item.id,
      name: item.name,
      values: item.values?.map((v: any) => ({ id: v.id, name: v.name })) || [],
    })),
    dateStart: dateValue.value?.length === 2
      ? Math.floor(new Date(dateValue.value[0]).getTime() / 1000) : 0,
    dateEnd: dateValue.value?.length === 2
      ? Math.floor(new Date(dateValue.value[1]).getTime() / 1000) : 0,
    page: pagination.current,
    limit: pagination.limit,
    hideAutoTask: hideAutoTask.value,
  });
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
// 标记是否正在恢复筛选状态（首次加载或从详情返回时），此期间不写回 URL
const isRestoringFromUrl = ref(false);

// 首次加载时从 URL 恢复筛选状态
if (!route.query.active) {
  router.replace({ query: { ...route.query, active: 'agent' } });
} else {
  restoreFiltersFromQuery(route.query);
  // 首次加载数据
  getTaskList();
  getWorkflowDistinct();
}

watch([() => active.value, () => route.fullPath], updateRouteSubTitle, { immediate: true });

// 监听后续 URL 变化（如浏览器后退、菜单导航、Tab 切换 router.replace）
watch(
  () => route.query,
  () => {
    const newActive = normalizeHistoryTab(route.query.active);
    if (active.value !== newActive) {
      active.value = newActive;
    }
    debounceGetTaskList.cancel();
    getTaskList();
    getWorkflowDistinct();
  },
);

watch(
  [
    () => searchSelectValue,
    () => mainStore.selectedBusinessId,
  ],
  async () => {
    // 筛选条件变化时同步到 URL
    if (!isRestoringFromUrl.value && route.query.active) {
      const newQuery = serializeFiltersToQuery();
      const currentQueryStr = JSON.stringify(route.query);
      const newQueryStr = JSON.stringify(newQuery);
      if (currentQueryStr !== newQueryStr) {
        router.replace({ query: newQuery });
      }
    }
    await debounceGetTaskList();
  },
  { deep: true },
);

// 日期、隐藏自动任务、分页变化时同步到 URL
watch(
  [() => dateValue.value, () => hideAutoTask.value, () => pagination.current, () => pagination.limit],
  () => {
    if (!isRestoringFromUrl.value && route.query.active) {
      const newQuery = serializeFiltersToQuery();
      const currentQueryStr = JSON.stringify(route.query);
      const newQueryStr = JSON.stringify(newQuery);
      if (currentQueryStr !== newQueryStr) {
        router.replace({ query: newQuery });
      }
    }
  },
  { deep: true },
);

// 监听 Tab 切换，只负责更新 URL
watch(
  () => active.value,
  (newActive) => {
    if (isRestoringFromUrl.value) return;
    // 仅当 URL 不同时才更新，避免循环触发
    if (route.query.active !== newActive) {
      // Tab 切换时移除任务类型筛选（不同 tab 的 type 值不同）
      const idx = searchSelectValue.value.findIndex((item: any) => item.id === 'type');
      if (idx > -1) {
        searchSelectValue.value.splice(idx, 1);
        filterOptionSource.type.checked = [];
      }
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
    getTaskList();
    getWorkflowDistinct();
  },
  { deep: true },
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
