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
  change_log_en: string;
  change_log_zh: string;
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
  change_log_en: string;
  change_log_zh: string;
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

// PackageUploadOriginPluginV2Req is the request for upload origin plugin v2.
export interface PackageUploadOriginPluginV2Req {
  overwrite: boolean;
}

// PackageUploadOriginPluginV2Resp is the response for upload origin plugin v2.
export interface PackageUploadOriginPluginV2Resp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  data: PackageUploadOriginPluginV2RespData;
}

export interface PackageUploadOriginPluginV2RespData {
  upload_id: string;
  existed: boolean;
  generated: boolean;
  name: string;
  size: number;
  md5: string;
  version: string;
  description: string;
  scenario: string;
  config_file: string;
  config_format: string;
  launch_node: string;
  platforms: Platform[];
}

// PackageUploadOriginExternalPluginV2Req is the request for upload origin
// external plugin v2.
export interface PackageUploadOriginExternalPluginV2Req {
  overwrite: boolean;
}

// PackageUploadOriginExternalPluginV2Resp is the response for upload origin
// external plugin v2.
export interface PackageUploadOriginExternalPluginV2Resp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  data: PackageUploadOriginExternalPluginV2RespData;
}

export interface PackageUploadOriginExternalPluginV2RespData {
  upload_id: string;
  existed: boolean;
  generated: boolean;
  name: string;
  size: number;
  md5: string;
  version: string;
  description: string;
  scenario: string;
  config_file: string;
  config_format: string;
  launch_node: string;
  platforms: Platform[];
}

// PackageUploadOriginPluginV3Req is the request for upload origin plugin v3.
export interface PackageUploadOriginPluginV3Req {
  overwrite: boolean;
}

// PackageUploadOriginPluginV3Resp is the response for upload origin plugin v3.
export interface PackageUploadOriginPluginV3Resp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  data: PackageUploadOriginPluginV3RespData;
}

export interface PackageUploadOriginPluginV3RespData {
  upload_id: string;
  existed: boolean;
  generated: boolean;
  name: string;
  size: number;
  md5: string;
  version: string;
  description: string;
  descriptionEn: string;
  scenario: string;
  scenarioEn: string;
  launch_node: string;
  template_renderer: string;
  platforms: Platform[];
}

// UploadOriginPluginBinToolReq is the request for upload origin plugin bin tool
// pkg.
export interface PackageUploadOriginPluginBinToolReq {
  overwrite: boolean;
}

// UploadOriginPluginBinToolResp is the response for upload origin plugin bin
// tool pkg.
export interface PackageUploadOriginPluginBinToolResp {
  code: number;
  message: string;
  request_id: string;
  data: PackageUploadOriginPluginBinToolRespData;
}

export interface PackageUploadOriginPluginBinToolRespData {
  upload_id: string;
  existed: boolean;
  generated: boolean;
  name: string;
  size: number;
  md5: string;
  v2: DataV2Info;
  v3: DataV3Info;
}

export interface DataV2Info {
  platforms: Platform[];
}

export interface DataV3Info {
  platforms: Platform[];
}

// PackagePublishReleasePluginV2Req is the request for publish release plugin v2
// pkg.
export interface PackagePublishReleasePluginV2Req {
  upload_id: string;
}

// PackagePublishReleasePluginV2Resp is the response for publish release plugin
// v2 pkg.
export interface PackagePublishReleasePluginV2Resp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  data: PackagePublishReleasePluginV2RespData;
}

export interface PackagePublishReleasePluginV2RespData {
}

// PackagePublishReleaseExternalPluginV2Req is the request for publish release
// external plugin v2 pkg.
export interface PackagePublishReleaseExternalPluginV2Req {
  upload_id: string;
}

// PackagePublishReleaseExternalPluginV2Resp is the response for publish release
// external plugin v2 pkg.
export interface PackagePublishReleaseExternalPluginV2Resp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  data: PackagePublishReleaseExternalPluginV2RespData;
}

export interface PackagePublishReleaseExternalPluginV2RespData {
}

// PackagePublishReleasePluginV3Req is the request for publish release plugin v3
// pkg.
export interface PackagePublishReleasePluginV3Req {
  upload_id: string;
}

// PackagePublishReleasePluginV3Resp is the response for publish release plugin
// v3 pkg.
export interface PackagePublishReleasePluginV3Resp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  data: PackagePublishReleasePluginV3RespData;
}

export interface PackagePublishReleasePluginV3RespData {
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

// PublishReleasePluginBinToolReq is the request for upload release plugin
// bintool pkg.
export interface PackagePublishReleasePluginBinToolReq {
  upload_id: string;
}

// PublishReleasePluginBinToolResp is the response for upload release plugin
// bintool pkg.
export interface PackagePublishReleasePluginBinToolResp {
  code: number;
  message: string;
  request_id: string;
  data: PackagePublishReleasePluginBinToolRespData;
}

export interface PackagePublishReleasePluginBinToolRespData {
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
  release_type: string;
  generation: number;
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

// PackageReleaseAgentListReq describes the HTTP request body when list package
// release.
export interface PackageReleaseAgentListReq {
  page: Page;
  generation: number;
  only_count: boolean;
  exact_include_conditions: PackageReleaseExactConditions;
}

// PackageReleaseAgentListResp describes the HTTP response body when list
// package release.
export interface PackageReleaseAgentListResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  data: PackageReleaseAgentListRespData;
}

export interface PackageReleaseAgentListRespData {
  total: number;
  items: ReleaseAgent[];
}

// PackageReleaseProxyListReq describes the HTTP request body when list package
// release.
export interface PackageReleaseProxyListReq {
  page: Page;
  generation: number;
  only_count: boolean;
  exact_include_conditions: PackageReleaseExactConditions;
}

// PackageReleaseProxyListResp describes the HTTP response body when list
// package release.
export interface PackageReleaseProxyListResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  data: PackageReleaseProxyListRespData;
}

export interface PackageReleaseProxyListRespData {
  total: number;
  items: ReleaseProxy[];
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

// PackageReleaseSetLabelsManyIdentity describes the HTTP request body.
export interface PackageReleaseSetLabelsManyIdentity {
  generation: number;
  platform: Platform;
  version: string;
}

// PackageReleaseSetLabelsManyReq describes the HTTP request body when set many
// labels.
export interface PackageReleaseSetLabelsManyReq {
  release_type: string;
  identify: PackageReleaseSetLabelsManyIdentity[];
  labels: string[];
}

// PackageReleaseSetLabelsManyResp describes the HTTP response body when set
// many labels.
export interface PackageReleaseSetLabelsManyResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  data: PackageReleaseSetLabelsManyRespData;
}

export interface PackageReleaseSetLabelsManyRespData {
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

// PackageReleaseAgentDownloadReq is the request for download agent pkg.
export interface PackageReleaseAgentDownloadReq {
  generation: number;
  platform: Platform;
  version: string;
}

// PackageReleaseProxyDownloadReq is the request for download proxy pkg.
export interface PackageReleaseProxyDownloadReq {
  generation: number;
  platform: Platform;
  version: string;
}

// PackageReleasePluginDownloadReq is the request for download plugin pkg.
export interface PackageReleasePluginDownloadReq {
  name: string;
  platform: Platform;
  version: string;
}

// FileChunk describes the file chunk data.
export interface FileChunk {
  content: bytes;
}

// PackageEventExactConditions describes the conditions when list event.
export interface PackageEventExactConditions {
  generation: number[];
  os_type: string[];
  cpu_arch: string[];
  release_type: string[];
  operator: string[];
  event_type: string[];
  version: string[];
}

// PackageEventFuzzyConditions describes the conditions when list event
export interface PackageEventFuzzyConditions {
}

// PackageEventListReq describes the HTTP request body when list event in
// Package service.
export interface PackageEventListReq {
  page: Page;
  only_count: boolean;
  exact_include_conditions: PackageEventExactConditions;
  fuzzy_include_conditions: PackageEventFuzzyConditions;
  operate_time_range: TimeRange;
}

// PackageEventListResp describes the HTTP response body when list event in
// Package service.
export interface PackageEventListResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  data: PackageEventListRespData;
}

export interface PackageEventListRespData {
  total: number;
  items: PackageEvent[];
}

// PackageEventDistinctReq describes the HTTP request body when distinct
// packageevent in package service.
export interface PackageEventDistinctReq {
  exact_include_conditions: PackageEventExactConditions;
  fuzzy_include_conditions: PackageEventFuzzyConditions;
  operate_time_range: TimeRange;
}

// PackageEventDistinctResp describes the HTTP response body when distinct
// packageevent in package service.
export interface PackageEventDistinctResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  data: PackageEventDistinctRespData;
}

export interface PackageEventDistinctRespData {
  release_type: string[];
  event_type: string[];
  os_type: string[];
  cpu_arch: string[];
  version: string[];
  operator: string[];
}

