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
	"errors"
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

	var addressing types.Addressing
	switch req.GetAddressing() {
	case protoBackend.NodeAgentInstallReq_static:
		addressing = types.AddressingStatic
	case protoBackend.NodeAgentInstallReq_dynamic:
		addressing = types.AddressingDynamic
	default:
		h.logger.Error("invalid addressing", errors.New("addressing must be static or dynamic"))

		return nil, errf.ErrWrap(errf.InvalidParameter, errors.New("addressing must be static or dynamic"))
	}

	networkUnit, err := h.iDaoNetworkUnit.GetNetworkUnit(tenantCtx, req.GetNetworkUnitId())
	if err != nil {
		h.logger.Error("get network unit failed", err)

		return nil, errf.ErrWrap(errf.Aborted, err)
	}

	nodeDeployment := types.NewNodeDeployment(&types.DeploymentInfo{
		Host: types.Host{
			TenantID: ctx.TenantID,
			Static: &types.HostStatic{
				BizID:         req.GetBizId(),
				NetworkAreaID: networkUnit.NetworkAreaID,
				InnerIP:       req.GetInnerIp(),
				InnerIPV6:     req.GetInnerIpv6(),
				OSType:        req.GetOsType(),
				Addressing:    addressing,
			},
			Dynamic: &types.HostDynamic{
				NodeRole:       types.NodeRoleAgent,
				NodeVersion:    req.GetTargetVersion(),
				NodeGeneration: types.NodeGeneration(req.GetTargetGeneration()),
				NetworkUnitID:  networkUnit.ID,
			},
		},
		LoginIP:   req.GetLoginIp(),
		LoginPort: req.GetLoginPort(),
		LoginUser: req.GetLoginUser(),
	})

	switch req.GetLoginMode() {
	case protoBackend.NodeAgentInstallReq_LOGIN_MODE_AUTO:
		return nil, errf.ErrWrap(errf.InvalidParameter,
			errors.New("this mode is no implement, please use login_mode_password or login_mode_key"))

	case protoBackend.NodeAgentInstallReq_LOGIN_MODE_KEY:
		nodeDeployment.Info.LoginMode = types.LoginModeKeyFile
		nodeDeployment.Info.LoginKeyFile, err = h.crypter.Encrypt(req.GetLoginKeyFile())
		if err != nil {
			h.logger.Error("encrypt key file failed", err)

			return nil, errf.ErrWrap(errf.InvalidParameter, err)
		}
	case protoBackend.NodeAgentInstallReq_LOGIN_MODE_PASSWORD:
		nodeDeployment.Info.LoginMode = types.LoginModePassword
		nodeDeployment.Info.LoginPassword, err = h.crypter.Encrypt([]byte(req.GetLoginPassword()))
		if err != nil {
			h.logger.Error("encrypt password failed", err)

			return nil, errf.ErrWrap(errf.InvalidParameter, err)
		}
	case protoBackend.NodeAgentInstallReq_LOGIN_MODE_NONE:
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
