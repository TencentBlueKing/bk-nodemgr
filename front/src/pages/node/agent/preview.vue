<template>
    <Sideslider
        v-model:isShow="isShow"
        :width="1600"
        :title="$t('platform.nodeMan.preview.title')"
        render-directive="if"
        :before-close="handleBeforeClose"
    >
        <template #default>
            <div class="py-[24px] px-[40px]">
                <div class="flex min-h-[56px] bg-[#FFF4E2] border border-[#FFF4E2] rounded-[2px] py-[6px] px-[9px] gap-[9px]">
                    <i class="nodeman-icon nc-tips pt-[2px] text-[#FF9C01]"></i>
                    <div class="flex-1 text-[12px] text-[#4D4F56]">
                        <p>{{ $t('platform.nodeMan.preview.tipTitle') }}</p>
                        <p>{{ $t('platform.nodeMan.preview.firstTip') }}</p>
                        <p>{{ $t('platform.nodeMan.preview.secondTip') }}</p>
                    </div>
                </div>
                <div class="flex justify-between mt-[16px]">
                    <div class="w-[50%] flex gap-[8px]">
                        <Input type="search" v-model="searchValue" :placeholder="$t('platform.nodeMan.preview.placeholder')"/>
                        <copy-ip-dropdown :type="'agent'" :disabled="!selection.length" :data="tableData"></copy-ip-dropdown>
                    </div>
                    <div class="flex gap-[8px]">
                        <Button>{{ $t('platform.nodeMan.preview.button.batchInstall') }}</Button>
                        <Button>{{ $t('platform.nodeMan.preview.button.batchRemove') }}</Button>
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
                                <close
                                    v-if="item.icon === 'wrong'"
                                    width="14px"
                                    height="14px"
                                    :fill="item.iconColor"
                                />
                                <i v-else :class="`nodeman-icon nc-${item.icon} text-[14px]`" :style="{ color: item.iconColor }"></i>
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
                        <template #panel>
                            <Table
                                :data="tableData"
                                :empty-text="$t('table.empty')"
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
                                <TableColumn field="bk_host_innerip" :title="t('platform.nodeMan.inner_ip')" min-width="150" fixed="left"></TableColumn>
                                <TableColumn field="bk_host_innerip_v6" :title="t('platform.nodeMan.inner_ipv6')" min-width="150"></TableColumn>
                                <TableColumn field="os_type" :title="t('platform.nodeMan.os_type')" min-width="100"></TableColumn>
                                <TableColumn field="bk_host_name" :title="t('platform.nodeMan.bk_host_name')" min-width="100"></TableColumn>
                                <TableColumn field="bk_networkarea_name" :title="t('platform.nodeMan.bk_cloud_name')" min-width="100"></TableColumn>
                                <TableColumn field="node_status" :title="t('platform.nodeMan.status')" min-width="344">
                                    <template #default="{ row }">
                                        <div class="flex">
                                            <i :class="`nodeman-icon nc-${row.node_status} text-[${row.node_status === 'confirm' ? '#FF9C01' : '#1CAB88'}]`"></i>
                                            <p>动态寻址下已存在相同 IP 的安装记录，去处理</p>
                                            <PopConfirm
                                                width="780"
                                                :title="$t('platform.nodeMan.preview.popConfirm.title')"
                                                trigger="click"
                                                @confirm="ensure"
                                            >
                                                <Button theme="primary" text>{{ t('platform.nodeMan.preview.button.deel') }}</Button>
                                                <template #content>
                                                    <div class="mt-[6px] mb-[8px] text-[12px]">{{ t('platform.nodeMan.preview.popConfirm.tip') }}</div>
                                                    <Table
                                                        class="mb-[24px]"
                                                        :data="tableData"
                                                        :empty-text="t('table.empty')"
                                                        show-overflow-tooltip
                                                        :column-config="{ resizable: true }"
                                                        :max-height="462"
                                                        :show-settings="isShowSetting"
                                                        :settings="settings"
                                                        @setting-change="handleSettingChange"
                                                        @checkbox-change="handleSelectChange"
                                                        @checkbox-all="handleSelectAllChange"
                                                    >
                                                        <template #prepend>
                                                            <div class="bg-[#F0F5FF] h-[32px] flex items-center pl-[16px]">
                                                                <Radio v-model="radioValue" label="cmdb">{{ t('platform.nodeMan.preview.popConfirm.prepend') }}</Radio>
                                                            </div>
                                                        </template>
                                                        <TableColumn field="bk_host_innerip" :title="t('platform.nodeMan.inner_ip')" min-width="200">
                                                            <template #default="{ row }">
                                                                <Radio v-model="row.check" label="cmdb">
                                                                    <span>{{ t('platform.nodeMan.preview.popConfirm.install') }}</span>
                                                                    <span>{{ row.bk_host_innerip }}</span>
                                                                </Radio>
                                                            </template>
                                                        </TableColumn>
                                                        <TableColumn field="bk_host_innerip_v6" :title="t('platform.nodeMan.inner_ipv6')" min-width="150"></TableColumn>
                                                        <TableColumn field="os_type" :title="t('platform.nodeMan.os_type')" min-width="100"></TableColumn>
                                                        <TableColumn field="bk_host_name" :title="t('主机名')" min-width="100"></TableColumn>
                                                        <TableColumn field="bk_networkarea_name" :title="t('platform.nodeMan.bk_cloud_name')" min-width="150"></TableColumn>
                                                    </Table>
                                                </template>
                                            </PopConfirm>
                                        </div>
                                    </template>
                                </TableColumn>
                                <TableColumn field="action" :title="t('platform.nodeMan.operate')" min-width="100" fixed="right">
                                    <template #default="{ row }">
                                        <Button theme="primary" text ext-cls="reinstall">{{ t('platform.nodeMan.preview.button.remove') }}</Button>
                                    </template>
                                </TableColumn>
                            </Table>
                        </template>
                    </Tab.TabPanel>
                </Tab>
            </div>
        </template>
        <template #footer>
            <div class="flex justify-start gap-[8px]">
                <Button theme="primary" @click="handleSetup" :disabled="!tableData.length">{{ t('platform.nodeMan.preview.button.performInstallation') }}</Button>
                <Button @click="handleBeforeClose">{{ t('action.cancel') }}</Button>
            </div>
        </template>
    </Sideslider>
</template>
<script lang="ts" setup>
import { computed, watch, ref } from 'vue';
import { Sideslider, Input, Button, Tag, Tab, PopConfirm, Radio } from 'bkui-vue';
import { Table, TableColumn } from '@blueking/table';
import useTableSetting from '@/composables/use-table-setting';
import { useI18n } from 'vue-i18n';
import { NodeAgentService } from '@/api/modules/node_agent';
import { AgentInstallInfo } from '@/@types/node_agent.d'
import { useRoute, useRouter } from 'vue-router';
import { useNodeManageStore } from '@/stores/node-manage';
import { Close } from 'bkui-vue/lib/icon';

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

const tabs = computed(() => ([
    { label: t('platform.nodeMan.preview.label.all'), name: 'all', count: tableData.value.length},
    { label: t('platform.nodeMan.preview.label.pendingConfirmation'), name: 'confirm', count: 3, icon: 'danger-fill', iconColor: '#FF9C01' },
    { label: t('platform.nodeMan.preview.label.error'), name: 'error', count: 2, icon: 'wrong', iconColor: '#EA3636'  },
    { label: t('platform.nodeMan.preview.label.cleanInstallation'), name: 'CMDB', count: 3, icon: 'check-circle-fill', iconColor: '#1CAB88' },
    { label: t('platform.nodeMan.preview.label.normalInstallation'), name: 'setup', count: 3, icon: 'check-circle-fill', iconColor: '#1CAB88'  }
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
const ensure = () => {

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
            bk_networkunit_id: Number(props.data.bk_networkunit_id),
            bk_host_id: Number(item.bk_host_id),
            bk_networkarea_name: props.data.bk_networkarea_name
        }));
    }
}, {immediate: true, deep: true});
</script>
<style lang="postcss" scoped>
:deep(.bk-tab-header) {
    background: #F0F1F5;
}
:deep(.bk-tab-content) {
    padding: 0;
}

</style>