// gen-api.js 自动生成，请勿手动修改
import { type Config } from '../interceptors';

import type { PluginInstallReq, PluginInstallResp, PluginApplySubConfigReq, PluginApplySubConfigResp, PluginListReq, PluginListResp } from '@/@types/plugin';
import Fetch from '@/api/fetch';

const fetch = new Fetch({
  prefix: `${import.meta.env.BK_API_PREFIX}`,
});

export const PluginService = {
  // InstallPlugin installs a plugin on specified hosts.
  InstallPlugin: async <Request = PluginInstallReq, ResponseData = PluginInstallResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/plugin/install')(params, config),
  // ApplyPluginSubConfig apply sub-configuration for a plugin on specified
  // hosts.
  ApplyPluginSubConfig: async <Request = PluginApplySubConfigReq, ResponseData = PluginApplySubConfigResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/plugin/apply_subconfig')(params, config),
  // ListPlugins lists plugins based on the given conditions.
  ListPlugins: async <Request = PluginListReq, ResponseData = PluginListResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/plugin/list')(params, config),
};

