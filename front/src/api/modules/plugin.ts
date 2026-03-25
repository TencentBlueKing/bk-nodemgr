// gen-api.js 自动生成，请勿手动修改
import { type Config } from '../interceptors';

import type { PluginInstallReq, PluginInstallResp, PluginUpgradeReq, PluginUpgradeResp, PluginUninstallReq, PluginUninstallResp, PluginApplySubConfigReq, PluginApplySubConfigResp, PluginListReq, PluginListResp, PluginSetMemoReq, PluginSetMemoResp, PluginListPermittedOperationReq, PluginListPermittedOperationResp } from '@/@types/plugin';
import Fetch from '@/api/fetch';

const fetch = new Fetch({
  prefix: `${import.meta.env.BK_API_PREFIX}`,
});

export const PluginAPIService = {
  // InstallPlugin installs a plugin on specified hosts.
  InstallPlugin: async <Request = PluginInstallReq, ResponseData = PluginInstallResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/plugin/install')(params, config),
  // UpgradePlugin upgrades a plugin on specified hosts.
  UpgradePlugin: async <Request = PluginUpgradeReq, ResponseData = PluginUpgradeResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/plugin/upgrade')(params, config),
  // UninstallPlugin uninstalls a plugin on specified hosts.
  UninstallPlugin: async <Request = PluginUninstallReq, ResponseData = PluginUninstallResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/plugin/uninstall')(params, config),
  // ApplyPluginSubConfig apply sub-configuration for a plugin on specified
  // hosts.
  ApplyPluginSubConfig: async <Request = PluginApplySubConfigReq, ResponseData = PluginApplySubConfigResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/plugin/apply_subconfig')(params, config),
  // ListPlugins lists plugins based on the given conditions.
  ListPlugins: async <Request = PluginListReq, ResponseData = PluginListResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/plugin/list')(params, config),
  // SetPluginMemo sets the memo for a plugin.
  SetPluginMemo: async <Request = PluginSetMemoReq, ResponseData = PluginSetMemoResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/plugin/set_memo')(params, config),
  // ListPluginPermittedOperation lists the plugin operations that the user has
  // permission to perform.
  ListPluginPermittedOperation: async <Request = PluginListPermittedOperationReq, ResponseData = PluginListPermittedOperationResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/plugin/list_permitted_operations')(params, config),
};

