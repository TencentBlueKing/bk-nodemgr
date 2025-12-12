// gen-api.js 自动生成，请勿手动修改
import { type Config } from '../interceptors';

import type { ProcessListReq, ProcessListResp, GetProcessDistributionByHostIDReq, GetProcessDistributionByHostIDResp, GetProcessDistributionByPluginNameReq, GetProcessDistributionByPluginNameResp } from '@/@types/process';
import Fetch from '@/api/fetch';

const fetch = new Fetch({
  prefix: `${import.meta.env.BK_API_PREFIX}`,
});

export const ProcessAPIService = {
  // ListProcesses lists the processes.
  ListProcesses: async <Request = ProcessListReq, ResponseData = ProcessListResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/process/list')(params, config),
  // GetProcessDistributionByHostID describes the process statistics request.
  GetProcessDistributionByHostID: async <Request = GetProcessDistributionByHostIDReq, ResponseData = GetProcessDistributionByHostIDResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/process/get_distribution_by_host_id')(params, config),
  // GetProcessDistributionByPluginName describes the process statistics
  // request.
  GetProcessDistributionByPluginName: async <Request = GetProcessDistributionByPluginNameReq, ResponseData = GetProcessDistributionByPluginNameResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/process/get_distribution_by_plugin_name')(params, config),
};

