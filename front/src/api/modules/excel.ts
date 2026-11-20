// gen-api.js 自动生成，请勿手动修改
import { type Config } from '../interceptors';

import type { ExcelDownloadAgentInstallTemplateReq, ExcelDownloadOpsTemplateReq, ExcelDownloadProxyInstallTemplateReq, ExcelUploadAgentInstallTemplateReq, ExcelUploadAgentInstallTemplateResp, ExcelUploadOpsTemplateReq, ExcelUploadOpsTemplateResp, ExcelUploadProxyInstallTemplateReq, ExcelUploadProxyInstallTemplateResp } from '@/@types/excel';
import Fetch from '@/api/fetch';

const fetch = new Fetch({
  prefix: `${import.meta.env.BK_API_PREFIX}`,
});

export const ExcelService = {
  // 解析 Agent 安装 Excel 模板，逐行返回安装参数。
  //
  // 模板由下载接口生成，需保持表头与必填标注一致；单行解析失败会返回该行的
  // 错误原因，不影响其他行。
  ExcelUploadAgentInstallTemplate: async <Request = ExcelUploadAgentInstallTemplateReq, ResponseData = ExcelUploadAgentInstallTemplateResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/excel/agent_install_template/upload')(params, config),
  // 下载 Agent 安装 Excel 模板。
  ExcelDownloadAgentInstallTemplate: async <Request = ExcelDownloadAgentInstallTemplateReq, ResponseData = FileChunk>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/excel/agent_install_template/download')(params, config),
  // 解析 Proxy 安装 Excel 模板，逐行返回安装参数。
  //
  // 模板由下载接口生成，需保持表头与必填标注一致；单行解析失败会返回该行的
  // 错误原因，不影响其他行。
  ExcelUploadProxyInstallTemplate: async <Request = ExcelUploadProxyInstallTemplateReq, ResponseData = ExcelUploadProxyInstallTemplateResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/excel/proxy_install_template/upload')(params, config),
  // 下载 Proxy 安装 Excel 模板。
  ExcelDownloadProxyInstallTemplate: async <Request = ExcelDownloadProxyInstallTemplateReq, ResponseData = FileChunk>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/excel/proxy_install_template/download')(params, config),
  // 解析带外设置 Excel 模板，按业务、管控区域与内网 IP
  // 匹配已有主机，并解析主控机与带外字段。
  //
  // 逐行返回解析结果：成功的行携带主机与主控机 ID，失败的行携带行号与错误
  // 原因，单行失败不影响其他行。
  ExcelUploadOpsTemplate: async <Request = ExcelUploadOpsTemplateReq, ResponseData = ExcelUploadOpsTemplateResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/excel/ops_template/upload')(params, config),
  // 下载带外设置 Excel 模板。
  ExcelDownloadOpsTemplate: async <Request = ExcelDownloadOpsTemplateReq, ResponseData = FileChunk>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/excel/ops_template/download')(params, config),
};

