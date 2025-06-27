// gen-api.js 自动生成，请勿手动修改
import { type Config } from '../interceptors';

import type { NodeAgentInstallReq, NodeAgentInstallResp } from '@/@types/node_agent';
import Fetch from '@/api/fetch';

const fetch = new Fetch({
  prefix: `${import.meta.env.BK_API_PREFIX}`,
});

export const NodeAgentService = {
  // NodeAgentInstall installs node agent.
  NodeAgentInstall: async <Request = NodeAgentInstallReq, ResponseData = NodeAgentInstallResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/node_agent/install')(params, config),
};

