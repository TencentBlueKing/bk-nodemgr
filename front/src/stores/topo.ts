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
    allWorkGraphInfos.value.forEach((info) => {
      unitProxyAgentMap.set(info.bk_networkunit_id, info);
    });

    // 5. 给基础单元数据补充 proxy 和 agent
    const unitsWithProxyAgent = baseUnits.map((unit) => {
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
    const resultMap = new Map<number, number[]>();
    const allAreaIds = new Set<number>();

    // 第一步：收集所有区域ID和接入点映射
    const accessPointToArea = new Map<number, number>(); // 接入点ID -> 区域ID

    workUnitByArea.value.forEach((unit) => {
      allAreaIds.add(unit.bk_networkarea_id);

      // 收集该单元的所有接入点对应的区域ID
      unit.accesspoints?.forEach((ap) => {
        accessPointToArea.set(ap.accesspoint_id, unit.bk_networkarea_id);
      });
    });

    // 第二步：根据单元的链接关系建立区域间连接
    const areaConnections = new Map<number, Set<number>>();
    allAreaIds.forEach((areaId) => {
      areaConnections.set(areaId, new Set([areaId])); // 每个区域都包含自身
    });

    workUnitByArea.value.forEach((unit) => {
      const currentAreaId = unit.bk_networkarea_id;

      // 检查所有类型的链接
      const linkTypes = ['cluster', 'file', 'data'] as const;
      linkTypes.forEach((type) => {
        const link = unit.links[type];
        if (link?.accesspoint_id !== undefined) {
          // 通过接入点ID找到对应的区域ID
          const linkedAreaId = accessPointToArea.get(link.accesspoint_id);
          if (linkedAreaId !== undefined && linkedAreaId !== currentAreaId) {
            // 建立双向连接关系
            areaConnections.get(currentAreaId)!.add(linkedAreaId);
            areaConnections.get(linkedAreaId)!.add(currentAreaId);
          }
        }
      });
    });

    // 第三步：为每个区域构建完整的依赖关系
    allAreaIds.forEach((areaId) => {
      const relatedAreas = new Set<number>();
      const visited = new Set<number>();

      const collectRelatedAreas = (currentAreaId: number) => {
        if (visited.has(currentAreaId)) return;
        visited.add(currentAreaId);
        relatedAreas.add(currentAreaId);

        areaConnections.get(currentAreaId)?.forEach((connectedAreaId) => {
          if (!visited.has(connectedAreaId)) {
            collectRelatedAreas(connectedAreaId);
          }
        });
      };

      collectRelatedAreas(areaId);
      resultMap.set(areaId, Array.from(relatedAreas));
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
