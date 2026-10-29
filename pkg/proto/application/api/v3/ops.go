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
	"errors"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// AutoConvert leaves operation history filters unchanged.
func (x *NodeOpsHistoryListReq) AutoConvert() {
	// Intentionally empty: no request defaults are required.
}

// Validate checks pagination and node roles.
func (x *NodeOpsHistoryListReq) Validate() error {
	if err := validatePage(x.GetPage()); err != nil {
		return err
	}
	for _, role := range x.GetExactIncludeConditions().GetNodeRole() {
		switch types.NodeRole(role) {
		case types.NodeRoleBlank, types.NodeRoleAgent, types.NodeRoleProxy:
		default:
			return fmt.Errorf("unsupported node role: %s", role)
		}
	}
	return nil
}

// ConvertPageToTypes converts pagination to the domain query.
func (x *NodeOpsHistoryListReq) ConvertPageToTypes() (types.Page, error) {
	return convPageToTypes(x.GetPage())
}

// ConvertConditionsToTypes converts history filters to domain conditions.
func (x *NodeOpsHistoryListReq) ConvertConditionsToTypes() (*types.NodeOpsHistoryCondition, error) {
	timeRange, err := convertTimeRangeToTypes(x.GetOperateTimeRange())
	if err != nil {
		return nil, fmt.Errorf("failed to convert operate time range: %w", err)
	}
	exact := x.GetExactIncludeConditions()
	return &types.NodeOpsHistoryCondition{
		OperateTimeRange: timeRange,
		ExactInclude: &types.NodeOpsHistoryExactFields{
			BizID:     exact.GetBkBizId(),
			OperateID: exact.GetOperateId(),
			Operator:  exact.GetOperator(),
			NodeRole:  exact.GetNodeRole(),
		},
	}, nil
}

// ConvertNodeOpsHistoriesFromTypes converts operation metadata without snapshots.
func (x *NodeOpsHistoryListResp) ConvertNodeOpsHistoriesFromTypes(total int64, histories []*types.NodeOpsHistory) {
	items := make([]*NodeOpsHistoryInfo, len(histories))
	for idx, history := range histories {
		operateTime := timeToUnixMilli(history.Info.OperateTime)
		nodeRole := string(history.Info.NodeRole)
		items[idx] = &NodeOpsHistoryInfo{
			OperateId:   &history.Info.OperateID,
			Operator:    &history.Info.Operator,
			OperateTime: &operateTime,
			NodeRole:    &nodeRole,
			BkBizId:     history.Info.BizIDs,
		}
	}
	x.Data = &NodeOpsHistoryListResp_Data{Total: total, Items: items}
}

// AutoConvert leaves the operation ID unchanged.
func (x *NodeOpsHistoryDetailReq) AutoConvert() {
	// Intentionally empty: operation IDs have no defaults.
}

// Validate requires the operation ID.
func (x *NodeOpsHistoryDetailReq) Validate() error {
	if x.GetOperateId() == "" {
		return errors.New("operate_id is required")
	}
	return nil
}

// ConvertSnapshotsFromTypes converts the permitted archived snapshots.
func (x *NodeOpsHistoryDetailResp) ConvertSnapshotsFromTypes(snapshots []*types.NodeOpsHistoryHostSnapshot) {
	items := make([]*NodeOpsHistoryHostSnapshot, len(snapshots))
	for idx, snapshot := range snapshots {
		items[idx] = &NodeOpsHistoryHostSnapshot{
			BkHostId:                    &snapshot.HostID,
			OpsConsoleHostId:            &snapshot.OpsConsoleHostID,
			OpsConsoleHostInneripList:   formatRespSlice(snapshot.OpsConsoleHostInnerIP),
			OpsConsoleHostInneripV6List: formatRespSlice(snapshot.OpsConsoleHostInnerIPV6),
			OpsOutBandIp:                &snapshot.OpsOutBandIP,
			OpsOutBandPort:              &snapshot.OpsOutBandPort,
			OpsOutBandProtocol:          &snapshot.OpsOutBandProtocol,
			OpsOutBandType:              &snapshot.OpsOutBandType,
		}
	}
	x.Data = &NodeOpsHistoryDetailResp_Data{Total: int64(len(items)), Items: items}
}

// Validate checks the update ops fields request body.
func (x *NodeOpsUpdateFieldsAgentReq) Validate() error {
	if len(x.GetHosts()) == 0 {
		return errors.New("hosts can not be empty")
	}

	for _, h := range x.GetHosts() {
		if h.GetBkHostId() <= 0 {
			return errors.New("bk_host_id must be greater than 0")
		}
	}
	hostIDs := conv.SliceToSlice(x.GetHosts(), func(host *NodeOpsInfo) int64 {
		return host.GetBkHostId()
	})
	if len(conv.SliceUnique(hostIDs)) != len(hostIDs) {
		return errors.New("duplicated bk_host_id")
	}

	return nil
}

// AutoConvert auto convert.
func (x *NodeOpsUpdateFieldsAgentReq) AutoConvert() {}

// ConvertParamToTypes converts update ops fields request to types hosts.
func (x *NodeOpsUpdateFieldsAgentReq) ConvertParamToTypes() []*types.Host {
	hosts := x.GetHosts()
	if hosts == nil {
		return nil
	}

	hostsParam := make([]*types.Host, len(hosts))
	for idx, host := range hosts {
		hostsParam[idx] = &types.Host{
			HostID: host.GetBkHostId(),
			Dynamic: &types.HostDynamic{
				OpsConsoleHostID:   host.GetOpsConsoleHostId(),
				OpsOutBandType:     host.GetOpsOutBandType(),
				OpsOutBandProtocol: host.GetOpsOutBandProtocol(),
				OpsOutBandIP:       host.GetOpsOutBandIp(),
				OpsOutBandPort:     host.GetOpsOutBandPort(),
			},
		}
	}

	return hostsParam
}

// Validate checks the update ops fields request body.
func (x *NodeOpsUpdateFieldsProxyReq) Validate() error {
	if len(x.GetHosts()) == 0 {
		return errors.New("hosts can not be empty")
	}

	for _, h := range x.GetHosts() {
		if h.GetBkHostId() <= 0 {
			return errors.New("bk_host_id must be greater than 0")
		}
	}
	hostIDs := conv.SliceToSlice(x.GetHosts(), func(host *NodeOpsInfo) int64 {
		return host.GetBkHostId()
	})
	if len(conv.SliceUnique(hostIDs)) != len(hostIDs) {
		return errors.New("duplicated bk_host_id")
	}

	return nil
}

// AutoConvert auto convert.
func (x *NodeOpsUpdateFieldsProxyReq) AutoConvert() {}

// ConvertParamToTypes converts update ops fields request to types hosts.
func (x *NodeOpsUpdateFieldsProxyReq) ConvertParamToTypes() []*types.Host {
	hosts := x.GetHosts()
	if hosts == nil {
		return nil
	}

	hostsParam := make([]*types.Host, len(hosts))
	for idx, host := range hosts {
		hostsParam[idx] = &types.Host{
			HostID: host.GetBkHostId(),
			Dynamic: &types.HostDynamic{
				OpsConsoleHostID:   host.GetOpsConsoleHostId(),
				OpsOutBandType:     host.GetOpsOutBandType(),
				OpsOutBandProtocol: host.GetOpsOutBandProtocol(),
				OpsOutBandIP:       host.GetOpsOutBandIp(),
				OpsOutBandPort:     host.GetOpsOutBandPort(),
			},
		}
	}

	return hostsParam
}
