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
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth"
	authRouter "github.com/TencentBlueKing/bk-nodemgr/internal/backend/router/api-v3/auth"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/goasync"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/identifier"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// UpdateFieldsAgent updates agent out-of-band settings and records the operation.
func (h *handler) UpdateFieldsAgent(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.NodeOpsUpdateFieldsAgentReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to update ops fields, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	hostIDs := req.GetHostIDs()
	bizIDs, err := h.getUpdateOpsFieldsNodeBizs(rCtx, hostIDs...)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to update ops fields, failed to fetch host bizs")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resources := authRouter.BuildBizResources(bizIDs...)
	if authErr := h.authorizer.Check(rCtx, auth.ActionAgentOperate, resources); authErr != nil {
		logger.G.Biz(rCtx).WithErr(authErr).Error("failed to update ops fields, permission denied")
		return nil, resterrf.ErrWrap(resterrf.PermissionDenied, authErr)
	}

	opsHosts := req.ConvertHostToTypes()
	consoleHosts, err := h.getUpdateOpsFieldsConsoleHosts(rCtx, opsHosts)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to update ops fields, failed to fetch ops console hosts")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	if err := h.updateOpsFields(rCtx, bizIDs, opsHosts); err != nil {
		return nil, err
	}

	h.recordOpsHistory(rCtx, types.NodeRoleAgent, bizIDs, opsHosts, consoleHosts)

	resp := new(protoBackend.NodeOpsUpdateFieldsAgentResp)

	return resp.GetData(), nil
}

// UpdateFieldsProxy updates proxy out-of-band settings and records the operation.
func (h *handler) UpdateFieldsProxy(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.NodeOpsUpdateFieldsProxyReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to update ops fields, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	hostIDs := req.GetHostIDs()
	bizIDs, err := h.getUpdateOpsFieldsNodeBizs(rCtx, hostIDs...)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to update ops fields, failed to fetch host bizs")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resources := authRouter.BuildBizResources(bizIDs...)
	if authErr := h.authorizer.Check(rCtx, auth.ActionProxyOperate, resources); authErr != nil {
		logger.G.Biz(rCtx).WithErr(authErr).Error("failed to update ops fields, permission denied")
		return nil, resterrf.ErrWrap(resterrf.PermissionDenied, authErr)
	}

	opsHosts := req.ConvertHostToTypes()
	consoleHosts, err := h.getUpdateOpsFieldsConsoleHosts(rCtx, opsHosts)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to update ops fields, failed to fetch ops console hosts")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	if err := h.updateOpsFields(rCtx, bizIDs, opsHosts); err != nil {
		return nil, err
	}

	h.recordOpsHistory(rCtx, types.NodeRoleProxy, bizIDs, opsHosts, consoleHosts)

	resp := new(protoBackend.NodeOpsUpdateFieldsProxyResp)

	return resp.GetData(), nil
}

func (h *handler) getUpdateOpsFieldsNodeBizs(rCtx restserver.IContext, hostIDs ...int64) ([]int64, error) {
	hostIDs = conv.SliceUnique(hostIDs)
	cond := &types.HostCondition{
		StaticExactInclude: &types.HostStaticExactFields{HostID: hostIDs},
	}
	selection := &types.HostFieldSelection{HostID: true, BizID: true}
	hosts, _, err := h.storageHost.ListHostWithFields(rCtx, types.UnlimitedPage(), selection, cond)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch hosts for updating ops fields: %w", err)
	}

	dbHostIDs := conv.SliceToSlice(hosts, func(host *types.Host) int64 {
		return host.HostID
	})
	if len(hostIDs) != len(hosts) {
		return nil, fmt.Errorf("some hosts not found when fetching hosts for updating ops fields, hostIDs: %v, dbHostIDs: %v",
			hostIDs, dbHostIDs)
	}

	bizIDs := conv.SliceUnique(conv.SliceToSlice(hosts, func(host *types.Host) int64 {
		return host.Static.BizID
	}))

	return bizIDs, nil
}

func (h *handler) getUpdateOpsFieldsConsoleHosts(rCtx restserver.IContext, opsHosts []*types.Host) (map[int64]*types.Host, error) {
	consoleHostIDs := make([]int64, 0, len(opsHosts))
	for _, host := range opsHosts {
		if host.Dynamic.OpsConsoleHostID > 0 {
			consoleHostIDs = append(consoleHostIDs, host.Dynamic.OpsConsoleHostID)
		}
	}

	consoleHosts := make(map[int64]*types.Host)
	if len(consoleHostIDs) > 0 {
		cond := &types.HostCondition{
			StaticExactInclude: &types.HostStaticExactFields{
				HostID: conv.SliceUnique(consoleHostIDs),
			},
		}
		selection := &types.HostFieldSelection{
			HostID:        true,
			InnerIPList:   true,
			InnerIPV6List: true,
		}
		hosts, _, err := h.storageHost.ListHostWithFields(rCtx, types.UnlimitedPage(), selection, cond)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch ops console hosts: %w", err)
		}

		consoleHosts, err = conv.SliceToMap(hosts, func(host *types.Host) int64 { return host.HostID })
		if err != nil {
			return nil, fmt.Errorf("failed to map ops console hosts: %w", err)
		}
		for _, hostID := range consoleHostIDs {
			if _, ok := consoleHosts[hostID]; !ok {
				return nil, fmt.Errorf("ops console host not found, host id: %d", hostID)
			}
		}
	}

	return consoleHosts, nil
}

func (h *handler) updateOpsFields(rCtx restserver.IContext, bizIDs []int64, opsHosts []*types.Host) error {
	for _, bizID := range bizIDs {
		if err := h.cmdbHandler.EnsureBizOpsCustomFields(rCtx, bizID); err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to update ops fields, failed to ensure CMDB custom fields")
			return resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
		}
	}

	hostIDs := conv.SliceToSlice(opsHosts, func(host *types.Host) int64 {
		return host.HostID
	})
	if err := h.cmdbHandler.UpdateHostOpsFields(rCtx, opsHosts...); err != nil {
		logger.G.Biz(rCtx).With("host-ids", hostIDs).WithErr(err).Error("failed to update ops fields to CMDB")
		return resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	dynamicFields := types.HostDynamicFields{
		OpsConsoleHostID:   true,
		OpsOutBandType:     true,
		OpsOutBandProtocol: true,
		OpsOutBandIP:       true,
		OpsOutBandPort:     true,
	}
	if err := h.storageHost.UpdateHostDynamicFields(rCtx, dynamicFields, opsHosts...); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to update ops fields in DB")
		return resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	logger.G.Biz(rCtx).With("host-ids", hostIDs).Info("updated ops fields")

	return nil
}

// recordOpsHistory archives successful updates without changing their response on recording failures.
func (h *handler) recordOpsHistory(
	rCtx restserver.IContext, nodeRole types.NodeRole, bizIDs []int64, opsHosts []*types.Host, consoleHosts map[int64]*types.Host) {

	if len(opsHosts) == 0 {
		return
	}

	snapshots := conv.SliceToSlice(opsHosts, func(host *types.Host) *types.NodeOpsHistoryHostSnapshot {
		snapshot := &types.NodeOpsHistoryHostSnapshot{
			HostID:                  host.HostID,
			OpsConsoleHostID:        host.Dynamic.OpsConsoleHostID,
			OpsConsoleHostInnerIP:   make([]string, 0),
			OpsConsoleHostInnerIPV6: make([]string, 0),
			OpsOutBandIP:            host.Dynamic.OpsOutBandIP,
			OpsOutBandPort:          host.Dynamic.OpsOutBandPort,
			OpsOutBandProtocol:      host.Dynamic.OpsOutBandProtocol,
			OpsOutBandType:          host.Dynamic.OpsOutBandType,
		}
		// Unset console hosts retain empty address lists.
		if consoleHost, ok := consoleHosts[host.Dynamic.OpsConsoleHostID]; ok {
			snapshot.OpsConsoleHostInnerIP = append(snapshot.OpsConsoleHostInnerIP, consoleHost.Static.InnerIPList...)
			snapshot.OpsConsoleHostInnerIPV6 = append(snapshot.OpsConsoleHostInnerIPV6, consoleHost.Static.InnerIPV6List...)
		}

		return snapshot
	})

	history := &types.NodeOpsHistory{
		TenantID: rCtx.TenantID(),
		Info: types.NodeOpsHistoryInfo{
			OperateID:   identifier.GenOpsOperateID(),
			Operator:    rCtx.BKUsername(),
			OperateTime: time.Now(),
			NodeRole:    nodeRole,
			BizIDs:      bizIDs,
		},
		Snapshots: snapshots,
	}
	err := h.goAsyncPool.Run(
		rCtx,
		func(nCtx contextx.IContext) error {
			return h.daoNodeOpsHistory.CreateNodeOpsHistory(nCtx, history)
		},
		goasync.WithName("record_node_ops_history"),
	)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).With("operate-id", history.Info.OperateID).Error("failed to submit node ops history recording task")
	}
}
