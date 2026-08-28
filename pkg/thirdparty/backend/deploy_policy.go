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

package backend

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IHandlerDeployPolicy defines the backend Handler for deploy policy.
type IHandlerDeployPolicy interface {
	// ListDeployPolicy lists deploy policy by page and conditions.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param page describes the page info when listing.
	// @param condition the filter conditions.
	// @return the deploy policy list with page and the total count with filter and error.
	ListDeployPolicy(nCtx contextx.IContext, page types.Page, condition *types.DeployPolicyCondition) (
		[]*types.DeployPolicy, int64, error)

	// CountDeployPolicy counts deploy policy by conditions.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param condition the filter conditions.
	// @return the deploy policy count with filter and error.
	CountDeployPolicy(nCtx contextx.IContext, condition *types.DeployPolicyCondition) (int64, error)

	// ExecuteDeployPolicy executes deploy policy by deploy policy id.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param deployPolicyID the deploy policy id.
	// @return the trigger id and error.
	ExecuteDeployPolicy(nCtx contextx.IContext, deployPolicyID int64) (string, error)
}

// ListDeployPolicy lists deploy policy.
func (h *Handler) ListDeployPolicy(nCtx contextx.IContext, page types.Page, condition *types.DeployPolicyCondition) (
	[]*types.DeployPolicy, int64, error) {

	req := &protoBackend.DeployPolicyListReq{
		Page: convertPage(page),
	}
	req.ConvertConditionsFromTypes(condition)

	resp, err := h.cli.listDeployPolicy(nCtx, req)
	if err != nil {
		return nil, 0, err
	}

	total, deployPolicies, err := resp.ConvertDeployPoliciesToTypes()
	if err != nil {
		return nil, 0, err
	}

	return deployPolicies, total, nil
}

// CountDeployPolicy counts deploy policy.
func (h *Handler) CountDeployPolicy(nCtx contextx.IContext, condition *types.DeployPolicyCondition) (int64, error) {
	req := &protoBackend.DeployPolicyListReq{
		OnlyCount: true,
	}
	req.ConvertConditionsFromTypes(condition)

	resp, err := h.cli.listDeployPolicy(nCtx, req)
	if err != nil {
		return 0, err
	}

	return resp.GetData().GetTotal(), nil
}

// ExecuteDeployPolicy executes deploy policy.
func (h *Handler) ExecuteDeployPolicy(nCtx contextx.IContext, deployPolicyID int64) (string, error) {
	req := &protoBackend.DeployPolicyExecuteReq{
		DeployPolicyId: &deployPolicyID,
	}

	resp, err := h.cli.executeDeployPolicy(nCtx, req)
	if err != nil {
		return "", err
	}

	return resp.GetData().GetTriggerId(), nil
}
