// gen-api.js 自动生成，请勿手动修改
import { type Config } from '../interceptors';

import type { PluginWorkflowListReq, PluginWorkflowListResp, PluginWorkflowStatisticsReq, PluginWorkflowStatisticsResp, PluginWorkflowDistinctReq, PluginWorkflowDistinctResp, PluginWorkflowOperationListReq, PluginWorkflowOperationListResp, PluginWorkflowOperationInstanceListReq, PluginWorkflowOperationInstanceListResp, PluginWorkflowOperationInstanceLogGetReq, PluginWorkflowOperationInstanceLogGetResp, PluginWorkflowOperationRetryReq, PluginWorkflowOperationRetryResp, PluginWorkflowOperationTerminateReq, PluginWorkflowOperationTerminateResp } from '@/@types/plugin_workflow';
import Fetch from '@/api/fetch';

const fetch = new Fetch({
  prefix: `${import.meta.env.BK_API_PREFIX}`,
});

export const PluginWorkflowService = {
  // PluginWorkflowList provides plugin workflow list.
  PluginWorkflowList: async <Request = PluginWorkflowListReq, ResponseData = PluginWorkflowListResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/plugin/workflow/list')(params, config),
  // PluginWorkflowStatistics provides plugin workflow statistics.
  PluginWorkflowStatistics: async <Request = PluginWorkflowStatisticsReq, ResponseData = PluginWorkflowStatisticsResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/plugin/workflow/statistics')(params, config),
  // PluginWorkflowDistinct provides plugin workflow distinct.
  PluginWorkflowDistinct: async <Request = PluginWorkflowDistinctReq, ResponseData = PluginWorkflowDistinctResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/plugin/workflow/distinct')(params, config),
  // PluginWorkflowOperationList provides plugin workflow operation list.
  PluginWorkflowOperationList: async <Request = PluginWorkflowOperationListReq, ResponseData = PluginWorkflowOperationListResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/plugin/workflow/operation/list')(params, config),
  // PluginWorkflowOperationInstanceList provides plugin workflow operation
  // instance list.
  PluginWorkflowOperationInstanceList: async <Request = PluginWorkflowOperationInstanceListReq, ResponseData = PluginWorkflowOperationInstanceListResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/plugin/workflow/operation/instance/list')(params, config),
  // PluginWorkflowOperationInstanceLogGet provides plugin operation instance
  // log get.
  PluginWorkflowOperationInstanceLogGet: async <Request = PluginWorkflowOperationInstanceLogGetReq, ResponseData = PluginWorkflowOperationInstanceLogGetResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/plugin/workflow/operation/instance/log/get')(params, config),
  // PluginWorkflowOperationRetry provides plugin operation retry.
  PluginWorkflowOperationRetry: async <Request = PluginWorkflowOperationRetryReq, ResponseData = PluginWorkflowOperationRetryResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/plugin/workflow/operation/retry')(params, config),
  // PluginWorkflowOperationTerminate provides plugin operation terminate.
  PluginWorkflowOperationTerminate: async <Request = PluginWorkflowOperationTerminateReq, ResponseData = PluginWorkflowOperationTerminateResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/plugin/workflow/operation/terminate')(params, config),
};

