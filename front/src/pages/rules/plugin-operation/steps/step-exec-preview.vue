<template>
  <div class="step-exec-preview bg-[#fff] p-[24px] rounded-[2px]">
    <!-- 顶部操作栏 -->

    <!-- Tab -->
    <Tab v-model:active="activeTab" type="unborder-card" :label-height="41" class="preview-tab">
      <Tab.TabPanel name="all" :label="allTabLabel">
        <div class="mt-[4px]">
          <Table
            :data="displayData"
            :empty-text="$t('table.empty')"
            :pagination="pagination"
            :max-height="tableMaxHeight"
            @page-limit-change="pageLimitChange"
            @page-value-change="pageValueChange"
          >
            <TableColumn
              field="ip"
              :title="$t('pluginOperation.preview.hostIP')"
              :min-width="140"
              fixed="left"
            />
            <TableColumn
              field="business"
              :title="$t('pluginOperation.preview.business')"
              :min-width="110"
            />
            <TableColumn
              field="networkArea"
              :title="$t('pluginOperation.preview.networkArea')"
              :min-width="110"
            />
            <TableColumn
              field="networkUnit"
              :title="$t('pluginOperation.preview.networkUnit')"
              :min-width="140"
            />
            <TableColumn
              field="opType"
              :title="$t('pluginOperation.preview.opType')"
              :min-width="100"
            />
            <TableColumn
              field="osType"
              :title="$t('pluginOperation.preview.osType')"
              :min-width="100"
            />
            <TableColumn
              field="agentStatus"
              :title="$t('pluginOperation.preview.agentStatus')"
              :min-width="120"
            >
              <template #default="{ row }">
                <div class="flex items-center">
                  <i :class="getStatusIconClass(row.agentStatus)"></i>
                  <span>{{ row.agentStatusLabel }}</span>
                </div>
              </template>
            </TableColumn>
            <TableColumn
              v-if="showTargetVersion"
              field="targetVersion"
              :title="$t('pluginOperation.preview.targetVersion')"
              :min-width="120"
            >
              <template #default="{ row }">
                <span>{{ row.targetVersion }}</span>
              </template>
            </TableColumn>
          </Table>
        </div>
      </Tab.TabPanel>
    </Tab>
  </div>
</template>

<script lang="ts" setup>
import { Tab } from 'bkui-vue';
import { computed, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';

import { Table, TableColumn } from '@blueking/table';

import { TopoService } from '@/api/modules/topo';
import { useMainStore } from '@/stores/main';

const { t } = useI18n();
const mainStore = useMainStore();

const props = defineProps<{
  formData: {
    strategyName: string;
    selectedHosts: any[];
    selectedVersion: string;
    paramConfig: Record<string, any>;
  };
  operationType?: string;
}>();

const activeTab = ref('all');


const pagination = reactive({ count: 0, limit: 50, current: 1, remote: true });
const tableMaxHeight = computed(() => mainStore.windowInnerHeight - 380 - (mainStore.noticeShow ? 40 : 0));

// 操作类型标签映射
const opTypeLabel = computed(() => {
  const map: Record<string, string> = {
    install: t('pluginManagement.plugin.operate.install'),
    upgrade: t('pluginManagement.plugin.operate.upgrade'),
    reload: t('pluginManagement.plugin.operate.reload'),
    restart: t('pluginManagement.plugin.operate.restart'),
    stop: t('pluginManagement.plugin.operate.stop'),
    reinstall: t('pluginManagement.plugin.operate.reinstall'),
  };
  return map[props.operationType || ''] || t('pluginManagement.plugin.operate.install');
});

// 重启/停止操作不需要显示目标版本列
const showTargetVersion = computed(() => !['restart', 'stop'].includes(props.operationType || ''));

// ===== 管控单元 id→name Map（学习 agent 列表 distinct + brief 模式）=====
const networkUnitMap = ref(new Map<number, string>());

const fetchNetworkUnitNames = async (ids: number[]) => {
  // -1 = 未分配（不用查），只查 >=0 的有效 ID
  const uniqueIds = [...new Set(ids.filter(id => id != null && id !== -1))];
  if (uniqueIds.length === 0) return;
  const res = await TopoService.NetworkUnitListBrief({
    exact_include_conditions: { bk_networkunit_id: uniqueIds },
  }).catch(() => ({ total: 0, items: [] }));
  const map = new Map<number, string>();
  (res.items || []).forEach((item: any) => {
    map.set(item.bk_networkunit_id, item.bk_networkunit_name);
  });
  networkUnitMap.value = map;
};

watch(() => props.formData.selectedHosts, (hosts) => {
  const ids = (hosts || []).map((h: any) => h.bk_networkunit_id ?? 0);
  fetchNetworkUnitNames(ids);
}, { immediate: true, deep: false });

// 预览数据 — 基于选中的主机生成
const previewData = computed(() => props.formData.selectedHosts.map((host: any) => {
    const rawAlive = host.alive;
    // 0=离线, 1=在线, undefined/null 视为未知
    const agentStatus = (rawAlive === 0 || rawAlive === 1) ? rawAlive : 2;
    const bizId = host.bk_biz_id ?? host.meta?.bk_biz_id ?? 0;
    const bizName = bizId ? (mainStore.businessList.find(b => b.bk_biz_id === bizId)?.bk_biz_name || `[${bizId}]`) : '—';
    const networkUnitId = host.bk_networkunit_id ?? host.info?.bk_networkunit_id;
    return {
      ip: host.ip || host.bk_host_innerip || host.bk_host_innerip_list?.[0] || '—',
      business: bizName,
      networkArea: host.cloud_area?.name || host.cloudArea?.name || t('pluginOperation.preview.directArea'),
      networkUnit: networkUnitId === -1
        ? t('pluginOperation.preview.unassigned')
        : networkUnitId != null
          ? (networkUnitMap.value.get(networkUnitId) || `[${networkUnitId}]`)
          : '—',
      opType: opTypeLabel.value,
      osType: host.os_name || host.os_type || host.osName || host.osType || '—',
      agentStatus,
      agentStatusLabel: rawAlive === 1 ? t('pluginOperation.preview.normal') : rawAlive === 0 ? t('pluginOperation.preview.abnormal') : t('pluginOperation.preview.unknownStatus'),
      targetVersion: props.formData.selectedVersion || '—',
    };
  }));

const displayData = computed(() => {
  const data = previewData.value;
  pagination.count = data.length;
  const start = (pagination.current - 1) * pagination.limit;
  return data.slice(start, start + pagination.limit);
});

const allTabLabel = computed(() => `${t('pluginOperation.preview.allTab')} (${previewData.value.length})`);

const getStatusIconClass = (status: number) => {
  switch (status) {
    case 1: return 'nodeman-icon nc-running status-icon';
    case 0: return 'nodeman-icon nc-terminated status-icon';
    case 2:
    default: return 'nodeman-icon nc-unknown status-icon';
  }
};

const pageLimitChange = (limit: number) => {
  pagination.limit = limit;
  pagination.current = 1;
};

const pageValueChange = (current: number) => {
  pagination.current = current;
};
</script>

<style lang="postcss" scoped>
.preview-tab {
  :deep(.bk-tab-header) {
    background: #fff;
  }
}

/* Agent 状态图标 — 与 detail-table.vue / process-sideslider.vue 统一 */
.status-icon::before {
  content: "";
  display: inline-block;
  margin-right: 8px;
  width: 14px;
  height: 14px;
  border: 3px solid #f0f1f5;
  border-radius: 6.5px;
  background: #b2b5bd;
  flex-shrink: 0;
  vertical-align: middle;
}
.nc-running::before {
  background: #3fc06d;
  border-color: #e5f6ea;
}
.nc-terminated::before {
  border-color: #ffe6e6;
  background: #ea3636;
}
.nc-unknown::before {
  border-color: #f0f1f5;
  background: #b2b5bd;
}
</style>
