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
	"errors"
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// Workflow lists use the same page bound as node and plugin workflows.
const maxDeployPolicyWorkflowLimit = 500

// Validate checks the request page.
func (x *DeployPolicyWorkflowListReq) Validate() error {
	return validatePage(x.GetPage())
}

// AutoConvert leaves optional filters unchanged.
func (x *DeployPolicyWorkflowListReq) AutoConvert() {
	// Intentionally empty: no request defaults are required.
}

// PageLimit returns the backend page limit.
func (x *DeployPolicyWorkflowListReq) PageLimit() int {
	return maxDeployPolicyWorkflowLimit
}

// PageTimeout returns the paging timeout.
func (x *DeployPolicyWorkflowListReq) PageTimeout() time.Duration {
	return backendPagingListTimeout
}

// ConvertPageToTypes converts the request page.
func (x *DeployPolicyWorkflowListReq) ConvertPageToTypes() (types.Page, error) {
	return convPageToTypes(x.GetPage(), x.PageLimit())
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

// ConvertConditionsFromTypes converts domain filters for a backend request.
func (x *DeployPolicyWorkflowListReq) ConvertConditionsFromTypes(condition *types.DeployPolicyWorkflowCondition) error {
	if condition == nil {
		return nil
	}
	if condition.ExactExclude != nil {
		return errors.New("exact-exclude not supported")
	}

	x.OperateTimeRange = convertDeployPolicyTimeRangeFromTypes(condition.OperateTimeRange)
	if exact := condition.ExactInclude; exact != nil {
		x.ExactIncludeConditions = &DeployPolicyWorkflowExactConditions{
			WorkflowId:     exact.WorkflowID,
			DeployPolicyId: exact.DeployPolicyID,
			Operator:       exact.Operator,
			Status: conv.SliceToSlice(exact.Status, func(status types.DeployPolicyWorkflowStatus) string {
				return string(status)
			}),
		}
	}

	return nil
}

// ConvertDeployPolicyWorkflowsFromTypes converts stored parent records without querying children.
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

// ConvertDeployPolicyWorkflowsToTypes converts backend records to domain models.
func (x *DeployPolicyWorkflowListResp) ConvertDeployPolicyWorkflowsToTypes() ([]*types.DeployPolicyWorkflow, int64) {
	data := x.GetData()
	items := conv.SliceToSlice(data.GetItems(), func(item *DeployPolicyWorkflowInfo) *types.DeployPolicyWorkflow {
		workflow := &types.DeployPolicyWorkflow{
			WorkflowID:     item.GetWorkflowId(),
			TriggerID:      item.GetTriggerId(),
			DeployPolicyID: item.GetDeployPolicyId(),
			Operator:       item.GetOperator(),
			OperateTime:    time.UnixMilli(item.GetOperateTime()),
			Status:         types.DeployPolicyWorkflowStatus(item.GetStatus()),
			Children: conv.SliceToSlice(item.GetChildren(), func(child *DeployPolicyChildWorkflow) types.DeployPolicyWorkflowChild {
				return types.DeployPolicyWorkflowChild{
					WorkflowDomain: types.WorkflowDomain(child.GetType()),
					WorkflowID:     child.GetWorkflowId(),
				}
			}),
		}
		// Zero means unfinished, not the Unix epoch.
		if item.GetFinishTime() != 0 {
			workflow.FinishTime = time.UnixMilli(item.GetFinishTime())
		}

		return workflow
	})

	return items, data.GetTotal()
}
