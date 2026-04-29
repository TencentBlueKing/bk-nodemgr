// gen-api.js 自动生成，请勿手动修改
// NodeProxyInstallHost describes the node proxy install host.
export interface NodeProxyInstallHost {
  bk_biz_id: number;
  bk_networkunit_id: number;
  bk_host_id: number;
  bk_addressing: string;
  bk_host_innerip: string;
  bk_host_innerip_v6: string;
  os_type: string;
  login_ip: string;
  login_port: number;
  login_user: string;
  // support: password_vault, password, keyfile
  login_mode: string;
  login_password: string;
  login_key_file: string;
  export_ip: string;
  export_ip_v6: string;
  advertise_ip: string;
  advertise_ip_v6: string;
  re_register: boolean;
  install_pre_ordered_plugins: boolean;
  proxy_tags: string[];
  proxy_install_origin_unit_id: number;
  // credit expired interval in seconds, default is 86400 (24 hours)
  credit_expired_interval_sec: number;
  relay_download_port: number;
  relay_callback_port: number;
  cpu_arch: string;
}

// NodeProxyInstallReq describes the node proxy install request.
export interface NodeProxyInstallReq {
  host: NodeProxyInstallHost[];
  target_version: TargetVersion[];
  is_manual: boolean;
  is_offline: boolean;
}

// NodeProxyInstallResp describes the node proxy install response.
export interface NodeProxyInstallResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: NodeProxyInstallRespData;
}

export interface NodeProxyInstallRespData {
  workflow_id: string;
}

export interface NodeProxyUpgradeHost {
  bk_host_id: number;
  force: boolean;
  graceful_restart_timeout_sec: number;
  bk_networkunit_id: number;
  cpu_arch: string;
}

// NodeProxyUpgradeReq describes the node proxy upgrade request.
export interface NodeProxyUpgradeReq {
  host: NodeProxyUpgradeHost[];
  target_version: TargetVersion[];
}

// NodeProxyUpgradeResp describes the node proxy upgrade response.
export interface NodeProxyUpgradeResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: NodeProxyUpgradeRespData;
}

export interface NodeProxyUpgradeRespData {
  workflow_id: string;
}

export interface NodeProxyReconfigHost {
  bk_host_id: number;
  force: boolean;
  graceful_restart_timeout_sec: number;
}

// NodeProxyReconfigReq describes the node proxy reconfig request.
export interface NodeProxyReconfigReq {
  host: NodeProxyReconfigHost[];
}

// NodeProxyReconfigResp describes the node proxy reconfig response.
export interface NodeProxyReconfigResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: NodeProxyReconfigRespData;
}

export interface NodeProxyReconfigRespData {
  workflow_id: string;
}

export interface NodeProxyRestartHost {
  bk_host_id: number;
  force: boolean;
  graceful_restart_timeout_sec: number;
}

// NodeProxyRestartReq describes the node proxy restart request.
export interface NodeProxyRestartReq {
  host: NodeProxyRestartHost[];
}

// NodeProxyRestartResp describes the node proxy restart response.
export interface NodeProxyRestartResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: NodeProxyRestartRespData;
}

export interface NodeProxyRestartRespData {
  workflow_id: string;
}

// NodeProxyUpdateHost describes the node proxy update host.
export interface NodeProxyUpdateHost {
  bk_host_id: number;
  login_ip: string;
  login_port: number;
  login_user: string;
  // support: password_vault, password, keyfile
  login_mode: string;
  login_password: string;
  login_key_file: string;
  export_ip: string;
  export_ip_v6: string;
  advertise_ip: string;
  advertise_ip_v6: string;
  proxy_tags: string[];
  relay_download_port: number;
  relay_callback_port: number;
}

export interface NodeProxyUninstallHost {
  bk_host_id: number;
}

// NodeProxyUninstallReq describes the node proxy uninstall request.
export interface NodeProxyUninstallReq {
  host: NodeProxyUninstallHost[];
}

// NodeProxyUninstallResp describes the node proxy uninstall response.
export interface NodeProxyUninstallResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: NodeProxyUninstallRespData;
}

export interface NodeProxyUninstallRespData {
  workflow_id: string;
}

// NodeProxyUpdateReq describes the node proxy update request.
export interface NodeProxyUpdateReq {
  Host: NodeProxyUpdateHost[];
}

// NodeProxyUpdateResp describes the node proxy update response.
export interface NodeProxyUpdateResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: NodeProxyUpdateRespData;
}

export interface NodeProxyUpdateRespData {
}

// ProxyInstallCheckInfo describes the node proxy install check parameter.
export interface ProxyInstallCheckInfo {
  bk_biz_id: number;
  bk_host_id: number;
  bk_host_innerip_list: string[];
  bk_host_innerip_v6_list: string[];
  bk_networkunit_id: number;
}

// NodeProxyInstallCheckReq describes the node proxy install check request.
export interface NodeProxyInstallCheckReq {
  host: ProxyInstallCheckInfo[];
}

// NodeProxyInstallCheckResp describes the response for node proxy installation
// check.
export interface NodeProxyInstallCheckResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: NodeProxyInstallCheckRespData;
}

export interface NodeProxyInstallCheckRespData {
  results: NodeProxyInstallCheckResult[];
}

// NodeProxyInstallCheckMatchedItem describes the matched item for proxy node.
export interface NodeProxyInstallCheckMatchedItem {
  bk_host_id: number;
  bk_biz_id: number;
  bk_networkarea_id: number;
  bk_networkunit_id: number;
  os_type: string;
  node_role: string;
  bk_host_innerip_list: string[];
  bk_host_innerip_v6_list: string[];
}

// NodeProxyInstallCheckResult describes the check result for node proxy
// installation.
export interface NodeProxyInstallCheckResult {
  status: string;
  matched: NodeProxyInstallCheckMatchedItem;
  message_en: string;
  message_zh: string;
  category: string;
}

// ProxyUpgradeCheckInfo describes the node proxy upgrade check parameter.
export interface ProxyUpgradeCheckInfo {
  bk_host_id: number;
  bk_networkunit_id: number;
  cpu_arch: string;
}

// NodeProxyUpgradeCheckReq describes the node proxy upgrade check request.
export interface NodeProxyUpgradeCheckReq {
  host: ProxyUpgradeCheckInfo[];
  target_version: TargetVersion[];
}

// NodeProxyUpgradeCheckResp describes the response for node proxy upgrade
// check.
export interface NodeProxyUpgradeCheckResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: NodeProxyUpgradeCheckRespData;
}

export interface NodeProxyUpgradeCheckRespData {
  results: NodeProxyUpgradeCheckResult[];
}

// NodeProxyUpgradeCheckMatchedItem describes the matched item for proxy upgrade
// check.
export interface NodeProxyUpgradeCheckMatchedItem {
  bk_host_id: number;
  bk_biz_id: number;
  bk_networkarea_id: number;
  bk_networkunit_id: number;
  os_type: string;
  node_role: string;
  bk_host_innerip_list: string[];
  bk_host_innerip_v6_list: string[];
}

// NodeProxyUpgradeCheckResult describes the check result for node proxy
// upgrade.
export interface NodeProxyUpgradeCheckResult {
  status: string;
  matched: NodeProxyUpgradeCheckMatchedItem;
  message_en: string;
  message_zh: string;
  category: string;
}

// NodeProxyAssignUnitReq describes the request body for batch-assigning
// a network unit to proxy hosts.
export interface NodeProxyAssignUnitReq {
  bk_host_id: number[];
  bk_networkunit_id: number;
}

// NodeProxyAssignUnitResp describes the response for batch-assigning
// a network unit to proxy hosts.
export interface NodeProxyAssignUnitResp {
  data: NodeProxyAssignUnitRespData;
}

export interface NodeProxyAssignUnitRespData {
  success_count: number;
  failed_count: number;
  failed_reasons: string[];
  workflow_id: string;
}

