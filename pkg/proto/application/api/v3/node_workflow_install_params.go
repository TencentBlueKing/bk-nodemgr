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

import (
	"fmt"
	"strings"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// AutoConvert keeps historical identifiers unchanged.
func (x *NodeWorkflowOperationInstallParamsGetReq) AutoConvert() {}

// Validate checks the selected workflow and operation identifiers.
func (x *NodeWorkflowOperationInstallParamsGetReq) Validate() error {
	if strings.TrimSpace(x.GetWorkflowId()) == "" {
		return fmt.Errorf("workflow_id is required")
	}
	if len(x.GetOperationIds()) == 0 {
		return fmt.Errorf("operation_ids is required")
	}
	seen := make(map[string]struct{}, len(x.GetOperationIds()))
	for index, operationID := range x.GetOperationIds() {
		if strings.TrimSpace(operationID) == "" {
			return fmt.Errorf("operation_ids[%d] is required", index)
		}
		if _, exists := seen[operationID]; exists {
			return fmt.Errorf("operation_ids[%d] is duplicated", index)
		}
		seen[operationID] = struct{}{}
	}
	return nil
}

// ConvertResultFromTypes converts the safe domain projection to response data.
func (x *NodeWorkflowOperationInstallParamsGetResp) ConvertResultFromTypes(result *types.NodeWorkflowInstallParamsResult) {
	x.Data = &NodeWorkflowOperationInstallParamsGetResp_Data{
		Items:                   make([]*NodeWorkflowOperationInstallParams, 0),
		UnavailableOperationIds: make([]string, 0),
	}
	if result == nil {
		return
	}
	x.Data.Type = string(result.Type)
	x.Data.UnavailableOperationIds = formatRespSlice(result.UnavailableOperationIDs)
	x.Data.Items = conv.SliceToSlice(result.Items, func(item *types.NodeWorkflowInstallParams) *NodeWorkflowOperationInstallParams {
		if item == nil {
			return nil
		}
		host := new(NodeWorkflowInstallHost)
		host.ConvertFromTypes(item.Host)
		return &NodeWorkflowOperationInstallParams{
			OperationId: item.OperationID,
			Host:        host,
			TargetVersion: conv.SliceToSlice(item.TargetVersion, func(version types.TargetVersion) *TargetVersion {
				return &TargetVersion{OsType: string(version.OsType), CpuArch: string(version.CPUArch), Version: version.Version}
			}),
			IsManual:      item.IsManual,
			IsOffline:     item.IsOffline,
			MissingFields: formatRespSlice(item.MissingFields),
		}
	})
}

// ConvertFromTypes maps only the explicitly allowed historical host fields.
func (x *NodeWorkflowInstallHost) ConvertFromTypes(host types.NodeWorkflowInstallHost) {
	x.BkHostId = host.HostID
	x.BkBizId = host.BizID
	x.BkNetworkareaId = host.NetworkAreaID
	x.BkNetworkunitId = host.NetworkUnitID
	x.BkAddressing = string(host.Addressing)
	x.BkHostInnerip = formatRespSlice(host.InnerIPList)
	x.BkHostInneripV6 = formatRespSlice(host.InnerIPV6List)
	x.OsType = host.OSType
	x.LoginIp = host.LoginIP
	x.LoginPort = host.LoginPort
	x.LoginUser = host.LoginUser
	x.LoginMode = string(host.LoginMode)
	x.InstallMethod = string(host.InstallMethod)
	x.IsSudo = host.IsSudo
	x.ReRegister = host.ReRegister
	x.InstallPreOrderedPlugins = host.InstallPreOrderedPlugins
	x.RenewGseTask = host.RenewGSETask
	x.RenewGseProc = host.RenewGSEProc
	x.ExportIp = host.ExportIP
	x.ExportIpV6 = host.ExportIPV6
	x.AdvertiseIp = host.AdvertiseIP
	x.AdvertiseIpV6 = host.AdvertiseIPV6
	x.ProxyTags = conv.SliceToSlice(host.ProxyTags, func(tag types.ProxyTag) string { return string(tag) })
	x.ProxyInstallOriginUnitId = host.ProxyInstallOriginUnitID
	x.RelayDownloadPort = host.RelayDownloadPort
	x.RelayCallbackPort = host.RelayCallbackPort
	x.CpuArch = host.CPUArch
}
