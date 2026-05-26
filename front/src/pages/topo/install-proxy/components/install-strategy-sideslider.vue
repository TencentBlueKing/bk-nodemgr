<template>
  <Sideslider
    v-model:is-show="isShow"
    :title="$t('topoManager.installProxy.strategyTitle')"
    width="900"
    render-directive="if"
  >
    <div class="py-[20px] px-[40px]">
      <!-- 提示信息 -->
      <div class="text-[12px] leading-[20px] text-[#4d4f56] mb-[24px]">
        <p>{{ $t('topoManager.installProxy.strategyTip1') }}</p>
        <p>{{ $t('topoManager.installProxy.strategyTip2') }}</p>
        <p>{{ $t('topoManager.installProxy.strategyTip3', { ip: '<IP>', port: '<PORT>' }) }}</p>
      </div>

      <!-- 区域可折叠标题 -->
      <div
        v-if="areaName"
        class="flex items-center h-[40px] bg-[#EAEBF0] px-[12px] cursor-pointer select-none rounded-[2px]"
        @click="toggleAreaCollapse"
      >
        <i
          class="nodeman-icon text-[12px] text-[#63656E] mr-[8px]"
          :class="isAreaCollapsed ? 'nc-angle-right' : 'nc-angle-down'"
        ></i>
        <span class="text-[14px] text-[#313238] font-medium">{{ areaName }}</span>
      </div>

      <Table
        v-show="!isAreaCollapsed"
        :data="strategyTableData"
        :border="true"
        :empty-text="$t('table.empty')"
      >
          <TableColumn
            field="source"
            :title="$t('topoManager.installProxy.sourceAddress')"
            min-width="200"
          ></TableColumn>
          <TableColumn
            field="target"
            :title="$t('topoManager.installProxy.targetAddress')"
            min-width="200"
          ></TableColumn>
          <TableColumn
            field="port"
            :title="$t('topoManager.installProxy.targetPort')"
            width="100"
          ></TableColumn>
          <TableColumn
            field="protocol"
            :title="$t('topoManager.installProxy.protocol')"
            width="80"
          ></TableColumn>
          <TableColumn
            field="purpose"
            :title="$t('topoManager.installProxy.purpose')"
            width="120"
          ></TableColumn>
        </Table>
    </div>
  </Sideslider>
</template>

<script setup lang="ts">
import { Sideslider } from 'bkui-vue';
import type { PropType } from 'vue';
import { computed, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';

import { Table, TableColumn } from '@blueking/table';

import { useTopoStore } from '@/stores/topo';
import { TopoService } from '@/api/modules/topo';

const isShow = defineModel<boolean>('isShow', { default: false });

const props = defineProps({
  tableData: {
    type: Array as PropType<Array<Record<string, any>>>,
    default: () => [],
  },
  networkUnitId: {
    type: [String, Number],
    default: '',
  },
  areaName: {
    type: String,
    default: '',
  },
});

const { t } = useI18n();
const topoStore = useTopoStore();

// 存储上游接入点信息（accesspoint_id -> AccessPoint）
const upstreamAccessPoints = ref<Map<number, AccessPoint>>(new Map());

// 从 store 缓存读取管控单元详情
const unitDetail = computed<NetworkUnitDetail | null>(() => {
  const id = Number(props.networkUnitId);
  const detail = id ? topoStore.networkUnitDetailMap.get(id) || null : null;
  return detail;
});

// 区域折叠状态
const isAreaCollapsed = ref(false);
const toggleAreaCollapse = () => {
  isAreaCollapsed.value = !isAreaCollapsed.value;
};

// 策略配置定义（端口从接入点数据动态获取）
const strategyConfigs = [
  {
    key: 'cluster' as const,
    sourcePrefix: 'Proxy(gse_agent)',
    targetPrefix: 'GSE_task',
    protocol: 'TCP',
    purpose: t('topoManager.installProxy.purposeCluster'),
  },
  {
    key: 'file' as const,
    sourcePrefix: 'Proxy(gse_file)',
    targetPrefix: 'GSE_file',
    protocol: 'TCP',
    purpose: t('topoManager.installProxy.purposeFile'),
  },
  {
    key: 'data' as const,
    sourcePrefix: 'Proxy(gse_data)',
    targetPrefix: 'GSE_data',
    protocol: 'TCP',
    purpose: t('topoManager.installProxy.purposeData'),
  },
];

// 解析 endpoint 字符串（格式："ip:port"）
const parseEndpoint = (endpoint: string): { ip: string; port: number } => {
  const parts = endpoint.split(':');
  return {
    ip: parts[0],
    port: parts.length > 1 ? Number(parts[1]) : 0,
  };
};

// 加载所有需要的接入点信息（只调用一次接口）
const loadAccessPoints = async (): Promise<Map<number, AccessPoint>> => {
  if (!unitDetail.value) return new Map();
  const links = unitDetail.value.links;
  const isDirect = unitDetail.value.is_direct as boolean;
  if (isDirect) return new Map();

  // 收集所有需要的 accesspoint_id
  const apIds = new Set<number>();
  for (const key of ['cluster', 'file', 'data'] as const) {
    const link = links?.[key];
    if (link?.accesspoint_id != null) {
      apIds.add(link.accesspoint_id);
    }
  }
  if (apIds.size === 0) return new Map();

  // 过滤掉已缓存的
  const uncachedIds = [...apIds].filter(id => !upstreamAccessPoints.value.has(id));
  if (uncachedIds.length > 0) {
    const apResult = await TopoService.AccessPointList({
      page: { offset: 0, limit: uncachedIds.length },
      only_count: false,
      exact_include_conditions: { accesspoint_id: uncachedIds },
    }).catch(() => null);
    if (apResult?.items?.length) {
      for (const ap of apResult.items as AccessPoint[]) {
        upstreamAccessPoints.value.set(ap.accesspoint_id, ap);
      }
      // 触发响应式更新
      upstreamAccessPoints.value = new Map(upstreamAccessPoints.value);
    }
  }
  return upstreamAccessPoints.value;
};

// 根据 link 类型获取 upstream IP 和端口列表
const getEndpoints = (key: 'cluster' | 'file' | 'data'): Array<{ ip: string; port: number }> => {
  if (!unitDetail.value) {
    return [];
  }
  const links = unitDetail.value.links;
  const isDirect = unitDetail.value.is_direct as boolean;
  // 直连单元使用 direct_endpoints
  if (isDirect) {
    const endpoints = (unitDetail.value as any).direct_endpoints?.[key] || [];
    return endpoints.map(parseEndpoint);
  }
  // 非直连单元通过 links 找到上游接入点
  const link = links?.[key];
  if (link?.accesspoint_id == null) {
    return [];
  }
  const ap = upstreamAccessPoints.value.get(link.accesspoint_id);
  if (ap?.endpoints?.[key]?.length) {
    return ap.endpoints[key].map(parseEndpoint);
  }
  return [];
};

// 策略表格数据（使用 ref 而非 computed，因为需要异步加载接入点）
const strategyTableData = ref<Array<{
  source: string;
  target: string;
  port: number;
  protocol: string;
  purpose: string;
}>>([]);

// 加载策略表格数据
const loadStrategyTableData = async () => {
  // 先一次性加载所有接入点信息
  await loadAccessPoints();

  const rows: typeof strategyTableData.value = [];
  // 按 export_ip 去重，相同的出口 IP 只展示一次策略
  const seenExportIps = new Set<string>();
  const proxies = props.tableData.filter((item: any) => {
    if (!item.export_ip) return false;
    if (seenExportIps.has(item.export_ip)) return false;
    seenExportIps.add(item.export_ip);
    return true;
  });

  for (const proxy of proxies) {
    const exportIp = proxy.export_ip;
    for (const config of strategyConfigs) {
      const endpoints = getEndpoints(config.key);
      if (endpoints.length > 0) {
        endpoints.forEach(({ ip, port }) => {
          rows.push({
            source: `${config.sourcePrefix}: ${exportIp}`,
            target: `${config.targetPrefix}: ${ip}`,
            port,
            protocol: config.protocol,
            purpose: config.purpose,
          });
        });
      } else {
        // 没有 upstream IP 时也显示一条
        rows.push({
          source: `${config.sourcePrefix}: ${exportIp}`,
          target: `${config.targetPrefix}: -`,
          port: 0,
          protocol: config.protocol,
          purpose: config.purpose,
        });
      }
    }
  }

  strategyTableData.value = rows;
};

watch(
  () => [isShow.value, props.networkUnitId, props.tableData],
  async ([show, unitId]) => {
    // 兜底：如果 store 中没有该 unit 详情，自己触发加载
    const id = Number(unitId);
    if (show && id && !topoStore.networkUnitDetailMap.has(id)) {
      await topoStore.handleFetchNetworkUnitDetail(id);
    }
    // 打开时默认展开
    if (show) {
      isAreaCollapsed.value = false;
    }
    // 加载策略表格数据
    if (show && unitDetail.value) {
      await loadStrategyTableData();
    } else if (!show) {
      strategyTableData.value = [];
    }
  },
  { immediate: true, deep: true },
);
</script>
