/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package agent ...
package agent

import (
	"context"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// DefaultNodeGeneration default node generation.
const DefaultNodeGeneration = 2

// AgentInstall install agent.
func (h *handler) AgentInstall(ctx *rest.Context) (interface{}, error) {
	sCtx, err := ctx.GetContext()
	if err != nil {
		h.logger.Errorf("failed to install agent, failed to get request context. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	req := new(protoBackend.NodeAgentInstallReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(sCtx, "failed to install agent, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	hosts := req.GetHost()
	nodeDeploys := make([]*types.NodeDeployment, len(hosts))
	bizIDs := make(map[int64]struct{})
	for idx := range hosts {
		reqHost := hosts[idx]

		nodeDeploy, err := h.generatesInstallDeploys(sCtx, ctx.TenantID, reqHost)
		if err != nil {
			h.logger.ErrorCtxf(sCtx, "failed to install agent, failed to generate node deployment. err: %v", err)

			return nil, errf.ErrWrap(errf.InvalidParameter, err)
		}

		nodeDeploys[idx] = nodeDeploy
		bizIDs[reqHost.GetBkBizId()] = struct{}{}
	}

	workflowID, err := h.manager.LaunchInstallNode(sCtx, manager.InstallNodeParam{
		Type:            types.NodeWorkflowTypeInstallAgent,
		BizIDs:          conv.MapKeyToSlice[int64, struct{}](bizIDs),
		Operator:        ctx.Username,
		NodeDeployments: nodeDeploys,
	})
	if err != nil {
		h.logger.ErrorCtxf(sCtx, "failed to install agent: %v", err)
		return nil, errf.ErrWrap(errf.BackendOperateFailed, err)
	}

	resp := new(protoBackend.NodeAgentInstallResp)
	resp.ConvertWorkflowID(workflowID)

	h.logger.InfoCtxf(sCtx, "launched install agent workflow: %s", workflowID)

	return resp.GetData(), nil
}

func (h *handler) generatesInstallDeploys(
	tenantCtx context.Context,
	tenantID string,
	reqHost *protoBackend.NodeAgentInstallReq_Host) (*types.NodeDeployment, error) {

	nodeDeployment, err := h.convAgentInstallReqToNodeDeployment(tenantCtx, tenantID, reqHost)
	if err != nil {
		h.logger.Error("conv agent install reqHost to node deployment failed", err)

		return nil, err
	}

	nodeDeployment.Info.LoginInfo.Mode = types.LoginMode(reqHost.GetLoginMode())
	switch nodeDeployment.Info.LoginInfo.Mode {
	case types.LoginModeKeyFile:
		nodeDeployment.Info.LoginInfo.KeyFile, err = h.crypter.Encrypt(reqHost.GetLoginKeyFile())
		if err != nil {
			h.logger.Error("encrypt key file failed", err)

			return nil, err
		}
	case types.LoginModePassword:
		nodeDeployment.Info.LoginInfo.Mode = types.LoginModePassword
		nodeDeployment.Info.LoginInfo.Password, err = h.crypter.Encrypt([]byte(reqHost.GetLoginPassword()))
		if err != nil {
			h.logger.Error("encrypt password failed", err)

			return nil, err
		}
	case types.LoginModeNone:
		nodeDeployment.Info.LoginInfo.Mode = types.LoginModeNone
	default:
		err = fmt.Errorf("unsupported login mode %s", reqHost.GetLoginMode())
		h.logger.Error(err)

		return nil, err
	}

	return nodeDeployment, nil
}

func (h *handler) convAgentInstallReqToNodeDeployment(tenantCtx context.Context,
	tenantID string,
	reqHost *protoBackend.NodeAgentInstallReq_Host,
) (*types.NodeDeployment, error) {

	networkUnit, err := h.storageNetworkUnit.GetNetworkUnit(tenantCtx, reqHost.GetBkNetworkunitId())
	if err != nil {
		return nil, fmt.Errorf("get network unit failed, err: %w", err)
	}

	nodeDeployment := types.NewNodeDeployment(&types.DeploymentInfo{
		Host: types.Host{
			TenantID: tenantID,
			HostID:   reqHost.GetBkHostId(),
			Static: &types.HostStatic{
				BizID:         reqHost.GetBkBizId(),
				NetworkAreaID: networkUnit.NetworkAreaID,
				InnerIP:       reqHost.GetBkHostInnerip(),
				InnerIPV6:     reqHost.GetBkHostInneripV6(),
				OSType:        reqHost.GetOsType(),
				Addressing:    types.Addressing(reqHost.GetBkAddressing()),
			},
			Dynamic: &types.HostDynamic{
				NodeRole:       types.NodeRoleAgent,
				NodeVersion:    reqHost.GetTargetVersion(),
				NodeGeneration: DefaultNodeGeneration,
				NetworkUnitID:  networkUnit.ID,
			},
		},
		InstallOptions: types.InstallOptions{
			ReRegister: reqHost.GetReRegister(),
		},
		LoginInfo: types.LoginInfo{
			IP:   reqHost.GetLoginIp(),
			Port: reqHost.GetLoginPort(),
			User: reqHost.GetLoginUser(),
		},
	})

	return nodeDeployment, nil
}
