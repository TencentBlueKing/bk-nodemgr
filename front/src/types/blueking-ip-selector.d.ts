/**
 * @blueking/ip-selector Vue 3 类型声明
 * 为第三方库提供基础类型支持
 */

declare module '@blueking/ip-selector/dist/vue3.x' {
  import type { Component } from 'vue';

  /**
   * IP 选择器配置选项
   */
  interface IpSelectorConfig {
    version?: string;
    panelList?: string[];
    unqiuePanelValue?: boolean;
    nameStyle?: 'camelCase' | 'kebabCase' | 'PascalCase';
    hostTableDefaultSelectAllMode?: boolean;
    hostTableRenderColumnList?: string[];
    serviceConfigError?: boolean;

    // 拓扑相关
    fetchTopologyHostCount?: (node?: any) => Promise<{ data: any[] }>;
    fetchTopologyHostsNodes?: (params: any) => Promise<{ data: any[]; total: number }>;
    fetchTopologyHostIdsNodes?: (params: any) => Promise<{ data: any[] }>;
    fetchHostsDetails?: (params: any) => Promise<{ data: any[] }>;
    fetchHostCheck?: (params: any) => Promise<{ data: any }>;
    fetchNodesQueryPath?: (params: any) => Promise<{ data: any[] }>;
    fetchHostAgentStatisticsNodes?: (params: any) => Promise<{ data: any[] }>;

    // 动态分组相关
    fetchDynamicGroups?: (params: any) => Promise<{ data: any[] }>;
    fetchHostsDynamicGroup?: (params: any) => Promise<{ data: any[] }>;
    fetchHostAgentStatisticsDynamicGroups?: (params: any) => Promise<{ data: any[] }>;

    // 服务模板相关
    fetchServiceTemplates?: (params: any) => Promise<{ data: any[] }>;
    fetchNodesServiceTemplate?: (params: any) => Promise<{ data: any[] }>;
    fetchHostServiceTemplate?: (params: any) => Promise<{ data: any[] }>;
    fetchHostAgentStatisticsServiceTemplate?: (params: any) => Promise<{ data: any[] }>;

    // 集群模板相关
    fetchSetTemplates?: (params: any) => Promise<{ data: any[] }>;
    fetchNodesSetTemplate?: (params: any) => Promise<{ data: any[] }>;
    fetchHostSetTemplate?: (params: any) => Promise<{ data: any[] }>;
    fetchHostAgentStatisticsSetTemplate?: (params: any) => Promise<{ data: any[] }>;

    // 服务实例相关
    fetchSeriviceInstanceList?: (params: any) => Promise<{ data: any[] }>;
    fetchSeriviceInstanceDetails?: (params: any) => Promise<{ data: any[] }>;

    // DBM 白名单
    fetchDBMWhitelist?: (params: any) => Promise<{ data: any[] }>;

    // 自定义配置
    fetchCustomSettings?: () => Promise<{ data: any }>;
    updateCustomSettings?: (params: any) => Promise<{ data: any }>;
    fetchConfig?: () => Promise<any>;

    [key: string]: any;
  }

  /**
   * IP 选择器实例
   */
  interface IpSelectorInstance {
    IpSelector: Component;
    IpSelectorValue: Component;
    IpSelectorTable: Component;
    [key: string]: any;
  }

  /**
   * 创建 IP 选择器工厂函数
   */
  function createFactory(config: IpSelectorConfig): IpSelectorInstance;

  export default createFactory;
  export { createFactory, IpSelectorConfig, IpSelectorInstance };
}

declare module '@blueking/ip-selector/dist/styles/index.css' {
  const content: string;
  export default content;
}
