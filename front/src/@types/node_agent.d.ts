// gen-api.js 自动生成，请勿手动修改
// AgentInstallInfo ...
export interface AgentInstallInfo {
  bk_addressing: string;
  bk_biz_id: number;
  bk_host_innerip: string[];
  bk_host_innerip_v6: string[];
  login_ip: string;
  login_port: number;
  login_user: string;
  // support: password_vault, password, keyfile
  login_mode: string;
  login_password: string;
  login_key_file: string;
  bk_networkunit_id: number;
  os_type: string;
  bk_host_id: number;
  re_register: boolean;
  install_pre_ordered_plugins: boolean;
}

// NodeAgentInstallReq describes the HTTP request body when install node agent.
export interface NodeAgentInstallReq {
  info: AgentInstallInfo[];
  target_version: TargetVersion[];
  is_manual: boolean;
  enable_compatibility_mode: boolean;
}

// NodeAgentInstallResp describes the node agent install response.
export interface NodeAgentInstallResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: NodeAgentInstallRespData;
}

export interface NodeAgentInstallRespData {
  workflow_id: string;
}

export interface NodeAgentUpgradeHost {
  bk_host_id: number;
  target_version: string;
  force: boolean;
  graceful_restart_timeout_sec: number;
  bk_networkunit_id: number;
  cpu_arch: string;
}

// NodeAgentUpgradeReq describes the node agent upgrade request.
export interface NodeAgentUpgradeReq {
  host: NodeAgentUpgradeHost[];
}

// NodeAgentUpgradeResp describes the node agent upgrade response.
export interface NodeAgentUpgradeResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: NodeAgentUpgradeRespData;
}

export interface NodeAgentUpgradeRespData {
  workflow_id: string;
}

export interface NodeAgentReconfigHost {
  bk_host_id: number;
  force: boolean;
  graceful_restart_timeout_sec: number;
}

// NodeAgentReconfigReq describes the node agent reconfigure request.
export interface NodeAgentReconfigReq {
  host: NodeAgentReconfigHost[];
}

// NodeAgentReconfigResp describes the node agent reconfigure response.
export interface NodeAgentReconfigResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: NodeAgentReconfigRespData;
}

export interface NodeAgentReconfigRespData {
  workflow_id: string;
}

export interface NodeAgentRestartHost {
  bk_host_id: number;
  force: boolean;
  graceful_restart_timeout_sec: number;
}

// NodeAgentRestartReq describes the node agent restart request.
export interface NodeAgentRestartReq {
  host: NodeAgentRestartHost[];
}

// NodeAgentRestartResp describes the node agent restart response.
export interface NodeAgentRestartResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: NodeAgentRestartRespData;
}

export interface NodeAgentRestartRespData {
  workflow_id: string;
}

export interface NodeAgentUninstallHost {
  bk_host_id: number;
}

// NodeAgentUninstallReq describes the node agent uninstall request.
export interface NodeAgentUninstallReq {
  host: NodeAgentUninstallHost[];
}

// NodeAgentUninstallResp describes the node agent uninstall response.
export interface NodeAgentUninstallResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: NodeAgentUninstallRespData;
}

export interface NodeAgentUninstallRespData {
  workflow_id: string;
}

// AgentInstallCheckInfo describes the node agent install check parameter.
export interface AgentInstallCheckInfo {
  bk_biz_id: number;
  bk_host_id: number;
  bk_host_innerip_list: string[];
  bk_host_innerip_v6_list: string[];
  bk_networkunit_id: number;
}

// NodeAgentInstallCheckReq describes the node agent install check request.
export interface NodeAgentInstallCheckReq {
  host: AgentInstallCheckInfo[];
}

// NodeAgentInstallCheckResp describes the response for node agent installation
// check.
export interface NodeAgentInstallCheckResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: NodeAgentInstallCheckRespData;
}

export interface NodeAgentInstallCheckRespData {
  results: NodeAgentInstallCheckResult[];
}

// NodeAgentInstallCheckMatchedItem describes the matched item for node.
export interface NodeAgentInstallCheckMatchedItem {
  bk_host_id: number;
  bk_biz_id: number;
  bk_networkarea_id: number;
  bk_networkunit_id: number;
  os_type: string;
  node_role: string;
  bk_host_innerip_list: string[];
  bk_host_innerip_v6_list: string[];
}

// NodeAgentInstallCheckResult describes the check result for node agent
// installation.
export interface NodeAgentInstallCheckResult {
  status: string;
  matched: NodeAgentInstallCheckMatchedItem;
  message_en: string;
  message_zh: string;
  category: string;
}

// AgentUpgradeCheckInfo describes the node agent upgrade check parameter.
export interface AgentUpgradeCheckInfo {
  bk_host_id: number;
  bk_networkunit_id: number;
  cpu_arch: string;
}

// NodeAgentUpgradeCheckReq describes the node agent upgrade check request.
export interface NodeAgentUpgradeCheckReq {
  host: AgentUpgradeCheckInfo[];
  target_version: TargetVersion[];
}

// NodeAgentUpgradeCheckResp describes the response for node agent upgrade
// check.
export interface NodeAgentUpgradeCheckResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: NodeAgentUpgradeCheckRespData;
}

export interface NodeAgentUpgradeCheckRespData {
  results: NodeAgentUpgradeCheckResult[];
}

// NodeAgentUpgradeCheckMatchedItem describes the matched item for upgrade
// check.
export interface NodeAgentUpgradeCheckMatchedItem {
  bk_host_id: number;
  bk_biz_id: number;
  bk_networkarea_id: number;
  bk_networkunit_id: number;
  os_type: string;
  node_role: string;
  bk_host_innerip_list: string[];
  bk_host_innerip_v6_list: string[];
}

// NodeAgentUpgradeCheckResult describes the check result for node agent
// upgrade.
export interface NodeAgentUpgradeCheckResult {
  status: string;
  matched: NodeAgentUpgradeCheckMatchedItem;
  message_en: string;
  message_zh: string;
  category: string;
}

// UploadAgentInstallTemplateReq is the request for upload agent install
// template file.
export interface UploadAgentInstallTemplateReq {
}

// AgentInstallParsedInfo describes the result for agent install template file
// parsed info.
export interface AgentInstallParsedInfo {
  bk_host_innerip: string;
  bk_host_innerip_v6: string;
  os_type: string;
  login_ip: string;
  login_port: number;
  login_user: string;
  login_mode: string;
  credit: string;
  bk_addressing: string;
  install_pre_ordered_plugins: boolean;
  re_register: boolean;
}

// UploadAgentInstallTemplateResp ...
export interface UploadAgentInstallTemplateResp {
  data: UploadAgentInstallTemplateRespData;
}

export interface UploadAgentInstallTemplateRespData {
  info: AgentInstallParsedInfo[];
}

// NodeAgentAssignUnitReq describes the request body for batch-assigning
// a network unit to hosts.
export interface NodeAgentAssignUnitReq {
  bk_host_id: number[];
  bk_networkunit_id: number;
}

// NodeAgentAssignUnitResp describes the response for batch-assigning
// a network unit.
export interface NodeAgentAssignUnitResp {
  data: NodeAgentAssignUnitRespData;
}

export interface NodeAgentAssignUnitRespData {
  success_count: number;
  failed_count: number;
  failed_reasons: string[];
}

