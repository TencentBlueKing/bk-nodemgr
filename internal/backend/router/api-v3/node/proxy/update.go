/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package proxy

import (
	"errors"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth"
	authRouter "github.com/TencentBlueKing/bk-nodemgr/internal/backend/router/api-v3/auth"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// Update updates proxy.
func (h *handler) Update(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.NodeProxyUpdateReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to update proxy, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}
	hosts, err := h.getUpdateNodeHosts(rCtx, req.GetHost())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to update proxy, failed to get host list")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}
	bizIDs := conv.SliceUnique(
		conv.SliceToSlice[*types.Host, int64](hosts, func(host *types.Host) int64 {
			return host.Static.BizID
		}))
	resources := authRouter.BuildBizResources(bizIDs...)
	if authErr := h.authorizer.Check(rCtx, auth.ActionProxyOperate, resources); authErr != nil {
		logger.G.Biz(rCtx).WithErr(authErr).Error("failed to update proxy, permission denied")
		return nil, resterrf.ErrWrap(resterrf.PermissionDenied, authErr)
	}

	if err := h.storageHost.UpdateHostDynamicFields(rCtx, req.ConvertHostFieldsToTypes(), req.ConvertHostToTypes()...); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to update proxy, failed to update host")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	hostIDs := make([]int64, 0, len(req.GetHost()))
	for _, host := range req.GetHost() {
		hostIDs = append(hostIDs, host.GetBkHostId())
	}
	if err := h.storageHost.TouchHostOperationTime(rCtx, hostIDs...); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Warn("failed to touch host operation time after proxy update")
	}

	resp := new(protoBackend.NodeProxyUpdateResp)

	return resp.GetData(), nil
}

func (h *handler) getUpdateNodeHosts(
	nCtx contextx.IContext, reqHosts []*protoBackend.NodeProxyUpdateHost) ([]*types.Host, error) {

	if len(reqHosts) == 0 {
		return nil, errors.New("empty host list")
	}

	hostIDs := make(map[int64]struct{})
	for _, host := range reqHosts {
		hostIDs[host.GetBkHostId()] = struct{}{}
	}

	hosts, _, err := h.storageHost.ListHost(nCtx,
		types.UnlimitedPage(),
		&types.HostCondition{
			StaticExactInclude: &types.HostStaticExactFields{
				HostID: conv.MapKeyToSlice(hostIDs),
			}})
	if err != nil {
		return nil, err
	}

	for _, host := range hosts {
		if host.Dynamic.NodeRole != types.NodeRoleProxy {
			return nil, fmt.Errorf("node role is not proxy. host-id(%d), node-role(%s)", host.HostID, host.Dynamic.NodeRole)
		}
	}

	return hosts, nil
}
