// gen-api.js 自动生成，请勿手动修改
import { type Config } from '../interceptors';

import type { PluginApplySubConfigReq, PluginApplySubConfigResp, PluginDistinctReq, PluginDistinctResp, PluginInstallReq, PluginInstallResp, PluginListPermittedOperationReq, PluginListPermittedOperationResp, PluginListReq, PluginListResp, PluginRestartReq, PluginRestartResp, PluginSetMemoReq, PluginSetMemoResp, PluginStartReq, PluginStartResp, PluginStopReq, PluginStopResp, PluginUninstallReq, PluginUninstallResp, PluginUpgradeReq, PluginUpgradeResp } from '@/@types/plugin';
import Fetch from '@/api/fetch';

const fetch = new Fetch({
  prefix: `${import.meta.env.BK_API_PREFIX}`,
});

export const PluginAPIService = {
  // 在指定主机上安装插件。
  InstallPlugin: async <Request = PluginInstallReq, ResponseData = PluginInstallResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/plugin/install')(params, config),
  // 在指定主机上升级插件。
  UpgradePlugin: async <Request = PluginUpgradeReq, ResponseData = PluginUpgradeResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/plugin/upgrade')(params, config),
  // 在指定主机上卸载插件。
  UninstallPlugin: async <Request = PluginUninstallReq, ResponseData = PluginUninstallResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/plugin/uninstall')(params, config),
  // 在指定主机上应用插件子配置。
  ApplyPluginSubConfig: async <Request = PluginApplySubConfigReq, ResponseData = PluginApplySubConfigResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/plugin/apply_subconfig')(params, config),
  // 按条件查询插件列表。
  ListPlugins: async <Request = PluginListReq, ResponseData = PluginListResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/plugin/list')(params, config),
  // 查询当前租户可见插件的属性去重值，包含尚未安装的插件。
  DistinctPlugin: async <Request = PluginDistinctReq, ResponseData = PluginDistinctResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/plugin/distinct')(params, config),
  // 设置插件备注。
  SetPluginMemo: async <Request = PluginSetMemoReq, ResponseData = PluginSetMemoResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/plugin/set_memo')(params, config),
  // 查询当前用户有权执行的插件操作。
  ListPluginPermittedOperation: async <Request = PluginListPermittedOperationReq, ResponseData = PluginListPermittedOperationResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/plugin/list_permitted_operations')(params, config),
  // 在指定主机上启动插件进程。
  StartPlugin: async <Request = PluginStartReq, ResponseData = PluginStartResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/plugin/start')(params, config),
  // 在指定主机上重启插件进程。
  RestartPlugin: async <Request = PluginRestartReq, ResponseData = PluginRestartResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/plugin/restart')(params, config),
  // 在指定主机上停止插件进程。
  StopPlugin: async <Request = PluginStopReq, ResponseData = PluginStopResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/plugin/stop')(params, config),
};

