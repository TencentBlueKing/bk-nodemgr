// gen-api.js 自动生成，请勿手动修改
import { type Config } from '../interceptors';

import type { PluginWorkflowDistinctReq, PluginWorkflowDistinctResp, PluginWorkflowListReq, PluginWorkflowListResp, PluginWorkflowOperationDistinctReq, PluginWorkflowOperationDistinctResp, PluginWorkflowOperationInstallParamsGetReq, PluginWorkflowOperationInstallParamsGetResp, PluginWorkflowOperationInstanceListReq, PluginWorkflowOperationInstanceListResp, PluginWorkflowOperationInstanceLogGetReq, PluginWorkflowOperationInstanceLogGetResp, PluginWorkflowOperationListReq, PluginWorkflowOperationListResp, PluginWorkflowOperationRetryReq, PluginWorkflowOperationRetryResp, PluginWorkflowOperationTerminateReq, PluginWorkflowOperationTerminateResp, PluginWorkflowStatisticsReq, PluginWorkflowStatisticsResp } from '@/@types/plugin_workflow';
import Fetch from '@/api/fetch';

const fetch = new Fetch({
  prefix: `${import.meta.env.BK_API_PREFIX}`,
});

export const PluginWorkflowService = {
  // 读取选中历史插件安装操作的参数，供新安装使用，不执行安装或修改历史任务。仅支持
  // install_plugin 工作流。
  PluginWorkflowOperationInstallParamsGet: async <Request = PluginWorkflowOperationInstallParamsGetReq, ResponseData = PluginWorkflowOperationInstallParamsGetResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/plugin/workflow/operation/install_params/get')(params, config),
  // 查询插件工作流列表。
  PluginWorkflowList: async <Request = PluginWorkflowListReq, ResponseData = PluginWorkflowListResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/plugin/workflow/list')(params, config),
  // 统计插件工作流的操作实例状态分布。
  PluginWorkflowStatistics: async <Request = PluginWorkflowStatisticsReq, ResponseData = PluginWorkflowStatisticsResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/plugin/workflow/statistics')(params, config),
  // 查询插件工作流筛选字段的去重值。
  PluginWorkflowDistinct: async <Request = PluginWorkflowDistinctReq, ResponseData = PluginWorkflowDistinctResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/plugin/workflow/distinct')(params, config),
  // 查询插件工作流操作列表。
  PluginWorkflowOperationList: async <Request = PluginWorkflowOperationListReq, ResponseData = PluginWorkflowOperationListResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/plugin/workflow/operation/list')(params, config),
  // 查询插件工作流操作筛选字段的去重值。
  PluginWorkflowOperationDistinct: async <Request = PluginWorkflowOperationDistinctReq, ResponseData = PluginWorkflowOperationDistinctResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/plugin/workflow/operation/distinct')(params, config),
  // 查询插件工作流操作实例列表。
  PluginWorkflowOperationInstanceList: async <Request = PluginWorkflowOperationInstanceListReq, ResponseData = PluginWorkflowOperationInstanceListResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/plugin/workflow/operation/instance/list')(params, config),
  // 获取插件工作流操作实例日志。
  PluginWorkflowOperationInstanceLogGet: async <Request = PluginWorkflowOperationInstanceLogGetReq, ResponseData = PluginWorkflowOperationInstanceLogGetResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/plugin/workflow/operation/instance/log/get')(params, config),
  // 重试插件工作流操作。
  PluginWorkflowOperationRetry: async <Request = PluginWorkflowOperationRetryReq, ResponseData = PluginWorkflowOperationRetryResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/plugin/workflow/operation/retry')(params, config),
  // 终止插件工作流操作。
  PluginWorkflowOperationTerminate: async <Request = PluginWorkflowOperationTerminateReq, ResponseData = PluginWorkflowOperationTerminateResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/plugin/workflow/operation/terminate')(params, config),
};

