<template>
  <div>
    <bk-loading
      title="数据加载中"
      :loading="loading"
      class="w-full overflow-auto mt-[16px]"
    >
      <Table
        class="w-full filterTable"
        ref="tableRef"
        :data="list"
        :empty-text="$t('table.empty')"
        empty-cell-text="--"
        :pagination="pagination"
        :sort-config="sortConfig"
        :show-settings="isShowSetting"
        :settings="settings"
        :max-height="maxHeight"
        @setting-change="handleSettingChange"
        @column-filter="handleFilter">
        <template #prepend>
          <div v-if="hasSelection" class="flex items-center justify-center h-[30px] bg-[#ebecf0] text-[12px]">
            <template v-if="isCrossPageSelection">
              {{ $t('taskDetail.table.crossPageSelected') }}
              <span class="font-bold mx-1"> {{ pagination.count - excludedIds.size }} </span>
              <span>{{ $t('taskDetail.table.items') }}</span>
              <Button text theme="primary" @click="handleClearSelection">
                {{ $t('taskDetail.table.cancelSelection') }}
              </Button>
            </template>
            <template v-else>
              <span>{{ $t('taskDetail.table.selected') }}</span>
              <span class="font-bold mx-1"> {{ selection.length }} </span>
              <span>{{ $t('taskDetail.table.items') }}</span>
              <Button
                v-if="pagination.count > pagination.limit"
                text theme="primary"
                @click="handleSelectAllCrossPage">
                {{ $t('taskDetail.table.selectAllPages', { x: pagination.count }) }}
              </Button>
              <Button v-else text theme="primary" @click="handleClearSelection">
                {{ $t('taskDetail.table.cancelSelection') }}
              </Button>
            </template>
          </div>
        </template>
        <TableColumn width="80" fixed="left">
          <template #header>
            <Button text class="flex items-center justify-start">
              <Checkbox
                :model-value="isCurrentPageAllChecked"
                :indeterminate="isIndeterminate"
                @change="handleHeaderClick" />
              <Dropdown trigger="click" placement="bottom-start">
                <i class="nodeman-icon nc-arrow-down ml-1 text-[18px]"></i>
                <template #content>
                  <Dropdown.DropdownMenu>
                    <Dropdown.DropdownItem @click="handleSelectCurrentPage">
                      {{ $t('taskDetail.filter.currentPage') }}
                    </Dropdown.DropdownItem>
                    <Dropdown.DropdownItem @click="handleSelectAllCrossPage">
                      <Button text>{{ $t('taskDetail.filter.crossSelected') }}</Button>
                    </Dropdown.DropdownItem>
                  </Dropdown.DropdownMenu>
                </template>
              </Dropdown>
            </Button>
          </template>
          <template #default="{ row }">
            <Checkbox class="mt-[6px]" :model-value="row.checked" @change="(val) => handleRowCheck(val, row)" />
          </template>
        </TableColumn>
        <TableColumn
          field="bk_host_id"
          title="Host ID"
          fixed="left"
          :min-width="100"
        ></TableColumn>
        <TableColumn
          :label="$t('topoManager.workAreaDetail.table.ipv4')"
          field="bk_host_innerip"
          show-overflow="tooltip"
          fixed="left"
          :min-width="150">
        </TableColumn>
        <TableColumn
          :label="$t('topoManager.workAreaDetail.table.ipv6')"
          field="bk_host_innerip_v6"
          show-overflow="tooltip"
          :min-width="130">
        </TableColumn>
        <TableColumn
          :label="$t('installProxy.exportIP')"
          field="export_ip"
          show-overflow="tooltip"
          :min-width="130">
        </TableColumn>
        <TableColumn
          :label="$t('installProxy.serviceIP')"
          field="advertise_ip"
          show-overflow="tooltip"
          :min-width="130">
        </TableColumn>
        <TableColumn
          :label="$t('installProxy.loginIP')"
          field="login_ip"
          show-overflow="tooltip"
          :min-width="130">
        </TableColumn>
        <TableColumn
          :label="$t('installProxy.businessName')"
          field="bk_biz_id"
          show-overflow="tooltip"
          :min-width="100">
          <template #default="{ row }">
            {{ businessList.find(item => item.bk_biz_id === row.bk_biz_id)?.bk_biz_name }}
          </template>
        </TableColumn>
        <TableColumn
          field="bk_networkarea_id"
          :title="t('platform.nodeMan.bk_cloud_name')"
          :filter="filterOptionSource.bk_networkarea_id"
          :min-width="120"
          show-overflow
        >
          <template #default="{ row }">
            {{ row.bk_networkarea_name }}
          </template>
        </TableColumn>
        <TableColumn
          show-overflow
          field="bk_networkunit_id"
          :title="t('platform.nodeMan.bk_cloud_unit')"
          :filter="filterOptionSource.bk_networkunit_id"
          :min-width="120"
        >
          <template #default="{ row }">
            {{ networkUnitListMap.get(row.bk_networkunit_id) || row.bk_networkunit_name }}
          </template>
        </TableColumn>
        <TableColumn
          field="dept_name"
          :title="t('platform.nodeMan.dept_name')"
          :filter="filterOptionSource.dept_name"
          :min-width="120"
        ></TableColumn>
        <TableColumn
          label="Agent ID"
          field="bk_agent_id"
          show-overflow="tooltip"
          :min-width="300">
        </TableColumn>
        <TableColumn
          :label="$t('topoManager.workAreaDetail.table.proxyVersion')"
          field="node_version"
          show-overflow="tooltip"
          :filter="filterOptionSource.node_version"
          :min-width="150">
        </TableColumn>
        <TableColumn
          :label="$t('topoManager.workAreaDetail.table.proxyStatus')"
          field="node_status"
          show-overflow="tooltip"
          :filter="filterOptionSource.node_status"
          :min-width="130">
          <template #default="{ row }">
            <div class="flex items-center" v-if="row.node_status">
              <i
                :class="`nodeman-icon nc-${row.node_status.toLowerCase()} status-icon`"
              ></i>
              <div>{{ statusMap.get(row.node_status) || row.node_status }}</div>
            </div>
            <div class="flex items-center" v-else>
              <span>--</span>
            </div>
          </template>
        </TableColumn>
        <TableColumn
          :label="$t('installProxy.proxyTags')"
          field="proxy_tags"
          show-overflow="tooltip"
          :min-width="300">
          <template #default="{ row }">
            <div class="flex items-center gap-[4px]">
              <div v-for="tag in getOrderedProxyTags(row.proxy_tags)" :key="tag">
                <Popover
                  theme="light"
                  trigger="hover"
                  placement="top"
                  :arrow="true"
                  :max-width="280"
                  :offset="8"
                  :popover-delay="[0, 100]"
                  :component-event-delay="0"
                >
                  <Tag>{{ proxyTagMap[tag] }}</Tag>
                  <template #content>
                    <div class="text-[12px] leading-[20px]">
                      <p>{{ proxyTagListTooltipMap[tag] }}</p>
                    </div>
                  </template>
                </Popover>
              </div>
            </div>
          </template>
        </TableColumn>
        <TableColumn
          :title="$t('installProxy.pluginNum')"
          field="pluginNum"
          :min-width="122"
        >
          <template #default="{ row }">
            <Button text theme="primary" @click="openSidebar(row)">
              {{ row.pluginNum || 0 }}
              <i class="nodeman-icon nc-plug-in ml-[5px]"></i>
            </Button>
          </template>
        </TableColumn>
        <TableColumn
          :label="$t('topoManager.workAreaDetail.table.action')"
          field="action"
          fixed="right"
          show-overflow="tooltip"
          :min-width="140">
          <template #default="{ row }">
            <div class="flex">
              <Button theme="primary" text class="mr-[12px]" @click="handleEdit(row)">
                {{ $t('topoManager.workAreaDetail.table.edit') }}
              </Button>
              <MoreAction
                :ipv4="row.bk_host_innerip"
                :data="[row]"
                @reinstall="handleReinstall(row)">
                <i class="nodeman-icon nc-more cursor"></i>
              </MoreAction>
            </div>
          </template>
        </TableColumn>
      </Table>
    </bk-loading>
    <edit-proxy-sideslider
      v-model:is-show="sidesliderData.isShow"
      :data="sidesliderData.data"
      @update="handleUpdate"
    >
    </edit-proxy-sideslider>
    <ReinstallProxy
      v-model:is-show="isShowInstallProxy"
      :data="reinstallData"
      :bk_networkunit_id="bkNetworkunitId"
    />
    <!-- 侧边栏 -->
    <processSideslider
      v-model:is-show="isShowSideslider"
      type="node"
      :node="currentProxy"
    ></processSideslider>
  </div>
</template>

<script lang="ts" setup>
import { Button, Checkbox, Dropdown, Popover, Tag } from 'bkui-vue';
import { debounce } from 'lodash';
import { computed, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute } from 'vue-router';

import { Table, TableColumn } from '@blueking/table';

import processSideslider from '../../../node/plugin/process-sideslider.vue';
import ReinstallProxy from '../../install-proxy/reinstall-proxy.vue';

import EditProxySideslider from './edit-proxy-sideslider.vue';
import MoreAction from './more-action.vue';

import type { TopoHostDistinctRespData } from '@/@types/topo';
import type {
  TopoHostExactConditions,
  TopoHostFuzzyConditions,
} from '@/@types/topo.d';
import { ProcessAPIService } from '@/api/modules/process';
import { TopoService } from '@/api/modules/topo';
import useDynamicsHeight from '@/composables/use-table-height';
import useTableSetting from '@/composables/use-table-setting';
import { useMainStore } from '@/stores/main';

interface FilterOption {
  list: { text: string; value: string }[];
  checked: string[];
  filterScope: string;
}

const props = defineProps({
  searchSelectValue: {
    type: Array,
    default: [],
  },
  bkNetworkunitId: {
    type: Number,
    default: 0,
  },
});
const emit = defineEmits(['update:searchSelectValue', 'selectChange', 'getData', 'excludedIdsChange', 'updateCrossPage', 'updateSearchSelectData']);

const { t } = useI18n();
const route = useRoute();
const workAreaId = Number(route.params.workarea);
const list = ref<Host[]>([]);
const pagination = reactive({ count: 0, limit: 20, current: 1 });
const sortConfig = ref({ multiple: true });
const mainStore = useMainStore();
const businessList = computed(() => mainStore.businessList);
const sidesliderData = reactive<{
  isShow: boolean,
  data: Host | null
}>({
  isShow: false,
  data: null,
});
const proxyTagMap = ref({
  dedicated_installer: t('installProxy.installJump'),
  cluster_tunnel: t('installProxy.agentControl'),
  file_tunnel: t('installProxy.fileTransfer'),
  data_tunnel: t('installProxy.dataReport'),
});
const proxyTagListTooltipMap = ref({
  dedicated_installer: t('installProxy.installJumpListTooltip'),
  cluster_tunnel: t('installProxy.agentControlListTooltip'),
  file_tunnel: t('installProxy.fileTransferListTooltip'),
  data_tunnel: t('installProxy.dataReportListTooltip'),
});
const proxyTagDisplayOrder = ['dedicated_installer', 'cluster_tunnel', 'file_tunnel', 'data_tunnel'];
const proxyTagOrderMap = new Map(proxyTagDisplayOrder.map((tag, index) => [tag, index]));
const getOrderedProxyTags = (tags: string[] = []) => [...tags].sort((a, b) => {
  const aIndex = proxyTagOrderMap.get(a) ?? Number.MAX_SAFE_INTEGER;
  const bIndex = proxyTagOrderMap.get(b) ?? Number.MAX_SAFE_INTEGER;
  if (aIndex === bIndex) {
    return a.localeCompare(b);
  }
  return aIndex - bIndex;
});
const { isShowSetting, settings, handleSettingChange } = useTableSetting({
  checked: [
    'bk_host_innerip',
    'bk_host_innerip_v6',
    'export_ip',
    'advertise_ip',
    'login_ip',
    'bk_biz_id',
    'bk_networkarea_id',
    'bk_networkunit_id',
    'dept_name',
    'bk_agent_id',
    'node_version',
    'node_status',
    'proxy_tags',
    'pluginNum',
    'action',
  ],
  disabled: ['action'],
}, `topoMng-workarea-detail-${String(route.name)}`);
// 表格勾选
const selection = computed(() => list.value.filter((item: any) => item.checked));
const setRowCheckedByHostId = (hostId: number, checked: boolean) => {
  const target = list.value.find(item => item.bk_host_id === hostId);
  if (target) {
    target.checked = checked;
  }
};
const handleSelectChange = ({
  checked,
  row,
}: {
  checked: boolean;
  row: any;
}) => {
  setRowCheckedByHostId(row.bk_host_id, checked);
  emit('selectChange', selection.value);
};

// 表格全选
const handleSelectAllChange = ({ checked }: { checked: boolean }) => {
  list.value.forEach((item: any) => (item.checked = checked));
  emit('selectChange', selection.value);
};

// --- 跨页全选核心状态 ---
const isCrossPageSelection = ref(false); // 是否开启跨页全选模式
const excludedIds = ref<Set<number>>(new Set()); // 全选模式下，用户手动"取消勾选"的 ID 集合
// 计算属性：是否有任何选中（用于禁用批量按钮）
const hasSelection = computed(() => list.value.some(item => item.checked) || isCrossPageSelection.value);

// 计算属性：当前页是否全选（用于表头 Checkbox 状态）
// eslint-disable-next-line max-len
const isCurrentPageAllChecked = computed(() => list.value.length > 0 && list.value.every(item => item.checked));
const isIndeterminate = computed(() => {
  const selectedCount = list.value.filter(item => item.checked).length;
  return selectedCount > 0 && selectedCount < list.value.length;
});

// 1. 处理单行勾选
const handleRowCheck = (checked: boolean, row: any) => {
  setRowCheckedByHostId(row.bk_host_id, checked);
  if (isCrossPageSelection.value) {
    if (!checked) {
      excludedIds.value.add(row.bk_host_id);
    } else {
      excludedIds.value.delete(row.bk_host_id);
    }
  }
};

// 2. 跨页全选
const handleSelectAllCrossPage = async () => {
  isCrossPageSelection.value = true;
  excludedIds.value.clear();

  // 更新当前页的选中状态
  list.value.forEach(item => (item.checked = true));
};

// 3. 取消选择
const handleClearSelection = () => {
  isCrossPageSelection.value = false;
  excludedIds.value.clear();
  list.value.forEach(item => (item.checked = false));
};

// 4. 本页全选
const handleSelectCurrentPage = () => {
  isCrossPageSelection.value = false;
  list.value.forEach(item => (item.checked = true));
};

// 5. 表头 Checkbox 快速切换
const handleHeaderClick = () => {
  isCurrentPageAllChecked.value ? handleClearSelection() : handleSelectCurrentPage();
};

const tableRef = ref();
const loading = ref(false);
// 待优化 各影响table最大高度的元素的高度
const tableOffset = 445;
const { maxHeight } = useDynamicsHeight(tableOffset);

const handleFilter = ({ checked, field }: { checked: string[]; field: string }) => {
  // 1. 克隆一份数据，避免直接修改 props
  const newValue = [...props.searchSelectValue];

  const index = newValue.findIndex((item: any) => item.id === field);
  if (index > -1) newValue.splice(index, 1);

  if (checked.length) {
    newValue.push({
      id: field,
      name: field,
      values: checked.map((item: any) => {
        let name = item;
        if (field === 'bk_networkarea_id') name = networkAreaListMap.value.get(Number(item)) || item;
        if (field === 'bk_networkunit_id') name = networkUnitListMap.value.get(Number(item)) || item;
        if (field === 'node_status') name = statusMap.value.get(item) || item;
        return { id: item, name };
      }),
    });
  }

  // 2. 通过 emit 通知父组件更新
  emit('update:searchSelectValue', newValue);
};

const statusMap = ref(new Map<string, string>([
  ['init', t('platform.nodeMan.agentStatus.init')],
  ['running', t('platform.nodeMan.agentStatus.running')],
  ['damaged', t('platform.nodeMan.agentStatus.damaged')],
  ['unknown', t('platform.nodeMan.agentStatus.unknown')],
]));

const searchSelectValue = computed(() => props.searchSelectValue);
const filterOptionSource: Record<string, FilterOption> = reactive({
  bk_networkarea_id: { list: [], checked: [], filterScope: 'all' },
  bk_networkunit_id: { list: [], checked: [], filterScope: 'all' },
  dept_name: { list: [], checked: [], filterScope: 'all' },
  node_version: { list: [], checked: [], filterScope: 'all' },
  node_status: { list: [], checked: [], filterScope: 'all' },
});

// ---------- 侧边栏 ----------
const isShowSideslider = ref(false);
const currentProxy = ref();
// 打开侧边栏并加载进程列表
const openSidebar = async (proxy) => {
  isShowSideslider.value = true;
  currentProxy.value = proxy;
};

const handleEdit = (row: Host) => {
  sidesliderData.isShow = true;
  sidesliderData.data = row;
};
const fuzzyKeys = new Set([
  'bk_host_innerip',
  'bk_host_innerip_v6',
  'dept_name',
]);
const getParams = () => {
  const params = {
    page: {
      limit: pagination.limit,
      offset: (pagination.current - 1) * pagination.limit,
    },
    exact_include_conditions: {
      node_role: ['proxy'],
    } as TopoHostExactConditions,
    fuzzy_include_conditions: {} as TopoHostFuzzyConditions,
  };
  if (route.name === 'proxy') {
    params.exact_include_conditions.bk_biz_id = mainStore.selectedBusinessId;
  }
  if (props.bkNetworkunitId) {
    params.exact_include_conditions.bk_networkunit_id = [props.bkNetworkunitId];
  }
  searchSelectValue.value.forEach((item: any) => {
    const target = fuzzyKeys.has(item.id)
      ? params.fuzzy_include_conditions
      : params.exact_include_conditions;
    target[item.id] = item.values?.map((value: any) => value.id);
  });
  return params;
};
const getProxyList = async () => {
  loading.value = true;
  const res = await TopoService.HostList(getParams()).catch((err) => {
    console.log(err);
    return {
      total: 0,
      items: [],
    };
  });
  pagination.count = res.total;

  const pluginNumMap = await ProcessAPIService.GetProcessDistributionByHostID({
    exact_include_conditions: {
      bk_host_id: res.items.map((item: any) => item.bk_host_id),
    },
  }).catch((err: any) => {
    console.error('获取插件数量失败:', err);
    return {} as Record<number, number>;
  });

  list.value = res.items.map((item: any) => {
    let isChecked = false;
    if (isCrossPageSelection.value) {
      // 如果是跨页模式，只要不在排除名单里就是选中
      isChecked = !excludedIds.value.has(item.bk_host_id);
    }
    return {
      ...item.state,
      ...item.info,
      ...item,
      bk_host_innerip: item.info.bk_host_innerip_list?.join(',') || '',
      bk_host_innerip_v6: item.info.bk_host_innerip_v6_list?.join(',') || '',
      pluginNum: pluginNumMap[item.bk_host_id] || 0,
      checked: isChecked,
    };
  });
  emit('getData', list.value);
  loading.value = false;
};
const handleUpdate = async () => {
  await getProxyList();
};
const isShowInstallProxy = ref(false);
const reinstallData = ref<Host[]>([]);
const handleReinstall = (row: Host) => {
  isShowInstallProxy.value = true;
  reinstallData.value = [row];
};

/**
 * 获取主机筛选条件的唯一值
 */
const networkAreaListMap = ref(new Map<number, string>([]));
const networkUnitListMap = ref(new Map<number, string>([]));
const hostDistinct = ref<TopoHostDistinctRespData | null>();
function getUniqueChildrenFrom <K extends keyof TopoHostDistinctRespData>(
  prop: K,
  keyMap?: Map<number | string, string | number>,
) {
  const uniqueValues = hostDistinct.value?.[prop] || [];
  return uniqueValues
    .filter((item: any) => item !== '')
    .map((value: any) => ({
      id: value,
      name: keyMap?.get(value) || String(value),
    }));
}
const getNetworkAreaList = async (data: {bk_networkarea_id: number[]} | null) => {
  const res = await TopoService.NetworkAreaList({
    page: { limit: 0 },
    exact_include_conditions: { bk_networkarea_id: data?.bk_networkarea_id || [] },
  }).catch((err: any) => {
    console.error('获取管控区域列表失败:', err);
    return { total: 0, items: [] };
  });
  networkAreaListMap.value.set(-1, t('platform.nodeMan.agentStatus.unassigned'));
  res.items.forEach((item) => {
    networkAreaListMap.value.set(item.bk_networkarea_id, item.bk_networkarea_name);
  });
};

const getNetworkUnitList = async (data: {bk_networkunit_id: number[]} | null) => {
  const res = await TopoService.NetworkUnitListBrief({
    exact_include_conditions: { bk_networkunit_id: data?.bk_networkunit_id || [] },
  }).catch((err: any) => {
    console.error('获取管控单元列表失败:', err);
    return { total: 0, items: [] };
  });
  networkUnitListMap.value.set(-1, t('platform.nodeMan.agentStatus.unassigned'));
  res.items.forEach((item) => {
    networkUnitListMap.value.set(item.bk_networkunit_id, item.bk_networkunit_name);
  });
};

const getHostDistinct = async () => {
  const params = {
    exact_include_conditions: {
      node_role: ['proxy'],
      bk_biz_id: mainStore.selectedBusinessId,
    },
  };
  const res = await TopoService.HostDistinct(params).catch((err: any) => {
    console.error('获取主机筛选条件唯一值失败:', err);
    return null;
  });
  await Promise.all([
    getNetworkAreaList(res),
    getNetworkUnitList(res),
  ]);
  if (res) {
    hostDistinct.value = res;
    Object.keys(res).forEach((key: any) => {
      if (filterOptionSource[key]) {
        filterOptionSource[key].list = res[key]
          .filter((item: any) => item !== '')
          .map((value: string | number) => {
            let text = value;
            if (key === 'bk_networkarea_id') text = networkAreaListMap.value.get(Number(value)) || value;
            if (key === 'bk_networkunit_id') text = networkUnitListMap.value.get(Number(value)) || value;
            if (key === 'node_status') text = statusMap.value.get(value as string) || value;
            return { text, value };
          });
      }
    });

    // 更新searchSelectData
    const searchSelectData = [
      {
        name: t('topoManager.workAreaDetail.table.ipv4'),
        id: 'bk_host_innerip',
      },
      {
        name: t('topoManager.workAreaDetail.table.ipv6'),
        id: 'bk_host_innerip_v6',
      },
      {
        name: 'AgentID',
        id: 'bk_agent_id',
      },
      {
        name: t('installProxy.proxyVersion'),
        id: 'node_version',
        children: getUniqueChildrenFrom('node_version'),
        multiple: true,
      },
      {
        name: t('topoManager.workAreaDetail.table.proxyStatus'),
        id: 'node_status',
        children: getUniqueChildrenFrom('node_status', statusMap.value),
        multiple: true,
      },
      {
        id: 'dept_name',
        name: t('platform.nodeMan.dept_name'),
        children: getUniqueChildrenFrom('dept_name'),
        multiple: true,
      },
    ];
    emit('updateSearchSelectData', searchSelectData);
  };
};
watch(route, async () => {
  if (route.query.os_type) {
    // 1. 拷贝当前已有数据，避免直接修改 props
    const nextSearchValue = [...props.searchSelectValue];

    // 2. 构造从路由获取的新标签
    const routeTags = [
      {
        id: 'os_type',
        name: t('topoManager.installProxy.form.os'),
        values: [{ id: route.query.os_type, name: route.query.os_type }],
      },
      {
        id: 'cpu_arch',
        name: t('agentStrategy.table.arch'),
        values: [{ id: route.query.cpu_arch, name: route.query.cpu_arch }],
      },
      {
        id: 'node_version',
        name: t('installProxy.proxyVersion'),
        values: [{ id: route.query.node_version, name: route.query.node_version }],
      },
    ];

    // 3. 过滤掉重复的标签（防止重复 push）
    routeTags.forEach((tag) => {
      const isExist = nextSearchValue.some(item => item.id === tag.id);
      if (!isExist) {
        nextSearchValue.push(tag);
      }
    });

    // 4. 通知父组件更新 searchKey
    emit('update:searchSelectValue', nextSearchValue);
  }
}, { immediate: true, deep: true });
const debounceGetProxyList = debounce(() => {
  getProxyList();
}, 300);
watch(
  () => mainStore.selectedBusinessId,
  async () => {
    getHostDistinct();
  },
  { immediate: true, deep: true },
);
watch(
  [
    () => mainStore.selectedBusinessId,
    () => searchSelectValue.value,
  ],
  async () => {
    await debounceGetProxyList();
  },
  { immediate: true, deep: true },
);
watch(
  () => selection.value,
  () => {
    emit('selectChange', selection.value);
  },
  { deep: true },
);
watch(
  () => excludedIds.value,
  () => {
    emit('excludedIdsChange', excludedIds.value);
  },
  { deep: true },
);
watch(
  () => isCrossPageSelection.value,
  () => {
    emit('updateCrossPage', isCrossPageSelection.value);
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
</style>
