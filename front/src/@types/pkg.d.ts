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
  permission: Permission;
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
  permission: Permission;
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

// PackageUploadOriginProxyReq is the request for upload origin proxy pkg.
export interface PackageUploadOriginProxyReq {
  generation: number;
  overwrite: boolean;
}

// PackageUploadOriginProxyResp is the response for upload origin proxy pkg.
export interface PackageUploadOriginProxyResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: PackageUploadOriginProxyRespData;
}

export interface PackageUploadOriginProxyRespData {
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
  permission: Permission;
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
  permission: Permission;
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
  permission: Permission;
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
  description_en: string;
  scenario: string;
  scenario_en: string;
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
  permission: Permission;
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
  description_en: string;
  scenario: string;
  scenario_en: string;
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
  permission: Permission;
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
  description_en: string;
  scenario: string;
  scenario_en: string;
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
  permission: Permission;
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
  permission: Permission;
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
  permission: Permission;
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
  permission: Permission;
  data: PackagePublishReleaseAgentRespData;
}

export interface PackagePublishReleaseAgentRespData {
}

// PackagePublishReleaseProxyReq is the request for upload release proxy pkg.
export interface PackagePublishReleaseProxyReq {
  upload_id: string;
  // upload_origin_pkg_type indicates the origin category of the uploaded
  // package. only accepts "origin_server" and "origin_proxy", no default value.
  upload_origin_pkg_type: string;
}

// PackagePublishReleaseProxyResp is the response for upload release proxy pkg.
export interface PackagePublishReleaseProxyResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
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
  permission: Permission;
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
  permission: Permission;
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
  platform: Platform[];
  version: string[];
  as_default: boolean[];
  enabled: boolean[];
  name: string[];
  file_name: string[];
  is_visible: boolean[];
}

// PackageReleaseDistinctField describes the release distinct field.
export interface PackageReleaseDistinctField {
  os_type: boolean;
  cpu_arch: boolean;
  name: boolean;
  version: boolean;
}

// PackageReleaseDistinctData describes the release distinct data.
export interface PackageReleaseDistinctData {
  os_type: string[];
  cpu_arch: string[];
  name: string[];
  version: string[];
}

// PackageReleasePluginDistinctReq describes the request for distinct plugin
// releases.
export interface PackageReleasePluginDistinctReq {
  generation: number;
  exact_include_conditions: PackageReleaseExactConditions;
  distinct_field: PackageReleaseDistinctField;
}

// PackageReleasePluginDistinctResp describes the response for distinct plugin
// releases.
export interface PackageReleasePluginDistinctResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: PackageReleaseDistinctData;
}

// PackageReleaseAgentListReq describes the HTTP request body when list agent
// release.
export interface PackageReleaseAgentListReq {
  page: Page;
  generation: number;
  only_count: boolean;
  exact_include_conditions: PackageReleaseExactConditions;
}

// PackageReleaseAgentListResp describes the HTTP response body when list agent
// release.
export interface PackageReleaseAgentListResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: PackageReleaseAgentListRespData;
}

export interface PackageReleaseAgentListRespData {
  total: number;
  items: ReleaseAgent[];
}

// PackageReleaseAgentListBriefReq describes the HTTP request body when list
// agent release brief.
export interface PackageReleaseAgentListBriefReq {
  page: Page;
  generation: number;
  only_count: boolean;
  exact_include_conditions: PackageReleaseExactConditions;
}

// PackageReleaseAgentListBriefResp describes the HTTP response body when list
// agent release brief.
export interface PackageReleaseAgentListBriefResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: PackageReleaseAgentListBriefRespData;
}

export interface PackageReleaseAgentListBriefRespData {
  total: number;
  items: ReleaseAgentBrief[];
}

// PackageReleaseAgentDistinctReq describes the HTTP request body when distinct
// agent release.
export interface PackageReleaseAgentDistinctReq {
  generation: number;
  exact_include_conditions: PackageReleaseExactConditions;
  distinct_field: PackageReleaseDistinctField;
}

// PackageReleaseAgentDistinctResp describes the HTTP response body when
// distinct agent release.
export interface PackageReleaseAgentDistinctResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: PackageReleaseDistinctData;
}

// PackageReleaseAgentSetLabelsManyReq describes the HTTP request body when set
// many labels to agent release.
export interface PackageReleaseAgentSetLabelsManyReq {
  generation: number;
  exact_include_conditions: PackageReleaseExactConditions;
  labels: string[];
}

// PackageReleaseAgentSetLabelsManyResp describes the HTTP response body when
// set many labels to agent release.
export interface PackageReleaseAgentSetLabelsManyResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: PackageReleaseAgentSetLabelsManyRespData;
}

export interface PackageReleaseAgentSetLabelsManyRespData {
}

// PackageReleaseAgentEnableReq describes the HTTP request body when enable
// agent release.
export interface PackageReleaseAgentEnableReq {
  generation: number;
  release_type: string;
  platform: Platform;
  version: string;
}

// PackageReleaseAgentEnableResp describes the HTTP response body when enable
// agent release.
export interface PackageReleaseAgentEnableResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: PackageReleaseAgentEnableRespData;
}

export interface PackageReleaseAgentEnableRespData {
}

// PackageReleaseAgentDisableReq describes the HTTP request body when disable
// agent release.
export interface PackageReleaseAgentDisableReq {
  generation: number;
  release_type: string;
  platform: Platform;
  version: string;
}

// PackageReleaseAgentDisableResp describes the HTTP response body when disable
// agent release.
export interface PackageReleaseAgentDisableResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: PackageReleaseAgentDisableRespData;
}

export interface PackageReleaseAgentDisableRespData {
}

export interface PackageReleaseAgentVisibleReq {
  generation: number;
  release_type: string;
  platform: Platform;
  version: string;
}

export interface PackageReleaseAgentVisibleResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: PackageReleaseAgentVisibleRespData;
}

export interface PackageReleaseAgentVisibleRespData {
}

export interface PackageReleaseAgentUnvisibleReq {
  generation: number;
  release_type: string;
  platform: Platform;
  version: string;
}

export interface PackageReleaseAgentUnvisibleResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: PackageReleaseAgentUnvisibleRespData;
}

export interface PackageReleaseAgentUnvisibleRespData {
}

// PackageReleaseAgentSetAsDefaultReq describes the HTTP request body when set
// default agent release.
export interface PackageReleaseAgentSetAsDefaultReq {
  generation: number;
  release_type: string;
  platform: Platform;
  version: string;
}

// PackageReleaseAgentSetAsDefaultResp describes the HTTP response body when set
// default agent release.
export interface PackageReleaseAgentSetAsDefaultResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: PackageReleaseAgentSetAsDefaultRespData;
}

export interface PackageReleaseAgentSetAsDefaultRespData {
}

// PackageReleaseAgentCancelAsDefaultReq describes the HTTP request body when
// cancel default agent release.
export interface PackageReleaseAgentCancelAsDefaultReq {
  generation: number;
  release_type: string;
  platform: Platform;
  version: string;
}

// PackageReleaseAgentCancelAsDefaultResp describes the HTTP response body when
// cancel default agent release.
export interface PackageReleaseAgentCancelAsDefaultResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: PackageReleaseAgentCancelAsDefaultRespData;
}

export interface PackageReleaseAgentCancelAsDefaultRespData {
}

// PackageReleaseAgentDeleteReq describes the HTTP request body when delete
// agent release.
export interface PackageReleaseAgentDeleteReq {
  generation: number;
  release_type: string;
  platform: Platform;
  version: string;
}

// PackageReleaseAgentDeleteResp describes the HTTP response body when delete
// agent release.
export interface PackageReleaseAgentDeleteResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: PackageReleaseAgentDeleteRespData;
}

export interface PackageReleaseAgentDeleteRespData {
}

// PackageReleaseAgentCountDeployedReq describes the HTTP request body when
// count agent deployed.
export interface PackageReleaseAgentCountDeployedReq {
  items: Item[];
}

export interface PackageReleaseAgentCountDeployedReqItem {
  generation: number;
  platform: Platform;
  version: string;
}

// PackageReleaseAgentCountDeployedResp describes the HTTP response body when
// count agent deployed.
export interface PackageReleaseAgentCountDeployedResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: PackageReleaseAgentCountDeployedRespData;
}

export interface PackageReleaseAgentCountDeployedRespData {
  counts: number[];
}

// PackageReleaseAgentDownloadReq is the request for download agent pkg.
export interface PackageReleaseAgentDownloadReq {
  generation: number;
  platform: Platform;
  version: string;
}

// PackageReleaseProxyListReq describes the HTTP request body when list proxy
// release.
export interface PackageReleaseProxyListReq {
  page: Page;
  generation: number;
  only_count: boolean;
  exact_include_conditions: PackageReleaseExactConditions;
}

// PackageReleaseProxyListResp describes the HTTP response body when list proxy
// release.
export interface PackageReleaseProxyListResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: PackageReleaseProxyListRespData;
}

export interface PackageReleaseProxyListRespData {
  total: number;
  items: ReleaseProxy[];
}

// PackageReleaseProxyListBriefReq describes the HTTP request body when list
// proxy release brief.
export interface PackageReleaseProxyListBriefReq {
  page: Page;
  generation: number;
  only_count: boolean;
  exact_include_conditions: PackageReleaseExactConditions;
}

// PackageReleaseProxyListBriefResp describes the HTTP response body when list
// proxy release brief.
export interface PackageReleaseProxyListBriefResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: PackageReleaseProxyListBriefRespData;
}

export interface PackageReleaseProxyListBriefRespData {
  total: number;
  items: ReleaseProxyBrief[];
}

// PackageReleaseProxyDistinctReq describes the HTTP request body when distinct
// proxy release.
export interface PackageReleaseProxyDistinctReq {
  generation: number;
  exact_include_conditions: PackageReleaseExactConditions;
  distinct_field: PackageReleaseDistinctField;
}

// PackageReleaseProxyDistinctResp describes the HTTP response body when
// distinct proxy release.
export interface PackageReleaseProxyDistinctResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: PackageReleaseDistinctData;
}

// PackageReleaseProxySetLabelsManyReq describes the HTTP request body when set
// many labels to proxy release.
export interface PackageReleaseProxySetLabelsManyReq {
  generation: number;
  exact_include_conditions: PackageReleaseExactConditions;
  labels: string[];
}

// PackageReleaseProxySetLabelsManyResp describes the HTTP response body when
// set many labels to proxy release.
export interface PackageReleaseProxySetLabelsManyResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: PackageReleaseProxySetLabelsManyRespData;
}

export interface PackageReleaseProxySetLabelsManyRespData {
}

// PackageReleaseProxyEnableReq describes the HTTP request body when enable
// proxy release.
export interface PackageReleaseProxyEnableReq {
  generation: number;
  release_type: string;
  platform: Platform;
  version: string;
}

// PackageReleaseProxyEnableResp describes the HTTP response body when enable
// proxy release.
export interface PackageReleaseProxyEnableResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: PackageReleaseProxyEnableRespData;
}

export interface PackageReleaseProxyEnableRespData {
}

// PackageReleaseProxyDisableReq describes the HTTP request body when disable
// proxy release.
export interface PackageReleaseProxyDisableReq {
  generation: number;
  release_type: string;
  platform: Platform;
  version: string;
}

// PackageReleaseProxyDisableResp describes the HTTP response body when disable
// proxy release.
export interface PackageReleaseProxyDisableResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: PackageReleaseProxyDisableRespData;
}

export interface PackageReleaseProxyDisableRespData {
}

export interface PackageReleaseProxyVisibleReq {
  generation: number;
  release_type: string;
  platform: Platform;
  version: string;
}

export interface PackageReleaseProxyVisibleResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: PackageReleaseProxyVisibleRespData;
}

export interface PackageReleaseProxyVisibleRespData {
}

export interface PackageReleaseProxyUnvisibleReq {
  generation: number;
  release_type: string;
  platform: Platform;
  version: string;
}

export interface PackageReleaseProxyUnvisibleResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: PackageReleaseProxyUnvisibleRespData;
}

export interface PackageReleaseProxyUnvisibleRespData {
}

// PackageReleaseProxySetAsDefaultReq describes the HTTP request body when set
// default proxy release.
export interface PackageReleaseProxySetAsDefaultReq {
  generation: number;
  release_type: string;
  platform: Platform;
  version: string;
}

// PackageReleaseProxySetAsDefaultResp describes the HTTP response body when set
// default proxy release.
export interface PackageReleaseProxySetAsDefaultResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: PackageReleaseProxySetAsDefaultRespData;
}

export interface PackageReleaseProxySetAsDefaultRespData {
}

// PackageReleaseProxyCancelAsDefaultReq describes the HTTP request body when
// cancel default proxy release.
export interface PackageReleaseProxyCancelAsDefaultReq {
  generation: number;
  release_type: string;
  platform: Platform;
  version: string;
}

// PackageReleaseProxyCancelAsDefaultResp describes the HTTP response body when
// cancel default proxy release.
export interface PackageReleaseProxyCancelAsDefaultResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: PackageReleaseProxyCancelAsDefaultRespData;
}

export interface PackageReleaseProxyCancelAsDefaultRespData {
}

// PackageReleaseProxyDeleteReq describes the HTTP request body when delete
// proxy release.
export interface PackageReleaseProxyDeleteReq {
  generation: number;
  release_type: string;
  platform: Platform;
  version: string;
}

// PackageReleaseProxyDeleteResp describes the HTTP response body when delete
// proxy release.
export interface PackageReleaseProxyDeleteResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: PackageReleaseProxyDeleteRespData;
}

export interface PackageReleaseProxyDeleteRespData {
}

// PackageReleaseProxyCountDeployedReq describes the HTTP request body when
// count agent deployed.
export interface PackageReleaseProxyCountDeployedReq {
  items: Item[];
}

export interface PackageReleaseProxyCountDeployedReqItem {
  generation: number;
  platform: Platform;
  version: string;
}

// PackageReleaseProxyCountDeployedResp describes the HTTP response body when
// count agent deployed.
export interface PackageReleaseProxyCountDeployedResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: PackageReleaseProxyCountDeployedRespData;
}

export interface PackageReleaseProxyCountDeployedRespData {
  counts: number[];
}

// PackageReleaseProxyDownloadReq is the request for download proxy pkg.
export interface PackageReleaseProxyDownloadReq {
  generation: number;
  platform: Platform;
  version: string;
}

// PackageReleasePluginListReq describes the HTTP request body when list plugin
// release.
export interface PackageReleasePluginListReq {
  page: Page;
  generation: number;
  only_count: boolean;
  exact_include_conditions: PackageReleaseExactConditions;
}

// PackageReleasePluginListResp describes the HTTP response body when list
// plugin release.
export interface PackageReleasePluginListResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: PackageReleasePluginListRespData;
}

export interface PackageReleasePluginListRespData {
  total: number;
  items: ReleasePlugin[];
}

// PackageReleasePluginListBriefReq describes the HTTP request body when list
// plugin release brief.
export interface PackageReleasePluginListBriefReq {
  page: Page;
  generation: number;
  only_count: boolean;
  exact_include_conditions: PackageReleaseExactConditions;
}

// PackageReleasePluginListBriefResp describes the HTTP response body when list
// plugin release brief.
export interface PackageReleasePluginListBriefResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: PackageReleasePluginListBriefRespData;
}

export interface PackageReleasePluginListBriefRespData {
  total: number;
  items: ReleasePluginBrief[];
}

// PackageReleasePluginEnableReq describes the HTTP request body when enable
// plugin release.
export interface PackageReleasePluginEnableReq {
  generation: number;
  name: string;
  platform: Platform;
  version: string;
}

// PackageReleasePluginEnableResp describes the HTTP response body when enable
// plugin release.
export interface PackageReleasePluginEnableResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: PackageReleasePluginEnableRespData;
}

export interface PackageReleasePluginEnableRespData {
}

// PackageReleasePluginDisableReq describes the HTTP request body when disable
// plugin release.
export interface PackageReleasePluginDisableReq {
  generation: number;
  name: string;
  platform: Platform;
  version: string;
}

// PackageReleasePluginDisableResp describes the HTTP response body when disable
// plugin release.
export interface PackageReleasePluginDisableResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: PackageReleasePluginDisableRespData;
}

export interface PackageReleasePluginDisableRespData {
}

export interface PackageReleasePluginVisibleReq {
  generation: number;
  name: string;
  platform: Platform;
  version: string;
}

export interface PackageReleasePluginVisibleResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: PackageReleasePluginVisibleRespData;
}

export interface PackageReleasePluginVisibleRespData {
}

export interface PackageReleasePluginUnvisibleReq {
  generation: number;
  name: string;
  platform: Platform;
  version: string;
}

export interface PackageReleasePluginUnvisibleResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: PackageReleasePluginUnvisibleRespData;
}

export interface PackageReleasePluginUnvisibleRespData {
}

// PackageReleasePluginSetAsDefaultReq describes the HTTP request body when set
// default plugin release.
export interface PackageReleasePluginSetAsDefaultReq {
  generation: number;
  name: string;
  platform: Platform;
  version: string;
}

// PackageReleasePluginSetAsDefaultResp describes the HTTP response body when
// set default plugin release.
export interface PackageReleasePluginSetAsDefaultResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: PackageReleasePluginSetAsDefaultRespData;
}

export interface PackageReleasePluginSetAsDefaultRespData {
}

// PackageReleasePluginCancelAsDefaultReq describes the HTTP request body when
// cancel default plugin release.
export interface PackageReleasePluginCancelAsDefaultReq {
  generation: number;
  name: string;
  platform: Platform;
  version: string;
}

// PackageReleasePluginCancelAsDefaultResp describes the HTTP response body when
// cancel default plugin release.
export interface PackageReleasePluginCancelAsDefaultResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: PackageReleasePluginCancelAsDefaultRespData;
}

export interface PackageReleasePluginCancelAsDefaultRespData {
}

// PackageReleasePluginDeleteReq describes the HTTP request body when delete
// plugin release.
export interface PackageReleasePluginDeleteReq {
  generation: number;
  name: string;
  platform: Platform;
  version: string;
}

// PackageReleasePluginDeleteResp describes the HTTP response body when delete
// plugin release.
export interface PackageReleasePluginDeleteResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: PackageReleasePluginDeleteRespData;
}

export interface PackageReleasePluginDeleteRespData {
}

// PackageReleasePluginDownloadReq is the request for download plugin pkg.
export interface PackageReleasePluginDownloadReq {
  name: string;
  platform: Platform;
  version: string;
}

// PackageReleasePluginGetConfigVariablesReq describes the HTTP request body
// when get config variables of plugin release.
export interface PackageReleasePluginGetConfigVariablesReq {
  generation: number;
  name: string;
  platforms: Platform[];
  version: string;
}

// ConfigVariables describes the config variables of plugin release.
export interface ConfigVariables {
  name: string;
  file_path: string;
  source_path: string;
  is_main_config: boolean;
  source_content: string;
  variables: Record<string, Property>;
}

export interface ConfigVariablesProperty {
  title: string;
  type: string;
  required: boolean;
  default: google.protobuf.Value;
  description: string;
  description_en: string;
  properties: Record<string, Property>;
}

// PackageReleasePluginGetConfigVariablesResp describes the HTTP response body
// when get config variables of plugin release.
export interface PackageReleasePluginGetConfigVariablesResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: PackageReleasePluginGetConfigVariablesRespData;
}

export interface PackageReleasePluginGetConfigVariablesRespConfigVariablesList {
  items: ConfigVariables[];
}

export interface PackageReleasePluginGetConfigVariablesRespData {
  config_variables: Record<string, ConfigVariablesList>;
}

// PackageReleaseCertListReq describes the HTTP request body when list cert
// release.
export interface PackageReleaseCertListReq {
  generation: number;
}

// PackageReleaseCertListResp describes the HTTP response body when list cert
// release.
export interface PackageReleaseCertListResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: PackageReleaseCertListRespData;
}

export interface PackageReleaseCertListRespData {
  total: number;
  items: ReleaseCert[];
}

// PackageReleaseCertDeleteReq describes the HTTP request body when delete
// cert release.
export interface PackageReleaseCertDeleteReq {
  generation: number;
}

// PackageReleaseCertDeleteResp describes the HTTP response body when delete
// cert release.
export interface PackageReleaseCertDeleteResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: PackageReleaseCertDeleteRespData;
}

export interface PackageReleaseCertDeleteRespData {
}

// PackageReleaseCertDownloadReq is the request for download cert pkg.
export interface PackageReleaseCertDownloadReq {
  generation: number;
}

// PackageReleaseBinToolListReq describes the HTTP request body when list
// bintool release.
export interface PackageReleaseBinToolListReq {
  generation: number;
}

// PackageReleaseBinToolListResp describes the HTTP response body when list
// bintool release.
export interface PackageReleaseBinToolListResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: PackageReleaseBinToolListRespData;
}

export interface PackageReleaseBinToolListRespData {
  total: number;
  items: ReleaseBinTool[];
}

// PackageReleaseBinToolDeleteReq describes the HTTP request body when delete
// bintool release.
export interface PackageReleaseBinToolDeleteReq {
  generation: number;
}

// PackageReleaseBinToolDeleteResp describes the HTTP response body when delete
// bintool release.
export interface PackageReleaseBinToolDeleteResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: PackageReleaseBinToolDeleteRespData;
}

export interface PackageReleaseBinToolDeleteRespData {
}

// PackageReleaseBinToolDownloadReq is the request for download bintool pkg.
export interface PackageReleaseBinToolDownloadReq {
  generation: number;
}

// PackageReleasePluginBinToolListReq describes the HTTP request body when list
// plugin-bintool release.
export interface PackageReleasePluginBinToolListReq {
  generation: number;
}

// PackageReleasePluginBinToolListResp describes the HTTP response body when
// list plugin-bintool release.
export interface PackageReleasePluginBinToolListResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: PackageReleasePluginBinToolListRespData;
}

export interface PackageReleasePluginBinToolListRespData {
  total: number;
  items: ReleasePluginBinTool[];
}

// PackageReleasePluginBinToolDeleteReq describes the HTTP request body when
// delete plugin-bintool release.
export interface PackageReleasePluginBinToolDeleteReq {
  generation: number;
  name: string;
}

// PackageReleasePluginBinToolDeleteResp describes the HTTP response body when
// delete plugin-bintool release.
export interface PackageReleasePluginBinToolDeleteResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: PackageReleasePluginBinToolDeleteRespData;
}

export interface PackageReleasePluginBinToolDeleteRespData {
}

// PackageReleasePluginBinToolDownloadReq is the request for download plugin
// bintool pkg.
export interface PackageReleasePluginBinToolDownloadReq {
  generation: number;
  name: string;
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
  permission: Permission;
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
  permission: Permission;
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

