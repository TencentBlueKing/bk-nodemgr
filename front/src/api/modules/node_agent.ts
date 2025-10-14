// gen-api.js 自动生成，请勿手动修改
import { type Config } from '../interceptors';

import type { NodeAgentInstallReq, NodeAgentInstallResp, NodeAgentUpgradeReq, NodeAgentUpgradeResp, NodeAgentInstallCheckReq, NodeAgentInstallCheckResp, UploadAgentTemplateReq, UploadAgentTemplateResp } from '@/@types/node_agent';
import Fetch from '@/api/fetch';

const fetch = new Fetch({
  prefix: `${import.meta.env.BK_API_PREFIX}`,
});

export const NodeAgentService = {
  // NodeAgentInstall installs node agent.
  NodeAgentInstall: async <Request = NodeAgentInstallReq, ResponseData = NodeAgentInstallResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/node/agent/install')(params, config),
  // NodeAgentUpgrade upgrades node agent.
  NodeAgentUpgrade: async <Request = NodeAgentUpgradeReq, ResponseData = NodeAgentUpgradeResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/node/agent/upgrade')(params, config),
  // NodeAgentInstallCheck  check installs node agent.
  NodeAgentInstallCheck: async <Request = NodeAgentInstallCheckReq, ResponseData = NodeAgentInstallCheckResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/node/agent/install_check')(params, config),
  UploadAgentTemplate: async <Request = UploadAgentTemplateReq, ResponseData = UploadAgentTemplateResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/node/agent/upload_template')(params, config),
};

