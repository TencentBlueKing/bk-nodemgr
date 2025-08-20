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
	"encoding/base64"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// DefaultNodeGeneration default node generation.
const DefaultNodeGeneration = 2

// AgentInstall install agent.
func (h *handler) AgentInstall(ctx *restserver.Context) (interface{}, error) {
	req := new(protoBackend.NodeAgentInstallReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to install agent, failed to decode request body. err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	targetVersions := make([]types.TargetVersion, len(req.GetTargetVersion()))
	for idx, version := range req.GetTargetVersion() {
		targetVersions[idx] = types.TargetVersion{
			OsType:  criteria.OSType(version.GetOsType()),
			CPUArch: criteria.CPUArch(version.GetCpuArch()),
			Version: version.GetVersion(),
		}
	}

	nodeDeploys := make([]*types.NodeDeployment, len(req.GetHost()))
	for idx := range req.GetHost() {
		reqHost := req.GetHost()[idx]

		nodeDeploy, err := h.handlerHost(ctx, ctx.TenantID(), reqHost, targetVersions)
		if err != nil {
			h.logger.ErrorCtxf(ctx, "failed to install agent, failed to generate node deployment. err: %v", err)

			return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
		}

		nodeDeploys[idx] = nodeDeploy
	}

	bizIDs := make(map[int64]struct{})
	for _, host := range req.GetHost() {
		bizIDs[host.GetBkBizId()] = struct{}{}
	}

	workflowID, err := h.manager.LaunchInstallNode(ctx, manager.InstallNodeParam{
		Type:            types.NodeWorkflowTypeInstallAgent,
		BizIDs:          conv.MapKeyToSlice[int64, struct{}](bizIDs),
		Operator:        ctx.LoginName(),
		NodeDeployments: nodeDeploys,
	})
	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to install agent: %v", err)
		return nil, resterrf.ErrWrap(resterrf.BackendOperateFailed, err)
	}

	resp := new(protoBackend.NodeAgentInstallResp)
	resp.ConvertWorkflowID(workflowID)

	h.logger.InfoCtxf(ctx, "launched install agent workflow: %s", workflowID)

	return resp.GetData(), nil
}

func (h *handler) handlerHost(
	tenantCtx context.Context,
	tenantID string,
	reqHost *protoBackend.NodeAgentInstallReq_Host,
	targetVersions []types.TargetVersion,
) (*types.NodeDeployment, error) {

	nodeDeployment, err := h.genNodeDeployment(tenantCtx, tenantID, reqHost, targetVersions)
	if err != nil {
		h.logger.Error("conv agent install reqHost to node deployment failed", err)

		return nil, err
	}

	return nodeDeployment, nil
}

// nolint: funlen
func (h *handler) genNodeDeployment(
	tenantCtx context.Context,
	tenantID string,
	reqHost *protoBackend.NodeAgentInstallReq_Host,
	targetVersions []types.TargetVersion,
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
				NodeStatus:     types.NodeStatusInit,
				NodeGeneration: DefaultNodeGeneration,
				NetworkUnitID:  networkUnit.ID,
			},
		},
		LoginInfo: types.LoginInfo{
			IP:   reqHost.GetLoginIp(),
			Port: reqHost.GetLoginPort(),
			User: reqHost.GetLoginUser(),
			Mode: types.LoginMode(reqHost.GetLoginMode()),
		},
		CurrentVersionSupports: types.DeploymentVersionSupports{},
		InstallOptions: types.DeploymentInstallOptions{
			ReRegister: reqHost.GetReRegister(),
		},
		UpgradeOptions:  types.DeploymentUpgradeOptions{},
		RestartOptions:  types.DeploymentRestartOptions{},
		TransferOptions: types.DeploymentTransferOptions{},
		TargetVersion:   targetVersions,
	})

	switch nodeDeployment.Info.LoginInfo.Mode {
	case types.LoginModeKeyFile:
		loginKeyFile, err := base64.StdEncoding.DecodeString(reqHost.GetLoginKeyFile())
		if err != nil {
			h.logger.Errorf("use base64 decode key file failed, err: %v", err)

			return nil, fmt.Errorf("failed to decode key file, err: %w", err)
		}

		err = h.storageHostCredit.StoreHostCredit(
			tenantCtx,
			nodeDeployment.Info.Host.Static.NetworkAreaID,
			nodeDeployment.Info.LoginInfo.IP,
			nodeDeployment.Info.LoginInfo.User,
			nodeDeployment.Info.LoginInfo.Mode,
			loginKeyFile,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to gen node deployment: %w", err)
		}
	case types.LoginModePassword:
		loginPassword := reqHost.GetLoginPassword()

		err = h.storageHostCredit.StoreHostCredit(
			tenantCtx,
			nodeDeployment.Info.Host.Static.NetworkAreaID,
			nodeDeployment.Info.LoginInfo.IP,
			nodeDeployment.Info.LoginInfo.User,
			nodeDeployment.Info.LoginInfo.Mode,
			[]byte(loginPassword),
		)

		if err != nil {
			return nil, fmt.Errorf("failed to gen node deployment: %w", err)
		}
	case types.LoginModePasswordVault:
		// notice: password vault don't need to store password.
	default:
		err = fmt.Errorf("unsupported this login mode. login-mode(%s)", reqHost.GetLoginMode())
		h.logger.Error(err)

		return nil, err
	}

	return nodeDeployment, nil
}
