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

package node

import (
	"errors"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	daoNodeDeployment "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/node-deployment"
	daoNodeWorkflow "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/node-workflow"
	daoOperation "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/operation"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/topoevent"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

const (
	// noSuchWorkflowID is a sentinel value used when IP filter returns no results.
	noSuchWorkflowID = "__no_such_workflow__"
)

// listNodeWorkflow lists node workflow by page and conditions.
func (s *Storage) listNodeWorkflow(nCtx contextx.IContext, page types.Page, conditions ...*types.NodeWorkflowCondition) (
	[]*types.NodeWorkflow, int64, error) {

	var results []*types.NodeWorkflow
	var num int64
	var err error

	page.Sort = types.WithSortFields(page.Sort,
		types.WithFieldDesc(daoNodeWorkflow.FieldKeyOperateTime))
	opts, err := s.convertNodeWorkflowConditionsToOptions(nCtx, conditions...)
	if err != nil {
		return nil, 0, err
	}
	if results, num, err = s.daoNodeWorkflow.List(nCtx, page, opts...); err != nil {
		return nil, 0, err
	}

	return results, num, nil
}

// countNodeWorkflow counts node workflow by conditions.
func (s *Storage) countNodeWorkflow(nCtx contextx.IContext, conditions ...*types.NodeWorkflowCondition) (int64, error) {
	var num int64
	var err error

	opts, err := s.convertNodeWorkflowConditionsToOptions(nCtx, conditions...)
	if err != nil {
		return 0, err
	}
	if num, err = s.daoNodeWorkflow.Count(nCtx, opts...); err != nil {
		return 0, err
	}

	return num, nil
}

// distinctNodeWorkflow distincts node workflow fields.
func (s *Storage) distinctNodeWorkflow(
	nCtx contextx.IContext, request types.NodeWorkflowDistinctRequest, conditions ...*types.NodeWorkflowCondition) (
	*types.NodeWorkflowDistinctResult, error) {

	var err error

	result := new(types.NodeWorkflowDistinctResult)
	opts, err := s.convertNodeWorkflowConditionsToOptions(nCtx, conditions...)
	if err != nil {
		return nil, err
	}

	gp := gopool.NewPool()
	if request.BizID {
		gp.Go(func() error {
			var err error
			result.BizID, err = s.daoNodeWorkflow.DistinctNodeWorkflowBkBizID(nCtx, opts...)

			return err
		})
	}
	if request.Operator {
		gp.Go(func() error {
			var err error
			result.Operator, err = s.daoNodeWorkflow.DistinctNodeWorkflowOperator(nCtx, opts...)

			return err
		})
	}
	if request.Status {
		gp.Go(func() error {
			var err error
			result.Status, err = s.daoNodeWorkflow.DistinctNodeWorkflowStatus(nCtx, opts...)

			return err
		})
	}
	if request.Type {
		gp.Go(func() error {
			var err error
			result.Type, err = s.daoNodeWorkflow.DistinctNodeWorkflowType(nCtx, opts...)

			return err
		})
	}

	if err = gp.Wait(); err != nil {
		return nil, err
	}

	return result, nil
}

// getNodeWorkflow gets a node workflow by workflow-id.
func (s *Storage) getNodeWorkflow(nCtx contextx.IContext, workflowID string) (*types.NodeWorkflow, error) {
	if nCtx == nil {
		return nil, basestorage.ErrNilContent()
	}

	if workflowID == "" {
		return nil, errors.New("workflowID cannot be empty")
	}

	workflow, err := s.daoNodeWorkflow.Get(nCtx, workflowID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow by id: %w", err)
	}

	if workflow == nil {
		return nil, fmt.Errorf("workflow not found results, workflow id: %s", workflowID)
	}

	return workflow, nil
}

// createNodeWorkflow creates a new node workflow.
func (s *Storage) createNodeWorkflow(nCtx contextx.IContext, workflow *types.NodeWorkflow) error {
	if nCtx == nil {
		return basestorage.ErrNilContent()
	}

	if workflow == nil {
		return errors.New("workflow cannot be nil")
	}

	if err := workflow.Type.Validate(); err != nil {
		return fmt.Errorf("invalid workflow data.type: %w", err)
	}

	if err := workflow.Status.Validate(); err != nil {
		return fmt.Errorf("invalid workflow data.status: %w", err)
	}

	if err := s.daoNodeWorkflow.Create(nCtx, workflow); err != nil {
		return fmt.Errorf("failed to createNodeDeployment workflow: %w", err)
	}

	return nil
}

// updateNodeWorkflowStatus updates the status of a node workflow.
func (s *Storage) updateNodeWorkflowStatus(nCtx contextx.IContext, workflowID string, status types.NodeWorkflowStatus) error {
	if nCtx == nil {
		return basestorage.ErrNilContent()
	}

	if workflowID == "" {
		return errors.New("workflowID cannot be empty")
	}

	if err := status.Validate(); err != nil {
		return fmt.Errorf("invalid workflow status: %s", status)
	}

	if err := s.daoNodeWorkflow.UpdateStatus(nCtx, workflowID, status); err != nil {
		return err
	}

	return nil
}

// convertNodeWorkflowConditionsToOptions converts node workflow conditions to options.
func (s *Storage) convertNodeWorkflowConditionsToOptions(
	nCtx contextx.IContext, conditions ...*types.NodeWorkflowCondition) ([]daoNodeWorkflow.OptFn, error) {

	opts := make([]daoNodeWorkflow.OptFn, 0)
	for _, condition := range conditions {
		if condition == nil {
			continue
		}

		if condition.OperateTimeRange != nil {
			opts = append(opts, topoevent.WithOperateTimeRange(*condition.OperateTimeRange))
		}

		if condition.ExactInclude == nil {
			continue
		}

		// Handle IP filter condition
		if len(condition.ExactInclude.HostInnerIP) > 0 || len(condition.ExactInclude.HostInnerIPV6) > 0 {
			triggerIDs, err := s.queryTriggerIDsByHostIP(nCtx,
				condition.ExactInclude.HostInnerIP, condition.ExactInclude.HostInnerIPV6)
			if err != nil {
				return nil, err
			}

			if len(triggerIDs) == 0 {
				opts = append(opts, daoNodeWorkflow.WithWorkflowID(noSuchWorkflowID))
				continue
			}

			opts = append(opts, daoNodeWorkflow.WithTriggerID(triggerIDs...))
		}

		opts = append(opts,
			daoNodeWorkflow.WithWorkflowID(condition.ExactInclude.WorkflowID...),
			daoNodeWorkflow.WithBizID(condition.ExactInclude.BizID...),
			daoNodeWorkflow.WithNodeRole(condition.ExactInclude.NodeRole...),
			daoNodeWorkflow.WithType(condition.ExactInclude.Type...),
			daoNodeWorkflow.WithOperator(condition.ExactInclude.Operator...),
			daoNodeWorkflow.WithStatus(condition.ExactInclude.Status...))

		if condition.ExactExclude != nil {
			opts = append(opts,
				daoNodeWorkflow.WithoutWorkflowID(condition.ExactExclude.WorkflowID...),
				daoNodeWorkflow.WithoutBizID(condition.ExactExclude.BizID...),
				daoNodeWorkflow.WithoutType(condition.ExactExclude.Type...),
				daoNodeWorkflow.WithoutOperator(condition.ExactExclude.Operator...),
				daoNodeWorkflow.WithoutStatus(condition.ExactExclude.Status...),
				daoNodeWorkflow.WithoutNodeRole(condition.ExactExclude.NodeRole...))
		}
	}

	return opts, nil
}

func buildNodeDeploymentIPOptions(hostInnerIPs, hostInnerIPV6s []string) []daoNodeDeployment.OptFn {
	opts := make([]daoNodeDeployment.OptFn, 0)
	if len(hostInnerIPs) > 0 {
		opts = append(opts, daoNodeDeployment.WithInfoInnerIP(hostInnerIPs...))
	}
	if len(hostInnerIPV6s) > 0 {
		opts = append(opts, daoNodeDeployment.WithInfoInnerIPV6(hostInnerIPV6s...))
	}

	return opts
}

func (s *Storage) queryTriggerIDsByHostIP(
	nCtx contextx.IContext, hostInnerIPs, hostInnerIPV6s []string) ([]string, error) {

	deployments, _, err := s.daoNodeDeployment.ListNodeDeployment(
		nCtx, types.UnlimitedPage(), buildNodeDeploymentIPOptions(hostInnerIPs, hostInnerIPV6s)...)
	if err != nil {
		return nil, fmt.Errorf("failed to list node deployment by ip: %w", err)
	}
	if len(deployments) == 0 {
		return nil, nil
	}

	tokens := make([]string, 0, len(deployments))
	for _, deployment := range deployments {
		if deployment == nil || deployment.Token == "" {
			continue
		}
		tokens = append(tokens, deployment.Token)
	}
	if len(tokens) == 0 {
		return nil, nil
	}

	operations, _, err := s.daoOperation.List(nCtx, types.UnlimitedPage(), daoOperation.WithInitContentToken(tokens...))
	if err != nil {
		return nil, fmt.Errorf("failed to list operation by deployment token: %w", err)
	}

	triggerSet := make(map[string]struct{}, len(operations))
	for _, op := range operations {
		if op == nil || op.TriggerID == "" {
			continue
		}
		triggerSet[op.TriggerID] = struct{}{}
	}

	triggerIDs := make([]string, 0, len(triggerSet))
	for triggerID := range triggerSet {
		triggerIDs = append(triggerIDs, triggerID)
	}

	return triggerIDs, nil
}
