// gen-api.js 自动生成，请勿手动修改
import { type Config } from '../interceptors';

import type { AuthVerifyReq, AuthVerifyResp, AuthorizedReq, AuthorizedResp } from '@/@types/auth';
import Fetch from '@/api/fetch';

const fetch = new Fetch({
  prefix: `${import.meta.env.BK_API_PREFIX}`,
});

export const AuthService = {
  // Verify checks whether the current user has permissions for the requested
  // resources.
  Verify: async <Request = AuthVerifyReq, ResponseData = AuthVerifyResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/auth/verify')(params, config),
  // Authorized queries the authorized resource scope for the requested
  // action-resource type pairs.
  Authorized: async <Request = AuthorizedReq, ResponseData = AuthorizedResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/auth/authorized')(params, config),
};

