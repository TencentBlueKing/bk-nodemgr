// gen-api.js 自动生成，请勿手动修改
// GetCurrentAnnouncementsReq describes the request to get current
// announcements. All parameters (platform and username) are automatically
// obtained from config and context.
export interface GetCurrentAnnouncementsReq {
}

// AnnouncementContent describes announcement content in a specific language.
export interface AnnouncementContent {
  // content is the announcement content text in the specified language.
  content: string;
  // language is the language code (e.g., "zh-cn", "en").
  language: string;
}

// Announcement describes a platform announcement from notice center.
export interface Announcement {
  // id is the unique identifier of the announcement.
  id: number;
  // title is the announcement title.
  title: string;
  // content_list contains announcement content in multiple languages.
  content_list: AnnouncementContent[];
  // content is the default announcement content (for backward compatibility).
  content: string;
  // announce_type is the type of announcement.
  // Valid values: "event" (activity notification) or "announce" (platform
  // announcement).
  announce_type: string;
  // start_time is when the announcement becomes active (RFC3339 format).
  start_time: string;
  // end_time is when the announcement expires (RFC3339 format).
  end_time: string;
}

// GetCurrentAnnouncementsResp describes the response for getting current
// announcements.
export interface GetCurrentAnnouncementsResp {
  // code is the response status code.
  code: number;
  // message is the response message.
  message: string;
  // request_id is the unique request identifier for tracing.
  request_id: string;
  // error contains error details if the request failed.
  error: Error;
  permission: Permission;
  // data contains the list of current announcements.
  data: Announcement[];
}

