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
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

const maxNodeOpsHistoryLimit = 500

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

// PageLimit returns the backend history page limit.
func (x *NodeOpsHistoryListReq) PageLimit() int {
	return maxNodeOpsHistoryLimit
}

// PageTimeout returns the timeout for paginated history queries.
func (x *NodeOpsHistoryListReq) PageTimeout() time.Duration {
	return backendPagingListTimeout
}

// ConvertPageToTypes converts pagination to the domain query.
func (x *NodeOpsHistoryListReq) ConvertPageToTypes() (types.Page, error) {
	return convPageToTypes(x.GetPage(), x.PageLimit())
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

// ConvertConditionsFromTypes converts history filters to the backend request.
func (x *NodeOpsHistoryListReq) ConvertConditionsFromTypes(condition *types.NodeOpsHistoryCondition) error {
	if condition == nil {
		return nil
	}
	if condition.ExactExclude != nil {
		return errors.New("exact-exclude not supported")
	}
	if exact := condition.ExactInclude; exact != nil {
		x.ExactIncludeConditions = &NodeOpsHistoryExactConditions{
			BkBizId:   exact.BizID,
			OperateId: exact.OperateID,
			Operator:  exact.Operator,
			NodeRole:  exact.NodeRole,
		}
	}
	if timeRange := condition.OperateTimeRange; timeRange != nil {
		x.OperateTimeRange = &TimeRange{
			StartTimestampSec: timeRange.StartTime.Unix(),
			EndTimestampSec:   timeRange.EndTime.Unix(),
		}
	}
	return nil
}

// ConvertNodeOpsHistoriesToTypes converts operation metadata from the backend response.
func (x *NodeOpsHistoryListResp) ConvertNodeOpsHistoriesToTypes() ([]*types.NodeOpsHistory, int64) {
	items := conv.SliceToSlice(x.GetData().GetItems(), func(item *NodeOpsHistoryInfo) *types.NodeOpsHistory {
		return &types.NodeOpsHistory{
			Info: types.NodeOpsHistoryInfo{
				OperateID:   item.GetOperateId(),
				Operator:    item.GetOperator(),
				OperateTime: time.UnixMilli(item.GetOperateTime()),
				NodeRole:    types.NodeRole(item.GetNodeRole()),
				BizIDs:      item.GetBkBizId(),
			},
		}
	})
	return items, x.GetData().GetTotal()
}

// ConvertSnapshotsToTypes converts archived snapshots from the backend response.
func (x *NodeOpsHistoryDetailResp) ConvertSnapshotsToTypes() []*types.NodeOpsHistoryHostSnapshot {
	return conv.SliceToSlice(x.GetData().GetItems(), func(item *NodeOpsHistoryHostSnapshot) *types.NodeOpsHistoryHostSnapshot {
		return &types.NodeOpsHistoryHostSnapshot{
			HostID:                  item.GetBkHostId(),
			OpsConsoleHostID:        item.GetOpsConsoleHostId(),
			OpsConsoleHostInnerIP:   item.GetOpsConsoleHostInneripList(),
			OpsConsoleHostInnerIPV6: item.GetOpsConsoleHostInneripV6List(),
			OpsOutBandIP:            item.GetOpsOutBandIp(),
			OpsOutBandPort:          item.GetOpsOutBandPort(),
			OpsOutBandProtocol:      item.GetOpsOutBandProtocol(),
			OpsOutBandType:          item.GetOpsOutBandType(),
		}
	})
}

// AutoConvert is a no-op for update ops fields requests.
func (x *NodeOpsUpdateFieldsAgentReq) AutoConvert() {}

// Validate validates the update ops fields request.
func (x *NodeOpsUpdateFieldsAgentReq) Validate() error {
	if len(x.GetHosts()) == 0 {
		return errors.New("hosts can not be empty")
	}

	for _, h := range x.GetHosts() {
		if h.GetBkHostId() <= 0 {
			return errors.New("bk_host_id must be greater than 0")
		}
	}
	if hostIDs := x.GetHostIDs(); len(conv.SliceUnique(hostIDs)) != len(hostIDs) {
		return errors.New("duplicated bk_host_id")
	}

	return nil
}

// ConvertHostToTypes converts update ops hosts to types hosts.
func (x *NodeOpsUpdateFieldsAgentReq) ConvertHostToTypes() []*types.Host {
	hosts := make([]*types.Host, 0, len(x.GetHosts()))
	for _, host := range x.GetHosts() {
		typeHost := &types.Host{
			HostID: host.GetBkHostId(),
			Dynamic: &types.HostDynamic{
				OpsConsoleHostID:   host.GetOpsConsoleHostId(),
				OpsOutBandType:     host.GetOpsOutBandType(),
				OpsOutBandProtocol: host.GetOpsOutBandProtocol(),
				OpsOutBandIP:       host.GetOpsOutBandIp(),
				OpsOutBandPort:     host.GetOpsOutBandPort(),
			},
		}

		hosts = append(hosts, typeHost)
	}

	return hosts
}

// ConvertParamFromTypes converts update ops fields hosts from types.
func (x *NodeOpsUpdateFieldsAgentReq) ConvertParamFromTypes(hosts []*types.Host) {
	if hosts == nil {
		return
	}

	items := make([]*NodeOpsInfo, len(hosts))
	for idx, host := range hosts {
		item := &NodeOpsInfo{}
		if host != nil {
			item.BkHostId = host.HostID
			if host.Dynamic != nil {
				item.OpsConsoleHostId = host.Dynamic.OpsConsoleHostID
				item.OpsOutBandType = host.Dynamic.OpsOutBandType
				item.OpsOutBandProtocol = host.Dynamic.OpsOutBandProtocol
				item.OpsOutBandIp = host.Dynamic.OpsOutBandIP
				item.OpsOutBandPort = host.Dynamic.OpsOutBandPort
			}
		}
		items[idx] = item
	}

	x.Hosts = items
}

// GetHostIDs extracts the host IDs from the update ops fields request.
func (x *NodeOpsUpdateFieldsAgentReq) GetHostIDs() []int64 {
	hostIDs := conv.SliceToSlice(x.GetHosts(), func(h *NodeOpsInfo) int64 {
		return h.GetBkHostId()
	})

	return hostIDs
}

// AutoConvert auto convert.
func (x *NodeOpsUpdateFieldsProxyReq) AutoConvert() {}

// Validate validates the update ops fields request.
func (x *NodeOpsUpdateFieldsProxyReq) Validate() error {
	if len(x.GetHosts()) == 0 {
		return errors.New("hosts can not be empty")
	}

	for _, h := range x.GetHosts() {
		if h.GetBkHostId() <= 0 {
			return errors.New("bk_host_id must be greater than 0")
		}
	}
	if hostIDs := x.GetHostIDs(); len(conv.SliceUnique(hostIDs)) != len(hostIDs) {
		return errors.New("duplicated bk_host_id")
	}

	return nil
}

// ConvertParamFromTypes converts update ops fields hosts from types.
func (x *NodeOpsUpdateFieldsProxyReq) ConvertParamFromTypes(hosts []*types.Host) {
	if hosts == nil {
		return
	}

	items := make([]*NodeOpsInfo, len(hosts))
	for idx, host := range hosts {
		item := &NodeOpsInfo{}
		if host != nil {
			item.BkHostId = host.HostID
			if host.Dynamic != nil {
				item.OpsConsoleHostId = host.Dynamic.OpsConsoleHostID
				item.OpsOutBandType = host.Dynamic.OpsOutBandType
				item.OpsOutBandProtocol = host.Dynamic.OpsOutBandProtocol
				item.OpsOutBandIp = host.Dynamic.OpsOutBandIP
				item.OpsOutBandPort = host.Dynamic.OpsOutBandPort
			}
		}
		items[idx] = item
	}

	x.Hosts = items
}

// ConvertHostToTypes converts update ops hosts to types hosts.
func (x *NodeOpsUpdateFieldsProxyReq) ConvertHostToTypes() []*types.Host {
	hosts := make([]*types.Host, 0, len(x.GetHosts()))
	for _, host := range x.GetHosts() {
		typeHost := &types.Host{
			HostID: host.GetBkHostId(),
			Dynamic: &types.HostDynamic{
				OpsConsoleHostID:   host.GetOpsConsoleHostId(),
				OpsOutBandType:     host.GetOpsOutBandType(),
				OpsOutBandProtocol: host.GetOpsOutBandProtocol(),
				OpsOutBandIP:       host.GetOpsOutBandIp(),
				OpsOutBandPort:     host.GetOpsOutBandPort(),
			},
		}

		hosts = append(hosts, typeHost)
	}

	return hosts
}

// GetHostIDs extracts the host IDs from the update ops fields request.
func (x *NodeOpsUpdateFieldsProxyReq) GetHostIDs() []int64 {
	hostIDs := conv.SliceToSlice(x.GetHosts(), func(h *NodeOpsInfo) int64 {
		return h.GetBkHostId()
	})

	return hostIDs
}
