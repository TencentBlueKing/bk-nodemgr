// gen-api.js 自动生成，请勿手动修改
// ConfigPolicyExactConditions describes config policy exact conditions.
export interface ConfigPolicyExactConditions {
  configpolicy_id: number[];
  bk_biz_id: number[];
  configpolicy_type: string[];
  enabled: boolean[];
}

// ConfigPolicyFuzzyConditions describes config policy fuzzy conditions.
export interface ConfigPolicyFuzzyConditions {
  configpolicy_name: string[];
  operator: string[];
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
  error: Error;
  permission: Permission;
  data: ConfigPolicyListRespData;
}

export interface ConfigPolicyListRespData {
  total: number;
  items: ConfigPolicy[];
}

// ConfigPolicyGetTemplateReq describes HTTP request body when get config policy
// template.
export interface ConfigPolicyGetTemplateReq {
  configpolicy_type: string;
}

// ConfigPolicyGetTemplateResp describes HTTP response body when get config
// policy template.
export interface ConfigPolicyGetTemplateResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: ConfigPolicyGetTemplateRespData;
}

export interface ConfigPolicyGetTemplateRespData {
  templates: ConfigPolicyConfigBlock[];
}

// ConfigPolicyListPlatformReq describes HTTP request body when list config
// policy platform.
export interface ConfigPolicyListPlatformReq {
  configpolicy_type: string;
  generation: number;
}

// ConfigPolicyListPlatformResp describes HTTP response body when list config
// policy platform.
export interface ConfigPolicyListPlatformResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
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
  configpolicy_type: string;
  bk_biz_id: number;
  remark: string;
  scopes: ConfigPolicyScope[];
  configs: ConfigPolicyConfigBlock[];
  operator: string;
  target_host_ids: number[];
}

// ConfigPolicyCreateResp describes HTTP response body when create config
// policy.
export interface ConfigPolicyCreateResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: ConfigPolicyCreateRespData;
}

export interface ConfigPolicyCreateRespData {
  configpolicy_id: number;
}

// ConfigPolicyUpdateReq describes HTTP request body when update config policy.
export interface ConfigPolicyUpdateReq {
  configpolicy_id: number;
  configpolicy_name: string;
  configpolicy_type: string;
  bk_biz_id: number;
  remark: string;
  scopes: ConfigPolicyScope[];
  configs: ConfigPolicyConfigBlock[];
  enabled: boolean;
  operator: string;
  priority: number;
  target_host_ids: number[];
}

// ConfigPolicyUpdateResp describes HTTP response body when update config
// policy.
export interface ConfigPolicyUpdateResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
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
  error: Error;
  permission: Permission;
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
  error: Error;
  permission: Permission;
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
  error: Error;
  permission: Permission;
  data: ConfigPolicyDeleteRespData;
}

export interface ConfigPolicyDeleteRespData {
}

// ConfigPolicyPriorityReorderReq describes HTTP request body when reordering
// config policy priorities within a single (biz, type) scope.
// Listed IDs are assigned priority 1..N; remaining enabled policies in the
// same (biz, type) preserve their relative order starting from N+1.
export interface ConfigPolicyPriorityReorderReq {
  bk_biz_id: number;
  configpolicy_type: string;
  ordered_configpolicy_id: number[];
}

// ConfigPolicyPriorityReorderResp describes HTTP response body when
// reordering config policy priorities.
export interface ConfigPolicyPriorityReorderResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: ConfigPolicyPriorityReorderRespData;
}

export interface ConfigPolicyPriorityReorderRespData {
}

// PreviewHost carries a host ID with optional attribute overrides for preview.
export interface PreviewHost {
  bk_host_id: number;
  bk_networkunit_id: number;
  bk_networkarea_id: number;
  os_type: string;
  cpu_arch: string;
}

// ConfigPolicyPreviewReq describes HTTP request body for previewing merged
// config result by host list.
export interface ConfigPolicyPreviewReq {
  bk_biz_id: number;
  policy_type: string;
  hosts: PreviewHost[];
}

// ConfigPolicyPreviewResp describes HTTP response body for preview result.
export interface ConfigPolicyPreviewResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: ConfigPolicyPreviewRespData;
}

export interface ConfigPolicyPreviewRespMatchedPolicy {
  configpolicy_id: number;
  configpolicy_name: string;
  priority: number;
}

export interface ConfigPolicyPreviewRespPreviewItem {
  bk_host_id: number;
  matched_policies: MatchedPolicy[];
  merged_configs_string: Record<string, string>;
  merged_configs_int: Record<string, number>;
  merged_configs_bool: Record<string, boolean>;
}

export interface ConfigPolicyPreviewRespData {
  reliable_items: PreviewItem[];
  unreliable_items: PreviewItem[];
}

// ConfigPolicyEventExactConditions describes the conditions when list event
export interface ConfigPolicyEventExactConditions {
  configpolicy_id: number[];
  configpolicy_type: string[];
  version: number[];
  type: string[];
  bk_biz_id: number[];
}

// ConfigPolicyEventFuzzyConditions describes the conditions when list event
export interface ConfigPolicyEventFuzzyConditions {
  configpolicy_name: string[];
  operator: string[];
}

// ConfigPolicyEventListReq describes the HTTP request body when list event in
// policy service.
export interface ConfigPolicyEventListReq {
  page: Page;
  only_count: boolean;
  exact_include_conditions: ConfigPolicyEventExactConditions;
  fuzzy_include_conditions: ConfigPolicyEventFuzzyConditions;
  operate_time_range: TimeRange;
}

// ConfigPolicyEventListResp describes the HTTP response body when list event in
// policy service.
export interface ConfigPolicyEventListResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: ConfigPolicyEventListRespData;
}

export interface ConfigPolicyEventListRespData {
  total: number;
  items: ConfigPolicyEvent[];
}

// ConfigPolicyEventDistinctReq describes the HTTP request body when distinct
// policyevent in policy service.
export interface ConfigPolicyEventDistinctReq {
  exact_include_conditions: ConfigPolicyEventExactConditions;
  fuzzy_include_conditions: ConfigPolicyEventFuzzyConditions;
  operate_time_range: TimeRange;
}

// ConfigPolicyEventDistinctResp describes the HTTP response body when distinct
// policyevent in policy service.
export interface ConfigPolicyEventDistinctResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: ConfigPolicyEventDistinctRespData;
}

export interface ConfigPolicyEventDistinctRespData {
  configpolicy_id: number[];
  configpolicy_name: string[];
  configpolicy_type: string[];
  type: string[];
  version: number[];
  operator: string[];
  bk_biz_id: number[];
}

