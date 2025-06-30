<template>
    <PageHeader
    class="w-full sticky top-0 z-1"
    :title="t('任务详情')"
    :back="true"
    >
        <span class="mx-[6px] text-[#979BA5] text-[14px]">-</span>
        <span class="text-[#979BA5] text-[14px] mr-[14px]" v-if="currentData">{{ currentData?.workflow_id }}</span>
        <Tag theme="warning" type="filled" v-if="currentStatus">{{ statusMap[currentStatus]?.text || '' }}</Tag>
    </PageHeader>
    <div class="p-[24px]">
        <div class="flex">
            <div v-for="item in taskInfoList" :key="item.name" class="leading-[30px] mr-[52px] text-[12px]">
                <div class="w-[50px]">{{ item.name }}</div>
                <div v-if="item.prop.includes('time')">{{ timeFormatter(nodeManageStore.taskHistoryTableRowData?.[item.prop]) }}</div>
                <div v-else>{{ nodeManageStore.taskHistoryTableRowData?.[item.prop] }}</div>
            </div>
        </div>
        <div class="mt-[24px] mb-[14px] flex justify-between">
            <div class="flex gap-[12px]">
                <Button>全部失败重试</Button>
                <copy-ip-dropdown
                    type="agent"
                    :list="list"
                    :data="tableData"
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
            <div class="flex-1 min-w-[300px] gap-[8px]">
                <Input
                    type="search"
                    v-model="searchValue"
                    :placeholder="t('请输入IP、云区域、业务、目标版本 搜索')"
                ></Input>
            </div>
        </div>
        <bk-loading title="数据加载中" :loading="loading">
            <ResizeLayout
                :initial-divide="logData.total ? '20%' : '100%'"
                placement="left"
                :disabled="logData.total"
            >
                <template #aside>
                    <Table
                        :data="tableData"
                        :empty-text="'暂无数据'"
                        :pagination="pagination"
                        show-overflow-tooltip
                        :max-height="maxHeight"
                        @checkbox-change="handleSelectChange"
                        @checkbox-all="handleSelectAllChange"
                        :show-settings="isShowSetting"
                        :settings="settings"
                        @setting-change="handleSettingChange"
                    >
                        <TableColumn type="checkbox" width="80" fixed="left"></TableColumn>
                        <TableColumn field="inner_ipv4" :title="'IP'" width="150" fixed="left"></TableColumn>
                        <TableColumn field="area_id" :title="'云区域'"></TableColumn>
                        <TableColumn field="bk_biz_name" :title="t('业务')"></TableColumn>
                        <TableColumn field="node_version" :title="'目标版本'" :filter="versionFilterOption"></TableColumn>
                        <TableColumn field="timeout_second" :title="t('耗时')">
                            <template #default={row}>
                                <span>{{ formatTimeToMS(row.timeout_second) }}</span>
                            </template>
                        </TableColumn>
                        <TableColumn
                            field="state"
                            :title="t('执行状态')"
                            :filter="stateFilterOption"
                        >
                            <template #default="{ row }">
                                <div class="flex items-center" v-if="row.state">
                                    <Spinner v-if="row.state === 'running'" class="mr-[8px]"/>
                                    <template v-else>
                                        <i :class="`nodeman-icon nc-${statusMap[row.state]?.icon} status-icon`"></i>
                                    </template>
                                    <span>{{ statusMap[row.state]?.text }}</span>
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
                            width="100"
                        >
                            <template #default={row}>
                                <div class="flex items-center">
                                    <Button text theme="primary" class="mr-[11px]" @click="handleViewLog(row)">查看日志</Button>
                                    <Button text theme="primary" v-if="row.state === 'failed'">
                                        <right-turn-line fill="#3A84FF"/>
                                        <span class="ml-[3px]">重试</span>
                                    </Button>
                                </div>
                            </template>
                        </TableColumn>
                    </Table>
                </template>
                <template #main>
                    <div ref="contentRef" class="h-[500px] flex flex-col overflow-y-auto" :class="{'text-[12px]': !isFullscreen, 'text-[16px]': isFullscreen}">
                        <div class="h-[50px] flex flex-shrink-0 justify-between items-center px-[16px] bg-[#2E2E2E] text-[#C4C6CC]">
                            <div>{{ t('执行日志') }}</div>
                            <div>
                                <Dropdown
                                    :popover-options="{
                                        clickContentAutoHide: true,
                                        boundary: 'body',
                                        trigger: 'click'
                                    }"
                                >
                                    <Button text class="mr-[16px]">
                                        <span class="text-[#C4C6CC] mr-[6px]">{{ curOperInstVal }}</span>
                                        <angle-up-fill class="text-[16px] text-[#C4C6CC]" />
                                    </Button>
                                    <template #content>
                                        <ul class="w-[80px] py-[4px]">
                                            <li
                                                v-for="item in operInstList"
                                                :key="item.name"
                                                @click="handleClick(item)"
                                                :class="['w-full h-[32px] flex items-center px-[12px]',
                                                    {'bg-[#E1ECFF] text-[#3A84FF]': curOperInstVal === item.name}
                                                ]"
                                            >
                                                {{ item.name }}
                                            </li>
                                        </ul>
                                    </template>
                                </Dropdown>
                                <Button text v-if="isFullscreen" @click="switchFullScreen">
                                    <i class="nodeman-icon nc-icon-un-full-screen text-[16px] text-[#C4C6CC]"></i>
                                </Button>
                                <Button text v-else @click="switchFullScreen">
                                    <i class="nodeman-icon nc-icon-full-screen text-[16px] text-[#C4C6CC]"></i>
                                </Button>
                            </div>
                        </div>
                        <div class="bg-[#1A1A1A] flex-1 flex">
                            <div class="w-[258px] border-r border-[#0A0A0A] text-[#a8acb8]">
                                <div
                                    v-for="(item, key, index) in logData"
                                    :key="key"
                                    @click="handleToggleLogItem(key)"
                                    :class="['h-[32px] flex items-center pl-[16.8px] cursor-pointer', {'bg-[#242424]': activeKey === key}]"
                                >
                                    <template v-if="key !== 'total'">
                                        <success :fill="activeKey === key ? '#24954f' : '#4D4F56'" v-if="item.life_cycle?.state === 'success'"/>
                                        <close :fill="activeKey === key ? '#993D3D' : '#4D4F56'" v-if="item.life_cycle?.state === 'failed'"/>
                                        <span class="ml-[7px]">{{ index + 1 }}.</span>
                                        <span class="ml-[2px] mr-[4px]">{{ key }}</span>
                                        <span>{{ item.life_cycle?.end_time - item.life_cycle?.start_time }}s</span>
                                    </template>
                                </div>
                            </div>
                            <div class="flex-1 text-[#a8acb8]" v-if="logs">
                                <div v-for="(item, index) in logs" :key="index" class="mx-[24px] my-[8px]">
                                    <span class="mr-[8px]">
                                        [{{ timeFormatter(item.time) }}]
                                    </span>
                                    <span>{{ item.text }}</span>
                                </div>
                            </div>
                        </div>
                    </div>
                </template>
            </ResizeLayout>
        </bk-loading>
    </div>
</template>
<script setup lang="ts">
import { RightTurnLine, Success, Close, AngleUpFill } from 'bkui-vue/lib/icon';
import { ref, reactive, computed, onMounted } from 'vue';
import { Table, TableColumn } from '@blueking/table';
import { Button, Input, Radio, Tag, ResizeLayout, Dropdown } from 'bkui-vue';
import usePage from '@/composables/use-page';
import { useMainStore } from '@/stores/main';
import { useI18n } from 'vue-i18n';
import { NodeWorkflowService } from '@/api/modules/node_workflow';
import { useRoute } from 'vue-router';
import { useNodeManageStore } from '@/stores/node-manage';
import useTableSetting from '@/composables/use-table-setting';
import useFullScreen from '@/composables/use-fullscreen';
import dayjs from 'dayjs';

const { t } = useI18n();
const route = useRoute();
const mainStore = useMainStore();
const nodeManageStore = useNodeManageStore();
// 全屏
const { contentRef, isFullscreen, switchFullScreen } = useFullScreen();
const maxHeight = computed(() => mainStore.windowInnerHeight - 214);
const currentData = nodeManageStore.taskHistoryTableRowData;
const currentStatus = nodeManageStore.currentStatus;
const statusMap = {
  running: {
    text: t('执行中')
  },
  failed: {
    text: t('失败'),
    icon: 'terminated'
  },
  success: {
    text: t('成功'),
    icon: 'running'
  },
  partial_failed: {
    text: t('部分失败'),
    icon: 'warning'
  },
  ignored: {
    text: t('已忽略（没有需要变更的实例）'),
    icon: 'warning'
  }
}
const taskInfoList = ref([
    {prop: 'type', name: t('任务类型'), value: ''},
    {prop: 'cost_time', name: t('总耗时'), value: ''},
    {prop: 'workflow_id', name: t('任务ID'), value: ''},
    {prop: 'operator', name: t('执行人'), value: ''},
    {prop: 'operate_time', name: t('执行时间'), value: ''},
]);
const searchValue = ref('');
const tableData = ref([]);
const radioGroupValue = ref('all');
const curOperationId = ref('');
const curOperInstId = ref('');
const curOperInstVal = ref('LATEST');
const operInstList = ref([]);
const logData = ref<Record<string, ActionMessage>>({
    total: 0,
    oper_inst_logs: {}
});
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
const filterOptionConfig = (prop: string, map?: Record<string, any>) => {
    const uniqueValues = Array.from(new Set(tableData.value.map((item: any) => item[prop]).filter((item: any) => item)));
    return {
      list: uniqueValues.map(value => ({
        text: map ? map[value as string].text : value,
        value: value
      })),
      checked: [] as string[],
      filterScope: 'all',
    }
}
const versionFilterOption = computed(() => filterOptionConfig('node_version'));
const stateFilterOption = computed(() => filterOptionConfig('state', statusMap));
const loading = ref(false);
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
    'inner_ipv4',
    'area_id',
    'bk_biz_name',
    'node_version',
    'timeout_second',
    'state',
  ],
  disabled: [],
});
const formatTimeToMS = (duration: number) => {
  const minutes = Math.floor(duration / 60000);
  const seconds = Math.floor((duration % 60000) / 1000);
  return `${minutes}m ${seconds}s`;
}
const timeFormatter = (val: string, format = 'YYYY-MM-DD HH:mm:ss') => {
  return val ? dayjs(val).format(format) : '--';
}
const handleClick = async (item: {name: string, id: string}) => {
    curOperInstId.value = item.id;
    await getLog();
}
const getOperateList = async () => {
    loading.value = true;
    const currentRowData = nodeManageStore.taskHistoryTableRowData
    const res = await NodeWorkflowService.NodeWorkflowOperationList({
        workflow_id: route.params.taskId
    }).catch(() => ({
        operations: [],
        total_count: 0
    }));
    tableData.value = res.operations.map(item => ({
        ...item.param,
        ...item.status,
        bk_biz_name: currentRowData?.bk_biz_name,
        operation_id: item.operation_id
    }));
    loading.value = false;
}
const getInstance = async () => {
    const res = await NodeWorkflowService.NodeWorkflowOperationInstanceList({
        operation_id: curOperationId.value
    }).catch(() => ({
        oper_inst_data: [],
        total: 0
    }));
    curOperInstId.value = res.total > 0 ?  res.oper_inst_data[res.total -1].oper_inst_id : '';
    operInstList.value = [];
    for(let i = 1; i <= res.total; i++){
        operInstList.value.push({
            name: i < res.total ? `${i}nd` : 'LATEST',
            id: res.oper_inst_data[i - 1].oper_inst_id,
        });
    }
}
const logs = ref([]);
const activeKey = ref('');
const getLog = async () => {
    const res = await NodeWorkflowService.NodeWorkflowOperationInstanceLogGet({
        oper_inst_id: curOperInstId.value
    }).catch(() => ({
        total: 0,
        oper_inst_logs: {}
    }));
    logData.value = res.oper_inst_logs;
    logData.value.total = 1;
    const firstKey = Object.keys(res.oper_inst_logs)[0];
    logs.value = { ...res.oper_inst_logs[firstKey].message.logs };
    activeKey.value = firstKey;
}
const handleViewLog = async (row: any) => {
    curOperationId.value = row.operation_id;
    await getInstance();
    await getLog();
}
const handleToggleLogItem = (key: string) => {
    logs.value = logData.value[key].message.logs;
    activeKey.value = key;
}
onMounted(async() => {
    await getOperateList();
})
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
</style>