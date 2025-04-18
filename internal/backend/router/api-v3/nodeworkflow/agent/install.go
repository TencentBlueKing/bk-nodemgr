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
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/nodeinstall"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// DefaultNodeGeneration default node generation.
const DefaultNodeGeneration = 2

// AgentInstall install agent.
func (h *handler) AgentInstall(ctx *rest.Context) (interface{}, error) {
	req := new(protoBackend.NodeAgentInstallReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.Error("bind json failed", err)

		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	tenantCtx, err := ctx.GetContext()
	if err != nil {
		h.logger.Error("get tenant context failed", err)

		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	hosts := req.GetHost()
	resp := &protoBackend.NodeAgentInstallResp_Data{
		WorkflowId: make([]string, len(hosts)),
	}

	for idx := range hosts {
		reqHost := hosts[idx]

		// push workflow
		triggerID, err := h.pushWorkflow(tenantCtx, ctx.TenantID, reqHost)
		if err != nil {
			return nil, err
		}

		resp.WorkflowId[idx] = triggerID
	}

	return resp, nil
}

func (h *handler) pushWorkflow(tenantCtx context.Context, tenantID string,
	reqHost *protoBackend.NodeAgentInstallReq_Host) (string, error) {

	nodeDeployment, err := h.convAgentInstallReqToNodeDeployment(tenantCtx, tenantID, reqHost)
	if err != nil {
		h.logger.Error("conv agent install reqHost to node deployment failed", err)

		return "", errf.ErrWrap(errf.Aborted, err)
	}

	nodeDeployment.Info.LoginMode = types.LoginMode(reqHost.GetLoginMode())
	switch nodeDeployment.Info.LoginMode {
	case types.LoginModeKeyFile:
		nodeDeployment.Info.LoginKeyFile, err = h.crypter.Encrypt(reqHost.GetLoginKeyFile())
		if err != nil {
			h.logger.Error("encrypt key file failed", err)

			return "", errf.ErrWrap(errf.InvalidParameter, err)
		}
	case types.LoginModePassword:
		nodeDeployment.Info.LoginMode = types.LoginModePassword
		nodeDeployment.Info.LoginPassword, err = h.crypter.Encrypt([]byte(reqHost.GetLoginPassword()))
		if err != nil {
			h.logger.Error("encrypt password failed", err)

			return "", errf.ErrWrap(errf.InvalidParameter, err)
		}
	case types.LoginModeNone:
		nodeDeployment.Info.LoginMode = types.LoginModeNone
	default:
		err = fmt.Errorf("unsupported login mode %s", reqHost.GetLoginMode())
		h.logger.Error(err)

		return "", errf.ErrWrap(errf.InvalidParameter, err)
	}

	if err := h.iDaoNodeDeployment.Create(tenantCtx, nodeDeployment); err != nil {
		h.logger.Error("create node deployment failed", err)

		return "", errf.ErrWrap(errf.Aborted, err)
	}

	triggerID, err := h.manager.Execute(
		tenantCtx,
		&nodeinstall.OperInstallNodeBySSH{
			Token: nodeDeployment.Token,
		},
	)

	if err != nil {
		h.logger.Error("execute operation failed", err)

		return "", errf.ErrWrap(errf.Aborted, err)
	}

	return triggerID, nil
}

func (h *handler) convAgentInstallReqToNodeDeployment(tenantCtx context.Context,
	tenantID string,
	reqHost *protoBackend.NodeAgentInstallReq_Host,
) (*types.NodeDeployment, error) {

	networkUnit, err := h.iDaoNetworkUnit.GetNetworkUnit(tenantCtx, reqHost.GetBkNetworkunitId())
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
		LoginIP:   reqHost.GetLoginIp(),
		LoginPort: reqHost.GetLoginPort(),
		LoginUser: reqHost.GetLoginUser(),
	})

	return nodeDeployment, nil
}
