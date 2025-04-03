<template>
    <div class="p-[24px]">
        <div class="bg-[#f0f1f5] w-full rounded-[2px]">
            <div class="flex p-[24px]">
                <div v-for="(item, index) in info" :key="item" class="leading-[30px] mr-[52px] text-[12px]">
                    <div class="w-[50px]">{{ index }}</div>
                    <div class="w-[200px]">:{{ info[index] }}</div>
                </div>
            </div>
        </div>
        <div class="mt-[24px] mb-[14px] flex justify-between">
            <div class="flex w-[50%] gap-[12px]">
                <Button>全部失败重试</Button>
                <copy-ip-dropdown type="agent" :list="list" :data="tableData" :disabled="!selection.length"></copy-ip-dropdown>
                <div class="w-[269px] h-[32px] bg-[#EAEBF0] rounded-[2px]">
                    <Button size="small" theme="primary">全部(123)</Button> 
                    <Button size="small">成功(89)</Button>
                    <Button size="small">失败(45)</Button>
                </div>
            </div>
            <div class="w-[30%] min-w-[300px] gap-[8px]">
                <Input type="search" v-model="searchValue"></Input>
            </div>
        </div>
        <bk-loading title="数据加载中" :loading="loading">
            <Table
                :data="tableData"
                :empty-text="'暂无数据'"
                :pagination="pagination"
                show-overflow-tooltip
                :max-height="maxHeight"
                @checkbox-change="handleSelectChange"
                @checkbox-all="handleSelectAllChange"
            >
                <TableColumn type="checkbox" width="80" fixed="left"></TableColumn>
                <TableColumn field="bk_host_innerip" :title="'IP'" width="150"></TableColumn>
                <TableColumn field="bk_networkarea_name" :title="'云区域'" width="150"></TableColumn>
                <TableColumn field="bk_bussiness" :title="t('业务')" width="150"></TableColumn>
                <TableColumn field="bk_agent_id" :title="'目标版本'" width="295"></TableColumn>
                <TableColumn field="cost_time" :title="t('耗时')" width="150"></TableColumn>
                <TableColumn field="task_status" :title="t('执行状态')" width="150"></TableColumn>
            </Table>
        </bk-loading>
    </div>
</template>
<script setup lang="ts">
import search from 'bkui-vue/lib/icon/search';
import { ref, reactive, computed } from 'vue';
import { Table, TableColumn } from '@blueking/table';
import { Button, Input } from 'bkui-vue';
import usePage from '@/composables/use-page';
import { useMainStore } from '@/stores/main';
import { useI18n } from 'vue-i18n';

const { t } = useI18n();
const mainStore = useMainStore();
const maxHeight = computed(() => mainStore.windowInnerHeight - 214);
const info = reactive({
    '操作类型': '重装',
    '任务类型': 'Agent',
    '总耗时': '6m 19s',
    '任务ID': '123',
    '执行人': 'admin',
    '执行时间': '2021-12-12 12:12:12',
});
const searchValue = ref('');
const tableData = ref([
    
]);
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
</script>