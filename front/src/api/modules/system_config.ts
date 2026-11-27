// gen-api.js 自动生成，请勿手动修改
import { type Config } from '../interceptors';

import type { GetSystemConfigReq, GetSystemConfigResp } from '@/@types/system_config';
import Fetch from '@/api/fetch';

const fetch = new Fetch({
  prefix: `${import.meta.env.BK_API_PREFIX}`,
});

export const SystemConfigQueryService = {
  // 查询 Backend
  // 实例加载的业务配置；须具有系统访问和平台管理的系统配置资源权限。
  GetSystemConfig: async <Request = GetSystemConfigReq, ResponseData = GetSystemConfigResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/system_config/get')(params, config),
};

