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

package pkg

import (
	"errors"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	packageworkflow "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/package-workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

func (s *Storage) createPackageWorkflow(nCtx contextx.IContext, workflow *types.PackageWorkflow) error {
	if nCtx == nil {
		return basestorage.ErrNilContent()
	}

	if workflow == nil {
		return errors.New("workflow is nil")
	}

	if err := workflow.Type.Validate(); err != nil {
		return fmt.Errorf("invalid workflow data.type: %w", err)
	}

	if err := workflow.Status.Validate(); err != nil {
		return fmt.Errorf("invalid workflow data.status: %w", err)
	}

	if err := s.daoPackageWorkflow.Create(nCtx, workflow); err != nil {
		return fmt.Errorf("failed to create package workflow: %w", err)
	}

	return nil
}

func (s *Storage) getPackageWorkflow(nCtx contextx.IContext, workflowID string) (*types.PackageWorkflow, error) {
	if nCtx == nil {
		return nil, basestorage.ErrNilContent()
	}

	if workflowID == "" {
		return nil, errors.New("workflow id should not be empty")
	}

	return s.daoPackageWorkflow.Get(nCtx, workflowID)
}

func (s *Storage) getPackageExportWorkflowByTriggerID(nCtx contextx.IContext, triggerID string) (*types.PackageWorkflow, error) {
	if nCtx == nil {
		return nil, basestorage.ErrNilContent()
	}
	if triggerID == "" {
		return nil, errors.New("trigger id should not be empty")
	}

	workflows, count, err := s.daoPackageWorkflow.List(
		nCtx,
		types.SingleItemPage(),
		packageworkflow.WithTriggerID(triggerID),
		packageworkflow.WithType(types.PackageWorkflowTypeExport),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list package export workflows: %w", err)
	}
	if count != 1 || len(workflows) != 1 {
		return nil, fmt.Errorf("package export workflow is not unique for trigger id %s", triggerID)
	}

	return workflows[0], nil
}
