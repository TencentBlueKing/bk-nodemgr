// gen-api.js 自动生成，请勿手动修改
import { type Config } from '../interceptors';

import type { WorkflowAgentInstallReq, WorkflowAgentInstallResp, WorkflowAgentUpgradeReq, WorkflowAgentUpgradeResp, WorkflowAgentReconfigReq, WorkflowAgentReconfigResp, WorkflowAgentRestartReq, WorkflowAgentRestartResp, WorkflowAgentUninstallReq, WorkflowAgentUninstallResp, WorkflowProxyInstallReq, WorkflowProxyInstallResp, WorkflowProxyUpgradeReq, WorkflowProxyUpgradeResp, WorkflowProxyReconfigReq, WorkflowProxyReconfigResp, WorkflowProxyRestartReq, WorkflowProxyRestartResp, WorkflowProxyUninstallReq, WorkflowProxyUninstallResp, WorkflowListReq, WorkflowListResp, WorkflowGetReq, WorkflowGetResp, WorkflowOperationGetReq, WorkflowOperationGetResp } from '@/@types/workflow';
import Fetch from '@/api/fetch';

const fetch = new Fetch({
  prefix: `${import.meta.env.BK_API_PREFIX}`,
});

export const WorkflowService = {
  // AgentInstall provides multi agent installing.
  AgentInstall: async <Request = WorkflowAgentInstallReq, ResponseData = WorkflowAgentInstallResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/workflow/agent/install')(params, config),
  // AgentUpgrade provides multi agent upgrading.
  AgentUpgrade: async <Request = WorkflowAgentUpgradeReq, ResponseData = WorkflowAgentUpgradeResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/workflow/agent/upgrade')(params, config),
  // AgentReconfig provides multi agent reconfiging.
  AgentReconfig: async <Request = WorkflowAgentReconfigReq, ResponseData = WorkflowAgentReconfigResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/workflow/agent/reconfig')(params, config),
  // AgentRestart provides multi agent restarting.
  AgentRestart: async <Request = WorkflowAgentRestartReq, ResponseData = WorkflowAgentRestartResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/workflow/agent/restart')(params, config),
  // AgentUninstall provides multi agent uninstalling.
  AgentUninstall: async <Request = WorkflowAgentUninstallReq, ResponseData = WorkflowAgentUninstallResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/workflow/agent/uninstall')(params, config),
  // ProxyInstall provides multi  proxy installing.
  ProxyInstall: async <Request = WorkflowProxyInstallReq, ResponseData = WorkflowProxyInstallResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/workflow/proxy/install')(params, config),
  // ProxyUpgrade provides multi proxy upgrading.
  ProxyUpgrade: async <Request = WorkflowProxyUpgradeReq, ResponseData = WorkflowProxyUpgradeResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/workflow/proxy/upgrade')(params, config),
  // ProxyReconfig provides multi proxy reconfiging.
  ProxyReconfig: async <Request = WorkflowProxyReconfigReq, ResponseData = WorkflowProxyReconfigResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/workflow/proxy/reconfig')(params, config),
  // ProxyRestart provides multi proxy restarting.
  ProxyRestart: async <Request = WorkflowProxyRestartReq, ResponseData = WorkflowProxyRestartResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/workflow/proxy/restart')(params, config),
  // ProxyUninstall provides multi proxy uninstalling.
  ProxyUninstall: async <Request = WorkflowProxyUninstallReq, ResponseData = WorkflowProxyUninstallResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/workflow/proxy/uninstall')(params, config),
  // WorkflowList provides workflow listing.
  WorkflowList: async <Request = WorkflowListReq, ResponseData = WorkflowListResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/workflow/list')(params, config),
  // WorkflowGet provides getting an existing workflow by id.
  WorkflowGet: async <Request = WorkflowGetReq, ResponseData = WorkflowGetResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/workflow/get')(params, config),
  // OperationGet provides getting an existing operation by id.
  OperationGet: async <Request = WorkflowOperationGetReq, ResponseData = WorkflowOperationGetResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/workflow/operation/get')(params, config),
};

