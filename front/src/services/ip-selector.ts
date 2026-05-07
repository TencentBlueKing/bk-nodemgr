/**
 * IP选择器专用服务 - 对接后端真实接口
 * 数据源对接 TopoService 中的主机相关 API
 *
 * 支持两种使用场景：
 * 1. 策略管理：通过 setType / setBizId 限定策略类型和业务，
 *    自动按 node_role 过滤（agent→['agent','blank']，proxy→['proxy']）
 * 2. 插件手动拓扑：不设置策略参数，走通用查询路径
 */

import { TopoService } from '@/api/modules/topo';
import { useMainStore } from '@/stores/main';
import { useAuthStore } from '@/stores/auth';

// 当前策略类型（由外部设置，影响 node_role 查询条件）
// 支持 config_policy_agent / config_policy_proxy / agent / proxy 等
// 判断时使用 type.includes('agent') 区分 agent 和 proxy
// 默认空字符串 = 不过滤 node_role（插件安装场景）
export let currentPolicyType = '';

/** 设置当前策略类型（打开 IP 选择器前调用） */
export const setPolicyType = (type: string) => {
  currentPolicyType = type;
};

/** 统一别名：设置类型 */
export const setType = setPolicyType;

// 当前策略业务 ID 数组（由外部设置，从 store 或 localStorage 获取）
let currentBizIds: number[] = [];

/** 设置当前业务 ID 数组（插件安装等场景使用）
 *  调用时会重置 currentPolicyType 为空，确保走权限自适应分支而非残留的 type 值 */
export const setBizId = (bizIds: number[]) => {
  currentPolicyType = '';  // 重置：避免策略管理页面的残留 type 污染
  currentBizIds = bizIds;
};

// 当前策略业务 ID（由外部设置，策略管理页面业务固定，拓扑树只展示当前业务）
// 兼容旧接口，内部复用 currentBizIds[0]
let currentStrategyBizId: number | undefined;

/** 设置当前策略业务 ID（策略管理页面打开 IP 选择器前调用）【兼容旧接口】 */
export const setStrategyBizId = (bizId: number | undefined) => {
  currentStrategyBizId = bizId;
  if (bizId !== undefined && !currentBizIds.includes(bizId)) {
    currentBizIds.push(bizId);
  }
};

/** 获取当前有效的业务 ID 数组 */
const getEffectiveBizIds = (): number[] => {
  return currentBizIds.length > 0 ? currentBizIds : (currentStrategyBizId ? [currentStrategyBizId] : []);
};

/**
 * 根据 currentPolicyType 和权限自动决定 node_role 过滤值
 *
 * 场景说明：
 * 1. 有显式 setType → 按 type.includes('proxy') 判断（策略管理 / 插件进程安装）
 * 2. 无 setType（插件安装按钮进入）→ 根据 authorized 权限自适应：
 *    - agent_view 有权 + proxy_view 无权 → ['agent', 'blank']
 *    - proxy_view 有权 + agent_view 无权 → ['proxy']
 *    - 都有权 → [] 不过滤
 *    - 都无权或权限未加载 → ['agent', 'blank']（保守过滤，避免暴露无权限的 proxy）
 */
const resolveNodeRoleFilter = (): string[] => {
  // 显式设置了 type → 按原有逻辑判断
  if (currentPolicyType) {
    return currentPolicyType.includes('proxy') ? ['proxy'] : ['agent', 'blank'];
  }

  // 无 type → 插件安装场景，根据 authorized 权限自适应过滤，避免无权 action 导致 403
  const authStore = useAuthStore();
  const bizIds = getEffectiveBizIds();
  if (bizIds.length === 0) return [];

  // 权限数据尚未加载完成（如页面刷新后的初始请求）→ 保守过滤：只显示 agent/blank
  // 避免 authorizedMap 为空时两个 hasAuthorizedResource 都返回 false，导致不过滤而暴露无权 proxy
  if (!authStore.authorizedLoaded) {
    return ['agent', 'blank'];
  }

  let hasAgentView = false;
  let hasProxyView = false;
  for (const bizId of bizIds) {
    if (!hasAgentView && authStore.hasAuthorizedResource('agent_view', bizId)) hasAgentView = true;
    if (!hasProxyView && authStore.hasAuthorizedResource('proxy_view', bizId)) hasProxyView = true;
    if (hasAgentView && hasProxyView) break;
  }

  if (hasAgentView && !hasProxyView) return ['agent', 'blank'];
  if (hasProxyView && !hasAgentView) return ['proxy'];
  // 都有权 → [] 不过滤；都无权 → 保守过滤，避免暴露无权限节点
  if (!hasAgentView && !hasProxyView) return ['agent', 'blank'];
  return [];
};

// 类型定义
export interface IMeta {
  bk_biz_id: number;
  scope_id: string;
  scope_type: 'biz';
}

export interface INode {
  instance_id: number;
  object_id: 'module' | 'set' | 'biz';
  meta: IMeta;
}

export interface IHost {
  host_id: number;
  meta: IMeta;
}

export interface ITreeItem extends INode {
  count: number;
  expanded?: boolean;
  object_name: string;
  instance_name: string;
  child: ITreeItem[];
  lazy?: boolean;
}

export interface IQuery {
  start?: number;
  page_size?: number;
  search_content?: string;
  node_list: INode[];
  all_scope?: boolean;
}

export interface IStatistics {
  alive_count: number;
  not_alive_count: number;
  total_count: number;
}

// ======== Host/List 请求构建辅助 ========

const buildHostListParams = (params?: Partial<{
  page: { limit: number; offset: number };
  exact: any;
  fuzzy: any;
}>) => ({
  page: params?.page || { limit: 20, offset: 0 },
  only_count: false,
  exact_include_conditions: {
    bk_host_id: [],
    bk_biz_id: [],
    bk_networkarea_id: [],
    os_type: [],
    node_role: [],
    node_status: [],
    node_version: [],
    bk_agent_id: [],
    bk_networkunit_id: [],
    node_generation: [],
    ...(params?.exact || {}),
  },
  fuzzy_include_conditions: {
    bk_host_name: [],
    dept_name: [],
    bk_host_innerip: [],
    bk_host_innerip_v6: [],
    bk_host_outerip: [],
    bk_host_outerip_v6: [],
    ...(params?.fuzzy || {}),
  },
});

/** 解析 "cloudId:ip" 格式输入 */
const parseCloudInput = (input: string) => {
  const trimmed = (input || '').trim();
  if (!trimmed) return { cloudId: undefined, value: '' };
  const match = trimmed.match(/^([0-9]+):(.*)$/);
  if (match) {
    return { cloudId: Number(match[1]), value: match[2] };
  }
  return { cloudId: undefined, value: trimmed };
};

/** 去除 IPv6 方括号 */
const normalizeIpValue = (value: string) => value.replace(/^\[(.*)\]$/, '$1');

/** 将后端状态值映射为 IP 选择器 alive 值 (0=离线, 1=在线) */
const mapAgentAlive = (status: any): number => {
  if (typeof status === 'number') return status;
  if (!status) return 0;
  const normalized = String(status).toUpperCase();
  return ['RUNNING', 'ALIVE', 'HEALTHY', 'OK'].includes(normalized) ? 1 : 0;
};

/** 将后端 host item 映射为 IP 选择器标准 Host 格式 */
const mapHostItem = (item: any) => {
  const info = item?.info || {};
  const state = item?.state || {};
  const bkHostId = item?.bk_host_id ?? item?.host_id ?? info?.bk_host_id;
  const innerIp = Array.isArray(info?.bk_host_innerip_list)
    ? info.bk_host_innerip_list[0]
    : info?.bk_host_innerip;
  const innerIpv6 = Array.isArray(info?.bk_host_innerip_v6_list)
    ? info.bk_host_innerip_v6_list[0]
    : info?.bk_host_innerip_v6;
  const bkBizId = info?.bk_biz_id ?? 0;
  const networkAreaId = info?.bk_networkarea_id ?? 0;
  const agentAlive = mapAgentAlive(state?.node_status);

  return {
    id: String(bkHostId),
    host_id: bkHostId,
    ip: innerIp || '',
    ipv6: innerIpv6 || '',
    host_name: info?.bk_host_name || '',
    os_name: info?.os_type || '',
    os_type: info?.os_type || '',
    alive: agentAlive,
    cloud_area: { id: networkAreaId, name: info?.bk_networkarea_name || '' },
    cloud_id: networkAreaId,
    biz: { id: bkBizId, name: '' },
    agent_id: state?.bk_agent_id || '',
    meta: { scope_type: 'biz' as const, scope_id: String(bkBizId || ''), bk_biz_id: bkBizId },
    bk_host_id: bkHostId,
    bk_biz_id: bkBizId,
    bk_cloud_id: networkAreaId,
    bk_agent_id: state?.bk_agent_id || '',
    bk_agent_alive: agentAlive,
  };
};

/** 获取默认业务 ID（优先 mainStore，兼容读取 localStorage） */
const getDefaultBizId = () => {
  // 优先从 Pinia store 读取
  try {
    const store = useMainStore();
    if (store.selectedBusinessId?.[0]) return store.selectedBusinessId[0];
  } catch {
    // ignore
  }
  // 兼容：store 未就绪时（如页面刷新），从 localStorage 读取
  try {
    const saved = localStorage.getItem('bk_biz_id');
    if (saved) {
      const parsed = JSON.parse(saved);
      if (Array.isArray(parsed) && parsed.length > 0) return parsed[0];
    }
  } catch {
    // ignore
  }
  return undefined;
};

/** 通用主机列表查询（自动填充默认业务 ID） */
const listHosts = async (params?: Partial<{
  page: { limit: number; offset: number };
  exact: any;
  fuzzy: any;
}>) => {
  const request = buildHostListParams(params);
  const defaultBizId = getDefaultBizId();
  if (defaultBizId && (!request.exact_include_conditions.bk_biz_id
    || request.exact_include_conditions.bk_biz_id.length === 0)) {
    request.exact_include_conditions.bk_biz_id = [defaultBizId];
  }
  return TopoService.HostList(request).catch(() => ({ total: 0, items: [] }));
};

/**
 * 拉取拓扑树（业务列表 + 子节点）
 * 层级结构：业务(biz) → 管控区域(set) → 管控单元(module)
 *
 * 策略管理场景：currentStrategyBizId 有值时，根节点只返回当前业务，
 * 并预加载管控区域节点，每个管控区域显示主机数量
 *
 * 库传参（camelCase 懒加载子节点时）: { objectId, instanceId, meta }
 * 库期望返回: ITreeItem[] 数组
 */
export const fetchTopologyTree = async (node?: any): Promise<ITreeItem[]> => {
  // 如果传入了 node，说明是懒加载子节点
  if (node) {
    const objectId = node.objectId || node.object_id;
    const instanceId = node.instanceId ?? node.instance_id;
    const meta = node.meta || { bk_biz_id: 0, scope_id: '0', scope_type: 'biz' as const };

    // 业务 → 加载管控区域（作为 set 节点）
    if (objectId === 'biz') {
      try {
        const areaRes = await TopoService.NetworkAreaList({
          page: { offset: 0, limit: 500 },
          only_count: false,
          exact_include_conditions: { bk_networkarea_id: [], cloud_vendor: [] },
          fuzzy_include_conditions: { bk_networkarea_name: [] },
        });
        const areas = areaRes?.items || [];
        if (areas.length > 0) {
          return areas.map((area: any) => ({
            instance_id: area.bk_networkarea_id,
            instance_name: area.bk_networkarea_name,
            object_id: 'set',
            object_name: '管控区域',
            meta,
            child: [],
            count: 0,
            lazy: true,
            expanded: false,
          })) as ITreeItem[];
        }
      } catch {
        // ignore
      }
      return [];
    }

    // 集群/管控区域 → 加载管控单元（作为 module 节点）
    if (objectId === 'set') {
      try {
        const unitRes = await TopoService.NetworkUnitListBrief({
          page: { offset: 0, limit: 500 },
          only_count: false,
          exact_include_conditions: {
            bk_networkunit_id: [],
            bk_networkarea_id: [instanceId],
            is_direct: [],
            generation: [],
          },
        });
        const units = unitRes?.items || [];
        if (units.length > 0) {
          return units.map((unit: any) => ({
            instance_id: unit.bk_networkunit_id,
            instance_name: unit.bk_networkunit_name,
            object_id: 'module',
            object_name: '管控单元',
            meta,
            child: [],
            count: 0,
          })) as ITreeItem[];
        }
      } catch {
        // ignore
      }
      return [];
    }

    return [];
  }

  // 根节点
  // 策略管理/插件安装场景：只返回当前策略业务（支持单业务或业务数组）
  let effectiveBizIds = getEffectiveBizIds();

  // 兼容刷新后模块变量丢失：从 store / localStorage 补充
  if (effectiveBizIds.length === 0) {
    const fallbackBizId = getDefaultBizId();
    if (fallbackBizId) effectiveBizIds = [fallbackBizId];
  }

  if (effectiveBizIds.length > 0) {
    try {
      const bizPromises = effectiveBizIds.map(bizId =>
        Promise.all([
          TopoService.BusinessList({
            page: { offset: 0, limit: 500 },
            only_count: false,
            exact_include_conditions: { bk_biz_id: [bizId] },
            fuzzy_include_conditions: { bk_biz_name: [] },
          }),
          TopoService.HostList({
            page: { offset: 0, limit: 0 },
            only_count: true,
            exact_include_conditions: {
              bk_biz_id: [bizId],
              node_role: resolveNodeRoleFilter(),
            },
            fuzzy_include_conditions: {},
          }).catch(() => ({ total: 0 })),
        ])
      );
      const allResults = await Promise.all(bizPromises);
      const areaRes = await TopoService.NetworkAreaList({
        page: { offset: 0, limit: 500 },
        only_count: false,
        exact_include_conditions: { bk_networkarea_id: [], cloud_vendor: [] },
        fuzzy_include_conditions: { bk_networkarea_name: [] },
      });
      const areas = areaRes?.items || [];

      const bizNodes: ITreeItem[] = [];
      for (const [bizRes, hostCountRes] of allResults) {
        const businesses = bizRes?.items || [];
        if (businesses.length > 0) {
          const biz = businesses[0];
          const areaNodes = areas.map((area: any) => ({
            instance_id: area.bk_networkarea_id,
            instance_name: area.bk_networkarea_name,
            object_id: 'set',
            object_name: '管控区域',
            meta: {
              bk_biz_id: biz.bk_biz_id,
              scope_id: String(biz.bk_biz_id),
              scope_type: 'biz',
            },
            child: [],
            count: 0,
            lazy: true,
            expanded: false,
          }));

          bizNodes.push({
            instance_id: biz.bk_biz_id,
            instance_name: biz.bk_biz_name,
            object_id: 'biz',
            object_name: '业务',
            meta: {
              bk_biz_id: biz.bk_biz_id,
              scope_id: String(biz.bk_biz_id),
              scope_type: 'biz',
            },
            child: areaNodes as ITreeItem[],
            count: hostCountRes?.total ?? 0,
            lazy: false,
            expanded: true,
          });
        }
      }
      return bizNodes;
    } catch {
      // ignore
    }
    return [];
  }

  // 普通场景（插件手动拓扑等）：返回所有业务列表
  try {
    const bizRes = await TopoService.BusinessList({
      page: { offset: 0, limit: 500 },
      only_count: false,
      exact_include_conditions: { bk_biz_id: [] },
      fuzzy_include_conditions: { bk_biz_name: [] },
    });
    const businesses = bizRes?.items || [];
    if (businesses.length > 0) {
      return businesses.map((biz: any) => ({
        instance_id: biz.bk_biz_id,
        instance_name: biz.bk_biz_name,
        object_id: 'biz',
        object_name: '业务',
        meta: {
          bk_biz_id: biz.bk_biz_id,
          scope_id: String(biz.bk_biz_id),
          scope_type: 'biz',
        },
        child: [],
        count: 0,
        lazy: true,
        expanded: false,
      }));
    }
  } catch {
    // ignore
  }

  return [];
};

/**
 * 根据拓扑节点查询主机列表
 * 对接 TopoService.HostList
 *
 * 库传参（camelCase）: { nodeList: [...], pageSize: number, start: number, searchContent?: string }
 * 库期望返回: { data: Host[], total: number }
 */
export const fetchHostsByNodes = async (query: any): Promise<any> => {
  const searchContent = query.searchContent || query.search_content || '';
  const start = query.start ?? 0;
  const pageSize = query.pageSize ?? query.page_size ?? 20;

  // 从选中的拓扑节点提取过滤条件
  // IpSelector 传入的 nodeList 格式：[{ objectId, instanceId, instanceName, meta }]
  const nodeList = query.nodeList || [];
  const exact: Record<string, (string | number)[]> = {};
  const bizIds: number[] = [];
  const networkAreaIds: number[] = [];
  const networkUnitIds: number[] = [];

  for (const node of nodeList) {
    const objectId = node.objectId || node.object_id || '';
    const instanceId = node.instanceId ?? node.instance_id ?? node.id ?? null;
    if (instanceId === null || instanceId === undefined) continue;

    switch (objectId) {
      case 'biz':
        bizIds.push(Number(instanceId));
        break;
      case 'set':
        networkAreaIds.push(Number(instanceId));
        // 从 meta 中提取所属业务的 bk_biz_id
        if (node.meta?.bk_biz_id) {
          const metaBizId = Number(node.meta.bk_biz_id);
          if (!bizIds.includes(metaBizId)) bizIds.push(metaBizId);
        }
        break;
      case 'module':
      case 'networkunit':
        networkUnitIds.push(Number(instanceId));
        // 从 meta 中提取所属业务的 bk_biz_id
        if (node.meta?.bk_biz_id) {
          const metaBizId = Number(node.meta.bk_biz_id);
          if (!bizIds.includes(metaBizId)) bizIds.push(metaBizId);
        }
        break;
    }
  }

  if (bizIds.length > 0) exact.bk_biz_id = bizIds;
  if (networkAreaIds.length > 0) exact.bk_networkarea_id = networkAreaIds;
  if (networkUnitIds.length > 0) exact.bk_networkunit_id = networkUnitIds;
  // 根据 currentPolicyType 和权限自适应设置 node_role 过滤
  const nodeRoleFilter = resolveNodeRoleFilter();
  if (nodeRoleFilter.length > 0) {
    exact.node_role = nodeRoleFilter;
  }
  // 强制限制为当前业务（优先使用 setBizId 设置的数组，其次用 setStrategyBizId）
  const effectiveBizIds = getEffectiveBizIds();
  if (effectiveBizIds.length > 0) {
    exact.bk_biz_id = effectiveBizIds;
  }

  const fuzzy: Record<string, string[]> = {};
  if (searchContent) {
    fuzzy.bk_host_innerip = [searchContent];
    fuzzy.bk_host_name = [searchContent];
    fuzzy.bk_host_innerip_v6 = [searchContent];
  }

  try {
    const res = await TopoService.HostList({
      page: { offset: start, limit: pageSize },
      only_count: false,
      exact_include_conditions: exact,
      fuzzy_include_conditions: fuzzy,
    });

    const hostList = (res.items || []).map(mapHostItem);

    return {
      total: res.total ?? 0,
      data: hostList,
    };
  } catch {
    return { total: 0, data: [] };
  }
};

/**
 * 根据拓扑节点查询主机ID列表
 * 对接 TopoService.HostSelectHostID（策略管理）或 HostList（通用场景）
 *
 * 库传参（camelCase）: { nodeList: [...] }
 * 库期望返回: { data: number[] }
 */
export const fetchHostIdsByNodes = async (query: any): Promise<any> => {
  // 从选中的拓扑节点提取过滤条件
  const nodeList = query?.nodeList || [];
  const bizIds: number[] = [];
  const networkAreaIds: number[] = [];
  const networkUnitIds: number[] = [];

  for (const node of nodeList) {
    const objectId = node.objectId || node.object_id || '';
    const instanceId = node.instanceId ?? node.instance_id ?? node.id ?? null;
    if (instanceId === null || instanceId === undefined) continue;

    switch (objectId) {
      case 'biz':
        bizIds.push(Number(instanceId));
        break;
      case 'set':
        networkAreaIds.push(Number(instanceId));
        break;
      case 'module':
      case 'networkunit':
        networkUnitIds.push(Number(instanceId));
        break;
    }
  }

  // 策略管理/插件安装场景：使用 HostSelectHostID 精确接口
  const effectiveBizIds = getEffectiveBizIds();
  if (currentPolicyType || effectiveBizIds.length > 0) {
    try {
      const res = await TopoService.HostSelectHostID({
        exact_include_conditions: {
          bk_host_id: [],
          bk_biz_id: effectiveBizIds.length > 0 ? effectiveBizIds : bizIds,
          bk_networkarea_id: networkAreaIds,
          os_type: [],
          node_role: resolveNodeRoleFilter(),
          node_status: [],
          node_version: [],
          bk_agent_id: [],
          bk_networkunit_id: networkUnitIds,
          node_generation: [],
        },
        fuzzy_include_conditions: { bk_host_name: [], dept_name: [], bk_host_innerip: [], bk_host_innerip_v6: [], bk_host_outerip: [], bk_host_outerip_v6: [] },
        exact_exclude_conditions: { bk_host_id: [], bk_biz_id: [], bk_networkarea_id: [], os_type: [], node_role: [], node_status: [], node_version: [], bk_agent_id: [], bk_networkunit_id: [], node_generation: [] },
      });
      return {
        data: res?.items || [],
      };
    } catch {
      return { data: [] };
    }
  }

  // 通用场景（插件手动拓扑等）：使用 HostList 获取 host_id 列表
  try {
    const bkBizId = bizIds.length > 0 ? bizIds[0] : undefined;
    const res = await listHosts({
      page: { limit: 500, offset: 0 },
      exact: bkBizId ? { bk_biz_id: [bkBizId] } : {},
    });

    return {
      data: (res.items || [])
        .map((item: any) => item?.bk_host_id ?? item?.host_id)
        .filter(Boolean),
    };
  } catch {
    return { data: [] };
  }
};

/**
 * 查询多个节点的拓扑路径
 * 后端暂无对应接口，保留 mock
 */
export const fetchNodePath = async (params: any): Promise<Array<any[]>> => {
  const nodeList = params.nodeList || params.node_list || [];
  return nodeList.map((node: any) => [node]);
};

/**
 * 获取多个拓扑节点的主机 Agent 状态统计信息
 * 策略管理场景：返回 mock（暂不支持按节点统计）
 * 通用场景：通过 HostList 查询实际存活数
 */
export const fetchAgentStatistics = async (params: any): Promise<any[]> => {
  const nodeList = params.nodeList || params.node_list || [];

  // 策略管理/插件安装场景：返回 mock
  const effectiveBizIds = getEffectiveBizIds();
  if (currentPolicyType || effectiveBizIds.length > 0) {
    return nodeList.map((node: any) => ({
      node,
      agent_statistics: {
        alive_count: 0,
        not_alive_count: 0,
        total_count: 0,
      },
    }));
  }

  // 通用场景：通过 HostList 查询实际存活数
  const results = await Promise.all(nodeList.map(async (node: any) => {
    const bkBizId = node?.meta?.bk_biz_id;
    const res = await listHosts({
      page: { limit: 200, offset: 0 },
      exact: bkBizId ? { bk_biz_id: [bkBizId] } : {},
    });

    const total = res.total || (res.items || []).length;
    const alive = (res.items || []).filter((item: any) => {
      const status = item?.state?.node_status;
      return status === 'RUNNING' || status === 'ALIVE' || status === 'healthy' || status === 'HEALTHY';
    }).length;

    return {
      node,
      agent_statistics: {
        alive_count: alive,
        not_alive_count: Math.max(total - alive, 0),
        total_count: total,
      },
    };
  }));

  return results;
};

/**
 * 根据主机ID查询主机详情
 * 对接 TopoService.HostList（用 bk_host_id 精确查询）
 *
 * 库传参（camelCase）: { hostList: [{ hostId: number, meta: IMeta }] }
 * 库期望返回: Host[] 数组
 */
export const fetchHostDetails = async (params: any): Promise<any> => {
  const hostListParam = params.hostList || params.host_list || [];
  const ids = hostListParam.map((h: any) => h.hostId || h.host_id).filter(Boolean);

  if (ids.length === 0) {
    return { data: [] };
  }

  try {
    const res = await TopoService.HostList({
      page: { offset: 0, limit: ids.length },
      only_count: false,
      exact_include_conditions: {
        bk_host_id: ids,
        ...(Array.isArray(params.node_role) && params.node_role.length > 0
          ? { node_role: params.node_role }
          : {}),
      },
      fuzzy_include_conditions: {},
    });

    return {
      data: (res.items || []).map(mapHostItem),
    };
  } catch {
    return { data: [] };
  }
};

/**
 * 手动输入 - 根据用户输入的IP/主机名等查询主机信息
 * 对接 TopoService.HostList（模糊查询）
 *
 * 库传参（camelCase）: { ipList: string[], ipv6List: string[], keyList: string[] }
 * 库期望返回: Host[] 数组（IpSelector.ts 中会取 result.data.valid）
 */
export const fetchHostCheck = async (params: any): Promise<any> => {
  const ipList: string[] = params.ipList || params.ip_list || [];
  const ipv6List: string[] = params.ipv6List || params.ipv6_list || [];
  const keyList: string[] = params.keyList || params.key_list || [];

  // 解析所有 IP（支持 cloudId:ip 格式）
  const allIpv4: string[] = [];
  const allIpv6: string[] = [];

  ipList.forEach((input) => {
    const parsed = parseCloudInput(input);
    const value = normalizeIpValue(parsed.value);
    if (value) allIpv4.push(value);
  });

  ipv6List.forEach((input) => {
    const parsed = parseCloudInput(input);
    const value = normalizeIpValue(parsed.value);
    if (value) allIpv6.push(value);
  });

  // keyList 中纯数字视为 hostId，其余视为主机名
  const keywordIds = keyList.filter(item => /^\d+$/.test(item)).map(Number);
  const keywordNames = keyList.filter(item => !/^\d+$/.test(item));

  // 构建查询条件
  const exact: any = {};
  const fuzzy: any = {};

  if (keywordIds.length) exact.bk_host_id = keywordIds;
  if (allIpv4.length) fuzzy.bk_host_innerip = allIpv4;
  if (allIpv6.length) fuzzy.bk_host_innerip_v6 = allIpv6;
  if (keywordNames.length) fuzzy.bk_host_name = keywordNames;

  // 没有任何查询条件则直接返回
  if (!allIpv4.length && !allIpv6.length && !keywordIds.length && !keywordNames.length) {
    return { data: { valid: [], invalid: [] } };
  }

  try {
    const res = await listHosts({
      page: { limit: 200, offset: 0 },
      exact,
      fuzzy,
    });

    return {
      data: {
        valid: (res.items || []).map(mapHostItem),
        invalid: [],
      },
    };
  } catch {
    return { data: { valid: [], invalid: [] } };
  }
};
