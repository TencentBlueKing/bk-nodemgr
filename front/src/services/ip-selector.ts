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

// 当前激活面板（由外部通过 @panel-change 设置）
export let activeIpsPanel = 'staticTopo';
export const setActiveIpsPanel = (panel: string) => {
  activeIpsPanel = panel;
};

/** 统一别名：设置类型 */
export const setType = setPolicyType;

// 当前策略业务 ID 数组（由外部设置，从 store 或 localStorage 获取）
let currentBizIds: number[] = [];
// 是否已显式调用过 setBizId（用于区分"从未设置"和"主动设为全选空数组"）
let bizIdExplicitlySet = false;

/** 设置当前业务 ID 数组（插件安装等场景使用）
 *  调用时会重置 currentPolicyType 为空，确保走权限自适应分支而非残留的 type 值
 *  传入空数组表示全选（不限业务） */
export const setBizId = (bizIds: number[]) => {
  currentPolicyType = '';  // 重置：避免策略管理页面的残留 type 污染
  currentBizIds = bizIds;
  bizIdExplicitlySet = true;
};

// 当前策略业务 ID（由外部设置，策略管理页面业务固定，拓扑树只展示当前业务）
// 兼容旧接口，内部复用 currentBizIds[0]
let currentStrategyBizId: number | undefined;

/** 设置当前策略业务 ID（策略管理页面打开 IP 选择器前调用）【兼容旧接口】
 *  策略管理为单业务模式，每次打开 IP 选择器应只展示当前业务，替换而非累加旧业务 ID */
export const setStrategyBizId = (bizId: number | undefined) => {
  currentStrategyBizId = bizId;
  currentBizIds = bizId !== undefined ? [bizId] : [];
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
export const mapHostItem = (item: any) => {
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
    cpu_arch: info?.cpu_arch || '',
    alive: agentAlive,
    cloud_area: { id: networkAreaId, name: info?.bk_networkarea_name || '' },
    cloud_id: networkAreaId,
    biz: { id: bkBizId, name: '' },
    agent_id: state?.bk_agent_id || '',
    bk_networkunit_id: info?.bk_networkunit_id ?? 0,
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

// 拓扑节点缓存（非主机 set/module 节点）：bizId → Map<key, ITreeItem>
const topoNodeCache = new Map<number, Map<string, ITreeItem>>();
// 父级关系：bizId → Map<childKey, parentKey>
const topoParentMap = new Map<number, Map<string, string>>();
// 业务名称：bizId → bizName
const bizNameMap = new Map<number, string>();

interface IHostQueryNodeIDs {
  bizIds: number[];
  setIds: number[];
  moduleIds: number[];
}

const appendUniqueID = (ids: number[], value: unknown) => {
  const id = Number(value);
  if (!Number.isFinite(id) || ids.includes(id)) return;
  ids.push(id);
};

/** Collect set IDs below a custom topology node from the cached topology tree. */
const collectDescendantSetIDs = (node: ITreeItem, setIds: number[]) => {
  for (const child of node.child || []) {
    if (child.object_id === 'set') {
      appendUniqueID(setIds, child.instance_id);
    }
    collectDescendantSetIDs(child, setIds);
  }
};

/** Resolve host query IDs from the nodes selected in the IP selector. */
const resolveHostQueryNodeIDs = (nodeList: any[]): IHostQueryNodeIDs => {
  const result: IHostQueryNodeIDs = { bizIds: [], setIds: [], moduleIds: [] };

  for (const node of nodeList) {
    const objectId = node.objectId || node.object_id || '';
    const instanceId = node.instanceId ?? node.instance_id ?? node.id ?? null;
    if (instanceId === null || instanceId === undefined) continue;

    switch (objectId) {
      case 'biz':
        appendUniqueID(result.bizIds, instanceId);
        break;
      case 'set':
        appendUniqueID(result.setIds, instanceId);
        appendUniqueID(result.bizIds, node.meta?.bk_biz_id ?? node.bk_biz_id);
        break;
      case 'module':
        appendUniqueID(result.moduleIds, instanceId);
        appendUniqueID(result.bizIds, node.meta?.bk_biz_id ?? node.bk_biz_id);
        break;
      default: {
        const bizID = node.meta?.bk_biz_id ?? node.bk_biz_id;
        appendUniqueID(result.bizIds, bizID);
        const cachedNode = topoNodeCache.get(Number(bizID))?.get(`${objectId}:${instanceId}`);
        if (cachedNode) {
          collectDescendantSetIDs(cachedNode, result.setIds);
        }
      }
    }
  }

  return result;
};

/** 递归索引 ITreeItem 子树，写入 topoNodeCache / topoParentMap */
const indexTreeNodes = (children: ITreeItem[], bizId: number, bizName?: string) => {
  const nodeMap = new Map<string, ITreeItem>();
  const parentMap = new Map<string, string>();
  const walk = (n: ITreeItem, pKey?: string) => {
    const k = `${n.object_id}:${n.instance_id}`;
    nodeMap.set(k, n);
    if (pKey) parentMap.set(k, pKey);
    (n.child || []).forEach((c: ITreeItem) => walk(c, k));
  };
  children.forEach(c => walk(c));
  topoNodeCache.set(bizId, nodeMap);
  topoParentMap.set(bizId, parentMap);
  if (bizName !== undefined) bizNameMap.set(bizId, bizName);
};

/**
 * 将 CMDB 实例拓扑节点 TopoNodeInfo 递归转换为 IP 选择器树节点 ITreeItem
 * TopoNodeInfo: { topo_inst_id, topo_inst_name, topo_obj_id, host_count, children }
 * ITreeItem: { instance_id, instance_name, object_id, meta, child, count }
 */
const transformCmdbNode = (node: any, bizId: number): ITreeItem => ({
  instance_id: node.topo_inst_id,
  instance_name: node.topo_inst_name,
  object_id: node.topo_obj_id,       // 'set' | 'module'
  meta: {
    bk_biz_id: bizId,
    scope_id: String(bizId),
    scope_type: 'biz' as const,
  },
  child: (node.children || []).map((child: any) => transformCmdbNode(child, bizId)),
  count: node.host_count || 0,
  // set 节点预展开让用户看到 module 层级，module 为叶子节点
  expanded: node.topo_obj_id === 'set',
  lazy: false,
});

/**
 * 拉取拓扑树（业务列表 + 子节点）
 * 层级结构：业务(biz) → CMDB Set(set) → CMDB Module(module)
 *
 * 策略管理场景：currentStrategyBizId / setBizId 有值时只返回当前业务
 *
 * 库传参（camelCase 懒加载子节点时）: { objectId, instanceId, meta }
 * 库期望返回: ITreeItem[] 数组
 */
export const fetchTopologyTree = async (node?: any): Promise<ITreeItem[]> => {
  // 如果传入了 node，说明是懒加载子节点
  if (node) {
    const objectId = node.objectId || node.object_id;
    const meta = node.meta || { bk_biz_id: 0, scope_id: '0', scope_type: 'biz' as const };

    // 业务 → 加载 CMDB 实例拓扑（set + module 完整子树）
    if (objectId === 'biz') {
      try {
        const res = await TopoService.BusinessInstTopoGet({ bk_biz_id: meta.bk_biz_id });
        const items = (res as any)?.items; // TopoNodeInfo 根节点（biz 自身）
        if (items?.children?.length > 0) {
          const children = items.children.map((child: any) => transformCmdbNode(child, meta.bk_biz_id));
          indexTreeNodes(children, meta.bk_biz_id);
          return children;
        }
      } catch {
        // ignore
      }
      return [];
    }

    // set/module 已在 biz 展开时全量加载，无需再次懒加载
    return [];
  }

  // 根节点
  // 策略管理/插件安装场景：只返回当前策略业务（支持单业务或业务数组）
  let effectiveBizIds = getEffectiveBizIds();

  // 兼容刷新后模块变量丢失：从 store / localStorage 补充
  if (effectiveBizIds.length === 0 && !bizIdExplicitlySet) {
    const fallbackBizId = getDefaultBizId();
    if (fallbackBizId) effectiveBizIds = [fallbackBizId];
  }

  if (effectiveBizIds.length > 0) {
    try {
      // 并行调用：BusinessList 获取业务列表 + BusinessHostCountGet 获取主机数量
      const [bizRes, countRes] = await Promise.all([
        TopoService.BusinessList({
          page: { offset: 0, limit: 500 },
          only_count: false,
          exact_include_conditions: { bk_biz_id: effectiveBizIds },
          fuzzy_include_conditions: { bk_biz_name: [] },
        }),
        TopoService.BusinessHostCountGet({ bk_biz_id: effectiveBizIds }),
      ]);

      const businesses = bizRes?.items || [];
      const countMap = new Map((countRes?.items || []).map((item: any) => [item.bk_biz_id, item.host_count]));

      if (businesses.length > 0) {
        // 仅预加载第一个业务（active 业务）的 CMDB 实例拓扑，其余业务点击展开时懒加载
        const firstBiz = businesses[0];
        let firstBizChildren: any[] = [];
        try {
          const topoRes = await TopoService.BusinessInstTopoGet({ bk_biz_id: firstBiz.bk_biz_id });
          const data = (topoRes as any)?.items;
          firstBizChildren = data?.children || [];
        } catch {
          // ignore
        }

        // 缓存第一个业务的拓扑节点和名称
        bizNameMap.set(firstBiz.bk_biz_id, firstBiz.bk_biz_name);
        if (firstBizChildren.length > 0) {
          indexTreeNodes(firstBizChildren.map((c: any) => transformCmdbNode(c, firstBiz.bk_biz_id)), firstBiz.bk_biz_id);
        }

        // 缓存其余业务的名称
        for (let i = 1; i < businesses.length; i++) {
          bizNameMap.set(businesses[i].bk_biz_id, businesses[i].bk_biz_name);
        }

        return businesses.map((biz: any, idx: number) => {
          const isFirst = idx === 0;
          const children = isFirst ? firstBizChildren : [];
          return {
            instance_id: biz.bk_biz_id,
            instance_name: biz.bk_biz_name,
            object_id: 'biz',
            meta: {
              bk_biz_id: biz.bk_biz_id,
              scope_id: String(biz.bk_biz_id),
              scope_type: 'biz' as const,
            },
            child: isFirst ? children.map((child: any) => transformCmdbNode(child, biz.bk_biz_id)) : [],
            count: countMap.get(biz.bk_biz_id) ?? 0,
            // 第一个业务预加载拓扑，其余业务点击展开时懒加载
            lazy: !isFirst,
            expanded: false,
          };
        });
      }
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
      const bizIds = businesses.map((biz: any) => biz.bk_biz_id);
      // 获取所有业务的 host_count
      let countMap = new Map<number, number>();
      try {
        const countRes = await TopoService.BusinessHostCountGet({ bk_biz_id: bizIds });
        countMap = new Map((countRes?.items || []).map((item: any) => [item.bk_biz_id, item.host_count]));
      } catch {
        // ignore
      }

      // 仅预加载第一个业务（active 业务）的 CMDB 实例拓扑，其余业务点击展开时懒加载
      const firstBiz = businesses[0];
      let firstBizChildren: any[] = [];
      try {
        const topoRes = await TopoService.BusinessInstTopoGet({ bk_biz_id: firstBiz.bk_biz_id });
        const data = (topoRes as any)?.items;
        firstBizChildren = data?.children || [];
      } catch {
        // ignore
      }

      // 缓存第一个业务的拓扑节点和名称
      bizNameMap.set(firstBiz.bk_biz_id, firstBiz.bk_biz_name);
      if (firstBizChildren.length > 0) {
        indexTreeNodes(firstBizChildren.map((c: any) => transformCmdbNode(c, firstBiz.bk_biz_id)), firstBiz.bk_biz_id);
      }

      // 缓存其余业务的名称
      for (let i = 1; i < businesses.length; i++) {
        bizNameMap.set(businesses[i].bk_biz_id, businesses[i].bk_biz_name);
      }

      return businesses.map((biz: any, idx: number) => {
        const isFirst = idx === 0;
        const children = isFirst ? firstBizChildren : [];
        return {
          instance_id: biz.bk_biz_id,
          instance_name: biz.bk_biz_name,
          object_id: 'biz',
          meta: {
            bk_biz_id: biz.bk_biz_id,
            scope_id: String(biz.bk_biz_id),
            scope_type: 'biz',
          },
          child: isFirst ? children.map((child: any) => transformCmdbNode(child, biz.bk_biz_id)) : [],
          count: countMap.get(biz.bk_biz_id) ?? 0,
          // 第一个业务预加载拓扑，其余业务点击展开时懒加载
          lazy: !isFirst,
          expanded: false,
        };
      });
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
  // 拓扑层级：业务(biz) → 自定义层级 → CMDB Set(set) → CMDB Module(module)
  const nodeList = query.nodeList || query.node_list || [];
  const exact: Record<string, (string | number)[]> = {};
  const { bizIds, setIds, moduleIds } = resolveHostQueryNodeIDs(nodeList);

  if (bizIds.length > 0) exact.bk_biz_id = bizIds;
  if (setIds.length > 0) exact.bk_set_id = setIds;
  if (moduleIds.length > 0) exact.bk_module_id = moduleIds;
  // 根据 currentPolicyType 和权限自适应设置 node_role 过滤
  const nodeRoleFilter = resolveNodeRoleFilter();
  if (nodeRoleFilter.length > 0) {
    exact.node_role = nodeRoleFilter;
  }
  // 主机查询条件只传当前激活/选中节点所属的业务 ID，不传全部业务
  // （用户点击某个业务下的节点时，bk_biz_id 应该只有该业务）

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
 * 统一使用 HostSelectHostID 跨页全选专用接口
 * 后端内部通过 pageexecutor 自动分页（5000/页），无需前端分片
 *
 * 库传参（camelCase）: { nodeList: [...] }
 * 库内部 fetchNodeAllHostId 取 data.data 后直接遍历元素访问 .host_id，
 * 因此返回的 data 中每个元素必须是 { host_id: number } 格式的对象，而非纯数字。
 */
export const fetchHostIdsByNodes = async (query: any): Promise<any> => {
  // 从选中的拓扑节点提取过滤条件
  // 拓扑层级：业务(biz) → 自定义层级 → CMDB Set(set) → CMDB Module(module)
  const nodeList = query?.nodeList || query?.node_list || [];
  const { bizIds, setIds, moduleIds } = resolveHostQueryNodeIDs(nodeList);

  // 统一使用 HostSelectHostID 跨页全选专用接口
  // 后端通过 pageexecutor（5000/页，1分钟超时）自动分片，一次请求即可获取全量
  // 只传选中节点所属的业务 ID
  try {
    const res = await TopoService.HostSelectHostID({
      exact_include_conditions: {
        bk_host_id: [],
        bk_biz_id: bizIds.length > 0 ? bizIds : [],
        bk_networkarea_id: [],
        os_type: [],
        node_role: resolveNodeRoleFilter(),
        node_status: [],
        node_version: [],
        bk_agent_id: [],
        bk_networkunit_id: [],
        node_generation: [],
        bk_set_id: setIds,
        bk_module_id: moduleIds,
      },
      fuzzy_include_conditions: {
        bk_host_name: [], dept_name: [], bk_host_innerip: [],
        bk_host_innerip_v6: [], bk_host_outerip: [], bk_host_outerip_v6: [],
      },
      exact_exclude_conditions: {
        bk_host_id: [], bk_biz_id: [], bk_networkarea_id: [],
        os_type: [], node_role: [], node_status: [], node_version: [],
        bk_agent_id: [], bk_networkunit_id: [], node_generation: [],
        bk_set_id: [], bk_module_id: [],
      },
    });

    // 库 fetchNodeAllHostId 取 data.data 后直接遍历元素访问 .host_id，
    // 因此每个元素必须是 { host_id: number } 对象，不能是纯数字
    const items: number[] = res?.items || [];
    return {
      data: items.map((id: number) => ({ host_id: id })),
    };
  } catch {
    return { data: [] };
  }
};

/**
 * 查询多个节点的拓扑路径（仅非主机 set/module 节点）
 * 从 topoNodeCache / topoParentMap / bizNameMap 构建完整路径
 * 库消费：nodeStack 最后元素做 key，所有 instance_name 用 ' / ' 拼接做 namePath
 */
export const fetchNodePath = async (params: any): Promise<Array<any[]>> => {
  const nodeList = params.nodeList || params.node_list || [];
  return nodeList.map((node: any) => {
    const bizId = node.meta?.bk_biz_id ?? node.bk_biz_id;
    const objectId = node.objectId || node.object_id || '';
    const instanceId = node.instanceId ?? node.instance_id ?? node.id;
    const cache = bizId ? topoNodeCache.get(bizId) : undefined;
    const parentMap = bizId ? topoParentMap.get(bizId) : undefined;

    const stack: ITreeItem[] = [];
    let curKey = `${objectId}:${instanceId}`;
    const visited = new Set<string>();

    const curNode = cache?.get(curKey);
    if (curNode) stack.push(curNode);

    while (parentMap && curKey && !visited.has(curKey)) {
      visited.add(curKey);
      const pk = parentMap.get(curKey);
      if (!pk) break;
      const pn = cache?.get(pk);
      if (pn) stack.push(pn);
      curKey = pk;
    }

    // 栈顶加业务根节点
    const bizName = bizId ? bizNameMap.get(bizId) : undefined;
    if (bizId && bizName) {
      stack.push({
        instance_id: bizId,
        instance_name: bizName,
        object_id: 'biz',
        meta: { bk_biz_id: bizId, scope_id: String(bizId), scope_type: 'biz' },
        count: 0,
      } as ITreeItem);
    }

    return stack.reverse();
  });
};

/**
 * 获取多个拓扑节点的主机 Agent 状态统计信息
 *
 * 数据来源：BusinessInstTopoGet 接口返回的 TopoNodeInfo 已包含 host_count，
 * 经 transformCmdbNode 映射到 ITreeItem.count 字段。
 * 因此无需再调 HostList 查询，直接从 node 的 count 属性取值即可。
 */
/**
 * 获取多个拓扑节点的主机 Agent 状态统计信息（仅非主机 set/module 节点）
 *
 * 库传参 camelCase：{ nodeList: [{ objectId, instanceId, meta: { bk_biz_id } }] }
 * 库消费：取 item.agent_statistics.{alive_count,not_alive_count,total_count}
 *        用 genNodeKey(item.node) 做 Map key → 依赖 node.object_id + node.instance_id（snake_case）
 */
export const fetchAgentStatistics = async (params: any): Promise<any[]> => {
  const nodeList = params.nodeList || params.node_list || [];

  return nodeList.map((node: any) => {
    const objectId = node.objectId || node.object_id || '';
    const instanceId = node.instanceId ?? node.instance_id ?? node.id;

    // 从缓存取 host_count 作为 total_count（库传入的 node 无 count 字段）
    const bizId = node.meta?.bk_biz_id ?? node.bk_biz_id;
    const cache = bizId ? topoNodeCache.get(bizId) : undefined;
    const cached = cache?.get(`${objectId}:${instanceId}`);

    return {
      // 补齐 snake_case：genNodeKey 需要 object_id + instance_id
      node: { ...node, object_id: objectId, instance_id: instanceId },
      agent_statistics: {
        alive_count: 0,
        not_alive_count: 0,
        total_count: cached?.count ?? 0,
      },
    };
  });
};

/**
 * 根据主机ID查询主机详情
 * 对接 TopoService.HostList（用 bk_host_id 精确查询）
 *
 * 库传参（camelCase）: { hostList: [{ hostId: number, meta: IMeta }] }
 * 库期望返回: Host[] 数组
 *
 * 注意：跨页全选时 hostList 可能包含数万个 ID，
 * 必须分片请求避免单次 limit 超过后端上限。
 */
export const fetchHostDetails = async (params: any): Promise<any> => {
  const hostListParam = params.hostList || params.host_list || [];
  const ids = hostListParam.map((h: any) => h.hostId || h.host_id).filter(Boolean);

  if (ids.length === 0) {
    return { data: [] };
  }

  try {
    // 分片请求：每批最多 500 条，避免后端 limit 上限
    const CHUNK_SIZE = 500;
    const allItems: any[] = [];

    for (let i = 0; i < ids.length; i += CHUNK_SIZE) {
      const chunkIds = ids.slice(i, i + CHUNK_SIZE);
      const res = await TopoService.HostList({
        page: { offset: 0, limit: CHUNK_SIZE },
        only_count: false,
        exact_include_conditions: {
          bk_host_id: chunkIds,
          node_role: resolveNodeRoleFilter(),
        },
        fuzzy_include_conditions: {},
      });
      allItems.push(...(res.items || []));
    }

    return { data: allItems.map(mapHostItem) };
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
    // 与静态拓扑选择一致，根据 agent_view/proxy_view 权限过滤 node_role
    const nodeRoleFilter = resolveNodeRoleFilter();
    if (nodeRoleFilter.length > 0) {
      exact.node_role = nodeRoleFilter;
    }

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

// ========== host cache ==========
const hostCache = new Map<number, any>();

/** 缓存单个 host 条目 */
export const cacheHostItem = (item: any) => {
  const id = item?.host_id ?? item?.bk_host_id ?? item?.id;
  if (id != null) {
    hostCache.set(Number(id), item);
  }
};

/** 从缓存批量取 host 条目，返回 { cached, missIds } */
export const getCachedHosts = (ids: number[]): { cached: any[]; missIds: number[] } => {
  const cached: any[] = [];
  const missIds: number[] = [];
  for (const id of ids) {
    const item = hostCache.get(id);
    if (item) {
      cached.push(item);
    } else {
      missIds.push(id);
    }
  }
  return { cached, missIds };
};
