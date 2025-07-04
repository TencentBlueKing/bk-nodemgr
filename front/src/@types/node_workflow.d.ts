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
  data: NodeWorkflowStatisticsRespData;
}

export interface NodeWorkflowStatisticsRespStatisticsInfo {
  workflow_id: string;
  total_count: number;
  init_count: number;
  launched_count: number;
  running_count: number;
  success_count: number;
  failed_count: number;
  timeout_count: number;
  terminated_count: number;
}

export interface NodeWorkflowStatisticsRespData {
  items: StatisticsInfo[];
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
  data: NodeWorkflowDistinctRespData;
}

export interface NodeWorkflowDistinctRespData {
  type: string[];
  bk_biz_id: number[];
  operator: string[];
  status: string[];
}

export interface NodeWorkflowOperationParam {
  bk_networkarea_id: number;
  bk_host_inner: string;
  bk_host_innerip_v6: string;
  node_version: string;
  bk_biz_id: number;
}

export interface NodeWorkflowOperation {
  operation_id: string;
  instance_ids: string[];
  param: NodeWorkflowOperationParam;
  status: NodeWorkflowOperationStatus;
}

export interface NodeWorkflowOperationStatus {
  state: string;
  total_time_second: number;
}

export interface Lifecycle {
  state: string;
  operate_time: number;
  finish_time: number;
}

// NodeWorflowOperationInstance describes the node operation instance.
export interface NodeWorflowOperationInstance {
  operation_id: string;
  oper_inst_id: string;
  oper_inst_status: string;
  bk_host_innerip: string;
  bk_networkunit_id: string;
  bk_biz_id: string;
  node_version: string;
  action: Action[];
}

export interface NodeWorflowOperationInstanceAction {
  action_name: string;
  action_status: string;
  cost_time_sec: number;
}

export interface NodeWorflowOperationInstanceData {
  operation_id: string;
  oper_inst_id: string;
  oper_inst_status: string;
  operation_def_name: string;
  parent_operation_id: string;
  action_names: string[];
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
  data: NodeWorkflowOperationListRespData;
}

export interface NodeWorkflowOperationListRespData {
  operations: NodeWorkflowOperation[];
  total_count: number;
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
  data: NodeWorkflowOperationInstanceListRespData;
}

export interface NodeWorkflowOperationInstanceListRespData {
  total: number;
  oper_inst_data: NodeWorflowOperationInstanceData[];
}

// NodeWorkflowInstanceStatusExactConditions describes the exact conditions of
// node workflow
export interface NodeWorkflowInstanceStatusExactConditions {
  trigger_id: string[];
}

// NodeWorkflowInstanceStatusFuzzyConditions describes the fuzzy conditions of
// node workflow list request.
export interface NodeWorkflowInstanceStatusFuzzyConditions {
}

// NodeWorkflowOperationInstanceStatusListReq ...
export interface NodeWorkflowOperationInstanceListStatusReq {
  page: Page;
  exact_include_conditions: NodeWorkflowInstanceStatusExactConditions;
  fuzzy_include_conditions: NodeWorkflowInstanceStatusFuzzyConditions;
}

export interface NodeWorkflowOperationInstanceStatus {
  index: number;
  instance_id: string;
  status: string;
  trigger_id: string;
}

export interface NodeWorkflowOperationInstanceListStatusResp {
  code: number;
  message: string;
  request_id: string;
  data: NodeWorkflowOperationInstanceListStatusRespData;
}

export interface NodeWorkflowOperationInstanceListStatusRespData {
  items: NodeWorkflowOperationInstanceStatus[];
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
  data: NodeWorkflowOperationInstanceLogGetRespData;
}

export interface NodeWorkflowOperationInstanceLogGetRespData {
  total: number;
  oper_inst_logs: Record<string, NodeWorkflowActionData>;
}

export interface LifeCycle {
  state: string;
  create_time: number;
  start_time: number;
  end_time: number;
}

export interface NodeWorkflowActionMessage {
  logs: Message[];
}

export interface NodeWorkflowActionMessageMessage {
  time: number;
  text: string;
}

export interface NodeWorkflowActionData {
  life_cycle: LifeCycle;
  message: NodeWorkflowActionMessage;
}

// NodeWorkflowOperationRetryReq
export interface NodeWorkflowOperationRetryReq {
  workflow_id: string;
  retry_mod: string;
  operation_id: string[];
}

// NodeWorkflowOperationRetryResp
export interface NodeWorkflowOperationRetryResp {
  code: number;
  message: string;
  request_id: string;
  data: NodeWorkflowOperationRetryRespData;
}

export interface NodeWorkflowOperationRetryRespData {
  instance_ids: string[];
}

