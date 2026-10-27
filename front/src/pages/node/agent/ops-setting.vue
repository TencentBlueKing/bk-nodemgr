<template>
    <div class="ops-setting pt-[24px] pb-[48px]">
        <div class="m-[24px]">
            <p class="text-[14px] text-[#63656E] mb-[16px]">{{ $t('platform.nodeMan.agentNodeStatus.opsSettingPageDesc') }}
            </p>

            <Loading :loading="loading">
                <VxeTable ref="xTableRef" :data="tableData" :size="'medium'" :border="true" :max-height="600"
                    :scroll-y="{ enabled: true, gt: 20 }" :row-config="{ isHover: true, useKey: true }"
                    :edit-config="{ trigger: 'click', mode: 'row', showIcon: false }"
                    round>
                    <!-- 业务属性 -->
                    <VxeColgroup :title="$t('components.installTable.bizProperty')" align="center">
                        <VxeColumn field="bk_biz_id" :title="$t('components.installTable.bkBizId')" :min-width="150">
                            <template #default="{ row }">
                                <Select v-model="row.bk_biz_id" :disabled="true" filterable>
                                    <Select.Option v-for="item in businessList" :key="item.bk_biz_id"
                                        :name="item.bk_biz_name" :id="item.bk_biz_id">
                                        [{{ item.bk_biz_id }}] {{ item.bk_biz_name }}
                                    </Select.Option>
                                </Select>
                            </template>
                        </VxeColumn>
                    </VxeColgroup>

                    <!-- 拓扑属性 -->
                    <VxeColgroup :title="$t('components.installTable.topoProperty')" align="center">
                        <VxeColumn field="bk_networkarea_name" :title="$t('components.installTable.networkArea')"
                            :min-width="150">
                            <template #default="{ row }">
                                <Input v-model.trim="row.bk_networkarea_name" :disabled="true" />
                            </template>
                        </VxeColumn>
                    </VxeColgroup>

                    <!-- 主机 IP -->
                    <VxeColgroup align="center">
                        <template #header>
                            <span>{{ $t('components.installTable.hostIp') }}</span>
                        </template>
                        <VxeColumn field="bk_host_innerip" :title="$t('components.installTable.innerIPv4')"
                            :min-width="150">
                            <template #default="{ row }">
                                <Input v-model.trim="row.bk_host_innerip" :disabled="true" />
                            </template>
                        </VxeColumn>
                        <VxeColumn field="bk_host_innerip_v6" :title="$t('components.installTable.innerIPv6')"
                            :min-width="150">
                            <template #default="{ row }">
                                <Input v-model.trim="row.bk_host_innerip_v6" :disabled="true" />
                            </template>
                        </VxeColumn>
                    </VxeColgroup>

                    <!-- 带外管理 -->
                    <VxeColgroup :title="$t('platform.nodeMan.agentNodeStatus.opsSetting')" align="center">
                        <VxeColumn
                            field="ops_console_host_id"
                            :title="$t('platform.nodeMan.agentNodeStatus.opsConsoleHostId')"
                            :min-width="200"
                        >
                            <template #header>
                                <Popover
                                    theme="light"
                                    trigger="hover"
                                    placement="top"
                                    :arrow="true"
                                    :max-width="300"
                                    :offset="8"
                                    :popover-delay="[0, 100]"
                                    :component-event-delay="0"
                                >
                                    <span class="cursor-default" style="border-bottom: 1px dashed #c4c6cc">
                                        {{ $t('platform.nodeMan.agentNodeStatus.opsConsoleHostId') }}
                                    </span>
                                    <template #content>
                                        <div class="text-[12px] leading-[20px]">
                                            <p>{{ $t('platform.nodeMan.agentNodeStatus.opsConsoleHostIdTooltip') }}</p>
                                        </div>
                                    </template>
                                </Popover>
                            </template>
                            <template #default="{ row }">
                                <!-- 主控机通过 IP 选择器选取：表格回显 IP，实际提交 host-id -->
                                <div class="ops-console-host-cell" @click="handleOpenOpsConsoleHostSelector(row)">
                                    <span v-if="row.ops_console_host_ip">{{ row.ops_console_host_ip }}</span>
                                    <span v-else-if="row.ops_console_host_id">{{ row.ops_console_host_id }}</span>
                                    <span v-else class="cell-placeholder">{{ $t('components.installTable.selectPlaceholder') }}</span>
                                </div>
                            </template>
                        </VxeColumn>
                        <VxeColumn
                            field="ops_out_band_type"
                            :title="$t('platform.nodeMan.agentNodeStatus.opsOutBandType')"
                            :min-width="140"
                            :edit-render="{ name: 'VxeInput' }"
                        >
                            <template #header>
                                <Popover
                                    theme="light"
                                    trigger="hover"
                                    placement="top"
                                    :arrow="true"
                                    :max-width="300"
                                    :offset="8"
                                    :popover-delay="[0, 100]"
                                    :component-event-delay="0"
                                >
                                    <span class="mr-[5px] cursor-default" style="border-bottom: 1px dashed #c4c6cc">
                                        {{ $t('platform.nodeMan.agentNodeStatus.opsOutBandType') }}
                                    </span>
                                    <template #content>
                                        <div class="text-[12px] leading-[20px]">
                                            <p>{{ $t('platform.nodeMan.agentNodeStatus.opsOutBandTypeTooltip') }}</p>
                                        </div>
                                    </template>
                                </Popover>
                                <BatchEdit :title="$t('platform.nodeMan.agentNodeStatus.opsOutBandType')" type="input"
                                    @confirm="(value: string) => handleBatchEdit('ops_out_band_type', value)" />
                            </template>
                            <template #default="{ row }">
                                <span v-if="row.ops_out_band_type">{{ row.ops_out_band_type }}</span>
                                <span v-else class="cell-placeholder">{{ $t('components.installTable.inputPlaceholder') }}</span>
                            </template>
                            <template #edit="{ row }">
                                <Input v-model.trim="row.ops_out_band_type" />
                            </template>
                        </VxeColumn>
                        <VxeColumn
                            field="ops_out_band_protocol"
                            :title="$t('platform.nodeMan.agentNodeStatus.opsOutBandProtocol')"
                            :min-width="140"
                            :edit-render="{ name: 'VxeInput' }"
                        >
                            <template #header>
                                <Popover
                                    theme="light"
                                    trigger="hover"
                                    placement="top"
                                    :arrow="true"
                                    :max-width="300"
                                    :offset="8"
                                    :popover-delay="[0, 100]"
                                    :component-event-delay="0"
                                >
                                    <span class="mr-[5px] cursor-default" style="border-bottom: 1px dashed #c4c6cc">
                                        {{ $t('platform.nodeMan.agentNodeStatus.opsOutBandProtocol') }}
                                    </span>
                                    <template #content>
                                        <div class="text-[12px] leading-[20px]">
                                            <p>{{ $t('platform.nodeMan.agentNodeStatus.opsOutBandProtocolTooltip') }}</p>
                                        </div>
                                    </template>
                                </Popover>
                                <BatchEdit :title="$t('platform.nodeMan.agentNodeStatus.opsOutBandProtocol')" type="select"
                                    :options="opsOutBandProtocolOptions" allow-create
                                    @confirm="handleBatchProtocolEdit" />
                            </template>
                            <template #default="{ row }">
                                <span v-if="row.ops_out_band_protocol">{{ row.ops_out_band_protocol }}</span>
                                <span v-else class="cell-placeholder">{{ $t('components.installTable.inputPlaceholder') }}</span>
                            </template>
                            <template #edit="{ row }">
                                <Select
                                    v-model="row.ops_out_band_protocol"
                                    filterable
                                    allow-create
                                    :popover-options="{ extCls: 'vxe-table--ignore-clear' }"
                                    @change="(value: string) => handleProtocolChange(row, value)"
                                    @focusout="handleProtocolFocusout(row, $event)"
                                >
                                    <Select.Option
                                        v-for="item in opsOutBandProtocols"
                                        :id="item.name"
                                        :key="item.name"
                                        :name="item.name"
                                        :title="item.description"
                                    >
                                        {{ item.name }}
                                    </Select.Option>
                                </Select>
                            </template>
                        </VxeColumn>
                        <VxeColumn
                            field="ops_out_band_ip"
                            :title="$t('platform.nodeMan.agentNodeStatus.opsOutBandIP')"
                            :min-width="140"
                            :edit-render="{ name: 'VxeInput' }"
                        >
                            <template #header>
                                <Popover
                                    theme="light"
                                    trigger="hover"
                                    placement="top"
                                    :arrow="true"
                                    :max-width="300"
                                    :offset="8"
                                    :popover-delay="[0, 100]"
                                    :component-event-delay="0"
                                >
                                    <span class="mr-[5px] cursor-default" style="border-bottom: 1px dashed #c4c6cc">
                                        {{ $t('platform.nodeMan.agentNodeStatus.opsOutBandIP') }}
                                    </span>
                                    <template #content>
                                        <div class="text-[12px] leading-[20px]">
                                            <p>{{ $t('platform.nodeMan.agentNodeStatus.opsOutBandIPTooltip') }}</p>
                                        </div>
                                    </template>
                                </Popover>
                                <BatchEdit :title="$t('platform.nodeMan.agentNodeStatus.opsOutBandIP')" type="input"
                                    @confirm="(value: string) => handleBatchEdit('ops_out_band_ip', value)" />
                            </template>
                            <template #default="{ row }">
                                <span v-if="row.ops_out_band_ip">{{ row.ops_out_band_ip }}</span>
                                <span v-else class="cell-placeholder">{{ $t('components.installTable.inputPlaceholder') }}</span>
                            </template>
                            <template #edit="{ row }">
                                <Input v-model.trim="row.ops_out_band_ip" />
                            </template>
                        </VxeColumn>
                        <VxeColumn
                            field="ops_out_band_port"
                            :title="$t('platform.nodeMan.agentNodeStatus.opsOutBandPort')"
                            :min-width="140"
                            :edit-render="{ name: 'VxeInput' }"
                        >
                            <template #header>
                                <Popover
                                    theme="light"
                                    trigger="hover"
                                    placement="top"
                                    :arrow="true"
                                    :max-width="300"
                                    :offset="8"
                                    :popover-delay="[0, 100]"
                                    :component-event-delay="0"
                                >
                                    <span class="cursor-default" style="border-bottom: 1px dashed #c4c6cc">
                                        {{ $t('platform.nodeMan.agentNodeStatus.opsOutBandPort') }}
                                    </span>
                                    <template #content>
                                        <div class="text-[12px] leading-[20px]">
                                            <p>{{ $t('platform.nodeMan.agentNodeStatus.opsOutBandPortTooltip') }}</p>
                                        </div>
                                    </template>
                                </Popover>
                                <BatchEdit :title="$t('platform.nodeMan.agentNodeStatus.opsOutBandPort')" type="input"
                                    @confirm="(value: string) => handleBatchEdit('ops_out_band_port', value)" />
                            </template>
                            <template #default="{ row }">
                                <span v-if="row.ops_out_band_port">{{ row.ops_out_band_port }}</span>
                                <span v-else class="cell-placeholder">{{ $t('components.installTable.inputPlaceholder') }}</span>
                            </template>
                            <template #edit="{ row }">
                                <Input v-model.trim="row.ops_out_band_port" />
                            </template>
                        </VxeColumn>
                    </VxeColgroup>
                </VxeTable>
            </Loading>
        </div>

        <!-- 主控机 IP 选择器 -->
        <IpSelector
            mode="dialog"
            :show-dialog="isShowOpsConsoleHostSelector"
            :value="opsConsoleHostSelectorValue"
            :keep-host-field-output="true"
            @change="handleOpsConsoleHostSelectorChange"
            @panel-change="handleIpSelectorPanelChange"
            @close-dialog="isShowOpsConsoleHostSelector = false"
        />

        <!-- 底部提交栏 -->
        <div :class="['h-[48px] w-full flex items-center pl-[24px]', { 'fixed bottom-[0] bg-[#fff] z-[100]': isAtBottom }]"
            ref="footerRef">
            <Button class="min-w-[120px] mr-[8px]" theme="primary" :loading="submitting" @click="handleConfirm">
                {{ $t('action.confirm') }}
            </Button>
            <Button class="w-[88px]" @click="handleCancel">{{ $t('action.cancel') }}</Button>
        </div>
    </div>
</template>

<script lang="ts" setup>
import { Button, Input, Loading, Message, Popover, Select } from 'bkui-vue';
import { debounce } from 'lodash';
import { computed, onMounted, onUnmounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';

import { VxeColgroup, VxeColumn, VxeTable } from '@blueking/vxe-table';

import BatchEdit from '@/components/batch-edit.vue';
import IpSelector from '@/components/IpSelector';
import { NodeAgentService } from '@/api/modules/node_agent';
import { NodeProxyService } from '@/api/modules/node_proxy';
import { TopoService } from '@/api/modules/topo';
import * as IpSelectorService from '@/services/ip-selector';
import { useNodeManageStore } from '@/stores/node-manage';
import { useMainStore } from '@/stores/main';

const { t } = useI18n();
const router = useRouter();
const nodeManageStore = useNodeManageStore();
const mainStore = useMainStore();

const businessList = computed(() => mainStore.businessList);

interface OpsOutBandProtocol {
    name: string;
    port: number;
    description: string;
}

const opsOutBandProtocols = (() => {
    try {
        const protocols = JSON.parse(window.PROJECT_CONFIG.OPS_OUT_BAND_PROTOCOLS || '[]');
        if (!Array.isArray(protocols)) return [];
        return protocols.filter((item): item is OpsOutBandProtocol => (
            item && typeof item.name === 'string' && item.name.trim()
            && Number.isInteger(item.port) && item.port > 0 && item.port <= 65535
        ));
    } catch {
        return [];
    }
})();
const opsOutBandProtocolOptions = opsOutBandProtocols.map(item => ({ id: item.name, name: item.name }));

const tableData = ref<any[]>([]);
const loading = ref(false);
const submitting = ref(false);
const isAtBottom = ref(false);
const footerRef = ref<Element | null>(null);

// 节点类型：agent | proxy
const nodeType = ref<'agent' | 'proxy'>('agent');

const handleCancel = () => {
    if (nodeType.value === 'proxy') {
        router.push({ name: 'proxy' });
    } else {
        router.push({ name: 'agent' });
    }
};

// 批量写入单列
const handleBatchEdit = (field: string, value: string) => {
    for (const row of tableData.value) {
        row[field] = value;
    }
};

const handleProtocolChange = (row: any, value: string) => {
    row.ops_out_band_protocol = String(value ?? '').trim();
    const protocol = opsOutBandProtocols.find(item => item.name === row.ops_out_band_protocol);
    if (protocol) {
        row.ops_out_band_port = protocol.port;
    }
};

// allow-create 仅回车确认，失焦时编辑单元格会被销毁导致输入丢失，这里在失焦时把输入文本确认为值
const handleProtocolFocusout = (row: any, event: FocusEvent) => {
    const text = event.target instanceof HTMLInputElement ? event.target.value.trim() : '';
    if (!text || text === row.ops_out_band_protocol) return;
    handleProtocolChange(row, text);
};

const handleBatchProtocolEdit = (value: string) => {
    for (const row of tableData.value) {
        handleProtocolChange(row, value);
    }
};

// ======== 主控机 IP 选择器（表格回显 IP，提交转 host-id，不预设 node role 限制） ========
const isShowOpsConsoleHostSelector = ref(false);
const opsConsoleHostSelectorValue = ref({
    hostList: [] as any[],
    nodeList: [] as any[],
    dynamicGroupList: [] as any[],
    serviceTemplateList: [] as any[],
    setTemplateList: [] as any[],
});
const opsConsoleHostEditingRow = ref<any>(null);

const handleIpSelectorPanelChange = (panel: string) => {
    IpSelectorService.setActiveIpsPanel(panel);
};

const handleOpenOpsConsoleHostSelector = (row: any) => {
    opsConsoleHostEditingRow.value = row;
    // 不预设 node role 限制：清空策略类型与业务限定（含策略页可能残留的 strategyBizId），
    // 拓扑树展示全部业务、主机查询不带 node_role 过滤
    IpSelectorService.setStrategyBizId(undefined);
    IpSelectorService.setBizId([]);
    opsConsoleHostSelectorValue.value = {
        hostList: row.ops_console_host_id ? [{ hostId: row.ops_console_host_id }] : [],
        nodeList: [],
        dynamicGroupList: [],
        serviceTemplateList: [],
        setTemplateList: [],
    };
    isShowOpsConsoleHostSelector.value = true;
};

const handleOpsConsoleHostSelectorChange = (value: any) => {
    opsConsoleHostSelectorValue.value = value;
    const row = opsConsoleHostEditingRow.value;
    if (!row) return;
    // 主控机为单选：仅保留最后一次选择的主机
    const host = value?.hostList?.length ? value.hostList[value.hostList.length - 1] : null;
    if (host) {
        const hostId = host.hostId ?? host.host_id ?? host.bk_host_id ?? host.id;
        const ip = host.ip || host.ipv6 || '';
        row.ops_console_host_id = hostId;
        row.ops_console_host_ip = ip;
    } else {
        row.ops_console_host_id = undefined;
        row.ops_console_host_ip = undefined;
    }
};

// 提交
const handleConfirm = async () => {
    submitting.value = true;
    try {
        const hosts = tableData.value
            .filter((row: any) => row.bk_host_id)
            .map((row: any) => ({
                bk_host_id: row.bk_host_id,
                ops_console_host_id: Number(row.ops_console_host_id) || 0,
                ops_out_band_type: String(row.ops_out_band_type ?? ''),
                ops_out_band_protocol: String(row.ops_out_band_protocol ?? ''),
                ops_out_band_ip: String(row.ops_out_band_ip ?? ''),
                ops_out_band_port: Number(row.ops_out_band_port) || 0,
            }));

        if (nodeType.value === 'proxy') {
            await NodeProxyService.NodeProxyUpdateOpsFields({ hosts });
        } else {
            await NodeAgentService.NodeAgentUpdateOpsFields({ hosts });
        }

        Message({ theme: 'success', message: t('platform.nodeMan.agentNodeStatus.opsSettingSuccess') });
        setTimeout(() => {
            if (nodeType.value === 'proxy') {
                router.push({ name: 'proxy' });
            } else {
                router.push({ name: 'agent' });
            }
        }, 800);
    } catch (err: any) {
        Message({ theme: 'error', message: err?.message || String(err) });
    } finally {
        submitting.value = false;
    }
};

const checkIfAtBottom = () => {
    if (footerRef.value) {
        const { bottom } = footerRef.value.getBoundingClientRect();
        isAtBottom.value = bottom >= window.innerHeight;
    }
};
const debouncedCheck = debounce(checkIfAtBottom, 100);

onMounted(async () => {
    nodeType.value = nodeManageStore.opsFieldsParams.nodeType || 'agent';

    if (footerRef.value) {
        window.addEventListener('resize', debouncedCheck);
        checkIfAtBottom();
    }

    loading.value = true;
    const params = nodeManageStore.opsFieldsParams;

    if (params.isCrossPageSelection) {
        // 跨页全选：分页拉取所有数据
        const allHosts: any[] = [];
        const pageSize = 500;
        let offset = 0;
        let hasMore = true;

        while (hasMore) {
            const hostListData = await TopoService.HostList({
                page: { offset, limit: pageSize },
                only_count: false,
                exact_include_conditions: params.queryParams?.exact_include_conditions || {},
                exact_exclude_conditions: params.queryParams?.exact_exclude_conditions || {},
                fuzzy_include_conditions: params.queryParams?.fuzzy_include_conditions || {},
            }).catch(() => ({ total: 0, items: [] }));

            if (hostListData.items?.length > 0) {
                allHosts.push(...hostListData.items);
                offset += pageSize;
                if (hostListData.items.length < pageSize) hasMore = false;
            } else {
                hasMore = false;
            }
        }

        tableData.value = formatHostData(allHosts);
    } else {
        tableData.value = formatHostData(params.tableData);
    }
    loading.value = false;
});

// 格式化主机数据为表格行：与 assign-unit 一致，提取 info 字段并初始化 Ops 字段
function formatHostData(hosts: any[]): any[] {
    return hosts.map((item: any) => {
        const info = item.info || {};
        return {
            bk_host_id: item.bk_host_id || info.bk_host_id,
            bk_biz_id: info.bk_biz_id ?? item.bk_biz_id ?? 0,
            bk_networkarea_id: info.bk_networkarea_id ?? item.bk_networkarea_id ?? 0,
            bk_networkarea_name: info.bk_networkarea_name ?? item.bk_networkarea_name ?? '',
            bk_host_innerip: info.bk_host_innerip_list?.join(',') ?? item.bk_host_innerip ?? '',
            bk_host_innerip_v6: info.bk_host_innerip_v6_list?.join(',') ?? item.bk_host_innerip_v6 ?? '',
            // Ops 字段：优先取已有值，否则初始化为空
            ops_console_host_id: item.ops_console_host_id ?? info.ops_console_host_id ?? undefined,
            ops_console_host_ip: item.ops_console_host_ip ?? info.ops_console_host_ip ?? undefined,
            ops_out_band_type: item.ops_out_band_type ?? info.ops_out_band_type ?? undefined,
            ops_out_band_protocol: item.ops_out_band_protocol ?? info.ops_out_band_protocol ?? undefined,
            ops_out_band_ip: item.ops_out_band_ip ?? info.ops_out_band_ip ?? undefined,
            ops_out_band_port: item.ops_out_band_port ?? info.ops_out_band_port ?? undefined,
        };
    });
}

onUnmounted(() => {
    if (footerRef.value) {
        window.removeEventListener('resize', debouncedCheck);
    }
});
</script>

<style lang="postcss" scoped>
::v-deep(.vxe-header--column) {
    font-weight: normal !important;
    font-size: 12px !important;
}

::v-deep(.vxe-body--column) {
    height: 56px !important;
    font-size: 12px !important;
}
.cell-placeholder {
    color: #c4c6cc;
}
.ops-console-host-cell {
    cursor: pointer;
    width: 100%;
    min-height: 32px;
    display: flex;
    align-items: center;
    border-radius: 2px;
    padding: 0 4px;
    transition: background-color .2s;
}
.ops-console-host-cell:hover {
    background-color: #f0f1f5;
}
</style>
