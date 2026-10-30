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
                                    ref="protocolSelectRef"
                                    v-model="row.ops_out_band_protocol"
                                    filterable
                                    allow-create
                                    :popover-options="{ extCls: 'vxe-table--ignore-clear' }"
                                    @toggle="handleProtocolSelectToggle"
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
                                    <template #extension>
                                        <div class="flex items-center h-[32px] px-[8px]">
                                            <template v-if="showProtocolEdit">
                                                <Input
                                                    v-model.trim="protocolCreateText"
                                                    size="small"
                                                    @enter="handleProtocolCreate(row)"
                                                />
                                                <Done
                                                    style="font-size: 22px; color: #2dcb56; cursor: pointer; margin-left: 6px"
                                                    @click="handleProtocolCreate(row)"
                                                />
                                                <Error
                                                    style="font-size: 16px; color: #c4c6cc; cursor: pointer; margin-left: 2px"
                                                    @click="handleProtocolCreateCancel"
                                                />
                                            </template>
                                            <template v-else>
                                                <span
                                                    class="flex items-center cursor-pointer text-[#63656E] hover:text-[#3A84FF]"
                                                    @click="handleShowProtocolEdit"
                                                >
                                                    <Plus style="font-size: 20px" />
                                                    <span class="text-[12px] ml-[4px]">{{ $t('action.create') }}</span>
                                                </span>
                                            </template>
                                        </div>
                                    </template>
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
        <!-- key 绑定业务：切换业务后强制重建组件，让拓扑树按新业务范围重新拉取
             （库的拓扑树只在组件挂载时请求一次，重开弹窗不会自动刷新） -->
        <IpSelector
            :key="mainStore.selectedBusinessId.join(',')"
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
import { Done, Error, Plus } from 'bkui-vue/lib/icon';
import { debounce } from 'lodash';
import { computed, onMounted, onUnmounted, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';

import { VxeColgroup, VxeColumn, VxeTable } from '@blueking/vxe-table';

import type { NodeOpsInfo } from '@/@types/ops';
import { OpsService } from '@/api/modules/ops';
import { TopoService } from '@/api/modules/topo';
import BatchEdit from '@/components/batch-edit.vue';
import IpSelector from '@/components/IpSelector';
import * as IpSelectorService from '@/services/ip-selector';
import { useMainStore } from '@/stores/main';
import { useNodeManageStore } from '@/stores/node-manage';

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

// 协议选项列表（响应式）：预设协议来自平台配置注入，用户新建的自定义协议在会话内追加
function parseOpsOutBandProtocols(): OpsOutBandProtocol[] {
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
}
const opsOutBandProtocols = ref<OpsOutBandProtocol[]>(parseOpsOutBandProtocols());
const opsOutBandProtocolOptions = computed(() => opsOutBandProtocols.value.map(item => ({ id: item.name, name: item.name })));

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
    const protocol = opsOutBandProtocols.value.find(item => item.name === row.ops_out_band_protocol);
    // 仅预设协议带默认端口；自定义协议 port 为 0，不覆盖已填写的端口
    if (protocol && protocol.port > 0) {
        row.ops_out_band_port = protocol.port;
    }
};

// allow-create 仅回车确认，失焦时编辑单元格会被销毁导致输入丢失，这里在失焦时把输入文本确认为值
const handleProtocolFocusout = (row: any, event: FocusEvent) => {
    const text = event.target instanceof HTMLInputElement ? event.target.value.trim() : '';
    if (!text || text === row.ops_out_band_protocol) return;
    handleProtocolChange(row, text);
};

// 协议下拉"新增"入口：下拉底部固定入口，点击后输入 + 确认/取消（同标签新建交互）
const protocolSelectRef = ref();
const showProtocolEdit = ref(false);
const protocolCreateText = ref('');

// 下拉收起时重置新增编辑态，避免下次打开残留上一次的输入
const handleProtocolSelectToggle = (show: boolean) => {
    if (show) return;
    showProtocolEdit.value = false;
    protocolCreateText.value = '';
};

const handleShowProtocolEdit = () => {
    protocolCreateText.value = '';
    showProtocolEdit.value = true;
};

const handleProtocolCreateCancel = () => {
    showProtocolEdit.value = false;
    protocolCreateText.value = '';
};

const handleProtocolCreate = (row: any) => {
    const text = protocolCreateText.value.trim();
    showProtocolEdit.value = false;
    protocolCreateText.value = '';
    if (!text) return;
    // 与已有协议同名时等价于选中该选项（预设协议会自动带出默认端口）
    if (!opsOutBandProtocols.value.some(item => item.name === text)) {
        // 自定义协议无默认端口，会话内加入选项列表供本次编辑选择
        opsOutBandProtocols.value.push({ name: text, port: 0, description: '' });
    }
    handleProtocolChange(row, text);
    protocolSelectRef.value?.hidePopover?.();
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
// 主控机回显数据：host-id → IP 选择器主机对象（由带外设置中的主控机 ID 聚合查询得到）
const opsConsoleHostMap = ref(new Map<number, any>());

const handleIpSelectorPanelChange = (panel: string) => {
    IpSelectorService.setActiveIpsPanel(panel);
};

// 顶部切换业务时同步库的业务范围，保证已打开/下次打开的 IP 选择器与当前业务一致
watch(
    () => mainStore.selectedBusinessId,
    (bizIds) => {
        IpSelectorService.setBizId([...(bizIds || [])]);
    },
);

const handleOpenOpsConsoleHostSelector = (row: any) => {
    opsConsoleHostEditingRow.value = row;
    // 业务范围与顶部业务选择器保持一致：拓扑树与主机查询仅覆盖选中业务；
    // 不预设 node role 限制（主控机不要求安装 Agent）
    IpSelectorService.setStrategyBizId(undefined);
    IpSelectorService.setBizId([...(mainStore.selectedBusinessId || [])]);
    // 回显当前主控机：已聚合到主机详情时传入完整字段，选择器打开即可勾选该主机；
    // 未查到时退回只传 host-id，由选择器自行补全
    const consoleHostId = Number(row.ops_console_host_id);
    const selectedHost = Number.isSafeInteger(consoleHostId) ? opsConsoleHostMap.value.get(consoleHostId) : undefined;
    let consoleHostSelection: any[] = [];
    if (selectedHost) {
        consoleHostSelection = [selectedHost];
    } else if (consoleHostId > 0) {
        consoleHostSelection = [{ hostId: consoleHostId }];
    }
    opsConsoleHostSelectorValue.value = {
        hostList: consoleHostSelection,
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
        // 缓存本次选择，再次打开选择器时可直接回显勾选，无需再查 topo
        opsConsoleHostMap.value.set(Number(hostId), {
            hostId: Number(hostId),
            ip: host.ip || '',
            ipv6: host.ipv6 || '',
            cloudArea: host.cloudArea ?? host.cloud_area,
            meta: host.meta,
        });
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
            await OpsService.NodeOpsUpdateFieldsProxy({ hosts });
        } else {
            await OpsService.NodeOpsUpdateFieldsAgent({ hosts });
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
    try {
        let selectedHosts = params.tableData;
        if (params.isCrossPageSelection) {
            selectedHosts = [];
            const pageSize = 500;
            // 逐页拉取：单页失败时降级为已拉到的数据，避免整表空白；
            // 终止条件不依赖 total 字段，与拉取前的历史行为保持一致
            for (;;) {
                const hostListData = await TopoService.HostList({
                    page: { offset: selectedHosts.length, limit: pageSize },
                    only_count: false,
                    exact_include_conditions: params.queryParams?.exact_include_conditions || {},
                    exact_exclude_conditions: params.queryParams?.exact_exclude_conditions || {},
                    fuzzy_include_conditions: params.queryParams?.fuzzy_include_conditions || {},
                }).catch(() => ({ total: 0, items: [] }));
                if (!hostListData.items?.length) break;
                selectedHosts.push(...hostListData.items);
                if (hostListData.items.length < pageSize) break;
            }
        }
        const rows = formatHostData(selectedHosts);
        await loadCurrentOpsFields(rows);
        await loadOpsConsoleHostDetails(rows);
        tableData.value = rows;
    } catch (err: any) {
        Message({ theme: 'error', message: err?.message || String(err) });
    } finally {
        loading.value = false;
    }
});

// 读取所选主机的当前带外配置，回写到表格行
async function loadCurrentOpsFields(rows: any[]) {
    // 过滤非法 host-id，避免 NaN 以 null 形式传给后端导致整个请求被拒绝
    const hostIds = [...new Set<number>(rows
        .map(row => Number(row.bk_host_id))
        .filter(hostId => Number.isSafeInteger(hostId) && hostId > 0))];
    if (!hostIds.length) return;
    const fieldsByHost = new Map<number, NodeOpsInfo>();
    let result: { items?: NodeOpsInfo[] } | undefined;
    try {
        // 按节点类型调用对应接口，各自校验权限
        result = nodeType.value === 'proxy'
            ? await OpsService.NodeOpsFieldsListProxy({ bk_host_id: hostIds })
            : await OpsService.NodeOpsFieldsListAgent({ bk_host_id: hostIds });
    } catch {
        // 回显失败不阻塞表格展示：字段保持为空，仅丢失回显，与主控机详情降级逻辑保持一致
        return;
    }
    if (!result) return;
    for (const fields of result.items || []) {
        fieldsByHost.set(fields.bk_host_id, fields);
    }
    for (const row of rows) {
        const fields = fieldsByHost.get(Number(row.bk_host_id));
        if (!fields) continue;
        // 接口仅返回主控机 ID，IP 由 loadOpsConsoleHostDetails 单独回填
        Object.assign(row, fields);
        const protocol = fields.ops_out_band_protocol;
        if (protocol && !opsOutBandProtocols.value.some(item => item.name === protocol)) {
            opsOutBandProtocols.value.push({ name: protocol, port: 0, description: '' });
        }
    }
}

// 聚合查询主控机基础信息：一次查询覆盖所有已设置的主控机，避免逐行请求；
// 结果同时用于表格 IP 回显与选择器勾选。
async function loadOpsConsoleHostDetails(rows: any[]) {
    const hostIds = [...new Set<number>(rows
        .map(row => Number(row.ops_console_host_id))
        .filter(hostId => Number.isSafeInteger(hostId) && hostId > 0))];
    if (!hostIds.length) return;

    const hosts = new Map<number, any>();
    try {
        // 主控机不要求安装 Agent，这里不附加 node_role 过滤，避免被节点角色条件排除。
        const chunkSize = 500;
        for (let offset = 0; offset < hostIds.length; offset += chunkSize) {
            const chunk = hostIds.slice(offset, offset + chunkSize);
            const { items } = await TopoService.HostList({
                page: { offset: 0, limit: chunk.length },
                only_count: false,
                exact_include_conditions: { bk_host_id: chunk },
            });
            for (const item of items || []) {
                const mapped = IpSelectorService.mapHostItem(item);
                if (!Number.isSafeInteger(mapped.host_id) || mapped.host_id <= 0) continue;
                hosts.set(mapped.host_id, {
                    hostId: mapped.host_id,
                    ip: mapped.ip,
                    ipv6: mapped.ipv6,
                    cloudArea: mapped.cloud_area,
                    meta: mapped.meta,
                });
            }
        }
    } catch {
        // IP 仅用于回显：查询失败时保留 host-id 展示，不阻塞带外字段编辑与提交
        return;
    }

    opsConsoleHostMap.value = hosts;
    for (const row of rows) {
        const host = hosts.get(Number(row.ops_console_host_id));
        if (host) row.ops_console_host_ip = host.ip || host.ipv6 || '';
    }
}

// Format only display metadata; current Ops fields are loaded separately.
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
