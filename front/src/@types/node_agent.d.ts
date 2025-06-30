// gen-api.js 自动生成，请勿手动修改
// NodeAgentInstallReq describes the HTTP request body when install node agent.
export interface NodeAgentInstallReq {
  host: Host[];
}

export interface NodeAgentInstallReqHost {
  bk_addressing: string;
  bk_biz_id: number;
  bk_host_innerip: string;
  bk_host_innerip_v6: string;
  login_ip: string;
  login_port: number;
  login_user: string;
  // support: auto, password, keyfile, none
  login_mode: string;
  login_password: string;
  login_key_file: bytes;
  bk_networkunit_id: number;
  os_type: string;
  target_version: string;
  bk_host_id: number;
  re_register: boolean;
}

// NodeAgentInstallResp describes the node agent install response.
export interface NodeAgentInstallResp {
  code: number;
  message: string;
  request_id: string;
  data: NodeAgentInstallRespData;
}

export interface NodeAgentInstallRespData {
  workflow_id: string;
}

