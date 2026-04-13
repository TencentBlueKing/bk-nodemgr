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

