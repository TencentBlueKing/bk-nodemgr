// gen-api.js 自动生成，请勿手动修改
import { type Config } from '../interceptors';

import type { ListGlobalSettingsReq, ListGlobalSettingsResp } from '@/@types/globalsettings';
import Fetch from '@/api/fetch';

const fetch = new Fetch({
  prefix: `${import.meta.env.BK_API_PREFIX}`,
});

export const GlobalSettingsService = {
  // 查询跨租户共享的已存储配置及原始值，包含可能敏感的自定义项；须具有系统访问和平台管理的全局配置资源权限。
  ListGlobalSettings: async <Request = ListGlobalSettingsReq, ResponseData = ListGlobalSettingsResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/globalsettings/list')(params, config),
};
