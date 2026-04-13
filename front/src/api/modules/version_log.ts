// gen-api.js 自动生成，请勿手动修改
import { type Config } from '../interceptors';

import type { VersionLogsListReq, VersionLogsListResp, VersionLogDetailReq, VersionLogDetailResp } from '@/@types/version_log';
import Fetch from '@/api/fetch';

const fetch = new Fetch({
  prefix: `${import.meta.env.BK_API_PREFIX}`,
});

export const VersionLogService = {
  // List returns all available version changelogs.
  List: async <Request = VersionLogsListReq, ResponseData = VersionLogsListResp['data']>(params?: Request, config?: Config) => await fetch.get<Request, ResponseData>('/api/v3/version_log/version_logs_list')(params, config),
  // Detail returns the changelog for a specific version.
  Detail: async <Request = VersionLogDetailReq, ResponseData = VersionLogDetailResp['data']>(params?: Request, config?: Config) => await fetch.get<Request, ResponseData>('/api/v3/version_log/changelog/{version}')(params, config),
};

