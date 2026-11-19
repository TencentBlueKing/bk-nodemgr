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
	// Shared column names across the installation templates.
	colInnerIP                  = "内网 IPv4"
	colInnerIPV6                = "内网 IPv6"
	colAddressing               = "寻址方式"
	colOSType                   = "操作系统"
	colLoginIP                  = "登录IP"
	colLoginPort                = "登录端口"
	colLoginUser                = "登录用户"
	colIsSudo                   = "sudo 提权"
	colLoginMode                = "登录方式"
	colCredit                   = "密钥/密码"
	colInstallPreOrderedPlugins = "安装预设插件"
	colReRegister               = "重新注册AgentID"
	colInstallMethod            = "安装方式"
)

const (
	// Shared column descriptions across the installation templates.
	colDescInnerIP       = "目标主机内网 IPv4 地址，与「内网 IPv6」至少填写一个。"
	colDescInnerIPV6     = "目标主机内网 IPv6 地址，与「内网 IPv4」至少填写一个。"
	colDescAddressing    = "记录到 CMDB 的寻址方式。static：静态；dynamic：动态。"
	colDescLoginIP       = "用于登录目标主机执行安装的 IP 地址，支持 IPv4 / IPv6，可与内网 IP 不同。"
	colDescLoginPort     = "登录目标主机的 SSH 端口，例如 22。"
	colDescLoginUser     = "登录目标主机所使用的用户，例如 root。"
	colDescLoginMode     = "登录认证方式。password：密码；keyfile：密钥；password_vault：密码库（需开启密码库开关）。"
	colDescCredit        = "登录凭证。password 填密码；keyfile 填私钥内容；password_vault 可不填。"
	colDescReRegister    = "是否重新注册 AgentID，true / false，不填使用页面默认值。"
	colDescIsSudo        = "登录用户为非 root 且已配置免密 sudo 时填写 true，安装命令将通过 sudo -n 提权执行；不填默认 false。"
	colDescInstallMethod = "安装方式。ssh：SSH；wmi：WMI；不填使用系统默认（自动）。Windows 支持 ssh/wmi，非 Windows 仅支持 ssh。"
)

const (
	// Agent install template.
	agentInstallSheetName = "agent_install_template"
	agentInstallFileName  = "agent_install_template.xlsx"

	agentColDescOSType                   = "目标主机操作系统类型，可选项来自包管理中已启用的 Agent 包。"
	agentColDescInstallPreOrderedPlugins = "安装 Agent 后是否自动安装预设插件，true / false，不填使用页面默认值。"
)

const (
	// Proxy install template.
	proxyInstallSheetName = "proxy_install_template"
	proxyInstallFileName  = "proxy_install_template.xlsx"

	proxyColExportIP           = "出口 IP"
	proxyColCPUArch            = "CPU架构"
	proxyColAdvertiseIP        = "服务IP"
	proxyColDedicatedInstaller = "安装跳板"
	proxyColClusterTunnel      = "Agent控制"
	proxyColFileTunnel         = "文件传输"
	proxyColDataTunnel         = "数据上报"

	proxyColDescExportIP                 = "Proxy 对外提供服务的出口 IP 地址，供 Agent 连接。"
	proxyColDescOSType                   = "目标主机操作系统类型，可选项来自包管理中已启用的 Proxy 包。"
	proxyColDescCPUArch                  = "CPU 架构，可选项来自包管理中已启用的 Proxy 包。离线安装时必填；在线安装由系统自动探测，无需填写。"
	proxyColDescAdvertiseIP              = "Proxy 对外通告的服务 IP，不填使用出口 IP。"
	proxyColDescInstallPreOrderedPlugins = "安装后是否自动安装预设插件，true / false，不填使用页面默认值。"
	proxyColDescDedicatedInstaller       = "是否作为安装跳板，true / false，不填默认 true。"
	proxyColDescClusterTunnel            = "是否承载 Agent 控制通道，true / false，不填默认 true。"
	proxyColDescFileTunnel               = "是否承载文件传输通道，true / false，不填默认 true。"
	proxyColDescDataTunnel               = "是否承载数据上报通道，true / false，不填默认 true。"
)

const (
	// Ops template.
	opsTemplateSheetName = "ops_template"
	opsTemplateFileName  = "ops_template.xlsx"

	opsColBiz              = "业务"
	opsColNetworkArea      = "管控区域"
	opsColConsoleArea      = "主控机管控区域"
	opsColConsoleInnerIP   = "主控机 IPv4"
	opsColConsoleInnerIPV6 = "主控机 IPv6"
	opsColOutBandType      = "设备类型"
	opsColOutBandProtocol  = "设备协议"
	opsColOutBandIP        = "带外 IP"
	opsColOutBandPort      = "带外端口"

	opsColDescInnerIP          = "目标主机内网 IPv4 地址，与「内网 IPv6」至少填写一个，用于匹配已有主机。"
	opsColDescInnerIPV6        = "目标主机内网 IPv6 地址，与「内网 IPv4」至少填写一个，用于匹配已有主机。"
	opsColDescNetworkArea      = "目标主机所在管控区域名称，须与节点管理中的记录一致。"
	opsColDescBiz              = "目标主机归属业务名称，须与节点管理中的记录一致。"
	opsColDescConsoleArea      = "主控机所在管控区域名称。主控机通常为跳板机或运维机，由其对目标主机执行带外操作，请选择已安装蓝鲸 GSE Agent 的主机。"
	opsColDescConsoleInnerIP   = "主控机内网 IPv4 地址，与「主控机 IPv6」至少填写一个；与「主控机管控区域」共同定位主控机。"
	opsColDescConsoleInnerIPV6 = "主控机内网 IPv6 地址，与「主控机 IPv4」至少填写一个；与「主控机管控区域」共同定位主控机。"
	opsColDescOutBandType      = "目标主机的带外管理方式。使用板载带外管理时填 onboard；使用外接带外设备时填写该设备对应类型标识，例如接入 NanoKVM 时填 nanokvm。具体取值以实际接入的设备为准。"
	opsColDescOutBandProtocol  = "带外操作使用的通信协议，请依据实际设备类型选择。ipmi 是服务器带外管理的通用标准协议，绝大多数带外设备均支持；nanokvm-gpio 等则是平台针对特定设备额外支持的协议，仅在接入该类设备时选择。"
	opsColDescOutBandIP        = "设备的带外管理地址，通常与主机自身 IP 相互独立。主控机通过该地址访问并操作主机的带外管理功能，请确保该地址对主控机可达。"
	opsColDescOutBandPort      = "访问并操作设备带外管理功能所使用的端口，通常使用协议默认端口即可。例如 IPMI 一般使用 623（UDP）。若设备侧修改过端口配置，请填写实际端口。"
)
