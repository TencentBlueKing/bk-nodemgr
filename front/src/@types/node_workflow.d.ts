// gen-api.js 自动生成，请勿手动修改
// NodeWorkflowInfo describes the node workflow information.
export interface NodeWorkflowInfo {
  workflow_id: string;
  trigger_id: string;
  type: string;
  bk_biz_id: number[];
  operator: string;
  operate_time: number;
  finish_time: number;
  status: string;
  bk_biz_name: string[];
}

// NodeWorkflowExactConditions describes the exact conditions of node workflow
// list request.
export interface NodeWorkflowExactConditions {
  workflow_id: string[];
  type: string[];
  bk_biz_id: number[];
  status: string[];
  operator: string[];
}

// NodeWorkflowFuzzyConditions describes the fuzzy conditions of node workflow
// list request.
export interface NodeWorkflowFuzzyConditions {
}

// NodeWorkflowListReq describes the node workflow list request.
export interface NodeWorkflowListReq {
  page: Page;
  only_count: boolean;
  exact_include_conditions: NodeWorkflowExactConditions;
  fuzzy_include_conditions: NodeWorkflowFuzzyConditions;
  operate_time_range: TimeRange;
}

// NodeWorkflowListResp describes the node workflow list response.
export interface NodeWorkflowListResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  data: NodeWorkflowListRespData;
}

export interface NodeWorkflowListRespData {
  total: number;
  items: NodeWorkflowInfo[];
}

// NodeWorkflowStatisticReq describes the node workflow statistic request.
export interface NodeWorkflowStatisticsReq {
  workflow_id: string[];
}

// NodeWorkflowStatisticResp describes the node workflow statistic response.
export interface NodeWorkflowStatisticsResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  data: NodeWorkflowStatisticsRespData;
}

export interface NodeWorkflowStatisticsRespData {
  items: WorkflowStatisticsInfo[];
}

// NodeWorkflowDistinctReq describes the node workflow distinct request.
export interface NodeWorkflowDistinctReq {
  exact_include_conditions: NodeWorkflowExactConditions;
  fuzzy_include_conditions: NodeWorkflowFuzzyConditions;
}

// NodeWorkflowDistinctResp describes the node workflow distinct response.
export interface NodeWorkflowDistinctResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  data: NodeWorkflowDistinctRespData;
}

export interface NodeWorkflowDistinctRespData {
  type: string[];
  bk_biz_id: number[];
  operator: string[];
  status: string[];
}

export interface NodeWorkflowOperationParam {
  bk_host_id: number;
  bk_biz_id: number;
  bk_networkarea_id: number;
  bk_networkunit_id: number;
  bk_host_inner_list: string[];
  bk_host_innerip_v6_list: string[];
  node_version: string;
  operator: string;
}

export interface NodeWorkflowOperation {
  operation_id: string;
  instance_ids: string[];
  param: NodeWorkflowOperationParam;
  status: NodeWorkflowOperationStatus;
  latest_action_inst_brief_data: WorkflowActionInstBriefData;
}

export interface NodeWorkflowOperationStatus {
  state: string;
  total_time_second: number;
}

// NodeWorkflowOperationExactConditions describes the exact conditions of node
// workflow operation list request.
export interface NodeWorkflowOperationExactConditions {
  workflow_id: string;
  node_version: string[];
  bk_host_innerip: string[];
  bk_host_innerip_v6: string[];
  bk_biz_id: number[];
  bk_networkarea_id: number[];
  bk_networkunit_id: number[];
  state: string[];
}

// NodeWorkflowOperationFuzzyConditions describes the fuzzy conditions of node
// workflow list request.
export interface NodeWorkflowOperationFuzzyConditions {
}

// NodeWorkflowOperationListReq describes the node operation list
// request.
export interface NodeWorkflowOperationListReq {
  only_count: boolean;
  page: Page;
  exact_include_conditions: NodeWorkflowOperationExactConditions;
  fuzzy_include_conditions: NodeWorkflowOperationFuzzyConditions;
}

// NodeWorkflowOperationListResp describes the node operation list by
export interface NodeWorkflowOperationListResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  data: NodeWorkflowOperationListRespData;
}

export interface NodeWorkflowOperationListRespData {
  operations: NodeWorkflowOperation[];
  total: number;
}

// NodeWorkflowOperationInstanceListReq describes the node operation instance
// list by operation-id request.
export interface NodeWorkflowOperationInstanceListReq {
  only_count: boolean;
  operation_id: string;
}

// NodeWorkflowOperationInstanceListResp describes the node operation instance
// list by operation-id response.
export interface NodeWorkflowOperationInstanceListResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  data: NodeWorkflowOperationInstanceListRespData;
}

export interface NodeWorkflowOperationInstanceListRespData {
  total: number;
  oper_inst_data: WorflowOperationInstanceData[];
}

export interface NodeWorkflowOperationInstanceLogGetReq {
  oper_inst_id: string;
}

// NodeWorkflowOperationInstanceLogGetResp describes the node operation action
// instance log
export interface NodeWorkflowOperationInstanceLogGetResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  data: NodeWorkflowOperationInstanceLogGetRespData;
}

export interface NodeWorkflowOperationInstanceLogGetRespData {
  oper_inst_logs: Record<string, WorkflowActionData>;
}

// NodeWorkflowOperationRetryReq
export interface NodeWorkflowOperationRetryReq {
  workflow_id: string;
  retry_mod: string;
  operation_ids: string[];
}

// NodeWorkflowOperationRetryResp
export interface NodeWorkflowOperationRetryResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  data: NodeWorkflowOperationRetryRespData;
}

export interface NodeWorkflowOperationRetryRespData {
}

// NodeWorkflowOperationTerminateReq
export interface NodeWorkflowOperationTerminateReq {
  workflow_id: string;
  operation_ids: string[];
}

// NodeWorkflowOperationTerminateResp
export interface NodeWorkflowOperationTerminateResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  data: NodeWorkflowOperationTerminateRespData;
}

export interface NodeWorkflowOperationTerminateRespData {
}

// ManualSolutionStep describes the step of manual solution.
export interface ManualSolutionStep {
  type: string;
  name_en: string;
  name_zh: string;
  content_en: string;
  content_zh: string;
}

// ManualSolution describes the manual solution.
export interface ManualSolution {
  type: string;
  description_en: string;
  description_zh: string;
  steps: ManualSolutionStep[];
}

// NodeWorkflowOperationManualSolutionGetReq describes the node workflow
// operation get manual solution request.
export interface NodeWorkflowOperationManualSolutionGetReq {
  workflow_id: string;
  operation_id: string;
}

// NodeWorkflowOperationManualSolutionGetResp describes the node workflow
// operation get manual solution response.
export interface NodeWorkflowOperationManualSolutionGetResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  data: ManualSolution[];
}

