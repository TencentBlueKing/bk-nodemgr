<template>
    <div class="p-[24px]">
        <div class="bg-[#f0f1f5] w-full rounded-[2px]">
            <div class="min-w-[690px] max-w-[879px] flex flex-wrap py-[14px] px-[24px]">
                <div class="mb-[16px] text-[16px] flex w-full">
                    <span class="mr-[8px]">重装Agent</span>
                    <div class="bg-[#ea3636] text-[#fff]">执行失败</div>
                </div>
                <div v-for="(item, index) in info" :key="item" class="flex leading-[30px] mr-[10px] text-[12px]">
                    <div class="w-[50px]">{{ index }}</div>
                    <div class="w-[200px]">: {{ info[index] }}</div>
                </div>
            </div>
        </div>
        <div class="mt-[24px] mb-[14px] flex justify-between">
            <div class="flex w-[50%] gap-[12px]">
                <copy-ip-dropdown type="agent" :list="list" :data="tableData"></copy-ip-dropdown>
                <Button>终止</Button>
                <Button>失败重试</Button>
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
            >
            <TableColumn field="bk_host_innerip" :title="t('platform.nodeMan.inner_ip')" width="150" fixed="left"></TableColumn>
            <TableColumn field="bk_host_innerip_v6" :title="t('platform.nodeMan.inner_ipv6')" width="150"></TableColumn>
            <TableColumn field="bk_agent_id" :title="t('platform.nodeMan.agentId')" width="295"></TableColumn>
            <TableColumn field="bk_bussiness" :title="t('业务')" width="150" fixed="left"></TableColumn>
            <TableColumn field="bk_operate_type" :title="t('操作类型')" width="150"></TableColumn>
            <TableColumn field="install_type" :title="t('安装方式')" width="295"></TableColumn>
            <TableColumn field="cost_time" :title="t('耗时')" width="150" fixed="left"></TableColumn>
            <TableColumn field="task_status" :title="t('执行状态')" width="150"></TableColumn>
            <TableColumn field="action" :title="t('操作')" width="295"></TableColumn>
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
    '任务ID': '123',
    '开始时间': '2021-12-12 12:12:12',
    '总耗时': '6m 19s',
    '任务类型': 'Agent',
    '操作类型': '重装',
    '执行账号': 'admin'
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
</script>