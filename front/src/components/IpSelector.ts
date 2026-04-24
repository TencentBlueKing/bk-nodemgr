import createFactory from '@blueking/ip-selector/dist/vue3.x';
import '@blueking/ip-selector/dist/styles/vue2.6.x.css';
import '@blueking/ip-selector/dist/styles/index.css';
import * as IpSelectorService from '@/services/ip-selector';

/**
 * IP选择器配置 - Vue 3 版本
 * 基于蓝鲸官方 @blueking/ip-selector Vue3 版本
 *
 * 重要：使用 section 模式（非 dialog 模式），由外部 bkui-vue Dialog 包裹
 * 避免 Vue3→Vue2→Vue3 三层嵌套导致的 bk-dialog 渲染异常
 *
 * 数据源（Service）对应关系：
 * - fetchTopologyHostCount → 拉取拓扑树 + 主机数量
 * - fetchTopologyHostsNodes → 根据拓扑节点查询主机列表
 * - fetchTopologyHostIdsNodes → 根据拓扑节点查询主机ID列表
 * - fetchHostsDetails → 根据主机ID查询主机详情
 * - fetchHostCheck → 手动输入解析主机
 * - fetchNodesQueryPath → 查询节点拓扑路径
 * - fetchHostAgentStatisticsNodes → 拓扑节点 Agent 状态统计
 */

// 数据源服务配置
// 返回格式严格对齐 @blueking/ip-selector 源码中的消费方式
const Service = {
  /**
   * 拉取拓扑树（业务列表 + 子节点）
   * 源码: transformTopoTree(data) 直接对返回值 .map() → 必须返回纯数组
   */
  fetchTopologyHostCount: async (node?: any): Promise<any> => {
    try {
      const result = await IpSelectorService.fetchTopologyTree(node);
      // 库内部 transformTopoTree 直接对返回值 .map()，必须返回数组
      return Array.isArray(result) ? result : [];
    } catch (error) {
      console.error('[IpSelector] fetchTopologyHostCount error:', error);
      return [];
    }
  },

  /**
   * 根据拓扑节点查询主机列表
   * 源码: data.data / data.total → 返回 { data: Host[], total: number }
   */
  fetchTopologyHost: async (params: any) => {
    try {
      const result = await IpSelectorService.fetchHostsByNodes(params);
      const data = result.data || [];
      return { data, total: result.total || 0 };
    } catch (error) {
      console.error('[IpSelector] fetchTopologyHost error:', error);
      return { data: [], total: 0 };
    }
  },

  /**
   * 根据拓扑节点查询主机ID列表
   * 源码: data.data → 返回 { data: number[] }
   */
  fetchTopogyHostIdList: async (params: any) => {
    try {
      const result = await IpSelectorService.fetchHostIdsByNodes(params);
      return { data: result.data || [] };
    } catch (error) {
      console.error('[IpSelector] fetchTopogyHostIdList error:', error);
      return { data: [] };
    }
  },

  /**
   * 根据主机ID查询主机详情
   * 源码: 直接遍历返回值 → 返回 Host[] 数组
   */
  fetchHostInfoByHostId: async (params: any) => {
    try {
      const result = await IpSelectorService.fetchHostDetails(params);
      const data = result.data || [];
      return data;
    } catch (error) {
      console.error('[IpSelector] fetchHostInfoByHostId error:', error);
      return [];
    }
  },

  /**
   * 手动输入 - 解析用户输入的IP/主机名
   * 源码: 直接遍历返回值 → 返回 Host[] 数组
   */
  fetchInputParseHostList: async (params: any) => {
    try {
      const result = await IpSelectorService.fetchHostCheck(params);
      const data = result.data?.valid || [];
      return data;
    } catch (error) {
      console.error('[IpSelector] fetchInputParseHostList error:', error);
      return [];
    }
  },

  /**
   * 查询多个节点的拓扑路径
   */
  fetchNodePath: async (params: any) => {
    try {
      return (await IpSelectorService.fetchNodePath(params)) || [];
    } catch (error) {
      console.error('[IpSelector] fetchNodePath error:', error);
      return [];
    }
  },

  /**
   * 获取多个拓扑节点的主机 Agent 状态统计信息
   */
  fetchBatchNodeAgentStatistics: async (params: any) => {
    try {
      return (await IpSelectorService.fetchAgentStatistics(params)) || [];
    } catch (error) {
      console.error('[IpSelector] fetchBatchNodeAgentStatistics error:', error);
      return [];
    }
  },

  // 动态分组（暂不支持）
  fetchDynamicGroup: () => Promise.resolve([]),
  fetchDynamicGroupHost: () => Promise.resolve({ data: [], total: 0 }),
  fetchBatchGroupAgentStatistics: () => Promise.resolve([]),

  // 服务模板（暂不支持）
  fetchServiceTemplates: () => Promise.resolve([]),
  fetchNodesServiceTemplate: () => Promise.resolve([]),
  fetchHostServiceTemplate: () => Promise.resolve({ data: [], total: 0 }),
  fetchHostAgentStatisticsServiceTemplate: () => Promise.resolve([]),

  // 集群模板（暂不支持）
  fetchSetTemplates: () => Promise.resolve([]),
  fetchNodesSetTemplate: () => Promise.resolve([]),
  fetchHostSetTemplate: () => Promise.resolve({ data: [], total: 0 }),
  fetchHostAgentStatisticsSetTemplate: () => Promise.resolve([]),

  // 服务实例（暂不支持）
  fetchSeriviceInstanceList: () => Promise.resolve({ data: [], total: 0 }),
  fetchSeriviceInstanceDetails: () => Promise.resolve([]),

  // DBM白名单（暂不支持）
  fetchDBMWhitelist: () => Promise.resolve([]),

  // 自定义配置
  fetchAll: () => Promise.resolve({}),
  update: (params: any) => Promise.resolve(params),
};

// 创建IP选择器组件 — section 模式（外部自行包裹 Dialog）
const IpSelector = createFactory({
  // 组件版本
  version: '1.0.0',
  // 需要支持的面板：静态拓扑 + 手动输入
  panelList: ['staticTopo', 'manualInput'],
  // 面板选项的值是否唯一
  unqiuePanelValue: false,
  // 字段命名风格
  nameStyle: 'camelCase',
  // 主机列表全选模式
  hostTableDefaultSelectAllMode: false,
  // 主机表格每页条数
  hostTablePageSize: 10,
  // 主机列表显示列
  hostTableRenderColumnList: ['ip', 'ipv6', 'hostName', 'cloudArea', 'osName', 'alive'],
  // 主机预览字段
  hostViewFieldRender: (host: any) => host.host_id,

  // 创建时是否提示 service 信息
  serviceConfigError: false,

  // ========== 数据源配置 ==========
  // 主机拓扑
  fetchTopologyHostCount: Service.fetchTopologyHostCount,
  fetchTopologyHostsNodes: Service.fetchTopologyHost,
  fetchTopologyHostIdsNodes: Service.fetchTopogyHostIdList,
  fetchHostsDetails: Service.fetchHostInfoByHostId,
  fetchHostCheck: Service.fetchInputParseHostList,
  fetchNodesQueryPath: Service.fetchNodePath as any,
  fetchHostAgentStatisticsNodes: Service.fetchBatchNodeAgentStatistics as any,
  // 动态分组
  fetchDynamicGroups: Service.fetchDynamicGroup as any,
  fetchHostsDynamicGroup: Service.fetchDynamicGroupHost as any,
  fetchHostAgentStatisticsDynamicGroups: Service.fetchBatchGroupAgentStatistics as any,
  // 服务模板
  fetchServiceTemplates: Service.fetchServiceTemplates as any,
  fetchNodesServiceTemplate: Service.fetchNodesServiceTemplate as any,
  fetchHostServiceTemplate: Service.fetchHostServiceTemplate as any,
  fetchHostAgentStatisticsServiceTemplate: Service.fetchHostAgentStatisticsServiceTemplate as any,
  // 集群模板
  fetchSetTemplates: Service.fetchSetTemplates as any,
  fetchNodesSetTemplate: Service.fetchNodesSetTemplate as any,
  fetchHostSetTemplate: Service.fetchHostSetTemplate as any,
  fetchHostAgentStatisticsSetTemplate: Service.fetchHostAgentStatisticsSetTemplate as any,
  // 服务实例
  fetchSeriviceInstanceList: Service.fetchSeriviceInstanceList as any,
  fetchSeriviceInstanceDetails: Service.fetchSeriviceInstanceDetails as any,
  // DBM 白名单
  fetchDBMWhitelist: Service.fetchDBMWhitelist as any,
  // 自定义配置
  fetchCustomSettings: Service.fetchAll as any,
  updateCustomSettings: Service.update as any,
  // 系统配置
  fetchConfig: () => Promise.resolve({
    bk_cmdb_dynamic_group_url: '',
    bk_cmdb_static_topo_url: '',
    bk_dbm_whitelist: '',
    bk_cmdb_service_template_url: '',
    bk_user_selector_host: '',
  }),
});

export default IpSelector;
