// gen-api.js 自动生成，请勿手动修改
import { type Config } from '../interceptors';

import type { NodeOpsHistoryListReq, NodeOpsHistoryListResp, NodeOpsHistoryDetailReq, NodeOpsHistoryDetailResp, NodeOpsUpdateFieldsAgentReq, NodeOpsUpdateFieldsAgentResp, NodeOpsUpdateFieldsProxyReq, NodeOpsUpdateFieldsProxyResp } from '@/@types/ops';
import Fetch from '@/api/fetch';

const fetch = new Fetch({
  prefix: `${import.meta.env.BK_API_PREFIX}`,
});

export const OpsService = {
  // NodeOpsHistoryList lists operations across agent and proxy roles within
  // the caller's permission scope.
  NodeOpsHistoryList: async <Request = NodeOpsHistoryListReq, ResponseData = NodeOpsHistoryListResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/node/ops/history/list')(params, config),
  // NodeOpsHistoryDetail retrieves all archived snapshots for an operation
  // within the caller's permission scope.
  NodeOpsHistoryDetail: async <Request = NodeOpsHistoryDetailReq, ResponseData = NodeOpsHistoryDetailResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/node/ops/history/detail')(params, config),
  // NodeOpsUpdateFieldsAgent batch-updates out-of-band fields for agent hosts.
  NodeOpsUpdateFieldsAgent: async <Request = NodeOpsUpdateFieldsAgentReq, ResponseData = NodeOpsUpdateFieldsAgentResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/node/ops/update_fields/agent')(params, config),
  // NodeOpsUpdateFieldsProxy batch-updates out-of-band fields for proxy hosts.
  NodeOpsUpdateFieldsProxy: async <Request = NodeOpsUpdateFieldsProxyReq, ResponseData = NodeOpsUpdateFieldsProxyResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/node/ops/update_fields/proxy')(params, config),
};

