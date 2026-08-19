// gen-api.js 自动生成，请勿手动修改
// PluginInstallReq describes the plugin install request.
export interface PluginInstallReq {
  plugin: InstallInfo[];
}

export interface PluginInstallReqDeploymentSpec {
  resource: ProcessResource;
  monitor_policy: ProcessMonitorPolicy;
}

export interface PluginInstallReqInstallInfo {
  bk_host_id: number;
  plugin_name: string;
  version: string;
  config_template_name: string[];
  custom_config_context: Record<string, any>;
  custom_spec: DeploymentSpec;
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
  plugin: UpgradeInfo[];
}

export interface PluginUpgradeReqUpgradeInfo {
  bk_host_id: number;
  plugin_name: string;
  version: string;
  config_template_name: string[];
  custom_config_context: Record<string, any>;
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
  plugin: UninstallInfo[];
}

export interface PluginUninstallReqUninstallInfo {
  bk_host_id: number;
  plugin_name: string;
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
  plugin: ApplySubConfigInfo[];
}

export interface PluginApplySubConfigReqApplySubConfigInfo {
  bk_host_id: number;
  plugin_name: string;
  config_template_name: string[];
  custom_config_context: Record<string, any>;
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
  visible_biz_ids: number[];
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

// PluginRestartReq describes the plugin restart request.
export interface PluginRestartReq {
  plugin: RestartInfo[];
}

export interface PluginRestartReqRestartInfo {
  bk_host_id: number;
  plugin_name: string;
}

// PluginStartReq describes the plugin start request.
export interface PluginStartReq {
  plugin: StartInfo[];
}

export interface PluginStartReqStartInfo {
  bk_host_id: number;
  plugin_name: string;
}

// PluginStartResp describes the plugin start response.
export interface PluginStartResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: PluginStartRespData;
}

export interface PluginStartRespData {
  workflow_id: string;
}

// PluginRestartResp describes the plugin restart response.
export interface PluginRestartResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: PluginRestartRespData;
}

export interface PluginRestartRespData {
  workflow_id: string;
}

// PluginStopReq describes the plugin stop request.
export interface PluginStopReq {
  plugin: StopInfo[];
}

export interface PluginStopReqStopInfo {
  bk_host_id: number;
  plugin_name: string;
}

// PluginStopResp describes the plugin stop response.
export interface PluginStopResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: PluginStopRespData;
}

export interface PluginStopRespData {
  workflow_id: string;
}

