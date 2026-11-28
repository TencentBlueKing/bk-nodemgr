// gen-api.js 自动生成，请勿手动修改
// 周期工作流 ID、名称及操作人的精确筛选条件。
export interface ScheduleWorkflowExactConditions {
  workflow_id: string[];
  workflow_name: string[];
  operator: string[];
}

// 预留的模糊筛选条件，当前不支持模糊筛选。
export interface ScheduleWorkflowFuzzyConditions {
}

// 查询当前租户的周期工作流配置。
export interface ScheduleWorkflowListReq {
  // 分页条件。
  page: Page;
  // 是否仅返回总数。
  only_count: boolean;
  exact_include_conditions: ScheduleWorkflowExactConditions;
  fuzzy_include_conditions: ScheduleWorkflowFuzzyConditions;
  // 配置操作时间范围，不代表执行实例的起止时间。
  operate_time_range: TimeRange;
}

// 周期工作流的长期调度配置；启用状态不代表某轮执行已经完成。
export interface ScheduleWorkflowListResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: ScheduleWorkflowListRespData;
}

// 一项周期工作流调度配置。
export interface ScheduleWorkflowListRespScheduleWorkflow {
  workflow_id: string;
  workflow_name: string;
  trigger_id: string;
  operator: string;
  // 配置操作时间，Unix 毫秒时间戳。
  operate_time: number;
  // 是否启用周期调度。
  enabled: boolean;
  // 周期调度间隔。
  interval: string;
}

export interface ScheduleWorkflowListRespData {
  total: number;
  items: ScheduleWorkflow[];
}

// 查询周期工作流根操作，或下钻某轮父实例的直接子操作。
export interface ScheduleWorkflowOperationListReq {
  only_count: boolean;
  page: Page;
  // 周期工作流 ID，必填。
  workflow_id: string;
  // 父操作实例 ID；为空时查询工作流根操作。
  parent_oper_inst_id: string;
}

// 周期工作流的操作记录及最近一次执行状态。
export interface ScheduleWorkflowOperationListResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: ScheduleWorkflowOperationListRespData;
}

// 一条周期工作流操作记录。
export interface ScheduleWorkflowOperationListRespScheduleWorkflowOperation {
  operation_id: string;
  instance_ids: string[];
  // 操作初始化时记录的操作人；未记录时为空。
  operator: string;
  // 操作创建时间，Unix 毫秒时间戳。
  create_time: number;
  latest_oper_inst_brief_data: WorkflowOperInstBriefData;
  parent_operation_id: string;
  parent_oper_inst_id: string;
}

export interface ScheduleWorkflowOperationListRespData {
  operations: ScheduleWorkflowOperation[];
  total: number;
}

// 按操作 ID 或实例 ID 查询周期工作流执行实例；两类 ID 至少提供一类。
export interface ScheduleWorkflowOperationInstanceListReq {
  only_count: boolean;
  // 操作 ID 列表，每项均不能为空。
  operation_id: string[];
  // 操作实例 ID 列表，每项均不能为空。
  oper_inst_id: string[];
}

// 保留范围内的执行实例及生命周期；每个实例对应一轮执行。
export interface ScheduleWorkflowOperationInstanceListResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: ScheduleWorkflowOperationInstanceListRespData;
}

export interface ScheduleWorkflowOperationInstanceListRespData {
  total: number;
  oper_inst_data: WorflowOperationInstanceData[];
}

// 查询一个周期工作流操作实例的步骤日志。
export interface ScheduleWorkflowOperationInstanceLogGetReq {
  // 操作实例 ID，必填。
  oper_inst_id: string;
}

// 执行实例的步骤日志及额外执行日志，受历史清理策略保留范围限制。
export interface ScheduleWorkflowOperationInstanceLogGetResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: ScheduleWorkflowOperationInstanceLogGetRespData;
}

export interface ScheduleWorkflowOperationInstanceLogGetRespData {
  oper_inst_logs: Record<string, WorkflowActionData>;
  extra_execution_logs: WorkflowActionMessage;
}

// 修改当前租户内一个定时工作流的调度周期。
export interface ScheduleWorkflowUpdateIntervalReq {
  // 定时工作流 ID，必填。
  workflow_id: string;
  // 周期表达式，必填；支持含秒的六段 Cron 及 @every 10m 等描述符。
  interval: string;
}

// 周期配置已保存，不代表调度器已完成同步。
export interface ScheduleWorkflowUpdateIntervalResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: ScheduleWorkflowUpdateIntervalRespData;
}

export interface ScheduleWorkflowUpdateIntervalRespData {
}
