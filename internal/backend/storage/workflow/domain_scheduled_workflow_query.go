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

package workflow

import (
	"errors"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	daoOperation "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/operation"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
)

func (s *Storage) listScheduledWorkflowOperation(nCtx contextx.IContext, page types.Page,
	workflowID, parentOperInstID string, onlyCount bool) ([]*operation.Operation, int64, error) {

	if err := checkScheduledQueryContext(nCtx); err != nil {
		return nil, 0, err
	}
	if err := page.Validate(); err != nil {
		return nil, 0, fmt.Errorf("invalid operation page: %w", err)
	}
	workflow, err := s.getScheduledWorkflow(nCtx, workflowID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get scheduled workflow: %w", err)
	}
	if workflow.TriggerID == "" {
		return []*operation.Operation{}, 0, nil
	}

	opts := []daoOperation.OptFn{daoOperation.WithTriggerID(workflow.TriggerID)}
	if parentOperInstID != "" {
		parent, err := s.getOperationInstanceDataBriefData(nCtx, parentOperInstID)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to get parent instance: %w", err)
		}
		if err := s.checkScheduledInstance(nCtx, parent, map[string]bool{workflow.TriggerID: true}); err != nil {
			return nil, 0, err
		}
		// Child operations can use another trigger; the verified parent instance defines their scope.
		opts = []daoOperation.OptFn{daoOperation.WithParentOperInstID(parentOperInstID),
			daoOperation.WithParentOperationID(parent.Metadata.OperationID)}
	}
	if onlyCount {
		total, err := s.daoOperation.Count(nCtx, opts...)
		return nil, total, err
	}

	return s.daoOperation.List(nCtx, page, opts...)
}

func (s *Storage) listScheduledWorkflowOperationInstance(nCtx contextx.IContext,
	operationIDs, operInstIDs []string, onlyCount bool) ([]*operation.InstanceBriefData, int64, error) {

	roots, err := s.scheduledQueryRoots(nCtx)
	if err != nil {
		return nil, 0, err
	}
	if len(operationIDs) == 0 && len(operInstIDs) == 0 {
		return nil, 0, errors.New("operation_id or oper_inst_id is required")
	}
	for _, operationID := range operationIDs {
		if err := s.checkScheduledOperation(nCtx, operationID, roots); err != nil {
			return nil, 0, err
		}
	}
	for _, instanceID := range operInstIDs {
		instance, err := s.getOperationInstanceDataBriefData(nCtx, instanceID)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to get instance: %w", err)
		}
		if err := s.checkScheduledInstance(nCtx, instance, roots); err != nil {
			return nil, 0, err
		}
	}
	instances, total, err := s.listOperationInstanceBriefDataWithoutActionInst(nCtx, types.UnlimitedPage(),
		&types.OperInstDataCondition{ExactInclude: &types.OperInstDataExactFields{
			OperationID: operationIDs, OperInstID: operInstIDs,
		}})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list scheduled workflow instances: %w", err)
	}
	// Validate returned records too: operation and instance collections are not tenant-partitioned.
	for _, instance := range instances {
		if err := s.checkScheduledInstance(nCtx, instance, roots); err != nil {
			return nil, 0, err
		}
	}
	if onlyCount {
		return nil, total, nil
	}

	return instances, total, nil
}

func (s *Storage) getScheduledWorkflowOperationInstanceLog(nCtx contextx.IContext,
	operInstID string) (*operation.InstanceData, error) {

	roots, err := s.scheduledQueryRoots(nCtx)
	if err != nil {
		return nil, err
	}
	instance, err := s.getOperationInstanceDataBriefData(nCtx, operInstID)
	if err != nil {
		return nil, fmt.Errorf("failed to get scheduled workflow instance: %w", err)
	}
	if err := s.checkScheduledInstance(nCtx, instance, roots); err != nil {
		return nil, err
	}

	return s.getOperationInstanceData(nCtx, operInstID)
}

func checkScheduledQueryContext(nCtx contextx.IContext) error {
	if nCtx == nil {
		return errors.New("context is required")
	}
	if err := nCtx.CheckTenantID(); err != nil {
		return fmt.Errorf("invalid scheduled query tenant: %w", err)
	}

	return nil
}

func (s *Storage) scheduledQueryRoots(nCtx contextx.IContext) (map[string]bool, error) {
	if err := checkScheduledQueryContext(nCtx); err != nil {
		return nil, err
	}
	workflows, _, err := s.listScheduledWorkflow(nCtx, types.UnlimitedPage())
	if err != nil {
		return nil, fmt.Errorf("failed to list tenant scheduled workflows: %w", err)
	}
	roots := make(map[string]bool, len(workflows))
	for _, workflow := range workflows {
		if workflow.TriggerID != "" {
			roots[workflow.TriggerID] = true
		}
	}

	return roots, nil
}

func (s *Storage) checkScheduledInstance(nCtx contextx.IContext, instance *operation.InstanceBriefData,
	roots map[string]bool) error {

	if instance == nil || instance.Metadata == nil {
		return errors.New("scheduled workflow instance metadata is missing")
	}
	oper, err := s.getOperation(nCtx, instance.Metadata.OperationID)
	if err != nil {
		return fmt.Errorf("failed to get instance operation: %w", err)
	}
	if oper.TriggerID != instance.Metadata.TriggerID {
		return errors.New("instance does not belong to its operation trigger")
	}

	return s.checkScheduledOperation(nCtx, oper.OperationID, roots)
}

func (s *Storage) checkScheduledOperation(nCtx contextx.IContext, operationID string, roots map[string]bool) error {
	visited := make(map[string]bool)
	for operationID != "" {
		if err := nCtx.Err(); err != nil {
			return fmt.Errorf("scheduled workflow query canceled: %w", err)
		}
		if visited[operationID] {
			return errors.New("cyclic scheduled workflow operation ancestry")
		}
		visited[operationID] = true
		oper, err := s.getOperation(nCtx, operationID)
		if err != nil {
			return fmt.Errorf("failed to get scheduled workflow operation: %w", err)
		}
		if roots[oper.TriggerID] {
			return nil
		}
		if oper.Param.ParentOperationID == "" || oper.Param.ParentOperInstID == "" {
			return errors.New("operation is outside the current tenant's scheduled workflow chains")
		}
		if err := s.checkScheduledParentInstance(nCtx, oper); err != nil {
			return err
		}
		operationID = oper.Param.ParentOperationID
	}

	return errors.New("scheduled workflow operation id is required")
}

func (s *Storage) checkScheduledParentInstance(nCtx contextx.IContext, oper *operation.Operation) error {
	parent, err := s.getOperationInstanceDataBriefData(nCtx, oper.Param.ParentOperInstID)
	if err != nil {
		return fmt.Errorf("failed to get ancestor instance: %w", err)
	}
	if parent.Metadata == nil || parent.Metadata.OperationID != oper.Param.ParentOperationID {
		return errors.New("operation parent instance does not belong to its parent operation")
	}
	parentOper, err := s.getOperation(nCtx, oper.Param.ParentOperationID)
	if err != nil {
		return fmt.Errorf("failed to get ancestor operation: %w", err)
	}
	if parent.Metadata.TriggerID != parentOper.TriggerID {
		return errors.New("parent instance trigger does not match its operation")
	}

	return nil
}
