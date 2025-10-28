// gen-api.js 自动生成，请勿手动修改
import { type Config } from '../interceptors';

import type { ProcessListReq, ProcessListResp } from '@/@types/process';
import Fetch from '@/api/fetch';

const fetch = new Fetch({
  prefix: `${import.meta.env.BK_API_PREFIX}`,
});

export const PackageService = {
  // ListProcesses lists the processes.
  ListProcesses: async <Request = ProcessListReq, ResponseData = ProcessListResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/process/list')(params, config),
};

