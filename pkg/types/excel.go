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

package types

import "github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"

// ExcelAgentInstallTemplateRow holds the installation values from one Agent template row.
type ExcelAgentInstallTemplateRow struct {
	InnerIP                  string
	InnerIPV6                string
	Addressing               Addressing
	OSType                   criteria.OSType
	LoginIP                  string
	LoginPort                int64
	LoginUser                string
	IsSudo                   bool
	LoginMode                LoginMode
	Credit                   string
	InstallPreOrderedPlugins *bool
	ReRegister               *bool
	InstallMethod            NodeInstallMethod
}

// ExcelProxyInstallTemplateRow holds the installation values from one Proxy template row.
type ExcelProxyInstallTemplateRow struct {
	InnerIP                  string
	InnerIPV6                string
	ExportIP                 string
	Addressing               Addressing
	OSType                   criteria.OSType
	LoginIP                  string
	LoginPort                int64
	LoginUser                string
	IsSudo                   bool
	LoginMode                LoginMode
	Credit                   string
	CPUArch                  string
	AdvertiseIP              string
	InstallMethod            NodeInstallMethod
	InstallPreOrderedPlugins *bool
	ReRegister               *bool
	DedicatedInstaller       *bool
	ClusterTunnel            *bool
	FileTunnel               *bool
	DataTunnel               *bool
}

// ExcelOpsTemplateRowResult holds the matched hosts and parsed settings for one Ops template row.
type ExcelOpsTemplateRowResult struct {
	HostID             int64
	BizID              int64
	BizName            string
	NetworkAreaID      int64
	NetworkAreaName    string
	InnerIP            string
	InnerIPV6          string
	OpsConsoleHostID   int64
	OpsConsoleHostIP   string
	OpsOutBandType     string
	OpsOutBandProtocol string
	OpsOutBandIP       string
	OpsOutBandPort     int64
}
