// gen-api.js 自动生成，请勿手动修改
import { type Config } from '../interceptors';

import type { GetCurrentAnnouncementsReq, GetCurrentAnnouncementsResp } from '@/@types/notice';
import Fetch from '@/api/fetch';

const fetch = new Fetch({
  prefix: `${import.meta.env.BK_API_PREFIX}`,
});

export const NoticeAPIService = {
  // GetCurrentAnnouncements retrieves current active announcements.
  // Platform and username are automatically obtained from config and context.
  GetCurrentAnnouncements: async <Request = GetCurrentAnnouncementsReq, ResponseData = GetCurrentAnnouncementsResp['data']>(params?: Request, config?: Config) => await fetch.get<Request, ResponseData>('/api/v3/notice/announcements/current')(params, config),
};

