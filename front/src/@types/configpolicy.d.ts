// gen-api.js 自动生成，请勿手动修改
// ConfigPolicyExactConditions describes config policy exact conditions.
export interface ConfigPolicyExactConditions {
  configpolicy_id: number[];
  biz_id: number[];
  node_role: string[];
  enabled: boolean[];
}

// ConfigPolicyFuzzyConditions describes config policy fuzzy conditions.
export interface ConfigPolicyFuzzyConditions {
  configpolicy_name: string[];
}

// ConfigPolicyListReq describes HTTP request body when list config policy.
export interface ConfigPolicyListReq {
  page: Page;
  only_count: boolean;
  exact_include_conditions: ConfigPolicyExactConditions;
  fuzzy_include_conditions: ConfigPolicyFuzzyConditions;
}

// ConfigPolicyListResp describes HTTP response body when list config policy.
export interface ConfigPolicyListResp {
  code: number;
  message: string;
  request_id: string;
  data: ConfigPolicyListRespData;
}

export interface ConfigPolicyListRespData {
  total: number;
  items: ConfigPolicy[];
}

// ConfigPolicyGetTemplateReq describes HTTP request body when get config policy
// template.
export interface ConfigPolicyGetTemplateReq {
  node_role: string;
}

// ConfigPolicyGetTemplateResp describes HTTP response body when get config
// policy template.
export interface ConfigPolicyGetTemplateResp {
  code: number;
  message: string;
  request_id: string;
  data: ConfigPolicyGetTemplateRespData;
}

export interface ConfigPolicyGetTemplateRespData {
  templates: ConfigPolicyConfigBlock[];
}

// ConfigPolicyListPlatformReq describes HTTP request body when list config
// policy platform.
export interface ConfigPolicyListPlatformReq {
  node_role: string;
}

// ConfigPolicyListPlatformResp describes HTTP response body when list config
// policy platform.
export interface ConfigPolicyListPlatformResp {
  code: number;
  message: string;
  request_id: string;
  data: ConfigPolicyListPlatformRespData;
}

export interface ConfigPolicyListPlatformRespData {
  os_type: string[];
  cpu_arch: string[];
}

// ConfigPolicyGetReq describes HTTP request body when get config policy.
export interface ConfigPolicyGetReq {
  configpolicy_id: number;
}

// ConfigPolicyGetResp describes HTTP response body when get config policy.
export interface ConfigPolicyGetResp {
  code: number;
  message: string;
  request_id: string;
  data: ConfigPolicy;
}

// ConfigPolicyCreateReq describes HTTP request body when create config policy.
export interface ConfigPolicyCreateReq {
  configpolicy_name: string;
  node_role: string;
  biz_id: number[];
  remark: string;
  scopes: ConfigPolicyScope[];
  configs: ConfigPolicyConfigBlock[];
  operator: string;
}

// ConfigPolicyCreateResp describes HTTP response body when create config
// policy.
export interface ConfigPolicyCreateResp {
  code: number;
  message: string;
  request_id: string;
  data: ConfigPolicyCreateRespData;
}

export interface ConfigPolicyCreateRespData {
  configpolicy_id: number;
}

// ConfigPolicyUpdateReq describes HTTP request body when update config policy.
export interface ConfigPolicyUpdateReq {
  configpolicy_id: number;
  configpolicy_name: string;
  node_role: string;
  biz_id: number[];
  remark: string;
  scopes: ConfigPolicyScope[];
  configs: ConfigPolicyConfigBlock[];
  enabled: boolean;
  operator: string;
}

// ConfigPolicyUpdateResp describes HTTP response body when update config
// policy.
export interface ConfigPolicyUpdateResp {
  code: number;
  message: string;
  request_id: string;
  data: ConfigPolicyUpdateRespData;
}

export interface ConfigPolicyUpdateRespData {
  configpolicy_id: number;
}

// ConfigPolicyEnableReq describes HTTP request body when enable config policy.
export interface ConfigPolicyEnableReq {
  configpolicy_id: number[];
}

// ConfigPolicyEnableResp describes HTTP response body when enable config
// policy.
export interface ConfigPolicyEnableResp {
  code: number;
  message: string;
  request_id: string;
  data: ConfigPolicyEnableRespData;
}

export interface ConfigPolicyEnableRespData {
}

// ConfigPolicyDisableReq describes HTTP request body when disable config
// policy.
export interface ConfigPolicyDisableReq {
  configpolicy_id: number[];
}

// ConfigPolicyDisableResp describes HTTP response body when disable config
// policy.
export interface ConfigPolicyDisableResp {
  code: number;
  message: string;
  request_id: string;
  data: ConfigPolicyDisableRespData;
}

export interface ConfigPolicyDisableRespData {
}

// ConfigPolicyDeleteReq describes HTTP request body when delete config policy.
export interface ConfigPolicyDeleteReq {
  configpolicy_id: number[];
}

// ConfigPolicyDeleteResp describes HTTP response body when delete config
// policy.
export interface ConfigPolicyDeleteResp {
  code: number;
  message: string;
  request_id: string;
  data: ConfigPolicyDeleteRespData;
}

export interface ConfigPolicyDeleteRespData {
}

