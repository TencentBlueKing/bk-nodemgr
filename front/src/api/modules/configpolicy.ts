// gen-api.js 自动生成，请勿手动修改
import { type Config } from '../interceptors';

import type { ConfigPolicyListReq, ConfigPolicyListResp, ConfigPolicyGetReq, ConfigPolicyGetResp, ConfigPolicyListPlatformReq, ConfigPolicyListPlatformResp, ConfigPolicyGetTemplateReq, ConfigPolicyGetTemplateResp, ConfigPolicyCreateReq, ConfigPolicyCreateResp, ConfigPolicyUpdateReq, ConfigPolicyUpdateResp, ConfigPolicyEnableReq, ConfigPolicyEnableResp, ConfigPolicyDisableReq, ConfigPolicyDisableResp, ConfigPolicyDeleteReq, ConfigPolicyDeleteResp, ConfigPolicyEventListReq, ConfigPolicyEventListResp, ConfigPolicyEventDistinctReq, ConfigPolicyEventDistinctResp } from '@/@types/configpolicy';
import Fetch from '@/api/fetch';

const fetch = new Fetch({
  prefix: `${import.meta.env.BK_API_PREFIX}`,
});

export const ConfigPolicyAPIService = {
  // ListConfigPolicy lists config policy.
  ConfigPolicyList: async <Request = ConfigPolicyListReq, ResponseData = ConfigPolicyListResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/policy/config/list')(params, config),
  // GetConfigPolicy gets config policy.
  ConfigPolicyGet: async <Request = ConfigPolicyGetReq, ResponseData = ConfigPolicyGetResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/policy/config/get')(params, config),
  // ListConfigPolicyPlatform lists config policy platform.
  ConfigPolicyListPlatform: async <Request = ConfigPolicyListPlatformReq, ResponseData = ConfigPolicyListPlatformResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/policy/config/list_platform')(params, config),
  // TemplateConfigPolicy gets config policy template.
  ConfigPolicyTemplate: async <Request = ConfigPolicyGetTemplateReq, ResponseData = ConfigPolicyGetTemplateResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/policy/config/get_template')(params, config),
  // CreateConfigPolicy creates config policy.
  ConfigPolicyCreate: async <Request = ConfigPolicyCreateReq, ResponseData = ConfigPolicyCreateResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/policy/config/create')(params, config),
  // UpdateConfigPolicy updates config policy.
  ConfigPolicyUpdate: async <Request = ConfigPolicyUpdateReq, ResponseData = ConfigPolicyUpdateResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/policy/config/update')(params, config),
  // EnableConfigPolicy enables config policy.
  ConfigPolicyEnable: async <Request = ConfigPolicyEnableReq, ResponseData = ConfigPolicyEnableResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/policy/config/enable')(params, config),
  // DisableConfigPolicy disables config policy.
  ConfigPolicyDisable: async <Request = ConfigPolicyDisableReq, ResponseData = ConfigPolicyDisableResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/policy/config/disable')(params, config),
  // DeleteConfigPolicy deletes config policy.
  ConfigPolicyDelete: async <Request = ConfigPolicyDeleteReq, ResponseData = ConfigPolicyDeleteResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/policy/config/delete')(params, config),
  // ConfigPolicyEventList provides policy event listing.
  ConfigPolicyEventList: async <Request = ConfigPolicyEventListReq, ResponseData = ConfigPolicyEventListResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/policy/config/event/list')(params, config),
  // ConfigPolicyEventDistinct provides policy event distincting.
  ConfigPolicyEventDistinct: async <Request = ConfigPolicyEventDistinctReq, ResponseData = ConfigPolicyEventDistinctResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/policy/config/event/distinct')(params, config),
};

