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

// NodeWorkflowInstallHost contains recoverable installation inputs without credentials or runtime identifiers.
type NodeWorkflowInstallHost struct {
	HostID                   int64
	BizID                    int64
	NetworkAreaID            int64
	NetworkUnitID            int64
	Addressing               Addressing
	InnerIPList              []string
	InnerIPV6List            []string
	OSType                   string
	LoginIP                  string
	LoginPort                int64
	LoginUser                string
	LoginMode                LoginMode
	InstallMethod            NodeInstallMethod
	IsSudo                   bool
	ReRegister               bool
	InstallPreOrderedPlugins bool
	RenewGSETask             bool
	RenewGSEProc             bool
	ExportIP                 string
	ExportIPV6               string
	AdvertiseIP              string
	AdvertiseIPV6            string
	ProxyTags                []ProxyTag
	ProxyInstallOriginUnitID int64
	RelayDownloadPort        int64
	RelayCallbackPort        int64
	CPUArch                  string
}

// NodeWorkflowInstallParams carries recoverable form inputs for one historical installation operation.
// MissingFields identifies inputs that were not preserved; they must not be presented as historical defaults.
type NodeWorkflowInstallParams struct {
	OperationID   string
	Host          NodeWorkflowInstallHost
	TargetVersion []TargetVersion
	IsManual      bool
	IsOffline     bool
	MissingFields []string
}

// NodeWorkflowInstallParamsResult describes recovered inputs and operations whose deployment data is unavailable.
type NodeWorkflowInstallParamsResult struct {
	Type                    NodeWorkflowType
	Items                   []*NodeWorkflowInstallParams
	UnavailableOperationIDs []string
}
