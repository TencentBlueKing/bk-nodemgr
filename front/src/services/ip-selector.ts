/**
 * IP选择器专用服务 - 对接后端真实接口
 * 数据源对接 TopoService 中的主机相关 API
 */

import { TopoService } from '@/api/modules/topo';

// 当前策略类型（由外部设置，影响 node_role 查询条件）
// agent → node_role: ['agent', 'blank']
// proxy → node_role: ['proxy', 'blank']
export let currentPolicyType = 'config_policy_agent';

/** 设置当前策略类型（打开 IP 选择器前调用） */
export const setPolicyType = (type: string) => {
  currentPolicyType = type;
};

// 当前策略业务 ID（由外部设置，策略管理页面业务固定，拓扑树只展示当前业务）
let currentStrategyBizId: number | undefined;

/** 设置当前策略业务 ID（策略管理页面打开 IP 选择器前调用） */
export const setStrategyBizId = (bizId: number | undefined) => {
  currentStrategyBizId = bizId;
};

/** 根据策略类型获取 node_role 过滤值 */
const getNodeRoleFilter = (): string[] => {
  return currentPolicyType === 'config_policy_proxy'
    ? ['proxy', 'blank']
    : ['agent', 'blank'];
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

/**
 * 拉取拓扑树（业务列表 + 子节点）
 * 层级结构：业务(biz) → 管控区域(set) → 管控单元(module)
 * 对接真实 API：
 *   - 根节点：TopoService.BusinessList
 *   - 业务子节点(集群)：TopoService.NetworkAreaList
 *   - 集群子节点(模块)：TopoService.NetworkUnitListBrief
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
    const instanceId = node.instanceId || node.instance_id;
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
  // 策略管理场景：只返回当前策略业务
  if (currentStrategyBizId) {
    try {
      const [bizRes, areaRes, hostCountRes] = await Promise.all([
        TopoService.BusinessList({
          page: { offset: 0, limit: 500 },
          only_count: false,
          exact_include_conditions: { bk_biz_id: [currentStrategyBizId] },
          fuzzy_include_conditions: { bk_biz_name: [] },
        }),
        TopoService.NetworkAreaList({
          page: { offset: 0, limit: 500 },
          only_count: false,
          exact_include_conditions: { bk_networkarea_id: [], cloud_vendor: [] },
          fuzzy_include_conditions: { bk_networkarea_name: [] },
        }),
        // 用一次查询获取业务下的主机总数，作为业务节点的 count
        TopoService.HostList({
          page: { offset: 0, limit: 0 },
          only_count: true,
          exact_include_conditions: {
            bk_biz_id: [currentStrategyBizId],
            node_role: getNodeRoleFilter(),
          },
          fuzzy_include_conditions: {},
        }).catch(() => ({ total: 0 })),
      ]);
      const businesses = bizRes?.items || [];
      if (businesses.length > 0) {
        const biz = businesses[0];
        const areas = areaRes?.items || [];
        // 管控区域 count 不预加载，点击时由 fetchHostsByNodes 的 total 提供
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

        return [{
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
          lazy: false,  // 已预加载子节点，无需懒加载
          expanded: true, // 默认展开
        }];
      }
    } catch {
      // ignore
    }
    return [];
  }

  // 普通场景：返回所有业务列表
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
  // 根据策略类型设置 node_role 过滤
  exact.node_role = getNodeRoleFilter();
  // 策略管理场景：强制限制为当前策略业务
  if (currentStrategyBizId) {
    exact.bk_biz_id = [currentStrategyBizId];
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

    const hostList = (res.items || []).map((host: any) => ({
      id: String(host.bk_host_id),
      host_id: host.bk_host_id,
      ip: host.info?.bk_host_innerip_list?.[0] || '',
      ipv6: host.info?.bk_host_innerip_v6_list?.[0] || '',
      host_name: host.info?.bk_host_name || '',
      os_name: host.info?.os_type || '',
      os_type: host.info?.os_type || '',
      cpu_arch: host.info?.cpu_arch || '',
      alive: 1,
      cloud_area: {
        id: host.info?.bk_networkarea_id || 0,
        name: host.info?.bk_networkarea_name || '',
      },
      cloud_id: host.info?.bk_networkarea_id || 0,
      biz: { id: host.info?.bk_biz_id || 0, name: '' },
      meta: {
        scope_type: 'biz',
        scope_id: String(host.info?.bk_biz_id || 0),
        bk_biz_id: host.info?.bk_biz_id || 0,
      },
      bk_host_id: host.bk_host_id,
      bk_biz_id: host.info?.bk_biz_id || 0,
      bk_cloud_id: host.info?.bk_networkarea_id || 0,
      bk_agent_alive: 1,
    }));

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
 * 对接 TopoService.HostSelectHostID
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

  try {
    const res = await TopoService.HostSelectHostID({
      exact_include_conditions: {
        bk_host_id: [],
        bk_biz_id: currentStrategyBizId ? [currentStrategyBizId] : bizIds,
        bk_networkarea_id: networkAreaIds,
        os_type: [],
        node_role: getNodeRoleFilter(),
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
 * 后端暂无对应接口，保留 mock
 */
export const fetchAgentStatistics = async (params: any): Promise<any[]> => {
  const nodeList = params.nodeList || params.node_list || [];
  return nodeList.map((node: any) => ({
    node,
    agent_statistics: {
      alive_count: 0,
      not_alive_count: 0,
      total_count: 0,
    },
  }));
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
      exact_include_conditions: { bk_host_id: ids, bk_biz_id: [], bk_networkarea_id: [], os_type: [], node_role: [], node_status: [], node_version: [], bk_agent_id: [], bk_networkunit_id: [], node_generation: [] },
      fuzzy_include_conditions: { bk_host_name: [], dept_name: [], bk_host_innerip: [], bk_host_innerip_v6: [], bk_host_outerip: [], bk_host_outerip_v6: [] },
    });

    const hostList = (res.items || []).map((host: any) => ({
      id: String(host.bk_host_id),
      host_id: host.bk_host_id,
      ip: host.info?.bk_host_innerip_list?.[0] || '',
      ipv6: host.info?.bk_host_innerip_v6_list?.[0] || '',
      host_name: host.info?.bk_host_name || '',
      os_name: host.info?.os_type || '',
      os_type: host.info?.os_type || '',
      cpu_arch: host.info?.cpu_arch || '',
      alive: 1,
      cloud_area: {
        id: host.info?.bk_networkarea_id || 0,
        name: host.info?.bk_networkarea_name || '',
      },
      cloud_id: host.info?.bk_networkarea_id || 0,
      biz: { id: host.info?.bk_biz_id || 0, name: '' },
      meta: {
        scope_type: 'biz',
        scope_id: String(host.info?.bk_biz_id || 0),
        bk_biz_id: host.info?.bk_biz_id || 0,
      },
      bk_host_id: host.bk_host_id,
      bk_biz_id: host.info?.bk_biz_id || 0,
      bk_cloud_id: host.info?.bk_networkarea_id || 0,
      bk_agent_alive: 1,
    }));

    return { data: hostList };
  } catch {
    return { data: [] };
  }
};

/**
 * 手动输入 - 根据用户输入的IP/主机名等查询主机信息
 * 对接 TopoService.HostSelectInnerIP / HostSelectInnerIPV6
 *
 * 库传参（camelCase）: { ipList: string[], ipv6List: string[], keyList: string[] }
 * 库期望返回: Host[] 数组（IpSelector.ts 中会取 result.data.valid）
 */
export const fetchHostCheck = async (params: any): Promise<any> => {
  const ipList: string[] = params.ipList || params.ip_list || [];
  const ipv6List: string[] = params.ipv6List || params.ipv6_list || [];

  const valid: any[] = [];

  try {
    // 查询 IPv4
    if (ipList.length > 0) {
      const ipRes = await TopoService.HostSelectNetWorkareaIDAndInnerIP({
        exact_include_conditions: { bk_host_id: [], bk_biz_id: currentStrategyBizId ? [currentStrategyBizId] : [], bk_networkarea_id: [], os_type: [], node_role: [], node_status: [], node_version: [], bk_agent_id: [], bk_networkunit_id: [], node_generation: [] },
        fuzzy_include_conditions: { bk_host_name: [], dept_name: [], bk_host_innerip: ipList, bk_host_innerip_v6: [], bk_host_outerip: [], bk_host_outerip_v6: [] },
        exact_exclude_conditions: { bk_host_id: [], bk_biz_id: [], bk_networkarea_id: [], os_type: [], node_role: [], node_status: [], node_version: [], bk_agent_id: [], bk_networkunit_id: [], node_generation: [] },
      });

      // ipRes.items 是 "bk_networkarea_id:ip" 格式的字符串数组
      const ipItems: string[] = ipRes?.items || [];
      if (ipItems.length > 0) {
        // 用查到的 ip 再查 HostList 获取完整信息
        const hostRes = await TopoService.HostList({
          page: { offset: 0, limit: ipItems.length },
          only_count: false,
          exact_include_conditions: { bk_host_id: [], bk_biz_id: currentStrategyBizId ? [currentStrategyBizId] : [], bk_networkarea_id: [], os_type: [], node_role: [], node_status: [], node_version: [], bk_agent_id: [], bk_networkunit_id: [], node_generation: [] },
          fuzzy_include_conditions: { bk_host_name: [], dept_name: [], bk_host_innerip: ipList, bk_host_innerip_v6: [], bk_host_outerip: [], bk_host_outerip_v6: [] },
        });

        (hostRes.items || []).forEach((host: any) => {
          valid.push({
            id: String(host.bk_host_id),
            host_id: host.bk_host_id,
            ip: host.info?.bk_host_innerip_list?.[0] || '',
            ipv6: host.info?.bk_host_innerip_v6_list?.[0] || '',
            host_name: host.info?.bk_host_name || '',
            os_name: host.info?.os_type || '',
            os_type: host.info?.os_type || '',
            cpu_arch: host.info?.cpu_arch || '',
            alive: 1,
            cloud_area: {
              id: host.info?.bk_networkarea_id || 0,
              name: host.info?.bk_networkarea_name || '',
            },
            cloud_id: host.info?.bk_networkarea_id || 0,
            biz: { id: host.info?.bk_biz_id || 0, name: '' },
            meta: {
              scope_type: 'biz',
              scope_id: String(host.info?.bk_biz_id || 0),
              bk_biz_id: host.info?.bk_biz_id || 0,
            },
            bk_host_id: host.bk_host_id,
            bk_biz_id: host.info?.bk_biz_id || 0,
            bk_cloud_id: host.info?.bk_networkarea_id || 0,
            bk_agent_alive: 1,
          });
        });
      }
    }

    // 查询 IPv6
    if (ipv6List.length > 0) {
      const ipv6Res = await TopoService.HostSelectNetWorkareaIDAndInnerIPV6({
        exact_include_conditions: { bk_host_id: [], bk_biz_id: currentStrategyBizId ? [currentStrategyBizId] : [], bk_networkarea_id: [], os_type: [], node_role: [], node_status: [], node_version: [], bk_agent_id: [], bk_networkunit_id: [], node_generation: [] },
        fuzzy_include_conditions: { bk_host_name: [], dept_name: [], bk_host_innerip: [], bk_host_innerip_v6: ipv6List, bk_host_outerip: [], bk_host_outerip_v6: [] },
        exact_exclude_conditions: { bk_host_id: [], bk_biz_id: [], bk_networkarea_id: [], os_type: [], node_role: [], node_status: [], node_version: [], bk_agent_id: [], bk_networkunit_id: [], node_generation: [] },
      });

      const ipv6Items: string[] = ipv6Res?.items || [];
      if (ipv6Items.length > 0) {
        const hostRes = await TopoService.HostList({
          page: { offset: 0, limit: ipv6Items.length },
          only_count: false,
          exact_include_conditions: { bk_host_id: [], bk_biz_id: currentStrategyBizId ? [currentStrategyBizId] : [], bk_networkarea_id: [], os_type: [], node_role: [], node_status: [], node_version: [], bk_agent_id: [], bk_networkunit_id: [], node_generation: [] },
          fuzzy_include_conditions: { bk_host_name: [], dept_name: [], bk_host_innerip: [], bk_host_innerip_v6: ipv6List, bk_host_outerip: [], bk_host_outerip_v6: [] },
        });

        (hostRes.items || []).forEach((host: any) => {
          // 避免重复
          if (!valid.some((h) => h.host_id === host.bk_host_id)) {
            valid.push({
              id: String(host.bk_host_id),
              host_id: host.bk_host_id,
              ip: host.info?.bk_host_innerip_list?.[0] || '',
              ipv6: host.info?.bk_host_innerip_v6_list?.[0] || '',
              host_name: host.info?.bk_host_name || '',
              os_name: host.info?.os_type || '',
              os_type: host.info?.os_type || '',
              cpu_arch: host.info?.cpu_arch || '',
              alive: 1,
              cloud_area: {
                id: host.info?.bk_networkarea_id || 0,
                name: host.info?.bk_networkarea_name || '',
              },
              cloud_id: host.info?.bk_networkarea_id || 0,
              biz: { id: host.info?.bk_biz_id || 0, name: '' },
              meta: {
                scope_type: 'biz',
                scope_id: String(host.info?.bk_biz_id || 0),
                bk_biz_id: host.info?.bk_biz_id || 0,
              },
              bk_host_id: host.bk_host_id,
              bk_biz_id: host.info?.bk_biz_id || 0,
              bk_cloud_id: host.info?.bk_networkarea_id || 0,
              bk_agent_alive: 1,
            });
          }
        });
      }
    }
  } catch {
    // 查询失败返回空
  }

  return {
    data: {
      valid,
      invalid: [],
    },
  };
};
