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

package agent

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

// AgentUpdateOpsFields batch-updates out-of-band (ops) fields for hosts,
// writing them to local DB and syncing back to CMDB.
// nolint: gocognit
func (h *handler) AgentUpdateOpsFields(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.NodeAgentUpdateOpsFieldsReq)
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

	for _, bizID := range bizIDs {
		if err := h.cmdbHandler.EnsureBizOpsCustomFields(rCtx, bizID); err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to update ops fields, failed to ensure CMDB custom fields")
			return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
		}
	}

	opsHosts := req.ConvertHostToTypes()
	if err := h.cmdbHandler.UpdateHostOpsFields(rCtx, opsHosts...); err != nil {
		logger.G.Biz(rCtx).With("host-ids", hostIDs).WithErr(err).Error("failed to update ops fields to CMDB")
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	dynamicFields := types.HostDynamicFields{
		OpsConsoleHostID:   true,
		OpsOutBandType:     true,
		OpsOutBandProtocol: true,
		OpsBMCIP:           true,
		OpsBMCPort:         true,
	}
	if err := h.storageHost.UpdateHostDynamicFields(rCtx, dynamicFields, opsHosts...); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to update ops fields in DB")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.NodeAgentUpdateOpsFieldsResp)

	return resp.GetData(), nil
}

func (h *handler) getUpdateOpsFieldsNodeBizs(rCtx restserver.IContext, hostIDs ...int64) ([]int64, error) {
	hostIDs = conv.SliceUnique(hostIDs)
	cond := &types.HostCondition{
		StaticExactInclude:  &types.HostStaticExactFields{HostID: hostIDs},
		DynamicExactInclude: &types.HostDynamicExactFields{NodeRole: []types.NodeRole{types.NodeRoleAgent}},
	}
	selection := &types.HostFieldSelection{BizID: true}
	hosts, _, err := h.storageHost.ListHostWithFields(rCtx, types.UnlimitedPage(), selection, cond)
	if err != nil {
		logger.G.Biz(rCtx).With("host-ids", hostIDs).WithErr(err).Error("failed to fetch hosts for updating ops fields")
		return nil, err
	}

	dbHostIDs := conv.SliceToSlice(hosts, func(h *types.Host) int64 {
		return h.HostID
	})

	if len(hostIDs) != len(hosts) {
		diffResult := func() []int64 {
			dbHostMap := make(map[int64]struct{}, len(dbHostIDs))
			for _, h := range dbHostIDs {
				dbHostMap[h] = struct{}{}
			}

			missingIDs := make([]int64, 0)
			for _, id := range hostIDs {
				if _, ok := dbHostMap[id]; !ok {
					missingIDs = append(missingIDs, id)
				}
			}

			return missingIDs
		}()

		logger.G.Biz(rCtx).
			With("host-ids", hostIDs).
			With("fetched-hosts", dbHostIDs).
			With("missing-hosts", diffResult).
			Error("some hosts not found or not agent role when fetching hosts for updating ops fields")

		return nil, fmt.Errorf("some hosts not found or not agent role when fetching hosts for updating ops fields, "+
			"hostIDs: %v, dbHostIDs: %v, missingIDs: %v", hostIDs, dbHostIDs, diffResult)
	}

	return conv.SliceToSlice(hosts, func(h *types.Host) int64 {
		return h.Static.BizID
	}), nil
}
