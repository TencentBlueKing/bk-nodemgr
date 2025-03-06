/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package topo

import (
	proto "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

const (
	maxHostLimit = 1000
)

// ListHost lists hosts with page and conditions.
func (h *handler) ListHost(ctx *rest.Context) (interface{}, error) {
	req := new(proto.TopoHostListReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.Errorf("failed to list host, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	if err := validateTopoHostListReq(req); err != nil {
		h.logger.Errorf("failed to list host, failed to validate request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	sCtx, err := ctx.GetContext()
	if err != nil {
		h.logger.Errorf("failed to list host, failed to get request context. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	hosts, num, err := h.storage.ListHosts(
		sCtx,
		generatePage(req.GetPage(), maxHostLimit),
		generateHostConditions(req))
	if err != nil {
		h.logger.Errorf("failed to list host. err: %v", err)
		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	items := make([]*proto.Host, len(hosts))
	for idx, host := range hosts {
		item := newEmptyHost()
		*item.TenantId = host.TenantID
		*item.BkHostId = host.HostID

		*item.Info.BkBizId = host.Static.BizID
		*item.Info.BkNetworkareaId = host.Static.NetworkAreaID
		*item.Info.BkHostName = host.Static.HostName
		*item.Info.DeptName = host.Static.DeptName
		*item.Info.BkHostInnerip = host.Static.InnerIP
		*item.Info.BkHostInneripV6 = host.Static.InnerIPV6
		*item.Info.BkHostOuterip = host.Static.OuterIP
		*item.Info.BkHostOuteripV6 = host.Static.OuterIPV6
		*item.Info.BkMac = host.Static.Mac
		*item.Info.BkOsType = host.Static.OSType

		*item.State.NodeRole = string(host.Dynamic.NodeRole)
		*item.State.NodeStatus = string(host.Dynamic.NodeStatus)
		*item.State.NodeVersion = host.Dynamic.NodeVersion
		*item.State.BkAgentId = host.Dynamic.AgentID

		items[idx] = item
	}

	resp := &proto.TopoHostListResp_Data{
		Total: num,
		Items: items,
	}

	return resp, nil
}

func validateTopoHostListReq(req *proto.TopoHostListReq) error {
	if err := validateTopoPage(req.GetPage()); err != nil {
		return err
	}

	return nil
}

func generateHostConditions(req *proto.TopoHostListReq) types.HostCondition {
	// exact conditions.
	if exactCond := req.GetExactIncludeConditions(); exactCond != nil {
		conditions := types.HostCondition{
			Type: types.ConditionTypeExactInclude,
		}
		conditions.Exact = &types.HostExactFields{
			HostID:        exactCond.GetBkHostId(),
			BizID:         exactCond.GetBkBizId(),
			NetworkAreaID: exactCond.GetBkNetworkareaId(),
			OSType:        exactCond.GetBkOsType(),
			NodeRole: func(source []string) []types.NodeRole {
				target := make([]types.NodeRole, len(source))
				for idx, s := range source {
					target[idx] = types.NodeRole(s)
				}

				return target
			}(exactCond.GetNodeRole()),
			NodeStatus: func(source []string) []types.NodeStatus {
				target := make([]types.NodeStatus, len(source))
				for idx, s := range source {
					target[idx] = types.NodeStatus(s)
				}

				return target
			}(exactCond.GetNodeStatus()),
			NodeVersion: exactCond.GetNodeVersion(),
			AgentID:     exactCond.GetBkAgentId(),
		}

		return conditions
	}

	// fuzzy conditions.
	if fuzzyCond := req.GetFuzzyIncludeConditions(); fuzzyCond != nil {
		conditions := types.HostCondition{
			Type: types.ConditionTypeFuzzyInclude,
		}
		conditions.Fuzzy = &types.HostFuzzyFields{
			HostName:  fuzzyCond.GetBkHostName(),
			DeptName:  fuzzyCond.GetDeptName(),
			InnerIP:   fuzzyCond.GetBkHostInnerip(),
			InnerIPV6: fuzzyCond.GetBkHostInneripV6(),
			OuterIP:   fuzzyCond.GetBkHostOuterip(),
			OuterIPV6: fuzzyCond.GetBkHostOuteripV6(),
		}

		return conditions
	}

	// default empty conditions.
	return types.HostCondition{
		Type: types.ConditionTypeExactInclude,
	}
}

func newEmptyHost() *proto.Host {
	return &proto.Host{
		TenantId: new(string),
		BkHostId: new(int64),
		Info: &proto.HostInfo{
			BkBizId:         new(int64),
			BkNetworkareaId: new(int64),
			BkHostName:      new(string),
			DeptName:        new(string),
			BkHostInnerip:   new(string),
			BkHostInneripV6: new(string),
			BkHostOuterip:   new(string),
			BkHostOuteripV6: new(string),
			BkMac:           new(string),
			BkOsType:        new(string),
		},
		State: &proto.HostState{
			NodeRole:    new(string),
			NodeStatus:  new(string),
			NodeVersion: new(string),
			BkAgentId:   new(string),
		},
	}
}
