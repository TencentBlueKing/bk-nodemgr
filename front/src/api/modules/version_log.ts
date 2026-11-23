// gen-api.js 自动生成，请勿手动修改
import { type Config } from '../interceptors';

import type { VersionLogDetailReq, VersionLogDetailResp, VersionLogsListReq, VersionLogsListResp, VersionOverviewDetailReq, VersionOverviewDetailResp, VersionOverviewListReq, VersionOverviewListResp } from '@/@types/version_log';
import Fetch from '@/api/fetch';

const fetch = new Fetch({
  prefix: `${import.meta.env.BK_API_PREFIX}`,
});

export const VersionLogService = {
  // 查询所有可用版本的更新日志列表。
  List: async <Request = VersionLogsListReq, ResponseData = VersionLogsListResp['data']>(params?: Request, config?: Config) => await fetch.get<Request, ResponseData>('/api/v3/version_log/version_logs_list')(params, config),
  // 查询当前语言下已有的大版本概览列表及当前所属大版本。
  OverviewList: async <Request = VersionOverviewListReq, ResponseData = VersionOverviewListResp['data']>(params?: Request, config?: Config) => await fetch.get<Request, ResponseData>('/api/v3/version_log/overview_list')(params, config),
  // 查询指定大版本的概览文档；当前语言没有对应文档时返回记录不存在。
  OverviewDetail: async <Request = VersionOverviewDetailReq, ResponseData = VersionOverviewDetailResp['data']>(params?: Request, config?: Config) => await fetch.get<Request, ResponseData>('/api/v3/version_log/overview/{version}')(params, config),
  // 获取指定版本的更新日志。
  Detail: async <Request = VersionLogDetailReq, ResponseData = VersionLogDetailResp['data']>(params?: Request, config?: Config) => await fetch.get<Request, ResponseData>('/api/v3/version_log/changelog/{version}')(params, config),
};

