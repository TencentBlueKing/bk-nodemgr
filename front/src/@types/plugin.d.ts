// gen-api.js 自动生成，请勿手动修改
export interface PluginOperateFullInfo {
  bk_host_id: number;
  plugin_name: string;
  version: string;
  config_name: string[];
  custom_config_context: Record<string, any>;
}

export interface PluginOperateBasicInfo {
  bk_host_id: number;
  plugin_name: string;
}

// PluginInstallReq describes the plugin install request.
export interface PluginInstallReq {
  plugin: PluginOperateFullInfo[];
}

// PluginInstallResp describes the plugin install response.
export interface PluginInstallResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: PluginInstallRespData;
}

export interface PluginInstallRespData {
  workflow_id: string;
}

// PluginUpgradeReq describes the plugin upgrade request.
export interface PluginUpgradeReq {
  plugin: PluginOperateFullInfo[];
}

// PluginUpgradeResp describes the plugin upgrade response.
export interface PluginUpgradeResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: PluginUpgradeRespData;
}

export interface PluginUpgradeRespData {
  workflow_id: string;
}

// PluginUninstallReq describes the plugin uninstall request.
export interface PluginUninstallReq {
  plugin: PluginOperateBasicInfo[];
}

// PluginUninstallResp describes the plugin uninstall response.
export interface PluginUninstallResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: PluginUninstallRespData;
}

export interface PluginUninstallRespData {
  workflow_id: string;
}

// PluginApplySubConfigReq describes the plugin apply sub-configuration
// request.
export interface PluginApplySubConfigReq {
  plugin: PluginOperateFullInfo[];
}

// PluginApplySubConfigResp describes the plugin apply sub-configuration
// response.
export interface PluginApplySubConfigResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: PluginApplySubConfigRespData;
}

export interface PluginApplySubConfigRespData {
  workflow_id: string;
}

export interface PluginListExactConditions {
  name: string[];
  group: string[];
}

export interface PluginListFuzzyConditions {
  name: string[];
  pkg_name: string[];
}

// PluginListReq describes the plugin list request.
export interface PluginListReq {
  page: Page;
  only_count: boolean;
  exact_include_conditions: PluginListExactConditions;
  fuzzy_include_conditions: PluginListFuzzyConditions;
}

// PluginListResp describes the plugin list response.
export interface PluginListResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
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
  permission: Permission;
}

// PluginListPermittedOperationReq describes the plugin list permitted operation
// request.
export interface PluginListPermittedOperationReq {
  page: Page;
}

// PluginListPermittedOperationResp describes the plugin list permitted
// operation response.
export interface PluginListPermittedOperationResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: PluginListPermittedOperationRespData;
}

export interface PluginListPermittedOperationRespData {
  operations: Operation[];
}

export interface DataOperation {
  name: string;
  permission: string[];
}

