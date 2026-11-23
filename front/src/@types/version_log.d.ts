// gen-api.js 自动生成，请勿手动修改
// VersionLogEntry represents a single version's changelog entry.
export interface VersionLogEntry {
  // version is the version identifier (e.g. "v3.0.1-alpha.17").
  version: string;
  // date is the release date in YYYY-MM-DD format.
  date: string;
  // is_current indicates whether this is the currently running version.
  is_current: boolean;
}

// VersionLogsListReq describes the HTTP request body for listing versions.
export interface VersionLogsListReq {
}

// VersionLogsListResp describes the HTTP response body for version list.
export interface VersionLogsListResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: VersionLogsListRespData;
}

export interface VersionLogsListRespData {
  // version_logs lists all available versions sorted by recency (newest
  // first).
  version_logs: VersionLogEntry[];
}

// VersionLogDetailReq describes the HTTP request body for fetching a specific
// version's changelog.
export interface VersionLogDetailReq {
  // version is the version identifier to fetch (e.g. "v3.0.1-alpha.17").
  version: string;
}

// VersionLogDetailResp describes the HTTP response body for version detail.
export interface VersionLogDetailResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: VersionLogDetailRespData;
}

export interface VersionLogDetailRespData {
  // version is the requested version identifier.
  version: string;
  // date is the release date in YYYY-MM-DD format.
  date: string;
  // content is the markdown-formatted changelog content.
  content: string;
}

// 独立维护的大版本概览，不按发布日期区分。
export interface VersionOverviewEntry {
  // 三段基础版本号，例如 v3.0.1。
  version: string;
  // 是否为当前运行版本所属的大版本。
  is_current: boolean;
}

// 按当前语言查询独立维护的大版本概览。
export interface VersionOverviewListReq {
}

// 大版本概览列表及当前运行版本所属的大版本。
export interface VersionOverviewListResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: VersionOverviewListRespData;
}

export interface VersionOverviewListRespData {
  // 当前语言下已有概览文档的大版本，按版本号降序排列。
  version_logs: VersionOverviewEntry[];
  // 当前运行版本所属的大版本，例如 v3.0.1；可能尚无对应概览文档。
  // 运行版本不符合项目版本格式时为空。
  current_version: string;
}

// 查询指定大版本的独立概览文档。
export interface VersionOverviewDetailReq {
  // 三段基础版本号，例如 v3.0.1，不包含 alpha、beta 或 release 后缀。
  version: string;
}

// 当前语言下的大版本概览内容。
export interface VersionOverviewDetailResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: VersionOverviewDetailRespData;
}

export interface VersionOverviewDetailRespData {
  // 请求的大版本号，例如 v3.0.1。
  version: string;
  // 独立维护、可持续更新的 Markdown 展示文档，不由迭代日志自动拼接。
  content: string;
}

