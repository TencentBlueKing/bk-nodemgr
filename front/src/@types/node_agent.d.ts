// gen-api.js 自动生成，请勿手动修改
// AgentInstallInfo ...
export interface AgentInstallInfo {
  bk_addressing: string;
  bk_biz_id: number;
  bk_host_innerip: string;
  bk_host_innerip_v6: string;
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
}

// NodeAgentInstallReq describes the HTTP request body when install node agent.
export interface NodeAgentInstallReq {
  info: AgentInstallInfo[];
  target_version: TargetVersion[];
  disable_default_target_version: boolean;
}

// NodeAgentInstallResp describes the node agent install response.
export interface NodeAgentInstallResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  data: NodeAgentInstallRespData;
}

export interface NodeAgentInstallRespData {
  workflow_id: string;
}

// AgentInstallCheckInfo describes the node agent install check parameter.
export interface AgentInstallCheckInfo {
  bk_biz_id: number;
  bk_host_innerip: string;
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
  data: NodeAgentInstallCheckRespData;
}

export interface NodeAgentInstallCheckRespData {
  install_eligs: NodeAgentInstallElig[];
  total_count: number;
}

// NodeAgentInstallElig describes the eligibility for node agent
// installation.
export interface NodeAgentInstallElig {
  inner_ip: string;
  elig_status: string;
  duplicate_host_ids: number[];
}

// UploadAgentTemplateReq is the request for upload agent tempalte file.
export interface UploadAgentTemplateReq {
}

// ParsedInfo describes the result for tempalte file parsed info.
export interface ParsedInfo {
  inner_ip: string;
  inner_ipv6: string;
  os_type: string;
  login_ip: string;
  login_port: number;
  login_user: string;
  login_mode: string;
  credential: string;
}

// UploadAgentTemplateResp ...
export interface UploadAgentTemplateResp {
  data: UploadAgentTemplateRespData;
}

export interface UploadAgentTemplateRespData {
  info: ParsedInfo[];
  total_count: number;
}

