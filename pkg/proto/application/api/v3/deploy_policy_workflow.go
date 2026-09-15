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

// Validate checks the request page.
func (x *DeployPolicyWorkflowListReq) Validate() error {
	return validatePage(x.GetPage())
}

// AutoConvert leaves optional filters unchanged.
func (x *DeployPolicyWorkflowListReq) AutoConvert() {
	// Intentionally empty: no request defaults are required.
}

// ConvertPageToTypes converts the request page.
func (x *DeployPolicyWorkflowListReq) ConvertPageToTypes() (types.Page, error) {
	return convPageToTypes(x.GetPage())
}

// ConvertConditionsToTypes converts exact filters and the operation time range.
func (x *DeployPolicyWorkflowListReq) ConvertConditionsToTypes() (*types.DeployPolicyWorkflowCondition, error) {
	timeRange, err := convertTimeRangeToTypes(x.GetOperateTimeRange())
	if err != nil {
		return nil, fmt.Errorf("failed to convert operate time range: %w", err)
	}

	condition := &types.DeployPolicyWorkflowCondition{OperateTimeRange: timeRange}
	if exact := x.GetExactIncludeConditions(); exact != nil {
		condition.ExactInclude = &types.DeployPolicyWorkflowExactConditions{
			WorkflowID:     exact.GetWorkflowId(),
			DeployPolicyID: exact.GetDeployPolicyId(),
			Operator:       exact.GetOperator(),
			Status: conv.SliceToSlice(exact.GetStatus(), func(status string) types.DeployPolicyWorkflowStatus {
				return types.DeployPolicyWorkflowStatus(status)
			}),
		}
	}

	return condition, nil
}

// ConvertDeployPolicyWorkflowsFromTypes converts parent records without querying children.
func (x *DeployPolicyWorkflowListResp) ConvertDeployPolicyWorkflowsFromTypes(total int64, workflows []*types.DeployPolicyWorkflow) {
	items := make([]*DeployPolicyWorkflowInfo, 0, len(workflows))
	for _, workflow := range workflows {
		items = append(items, &DeployPolicyWorkflowInfo{
			WorkflowId:     workflow.WorkflowID,
			TriggerId:      workflow.TriggerID,
			DeployPolicyId: workflow.DeployPolicyID,
			Operator:       workflow.Operator,
			OperateTime:    timeToUnixMilli(workflow.OperateTime),
			FinishTime:     timeToUnixMilli(workflow.FinishTime),
			Status:         string(workflow.Status),
			Children: conv.SliceToSlice(workflow.Children, func(child types.DeployPolicyWorkflowChild) *DeployPolicyChildWorkflow {
				return &DeployPolicyChildWorkflow{Type: string(child.WorkflowDomain), WorkflowId: child.WorkflowID}
			}),
		})
	}
	x.Data = &DeployPolicyWorkflowListResp_Data{Total: total, Items: items}
}
