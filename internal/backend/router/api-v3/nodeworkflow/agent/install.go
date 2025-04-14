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
	"fmt"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/keys"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/nodeinstall"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operengine"
	"github.com/google/uuid"
	"time"
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

	addressing := types.Addressing(req.GetBkAddressing())

	networkUnit, err := h.iDaoNetworkUnit.GetNetworkUnit(tenantCtx, req.GetBkNetworkunitId())
	if err != nil {
		h.logger.Error("get network unit failed", err)

		return nil, errf.ErrWrap(errf.Aborted, err)
	}

	nodeDeployment := types.NewNodeDeployment(&types.DeploymentInfo{
		Host: types.Host{
			TenantID: ctx.TenantID,
			Static: &types.HostStatic{
				BizID:         req.GetBkBizId(),
				NetworkAreaID: networkUnit.NetworkAreaID,
				InnerIP:       req.GetBkHostInnerip(),
				InnerIPV6:     req.GetBkHostInneripV6(),
				OSType:        req.GetOsType(),
				Addressing:    addressing,
			},
			Dynamic: &types.HostDynamic{
				NodeRole:       types.NodeRoleAgent,
				NodeVersion:    req.GetTargetVersion(),
				NodeGeneration: DefaultNodeGeneration,
				NetworkUnitID:  networkUnit.ID,
			},
		},
		LoginIP:   req.GetLoginIp(),
		LoginPort: req.GetLoginPort(),
		LoginUser: req.GetLoginUser(),
	})

	nodeDeployment.Info.LoginMode = types.LoginMode(req.GetLoginMode())
	switch nodeDeployment.Info.LoginMode {
	case types.LoginModeKeyFile:
		nodeDeployment.Info.LoginKeyFile, err = h.crypter.Encrypt(req.GetLoginKeyFile())
		if err != nil {
			h.logger.Error("encrypt key file failed", err)

			return nil, errf.ErrWrap(errf.InvalidParameter, err)
		}
	case types.LoginModePassword:
		nodeDeployment.Info.LoginMode = types.LoginModePassword
		nodeDeployment.Info.LoginPassword, err = h.crypter.Encrypt([]byte(req.GetLoginPassword()))
		if err != nil {
			h.logger.Error("encrypt password failed", err)

			return nil, errf.ErrWrap(errf.InvalidParameter, err)
		}
	case types.LoginModeNone:
		nodeDeployment.Info.LoginMode = types.LoginModeNone
	default:
		err = fmt.Errorf("unsupported login mode %s", req.GetLoginMode())
		h.logger.Error(err)

		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	if err := h.iDaoNodeDeployment.Create(tenantCtx, nodeDeployment); err != nil {
		h.logger.Error("create node deployment failed", err)

		return nil, errf.ErrWrap(errf.Aborted, err)
	}

	triggerID := uuid.New().String()
	triggerID, err = h.manager.ExecuteOperation(
		tenantCtx,
		nodeinstall.OperDefNameInstallNodeBySSH,
		&operengine.OperInstParam{
			Timeout: 1 * time.Minute,
			InitContent: map[string]any{
				keys.CKeyToken: nodeDeployment.Token,
			},
			ParentOperationID: "",
		})

	if err != nil {
		h.logger.Error("execute operation failed", err)

		return nil, errf.ErrWrap(errf.Aborted, err)
	}

	resp := &protoBackend.NodeAgentInstallResp_Data{
		WorkflowId: triggerID,
	}

	return resp, nil
}
