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

package v3

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// Validate requires an explicit non-negative policy ID, matching execute.
func (x *DeployPolicyPreviewReq) Validate() error {
	if x.DeployPolicyId == nil || x.GetDeployPolicyId() < 0 {
		return fmt.Errorf("deploy policy id must be non-negative")
	}
	return nil
}

// AutoConvert preserves the execute request's missing-ID default.
func (x *DeployPolicyPreviewReq) AutoConvert() {
	if x.DeployPolicyId == nil {
		x.DeployPolicyId = new(int64)
		*x.DeployPolicyId = -1
	}
}

// ConvertPreviewFromTypes converts the complete read-only calculation result.
func (x *DeployPolicyPreviewResp) ConvertPreviewFromTypes(items []*types.DeployPolicySpecPreview) error {
	converted, err := conv.SliceToSliceWithError(items, convDeployPolicySpecPreviewFromTypes)
	if err != nil {
		return fmt.Errorf("failed to convert deploy policy preview: %w", err)
	}
	x.Data = &DeployPolicyPreviewResp_Data{Items: converted}
	return nil
}

func convDeployPolicySpecPreviewFromTypes(item *types.DeployPolicySpecPreview) (*DeployPolicySpecPreview, error) {
	spec, err := convSpecFromTypes(item.Spec)
	if err != nil {
		return nil, fmt.Errorf("failed to convert preview spec: %w", err)
	}
	return &DeployPolicySpecPreview{
		Spec:    spec,
		Results: conv.SliceToSlice(item.Results, convDeployPolicyTargetPreviewFromTypes),
	}, nil
}

func convDeployPolicyTargetPreviewFromTypes(result *types.DeployPolicyTargetPreview) *DeployPolicyTargetPreview {
	target := &DeployPolicyPreviewTarget{
		Host: &DeployPolicyPreviewTarget_Host{BkHostId: result.Target.Host.HostID},
	}
	// Scope host targets have no service instance; its zero ID represents absence.
	if service := result.Target.ServiceInstance; service.ID != 0 {
		target.ServiceInstance = &DeployPolicyPreviewTarget_ServiceInstance{
			Id:         service.ID,
			BkModuleId: service.ModuleID,
		}
	}
	return &DeployPolicyTargetPreview{Target: target, Status: string(result.Status)}
}
