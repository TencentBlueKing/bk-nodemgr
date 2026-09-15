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
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

func (s *Storage) ensureDeployPolicyWorkflow(nCtx contextx.IContext, execution types.DeployPolicyExecutionParam,
	policyID int64, operator string) (*types.DeployPolicyWorkflow, error) {

	workflow, err := s.daoDeployPolicyWorkflow.Ensure(nCtx, execution, policyID, operator)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure deploy policy workflow: %w", err)
	}

	return workflow, nil
}

func (s *Storage) getDeployPolicyWorkflow(nCtx contextx.IContext, workflowID string) (*types.DeployPolicyWorkflow, error) {
	workflow, err := s.daoDeployPolicyWorkflow.Get(nCtx, workflowID)
	if err != nil {
		return nil, fmt.Errorf("failed to get deploy policy workflow: %w", err)
	}

	return workflow, nil
}

func (s *Storage) updateDeployPolicyWorkflowAttempt(nCtx contextx.IContext, workflowIDs []string, operationInstanceID string,
	status types.DeployPolicyWorkflowAttemptStatus, attemptError string) error {

	if err := s.daoDeployPolicyWorkflow.UpdateAttempt(nCtx, workflowIDs, operationInstanceID, status, attemptError); err != nil {
		return fmt.Errorf("failed to update deploy policy workflow attempt: %w", err)
	}

	return nil
}

func (s *Storage) recordDeployPolicyWorkflowChild(nCtx contextx.IContext, workflowIDs []string,
	child types.DeployPolicyWorkflowChild) error {

	if err := s.daoDeployPolicyWorkflow.RecordChild(nCtx, workflowIDs, child); err != nil {
		return fmt.Errorf("failed to record deploy policy workflow child: %w", err)
	}

	return nil
}
