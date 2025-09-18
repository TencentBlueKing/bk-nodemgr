/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package plugin

import (
	"errors"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// createPluginWorkflow create plugin workflow.
func (s *Storage) createPluginWorkflow(ctx contextx.ITenantContext, workflow *types.PluginWorkflow) error {
	if ctx == nil {
		return basestorage.ErrNilContent()
	}

	if workflow == nil {
		return errors.New("workflow is nil")
	}

	if err := s.daoPluginWorkflow.Create(ctx, workflow); err != nil {
		return fmt.Errorf("failed to create plugin workflow: %v", err)
	}

	return nil
}

// getPluginWorkflow get plugin workflow.
func (s *Storage) getPluginWorkflow(ctx contextx.ITenantContext, workflowID string) (*types.PluginWorkflow, error) {
	if ctx == nil {
		return nil, basestorage.ErrNilContent()
	}

	if workflowID == "" {
		return nil, errors.New("workflowID is empty")
	}

	pluginWorkflow, err := s.daoPluginWorkflow.Get(ctx, workflowID)
	if err != nil {
		return nil, fmt.Errorf("failed to get plugin workflow: %v", err)
	}

	return pluginWorkflow, nil
}

// updatePluginWorkflowStatus update plugin workflow status.
func (s *Storage) updatePluginWorkflowStatus(ctx contextx.ITenantContext, workflowID string, status types.PluginWorkflowStatus) error {
	if ctx == nil {
		return basestorage.ErrNilContent()
	}

	if workflowID == "" {
		return errors.New("workflowID is empty")
	}

	if err := status.Validate(); err != nil {
		return fmt.Errorf("status is invalid: %v", err)
	}

	err := s.daoPluginWorkflow.UpdateStatus(ctx, workflowID, status)
	if err != nil {
		return fmt.Errorf("failed to update plugin workflow status: %v", err)
	}

	return nil
}
