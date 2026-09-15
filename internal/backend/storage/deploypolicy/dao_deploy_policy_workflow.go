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

package deploypolicy

import (
	"errors"
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/deploypolicy-workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/identifier"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
)

func (s *Storage) ensureDeployPolicyWorkflow(nCtx contextx.IContext, operationID, triggerID string,
	policyID int64, operator string) (*types.DeployPolicyWorkflow, error) {

	if operationID == "" || triggerID == "" || policyID <= 0 {
		return nil, base.ErrInvalidParam(errors.New("operation id, trigger id and positive policy id are required"))
	}
	opts := []deploypolicyworkflow.OptFn{
		deploypolicyworkflow.WithOperationID(operationID),
		deploypolicyworkflow.WithDeployPolicyID(policyID),
	}
	workflow, err := s.daoDeployPolicyWorkflow.Get(nCtx, opts...)
	if err != nil {
		if !errors.Is(err, base.ErrRecordNoFound()) {
			return nil, fmt.Errorf("failed to get operation workflow: %w", err)
		}
		workflow, err = s.createDeployPolicyWorkflow(nCtx, operationID, triggerID, policyID, operator)
		if err != nil {
			return nil, fmt.Errorf("failed to ensure operation workflow: %w", err)
		}
	}
	if workflow.TriggerID != triggerID {
		return nil, base.ErrInvalidParam(errors.New("operation trigger id does not match the persisted workflow"))
	}

	return workflow, nil
}

func (s *Storage) createDeployPolicyWorkflow(nCtx contextx.IContext, operationID, triggerID string,
	policyID int64, operator string) (*types.DeployPolicyWorkflow, error) {

	workflow := &types.DeployPolicyWorkflow{
		TenantID: nCtx.TenantID(), WorkflowID: identifier.GenWorkflowID(),
		OperationID: operationID, TriggerID: triggerID, DeployPolicyID: policyID,
		Operator: operator, OperateTime: time.Now(),
		Children: make([]types.DeployPolicyWorkflowChild, 0),
	}
	if err := s.daoDeployPolicyWorkflow.Create(nCtx, workflow); err != nil {
		if !mongo.IsDuplicateKeyError(err) {
			return nil, fmt.Errorf("failed to create operation workflow: %w", err)
		}
		// Concurrent creators must reuse the winner's generated workflow ID.
		existing, getErr := s.daoDeployPolicyWorkflow.Get(nCtx,
			deploypolicyworkflow.WithOperationID(operationID), deploypolicyworkflow.WithDeployPolicyID(policyID))
		if getErr != nil {
			return nil, fmt.Errorf("failed to recover concurrent workflow creation: %w", errors.Join(err, getErr))
		}

		return existing, nil
	}

	return workflow, nil
}

func (s *Storage) getDeployPolicyWorkflow(nCtx contextx.IContext, workflowID string) (*types.DeployPolicyWorkflow, error) {
	if workflowID == "" {
		return nil, base.ErrInvalidParam(errors.New("workflow id is empty"))
	}
	workflow, err := s.daoDeployPolicyWorkflow.Get(nCtx, deploypolicyworkflow.WithWorkflowID(workflowID))
	if err != nil {
		return nil, fmt.Errorf("failed to get deploy policy workflow: %w", err)
	}

	return workflow, nil
}

func (s *Storage) recordDeployPolicyWorkflowChild(nCtx contextx.IContext, workflowIDs []string,
	child types.DeployPolicyWorkflowChild) error {

	if err := s.daoDeployPolicyWorkflow.RecordChild(nCtx, workflowIDs, child); err != nil {
		return fmt.Errorf("failed to record deploy policy workflow child: %w", err)
	}

	return nil
}
