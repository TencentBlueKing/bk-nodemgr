// gen-api.js 自动生成，请勿手动修改
// WorkflowHostParameter describes the host action parameters in workflows
// service. Such as installing, upgrading.
export interface WorkflowHostParameter {
  bkHostId: number;
  topo: Topo;
  attributes: Attributes;
  config: Config;
}

// WorkflowAgentInstallReq describes the HTTP request body when install agent in
// workflow service.
export interface WorkflowAgentInstallReq {
  hosts: WorkflowHostParameter[];
  timeoutSec: number;
}

// WorkflowAgentInstallResp describes the HTTP response body when install agent
// in workflow service.
export interface WorkflowAgentInstallResp {
  code: number;
  message: string;
  requestId: string;
  data: Data;
}

// WorkflowAgentUpgradeReq describes the HTTP request body when upgrade agent in
// workflow service.
export interface WorkflowAgentUpgradeReq {
  hosts: WorkflowHostParameter[];
  timeoutSec: number;
  force: boolean;
}

// WorkflowAgentUpgradeResp describes the HTTP response body when upgrade agent
// in workflow service.
export interface WorkflowAgentUpgradeResp {
  code: number;
  message: string;
  requestId: string;
  data: Data;
}

// WorkflowAgentReconfigReq describes the HTTP request body when reconfig agent
// in workflow service.
export interface WorkflowAgentReconfigReq {
  hosts: WorkflowHostParameter[];
  timeoutSec: number;
  force: boolean;
}

// WorkflowAgentReconfigResp describes the HTTP response body when reconfig
// agent in workflow service.
export interface WorkflowAgentReconfigResp {
  code: number;
  message: string;
  requestId: string;
  data: Data;
}

// WorkflowAgentRestartReq describes the HTTP request body when restart agent in
// workflow service.
export interface WorkflowAgentRestartReq {
  hosts: WorkflowHostParameter[];
  timeoutSec: number;
  force: boolean;
}

// WorkflowAgentRestartResp describes the HTTP response body when restart agent
// in workflow service.
export interface WorkflowAgentRestartResp {
  code: number;
  message: string;
  requestId: string;
  data: Data;
}

// WorkflowAgentUninstallReq describes the HTTP request body when uninstall
// agent in workflow service.
export interface WorkflowAgentUninstallReq {
  hosts: WorkflowHostParameter[];
  timeoutSec: number;
}

// WorkflowAgentUninstallResp describes the HTTP response body when uninstall
// agent in workflow service.
export interface WorkflowAgentUninstallResp {
  code: number;
  message: string;
  requestId: string;
  data: Data;
}

// WorkflowProxyInstallReq describes the HTTP request body when install proxy in
// workflow service.
export interface WorkflowProxyInstallReq {
  hosts: WorkflowHostParameter[];
  timeoutSec: number;
}

// WorkflowProxyInstallResp describes the HTTP response body when install proxy
// in workflow service.
export interface WorkflowProxyInstallResp {
  code: number;
  message: string;
  requestId: string;
  data: Data;
}

// WorkflowProxyUpgradeReq describes the HTTP request body when upgrade proxy in
// workflow service.
export interface WorkflowProxyUpgradeReq {
  hosts: WorkflowHostParameter[];
  timeoutSec: number;
  force: boolean;
}

// WorkflowProxyUpgradeResp describes the HTTP response body when upgrade proxy
// in workflow service.
export interface WorkflowProxyUpgradeResp {
  code: number;
  message: string;
  requestId: string;
  data: Data;
}

// WorkflowProxyReconfigReq describes the HTTP request body when reconfig proxy
// in workflow service.
export interface WorkflowProxyReconfigReq {
  hosts: WorkflowHostParameter[];
  timeoutSec: number;
  force: boolean;
}

// WorkflowProxyReconfigResp describes the HTTP response body when reconfig
// proxy in workflow service.
export interface WorkflowProxyReconfigResp {
  code: number;
  message: string;
  requestId: string;
  data: Data;
}

// WorkflowProxyRestartReq describes the HTTP request body when restart proxy in
// workflow service.
export interface WorkflowProxyRestartReq {
  hosts: WorkflowHostParameter[];
  timeoutSec: number;
  force: boolean;
}

// WorkflowProxyRestartResp describes the HTTP response body when restart proxy
// in workflow service.
export interface WorkflowProxyRestartResp {
  code: number;
  message: string;
  requestId: string;
  data: Data;
}

// WorkflowProxyUninstallReq describes the HTTP request body when uninstall
// proxy in workflow service.
export interface WorkflowProxyUninstallReq {
  hosts: WorkflowHostParameter[];
  timeoutSec: number;
}

// WorkflowProxyUninstallResp describes the HTTP response body when uninstall
// proxy in workflow service.
export interface WorkflowProxyUninstallResp {
  code: number;
  message: string;
  requestId: string;
  data: Data;
}

// OperationInstance describes the operation instance informations. One
// operation may contains multi instances.
export interface OperationInstance {
  instanceId: string;
  createAt: number;
  startedAt: number;
  endedAt: number;
  stoppedAt: number;
  action: Action[];
}

// OperationBrief describes the operation informations in brief.
export interface OperationBrief {
  operationId: string;
  createAt: number;
  startedAt: number;
  endedAt: number;
  stoppedAt: number;
  bkHostId: number;
  bkHostInnerip: number;
  instanceCount: number;
}

// OperationDetail describes the operation informations in detail.
export interface OperationDetail {
  operationId: string;
  createAt: number;
  startedAt: number;
  endedAt: number;
  stoppedAt: number;
  bkHostId: number;
  bkHostInnerip: number;
  instanceCount: number;
  instances: OperationInstance[];
}

// WorkflowBrief describes the workflow informations in brief.
export interface WorkflowBrief {
  workflowId: string;
  status: string;
  operator: string;
  type: number;
  createAt: number;
  startedAt: number;
  endedAt: number;
  stoppedAt: number;
  operationCount: number;
}

// WorkflowDetail describes the workflow informations in detail.
export interface WorkflowDetail {
  workflowId: string;
  status: string;
  operator: string;
  type: number;
  createAt: number;
  startedAt: number;
  endedAt: number;
  stoppedAt: number;
  operationCount: number;
  operations: OperationBrief[];
}

// WorkflowListReq describes the HTTP request body when list workflows in
// workflow service.
export interface WorkflowListReq {
  page: Page;
  onlyCount: boolean;
  includeConditions: Conditions;
}

// WorkflowListResp describes the HTTP response body when list workflows in
// workflow service.
export interface WorkflowListResp {
  code: number;
  message: string;
  requestId: string;
  data: Data;
}

// WorkflowGetReq describes the HTTP request body when get workflow in workflow
// service.
export interface WorkflowGetReq {
  workflowId: string;
}

// WorkflowGetResp describes the HTTP response body when get workflow in
// workflow service.
export interface WorkflowGetResp {
  code: number;
  message: string;
  requestId: string;
  data: WorkflowDetail;
}

// WorkflowOperationGetReq describes the HTTP request body when get operation in
// workflow service.
export interface WorkflowOperationGetReq {
  workflowId: string;
  operationId: string;
}

// WorkflowOperationGetResp describes the HTTP response body when get operation
// in workflow service.
export interface WorkflowOperationGetResp {
  code: number;
  message: string;
  requestId: string;
  Data: OperationDetail;
}

