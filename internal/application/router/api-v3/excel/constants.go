/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 节点管理 (BlueKing - Node Management) available.
 * Copyright (C) Tencent. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.

 * We undertake not to change the open source license (MIT license) applicable

 * to the current version of the project delivered to anyone in the future.
 */

package excel

const (
	// Upload constraints.
	maxTemplateUploadFileSize = 10 * 1024 * 1024 // 10MB
)

const (
	// Agent install template.
	agentInstallSheetName = "agent_install_template"
	agentInstallFileName  = "bk_nodemgr_agent_install_template.xlsx"

	agentColDescOSTypeZh                   = "目标主机操作系统类型，可选项来自包管理中已启用的 Agent 包。"
	agentColDescOSTypeEn                   = "Operating system of the target host. Choices come from enabled Agent packages in Package Management."
	agentColDescInstallPreOrderedPluginsZh = "安装 Agent 后是否自动安装预设插件，true / false，不填使用页面默认值。"
	agentColDescInstallPreOrderedPluginsEn = "Whether to install preset plugins after installing the Agent. Enter true or false. Leave blank to use the page default." //nolint:lll // Keep the translation on one line.
)

const (
	// Proxy install template.
	proxyInstallSheetName = "proxy_install_template"
	proxyInstallFileName  = "bk_nodemgr_proxy_install_template.xlsx"

	proxyColExportIPZh           = "出口 IP"
	proxyColExportIPEn           = "Export IP"
	proxyColCPUArchZh            = "CPU架构"
	proxyColCPUArchEn            = "CPU Architecture"
	proxyColAdvertiseIPZh        = "服务IP"
	proxyColAdvertiseIPEn        = "Service IP"
	proxyColDedicatedInstallerZh = "安装跳板"
	proxyColDedicatedInstallerEn = "Install Jumper"
	proxyColClusterTunnelZh      = "Agent控制"
	proxyColClusterTunnelEn      = "Agent Control"
	proxyColFileTunnelZh         = "文件传输"
	proxyColFileTunnelEn         = "File Transfer"
	proxyColDataTunnelZh         = "数据上报"
	proxyColDataTunnelEn         = "Data Report"

	proxyColDescExportIPZh                 = "Proxy 对外提供服务的出口 IP 地址，供 Agent 连接。"
	proxyColDescExportIPEn                 = "External IP address that Agents use to connect to the Proxy."
	proxyColDescOSTypeZh                   = "目标主机操作系统类型，可选项来自包管理中已启用的 Proxy 包。"
	proxyColDescOSTypeEn                   = "Operating system of the target host. Choices come from enabled Proxy packages in Package Management."
	proxyColDescCPUArchZh                  = "CPU 架构，可选项来自包管理中已启用的 Proxy 包。离线安装时必填；在线安装由系统自动探测，无需填写。"
	proxyColDescCPUArchEn                  = "CPU architecture from enabled Proxy packages. Required for offline installation. Online installation detects it automatically." //nolint:lll // Keep the translation on one line.
	proxyColDescAdvertiseIPZh              = "Proxy 对外通告的服务 IP，不填使用出口 IP。"
	proxyColDescAdvertiseIPEn              = "Service IP advertised by the Proxy. Leave blank to use the Export IP."
	proxyColDescInstallPreOrderedPluginsZh = "安装后是否自动安装预设插件，true / false，不填使用页面默认值。"
	proxyColDescInstallPreOrderedPluginsEn = "Whether to install preset plugins after installation. Enter true or false. Leave blank to use the page default." //nolint:lll // Keep the translation on one line.
	proxyColDescDedicatedInstallerZh       = "是否作为安装跳板，true / false，不填默认 true。"
	proxyColDescDedicatedInstallerEn       = "Whether the Proxy acts as an install jumper. Enter true or false. Defaults to true."
	proxyColDescClusterTunnelZh            = "是否承载 Agent 控制通道，true / false，不填默认 true。"
	proxyColDescClusterTunnelEn            = "Whether the Proxy relays Agent control traffic. Enter true or false. Defaults to true."
	proxyColDescFileTunnelZh               = "是否承载文件传输通道，true / false，不填默认 true。"
	proxyColDescFileTunnelEn               = "Whether the Proxy relays file transfers. Enter true or false. Defaults to true."
	proxyColDescDataTunnelZh               = "是否承载数据上报通道，true / false，不填默认 true。"
	proxyColDescDataTunnelEn               = "Whether the Proxy relays data reports. Enter true or false. Defaults to true."
)

const (
	// Ops template.
	opsTemplateSheetName = "ops_template"
	opsTemplateFileName  = "bk_nodemgr_ops_template.xlsx"

	opsColBizZh              = "业务"
	opsColBizEn              = "Business"
	opsColNetworkAreaZh      = "管控区域"
	opsColNetworkAreaEn      = "Network Area"
	opsColConsoleAreaZh      = "主控机管控区域"
	opsColConsoleAreaEn      = "Console Host Network Area"
	opsColConsoleInnerIPZh   = "主控机 IPv4"
	opsColConsoleInnerIPEn   = "Console Host IPv4"
	opsColConsoleInnerIPV6Zh = "主控机 IPv6"
	opsColConsoleInnerIPV6En = "Console Host IPv6"
	opsColOutBandTypeZh      = "设备类型"
	opsColOutBandTypeEn      = "Device Type"
	opsColOutBandProtocolZh  = "设备协议"
	opsColOutBandProtocolEn  = "Device Protocol"
	opsColOutBandIPZh        = "带外 IP"
	opsColOutBandIPEn        = "Out-of-band IP"
	opsColOutBandPortZh      = "带外端口"
	opsColOutBandPortEn      = "Out-of-band Port"

	opsColDescBizZh              = "目标主机归属业务名称，须与节点管理中的记录一致。"
	opsColDescBizEn              = "Business that owns the target host. Use the name recorded in Node Management."
	opsColDescNetworkAreaZh      = "目标主机所在管控区域名称，须与节点管理中的记录一致。"
	opsColDescNetworkAreaEn      = "Network area of the target host. Use the name recorded in Node Management."
	opsColDescInnerIPZh          = "目标主机内网 IPv4 地址，与「内网 IPv6」至少填写一个，用于匹配已有主机。"
	opsColDescInnerIPEn          = "Internal IPv4 address of the target host. Enter at least one of Internal IPv4 and Internal IPv6 to match an existing host." //nolint:lll // Keep the translation on one line.
	opsColDescInnerIPV6Zh        = "目标主机内网 IPv6 地址，与「内网 IPv4」至少填写一个，用于匹配已有主机。"
	opsColDescInnerIPV6En        = "Internal IPv6 address of the target host. Enter at least one of Internal IPv4 and Internal IPv6 to match an existing host." //nolint:lll // Keep the translation on one line.
	opsColDescConsoleAreaZh      = "主控机所在管控区域名称。主控机通常为跳板机或运维机，由其对目标主机执行带外操作，请选择已安装蓝鲸 GSE Agent 的主机。"
	opsColDescConsoleAreaEn      = "Network area of the console host. This is usually a jump host or operations host that performs out-of-band operations on the target host. Select a host with the BlueKing GSE Agent installed." //nolint:lll // Keep the translation on one line.
	opsColDescConsoleInnerIPZh   = "主控机内网 IPv4 地址，与「主控机 IPv6」至少填写一个；与「主控机管控区域」共同定位主控机。"
	opsColDescConsoleInnerIPEn   = "Internal IPv4 address of the console host. Enter at least one of Console Host IPv4 and Console Host IPv6. The IP and Console Host Network Area identify the console host." //nolint:lll // Keep the translation on one line.
	opsColDescConsoleInnerIPV6Zh = "主控机内网 IPv6 地址，与「主控机 IPv4」至少填写一个；与「主控机管控区域」共同定位主控机。"
	opsColDescConsoleInnerIPV6En = "Internal IPv6 address of the console host. Enter at least one of Console Host IPv4 and Console Host IPv6. The IP and Console Host Network Area identify the console host." //nolint:lll // Keep the translation on one line.
	opsColDescOutBandTypeZh      = "目标主机的带外管理方式。使用板载带外管理时填 onboard；使用外接带外设备时填写该设备对应类型标识，例如接入 NanoKVM 时填 nanokvm。具体取值以实际接入的设备为准。"
	opsColDescOutBandTypeEn      = "Out-of-band management method of the target host. Enter onboard for an onboard controller. For an external device, enter its type identifier, such as nanokvm for NanoKVM." //nolint:lll // Keep the translation on one line.
	opsColDescOutBandProtocolZh  = "带外操作使用的通信协议，请依据实际设备类型选择。ipmi 是服务器带外管理的通用标准协议，绝大多数带外设备均支持；nanokvm-gpio 等则是平台针对特定设备额外支持的协议，仅在接入该类设备时选择。"
	opsColDescOutBandProtocolEn  = "Protocol for out-of-band operations. Choose one supported by the device. Most devices support ipmi. Use protocols such as nanokvm-gpio only with the corresponding device." //nolint:lll // Keep the translation on one line.
	opsColDescOutBandIPZh        = "设备的带外管理地址，通常与主机自身 IP 相互独立。主控机通过该地址访问并操作主机的带外管理功能，请确保该地址对主控机可达。"
	opsColDescOutBandIPEn        = "Out-of-band management address of the device. It is usually separate from the host IP. Make sure this address is reachable from the console host." //nolint:lll // Keep the translation on one line.
	opsColDescOutBandPortZh      = "访问并操作设备带外管理功能所使用的端口，通常使用协议默认端口即可。例如 IPMI 一般使用 623（UDP）。若设备侧修改过端口配置，请填写实际端口。"
	opsColDescOutBandPortEn      = "Port for out-of-band management. The protocol default usually works, for example 623 (UDP) for IPMI. If the device uses a different port, enter that port." //nolint:lll // Keep the translation on one line.
)

const (
	// Column copy shared by multiple templates.
	colInnerIPZh                  = "内网 IPv4"
	colInnerIPEn                  = "Internal IPv4"
	colInnerIPV6Zh                = "内网 IPv6"
	colInnerIPV6En                = "Internal IPv6"
	colAddressingZh               = "寻址方式"
	colAddressingEn               = "Addressing Mode"
	colOSTypeZh                   = "操作系统"
	colOSTypeEn                   = "Operating System"
	colLoginIPZh                  = "登录IP"
	colLoginIPEn                  = "Login IP"
	colLoginPortZh                = "登录端口"
	colLoginPortEn                = "Login Port"
	colLoginUserZh                = "登录用户"
	colLoginUserEn                = "Login Account"
	colIsSudoZh                   = "sudo 提权"
	colIsSudoEn                   = "Sudo Escalation"
	colLoginModeZh                = "登录方式"
	colLoginModeEn                = "Auth Method"
	colCreditZh                   = "密钥/密码"
	colCreditEn                   = "Password / Key"
	colInstallPreOrderedPluginsZh = "安装预设插件"
	colInstallPreOrderedPluginsEn = "Install Preset Plugins"
	colReRegisterZh               = "重新注册AgentID"
	colReRegisterEn               = "Re-register Agent ID"
	colInstallMethodZh            = "安装方式"
	colInstallMethodEn            = "Installation Method"
	colDescInnerIPZh              = "目标主机内网 IPv4 地址，与「内网 IPv6」至少填写一个。"
	colDescInnerIPEn              = "Internal IPv4 address of the target host. Enter at least one of Internal IPv4 and Internal IPv6."
	colDescInnerIPV6Zh            = "目标主机内网 IPv6 地址，与「内网 IPv4」至少填写一个。"
	colDescInnerIPV6En            = "Internal IPv6 address of the target host. Enter at least one of Internal IPv4 and Internal IPv6."
	colDescAddressingZh           = "记录到 CMDB 的寻址方式。static：静态；dynamic：动态。"
	colDescAddressingEn           = "Addressing mode recorded in CMDB. Enter static for static addressing or dynamic for dynamic addressing."
	colDescLoginIPZh              = "用于登录目标主机执行安装的 IP 地址，支持 IPv4 / IPv6，可与内网 IP 不同。"
	colDescLoginIPEn              = "IP address used to log in to the target host for installation. Supports IPv4 and IPv6 and may differ from the internal IP." //nolint:lll // Keep the translation on one line.
	colDescLoginPortZh            = "登录目标主机的 SSH 端口，例如 22。"
	colDescLoginPortEn            = "SSH port used to log in to the target host, for example 22."
	colDescLoginUserZh            = "登录目标主机所使用的用户，例如 root。"
	colDescLoginUserEn            = "Account used to log in to the target host, for example root."
	colDescLoginModeZh            = "登录认证方式。password：密码；keyfile：密钥；password_vault：密码库（需开启密码库开关）。"
	colDescLoginModeEn            = "Use password for a password, keyfile for a private key, or password_vault when Password Vault is enabled."
	colDescCreditZh               = "登录凭证。password 填密码；keyfile 填私钥内容；password_vault 可不填。"
	colDescCreditEn               = "Login credential. For password, enter the password. For keyfile, enter the private key. For password_vault, leave it blank." //nolint:lll // Keep the translation on one line.
	colDescReRegisterZh           = "是否重新注册 AgentID，true / false，不填使用页面默认值。"
	colDescReRegisterEn           = "Whether to re-register the Agent ID. Enter true or false. Leave blank to use the page default."
	colDescIsSudoZh               = "登录用户为非 root 且已配置免密 sudo 时填写 true，安装命令将通过 sudo -n 提权执行；不填默认 false。"
	colDescIsSudoEn               = "Enter true for a non-root account with passwordless sudo configured. Installation runs with sudo -n. Defaults to false." //nolint:lll // Keep the translation on one line.
	colDescInstallMethodZh        = "安装方式。ssh：SSH；wmi：WMI；不填使用系统默认（自动）。Windows 支持 ssh/wmi，非 Windows 仅支持 ssh。"
	colDescInstallMethodEn        = "Enter ssh for SSH or wmi for WMI. Leave blank for automatic selection. Windows supports both methods. Other systems support ssh." //nolint:lll // Keep the translation on one line.
)
