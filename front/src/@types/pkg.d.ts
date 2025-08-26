// gen-api.js 自动生成，请勿手动修改
// PackageUploadOriginAgentReq is the request for upload origin agent pkg.
export interface PackageUploadOriginAgentReq {
  generation: number;
  overwrite: boolean;
}

// PackageUploadOriginAgentResp is the response for upload origin agent pkg.
export interface PackageUploadOriginAgentResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  data: PackageUploadOriginAgentRespData;
}

export interface PackageUploadOriginAgentRespData {
  upload_id: string;
  existed: boolean;
  generated: boolean;
  name: string;
  size: number;
  md5: string;
  version: string;
  changelog_en: string;
  changelog_zh: string;
  platforms: Platform[];
}

// PackageUploadOriginServerReq is the request for upload origin server pkg.
export interface PackageUploadOriginServerReq {
  generation: number;
  overwrite: boolean;
}

// PackageUploadOriginServerResp is the response for upload origin server pkg.
export interface PackageUploadOriginServerResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  data: PackageUploadOriginServerRespData;
}

export interface PackageUploadOriginServerRespData {
  upload_id: string;
  existed: boolean;
  generated: boolean;
  name: string;
  size: number;
  md5: string;
  version: string;
  platforms: Platform[];
}

// PackageUploadOriginCertResp is the response for upload origin cert pkg.
export interface PackageUploadOriginCertReq {
  overwrite: boolean;
}

// PackageUploadOriginCertResp is the response for upload origin cert pkg.
export interface PackageUploadOriginCertResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  data: PackageUploadOriginCertRespData;
}

export interface PackageUploadOriginCertRespData {
  upload_id: string;
  existed: boolean;
  generated: boolean;
  name: string;
  size: number;
  md5: string;
  cert_files: string[];
}

// PackageUploadOriginBinToolReq is the request for upload origin bin tool pkg.
export interface PackageUploadOriginBinToolReq {
  generation: number;
  overwrite: boolean;
}

// PackageUploadOriginBinToolResp is the response for upload origin bin tool
// pkg.
export interface PackageUploadOriginBinToolResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  data: PackageUploadOriginBinToolRespData;
}

export interface PackageUploadOriginBinToolRespData {
  upload_id: string;
  existed: boolean;
  generated: boolean;
  name: string;
  size: number;
  md5: string;
  agent_platforms: Platform[];
  proxy_platforms: Platform[];
}

// PackagePublishReleaseAgentReq is the request for upload release agent pkg.
export interface PackagePublishReleaseAgentReq {
  upload_id: string;
}

// PackagePublishReleaseAgentResp is the response for upload release agent pkg.
export interface PackagePublishReleaseAgentResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  data: PackagePublishReleaseAgentRespData;
}

export interface PackagePublishReleaseAgentRespData {
}

// PackagePublishReleaseProxyReq is the request for upload release proxy pkg.
export interface PackagePublishReleaseProxyReq {
  upload_id: string;
}

// PackagePublishReleaseProxyResp is the response for upload release proxy pkg.
export interface PackagePublishReleaseProxyResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  data: PackagePublishReleaseProxyRespData;
}

export interface PackagePublishReleaseProxyRespData {
}

// PackagePublishReleaseCertReq is the request for upload release cert pkg.
export interface PackagePublishReleaseCertReq {
  upload_id: string;
}

// PackagePublishReleaseCertResp is the response for upload release cert pkg.
export interface PackagePublishReleaseCertResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  data: PackagePublishReleaseCertRespData;
}

export interface PackagePublishReleaseCertRespData {
}

// PackagePublishReleaseBinToolReq is the request for upload release bintool
// pkg.
export interface PackagePublishReleaseBinToolReq {
  upload_id: string;
}

// PackagePublishReleaseBinToolResp is the response for upload release bintool
// pkg.
export interface PackagePublishReleaseBinToolResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  data: PackagePublishReleaseBinToolRespData;
}

export interface PackagePublishReleaseBinToolRespData {
}

// PackageReleaseExactConditions describes release exact conditions.
export interface PackageReleaseExactConditions {
  generation: number[];
  release_type: string[];
  platform: Platform[];
  version: string[];
  as_default: boolean[];
  enabled: boolean[];
}

// PackageReleaseListReq describes the HTTP request body when list package
// release.
export interface PackageReleaseListReq {
  page: Page;
  only_count: boolean;
  exact_include_conditions: PackageReleaseExactConditions;
}

// PackageReleaseListResp describes the HTTP response body when list package
// release.
export interface PackageReleaseListResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  data: PackageReleaseListRespData;
}

export interface PackageReleaseListRespData {
  total: number;
  items: Release[];
}

// PackageReleaseSetLabelsReq describes the HTTP request body when set labels.
export interface PackageReleaseSetLabelsReq {
  generation: number;
  release_type: string;
  platform: Platform;
  version: string;
  labels: string[];
}

// PackageReleaseSetLabelsResp describes the HTTP response body when set
// labels.
export interface PackageReleaseSetLabelsResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  data: PackageReleaseSetLabelsRespData;
}

export interface PackageReleaseSetLabelsRespData {
}

// PackageReleaseEnableReq describes the HTTP request body when enable package
// release.
export interface PackageReleaseEnableReq {
  generation: number;
  release_type: string;
  platform: Platform;
  version: string;
}

// PackageReleaseEnableResp describes the HTTP response body when enable package
// release.
export interface PackageReleaseEnableResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  data: PackageReleaseEnableRespData;
}

export interface PackageReleaseEnableRespData {
}

// PackageReleaseDisableReq describes the HTTP request body when disable package
// release.
export interface PackageReleaseDisableReq {
  generation: number;
  release_type: string;
  platform: Platform;
  version: string;
}

// PackageReleaseDisableResp describes the HTTP response body when disable
// package release.
export interface PackageReleaseDisableResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  data: PackageReleaseDisableRespData;
}

export interface PackageReleaseDisableRespData {
}

// PackageReleaseSetAsDefaultReq describes the HTTP request body when set
// default package release.
export interface PackageReleaseSetAsDefaultReq {
  generation: number;
  release_type: string;
  platform: Platform;
  version: string;
}

// PackageReleaseSetAsDefaultResp describes the HTTP response body when set
// default package release.
export interface PackageReleaseSetAsDefaultResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  data: PackageReleaseSetAsDefaultRespData;
}

export interface PackageReleaseSetAsDefaultRespData {
}

// PackageReleaseCancelAsDefaultReq describes the HTTP request body when cancel
// default package release.
export interface PackageReleaseCancelAsDefaultReq {
  generation: number;
  release_type: string;
  platform: Platform;
  version: string;
}

// PackageReleaseCancelAsDefaultResp describes the HTTP response body when
// cancel default package release.
export interface PackageReleaseCancelAsDefaultResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  data: PackageReleaseCancelAsDefaultRespData;
}

export interface PackageReleaseCancelAsDefaultRespData {
}

// PackageReleaseDeleteReq describes the HTTP request body when delete release.
export interface PackageReleaseDeleteReq {
  generation: number;
  release_type: string;
  platform: Platform;
  version: string;
}

// PackageReleaseDeleteResp describes the HTTP response body when delete
// release.
export interface PackageReleaseDeleteResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  data: PackageReleaseDeleteRespData;
}

export interface PackageReleaseDeleteRespData {
}

// CountRequestItem describes  the HTTP request body when count release
// deployed_host.
export interface CountRequestItem {
  generation: number;
  release_type: string;
  platform: Platform;
  version: string;
}

// PackageReleaseDeployedHostCountReq describes the HTTP request body when count
// deployed host.
export interface PackageReleaseDeployedHostCountReq {
  request_items: CountRequestItem[];
}

// PackageReleaseDeployedHostCountResp describes the HTTP response body when
// count deployed host.
export interface PackageReleaseDeployedHostCountResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  data: PackageReleaseDeployedHostCountRespData;
}

export interface PackageReleaseDeployedHostCountRespData {
  total: number;
  items: number[];
}

