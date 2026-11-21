// gen-api.js 自动生成，请勿手动修改
// NodeWorkflowInfo describes the node workflow information.
export interface NodeWorkflowInfo {
  workflow_id: string;
  trigger_id: string;
  type: string;
  bk_biz_id: number[];
  bk_networkarea_id: number[];
  bk_networkunit_id: number[];
  operator: string;
  operate_time: number;
  finish_time: number;
  status: string;
  bk_biz_name: string[];
  node_role: string[];
  deploy_policy_ids: number[];
}

// NodeWorkflowExactConditions describes the exact conditions of node workflow
// list request.
export interface NodeWorkflowExactConditions {
  workflow_id: string[];
  type: string[];
  bk_biz_id: number[];
  status: string[];
  operator: string[];
  bk_host_innerip: string[];
  bk_host_innerip_v6: string[];
  node_role: string[];
  deploy_policy_id: number[];
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
  permission: Permission;
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
  permission: Permission;
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
  permission: Permission;
  data: NodeWorkflowDistinctRespData;
}

export interface NodeWorkflowDistinctRespData {
  type: string[];
  bk_biz_id: number[];
  operator: string[];
  status: string[];
}

// NodeDeploymentInfo describes the node workflow operation information.
export interface NodeDeploymentInfo {
  bk_host_id: number;
  bk_biz_id: number;
  bk_host_innerip_list: string[];
  bk_host_innerip_v6_list: string[];
  bk_networkarea_id: number;
  bk_networkunit_id: number;
  node_version: string;
}

// NodeWorkflowOperation describes the node workflow operation.
export interface NodeWorkflowOperation {
  operation_id: string;
  instance_ids: string[];
  operator: string;
  create_time: number;
  node_deployment_info: NodeDeploymentInfo;
  latest_oper_inst_brief_data: WorkflowOperInstBriefData;
}

// NodeWorkflowOperationExactConditions describes the exact conditions of node
// workflow operation list request.
export interface NodeWorkflowOperationExactConditions {
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
  workflow_id: string;
  exact_include_conditions: NodeWorkflowOperationExactConditions;
  fuzzy_include_conditions: NodeWorkflowOperationFuzzyConditions;
}

// NodeWorkflowOperationListResp describes the node operation list by
export interface NodeWorkflowOperationListResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: NodeWorkflowOperationListRespData;
}

export interface NodeWorkflowOperationListRespData {
  operations: NodeWorkflowOperation[];
  total: number;
}

// NodeWorkflowOperationDistinctReq describes the node workflow operation
// distinct request.
export interface NodeWorkflowOperationDistinctReq {
  workflow_id: string;
}

// NodeWorkflowOperationDistinctResp describes the node workflow operation
// distinct response.
export interface NodeWorkflowOperationDistinctResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: NodeWorkflowOperationDistinctRespData;
}

export interface NodeWorkflowOperationDistinctRespData {
  state: string[];
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
  permission: Permission;
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
  permission: Permission;
  data: NodeWorkflowOperationInstanceLogGetRespData;
}

export interface NodeWorkflowOperationInstanceLogGetRespData {
  oper_inst_logs: Record<string, WorkflowActionData>;
  extra_execution_logs: WorkflowActionMessage;
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
  permission: Permission;
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
  permission: Permission;
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
  permission: Permission;
  data: ManualSolution[];
}

// NodeWorkflowOperationOfflinePackageDownloadReq describes the offline package
// download request.
export interface NodeWorkflowOperationOfflinePackageDownloadReq {
  operation_id: string;
}

// NodeWorkflowOperationOfflineInstallResultSubmitReq describes the offline
// install result submit request.
export interface NodeWorkflowOperationOfflineInstallResultSubmitReq {
  operation_id: string;
  result_data: string;
}

// NodeWorkflowOperationOfflineInstallResultSubmitResp describes the offline
// install result submit response.
export interface NodeWorkflowOperationOfflineInstallResultSubmitResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
}

// 可复制到新安装表单的历史主机参数，不包含凭据内容。
export interface NodeWorkflowInstallHost {
  // 主机 ID。
  bk_host_id: number;
  // 业务 ID。
  bk_biz_id: number;
  // 管控区域 ID。
  bk_networkarea_id: number;
  // 管控单元 ID。
  bk_networkunit_id: number;
  // 寻址方式。
  bk_addressing: string;
  // 内网 IPv4 地址。
  bk_host_innerip: string[];
  // 内网 IPv6 地址。
  bk_host_innerip_v6: string[];
  // 操作系统类型。
  os_type: string;
  // 登录地址。
  login_ip: string;
  // 登录用户。
  login_user: string;
  // 登录认证方式，不包含认证凭据内容。
  login_mode: string;
  // 安装连接方式。
  install_method: string;
  // 登录端口。
  login_port: number;
  // 是否使用 sudo。
  is_sudo: boolean;
  // 是否重新注册主机。
  re_register: boolean;
  // 是否安装预订阅插件。
  install_pre_ordered_plugins: boolean;
  // 是否更新 GSE task 组件。
  renew_gse_task: boolean;
  // 是否更新 GSE proc 组件。
  renew_gse_proc: boolean;
  // 对外服务 IPv4 地址。
  export_ip: string;
  // 对外服务 IPv6 地址。
  export_ip_v6: string;
  // 广播 IPv4 地址。
  advertise_ip: string;
  // 广播 IPv6 地址。
  advertise_ip_v6: string;
  // Proxy 标签。
  proxy_tags: string[];
  // Proxy 安装来源管控单元 ID。
  proxy_install_origin_unit_id: number;
  // Relay 下载端口。
  relay_download_port: number;
  // Relay 回调端口。
  relay_callback_port: number;
  // CPU 架构。
  cpu_arch: string;
}

// 单个历史安装操作可复制的参数。
export interface NodeWorkflowOperationInstallParams {
  // 历史操作 ID。
  operation_id: string;
  // 安全的主机安装参数。
  host: NodeWorkflowInstallHost;
  // 历史目标版本。
  target_version: TargetVersion[];
  // 是否手动安装。
  is_manual: boolean;
  // 是否离线安装。
  is_offline: boolean;
  // 需要在新安装表单中补充的字段。
  missing_fields: string[];
}

// 按工作流及选中的操作读取历史安装参数。
export interface NodeWorkflowOperationInstallParamsGetReq {
  // 必填，历史工作流 ID。
  workflow_id: string;
  // 必填，非空且无空白或重复 ID 的操作列表。
  operation_ids: string[];
}

// 返回可复制的安装参数和无法恢复参数的操作 ID。
export interface NodeWorkflowOperationInstallParamsGetResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: NodeWorkflowOperationInstallParamsGetRespData;
}

export interface NodeWorkflowOperationInstallParamsGetRespData {
  // 历史安装工作流类型。
  type: string;
  // 可恢复的安装参数。
  items: NodeWorkflowOperationInstallParams[];
  // 无法恢复安装参数的操作 ID。
  unavailable_operation_ids: string[];
}

