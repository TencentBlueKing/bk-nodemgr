<template>
  <PageHeader
    class="w-full absolute top-0 z-100"
    :title="$t('route.opsHistoryDetail')"
    :back="true"
    :on-back="handleBack"
  >
    <span class="mx-[6px] text-[#979BA5] text-[14px]">-</span>
    <span class="text-[#979BA5] text-[14px] mr-[14px]">{{ operateId }}</span>
  </PageHeader>
  <div class="p-[24px] mt-[52px]">
    <!-- 操作批次信息（从列表行带入，刷新后显示 --） -->
    <div class="flex">
      <div
        v-for="item in infoList"
        :key="item.prop"
        class="leading-[30px] mr-[52px] text-[12px]"
      >
        <div>{{ item.name }}</div>
        <UserNameDisplay v-if="item.prop === 'operator'" :name="item.value || ''" />
        <div v-else>{{ item.value }}</div>
      </div>
    </div>
    <!-- 主机设置快照明细 -->
    <div class="relative mt-[24px]">
      <bk-loading :loading="loading" :title="$t('table.loading')">
        <Table
          class="filterTable"
          :data="tableData"
          :empty-text="$t('table.empty')"
          show-overflow-tooltip
          :max-height="maxHeight"
          :column-config="{ resizable: true }"
        >
          <TableColumn
            field="host_id"
            :title="$t('hostID')"
            min-width="120"
            fixed="left"
          />
          <TableColumn
            field="console_host"
            :title="$t('platform.nodeMan.agentNodeStatus.opsConsoleHostId')"
            min-width="140"
          />
          <TableColumn
            field="out_band_type"
            :title="$t('platform.nodeMan.agentNodeStatus.opsOutBandType')"
            min-width="120"
          />
          <TableColumn
            field="out_band_protocol"
            :title="$t('platform.nodeMan.agentNodeStatus.opsOutBandProtocol')"
            min-width="120"
          />
          <TableColumn
            field="out_band_ip"
            :title="$t('platform.nodeMan.agentNodeStatus.opsOutBandIP')"
            min-width="140"
          />
          <TableColumn
            field="out_band_port"
            :title="$t('platform.nodeMan.agentNodeStatus.opsOutBandPort')"
            min-width="100"
          />
        </Table>
      </bk-loading>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { computed, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';

import { Table, TableColumn } from '@blueking/table';

import type { NodeOpsHistoryHostSnapshot } from '@/@types/ops';
import { OpsService } from '@/api/modules/ops';
import { formatTimeByTimezone } from '@/common/util';
import { useMainStore } from '@/stores/main';
import { useNodeManageStore } from '@/stores/node-manage';

const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const nodeManageStore = useNodeManageStore();
const mainStore = useMainStore();

const operateId = computed(() => String(route.params.operateId || ''));
const loading = ref(false);
const snapshots = ref<NodeOpsHistoryHostSnapshot[]>([]);
const maxHeight = computed(() => mainStore.windowInnerHeight - 220);

// 行数据从列表页带入；直接刷新详情页时取不到，显示 --
const rowData = computed(() => nodeManageStore.taskHistoryTableRowData as any);
const infoList = computed(() => [
  {
    name: t('platform.nodeMan.taskHistory.label.operator'),
    prop: 'operator',
    value: rowData.value?.operator || '--',
  },
  {
    name: t('platform.nodeMan.taskHistory.label.operateTime'),
    prop: 'operate_time',
    value: rowData.value?.operate_time
      ? formatTimeByTimezone(rowData.value.operate_time, 'YYYY-MM-DD HH:mm:ss')
      : '--',
  },
  {
    name: t('platform.nodeMan.taskHistory.label.nodeRole'),
    prop: 'node_role',
    value: rowData.value?.node_role || '--',
  },
]);

const joinList = (list?: string[]) => (list && list.length > 0 ? list.join(', ') : '--');

const tableData = computed(() => snapshots.value.map(item => ({
  ...item,
  host_id: item.bk_host_id || '--',
  console_host: joinList(item.ops_console_host_innerip_list),
  out_band_type: item.ops_out_band_type || '--',
  out_band_protocol: item.ops_out_band_protocol || '--',
  out_band_ip: item.ops_out_band_ip || '--',
  out_band_port: item.ops_out_band_port || '--',
})));

const fetchDetail = async () => {
  if (!operateId.value) return;
  loading.value = true;
  const res = await OpsService.NodeOpsHistoryDetail({
    operate_id: operateId.value,
  }).catch((err) => {
    console.log(err);
    return { total: 0, items: [] };
  });
  snapshots.value = res.items || [];
  loading.value = false;
};
fetchDetail();

const handleBack = () => {
  router.push({ name: 'history', query: { active: 'ops' } });
};
</script>
