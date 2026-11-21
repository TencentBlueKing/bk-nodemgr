// gen-api.js 自动生成，请勿手动修改
// ExcelParseAgentInstallTemplateReq 通过 multipart 表单提交待解析的 Agent 安装
// Excel 文件。 file 为 Excel 文件，filename 须与文件名一致，metadata 可传 {}。
export interface ExcelParseAgentInstallTemplateReq {
}

// ExcelDownloadAgentInstallTemplateReq 下载 Agent 安装 Excel 模板的请求。
export interface ExcelDownloadAgentInstallTemplateReq {
}

// ExcelAgentInstallTemplateParsedInfo 描述 Agent 安装模板中一行数据的解析结果。
// 解析失败的行仅返回行号与错误原因，其余字段为空。
export interface ExcelAgentInstallTemplateParsedInfo {
  // Excel 中的行号，从 4 开始（前三行为表头、必填标注与填写说明）。
  row_index: number;
  // 解析失败原因；为空表示该行解析成功。
  error_message: string;
  bk_host_innerip: string;
  bk_host_innerip_v6: string;
  os_type: string;
  login_ip: string;
  login_port: number;
  login_user: string;
  login_mode: string;
  // 登录凭证，按登录方式填写密码、私钥内容或留空。
  credit: string;
  // 记录到 CMDB 的寻址方式，static 或 dynamic。
  bk_addressing: string;
  // 为空时使用页面上的安装预设插件选项。
  install_pre_ordered_plugins: boolean;
  // 为空时使用页面上的重新注册 AgentID 选项。
  re_register: boolean;
  // 登录用户为非 root 且已配置免密 sudo 时置为 true，安装命令将通过 sudo -n
  // 提权执行；默认为 false。
  is_sudo: boolean;
  // 安装方式，支持 ssh、wmi；不填使用系统默认（自动）。
  install_method: string;
}

// ExcelParseAgentInstallTemplateResp 逐行返回 Agent 安装 Excel 模板的解析结果。
export interface ExcelParseAgentInstallTemplateResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: ExcelParseAgentInstallTemplateRespData;
}

export interface ExcelParseAgentInstallTemplateRespData {
  items: ExcelAgentInstallTemplateParsedInfo[];
}

// ExcelParseProxyInstallTemplateReq 通过 multipart 表单提交待解析的 Proxy 安装
// Excel 文件。 file 为 Excel 文件，filename 须与文件名一致，metadata 可传 {}。
export interface ExcelParseProxyInstallTemplateReq {
}

// ExcelDownloadProxyInstallTemplateReq 下载 Proxy 安装 Excel 模板的请求。
export interface ExcelDownloadProxyInstallTemplateReq {
}

// ExcelProxyInstallTemplateParsedInfo 描述 Proxy 安装模板中一行数据的解析结果。
// 解析失败的行仅返回行号与错误原因，其余字段为空。
export interface ExcelProxyInstallTemplateParsedInfo {
  // Excel 中的行号，从 4 开始（前三行为表头、必填标注与填写说明）。
  row_index: number;
  // 解析失败原因；为空表示该行解析成功。
  error_message: string;
  bk_host_innerip: string;
  bk_host_innerip_v6: string;
  export_ip: string;
  bk_addressing: string;
  os_type: string;
  login_ip: string;
  login_port: number;
  login_user: string;
  login_mode: string;
  // 登录凭证，按登录方式填写密码、私钥内容或留空。
  credit: string;
  cpu_arch: string;
  advertise_ip: string;
  // 为空时使用页面上的安装预设插件选项。
  install_pre_ordered_plugins: boolean;
  // 为空时使用页面上的重新注册 AgentID 选项。
  re_register: boolean;
  // 为空时按页面默认值处理。
  dedicated_installer: boolean;
  cluster_tunnel: boolean;
  file_tunnel: boolean;
  data_tunnel: boolean;
  // 安装方式，支持 ssh、wmi；不填使用系统默认（自动）。
  install_method: string;
  // 登录用户为非 root 且已配置免密 sudo 时置为 true，安装命令将通过 sudo -n
  // 提权执行；默认为 false。
  is_sudo: boolean;
}

// ExcelParseProxyInstallTemplateResp 逐行返回 Proxy 安装 Excel 模板的解析结果。
export interface ExcelParseProxyInstallTemplateResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: ExcelParseProxyInstallTemplateRespData;
}

export interface ExcelParseProxyInstallTemplateRespData {
  items: ExcelProxyInstallTemplateParsedInfo[];
}

// ExcelParseOpsTemplateReq 通过 multipart 表单提交待解析的带外设置 Excel 文件。
// file 为 Excel 文件，filename 须与文件名一致，metadata 可传 {}。
export interface ExcelParseOpsTemplateReq {
}

// ExcelDownloadOpsTemplateReq 下载带外设置 Excel 模板的请求。
export interface ExcelDownloadOpsTemplateReq {
}

// ExcelOpsTemplateParsedInfo 描述带外设置模板中一行数据的解析结果。
// 解析成功的行返回匹配到的主机与主控机 ID，可直接用于回填带外设置表格；
// 解析失败的行仅返回行号与错误原因，其余字段为空。
export interface ExcelOpsTemplateParsedInfo {
  // Excel 中的行号，从 4 开始（前三行为表头、必填标注与填写说明）。
  row_index: number;
  // 解析失败原因；为空表示该行解析成功。
  error_message: string;
  // 按业务、管控区域与内网 IP 匹配到的目标主机 ID。
  bk_host_id: number;
  bk_biz_id: number;
  bk_biz_name: string;
  bk_networkarea_id: number;
  bk_networkarea_name: string;
  bk_host_innerip: string;
  bk_host_innerip_v6: string;
  // 按主控机管控区域与主控机 IP 匹配到的主控机 ID；未填写主控机时为 0。
  ops_console_host_id: number;
  // 主控机回显 IP，取匹配到的主控机内网 IPv4，无 IPv4 时取 IPv6。
  ops_console_host_ip: string;
  ops_out_band_type: string;
  ops_out_band_protocol: string;
  ops_out_band_ip: string;
  ops_out_band_port: number;
}

// ExcelParseOpsTemplateResp 逐行返回带外设置 Excel 模板的解析结果。
export interface ExcelParseOpsTemplateResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: ExcelParseOpsTemplateRespData;
}

export interface ExcelParseOpsTemplateRespData {
  items: ExcelOpsTemplateParsedInfo[];
}

