// gen-api.js 自动生成，请勿手动修改
import { type Config } from '../interceptors';

import type { NodeWorkflowListReq, NodeWorkflowListResp, NodeWorkflowStatisticsReq, NodeWorkflowStatisticsResp, NodeWorkflowDistinctReq, NodeWorkflowDistinctResp, NodeWorkflowOperationListReq, NodeWorkflowOperationListResp, NodeWorkflowOperationDistinctReq, NodeWorkflowOperationDistinctResp, NodeWorkflowOperationInstanceListReq, NodeWorkflowOperationInstanceListResp, NodeWorkflowOperationInstanceLogGetReq, NodeWorkflowOperationInstanceLogGetResp, NodeWorkflowOperationRetryReq, NodeWorkflowOperationRetryResp, NodeWorkflowOperationTerminateReq, NodeWorkflowOperationTerminateResp, NodeWorkflowOperationManualSolutionGetReq, NodeWorkflowOperationManualSolutionGetResp, NodeWorkflowOperationOfflinePackageDownloadReq, FileChunk, NodeWorkflowOperationOfflineInstallResultSubmitReq, NodeWorkflowOperationOfflineInstallResultSubmitResp } from '@/@types/node_workflow';
import Fetch from '@/api/fetch';

const fetch = new Fetch({
  prefix: `${import.meta.env.BK_API_PREFIX}`,
});

export const NodeWorkflowService = {
  // NodeWorkflowList provides node workflow list.
  NodeWorkflowList: async <Request = NodeWorkflowListReq, ResponseData = NodeWorkflowListResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/node/workflow/list')(params, config),
  // NodeWorkflowStatistics provides node workflow statistics.
  NodeWorkflowStatistics: async <Request = NodeWorkflowStatisticsReq, ResponseData = NodeWorkflowStatisticsResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/node/workflow/statistics')(params, config),
  // NodeWorkflowDistinct provides node workflow distinct.
  NodeWorkflowDistinct: async <Request = NodeWorkflowDistinctReq, ResponseData = NodeWorkflowDistinctResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/node/workflow/distinct')(params, config),
  // NodeWorkflowOperationList provides node workflow operation list.
  NodeWorkflowOperationList: async <Request = NodeWorkflowOperationListReq, ResponseData = NodeWorkflowOperationListResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/node/workflow/operation/list')(params, config),
  // NodeWorkflowOperationDistinct provides node workflow operation distinct.
  NodeWorkflowOperationDistinct: async <Request = NodeWorkflowOperationDistinctReq, ResponseData = NodeWorkflowOperationDistinctResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/node/workflow/operation/distinct')(params, config),
  // NodeWorkflowOperationInstanceList provides node workflow operation instance
  // list.
  NodeWorkflowOperationInstanceList: async <Request = NodeWorkflowOperationInstanceListReq, ResponseData = NodeWorkflowOperationInstanceListResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/node/workflow/operation/instance/list')(params, config),
  // NodeWorkflowOperationInstanceLogGet provides node operation instance log
  // get.
  NodeWorkflowOperationInstanceLogGet: async <Request = NodeWorkflowOperationInstanceLogGetReq, ResponseData = NodeWorkflowOperationInstanceLogGetResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/node/workflow/operation/instance/log/get')(params, config),
  // NodeWorkflowOperationRetry retry node operation.
  NodeWorkflowOperationRetry: async <Request = NodeWorkflowOperationRetryReq, ResponseData = NodeWorkflowOperationRetryResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/node/workflow/operation/retry')(params, config),
  // NodeWorkflowOperationTerminate provides node operation terminate.
  NodeWorkflowOperationTerminate: async <Request = NodeWorkflowOperationTerminateReq, ResponseData = NodeWorkflowOperationTerminateResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/node/workflow/operation/terminate')(params, config),
  // NodeWorkflowOperationManualSolutionGet provides node operation manual
  // solution get.
  NodeWorkflowOperationManualSolutionGet: async <Request = NodeWorkflowOperationManualSolutionGetReq, ResponseData = NodeWorkflowOperationManualSolutionGetResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/node/workflow/operation/manual/solution/get')(params, config),
  // NodeWorkflowOperationOfflinePackageDownload streams the tar.gz offline
  // install bundle for an operation (POST
  // /api/v3/node/workflow/operation/offline/download, body: operation_id).
  NodeWorkflowOperationOfflinePackageDownload: async <Request = NodeWorkflowOperationOfflinePackageDownloadReq, ResponseData = FileChunk['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/node/workflow/operation/offline/download')(params, config),
  // NodeWorkflowOperationOfflineInstallResultSubmit forwards offline install
  // result JSON to the backend (POST
  // /api/v3/node/workflow/operation/offline/result).
  NodeWorkflowOperationOfflineInstallResultSubmit: async <Request = NodeWorkflowOperationOfflineInstallResultSubmitReq, ResponseData = NodeWorkflowOperationOfflineInstallResultSubmitResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/node/workflow/operation/offline/result')(params, config),
};

