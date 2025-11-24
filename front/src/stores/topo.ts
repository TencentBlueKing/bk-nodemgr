import { defineStore } from 'pinia';
import { computed, ref } from 'vue';

import type { TopoNetworkAreaListReq } from '@/@types/topo';
import { TopoService } from '@/api/modules/topo';

// 定义类型（确保类型一致）
type NetworkArea = {
  bk_networkarea_id: number;
  bk_networkarea_name: string;
  // 其他区域字段...
};

type NetworkUnit = {
  bk_networkunit_id: number;
  bk_networkarea_id: number;
  bk_networkunit_name: string;
  is_direct: boolean;
  direct_endpoints: any;
  accesspoints: any[];
  links: {
    cluster?: { accesspoint_id?: number; bk_networkunit_id?: number; bk_networkarea_id?: number };
    file?: any;
    data?: any;
  };
  status?: string;
  latency?: number[];
  running_proxy: number;
  total_proxy: number;
  running_agent: number;
  total_agent: number;
  is_healthy: boolean;
  cycle_times: string[];
};

type NetworkUnitGraph = any;
type LinkGraph = any;

export const useTopoStore = defineStore('topo', () => {
  const allWorkareaList = ref<NetworkArea[]>([]);
  const workUnitByArea = ref<NetworkUnit[]>([]);
  const workareaTotalCount = ref(0);
  const allWorkGraphNodes = ref<NetworkUnitGraph[]>([]);
  const allWorkGraphEdges = ref<LinkGraph[]>([]);
  const allWorkGraphInfos = ref<{
    bk_networkunit_id: number;
    proxy: number;
    agent: number;
  }[]>([]);

  const handleFetchTopoWorkareaList = async () => {
    const params: Partial<TopoNetworkAreaListReq> = {
      page: {
        offset: 0,
        limit: 0, // limit为0即获取所有
      },
    };

    const result = await TopoService.NetworkAreaList(params);
    allWorkareaList.value = result.items;
    workareaTotalCount.value = result.total;
  };

  // 根据区域获取单元 + 补充 proxy 和 agent 数据（核心修改）
  const handleFetchAllWorkUnit = async (ids: number[] = []) => {
    // 1. 请求单元基础数据
    const unitResult = await TopoService.NetworkUnitList({
      exact_include_conditions: {
        bk_networkarea_id: ids,
      },
    }).catch(() => ({
      total: 0,
      items: [],
    }));

    const baseUnits = unitResult?.items || [];

    // 2. 获取所有单元ID，用于请求 proxy 和 agent 数据
    const allUnitIds = baseUnits.map(unit => unit.bk_networkunit_id);
    // 3. 请求 proxy 和 agent 数据（调用 handleFetchTopoWorkGraphInfo）
    await handleFetchTopoWorkGraphInfo(allUnitIds);

    // 4. 构建单元ID到 proxy/agent 的映射（方便快速查找）
    const unitProxyAgentMap = new Map<number, { proxy: number; agent: number }>();
    allWorkGraphInfos.value.forEach(info => {
      unitProxyAgentMap.set(info.bk_networkunit_id, info);
    });

    // 5. 给基础单元数据补充 proxy 和 agent
    const unitsWithProxyAgent = baseUnits.map(unit => {
      const proxyAgent = unitProxyAgentMap.get(unit.bk_networkunit_id) || { proxy: 0, agent: 0 };
      return {
        ...unit,
        ...proxyAgent,
      };
    });

    // 6. 赋值给 workUnitByArea
    workUnitByArea.value = unitsWithProxyAgent;
  };

  const handleFetchTopoWorkGraphNode = async (ids: Number[] = []) => {
    const result = await TopoService.GraphGet({
      bk_networkarea_id: ids,
    });
    allWorkGraphNodes.value = result.networkunit;
    allWorkGraphEdges.value = result.links;
  };

  const handleFetchTopoWorkGraphInfo = async (ids: Number[] = []) => {
    const result = await TopoService.TopoGraphNodeGetReq({
      bk_networkunit_id: ids,
    });
    allWorkGraphInfos.value = result.graph_node_info; // 这里返回的是包含 proxy 和 agent 的数据
  };

  const accessPointData = computed(() => {
    const accessPoints: any[] = [];
    const accessPointUnitCount = new Map(); // 统计每个接入点的下游单元数量

    // 第一次遍历：统计每个接入点被多少单元引用
    workUnitByArea.value.forEach((unit) => {
      if (unit.links && unit.links.cluster && unit.links.cluster.accesspoint_id !== null) {
        const apId = unit.links.cluster.accesspoint_id;
        accessPointUnitCount.set(apId, (accessPointUnitCount.get(apId) || 0) + 1);
      }
    });

    // 第二次遍历：构建接入点数据
    workUnitByArea.value.forEach((unit) => {
      if (unit.accesspoints && unit.accesspoints.length > 0) {
        unit.accesspoints.forEach((apInfo) => {
          const downstreamUnits = accessPointUnitCount.get(apInfo.accesspoint_id) || 0;
          accessPoints.push({
            id: `accessPoint-${apInfo.accesspoint_id}`,
            bk_accesspoint_id: apInfo.accesspoint_id,
            name: apInfo.accesspoint_name,
            bk_networkarea_id: unit.bk_networkarea_id,
            bk_networkunit_id: unit.bk_networkunit_id,
            endpoints: apInfo.endpoints || {},
            downstreamUnits,
            type: 'internal',
          });
        });
      }
    });

    return accessPoints;
  });

  const serverData = computed(() => workUnitByArea.value.filter(item => item.is_direct).map(item => ({
    id: `server-${item.bk_networkarea_id}`,
    bk_server_id: item.bk_networkarea_id,
    name: `server-${item.bk_networkarea_id}`,
    bk_networkarea_id: 0,
    endpoints: {
      cluster: item.direct_endpoints.cluster,
      file: item.direct_endpoints.file,
      data: item.direct_endpoints.data,
    },
  })));

  const areaDependencyMap = computed(() => {
    // 第一步：构建核心映射（包含所有单元，不管是否在 workUnitByArea 中）
    const unitToUpstreamUnit = new Map<number, number>(); // 单元→上游单元
    const unitToArea = new Map<number, number>(); // 单元→所属区域（关键：包含所有单元）
    const allAreaIds = new Set<number>(); // 所有区域ID
    const allUnitIds = new Set<number>(); // 所有单元ID（避免遗漏上游单元）

    // 1. 先处理 workUnitByArea 中的单元（下游单元）
    workUnitByArea.value.forEach((unit) => {
      const unitId = unit.bk_networkunit_id;
      const areaId = unit.bk_networkarea_id;

      allUnitIds.add(unitId);
      allAreaIds.add(areaId);
      unitToArea.set(unitId, areaId);

      // 提取上游单元
      const firstLink = unit.links.cluster || unit.links.file || unit.links.data;
      if (firstLink && firstLink.bk_networkunit_id) {
        const upstreamUnitId = firstLink.bk_networkunit_id;
        unitToUpstreamUnit.set(unitId, upstreamUnitId);
        allUnitIds.add(upstreamUnitId); // 收集上游单元ID
        allAreaIds.add(firstLink.bk_networkarea_id); // 收集上游单元的区域ID
      }
    });

    // 2. 补充上游单元的「单元→区域」映射（关键修复！）
    workUnitByArea.value.forEach((unit) => {
      const firstLink = unit.links.cluster || unit.links.file || unit.links.data;
      if (firstLink) {
        const upstreamUnitId = firstLink.bk_networkunit_id;
        const upstreamAreaId = firstLink.bk_networkarea_id;
        // 给上游单元绑定区域ID（即使上游单元不在 workUnitByArea 中）
        if (!unitToArea.has(upstreamUnitId) && upstreamAreaId) {
          unitToArea.set(upstreamUnitId, upstreamAreaId);
        }
      }
    });

    // 第二步：构建「区域→直接上游」和「区域→直接下游」（精准绑定）
    const areaToDirectUpstream = new Map<number, Set<number>>();
    const areaToDirectDownstream = new Map<number, Set<number>>();
    allAreaIds.forEach((areaId) => {
      areaToDirectUpstream.set(areaId, new Set());
      areaToDirectDownstream.set(areaId, new Set());
    });

    // 遍历所有单元，建立区域上下游关系
    allUnitIds.forEach((unitId) => {
      const currentAreaId = unitToArea.get(unitId);
      const upstreamUnitId = unitToUpstreamUnit.get(unitId);
      const upstreamAreaId = upstreamUnitId ? unitToArea.get(upstreamUnitId) : undefined;

      if (currentAreaId !== undefined && upstreamAreaId !== undefined && currentAreaId !== upstreamAreaId) {
        // 当前区域的直接上游 = 上游单元的区域
        areaToDirectUpstream.get(currentAreaId)!.add(upstreamAreaId);
        // 上游区域的直接下游 = 当前区域
        areaToDirectDownstream.get(upstreamAreaId)!.add(currentAreaId);
      }
    });

    // 第三步：以自身为中心，递归收集所有上下游（含间接）
    const getSelfCenteredRelations = (startAreaId: number): number[] => {
      const result = new Set<number>([startAreaId]); // 自身必含

      // 递归收集所有上游（含间接）
      const collectUpstream = (areaId: number) => {
        const directUpstreams = areaToDirectUpstream.get(areaId)!;
        directUpstreams.forEach((upAreaId) => {
          if (!result.has(upAreaId)) {
            result.add(upAreaId);
            collectUpstream(upAreaId); // 递归上游的上游
          }
        });
      };

      // 递归收集所有下游（含间接）
      const collectDownstream = (areaId: number) => {
        const directDownstreams = areaToDirectDownstream.get(areaId)!;
        directDownstreams.forEach((downAreaId) => {
          if (!result.has(downAreaId)) {
            result.add(downAreaId);
            collectDownstream(downAreaId); // 递归下游的下游
          }
        });
      };

      collectUpstream(startAreaId);
      collectDownstream(startAreaId);

      return Array.from(result);
    };

    // 第四步：生成最终Map
    const resultMap = new Map<number, number[]>();
    allAreaIds.forEach((areaId) => {
      const relations = getSelfCenteredRelations(areaId);
      resultMap.set(areaId, relations);
    });

    return resultMap;
  });

  return {
    allWorkareaList,
    workUnitByArea,
    accessPointData,
    serverData,
    areaDependencyMap,
    workareaTotalCount,
    allWorkGraphNodes,
    allWorkGraphEdges,
    allWorkGraphInfos,
    handleFetchTopoWorkareaList,
    handleFetchTopoWorkGraphNode,
    handleFetchTopoWorkGraphInfo,
    handleFetchAllWorkUnit,
  };
});
