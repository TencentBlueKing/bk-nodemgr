// gen-api.js 自动生成，请勿手动修改
// NodeConstantDeployGetReq describes the HTTP request body when get
// default deploy constant in node service.
export interface NodeConstantDeployGetReq {
  generation: number;
  os_type: string;
}

// NodeConstantDeployGetResp describes the HTTP response body when get
// default deploy constant in node service.
export interface NodeConstantDeployGetResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: NodeConstantDeployGetRespData;
}

export interface NodeConstantDeployGetRespData {
  default_deploy_config: CustomDeployConfig;
}

