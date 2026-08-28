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

package plugin

import (
	"errors"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	daoNodeDeployment "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/node-deployment"
	pluginworkflow "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/plugin-workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/topoevent"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// createPluginWorkflow create plugin workflow.
func (s *Storage) createPluginWorkflow(nCtx contextx.IContext, workflow *types.PluginWorkflow) error {
	if nCtx == nil {
		return basestorage.ErrNilContent()
	}

	if workflow == nil {
		return errors.New("workflow is nil")
	}

	if err := s.daoPluginWorkflow.Create(nCtx, workflow); err != nil {
		return fmt.Errorf("failed to create plugin workflow: %w", err)
	}

	return nil
}

// getPluginWorkflow get plugin workflow.
func (s *Storage) getPluginWorkflow(nCtx contextx.IContext, workflowID string) (*types.PluginWorkflow, error) {
	if nCtx == nil {
		return nil, basestorage.ErrNilContent()
	}

	if workflowID == "" {
		return nil, errors.New("workflowID is empty")
	}

	pluginWorkflow, err := s.daoPluginWorkflow.Get(nCtx, workflowID)
	if err != nil {
		return nil, fmt.Errorf("failed to get plugin workflow: %w", err)
	}

	return pluginWorkflow, nil
}

// getPluginWorkflowStatus get plugin workflow status.
func (s *Storage) getPluginWorkflowStatus(nCtx contextx.IContext, workflowID string) (types.PluginWorkflowStatus, error) {
	if nCtx == nil {
		return "", basestorage.ErrNilContent()
	}

	if workflowID == "" {
		return "", errors.New("workflowID is empty")
	}

	status, err := s.daoPluginWorkflow.GetStatus(nCtx, workflowID)
	if err != nil {
		return "", fmt.Errorf("failed to get plugin workflow status: %w", err)
	}

	return status, nil
}

// updatePluginWorkflowStatus update plugin workflow status.
func (s *Storage) updatePluginWorkflowStatus(nCtx contextx.IContext, workflowID string, status types.PluginWorkflowStatus) error {
	if nCtx == nil {
		return basestorage.ErrNilContent()
	}

	if workflowID == "" {
		return errors.New("workflowID is empty")
	}

	if err := status.Validate(); err != nil {
		return fmt.Errorf("status is invalid: %w", err)
	}

	err := s.daoPluginWorkflow.UpdateStatus(nCtx, workflowID, status)
	if err != nil {
		return fmt.Errorf("failed to update plugin workflow status: %w", err)
	}

	return nil
}

// countPluginWorkflow count plugin workflow.
func (s *Storage) countPluginWorkflow(nCtx contextx.IContext, condition ...*types.PluginWorkflowCondition) (int64, error) {
	if nCtx == nil {
		return 0, basestorage.ErrNilContent()
	}

	opts, err := s.convertPluginWorkflowConditionsToOptions(nCtx, condition...)
	if err != nil {
		return 0, fmt.Errorf("failed to convert conditions: %w", err)
	}

	num, err := s.daoPluginWorkflow.Count(nCtx, opts...)
	if err != nil {
		return 0, fmt.Errorf("failed to count plugin workflow: %w", err)
	}

	return num, nil
}

// listPluginWorkflow list plugin workflow.
func (s *Storage) listPluginWorkflow(nCtx contextx.IContext, page types.Page, condition ...*types.PluginWorkflowCondition) (
	[]*types.PluginWorkflow, int64, error) {

	if nCtx == nil {
		return nil, 0, basestorage.ErrNilContent()
	}

	page.Sort = types.WithSortFields(page.Sort, types.WithFieldDesc(pluginworkflow.FieldKeyOperateTime))
	opts, err := s.convertPluginWorkflowConditionsToOptions(nCtx, condition...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to convert conditions: %w", err)
	}

	workflows, count, err := s.daoPluginWorkflow.List(nCtx, page, opts...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list plugin workflow: %w", err)
	}

	return workflows, count, nil
}

// distinctPluginWorkflow distincts node workflow fields.
func (s *Storage) distinctPluginWorkflow(
	nCtx contextx.IContext, request types.PluginWorkflowDistinctRequest, conditions ...*types.PluginWorkflowCondition) (
	*types.PluginWorkflowDistinctResult, error) {

	var err error

	result := new(types.PluginWorkflowDistinctResult)
	opts, err := s.convertPluginWorkflowConditionsToOptions(nCtx, conditions...)
	if err != nil {
		return nil, fmt.Errorf("failed to convert conditions: %w", err)
	}

	gp := gopool.NewPool()
	if request.HostID {
		gp.Go(func() error {
			var err error
			result.HostID, err = s.daoPluginWorkflow.DistinctPluginWorkflowBkHostID(nCtx, opts...)

			return err
		})
	}
	if request.Operator {
		gp.Go(func() error {
			var err error
			result.Operator, err = s.daoPluginWorkflow.DistinctPluginWorkflowOperator(nCtx, opts...)

			return err
		})
	}
	if request.Status {
		gp.Go(func() error {
			var err error
			result.Status, err = s.daoPluginWorkflow.DistinctPluginWorkflowStatus(nCtx, opts...)

			return err
		})
	}
	if request.Type {
		gp.Go(func() error {
			var err error
			result.Type, err = s.daoPluginWorkflow.DistinctPluginWorkflowType(nCtx, opts...)

			return err
		})
	}

	if err = gp.Wait(); err != nil {
		return nil, err
	}

	return result, nil
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

func (s *Storage) queryHostIDsByIP(
	nCtx contextx.IContext, hostInnerIPs, hostInnerIPV6s []string) ([]int64, error) {

	deployments, _, err := s.daoNodeDeployment.ListNodeDeployment(
		nCtx, types.UnlimitedPage(), buildNodeDeploymentIPOptions(hostInnerIPs, hostInnerIPV6s)...)
	if err != nil {
		return nil, fmt.Errorf("failed to list node deployment by ip: %w", err)
	}
	if len(deployments) == 0 {
		return nil, nil
	}

	hostIDSet := make(map[int64]struct{}, len(deployments))
	for _, deployment := range deployments {
		if deployment == nil || deployment.Info == nil {
			continue
		}
		if deployment.Info.Host.HostID <= 0 {
			continue
		}
		hostIDSet[deployment.Info.Host.HostID] = struct{}{}
	}

	hostIDs := make([]int64, 0, len(hostIDSet))
	for hostID := range hostIDSet {
		hostIDs = append(hostIDs, hostID)
	}

	return hostIDs, nil
}

// handleIPFilter converts IP filter to workflow filter options.
func (s *Storage) handleIPFilter(
	nCtx contextx.IContext, hostInnerIP, hostInnerIPV6 []string) ([]pluginworkflow.OptFn, error) {

	if len(hostInnerIP) == 0 && len(hostInnerIPV6) == 0 {
		return nil, nil
	}

	hostIDs, err := s.queryHostIDsByIP(nCtx, hostInnerIP, hostInnerIPV6)
	if err != nil {
		return nil, fmt.Errorf("failed to query host ids by ip: %w", err)
	}

	if len(hostIDs) == 0 {
		return []pluginworkflow.OptFn{pluginworkflow.WithHostIDs(-1)}, nil
	}

	return []pluginworkflow.OptFn{pluginworkflow.WithHostIDs(hostIDs...)}, nil
}

func (s *Storage) convertPluginWorkflowConditionsToOptions(
	nCtx contextx.IContext, conditions ...*types.PluginWorkflowCondition) ([]pluginworkflow.OptFn, error) {

	opts := make([]pluginworkflow.OptFn, 0)
	for _, condition := range conditions {
		if condition == nil {
			continue
		}

		if condition.OperateTimeRange != nil {
			opts = append(opts, topoevent.WithOperateTimeRange(*condition.OperateTimeRange))
		}

		if condition.ExactInclude != nil {
			ipOpts, err := s.handleIPFilter(nCtx, condition.ExactInclude.HostInnerIP, condition.ExactInclude.HostInnerIPV6)
			if err != nil {
				return nil, err
			}
			if ipOpts != nil {
				opts = append(opts, ipOpts...)
			}

			opts = append(opts,
				pluginworkflow.WithWorkflowID(condition.ExactInclude.WorkflowID...),
				pluginworkflow.WithHostIDs(condition.ExactInclude.HostID...),
				pluginworkflow.WithBizIDs(condition.ExactInclude.BizID...),
				pluginworkflow.WithDeployPolicyID(condition.ExactInclude.DeployPolicyIDs...),
				pluginworkflow.WithType(condition.ExactInclude.Type...),
				pluginworkflow.WithOperator(condition.ExactInclude.Operator...),
				pluginworkflow.WithStatus(condition.ExactInclude.Status...))
		}

		if condition.ExactExclude != nil {
			opts = append(opts,
				pluginworkflow.WithoutWorkflowID(condition.ExactExclude.WorkflowID...),
				pluginworkflow.WithoutHostIDs(condition.ExactExclude.HostID...),
				pluginworkflow.WithoutBizIDs(condition.ExactExclude.BizID...),
				pluginworkflow.WithoutDeployPolicyID(condition.ExactExclude.DeployPolicyIDs...),
				pluginworkflow.WithoutType(condition.ExactExclude.Type...),
				pluginworkflow.WithoutOperator(condition.ExactExclude.Operator...),
				pluginworkflow.WithoutStatus(condition.ExactExclude.Status...))
		}
	}

	return opts, nil
}
