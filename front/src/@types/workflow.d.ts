// gen-api.js 自动生成，请勿手动修改
// WorkflowStatisticsInfo describes the workflow statistics information.
export interface WorkflowStatisticsInfo {
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

// WorflowOperationInstanceData describes the workflow operation instance data.
export interface WorflowOperationInstanceData {
  operation_id: string;
  oper_inst_id: string;
  oper_inst_status: string;
  operation_def_name: string;
  parent_operation_id: string;
  action_names: string[];
  life_cycle: WorkflowLifeCycle;
  latest_action_inst_brief_data: WorkflowActionInstBriefData;
}

// WorkflowLifeCycle describes the workflow life cycle.
export interface WorkflowLifeCycle {
  state: string;
  create_time: number;
  start_time: number;
  end_time: number;
  stop_time: number;
}

// WorkflowActionMessage describes the workflow action message logs.
export interface WorkflowActionMessage {
  logs: Message[];
}

export interface WorkflowActionMessageMessage {
  time: number;
  level: string;
  // Chinese log content
  text_zh: string;
  // English log content
  text_en: string;
}

// WorkflowActionInstBriefData describes the workflow action instance brief
// data.
export interface WorkflowActionInstBriefData {
  name: string;
  tags: string[];
}

// WorkflowActionData describes the workflow action data.
export interface WorkflowActionData {
  life_cycle: WorkflowLifeCycle;
  message: WorkflowActionMessage;
  // Chinese display name
  display_name_zh: string;
  // English display name
  display_name_en: string;
}

// WorkflowOperInstBriefData describes the brief data of operation instance.
export interface WorkflowOperInstBriefData {
  life_cycle: WorkflowLifeCycle;
  latest_action_inst_brief_data: WorkflowActionInstBriefData;
}

