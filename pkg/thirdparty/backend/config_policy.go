/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package backend

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IHandlerConfigPolicy defines the backend Handler for config policy.
type IHandlerConfigPolicy interface {
	// ListConfigPolicy lists config policy by page and conditions.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param page describes the page info when listing.
	// @param condition the filter conditions.
	// @return the config policy list with page and the total count with filter and error.
	ListConfigPolicy(nCtx contextx.IContext, page types.Page, condition *types.ConfigPolicyCondition) ([]*types.ConfigPolicy, int64, error)

	// GetConfigPolicy gets config policy by config policy id.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param configPolicyID the config policy id.
	// @return the config policy instance and error.
	GetConfigPolicy(nCtx contextx.IContext, configPolicyID int64) (*types.ConfigPolicy, error)

	// CountConfigPolicy counts config policy by conditions.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param condition the filter conditions.
	// @return the config policy count with filter and error.
	CountConfigPolicy(nCtx contextx.IContext, condition *types.ConfigPolicyCondition) (int64, error)

	// CreateConfigPolicy creates config policy.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param configPolicy the config policy instance to create.
	// @return the created config policy id and error.
	CreateConfigPolicy(nCtx contextx.IContext, configPolicy *types.ConfigPolicy) (int64, error)

	// UpdateConfigPolicy updates config policy.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param configPolicy the config policy instance to update.
	// @return the updated config policy id and error.
	UpdateConfigPolicy(nCtx contextx.IContext, configPolicy *types.ConfigPolicy) (int64, error)

	// EnableConfigPolicy enables config policy by config policy ids.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param configPolicyIDs the config policy ids to enable.
	// @return the error.
	EnableConfigPolicy(nCtx contextx.IContext, configPolicyIDs ...int64) error

	// DisableConfigPolicy disables config policy by config policy ids.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param configPolicyIDs the config policy ids to disable.
	// @return the error.
	DisableConfigPolicy(nCtx contextx.IContext, configPolicyIDs ...int64) error

	// DeleteConfigPolicy deletes config policy by config policy ids.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param configPolicyIDs the config policy ids to delete.
	// @return the error.
	DeleteConfigPolicy(nCtx contextx.IContext, configPolicyIDs ...int64) error

	// ReorderPrioritiesConfigPolicy reorders config policy priorities within a (biz, type) scope.
	// @param bizID the biz scope.
	// @param policyType the config policy type scope.
	// @param orderedPolicyIDs ordered config policy ids to assign priority 1..N;
	//        unlisted enabled policies preserve relative order from N+1.
	ReorderPrioritiesConfigPolicy(nCtx contextx.IContext, bizID int64, policyType types.ConfigPolicyType,
		orderedPolicyIDs []int64) error

	// PreviewConfigPolicy previews the merged config for each host.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param bizID the business id.
	// @param policyType the config policy type.
	// @param hosts the preview host entries.
	// @return the preview result and error.
	PreviewConfigPolicy(nCtx contextx.IContext, bizID int64, policyType types.ConfigPolicyType,
		hosts []types.ConfigPolicyPreviewHost) (*types.ConfigPolicyPreviewResult, error)
}

// ListConfigPolicy lists config policy.
func (h *Handler) ListConfigPolicy(nCtx contextx.IContext, page types.Page, condition *types.ConfigPolicyCondition) (
	[]*types.ConfigPolicy, int64, error) {

	req := &protoBackend.ConfigPolicyListReq{
		Page: convertPage(page),
	}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return nil, 0, err
	}

	resp, err := h.cli.listConfigPolicy(nCtx, req)
	if err != nil {
		return nil, 0, err
	}

	total, configPolicies := resp.ConvertConfigPoliciesToTypes()

	return configPolicies, total, nil
}

// GetConfigPolicy gets config policy by config policy id.
func (h *Handler) GetConfigPolicy(nCtx contextx.IContext, configPolicyID int64) (*types.ConfigPolicy, error) {
	req := &protoBackend.ConfigPolicyGetReq{
		ConfigpolicyId: configPolicyID,
	}

	resp, err := h.cli.getConfigPolicy(nCtx, req)
	if err != nil {
		return nil, err
	}

	return resp.ConvertConfigPolicyToTypes(), nil
}

// CountConfigPolicy counts config policy.
func (h *Handler) CountConfigPolicy(nCtx contextx.IContext, condition *types.ConfigPolicyCondition) (int64, error) {
	req := &protoBackend.ConfigPolicyListReq{
		OnlyCount: true,
	}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return 0, err
	}

	resp, err := h.cli.listConfigPolicy(nCtx, req)
	if err != nil {
		return 0, err
	}

	return resp.GetData().GetTotal(), nil
}

// CreateConfigPolicy creates config policy.
func (h *Handler) CreateConfigPolicy(nCtx contextx.IContext, configPolicy *types.ConfigPolicy) (int64, error) {
	req := new(protoBackend.ConfigPolicyCreateReq)
	req.ConvertConfigPolicyFromTypes(configPolicy)

	resp, err := h.cli.createConfigPolicy(nCtx, req)
	if err != nil {
		return -1, err
	}

	return resp.GetData().GetConfigpolicyId(), nil
}

// UpdateConfigPolicy updates config policy.
func (h *Handler) UpdateConfigPolicy(nCtx contextx.IContext, configPolicy *types.ConfigPolicy) (int64, error) {
	req := new(protoBackend.ConfigPolicyUpdateReq)
	req.ConvertConfigPolicyFromTypes(configPolicy)

	resp, err := h.cli.updateConfigPolicy(nCtx, req)
	if err != nil {
		return -1, err
	}

	return resp.GetData().GetConfigpolicyId(), nil
}

// EnableConfigPolicy enables config policy.
func (h *Handler) EnableConfigPolicy(nCtx contextx.IContext, configPolicyIDs ...int64) error {
	req := &protoBackend.ConfigPolicyEnableReq{ConfigpolicyId: configPolicyIDs}

	_, err := h.cli.enableConfigPolicy(nCtx, req)
	if err != nil {
		return err
	}

	return nil
}

// DisableConfigPolicy disables config policy.
func (h *Handler) DisableConfigPolicy(nCtx contextx.IContext, configPolicyIDs ...int64) error {
	req := &protoBackend.ConfigPolicyDisableReq{ConfigpolicyId: configPolicyIDs}

	_, err := h.cli.disableConfigPolicy(nCtx, req)
	if err != nil {
		return err
	}

	return nil
}

// DeleteConfigPolicy deletes config policy.
func (h *Handler) DeleteConfigPolicy(nCtx contextx.IContext, configPolicyIDs ...int64) error {
	req := &protoBackend.ConfigPolicyDeleteReq{ConfigpolicyId: configPolicyIDs}

	_, err := h.cli.deleteConfigPolicy(nCtx, req)
	if err != nil {
		return err
	}

	return nil
}

// PreviewConfigPolicy previews the merged config for each host.
func (h *Handler) PreviewConfigPolicy(nCtx contextx.IContext, bizID int64, policyType types.ConfigPolicyType,
	hosts []types.ConfigPolicyPreviewHost) (*types.ConfigPolicyPreviewResult, error) {

	req := new(protoBackend.ConfigPolicyPreviewReq)
	req.ConvertFromTypes(bizID, policyType, hosts)

	resp, err := h.cli.previewConfigPolicy(nCtx, req)
	if err != nil {
		return nil, err
	}

	return resp.ConvertMatchResultsToTypes(), nil
}

// ReorderPrioritiesConfigPolicy reorders config policy priorities within a (biz, type) scope.
func (h *Handler) ReorderPrioritiesConfigPolicy(nCtx contextx.IContext, bizID int64,
	policyType types.ConfigPolicyType, orderedPolicyIDs []int64) error {

	req := &protoBackend.ConfigPolicyPriorityReorderReq{
		BizId:                 bizID,
		ConfigpolicyType:      string(policyType),
		OrderedConfigpolicyId: orderedPolicyIDs,
	}

	_, err := h.cli.reorderPrioritiesConfigPolicy(nCtx, req)
	if err != nil {
		return err
	}

	return nil
}
