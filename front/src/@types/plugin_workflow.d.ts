// gen-api.js 自动生成，请勿手动修改
// PluginWorkflowInfo describes the plugin workflow information.
export interface PluginWorkflowInfo {
  tenant_id: string;
  workflow_id: string;
  trigger_id: string;
  type: string;
  bk_host_id: number[];
  operator: string;
  operate_time: number;
  finish_time: number;
  status: string;
}

// PluginWorkflowExactConditions describes the exact conditions of plugin
// workflow list request.
export interface PluginWorkflowExactConditions {
  workflow_id: string[];
  type: string[];
  bk_host_id: number[];
  status: string[];
  operator: string[];
}

// PluginWorkflowFuzzyConditions describes the fuzzy conditions of plugin
// workflow list request.
export interface PluginWorkflowFuzzyConditions {
}

// PluginWorkflowListReq describes the plugin workflow list request.
export interface PluginWorkflowListReq {
  page: Page;
  only_count: boolean;
  exact_include_conditions: PluginWorkflowExactConditions;
  fuzzy_include_conditions: PluginWorkflowFuzzyConditions;
  operate_time_range: TimeRange;
}

// PluginWorkflowListResp describes the plugin workflow list response.
export interface PluginWorkflowListResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  data: PluginWorkflowListRespData;
}

export interface PluginWorkflowListRespData {
  total: number;
  items: PluginWorkflowInfo[];
}

// PluginWorkflowStatisticReq describes the node workflow statistic request.
export interface PluginWorkflowStatisticsReq {
  workflow_id: string[];
}

// PluginWorkflowStatisticResp describes the node workflow statistic response.
export interface PluginWorkflowStatisticsResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  data: PluginWorkflowStatisticsRespData;
}

export interface PluginWorkflowStatisticsRespData {
  items: WorkflowStatisticsInfo[];
}

// PluginWorkflowDistinctReq describes the plugin workflow distinct request.
export interface PluginWorkflowDistinctReq {
  exact_include_conditions: PluginWorkflowExactConditions;
  fuzzy_include_conditions: PluginWorkflowFuzzyConditions;
}

// PluginWorkflowDistinctResp describes the plugin workflow distinct
export interface PluginWorkflowDistinctResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  data: PluginWorkflowDistinctRespData;
}

export interface PluginWorkflowDistinctRespData {
  type: string[];
  bk_host_id: number[];
  operator: string[];
  status: string[];
}

// PluginWorkflowOperationExactConditions describes the exact conditions of
// plugin workflow operation list request.
export interface PluginWorkflowOperationExactConditions {
  bk_host_id: number[];
  plugin_name: string[];
  plugin_version: string[];
  state: string[];
}

// PluginWorkflowOperationFuzzyConditions describes the fuzzy conditions of
// plugin workflow operation list request.
export interface PluginWorkflowOperationFuzzyConditions {
}

// PluginWorkflowOperationListReq describes the node operation list request.
export interface PluginWorkflowOperationListReq {
  only_count: boolean;
  page: Page;
  workflow_id: string;
  exact_include_conditions: PluginWorkflowOperationExactConditions;
  fuzzy_include_conditions: PluginWorkflowOperationFuzzyConditions;
}

// PluginWorkflowOperationDistinctReq describes the plugin workflow operation
// distinct request.
export interface PluginWorkflowOperationDistinctReq {
  workflow_id: string;
}

// PluginWorkflowOperationDistinctResp describes the plugin workflow operation
// distinct response.
export interface PluginWorkflowOperationDistinctResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  data: PluginWorkflowOperationDistinctRespData;
}

export interface PluginWorkflowOperationDistinctRespData {
  state: string[];
}

// PluginDeploymentInfo describes the plugin deployment information.
export interface PluginDeploymentInfo {
  bk_host_id: number;
  bk_biz_id: number;
  bk_networkarea_id: number;
  bk_networkunit_id: number;
  bk_host_innerip_list: string[];
  bk_host_innerip_v6_list: string[];
  plugin_name: string;
  plugin_version: string;
}

// PluginWorkflowOperation describes the node workflow operation.
export interface PluginWorkflowOperation {
  operation_id: string;
  instance_ids: string[];
  operator: string;
  create_time: number;
  plugin_deployment_info: PluginDeploymentInfo;
  latest_oper_inst_brief_data: WorkflowOperInstBriefData;
}

// PluginWorkflowOperationListResp describes the node operation list by
// workflow-id response.
export interface PluginWorkflowOperationListResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  data: PluginWorkflowOperationListRespData;
}

export interface PluginWorkflowOperationListRespData {
  operations: PluginWorkflowOperation[];
  total_count: number;
}

// PluginWorkflowOperationInstanceListReq describes the node operation instance
// list by operation-id request.
export interface PluginWorkflowOperationInstanceListReq {
  only_count: boolean;
  operation_id: string[];
}

// PluginWorkflowOperationInstanceListResp describes the node operation instance
// list by operation-id response.
export interface PluginWorkflowOperationInstanceListResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  data: PluginWorkflowOperationInstanceListRespData;
}

export interface PluginWorkflowOperationInstanceListRespData {
  total: number;
  oper_inst_data: WorflowOperationInstanceData[];
}

// PluginWorkflowOperationInstanceLogGetReq describes the node operation
// instance log request.
export interface PluginWorkflowOperationInstanceLogGetReq {
  oper_inst_id: string;
}

// PluginWorkflowOperationInstanceLogGetResp describes the node operation action
// instance log response.
export interface PluginWorkflowOperationInstanceLogGetResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  data: PluginWorkflowOperationInstanceLogGetRespData;
}

export interface PluginWorkflowOperationInstanceLogGetRespData {
  total: number;
  oper_inst_logs: Record<string, WorkflowActionData>;
}

// PluginWorkflowOperationRetryReq describes the plugin operation retry.
export interface PluginWorkflowOperationRetryReq {
  workflow_id: string;
  retry_mod: string;
  operation_ids: string[];
}

// PluginWorkflowOperationRetryResp describes the plugin operation retry.
export interface PluginWorkflowOperationRetryResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  data: PluginWorkflowOperationRetryRespData;
}

export interface PluginWorkflowOperationRetryRespData {
}

// PluginWorkflowOperationTerminateReq describes the plugin operation terminate.
export interface PluginWorkflowOperationTerminateReq {
  workflow_id: string;
  operation_ids: string[];
}

// PluginWorkflowOperationTerminateResp describes the plugin operation
// terminate.
export interface PluginWorkflowOperationTerminateResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  data: PluginWorkflowOperationTerminateRespData;
}

export interface PluginWorkflowOperationTerminateRespData {
}

