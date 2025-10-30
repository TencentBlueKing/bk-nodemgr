// gen-api.js 自动生成，请勿手动修改
import { type Config } from '../interceptors';

import type { PackageUploadOriginAgentReq, PackageUploadOriginAgentResp, PackageUploadOriginServerReq, PackageUploadOriginServerResp, PackageUploadOriginCertReq, PackageUploadOriginCertResp, PackageUploadOriginBinToolReq, PackageUploadOriginBinToolResp, PackagePublishReleaseAgentReq, PackagePublishReleaseAgentResp, PackagePublishReleaseProxyReq, PackagePublishReleaseProxyResp, PackagePublishReleaseCertReq, PackagePublishReleaseCertResp, PackagePublishReleaseBinToolReq, PackagePublishReleaseBinToolResp, PackageReleaseListReq, PackageReleaseListResp, PackageReleaseSetLabelsReq, PackageReleaseSetLabelsResp, PackageReleaseSetLabelsManyReq, PackageReleaseSetLabelsManyResp, PackageReleaseEnableReq, PackageReleaseEnableResp, PackageReleaseDisableReq, PackageReleaseDisableResp, PackageReleaseSetAsDefaultReq, PackageReleaseSetAsDefaultResp, PackageReleaseCancelAsDefaultReq, PackageReleaseCancelAsDefaultResp, PackageReleaseDeleteReq, PackageReleaseDeleteResp, PackageReleaseDeployedHostCountReq, PackageReleaseDeployedHostCountResp, PackageReleaseAgentDownloadReq, FileChunk, PackageReleaseProxyDownloadReq, PackageEventListReq, PackageEventListResp, PackageEventDistinctReq, PackageEventDistinctResp, PackageReleaseAgentListReq, PackageReleaseAgentListResp, PackageReleaseProxyListReq, PackageReleaseProxyListResp } from '@/@types/pkg';
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
  // PublishReleaseAgent provides release agent publish.
  PublishReleaseAgent: async <Request = PackagePublishReleaseAgentReq, ResponseData = PackagePublishReleaseAgentResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/package/publish/release/agent')(params, config),
  // PublishReleaseServer provides release server publish.
  PublishReleaseProxy: async <Request = PackagePublishReleaseProxyReq, ResponseData = PackagePublishReleaseProxyResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/package/publish/release/proxy')(params, config),
  // PublishReleaseCert provides release cert publish.
  PublishReleaseCert: async <Request = PackagePublishReleaseCertReq, ResponseData = PackagePublishReleaseCertResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/package/publish/release/cert')(params, config),
  // PublishReleaseBinTool provides release bintool publish.
  PublishReleaseBinTool: async <Request = PackagePublishReleaseBinToolReq, ResponseData = PackagePublishReleaseBinToolResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/package/publish/release/bintool')(params, config),
  // ListRelease lists releases.
  ListRelease: async <Request = PackageReleaseListReq, ResponseData = PackageReleaseListResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/package/release/list')(params, config),
  // SetReleaseLabels sets release labels.
  SetReleaseLabels: async <Request = PackageReleaseSetLabelsReq, ResponseData = PackageReleaseSetLabelsResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/package/release/set_labels')(params, config),
  // SetReleaseLabeslMany sets many release labels.
  SetReleaseLabelsMany: async <Request = PackageReleaseSetLabelsManyReq, ResponseData = PackageReleaseSetLabelsManyResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/package/release/set_labels_many')(params, config),
  // EnableRelease enables release.
  EnableRelease: async <Request = PackageReleaseEnableReq, ResponseData = PackageReleaseEnableResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/package/release/enable')(params, config),
  // DisableRelease disables release.
  DisableRelease: async <Request = PackageReleaseDisableReq, ResponseData = PackageReleaseDisableResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/package/release/disable')(params, config),
  // SetAsDefaultRelease sets release as default.
  SetAsDefaultRelease: async <Request = PackageReleaseSetAsDefaultReq, ResponseData = PackageReleaseSetAsDefaultResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/package/release/set_as_default')(params, config),
  // CancelAsDefaultRelease cancels release as default.
  CancelAsDefaultRelease: async <Request = PackageReleaseCancelAsDefaultReq, ResponseData = PackageReleaseCancelAsDefaultResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/package/release/cancel_as_default')(params, config),
  // DeleteRelease deletes release.
  DeleteRelease: async <Request = PackageReleaseDeleteReq, ResponseData = PackageReleaseDeleteResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/package/release/delete')(params, config),
  // DeployedHostCount count deployed host.
  DeployedHostCount: async <Request = PackageReleaseDeployedHostCountReq, ResponseData = PackageReleaseDeployedHostCountResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/package/release/deployed_host/count')(params, config),
  // DownloadAgent download release agent package.
  DownloadAgent: async <Request = PackageReleaseAgentDownloadReq, ResponseData = FileChunk['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/package/release/agent/download')(params, config),
  // DownloadProxy download release proxy package.
  DownloadProxy: async <Request = PackageReleaseProxyDownloadReq, ResponseData = FileChunk['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/package/release/proxy/download')(params, config),
  // PackageEventList provides package event listing.
  PackageEventList: async <Request = PackageEventListReq, ResponseData = PackageEventListResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/package/event/list')(params, config),
  // PackageEventDistinct provides package event distincting.
  PackageEventDistinct: async <Request = PackageEventDistinctReq, ResponseData = PackageEventDistinctResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/package/event/distinct')(params, config),
  // ListReleaseAgent lists releases.
  ListReleaseAgent: async <Request = PackageReleaseAgentListReq, ResponseData = PackageReleaseAgentListResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/package/release/agent/list')(params, config),
  // ListReleaseProxy lists releases.
  ListReleaseProxy: async <Request = PackageReleaseProxyListReq, ResponseData = PackageReleaseProxyListResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/package/release/proxy/list')(params, config),
};

