// gen-api.js 自动生成，请勿手动修改
// WorkflowHostParameter describes the host action parameters in workflows
// service. Such as installing, upgrading.
export interface WorkflowHostParameter {
  bk_host_id: number;
  topo: WorkflowHostParameterTopo;
  attributes: WorkflowHostParameterAttributes;
  config: WorkflowHostParameterConfig;
}

export interface WorkflowHostParameterTopo {
  bk_biz_id: number;
  bk_networkarea_id: number;
  bk_networkunit_id: number;
}

export interface WorkflowHostParameterAttributes {
  bk_host_innerip: string;
  bk_host_innerip_v6: string;
  bk_host_outerip: string;
  bk_host_outerip_v6: string;
  login_ip: string;
  login_port: number;
  login_password: string;
}

export interface WorkflowHostParameterConfig {
  version: string;
}

// WorkflowAgentInstallReq describes the HTTP request body when install agent in
// workflow service.
export interface WorkflowAgentInstallReq {
  hosts: WorkflowHostParameter[];
  timeout_sec: number;
}

// WorkflowAgentInstallResp describes the HTTP response body when install agent
// in workflow service.
export interface WorkflowAgentInstallResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  data: WorkflowAgentInstallRespData;
}

export interface WorkflowAgentInstallRespData {
  workflow_id: string;
}

// WorkflowAgentUpgradeReq describes the HTTP request body when upgrade agent in
// workflow service.
export interface WorkflowAgentUpgradeReq {
  hosts: WorkflowHostParameter[];
  timeout_sec: number;
  force: boolean;
}

// WorkflowAgentUpgradeResp describes the HTTP response body when upgrade agent
// in workflow service.
export interface WorkflowAgentUpgradeResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  data: WorkflowAgentUpgradeRespData;
}

export interface WorkflowAgentUpgradeRespData {
  workflow_id: string;
}

// WorkflowAgentReconfigReq describes the HTTP request body when reconfig agent
// in workflow service.
export interface WorkflowAgentReconfigReq {
  hosts: WorkflowHostParameter[];
  timeout_sec: number;
  force: boolean;
}

// WorkflowAgentReconfigResp describes the HTTP response body when reconfig
// agent in workflow service.
export interface WorkflowAgentReconfigResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  data: WorkflowAgentReconfigRespData;
}

export interface WorkflowAgentReconfigRespData {
  workflow_id: string;
}

// WorkflowAgentRestartReq describes the HTTP request body when restart agent in
// workflow service.
export interface WorkflowAgentRestartReq {
  hosts: WorkflowHostParameter[];
  timeout_sec: number;
  force: boolean;
}

// WorkflowAgentRestartResp describes the HTTP response body when restart agent
// in workflow service.
export interface WorkflowAgentRestartResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  data: WorkflowAgentRestartRespData;
}

export interface WorkflowAgentRestartRespData {
  workflow_id: string;
}

// WorkflowAgentUninstallReq describes the HTTP request body when uninstall
// agent in workflow service.
export interface WorkflowAgentUninstallReq {
  hosts: WorkflowHostParameter[];
  timeout_sec: number;
}

// WorkflowAgentUninstallResp describes the HTTP response body when uninstall
// agent in workflow service.
export interface WorkflowAgentUninstallResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  data: WorkflowAgentUninstallRespData;
}

export interface WorkflowAgentUninstallRespData {
  workflow_id: string;
}

// WorkflowProxyInstallReq describes the HTTP request body when install proxy in
// workflow service.
export interface WorkflowProxyInstallReq {
  hosts: WorkflowHostParameter[];
  timeout_sec: number;
}

// WorkflowProxyInstallResp describes the HTTP response body when install proxy
// in workflow service.
export interface WorkflowProxyInstallResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  data: WorkflowProxyInstallRespData;
}

export interface WorkflowProxyInstallRespData {
  workflow_id: string;
}

// WorkflowProxyUpgradeReq describes the HTTP request body when upgrade proxy in
// workflow service.
export interface WorkflowProxyUpgradeReq {
  hosts: WorkflowHostParameter[];
  timeout_sec: number;
  force: boolean;
}

// WorkflowProxyUpgradeResp describes the HTTP response body when upgrade proxy
// in workflow service.
export interface WorkflowProxyUpgradeResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  data: WorkflowProxyUpgradeRespData;
}

export interface WorkflowProxyUpgradeRespData {
  workflow_id: string;
}

// WorkflowProxyReconfigReq describes the HTTP request body when reconfig proxy
// in workflow service.
export interface WorkflowProxyReconfigReq {
  hosts: WorkflowHostParameter[];
  timeout_sec: number;
  force: boolean;
}

// WorkflowProxyReconfigResp describes the HTTP response body when reconfig
// proxy in workflow service.
export interface WorkflowProxyReconfigResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  data: WorkflowProxyReconfigRespData;
}

export interface WorkflowProxyReconfigRespData {
  workflow_id: string;
}

// WorkflowProxyRestartReq describes the HTTP request body when restart proxy in
// workflow service.
export interface WorkflowProxyRestartReq {
  hosts: WorkflowHostParameter[];
  timeout_sec: number;
  force: boolean;
}

// WorkflowProxyRestartResp describes the HTTP response body when restart proxy
// in workflow service.
export interface WorkflowProxyRestartResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  data: WorkflowProxyRestartRespData;
}

export interface WorkflowProxyRestartRespData {
  workflow_id: string;
}

// WorkflowProxyUninstallReq describes the HTTP request body when uninstall
// proxy in workflow service.
export interface WorkflowProxyUninstallReq {
  hosts: WorkflowHostParameter[];
  timeout_sec: number;
}

// WorkflowProxyUninstallResp describes the HTTP response body when uninstall
// proxy in workflow service.
export interface WorkflowProxyUninstallResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  data: WorkflowProxyUninstallRespData;
}

export interface WorkflowProxyUninstallRespData {
  workflow_id: string;
}

// OperationInstance describes the operation instance informations. One
// operation may contains multi instances.
export interface OperationInstance {
  instance_id: string;
  create_at: number;
  started_at: number;
  ended_at: number;
  stopped_at: number;
  action: Action[];
}

export interface OperationInstanceMessage {
  timestamp: number;
  text: string;
}

export interface OperationInstanceAction {
  create_at: number;
  started_at: number;
  ended_at: number;
  stopped_at: number;
  status: string;
  message: Message[];
}

// OperationBrief describes the operation informations in brief.
export interface OperationBrief {
  operation_id: string;
  create_at: number;
  started_at: number;
  ended_at: number;
  stopped_at: number;
  bk_host_id: number;
  bk_host_innerip: number;
  instance_count: number;
}

// OperationDetail describes the operation informations in detail.
export interface OperationDetail {
  operation_id: string;
  create_at: number;
  started_at: number;
  ended_at: number;
  stopped_at: number;
  bk_host_id: number;
  bk_host_innerip: number;
  instance_count: number;
  instances: OperationInstance[];
}

// WorkflowBrief describes the workflow informations in brief.
export interface WorkflowBrief {
  workflow_id: string;
  status: string;
  operator: string;
  type: number;
  create_at: number;
  started_at: number;
  ended_at: number;
  stopped_at: number;
  operation_count: number;
}

// WorkflowDetail describes the workflow informations in detail.
export interface WorkflowDetail {
  workflow_id: string;
  status: string;
  operator: string;
  type: number;
  create_at: number;
  started_at: number;
  ended_at: number;
  stopped_at: number;
  operation_count: number;
  operations: OperationBrief[];
}

// WorkflowListReq describes the HTTP request body when list workflows in
// workflow service.
export interface WorkflowListReq {
  page: Page;
  only_count: boolean;
  include_conditions: WorkflowListReqConditions;
}

export interface WorkflowListReqConditions {
}

// WorkflowListResp describes the HTTP response body when list workflows in
// workflow service.
export interface WorkflowListResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  data: WorkflowListRespData;
}

export interface WorkflowListRespData {
  total: number;
  items: WorkflowBrief[];
}

// WorkflowGetReq describes the HTTP request body when get workflow in workflow
// service.
export interface WorkflowGetReq {
  workflow_id: string;
}

// WorkflowGetResp describes the HTTP response body when get workflow in
// workflow service.
export interface WorkflowGetResp {
  code: number;
  message: string;
  request_id: string;
  data: WorkflowDetail;
}

// WorkflowOperationGetReq describes the HTTP request body when get operation in
// workflow service.
export interface WorkflowOperationGetReq {
  workflow_id: string;
  operation_id: string;
}

// WorkflowOperationGetResp describes the HTTP response body when get operation
// in workflow service.
export interface WorkflowOperationGetResp {
  code: number;
  message: string;
  request_id: string;
  Data: OperationDetail;
}

