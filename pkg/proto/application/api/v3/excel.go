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

package v3

import "github.com/TencentBlueKing/bk-nodemgr/pkg/types"

// AutoConvert auto convert.
func (x *ExcelParseAgentInstallTemplateReq) AutoConvert() {
	// Intentionally empty: the file is carried by the multipart form.
}

// Validate check body.
func (x *ExcelParseAgentInstallTemplateReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *ExcelParseProxyInstallTemplateReq) AutoConvert() {
	// Intentionally empty: the file is carried by the multipart form.
}

// Validate check body.
func (x *ExcelParseProxyInstallTemplateReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *ExcelParseOpsTemplateReq) AutoConvert() {
	// Intentionally empty: the file is carried by the multipart form.
}

// Validate check body.
func (x *ExcelParseOpsTemplateReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *ExcelDownloadAgentInstallTemplateReq) AutoConvert() {
	// Intentionally empty: download requests are carried by URL path only.
}

// Validate check body.
func (x *ExcelDownloadAgentInstallTemplateReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *ExcelDownloadProxyInstallTemplateReq) AutoConvert() {
	// Intentionally empty: download requests are carried by URL path only.
}

// Validate check body.
func (x *ExcelDownloadProxyInstallTemplateReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *ExcelDownloadOpsTemplateReq) AutoConvert() {
	// Intentionally empty: download requests are carried by URL path only.
}

// Validate check body.
func (x *ExcelDownloadOpsTemplateReq) Validate() error {
	return nil
}

// ConvertFromTypes copies the parsed Agent installation values into the response item.
func (x *ExcelAgentInstallTemplateParsedInfo) ConvertFromTypes(row *types.ExcelAgentInstallTemplateRow) {
	x.BkHostInnerip = row.InnerIP
	x.BkHostInneripV6 = row.InnerIPV6
	x.BkAddressing = string(row.Addressing)
	x.OsType = string(row.OSType)
	x.LoginIp = row.LoginIP
	x.LoginPort = row.LoginPort
	x.LoginUser = row.LoginUser
	x.IsSudo = row.IsSudo
	x.LoginMode = string(row.LoginMode)
	x.Credit = row.Credit
	x.InstallPreOrderedPlugins = row.InstallPreOrderedPlugins
	x.ReRegister = row.ReRegister
	x.InstallMethod = string(row.InstallMethod)
}

// ConvertFromTypes copies the parsed Proxy installation values into the response item.
func (x *ExcelProxyInstallTemplateParsedInfo) ConvertFromTypes(row *types.ExcelProxyInstallTemplateRow) {
	x.BkHostInnerip = row.InnerIP
	x.BkHostInneripV6 = row.InnerIPV6
	x.ExportIp = row.ExportIP
	x.BkAddressing = string(row.Addressing)
	x.OsType = string(row.OSType)
	x.LoginIp = row.LoginIP
	x.LoginPort = row.LoginPort
	x.LoginUser = row.LoginUser
	x.IsSudo = row.IsSudo
	x.LoginMode = string(row.LoginMode)
	x.Credit = row.Credit
	x.CpuArch = row.CPUArch
	x.AdvertiseIp = row.AdvertiseIP
	x.InstallMethod = string(row.InstallMethod)
	x.InstallPreOrderedPlugins = row.InstallPreOrderedPlugins
	x.ReRegister = row.ReRegister
	x.DedicatedInstaller = row.DedicatedInstaller
	x.ClusterTunnel = row.ClusterTunnel
	x.FileTunnel = row.FileTunnel
	x.DataTunnel = row.DataTunnel
}

// ConvertFromTypes copies the resolved Ops template values into the response item.
func (x *ExcelOpsTemplateParsedInfo) ConvertFromTypes(row *types.ExcelOpsTemplateRowResult) {
	x.BkHostId = row.HostID
	x.BkBizId = row.BizID
	x.BkBizName = row.BizName
	x.BkNetworkareaId = row.NetworkAreaID
	x.BkNetworkareaName = row.NetworkAreaName
	x.BkHostInnerip = row.InnerIP
	x.BkHostInneripV6 = row.InnerIPV6
	x.OpsConsoleHostId = row.OpsConsoleHostID
	x.OpsConsoleHostIp = row.OpsConsoleHostIP
	x.OpsOutBandType = row.OpsOutBandType
	x.OpsOutBandProtocol = row.OpsOutBandProtocol
	x.OpsOutBandIp = row.OpsOutBandIP
	x.OpsOutBandPort = row.OpsOutBandPort
}
