// gen-api.js 自动生成，请勿手动修改
import { type Config } from '../interceptors';

import type { DeployPolicyListReq, DeployPolicyListResp, DeployPolicyExecuteReq, DeployPolicyExecuteResp } from '@/@types/deploy_policy';
import Fetch from '@/api/fetch';

const fetch = new Fetch({
  prefix: `${import.meta.env.BK_API_PREFIX}`,
});

export const DeployPolicyAPIService = {
  // ListDeployPolicy lists deploy policy.
  ListDeployPolicy: async <Request = DeployPolicyListReq, ResponseData = DeployPolicyListResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/deploy_policy/list')(params, config),
  // ExecuteDeployPolicy executes deploy policy.
  ExecuteDeployPolicy: async <Request = DeployPolicyExecuteReq, ResponseData = DeployPolicyExecuteResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/deploy_policy/execute')(params, config),
};

