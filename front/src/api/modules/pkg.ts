// gen-api.js 自动生成，请勿手动修改
import { type Config } from '../interceptors';

import type { PackageUploadOriginAgentReq, PackageUploadOriginAgentResp, PackageUploadOriginServerReq, PackageUploadOriginServerResp, PackageUploadOriginCertReq, PackageUploadOriginCertResp, PackageUploadOriginBinToolReq, PackageUploadOriginBinToolResp, PackageUploadOriginPluginBinToolReq, PackageUploadOriginPluginBinToolResp, PackagePublishReleaseAgentReq, PackagePublishReleaseAgentResp, PackagePublishReleaseProxyReq, PackagePublishReleaseProxyResp, PackagePublishReleaseCertReq, PackagePublishReleaseCertResp, PackagePublishReleaseBinToolReq, PackagePublishReleaseBinToolResp, PackagePublishReleasePluginBinToolReq, PackagePublishReleasePluginBinToolResp, PackageEventListReq, PackageEventListResp, PackageEventDistinctReq, PackageEventDistinctResp, PackageUploadOriginPluginV2Req, PackageUploadOriginPluginV2Resp, PackageUploadOriginExternalPluginV2Req, PackageUploadOriginExternalPluginV2Resp, PackageUploadOriginPluginV3Req, PackageUploadOriginPluginV3Resp, PackagePublishReleasePluginV2Req, PackagePublishReleasePluginV2Resp, PackagePublishReleaseExternalPluginV2Req, PackagePublishReleaseExternalPluginV2Resp, PackagePublishReleasePluginV3Req, PackagePublishReleasePluginV3Resp, PackageReleaseAgentListReq, PackageReleaseAgentListResp, PackageReleaseAgentDistinctReq, PackageReleaseAgentDistinctResp, PackageReleaseAgentSetLabelsManyReq, PackageReleaseAgentSetLabelsManyResp, PackageReleaseAgentEnableReq, PackageReleaseAgentEnableResp, PackageReleaseAgentDisableReq, PackageReleaseAgentDisableResp, PackageReleaseAgentSetAsDefaultReq, PackageReleaseAgentSetAsDefaultResp, PackageReleaseAgentCancelAsDefaultReq, PackageReleaseAgentCancelAsDefaultResp, PackageReleaseAgentDeleteReq, PackageReleaseAgentDeleteResp, PackageReleaseAgentCountDeployedReq, PackageReleaseAgentCountDeployedResp, PackageReleaseAgentDownloadReq, FileChunk, PackageReleaseProxyListReq, PackageReleaseProxyListResp, PackageReleaseProxyDistinctReq, PackageReleaseProxyDistinctResp, PackageReleaseProxySetLabelsManyReq, PackageReleaseProxySetLabelsManyResp, PackageReleaseProxyEnableReq, PackageReleaseProxyEnableResp, PackageReleaseProxyDisableReq, PackageReleaseProxyDisableResp, PackageReleaseProxySetAsDefaultReq, PackageReleaseProxySetAsDefaultResp, PackageReleaseProxyCancelAsDefaultReq, PackageReleaseProxyCancelAsDefaultResp, PackageReleaseProxyDeleteReq, PackageReleaseProxyDeleteResp, PackageReleaseProxyCountDeployedReq, PackageReleaseProxyCountDeployedResp, PackageReleaseProxyDownloadReq, PackageReleasePluginListReq, PackageReleasePluginListResp, PackageReleasePluginEnableReq, PackageReleasePluginEnableResp, PackageReleasePluginDisableReq, PackageReleasePluginDisableResp, PackageReleasePluginSetAsDefaultReq, PackageReleasePluginSetAsDefaultResp, PackageReleasePluginCancelAsDefaultReq, PackageReleasePluginCancelAsDefaultResp, PackageReleasePluginDeleteReq, PackageReleasePluginDeleteResp, PackageReleasePluginDownloadReq, PackageReleaseCertListReq, PackageReleaseCertListResp, PackageReleaseCertDeleteReq, PackageReleaseCertDeleteResp, PackageReleaseCertDownloadReq, PackageReleaseBinToolListReq, PackageReleaseBinToolListResp, PackageReleaseBinToolDeleteReq, PackageReleaseBinToolDeleteResp, PackageReleaseBinToolDownloadReq, PackageReleasePluginBinToolListReq, PackageReleasePluginBinToolListResp, PackageReleasePluginBinToolDeleteReq, PackageReleasePluginBinToolDeleteResp, PackageReleasePluginBinToolDownloadReq } from '@/@types/pkg';
import Fetch from '@/api/fetch';

const fetch = new Fetch({
  prefix: `${import.meta.env.BK_API_PREFIX}`,
});

export const PackageService = {
  // UploadOriginAgent provides origin agent upload.
  UploadOriginAgent: async <Request = PackageUploadOriginAgentReq, ResponseData = PackageUploadOriginAgentResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/package/upload/origin/agent')(params, config),
  // UploadOriginServer provides origin server upload.
  UploadOriginServer: async <Request = PackageUploadOriginServerReq, ResponseData = PackageUploadOriginServerResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/package/upload/origin/server')(params, config),
  // UploadOriginCert provides origin cert upload.
  UploadOriginCert: async <Request = PackageUploadOriginCertReq, ResponseData = PackageUploadOriginCertResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/package/upload/origin/cert')(params, config),
  // UploadOriginBinTool provides origin bintool upload.
  UploadOriginBinTool: async <Request = PackageUploadOriginBinToolReq, ResponseData = PackageUploadOriginBinToolResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/package/upload/origin/bintool')(params, config),
  // UploadOriginPluginBinTool provides origin plugin bintool upload.
  UploadOriginPluginBinTool: async <Request = PackageUploadOriginPluginBinToolReq, ResponseData = PackageUploadOriginPluginBinToolResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/package/upload/origin/plugin_bintool')(params, config),
  // PublishReleaseAgent provides release agent publish.
  PublishReleaseAgent: async <Request = PackagePublishReleaseAgentReq, ResponseData = PackagePublishReleaseAgentResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/package/publish/release/agent')(params, config),
  // PublishReleaseServer provides release server publish.
  PublishReleaseProxy: async <Request = PackagePublishReleaseProxyReq, ResponseData = PackagePublishReleaseProxyResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/package/publish/release/proxy')(params, config),
  // PublishReleaseCert provides release cert publish.
  PublishReleaseCert: async <Request = PackagePublishReleaseCertReq, ResponseData = PackagePublishReleaseCertResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/package/publish/release/cert')(params, config),
  // PublishReleaseBinTool provides release bintool publish.
  PublishReleaseBinTool: async <Request = PackagePublishReleaseBinToolReq, ResponseData = PackagePublishReleaseBinToolResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/package/publish/release/bintool')(params, config),
  // PublishReleaseBinTool provides release bintool publish.
  PublishReleasePluginBinTool: async <Request = PackagePublishReleasePluginBinToolReq, ResponseData = PackagePublishReleasePluginBinToolResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/package/publish/release/plugin_bintool')(params, config),
  // PackageEventList provides package event listing.
  PackageEventList: async <Request = PackageEventListReq, ResponseData = PackageEventListResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/package/event/list')(params, config),
  // PackageEventDistinct provides package event distincting.
  PackageEventDistinct: async <Request = PackageEventDistinctReq, ResponseData = PackageEventDistinctResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/package/event/distinct')(params, config),
  // UploadOriginPluginV2 provides origin plugin v2 upload.
  UploadOriginPluginV2: async <Request = PackageUploadOriginPluginV2Req, ResponseData = PackageUploadOriginPluginV2Resp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/package/upload/origin/v2/plugin')(params, config),
  // UploadOriginExternalPluginV2 provides origin external plugin v2 upload.
  UploadOriginExternalPluginV2: async <Request = PackageUploadOriginExternalPluginV2Req, ResponseData = PackageUploadOriginExternalPluginV2Resp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/package/upload/origin/v2/external_plugin')(params, config),
  // UploadOriginPluginV3 provides origin plugin v3 upload.
  UploadOriginPluginV3: async <Request = PackageUploadOriginPluginV3Req, ResponseData = PackageUploadOriginPluginV3Resp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/package/upload/origin/v3/plugin')(params, config),
  // PublishReleasePluginV2 provides release plugin v2 publish.
  PublishReleasePluginV2: async <Request = PackagePublishReleasePluginV2Req, ResponseData = PackagePublishReleasePluginV2Resp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/package/publish/release/v2/plugin')(params, config),
  // PublishReleaseExternalPluginV2 provides release external plugin v2 publish.
  PublishReleaseExternalPluginV2: async <Request = PackagePublishReleaseExternalPluginV2Req, ResponseData = PackagePublishReleaseExternalPluginV2Resp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/package/publish/release/v2/external_plugin')(params, config),
  // PublishReleasePluginV3 provides release plugin v3 publish.
  PublishReleasePluginV3: async <Request = PackagePublishReleasePluginV3Req, ResponseData = PackagePublishReleasePluginV3Resp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/package/publish/release/v3/plugin')(params, config),
  // ListReleaseAgent lists agent releases.
  ListReleaseAgent: async <Request = PackageReleaseAgentListReq, ResponseData = PackageReleaseAgentListResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/package/release/agent/list')(params, config),
  // DistinctReleaseAgent distincts agent releases.
  DistinctReleaseAgent: async <Request = PackageReleaseAgentDistinctReq, ResponseData = PackageReleaseAgentDistinctResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/package/release/agent/distinct')(params, config),
  // SetReleaseAgentLabelsMany sets many agent releases labels.
  SetReleaseAgentLabelsMany: async <Request = PackageReleaseAgentSetLabelsManyReq, ResponseData = PackageReleaseAgentSetLabelsManyResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/package/release/agent/set_labels_many')(params, config),
  // EnableReleaseAgent enables agent release.
  EnableReleaseAgent: async <Request = PackageReleaseAgentEnableReq, ResponseData = PackageReleaseAgentEnableResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/package/release/agent/enable')(params, config),
  // DisableReleaseAgent disables agent release.
  DisableReleaseAgent: async <Request = PackageReleaseAgentDisableReq, ResponseData = PackageReleaseAgentDisableResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/package/release/agent/disable')(params, config),
  // SetAsDefaultReleaseAgent sets agent release as default.
  SetAsDefaultReleaseAgent: async <Request = PackageReleaseAgentSetAsDefaultReq, ResponseData = PackageReleaseAgentSetAsDefaultResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/package/release/agent/set_as_default')(params, config),
  // CancelAsDefaultReleaseAgent cancels agent release as default.
  CancelAsDefaultReleaseAgent: async <Request = PackageReleaseAgentCancelAsDefaultReq, ResponseData = PackageReleaseAgentCancelAsDefaultResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/package/release/agent/cancel_as_default')(params, config),
  // DeleteReleaseAgent deletes agent release.
  DeleteReleaseAgent: async <Request = PackageReleaseAgentDeleteReq, ResponseData = PackageReleaseAgentDeleteResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/package/release/agent/delete')(params, config),
  // CountDeployedReleasedAgent count deployed for release agent.
  CountDeployedReleasedAgent: async <Request = PackageReleaseAgentCountDeployedReq, ResponseData = PackageReleaseAgentCountDeployedResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/package/release/agent/count_deployed')(params, config),
  // DownloadReleaseAgent download release agent package.
  DownloadReleaseAgent: async <Request = PackageReleaseAgentDownloadReq, ResponseData = FileChunk['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/package/release/agent/download')(params, config),
  // ListReleaseProxy lists proxy releases.
  ListReleaseProxy: async <Request = PackageReleaseProxyListReq, ResponseData = PackageReleaseProxyListResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/package/release/proxy/list')(params, config),
  // DistinctReleaseProxy distincts proxy releases.
  DistinctReleaseProxy: async <Request = PackageReleaseProxyDistinctReq, ResponseData = PackageReleaseProxyDistinctResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/package/release/proxy/distinct')(params, config),
  // SetReleaseProxyLabelsMany sets many proxy releases labels.
  SetReleaseProxyLabelsMany: async <Request = PackageReleaseProxySetLabelsManyReq, ResponseData = PackageReleaseProxySetLabelsManyResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/package/release/proxy/set_labels_many')(params, config),
  // EnableReleaseProxy enables proxy release.
  EnableReleaseProxy: async <Request = PackageReleaseProxyEnableReq, ResponseData = PackageReleaseProxyEnableResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/package/release/proxy/enable')(params, config),
  // DisableReleaseProxy disables proxy release.
  DisableReleaseProxy: async <Request = PackageReleaseProxyDisableReq, ResponseData = PackageReleaseProxyDisableResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/package/release/proxy/disable')(params, config),
  // SetAsDefaultReleaseProxy sets proxy release as default.
  SetAsDefaultReleaseProxy: async <Request = PackageReleaseProxySetAsDefaultReq, ResponseData = PackageReleaseProxySetAsDefaultResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/package/release/proxy/set_as_default')(params, config),
  // CancelAsDefaultReleaseProxy cancels proxy release as default.
  CancelAsDefaultReleaseProxy: async <Request = PackageReleaseProxyCancelAsDefaultReq, ResponseData = PackageReleaseProxyCancelAsDefaultResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/package/release/proxy/cancel_as_default')(params, config),
  // DeleteReleaseProxy deletes proxy release.
  DeleteReleaseProxy: async <Request = PackageReleaseProxyDeleteReq, ResponseData = PackageReleaseProxyDeleteResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/package/release/proxy/delete')(params, config),
  // CountDeployedReleasedProxy count deployed for release proxy.
  CountDeployedReleasedProxy: async <Request = PackageReleaseProxyCountDeployedReq, ResponseData = PackageReleaseProxyCountDeployedResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/package/release/proxy/count_deployed')(params, config),
  // DownloadReleaseProxy download release proxy package.
  DownloadReleaseProxy: async <Request = PackageReleaseProxyDownloadReq, ResponseData = FileChunk['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/package/release/proxy/download')(params, config),
  // ListReleasePlugin lists plugin releases.
  ListReleasePlugin: async <Request = PackageReleasePluginListReq, ResponseData = PackageReleasePluginListResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/package/release/plugin/list')(params, config),
  // EnableReleasePlugin enables plugin release.
  EnableReleasePlugin: async <Request = PackageReleasePluginEnableReq, ResponseData = PackageReleasePluginEnableResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/package/release/plugin/enable')(params, config),
  // DisableReleasePlugin disables plugin release.
  DisableReleasePlugin: async <Request = PackageReleasePluginDisableReq, ResponseData = PackageReleasePluginDisableResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/package/release/plugin/disable')(params, config),
  // SetAsDefaultReleasePlugin sets plugin release as default.
  SetAsDefaultReleasePlugin: async <Request = PackageReleasePluginSetAsDefaultReq, ResponseData = PackageReleasePluginSetAsDefaultResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/package/release/plugin/set_as_default')(params, config),
  // CancelAsDefaultReleasePlugin cancels plugin release as default.
  CancelAsDefaultReleasePlugin: async <Request = PackageReleasePluginCancelAsDefaultReq, ResponseData = PackageReleasePluginCancelAsDefaultResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/package/release/plugin/cancel_as_default')(params, config),
  // DeleteReleasePlugin deletes plugin release.
  DeleteReleasePlugin: async <Request = PackageReleasePluginDeleteReq, ResponseData = PackageReleasePluginDeleteResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/package/release/plugin/delete')(params, config),
  // DownloadReleasePlugin download release plugin package.
  DownloadReleasePlugin: async <Request = PackageReleasePluginDownloadReq, ResponseData = FileChunk['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/package/release/plugin/download')(params, config),
  // ListReleaseCert lists cert releases.
  ListReleaseCert: async <Request = PackageReleaseCertListReq, ResponseData = PackageReleaseCertListResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/package/release/cert/list')(params, config),
  // DeleteReleaseCert deletes cert release.
  DeleteReleaseCert: async <Request = PackageReleaseCertDeleteReq, ResponseData = PackageReleaseCertDeleteResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/package/release/cert/delete')(params, config),
  // DownloadReleaseCert download release cert package.
  DownloadReleaseCert: async <Request = PackageReleaseCertDownloadReq, ResponseData = FileChunk['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/package/release/cert/download')(params, config),
  // ListReleaseBinTool lists bintool releases.
  ListReleaseBinTool: async <Request = PackageReleaseBinToolListReq, ResponseData = PackageReleaseBinToolListResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/package/release/bintool/list')(params, config),
  // DeleteReleaseBinTool deletes bintool release.
  DeleteReleaseBinTool: async <Request = PackageReleaseBinToolDeleteReq, ResponseData = PackageReleaseBinToolDeleteResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/package/release/bintool/delete')(params, config),
  // DownloadReleaseBinTool download release bintool package.
  DownloadReleaseBinTool: async <Request = PackageReleaseBinToolDownloadReq, ResponseData = FileChunk['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/package/release/bintool/download')(params, config),
  // ListReleasePluginBinTool lists plugin-bintool releases.
  ListReleasePluginBinTool: async <Request = PackageReleasePluginBinToolListReq, ResponseData = PackageReleasePluginBinToolListResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/package/release/plugin_bintool/list')(params, config),
  // DeleteReleasePluginBinTool deletes plugin-bintool release.
  DeleteReleasePluginBinTool: async <Request = PackageReleasePluginBinToolDeleteReq, ResponseData = PackageReleasePluginBinToolDeleteResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/package/release/plugin_bintool/delete')(params, config),
  // DownloadReleasePluginBinTool download release plugin bintool package.
  DownloadReleasePluginBinTool: async <Request = PackageReleasePluginBinToolDownloadReq, ResponseData = FileChunk['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/package/release/plugin_bintool/download')(params, config),
};

