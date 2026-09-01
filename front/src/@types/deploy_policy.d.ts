// gen-api.js 自动生成，请勿手动修改
// DeployPolicyMeta describes the meta of deploy policy.
export interface DeployPolicyMeta {
  name: string;
  description: string;
}

// PluginConfigDetail describes the plugin config detail.
export interface PluginConfigDetail {
  name: string;
  content: string;
  is_main_config: boolean;
}

// SpecifyAgentParam describes the specify agent param structure.
export interface SpecifyAgentParam {
  node_version: string;
}

// SpecifyProxyParam describes the specify proxy param structure.
export interface SpecifyProxyParam {
  node_version: string;
}

// SpecifyPluginParam describes the specify plugin param structure.
export interface SpecifyPluginParam {
  plugin_name: string;
  version: string;
  custom_config_context: Record<string, any>;
}

// SpecifyPluginPkgParam describes the specify plugin pkg param structure.
export interface SpecifyPluginPkgParam {
  plugin_pkg_name: string;
  version: string;
  custom_config_context: Record<string, any>;
}

// SpecifyPluginSubConfigParam describes the specify plugin sub config param
// structure.
export interface SpecifyPluginSubConfigParam {
  plugin_name: string;
  config_files_detail: PluginConfigDetail[];
  custom_config_context: Record<string, any>;
}

// DeploySpec describes the spec of deploy policy.
export interface DeploySpec {
  type: string;
  param: Record<string, any>;
}

// TargetFilter describes the target filter of deploy policy.
export interface TargetFilter {
}

// ScopeServiceTemplate describes the scope when type is "service_template".
export interface ScopeServiceTemplate {
  granularity: string;
  bk_biz_id: number;
  filter: TargetFilter;
  service_template_ids: number[];
  module_ids: number[];
}

// ScopeSetTemplate describes the scope when type is "set_template".
export interface ScopeSetTemplate {
  granularity: string;
  bk_biz_id: number;
  filter: TargetFilter;
  set_template_ids: number[];
  set_ids: number[];
}

// ScopeInstance describes the scope when type is "instance".
export interface ScopeInstance {
  granularity: string;
  bk_biz_id: number;
  filter: TargetFilter;
  instance_ids: number[];
}

// ScopeTopo describes the scope when type is "topo".
export interface ScopeTopo {
  granularity: string;
  bk_biz_id: number;
  filter: TargetFilter;
  paths: ScopeTopoNode[];
}

export interface ScopeTopoScopeTopoNode {
  topo_obj_id: string;
  topo_inst_id: number;
}

// ScopeDynamicGroup describes the scope when type is "dynamic_group".
export interface ScopeDynamicGroup {
  granularity: string;
  bk_biz_id: number;
  filter: TargetFilter;
  dynamic_group_ids: string[];
}

// Scope describes the scope of deploy policy.
export interface Scope {
  type: string;
  scope: Record<string, any>;
}

// DeployPolicy describes the deploy policy.
export interface DeployPolicy {
  deploy_policy_id: number;
  dsu_id: number;
  meta: DeployPolicyMeta;
  specs: DeploySpec[];
  scopes: Scope[];
  operator: string;
  enabled: boolean;
}

// DeployPolicyExactConditions describes exact conditions of deploy policy.
export interface DeployPolicyExactConditions {
  deploy_policy_id: number[];
  dsu_id: number[];
  deploy_policy_name: string[];
  operator: string[];
  enabled: boolean[];
}

// DeployPolicyFuzzyConditions describes fuzzy conditions of deploy policy.
export interface DeployPolicyFuzzyConditions {
  deploy_policy_name: string[];
  operator: string[];
}

// DeployPolicyListReq describes HTTP request body when list deploy policy.
export interface DeployPolicyListReq {
  page: Page;
  only_count: boolean;
  exact_include_conditions: DeployPolicyExactConditions;
  fuzzy_include_conditions: DeployPolicyFuzzyConditions;
  exact_exclude_conditions: DeployPolicyExactConditions;
  fuzzy_exclude_conditions: DeployPolicyFuzzyConditions;
  executed_time_range: TimeRange;
}

// DeployPolicyListResp describes HTTP response body when list deploy policy.
export interface DeployPolicyListResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: DeployPolicyListRespData;
}

export interface DeployPolicyListRespData {
  total: number;
  items: DeployPolicy[];
}

// DeployPolicyExecuteReq describes HTTP request body when execute deploy
// policy.
export interface DeployPolicyExecuteReq {
  deploy_policy_id: number;
}

// DeployPolicyExecuteResp describes HTTP response body when execute deploy
// policy.
export interface DeployPolicyExecuteResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: DeployPolicyExecuteRespData;
}

export interface DeployPolicyExecuteRespData {
  trigger_id: string;
}

