<template>
    <PageHeader
        class="w-full absolute top-0 z-100"
        :title="t('任务详情')"
        :back="true"
    >
        <span class="mx-[6px] text-[#979BA5] text-[14px]">-</span>
        <span class="text-[#979BA5] text-[14px] mr-[14px]" v-if="currentData">{{ currentData?.workflow_id }}</span>
        <Tag :theme="statusMap[currentStatus]?.tagTheme" type="filled" v-if="currentStatus">{{ statusMap[currentStatus]?.text || '' }}</Tag>
    </PageHeader>
    <div class="p-[24px] mt-[52px]">
        <div class="flex">
            <div v-for="item in taskInfoList" :key="item.name" class="leading-[30px] mr-[52px] text-[12px]">
                <div class="w-[50px]">{{ item.name }}</div>
                <div>{{ item.value }}</div>
            </div>
        </div>
        <div class="mt-[24px] mb-[14px] flex justify-between">
            <div class="flex gap-[12px]">
                <Dropdown
                    theme="light"
                    trigger="click"
                    :popover-options="{
                        clickContentAutoHide: true,
                    }">
                    <Button :disabled="!failedSelection.length">全部失败重试</Button>
                    <template #content>
                        <Dropdown.DropdownMenu extCls="dropDown-menu">
                            <Dropdown.DropdownItem
                                class="text-14px"
                                v-for="item in reTryType"
                                :key="item.id"
                                @click="handleFullRetry(item.id)">
                                {{ item.name }}
                            </Dropdown.DropdownItem>
                        </Dropdown.DropdownMenu>
                    </template>
                </Dropdown>
                <copy-ip-dropdown
                    type="agent"
                    :list="list"
                    :data="tableData"
                    filterProp="state"
                    :disabled="!selection.length"
                ></copy-ip-dropdown>
                <div class="h-[32px] bg-[#EAEBF0] rounded-[2px] flex items-center text-[12px] mr-[12px]">
                    <Radio.Group
                        v-model="radioGroupValue"
                        type="capsule"
                    >
                        <Radio.Button
                            v-for="item in radioGroup"
                            :label="item.name"
                            :key="item.name"
                        >
                            <i v-if="item.icon" :class="item.icon"></i>
                            <span>{{ item.label }} ({{ item.count }})</span>
                        </Radio.Button>
                    </Radio.Group>
                </div>
            </div>
            <SearchSelect
                class="flex-1"
                ref="searchSelect"
                :data="searchSelectData"
                v-model="searchSelectValue"
                :uniqueSelect="true"
                :placeholder="t('请输入IP、云区域、业务、目标版本、执行状态 搜索')"
                @update:modelValue="handleSearchSelectChange">
            </SearchSelect>
        </div>
        <bk-loading title="数据加载中" :loading="loading">
            <div class="relative">
                <Table
                    :data="filterTableData"
                    :empty-text="'暂无数据'"
                    :pagination="pagination"
                    show-overflow-tooltip
                    :max-height="maxHeight"
                    @checkbox-change="handleSelectChange"
                    @checkbox-all="handleSelectAllChange"
                    :show-settings="isShowSetting"
                    :settings="settings"
                    @setting-change="handleSettingChange"
                    @column-filter="handleFilter"
                >
                    <TableColumn type="checkbox" width="80" fixed="left"></TableColumn>
                    <TableColumn field="bk_host_inner" :title="t('IPv4')" width="150" fixed="left"></TableColumn>
                    <TableColumn field="bk_host_innerip_v6" :title="t('IPv6')" width="150"></TableColumn>
                    <TableColumn field="bk_networkarea_id" :title="t('云区域')" min-width="150"></TableColumn>
                    <TableColumn field="bk_biz_name" :title="t('业务')" min-width="150"></TableColumn>
                    <TableColumn field="node_version" :title="t('目标版本')" min-width="150" :filter="filterOptionSource.node_version"></TableColumn>
                    <TableColumn field="total_time_second" :title="t('耗时')">
                        <template #default={row}>
                            <span>{{ formatTimeToMS(row.total_time_second) }}</span>
                        </template>
                    </TableColumn>
                    <TableColumn
                        field="state"
                        :title="t('执行状态')"
                        :filter="filterOptionSource.state"
                        min-width="150"
                    >
                        <template #default="{ row }">
                            <div class="flex items-center" v-if="row.state && statusMap[row.state]">
                                <Spinner v-if="row.state === 'running'" class="mr-[8px]"/>
                                <template v-else>
                                    <i :class="`nodeman-icon nc-${statusMap[row.state].icon} status-icon`"></i>
                                </template>
                                <span>{{ statusMap[row.state].text }}</span>
                            </div>
                            <div class="flex items-center" v-else>
                                <span class="nodeman-icon nc-unknown status-icon"></span>
                                <span>{{ row.state }}</span>
                            </div>
                        </template>
                    </TableColumn>
                    <TableColumn
                        :title="t('操作')"
                        fixed="right"
                        width="150"
                    >
                        <template #default={row}>
                            <div class="flex items-center">
                                <Button text theme="primary" class="mr-[11px]" @click="handleViewLog(row)">查看日志</Button>
                                <Dropdown
                                    theme="light"
                                    trigger="click"
                                    ext-cls="dropdownCls"
                                    :popover-options="{
                                        clickContentAutoHide: true,
                                    }">
                                    <Button text theme="primary" v-if="!['success', 'running'].includes(row.state)">
                                        <right-turn-line fill="#3A84FF"/>
                                        <span>重试</span>
                                    </Button>
                                    <template #content>
                                        <Dropdown.DropdownMenu extCls="dropDown-menu">
                                            <Dropdown.DropdownItem
                                                class="text-14px"
                                                v-for="item in reTryType"
                                                :key="item.id"
                                                @click="handleRetry(row, item.id)">
                                                {{ item.name }}
                                            </Dropdown.DropdownItem>
                                        </Dropdown.DropdownMenu>
                                    </template>
                                </Dropdown>
                            </div>
                        </template>
                    </TableColumn>
                </Table>
                <Log :data="curRow" v-if="curRow" ref="logRef" @stop="handleStop"></Log>
            </div>
        </bk-loading>
    </div>
</template>
<script setup lang="ts">
interface FilterOption {
  list: { text: string, value: string }[];
  checked: string[];
  filterScope: string;
}
type taskType = 'install_agent' | 'install_plugin' | 'upgrade_agent' | 'upgrade_plugin';
type filterProp = 'state' | 'node_version';

import { RightTurnLine, Success, Close, AngleUpFill, Spinner } from 'bkui-vue/lib/icon';
import { ref, reactive, computed, onMounted, onBeforeUnmount, watch, nextTick } from 'vue';
import { Table, TableColumn } from '@blueking/table';
import { Button, Input, SearchSelect, Radio, Tag, ResizeLayout, Dropdown } from 'bkui-vue';
import usePage from '@/composables/use-page';
import { useMainStore } from '@/stores/main';
import { useI18n } from 'vue-i18n';
import { NodeWorkflowService } from '@/api/modules/node_workflow';
import { useRoute } from 'vue-router';
import { useNodeManageStore } from '@/stores/node-manage';
import useTableSetting from '@/composables/use-table-setting';
import Log from './log.vue';
import dayjs from 'dayjs';
import useInterval from "@/composables/use-interval";

const { t } = useI18n();
const route = useRoute();
const mainStore = useMainStore();
const nodeManageStore = useNodeManageStore();
const reTryType = [
    {
        id: 'full_node_instance_retry',
        name: t('全部重试')
    },
    {
        id: 'partial_node_instance_retry',
        name: t('部分重试')
    }
]
const maxHeight = computed(() => mainStore.windowInnerHeight - 214);
const currentData = computed(() => nodeManageStore.taskHistoryTableRowData);
const currentStatus = computed(() => nodeManageStore.currentStatus);
const statusMap = {
  running: {
    text: t('执行中'),
    tagTheme: 'info'
  },
  failed: {
    text: t('失败'),
    icon: 'terminated',
    tagTheme: 'danger'
  },
  success: {
    text: t('成功'),
    icon: 'running',
    tagTheme: 'success'
  },
  partial_failed: {
    text: t('部分失败'),
    icon: 'warning',
    tagTheme: 'warning'
  },
  ignored: {
    text: t('已忽略（没有需要变更的实例）'),
    icon: 'warning',
    tagTheme: ''
  },
  timeout: {
    text: t('超时'),
    icon: 'unknown',
    tagTheme: ''
  }
}
const typeMap = {
  install_agent: t('platform.nodeMan.taskHistory.taskType.install_agent'),
  install_plugin: t('platform.nodeMan.taskHistory.taskType.install_plugin'),
  upgrade_agent: t('platform.nodeMan.taskHistory.taskType.upgrade_agent'),
  upgrade_plugin: t('platform.nodeMan.taskHistory.taskType.upgrade_plugin'),
  reconfig_agent: t('platform.nodeMan.taskHistory.taskType.reconfig_agent'),
  restart_agent: t('platform.nodeMan.taskHistory.taskType.restart_agent'),
  uninstall_agent: t('platform.nodeMan.taskHistory.taskType.uninstall_agent'),
  install_proxy: t('platform.nodeMan.taskHistory.taskType.install_proxy'),
  upgrade_proxy: t('platform.nodeMan.taskHistory.taskType.upgrade_proxy'),
  reconfig_proxy: t('platform.nodeMan.taskHistory.taskType.reconfig_proxy'),
  restart_proxy: t('platform.nodeMan.taskHistory.taskType.restart_proxy'),
  uninstall_proxy: t('platform.nodeMan.taskHistory.taskType.uninstall_proxy'),
}
const formatTimeToMS = (duration: number = 0) => {
  const minutes = Math.floor(duration / 60000);
  const seconds = Math.floor((duration % 60000) / 1000);
  return `${minutes}m ${seconds}s`;
}
const timeFormatter = (val: number | string | undefined, format = 'YYYY-MM-DD HH:mm:ss') => {
  return val ? dayjs(val).format(format) : '--';
}

const sliceWorkflowId = (val: string) => {
    return '#' + val?.slice(-4);
}
const taskInfoList = computed(() => ([
    {prop: 'type', name: t('任务类型'), value: typeMap[nodeManageStore.taskHistoryTableRowData?.type as taskType]} || nodeManageStore.taskHistoryTableRowData?.type,
    {prop: 'cost_time', name: t('总耗时'), value: formatTimeToMS(nodeManageStore.taskHistoryTableRowData?.cost_time)},
    {prop: 'workflow_id', name: t('任务ID'), value: sliceWorkflowId(nodeManageStore.taskHistoryTableRowData?.workflow_id)},
    {prop: 'operator', name: t('执行人'), value: nodeManageStore.taskHistoryTableRowData?.operator},
    {prop: 'operate_time', name: t('执行时间'), value: timeFormatter(nodeManageStore.taskHistoryTableRowData?.operate_time)},
]));

const tableData = ref<any[]>([]);
const filterTableData = computed(() =>
    tableData.value.filter((item: any) =>
        radioGroupValue.value === 'all'
        || item.state === radioGroupValue.value
));
const radioGroupValue = ref('all');
const curOperationId = ref('');
const radioGroup = computed(() => ([ 
    {
        icon: '',
        label: t('全部'),
        name: 'all',
        count: tableData.value.length
    },
    {
        icon: 'nodeman-icon nc-running status-icon',
        label: t('成功'),
        name: 'success',
        count: tableData.value.filter((item: {state: string}) => item.state === 'success').length
    },
    {
        icon: 'nodeman-icon nc-terminated status-icon',
        label: t('失败'),
        name: 'failed',
        count: tableData.value.filter((item: {state: string}) => item.state === 'failed').length
    }
]));
const bussinessMap = computed(() => mainStore.businessList.map(item => ({
  id: item.bk_biz_id,
  name: item.bk_biz_name
})));
const getUniqueChildren = (prop: string) => {
  const uniqueValues = Array.from(new Set(tableData.value.map((item: any) => item[prop]).filter((item: any) => item)));
  return uniqueValues.map(value => ({
    id: value,
    name: prop === 'state' ? statusMap[value as string]?.text : String(value),
  }))
}
const filterOptionConfig = (prop: string, textMap?: Record<string, any>) => {
    const uniqueValues = Array.from(new Set(tableData.value.map((item: any) => item[prop])?.filter((item: any) => item)));
    return uniqueValues.map(value => ({
        text: textMap && textMap[value as string] ? textMap[value as string].text : value,
        value: value
    }));
}
const filterOptionSource = reactive<Record<string, FilterOption>>({
  node_version: {
    list: [],
    checked: [],
    filterScope: 'all'
  },
  state: {
    list: [],
    checked: [],
    filterScope: 'all'
  },
});
const loading = ref(false);
// 搜索
const searchSelectValue = ref<{id: string, name: string, values: any[]}[]>([]);
const searchSelectData = computed(() => [
  {id: 'bk_host_innerip', name: t('IP'), multiple: true},
  {id: 'bk_networkarea_id', name: t('云区域'), children: getUniqueChildren('bk_networkarea_id'), multiple: true},
  {id: 'bk_biz_id', name: t('业务'), children: bussinessMap.value, multiple: true},
  {id: 'node_version', name: t('目标版本'), children: getUniqueChildren('node_version'), multiple: true},
  {id: 'state', name: t('执行状态'), children: getUniqueChildren('state'), multiple: true},
]);
const handleSearchSelectChange = async (data: {id: string, name: string, values: {id: string,name: string}[]}[]) => {
    Object.keys(filterOptionSource).forEach(key => {
        filterOptionSource[key].checked = [];
    });
    data.forEach(item => {
        if (filterOptionSource[item.id as filterProp]) {
            filterOptionSource[item.id as filterProp].checked = item.values.map((item: any) => item.id) as string[];
        }
    });
}

// 分页
const {
  pagination
} = usePage(tableData);
// 复制
const list = [
    {
        id: 'all',
        name: '所有IP',
        children: [
            {
                id: 'ipv4',
                name: 'IPv4',
            },
            {
                id: 'ipv6',
                name: 'IPv6',
            },
        ],
    },
    {
        id: 'ignore',
        name: '被忽略IP',
        children: [
            {
                id: 'ipv4',
                name: 'IPv4',
            },
            {
                id: 'ipv6',
                name: 'IPv6',
            },
        ],
    },
    {
        id: 'failed',
        name: '失败IP',
        children: [
            {
                id: 'ipv4',
                name: 'IPv4',
            },
            {
                id: 'ipv6',
                name: 'IPv6',
            },
        ],
    },
    {
        id: 'success',
        name: '成功IP',
        children: [
            {
                id: 'ipv4',
                name: 'IPv4',
            },
            {
                id: 'ipv6',
                name: 'IPv6',
            },
        ],
    },
];
// 表格勾选
const selection = computed(() => tableData.value.filter((item: any) => item.checked));
const failedSelection = computed(() => selection.value.filter((item: any) => item.state === 'failed'));
const handleSelectChange = ({ checked, row }: {checked: boolean, row: any}) => {
  row.checked = checked;
}

// 表格全选
const handleSelectAllChange = ({ checked }: { checked: boolean}) => {
  tableData.value.forEach((item: any) => item.checked = checked);
}
// 表格设置
const { isShowSetting, settings, handleSettingChange } = useTableSetting({
  checked: [
    'bk_host_inner',
    'bk_host_innerip_v6',
    'bk_networkarea_id',
    'bk_biz_name',
    'node_version',
    'timeout_second',
    'state',
  ],
  disabled: [],
});

// 筛选
const handleFilter = ({checked, field}: {checked: string[], field: string}) => {
    const index = searchSelectValue.value.findIndex((item: any) => item.id === field);
    index > -1 && searchSelectValue.value.splice(index, 1);
    if (checked.length){
        searchSelectValue.value.push({ id: field, name: t(field), values: checked.map((item: any) => {
            const name = field === 'state' ? statusMap[item].text : item;
            return {
                id: item,
                name
            }
        })});
    }
}
// 重试
const handleRetry = async (row: any, type: string) => {
    const res = await NodeWorkflowService.NodeWorkflowOperationRetry({
        workflow_id: route.params.taskId,
        operation_id: [row.operation_id],
        retry_mod: type
    }).catch(() => false);
    if(res) {
        await getOperateList();
    }
}
const handleFullRetry = async (type: string) => {
    const res = await NodeWorkflowService.NodeWorkflowOperationRetry({
        workflow_id: route.params.taskId,
        operation_id: failedSelection.value.map(item => item.operation_id),
        retry_mod: type
    }).catch(() => false);
    if(res) {
        await getOperateList();
    }
}
const updataCurrentTaskInfo = async () => {
    const res = await NodeWorkflowService.NodeWorkflowList({
        exact_include_conditions: {}
    }).catch((err) => {
        console.log(err);
        return {
        total: 0,
        items: [],
        }
    });
    const list = res.items.map(item => {
        return {
        ...item,
        bk_biz_name: item.bk_biz_name.filter(item => item),
        cost_time: item.finish_time > 0 ? (item.finish_time - item.operate_time) : 0
        }
    });
    const findItem = list.find((item: any) => item.workflow_id === route.params.taskId);
    if (findItem) {
        nodeManageStore.updateCurrentRowData(findItem);
        nodeManageStore.updateCurrentStatus(findItem.status);
    }
}
const getParams = () => {
  const params = {
    page: {
        limit: 0,
        offset: 0
    },
    exact_include_conditions: {} as Record<string, string[] | string>,
    fuzzy_include_conditions: {} as Record<string, string[]>,
  };
  searchSelectValue.value.forEach((item: any) => {
    const target = params.exact_include_conditions;
    target[item.id] = item.values.map((value: any) => value.id);
  });
  params.exact_include_conditions['workflow_id'] = route.params.taskId;
  return params;
}
const needInterval = ref(false);
const getOperateList = async () => {
    const currentRowData = nodeManageStore.taskHistoryTableRowData;
    const searchParameters = getParams();
    const res = await NodeWorkflowService.NodeWorkflowOperationList(searchParameters).catch(() => ({
        operations: [],
        total_count: 0
    }));
    const mapList = res.operations.map(item => {
        needInterval.value = ['running', 'empty_instance'].includes(item.status.state);
        return {
            ...item.param,
            ...item.status,
            bk_biz_name: currentRowData?.bk_biz_name || item.bk_biz_id,
            operation_id: item.operation_id
        }
    });
    let equal = false;
    for (let index = 0; index < tableData.value.length; index++) {
        if (tableData.value[index].state === mapList[index]?.state) {
            equal = true;
            break;
        }
    }
    if (!equal) {
        tableData.value = mapList;
    }
}
const logRef = ref<InstanceType<typeof Log>>();
const curRow = ref(null);
const handleViewLog = async (row: any) => {
    curOperationId.value = row.operation_id;
    curRow.value = row;
    nextTick(() => {
        logRef.value?.show();
    });

}
const { start, stop } = useInterval(getOperateList, 10000); // 轮询
const handleStop = async () => {
    await updataCurrentTaskInfo();
    await getOperateList();
}
watch(() => searchSelectValue, async () => {
    await getOperateList();
}, { deep: true });
watch(() => tableData, () => {
    filterOptionSource.node_version.list = filterOptionConfig('node_version', typeMap);
    filterOptionSource.state.list = filterOptionConfig('state', statusMap);
}, { deep: true, immediate: true });
watch(() => needInterval.value, async (val: boolean) => {
    if(!val) {
        stop();
        await updataCurrentTaskInfo();
    }
}, { immediate: true });
onMounted(async() => {
    await updataCurrentTaskInfo();
    await getOperateList();
    if (nodeManageStore.currentStatus === 'running' || needInterval.value) {
        start();
    }
});
onBeforeUnmount(() => {
  stop();
});
</script>
<style lang="postcss" scoped>
.status-icon::before {
  content: '';
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
    background: #CBF0DA;
    border-color: #2CAF5E;
  }
}
.nc-terminated {
  &::before {
    border-color: #EA3636;
    background: #FFDDDD;
  }
}
.dropdownCls {
    :deep(span) {
        display: inline-block;
        vertical-align: middle;
    }
}
</style>