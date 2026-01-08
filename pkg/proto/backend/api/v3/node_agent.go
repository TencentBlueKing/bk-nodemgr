/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package v3

import (
	"errors"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// Validate check body.
func (x *NodeAgentInstallReq) Validate() error {
	_, err := conv.SliceToMap(x.GetTargetVersion(), func(v *NodeAgentInstallReq_TargetVersion) string {
		return fmt.Sprintf("%s:%s", v.GetOsType(), v.GetCpuArch())
	})
	if err != nil {
		return err
	}

	hosts := x.GetHost()
	if len(hosts) == 0 {
		return errors.New("host can not be empty")
	}

	for idx := range hosts {
		if err := hosts[idx].Validate(); err != nil {
			return err
		}
	}

	return nil
}

// Validate check body.
// nolint: protogetter
func (x *NodeAgentInstallReq_Host) Validate() error {
	if x.BkHostInnerip == "" && x.BkHostInneripV6 == "" {
		return errors.New("bk_innerip and bk_inneripv6 can not be empty at the same time")
	}

	if x.GetBkBizId() < 0 {
		return errors.New("biz_id must be equal or greater than 0")
	}

	if x.OsType == "" {
		return errors.New("os_type can not be empty")
	}

	if err := types.Addressing(x.GetBkAddressing()).Validate(); err != nil {
		return err
	}

	if x.GetBkNetworkunitId() < 0 {
		return errors.New("network_unit_id must be greater than or equal to 0")
	}

	if x.LoginIp == "" {
		return errors.New("login_ip can not be empty")
	}

	if x.GetLoginPort() < 0 {
		return errors.New("login_port must be greater than 0")
	}

	if len(x.GetLoginUser()) == 0 {
		return errors.New("login_user can not be empty")
	}

	if err := types.LoginMode(x.GetLoginMode()).Validate(); err != nil {
		return err
	}

	return nil
}

// AutoConvert auto convert.
func (x *NodeAgentInstallReq) AutoConvert() {
	hosts := x.GetHost()
	for idx := range hosts {
		hosts[idx].AutoConvert()
	}
}

// AutoConvert auto convert.
func (x *NodeAgentInstallReq_Host) AutoConvert() {
	if x.BkNetworkunitId == nil {
		x.BkNetworkunitId = new(int64)
		*x.BkNetworkunitId = -1
	}

	if x.BkBizId == nil {
		x.BkBizId = new(int64)
		*x.BkBizId = -1
	}

	if x.LoginPort == nil {
		x.LoginPort = new(int64)
		*x.LoginPort = -1
	}

	if x.BkHostId == nil {
		x.BkHostId = new(int64)
		*x.BkHostId = -1
	}
}

// ConvertWorkflowID convert workflow id.
func (x *NodeAgentInstallResp) ConvertWorkflowID(workflowID string) {
	x.Data = &NodeAgentInstallResp_Data{WorkflowId: workflowID}
}

// ConvertHostParamFromTypes convert host param from types.
func (x *NodeAgentInstallReq) ConvertHostParamFromTypes(installParam *types.NodeAgentInstallParam) {
	hostsParam := make([]*NodeAgentInstallReq_Host, len(installParam.NodeAgentInstallHosts))
	for idx, host := range installParam.NodeAgentInstallHosts {
		hostsParam[idx] = &NodeAgentInstallReq_Host{
			BkHostId:        &host.HostID,
			BkBizId:         &host.BizID,
			BkHostInnerip:   host.InnerIP,
			BkHostInneripV6: host.InnerIPV6,
			BkAddressing:    string(host.Addressing),
			LoginIp:         host.LoginIP,
			LoginPort:       &host.LoginPort,
			LoginUser:       host.LoginUser,
			LoginMode:       string(host.LoginMode),
			LoginPassword:   host.LoginPassword,
			LoginKeyFile:    host.LoginKeyFile,
			BkNetworkunitId: &host.NetworkUnitID,
			OsType:          host.OSType,
		}
	}

	targetVersion := make([]*NodeAgentInstallReq_TargetVersion, len(installParam.NodeInstallTargetVersion))
	for idx, version := range installParam.NodeInstallTargetVersion {
		targetVersion[idx] = &NodeAgentInstallReq_TargetVersion{
			Version: version.Version,
			CpuArch: string(version.CPUArch),
			OsType:  string(version.OsType),
		}
	}

	x.TargetVersion = targetVersion
	x.Host = hostsParam
	x.IsManual = installParam.IsManual
}

// ConvertResultToComm ...
func (x *NodeAgentInstallResp) ConvertResultToComm() string {
	return x.GetData().GetWorkflowId()
}

// Validate check body.
func (x *NodeAgentUpgradeReq) Validate() error {
	hosts := x.GetHost()
	if len(hosts) == 0 {
		return errors.New("host can not be empty")
	}

	for idx := range hosts {
		if err := hosts[idx].Validate(); err != nil {
			return err
		}
	}

	return nil
}

// Validate check body.
// nolint: protogetter
func (x *NodeAgentUpgradeReq_Host) Validate() error {
	if x.GetBkHostId() < 0 {
		return errors.New("bk_host_id must be equal or greater than 0")
	}

	if x.GetTargetVersion() == "" {
		return errors.New("target_version can not be empty")
	}

	return nil
}

// ConvertParamFromTypes convert param from types.
func (x *NodeAgentUpgradeReq) ConvertParamFromTypes(upgradeParam *types.NodeAgentUpgradeParam) {
	hostsParam := make([]*NodeAgentUpgradeReq_Host, len(upgradeParam.Hosts))
	for idx, host := range upgradeParam.Hosts {
		hostsParam[idx] = &NodeAgentUpgradeReq_Host{
			BkHostId:                  host.HostID,
			Force:                     host.Force,
			GracefulRestartTimeoutSec: int64(host.GracefulRestartTimeout.Seconds()),
			TargetVersion:             host.TargetVersion,
		}
	}

	x.Host = hostsParam
}

// AutoConvert auto convert.
func (x *NodeAgentUpgradeReq) AutoConvert() {
}

// ConvertWorkflowID convert workflow id.
func (x *NodeAgentUpgradeResp) ConvertWorkflowID(workflowID string) {
	x.Data = &NodeAgentUpgradeResp_Data{WorkflowId: workflowID}
}

// GetWorkflowID get workflow id.
func (x *NodeAgentUpgradeResp) GetWorkflowID() string {
	data := x.GetData()
	if data == nil {
		return ""
	}

	return data.GetWorkflowId()
}

// Validate check body.
func (x *NodeAgentReconfigReq) Validate() error {
	hosts := x.GetHost()
	if len(hosts) == 0 {
		return errors.New("host can not be empty")
	}

	return nil
}

// AutoConvert auto convert.
func (x *NodeAgentReconfigReq) AutoConvert() {
}

// ConvertParamFromTypes convert host param from types.
func (x *NodeAgentReconfigReq) ConvertParamFromTypes(reconfigParam *types.NodeAgentReconfigParam) {
	hostsParam := make([]*NodeAgentReconfigReq_Host, len(reconfigParam.Hosts))
	for idx, host := range reconfigParam.Hosts {
		hostsParam[idx] = &NodeAgentReconfigReq_Host{
			BkHostId:                  host.HostID,
			Force:                     host.Force,
			GracefulRestartTimeoutSec: int64(host.GracefulRestartTimeout.Seconds()),
		}
	}

	x.Host = hostsParam
}

// ConvertWorkflowID convert workflow id.
func (x *NodeAgentReconfigResp) ConvertWorkflowID(workflowID string) {
	x.Data = &NodeAgentReconfigResp_Data{WorkflowId: workflowID}
}

// GetWorkflowID get workflow id.
func (x *NodeAgentReconfigResp) GetWorkflowID() string {
	data := x.GetData()
	if data == nil {
		return ""
	}

	return data.GetWorkflowId()
}

// Validate check body.
func (x *NodeAgentRestartReq) Validate() error {
	hosts := x.GetHost()
	if len(hosts) == 0 {
		return errors.New("host can not be empty")
	}

	return nil
}

// AutoConvert auto convert.
func (x *NodeAgentRestartReq) AutoConvert() {
}

// ConvertParamFromTypes convert host param from types.
func (x *NodeAgentRestartReq) ConvertParamFromTypes(restartParam *types.NodeAgentRestartParam) {
	hostsParam := make([]*NodeAgentRestartReq_Host, len(restartParam.Hosts))
	for idx, host := range restartParam.Hosts {
		hostsParam[idx] = &NodeAgentRestartReq_Host{
			BkHostId:                  host.HostID,
			Force:                     host.Force,
			GracefulRestartTimeoutSec: int64(host.GracefulRestartTimeout.Seconds()),
		}
	}

	x.Host = hostsParam
}

// ConvertWorkflowID convert workflow id.
func (x *NodeAgentRestartResp) ConvertWorkflowID(workflowID string) {
	x.Data = &NodeAgentRestartResp_Data{WorkflowId: workflowID}
}

// GetWorkflowID get workflow id.
func (x *NodeAgentRestartResp) GetWorkflowID() string {
	data := x.GetData()
	if data == nil {
		return ""
	}

	return data.GetWorkflowId()
}

// Validate check body.
func (x *NodeAgentUninstallReq) Validate() error {
	hosts := x.GetHost()
	if len(hosts) == 0 {
		return errors.New("host can not be empty")
	}

	for idx := range hosts {
		if err := hosts[idx].Validate(); err != nil {
			return err
		}
	}

	return nil
}

// Validate check body.
func (x *NodeAgentUninstallReq_Host) Validate() error {
	if x.GetBkHostId() < 0 {
		return errors.New("bk_host_id must be equal or greater than 0")
	}

	return nil
}

// AutoConvert auto convert.
func (x *NodeAgentUninstallReq) AutoConvert() {
}

// ConvertParamFromTypes convert host param from types.
func (x *NodeAgentUninstallReq) ConvertParamFromTypes(restartParam *types.NodeAgentUninstallParam) {
	hostsParam := make([]*NodeAgentUninstallReq_Host, len(restartParam.Hosts))
	for idx, host := range restartParam.Hosts {
		hostsParam[idx] = &NodeAgentUninstallReq_Host{
			BkHostId: host.HostID,
		}
	}

	x.Host = hostsParam
}

// ConvertWorkflowID convert workflow id.
func (x *NodeAgentUninstallResp) ConvertWorkflowID(workflowID string) {
	x.Data = &NodeAgentUninstallResp_Data{WorkflowId: workflowID}
}

// GetWorkflowID get workflow id.
func (x *NodeAgentUninstallResp) GetWorkflowID() string {
	data := x.GetData()
	if data == nil {
		return ""
	}

	return data.GetWorkflowId()
}

// Validate check body.
func (x *NodeAgentInstallCheckReq) Validate() error {
	hosts := x.GetHost()
	if len(hosts) == 0 {
		return errors.New("host can not be empty")
	}

	for idx := range hosts {
		if err := hosts[idx].Validate(); err != nil {
			return err
		}
	}

	return nil
}

// Validate check body.
func (x *NodeAgentInstallCheckReq_Host) Validate() error {
	if x.GetBkBizId() < 0 {
		return errors.New("bk_biz_id is required")
	}

	if len(x.GetBkHostInneripList()) == 0 && len(x.GetBkHostInneripV6List()) == 0 {
		return errors.New("bk_host_innerip_list and bk_host_innerip_v6_list can not be both empty")
	}

	for _, ip := range x.GetBkHostInneripList() {
		if ip == "" {
			return errors.New("bk_host_innerip_list can not contain empty string")
		}
	}

	for _, ipv6 := range x.GetBkHostInneripV6List() {
		if ipv6 == "" {
			return errors.New("bk_host_innerip_v6_list can not contain empty string")
		}
	}

	return nil
}

// AutoConvert auto convert.
func (x *NodeAgentInstallCheckReq) AutoConvert() {
	hosts := x.GetHost()
	for idx := range hosts {
		hosts[idx].AutoConvert()
	}
}

// AutoConvert auto convert.
func (x *NodeAgentInstallCheckReq_Host) AutoConvert() {
	if x.BkNetworkunitId == nil {
		x.BkNetworkunitId = new(int64)
		*x.BkNetworkunitId = -1
	}

	if x.BkBizId == nil {
		x.BkBizId = new(int64)
		*x.BkBizId = -1
	}

	if x.BkHostId == nil {
		x.BkHostId = new(int64)
		*x.BkHostId = -1
	}
}

// ConvertParamFromTypes convert host param from types.
func (x *NodeAgentInstallCheckReq) ConvertParamFromTypes(checkParam []*types.NodeAgentInstallCheckParam) {
	hostsParam := make([]*NodeAgentInstallCheckReq_Host, len(checkParam))
	for idx, host := range checkParam {
		hostsParam[idx] = &NodeAgentInstallCheckReq_Host{
			BkHostId:            &host.HostID,
			BkBizId:             &host.BizID,
			BkNetworkunitId:     &host.NetworkUnitID,
			BkHostInneripList:   host.InnerIPList,
			BkHostInneripV6List: host.InnerIPV6List,
		}
	}

	x.Host = hostsParam
}

// ConvertResultFromTypes convert result from types.
func (x *NodeAgentInstallCheckResp) ConvertResultFromTypes(results []*types.NodeAgentInstallCheckResult) {
	items := make([]*NodeAgentInstallCheckResult, len(results))
	for idx, result := range results {
		item := &NodeAgentInstallCheckResult{
			Status: string(result.Status),
		}

		if result.Matched != nil {
			item.Matched = &NodeAgentInstallCheckMatchedItem{
				BkHostId:            &result.Matched.HostID,
				BkBizId:             &result.Matched.BizID,
				BkNetworkareaId:     &result.Matched.NetworkAreaID,
				BkNetworkunitId:     &result.Matched.NetworkUnitID,
				OsType:              string(result.Matched.OsType),
				NodeRole:            string(result.Matched.NodeRole),
				BkHostInneripList:   result.Matched.InnerIPList,
				BkHostInneripV6List: result.Matched.InnerIPV6List,
			}
		}

		items[idx] = item
	}

	x.Data = &NodeAgentInstallCheckResp_Data{
		Results: items,
	}
}

// ConvertResultToTypes convert result to types.
func (x *NodeAgentInstallCheckResp) ConvertResultToTypes() []*types.NodeAgentInstallCheckResult {
	if x.GetData() == nil {
		return nil
	}

	data := x.GetData()

	items := make([]*types.NodeAgentInstallCheckResult, len(data.GetResults()))
	for idx, result := range data.GetResults() {
		item := &types.NodeAgentInstallCheckResult{
			Status: types.NodeAgentInstallCheckStatus(result.GetStatus()),
		}

		if matched := result.GetMatched(); matched != nil {
			item.Matched = &types.NodeAgentInstallCheckMatchedItem{
				HostID:        matched.GetBkHostId(),
				BizID:         matched.GetBkBizId(),
				NetworkAreaID: matched.GetBkNetworkareaId(),
				NetworkUnitID: matched.GetBkNetworkunitId(),
				OsType:        criteria.OSType(matched.GetOsType()),
				NodeRole:      types.NodeRole(matched.GetNodeRole()),
				InnerIPList:   matched.GetBkHostInneripList(),
				InnerIPV6List: matched.GetBkHostInneripV6List(),
			}
		}

		items[idx] = item
	}

	return items
}
