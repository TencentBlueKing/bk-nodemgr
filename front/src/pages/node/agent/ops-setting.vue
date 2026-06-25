<template>
    <div class="ops-setting pt-[24px] pb-[48px]">
        <div class="m-[24px]">
            <p class="text-[14px] text-[#63656E] mb-[16px]">{{ $t('platform.nodeMan.agentStatus.opsSettingPageDesc') }}
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
                    <VxeColgroup :title="$t('platform.nodeMan.agentStatus.opsSetting')" align="center">
                        <VxeColumn
                            field="ops_console_host_id"
                            :title="$t('platform.nodeMan.agentStatus.opsConsoleHostId')"
                            :min-width="140"
                            :edit-render="{ name: 'VxeInput' }"
                        >
                            <template #header>
                                <span class="mr-[5px]">{{ $t('platform.nodeMan.agentStatus.opsConsoleHostId') }}</span>
                                <BatchEdit :title="$t('platform.nodeMan.agentStatus.opsConsoleHostId')" type="input"
                                    @confirm="(value: string) => handleBatchEdit('ops_console_host_id', value)" />
                            </template>
                            <template #default="{ row }">
                                <span v-if="row.ops_console_host_id">{{ row.ops_console_host_id }}</span>
                                <span v-else class="cell-placeholder">{{ $t('components.installTable.inputPlaceholder') }}</span>
                            </template>
                            <template #edit="{ row }">
                                <Input v-model.trim="row.ops_console_host_id" />
                            </template>
                        </VxeColumn>
                        <VxeColumn
                            field="ops_out_band_type"
                            :title="$t('platform.nodeMan.agentStatus.opsOutBandType')"
                            :min-width="140"
                            :edit-render="{ name: 'VxeInput' }"
                        >
                            <template #header>
                                <span class="mr-[5px]">{{ $t('platform.nodeMan.agentStatus.opsOutBandType') }}</span>
                                <BatchEdit :title="$t('platform.nodeMan.agentStatus.opsOutBandType')" type="input"
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
                            :title="$t('platform.nodeMan.agentStatus.opsOutBandProtocol')"
                            :min-width="140"
                            :edit-render="{ name: 'VxeInput' }"
                        >
                            <template #header>
                                <span class="mr-[5px]">{{ $t('platform.nodeMan.agentStatus.opsOutBandProtocol') }}</span>
                                <BatchEdit :title="$t('platform.nodeMan.agentStatus.opsOutBandProtocol')" type="input"
                                    @confirm="(value: string) => handleBatchEdit('ops_out_band_protocol', value)" />
                            </template>
                            <template #default="{ row }">
                                <span v-if="row.ops_out_band_protocol">{{ row.ops_out_band_protocol }}</span>
                                <span v-else class="cell-placeholder">{{ $t('components.installTable.inputPlaceholder') }}</span>
                            </template>
                            <template #edit="{ row }">
                                <Input v-model.trim="row.ops_out_band_protocol" />
                            </template>
                        </VxeColumn>
                        <VxeColumn
                            field="ops_bmc_ip"
                            :title="$t('platform.nodeMan.agentStatus.opsBMCIP')"
                            :min-width="140"
                            :edit-render="{ name: 'VxeInput' }"
                        >
                            <template #header>
                                <span class="mr-[5px]">{{ $t('platform.nodeMan.agentStatus.opsBMCIP') }}</span>
                                <BatchEdit :title="$t('platform.nodeMan.agentStatus.opsBMCIP')" type="input"
                                    @confirm="(value: string) => handleBatchEdit('ops_bmc_ip', value)" />
                            </template>
                            <template #default="{ row }">
                                <span v-if="row.ops_bmc_ip">{{ row.ops_bmc_ip }}</span>
                                <span v-else class="cell-placeholder">{{ $t('components.installTable.inputPlaceholder') }}</span>
                            </template>
                            <template #edit="{ row }">
                                <Input v-model.trim="row.ops_bmc_ip" />
                            </template>
                        </VxeColumn>
                        <VxeColumn
                            field="ops_bmc_port"
                            :title="$t('platform.nodeMan.agentStatus.opsBMCPort')"
                            :min-width="140"
                            :edit-render="{ name: 'VxeInput' }"
                        >
                            <template #header>
                                <span class="mr-[5px]">{{ $t('platform.nodeMan.agentStatus.opsBMCPort') }}</span>
                                <BatchEdit :title="$t('platform.nodeMan.agentStatus.opsBMCPort')" type="input"
                                    @confirm="(value: string) => handleBatchEdit('ops_bmc_port', value)" />
                            </template>
                            <template #default="{ row }">
                                <span v-if="row.ops_bmc_port">{{ row.ops_bmc_port }}</span>
                                <span v-else class="cell-placeholder">{{ $t('components.installTable.inputPlaceholder') }}</span>
                            </template>
                            <template #edit="{ row }">
                                <Input v-model.trim="row.ops_bmc_port" />
                            </template>
                        </VxeColumn>
                    </VxeColgroup>
                </VxeTable>
            </Loading>
        </div>

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
import { Button, Input, Loading, Message, Select } from 'bkui-vue';
import { debounce } from 'lodash';
import { computed, onMounted, onUnmounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';

import { VxeColgroup, VxeColumn, VxeTable } from '@blueking/vxe-table';

import BatchEdit from '@/components/batch-edit.vue';
import { NodeAgentService } from '@/api/modules/node_agent';
import { NodeProxyService } from '@/api/modules/node_proxy';
import { TopoService } from '@/api/modules/topo';
import { useNodeManageStore } from '@/stores/node-manage';
import { useMainStore } from '@/stores/main';

const { t } = useI18n();
const router = useRouter();
const nodeManageStore = useNodeManageStore();
const mainStore = useMainStore();

const businessList = computed(() => mainStore.businessList);

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
                ops_bmc_ip: String(row.ops_bmc_ip ?? ''),
                ops_bmc_port: Number(row.ops_bmc_port) || 0,
            }));

        if (nodeType.value === 'proxy') {
            await NodeProxyService.NodeProxyUpdateOpsFields({ hosts });
        } else {
            await NodeAgentService.NodeAgentUpdateOpsFields({ hosts });
        }

        Message({ theme: 'success', message: t('platform.nodeMan.agentStatus.opsSettingSuccess') });
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
            ops_out_band_type: item.ops_out_band_type ?? info.ops_out_band_type ?? undefined,
            ops_out_band_protocol: item.ops_out_band_protocol ?? info.ops_out_band_protocol ?? undefined,
            ops_bmc_ip: item.ops_bmc_ip ?? info.ops_bmc_ip ?? undefined,
            ops_bmc_port: item.ops_bmc_port ?? info.ops_bmc_port ?? undefined,
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
</style>
