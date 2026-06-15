import { defineStore } from 'pinia';
import { computed, ref } from 'vue';

import type { CycleTime, GraphNodeInfo, TopoNetworkAreaListReq } from '@/@types/topo';
import { TopoService } from '@/api/modules/topo';

// 定义类型（确保类型一致）
type NetworkArea = {
  bk_networkarea_id: number;
  bk_networkarea_name: string;
  // 其他区域字段...
};

type NetworkUnit = Omit<NetworkUnitBrief, 'accesspoints'> & {
  accesspoints: AccessPointBrief[];
  status?: string;
  latency?: number[];
  running_proxy: number;
  total_proxy: number;
  running_agent: number;
  total_agent: number;
  is_healthy: boolean;
  cycle_times: CycleTime[];
};

type NetworkUnitGraph = any;
type LinkGraph = any;

export const useTopoStore = defineStore('topo', () => {
  const allWorkareaList = ref<NetworkArea[]>([]);
  const workUnitByArea = ref<NetworkUnit[]>([]);
  const workareaTotalCount = ref(0);
  const allWorkGraphNodes = ref<NetworkUnitGraph[]>([]);
  const allWorkGraphEdges = ref<LinkGraph[]>([]);
  const allWorkGraphInfos = ref<GraphNodeInfo[]>([]);

  // 有权限的接入点完整信息（含 endpoints），key = accesspoint_id
  const accessPointDetailMap = ref<Map<number, AccessPoint>>(new Map());
  // 管控单元详情缓存（含 links / direct_endpoints），key = bk_networkunit_id
  const networkUnitDetailMap = ref<Map<number, NetworkUnitDetail>>(new Map());

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

  // 根据区域获取单元 + 补充 proxy 和 agent 数据
  // authorizedUnitIds: null = 全部有权限（isAny），string[] = 有权限的单元 ID 列表
  const handleFetchAllWorkUnit = async (ids: number[] = [], authorizedUnitIds: string[] | null = []) => {
    // 1. 请求单元基础数据（Brief，不含 direct_endpoints/custom_deploy_config）
    const unitResult = await TopoService.NetworkUnitListBrief({
      exact_include_conditions: {
        bk_networkarea_id: ids,
      },
    }).catch(() => ({
      total: 0,
      items: [],
    }));

    const baseUnits = unitResult?.items || [];
    // null 表示全部有权限
    const allAuthorized = authorizedUnitIds === null;
    const authorizedSet = new Set((authorizedUnitIds || []).map(Number));

    // 2. 按权限分组接入点 ID
    const authorizedApIds: number[] = [];
    const unauthorizedApIds: number[] = [];
    baseUnits.forEach((unit) => {
      (unit.accesspoints || []).forEach((apId) => {
        if (allAuthorized || authorizedSet.has(unit.bk_networkunit_id)) {
          authorizedApIds.push(apId);
        } else {
          unauthorizedApIds.push(apId);
        }
      });
    });
    const uniqueAuthorizedApIds = [...new Set(authorizedApIds)];
    const uniqueUnauthorizedApIds = [...new Set(unauthorizedApIds)].filter(id => !uniqueAuthorizedApIds.includes(id));

    // 3. 有权限的接入点用 AccessPointList（含 endpoints），无权限的用 AccessPointListBrief
    const [authorizedApResult, unauthorizedApResult] = await Promise.all([
      uniqueAuthorizedApIds.length > 0
        ? TopoService.AccessPointList({
          page: { offset: 0, limit: uniqueAuthorizedApIds.length },
          only_count: false,
          exact_include_conditions: { accesspoint_id: uniqueAuthorizedApIds },
        }).catch(() => ({ total: 0, items: [] }))
        : { total: 0, items: [] },
      uniqueUnauthorizedApIds.length > 0
        ? TopoService.AccessPointListBrief({
          page: { offset: 0, limit: uniqueUnauthorizedApIds.length },
          only_count: false,
          exact_include_conditions: { accesspoint_id: uniqueUnauthorizedApIds },
        }).catch(() => ({ total: 0, items: [] }))
        : { total: 0, items: [] },
    ]);

    // 4. 合并接入点信息（AccessPointBrief 是 AccessPoint 的子集，兼容）
    const accessPointMap = new Map<number, AccessPointBrief>();
    const newDetailMap = new Map<number, AccessPoint>();
    (authorizedApResult.items || []).forEach((item: AccessPoint) => {
      accessPointMap.set(item.accesspoint_id, item);
      newDetailMap.set(item.accesspoint_id, item);
    });
    (unauthorizedApResult.items || []).forEach((item: AccessPointBrief) => {
      accessPointMap.set(item.accesspoint_id, item);
    });
    accessPointDetailMap.value = newDetailMap;

    const detailedUnits = baseUnits.map(unit => ({
      ...unit,
      accesspoints: (unit.accesspoints || [])
        .map(accessPointID => accessPointMap.get(accessPointID))
        .filter((item): item is AccessPointBrief => !!item),
    }));

    // 2. 过滤出有权限的单元 ID，用于请求 proxy 和 agent 数据
    //    无权限的单元不调用 /graph/node/get，避免后端返回 permission denied
    const unitIdsToFetch = detailedUnits
      .map(unit => unit.bk_networkunit_id)
      .filter(id => allAuthorized || authorizedSet.has(id));
    // 3. 请求 proxy 和 agent 数据（调用 handleFetchTopoWorkGraphInfo）
    if (unitIdsToFetch.length > 0) {
      await handleFetchTopoWorkGraphInfo(unitIdsToFetch).catch(() => {});
    } else {
      allWorkGraphInfos.value = [];
    }

    // 4. 构建单元ID到 graph info 的映射（方便快速查找）
    const unitGraphInfoMap = new Map<number, any>();
    allWorkGraphInfos.value.forEach((info) => {
      unitGraphInfoMap.set(info.bk_networkunit_id, info);
    });

    // 5. 给基础单元数据补充 proxy、agent、延迟等信息
    const unitsWithProxyAgent = detailedUnits.map((unit) => {
      const graphInfo = unitGraphInfoMap.get(unit.bk_networkunit_id) || {};
      return {
        ...unit,
        running_proxy: graphInfo.running_proxy ?? 0,
        total_proxy: graphInfo.total_proxy ?? 0,
        running_agent: graphInfo.running_agent ?? 0,
        total_agent: graphInfo.total_agent ?? 0,
        is_healthy: graphInfo.is_healthy ?? true,
        cycle_times: graphInfo.cycle_times ?? [],
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

  const handleFetchNetworkUnitDetail = async (id: number) => {
    if (!id || networkUnitDetailMap.value.has(id)) {
      return networkUnitDetailMap.value.get(id) || null;
    }
    const rawRes = await TopoService.NetworkUnitGet({
      bk_networkunit_id: id,
    }).catch(() => null);

    // 适配后端可能返回 list 格式的情况（如 { items: [...] }）
    let res = rawRes;
    if (res && Array.isArray((res as any).items)) {
      const found = (res as any).items.find((u: any) => u.bk_networkunit_id === id);
      res = found || null;
    }
    if (res) {
      

      // 从 links 提取上游接入点 ID（安装策略需要这些）
      const links = (res as any).links;
      const upstreamApIds = new Set<number>();
      if (links) {
        for (const key of ['cluster', 'file', 'data'] as const) {
          const link = links[key];
          if (link?.accesspoint_id != null) {
            upstreamApIds.add(link.accesspoint_id);
          }
        }
      }

      // 查询上游接入点详情（含 endpoints），结果存为数组
      let upstreamApList: AccessPoint[] = [];
      if (upstreamApIds.size > 0) {
        const apIdArray = [...upstreamApIds];
        const apResult = await TopoService.AccessPointList({
          page: { offset: 0, limit: apIdArray.length },
          only_count: false,
          exact_include_conditions: { accesspoint_id: apIdArray },
        } as any).catch(() => ({ total: 0, items: [] }));
        upstreamApList = (apResult.items || []) as AccessPoint[];
      }

      // 挂载到 res 上，供 install-strategy 使用（数组格式，Vue 响应式友好）
      (res as any).upstreamAccessPoints = upstreamApList;

      const newMap = new Map(networkUnitDetailMap.value);
      newMap.set(id, res);
      networkUnitDetailMap.value = newMap;
    }
    return res;
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

    // 第二次遍历：构建接入点数据（合并 detail 中的 endpoints）
    workUnitByArea.value.forEach((unit) => {
      if (unit.accesspoints && unit.accesspoints.length > 0) {
        unit.accesspoints.forEach((apInfo) => {
          const downstreamUnits = accessPointUnitCount.get(apInfo.accesspoint_id) || 0;
          const detail = accessPointDetailMap.value.get(apInfo.accesspoint_id);
          accessPoints.push({
            id: `accessPoint-${apInfo.accesspoint_id}`,
            bk_accesspoint_id: apInfo.accesspoint_id,
            name: apInfo.accesspoint_name,
            bk_networkarea_id: apInfo.bk_networkarea_id || unit.bk_networkarea_id,
            bk_networkunit_id: unit.bk_networkunit_id,
            endpoints: detail?.endpoints || null,
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
  })));

  const areaDependencyMap = computed(() => {
    const resultMap = new Map<number, number[]>();
    const allAreaIds = new Set<number>();

    // 第一步：收集所有区域ID和接入点映射
    const accessPointToArea = new Map<number, number>(); // 接入点ID -> 区域ID

    workUnitByArea.value.forEach((unit) => {
      allAreaIds.add(unit.bk_networkarea_id);

      // 收集该单元提供的接入点对应的区域ID
      unit.accesspoints?.forEach((ap) => {
        accessPointToArea.set(ap.accesspoint_id, unit.bk_networkarea_id);
      });
    });

    // 第二步：建立单向上游依赖关系（当前区域 -> 上游区域）
    // 当前区域的单元通过 links 引用上游区域的接入点，所以上游区域是被引用的那个区域
    const upstreamConnections = new Map<number, Set<number>>();
    allAreaIds.forEach((areaId) => {
      upstreamConnections.set(areaId, new Set([areaId])); // 每个区域都包含自身
    });

    workUnitByArea.value.forEach((unit) => {
      const currentAreaId = unit.bk_networkarea_id;

      // 检查所有类型的链接
      const linkTypes = ['cluster', 'file', 'data'] as const;
      linkTypes.forEach((type) => {
        const link = unit.links[type];
        if (link?.accesspoint_id !== undefined) {
          // 通过接入点ID找到对应的上游区域ID
          const upstreamAreaId = accessPointToArea.get(link.accesspoint_id);
          if (upstreamAreaId !== undefined && upstreamAreaId !== currentAreaId) {
            // 只建立单向关系：当前区域依赖上游区域
            upstreamConnections.get(currentAreaId)!.add(upstreamAreaId);
          }
        }
      });
    });

    // 第三步：为每个区域递归收集所有上游区域（传递性上游）
    allAreaIds.forEach((areaId) => {
      const upstreamAreas = new Set<number>();
      const visited = new Set<number>();

      const collectUpstream = (currentAreaId: number) => {
        if (visited.has(currentAreaId)) return;
        visited.add(currentAreaId);
        upstreamAreas.add(currentAreaId);

        upstreamConnections.get(currentAreaId)?.forEach((upstreamId) => {
          if (!visited.has(upstreamId)) {
            collectUpstream(upstreamId);
          }
        });
      };

      collectUpstream(areaId);
      resultMap.set(areaId, Array.from(upstreamAreas));
    });

    return resultMap;
  });

  return {
    allWorkareaList,
    workUnitByArea,
    accessPointData,
    accessPointDetailMap,
    networkUnitDetailMap,
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
    handleFetchNetworkUnitDetail,
  };
});
