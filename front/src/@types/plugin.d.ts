// gen-api.js 自动生成，请勿手动修改
// PluginInstallReq describes the plugin install request.
export interface PluginInstallReq {
  process: Process[];
}

export interface PluginInstallReqProcess {
  bk_host_id: number;
  plugin_name: string;
  version: string;
}

// PluginInstallResp describes the plugin install response.
export interface PluginInstallResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  data: PluginInstallRespData;
}

export interface PluginInstallRespData {
  workflow_id: string;
}

// PluginListReq describes the plugin list request.
export interface PluginListReq {
  page: Page;
  only_count: boolean;
  exact_include_conditions: PluginListReqExactConditions;
  fuzzy_include_conditions: PluginListReqFuzzyConditions;
}

export interface PluginListReqExactConditions {
  plugin_name: string[];
  plugin_group: string[];
}

export interface PluginListReqFuzzyConditions {
  plugin_name: string[];
  plugin_pkg_name: string[];
}

// Plugin describes a plugin.
export interface Plugin {
  tenant_id: string;
  name: string;
  group: string;
  pkg_name: string;
}

// PluginListResp describes the plugin list response.
export interface PluginListResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  data: PluginListRespData;
}

export interface PluginListRespData {
  total: number;
  items: Plugin[];
}

