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
  advertise_ip: string;
  re_register: boolean;
  proxy_tags: string[];
  proxy_install_origin_unit_id: number;
}

// NodeProxyInstallReq describes the node proxy install request.
export interface NodeProxyInstallReq {
  host: NodeProxyInstallHost[];
  target_version: TargetVersion[];
}

// NodeProxyInstallResp describes the node proxy install response.
export interface NodeProxyInstallResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  data: NodeProxyInstallRespData;
}

export interface NodeProxyInstallRespData {
  workflow_id: string;
}

export interface NodeProxyUpgradeHost {
  bk_host_id: number;
  force: boolean;
  graceful_restart_timeout_sec: number;
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
  advertise_ip: string;
  proxy_tags: string[];
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
  data: NodeProxyUpdateRespData;
}

export interface NodeProxyUpdateRespData {
}

