// gen-api.js 自动生成，请勿手动修改
// PluginInstallReq describes the plugin install request.
export interface PluginInstallReq {
  plugin: Plugin[];
}

export interface PluginInstallReqPlugin {
  bk_host_id: number;
  plugin_name: string;
  version: string;
  config_name: string[];
  custom_config_context: Record<string, any>;
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

// PluginUpgradeReq describes the plugin upgrade request.
export interface PluginUpgradeReq {
  plugin: Plugin[];
}

export interface PluginUpgradeReqPlugin {
  bk_host_id: number;
  plugin_name: string;
  version: string;
  config_name: string[];
  custom_config_context: Record<string, any>;
}

// PluginUpgradeResp describes the plugin upgrade response.
export interface PluginUpgradeResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  data: PluginUpgradeRespData;
}

export interface PluginUpgradeRespData {
  workflow_id: string;
}

// PluginApplySubConfigReq describes the plugin apply sub-configuration
// request.
export interface PluginApplySubConfigReq {
  plugin: Plugin[];
}

export interface PluginApplySubConfigReqPlugin {
  bk_host_id: number;
  plugin_name: string;
  version: string;
  config_name: string[];
  custom_config_context: Record<string, any>;
}

// PluginApplySubConfigResp describes the plugin apply sub-configuration
// response.
export interface PluginApplySubConfigResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  data: PluginApplySubConfigRespData;
}

export interface PluginApplySubConfigRespData {
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
  name: string[];
  group: string[];
}

export interface PluginListReqFuzzyConditions {
  name: string[];
  pkg_name: string[];
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

// PluginSetMemoReq describes the plugin set memo request.
export interface PluginSetMemoReq {
  plugin_name: string;
  memo: string;
}

// PluginSetMemoResp describes the plugin set memo response.
export interface PluginSetMemoResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
}

