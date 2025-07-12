<template>
    <Sideslider
        v-model:isShow="isShow"
        :width="1200"
        :title="'安装预览'"
        render-directive="if"
        :before-close="handleBeforeClose"
    >
        <template #default>
            <div class="py-[24px] px-[40px]">
                <div class="flex min-h-[56px] bg-[#FFF4E2] border border-[#FFF4E2] rounded-[2px] py-[6px] px-[9px] gap-[9px]">
                    <i class="nodeman-icon nc-tips pt-[2px] text-[#FF9C01]"></i>
                    <div class="flex-1 text-[12px] text-[#4D4F56]">
                        <p>当前业务中发现部分主机 IP 已经在蓝鲸注册，需要手动确认是全新安装还是基于 CMDB 记录安装：</p>
                        <p>1. 全新安装会将该 IP 当作全新主机处理，执行安装并将该主机导入到 CMDB 中；</p>
                        <p>2. 若是已经在 CMDB 中注册过的主机，为了确保节点管理安装记录与 CMDB 主机记录之间的正确匹配，需要手动指定该 IP 所匹配的正确 CMDB 主机记录。</p>
                    </div>
                </div>
                <div class="flex justify-between mt-[16px]">
                    <div class="w-[50%] flex gap-[8px]">
                        <Input type="search" v-model="searchValue" placeholder="输入 关键词"/>
                        <copy-ip-dropdown :type="'agent'" :disabled="!selection.length" :data="tableData"></copy-ip-dropdown>
                    </div>
                    <div class="flex gap-[8px]">
                        <Button>一键处理为：全新安装</Button>
                        <Button>一键移除错误项</Button>
                    </div>
                </div>
                <Tab
                    class="mt-[16px]"
                    v-model:active="active"
                    type="card"
                >
                    <!-- <template #setting>
                        <div class="leading-[50px]"><i class="mr-[16px] nodeman-icon nc-setting"></i></div>
                    </template> -->
                    <Tab.TabPanel v-for="item in tabs" :key="item.name" :label="item.label" :name="item.name">
                        <template #label>
                            <div class="flex gap-[5px] items-center">
                                <i :class="`nodeman-icon nc-${item.icon} text-[${item.iconColor}]`"></i>
                                <span class="text-[14px] text-[#313238]">{{ item.label }}</span>
                                <div 
                                    :class="`rounded-[8px] w-[23px] h-[16px] border text-[12px] leading-[16px] text-center 
                                    bg-[${
                                        item.name === active ? '#E1ECFF' : '#DCDEE5'
                                    }]`">
                                    {{ item.count }}
                                </div>
                            </div>
                        </template>
                    </Tab.TabPanel>
                </Tab>
                <Table
                    :data="tableData"
                    :empty-text="'暂无数据'"
                    :column-config="{ resizable: true }"
                    show-overflow-tooltip
                    :max-height="462"
                    :show-settings="isShowSetting"
                    :settings="settings"
                    @setting-change="handleSettingChange"
                    @checkbox-change="handleSelectChange"
                    @checkbox-all="handleSelectAllChange"
                >
                    <TableColumn type="checkbox" width="80" fixed="left"></TableColumn>
                    <TableColumn field="bk_host_innerip" :title="t('platform.nodeMan.inner_ip')" width="150" fixed="left"></TableColumn>
                    <TableColumn field="bk_host_innerip_v6" :title="t('platform.nodeMan.inner_ipv6')" width="150"></TableColumn>
                    <TableColumn field="os_type" :title="t('platform.nodeMan.os_type')"></TableColumn>
                    <TableColumn field="bk_host_name" :title="t('主机名')"></TableColumn>
                    <TableColumn field="bk_networkarea_name" :title="t('platform.nodeMan.bk_cloud_name')"></TableColumn>
                    <TableColumn field="node_status" :title="t('platform.nodeMan.status')" width="150">
                        <template #default="{ row }">
                            <div class="flex">
                                <i :class="`nodeman-icon nc-${row.node_status} text-[${row.node_status === 'confirm' ? '#FF9C01' : '#1CAB88'}]`"></i>
                                <p>动态寻址下已存在相同 IP 的安装记录，去处理</p>
                                <Popover :is-show="isPopShow">
                                    <Button theme="primary" text @click="isPopShow = true">去处理</Button>
                                    <template #content>
                                        <div class="text-[16px]">当前业务下可能已存在您希望安装的相同 IP主机</div>
                                        <div class="mt-[6px] mb-[8px] text-[12px]">请选择处理方式：</div>
                                        <Table
                                            :data="tableData"
                                            :empty-text="'暂无数据'"
                                            :column-config="{ resizable: true }"
                                            show-overflow-tooltip
                                            :max-height="462"
                                            :show-settings="isShowSetting"
                                            :settings="settings"
                                            @setting-change="handleSettingChange"
                                            @checkbox-change="handleSelectChange"
                                            @checkbox-all="handleSelectAllChange"
                                        >
                                            <template #prepend>
                                                <Radio v-model="radioValue" label="cmdb">全新安装，并导入 CMDB</Radio>
                                            </template>
                                            <TableColumn field="bk_host_innerip" :title="t('platform.nodeMan.inner_ip')" width="150" fixed="left"></TableColumn>
                                            <TableColumn field="bk_host_innerip_v6" :title="t('platform.nodeMan.inner_ipv6')" width="150"></TableColumn>
                                            <TableColumn field="os_type" :title="t('platform.nodeMan.os_type')"></TableColumn>
                                            <TableColumn field="bk_host_name" :title="t('主机名')"></TableColumn>
                                            <TableColumn field="bk_networkarea_name" :title="t('platform.nodeMan.bk_cloud_name')"></TableColumn>
                                        </Table>
                                    </template>
                                </Popover>
                            </div>
                        </template>
                    </TableColumn>
                    <TableColumn field="action" :title="t('platform.nodeMan.operate')" width="100" fixed="right">
                        <template #default="{ row }">
                            <Button theme="primary" text ext-cls="reinstall">移除</Button>
                        </template>
                    </TableColumn>
                </Table>
            </div>
        </template>
        <template #footer>
            <div class="flex justify-start gap-[8px]">
                <Button theme="primary" @click="handleSetup" :disabled="!tableData.length">执行安装</Button>
                <Button @click="handleBeforeClose">取消</Button>
            </div>
        </template>
    </Sideslider>
</template>
<script lang="ts" setup>
import { computed, watch, ref } from 'vue';
import { Sideslider, Select, Input, Button, Tag, Tab, Popover, Radio } from 'bkui-vue';
import { Table, TableColumn } from '@blueking/table';
import useTableSetting from '@/composables/use-table-setting';
import { useI18n } from 'vue-i18n';
import { NodeAgentService } from '@/api/modules/node_agent';
import { AgentInstallInfo } from '@/@types/node_agent.d'
import { useRoute, useRouter } from 'vue-router';
import { useNodeManageStore } from '@/stores/node-manage';

const props = defineProps({
    data: {
        type: Object,
        default: () => {}
    }
});

const { t } = useI18n();
const nodeManageStore = useNodeManageStore();
const router = useRouter();
const isShow = defineModel('isShow', { type: Boolean });
const selection = ref([]);
const searchValue = ref('');
const tableData = ref<AgentInstallInfo[]>([]);
const isPopShow = ref(false);

const tabs = computed(() => ([
    { label: '全部', name: 'all', count: tableData.value.length},
    { label: '待确认', name: 'confirm', count: 3, icon: 'danger-fill', iconColor: '#FF9C01' },
    { label: '错误', name: 'error', count: 2, icon: 'wrong', iconColor: '#EA3636'  },
    { label: '全新安装并导入 CMDB', name: 'CMDB', count: 3, icon: 'check-circle-fill', iconColor: '#1CAB88' },
    { label: '正常安装', name: 'setup', count: 3, icon: 'check-circle-fill', iconColor: '#1CAB88'  }
]));
const active = ref('all');
const { isShowSetting, settings, handleSettingChange } = useTableSetting({
  checked: [
    'bk_host_innerip',
    'bk_host_innerip_v6',
    'bk_agent_id',
    'bk_host_name',
    'bk_networkarea_name',
    'bk_networkarea_id',
    'bk_networkunit_id',
    'os_type',
    'node_version',
    'node_status',
    'action',
  ],
  disabled: ['action'],
});

const handleSelectChange = () => {}
const handleSelectAllChange = () => {}

const radioValue = ref('cmdb');

const handleBeforeClose = () => {
    isShow.value = false;
}
const handleSetup = async () => {
    const res = await NodeAgentService.NodeAgentInstall({
        info: tableData.value,
        target_version: props.data.target_version,
        disable_default_target_version: props.data.disable_default_target_version
    }).catch(() => ({
        workflow_id: ''
    }));
    if(!res) return;
    if (res.workflow_id) {
        router.push({ 
            name: 'taskDetail', 
            params: { taskId: res.workflow_id },
        });
    }
}
watch(() => isShow, () => {
    if(isShow.value && props.data) {
        tableData.value = props.data.info.map(item => ({
            ...item,
            login_port: Number(item.login_port),
            bk_biz_id: props.data.bk_biz_id,
            bk_networkunit_id: props.data.bk_networkunit_id,
            bk_host_id: Number(item.bk_host_id)
        }));
    }
}, {immediate: true, deep: true});
</script>
<style lang="postcss" scoped>
:deep(.bk-tab-header) {
    background: #F0F1F5;
}
</style>