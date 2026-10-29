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

package ops

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth"
	authRouter "github.com/TencentBlueKing/bk-nodemgr/internal/backend/router/api-v3/auth"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// ListFieldsAgent reads current agent settings using the agent operation permission.
// It fails when any requested host is unmatched instead of returning a partial list.
func (h *handler) ListFieldsAgent(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.NodeOpsFieldsListAgentReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list agent node ops fields, invalid request")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	// Deduplicate before counting so duplicated IDs never cause a mismatch below.
	hostIDs := conv.SliceUnique(req.GetBkHostId())
	selection := &types.HostFieldSelection{
		HostID:             true,
		BizID:              true,
		OpsConsoleHostID:   true,
		OpsOutBandType:     true,
		OpsOutBandProtocol: true,
		OpsOutBandIP:       true,
		OpsOutBandPort:     true,
	}
	cond := &types.HostCondition{
		StaticExactInclude: &types.HostStaticExactFields{HostID: hostIDs},
	}
	hosts, _, err := h.storageHost.ListHostWithFields(rCtx, types.UnlimitedPage(), selection, cond)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list node ops fields, failed to fetch targets")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	bizIDs := conv.SliceToSlice(hosts, func(host *types.Host) int64 { return host.Static.BizID })
	if err := h.authorizer.Check(rCtx, auth.ActionAgentView, authRouter.BuildBizResources(bizIDs...)); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list node ops fields, permission denied")
		return nil, resterrf.ErrWrap(resterrf.PermissionDenied, err)
	}

	matched, _ := conv.SliceToMapIgnore(hosts, func(host *types.Host) int64 { return host.HostID })
	unmatched := make([]int64, 0, len(hostIDs))
	for _, hostID := range hostIDs {
		if _, ok := matched[hostID]; !ok {
			unmatched = append(unmatched, hostID)
		}
	}

	if len(unmatched) > 0 {
		logger.G.Biz(rCtx).
			With("requested-count", len(hostIDs)).
			With("matched-count", len(hosts)).
			With("unmatched-host-ids", unmatched).
			Error("failed to list node ops fields, requested hosts are not all matched")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, fmt.Errorf("unmatched host IDs: %v", unmatched))
	}

	resp := new(protoBackend.NodeOpsFieldsListAgentResp)
	resp.ConvertHostsFromTypes(hosts)

	return resp.GetData(), nil
}

// ListFieldsProxy reads current proxy settings using the proxy operation permission.
// It fails when any requested host is unmatched instead of returning a partial list.
func (h *handler) ListFieldsProxy(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.NodeOpsFieldsListProxyReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list proxy node ops fields, invalid request")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	// Deduplicate before counting so duplicated IDs never cause a mismatch below.
	hostIDs := conv.SliceUnique(req.GetBkHostId())
	selection := &types.HostFieldSelection{
		HostID:             true,
		BizID:              true,
		OpsConsoleHostID:   true,
		OpsOutBandType:     true,
		OpsOutBandProtocol: true,
		OpsOutBandIP:       true,
		OpsOutBandPort:     true,
	}
	cond := &types.HostCondition{
		StaticExactInclude: &types.HostStaticExactFields{HostID: hostIDs},
	}
	hosts, _, err := h.storageHost.ListHostWithFields(rCtx, types.UnlimitedPage(), selection, cond)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list node ops fields, failed to fetch targets")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	bizIDs := conv.SliceToSlice(hosts, func(host *types.Host) int64 { return host.Static.BizID })
	if err := h.authorizer.Check(rCtx, auth.ActionProxyView, authRouter.BuildBizResources(bizIDs...)); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list node ops fields, permission denied")
		return nil, resterrf.ErrWrap(resterrf.PermissionDenied, err)
	}

	matched, _ := conv.SliceToMapIgnore(hosts, func(host *types.Host) int64 { return host.HostID })
	unmatched := make([]int64, 0, len(hostIDs))
	for _, hostID := range hostIDs {
		if _, ok := matched[hostID]; !ok {
			unmatched = append(unmatched, hostID)
		}
	}

	if len(unmatched) > 0 {
		logger.G.Biz(rCtx).
			With("requested-count", len(hostIDs)).
			With("matched-count", len(hosts)).
			With("unmatched-host-ids", unmatched).
			Error("failed to list node ops fields, requested hosts are not all matched")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, fmt.Errorf("unmatched host IDs: %v", unmatched))
	}

	resp := new(protoBackend.NodeOpsFieldsListProxyResp)
	resp.ConvertHostsFromTypes(hosts)

	return resp.GetData(), nil
}
