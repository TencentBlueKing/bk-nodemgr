/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package workflow

import (
	"errors"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/operation"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	workoper "github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
)

// getOperation gets operation by operationID.
func (s *Storage) getOperation(nCtx contextx.IContext, operationID string) (*workoper.Operation, error) {
	if nCtx == nil {
		return nil, basestorage.ErrNilContent()
	}

	opers, _, err := s.daoOperation.List(nCtx, types.UnlimitedPage(), operation.WithOperationID(operationID))
	if err != nil {
		return nil, err
	}

	if len(opers) != 1 {
		return nil, fmt.Errorf("failed to get operation, match num not 1, operation-id(%s), count(%d)",
			operationID, len(opers))
	}

	return opers[0], nil
}

// upsertOperation upsert operation.
func (s *Storage) upsertOperation(nCtx contextx.IContext, operation *workoper.Operation) error {
	if nCtx == nil {
		return basestorage.ErrNilContent()
	}

	if operation == nil {
		return basestorage.ErrUpsertNilData()
	}

	return s.daoOperation.Upsert(nCtx, operation)
}

// listOperation lists operation.
func (s *Storage) listOperationByOperationID(nCtx contextx.IContext, operationID ...string) ([]*workoper.Operation, int64, error) {
	if nCtx == nil {
		return nil, 0, basestorage.ErrNilContent()
	}

	if len(operationID) == 0 {
		return nil, 0, basestorage.ErrEmptyOperationID()
	}

	return s.daoOperation.List(nCtx, types.UnlimitedPage(), operation.WithOperationID(operationID...))
}

// listOperationWithoutParametersByParentOperationID lists operation without parameters by parent operation id.
func (s *Storage) listOperationWithoutParametersByParentOperationID(
	nCtx contextx.IContext, page types.Page, parentOperationID ...string) (
	[]*workoper.Operation, int64, error) {

	if nCtx == nil {
		return nil, 0, basestorage.ErrNilContent()
	}

	if len(parentOperationID) == 0 {
		return nil, 0, basestorage.ErrEmptyOperationID()
	}

	return s.daoOperation.ListWithoutParameters(nCtx, page, operation.WithParentOperationID(parentOperationID...))
}

// listOperationWithoutParametersByParentOperInstID lists operation without parameters by parent operation instance id.
func (s *Storage) listOperationWithoutParametersByParentOperInstID(nCtx contextx.IContext, page types.Page, parentOperInstID ...string) (
	[]*workoper.Operation, int64, error) {

	if nCtx == nil {
		return nil, 0, basestorage.ErrNilContent()
	}

	if len(parentOperInstID) == 0 {
		return nil, 0, basestorage.ErrEmptyOperationID()
	}

	return s.daoOperation.ListWithoutParameters(nCtx, page, operation.WithParentOperInstID(parentOperInstID...))
}

// deleteOperationsByTriggerID deletes operations by trigger ids.
func (s *Storage) deleteOperationsByTriggerID(nCtx contextx.IContext, triggerID ...string) error {
	if nCtx == nil {
		return basestorage.ErrNilContent()
	}

	if len(triggerID) == 0 {
		return basestorage.ErrEmptyOperationID()
	}

	return s.daoOperation.DeleteByTriggerID(nCtx, triggerID...)
}

// pullOperationInstanceIDs pulls operation instance IDs from operation.
func (s *Storage) pullOperationInstanceIDs(nCtx contextx.IContext, operationID string, operInstIDs ...string) error {
	if nCtx == nil {
		return basestorage.ErrNilContent()
	}

	if len(operationID) == 0 {
		return basestorage.ErrEmptyOperationID()
	}

	return s.daoOperation.PullOperInstIDs(nCtx, operationID, operInstIDs...)
}

// listNeedInstantiateOperationByTriggerID lists operations need to be instantiated by trigger id.
func (s *Storage) listNeedInstantiateOperationByTriggerID(nCtx contextx.IContext, page types.Page, triggerID string) (
	[]*workoper.Operation, int64, error) {

	if nCtx == nil {
		return nil, 0, basestorage.ErrNilContent()
	}

	if err := page.Validate(); err != nil {
		return nil, 0, err
	}

	if triggerID == "" {
		return nil, 0, basestorage.ErrEmptyTriggerID()
	}

	return s.daoOperation.List(nCtx, page, operation.WithTriggerID(triggerID), operation.WithInstantiated(true))
}

// existNeedInstantiateOperationByTriggerID checks whether there are operations need to be instantiated by trigger id.
func (s *Storage) existNeedInstantiateOperationByTriggerID(nCtx contextx.IContext, triggerID string) (bool, error) {
	if nCtx == nil {
		return false, basestorage.ErrNilContent()
	}

	if triggerID == "" {
		return false, basestorage.ErrEmptyTriggerID()
	}

	return s.daoOperation.Exist(nCtx, operation.WithTriggerID(triggerID), operation.WithInstantiated(true))
}

// updateLatestInstBriefData updates operation's latest instance brief data.
func (s *Storage) updateLatestInstBriefData(nCtx contextx.IContext, operationID string, briefData *workoper.InstanceBriefData) error {
	if nCtx == nil {
		return basestorage.ErrNilContent()
	}

	if operationID == "" {
		return basestorage.ErrEmptyOperationID()
	}

	if briefData == nil {
		return basestorage.ErrUpsertNilData()
	}

	return s.daoOperation.UpdateLatestInstBriefData(nCtx, operationID, briefData)
}

// getLatestOperationInstanceStatusDistributionByTriggerID gets the latest operation instance status distribution by trigger ID.
func (s *Storage) getLatestOperationInstanceStatusDistributionByTriggerID(
	nCtx contextx.IContext, triggerID ...string) (
	map[string]*workoper.InstanceStatusDistribution, error) {

	if nCtx == nil {
		return nil, basestorage.ErrNilContent()
	}

	if len(triggerID) == 0 {
		return nil, basestorage.ErrEmptyTriggerID()
	}

	distribution, err := s.daoOperation.GetLatestOperationInstanceStatusDistributionByTriggerID(nCtx, triggerID...)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest operation instance status distribution by trigger id: %w", err)
	}

	return distribution, nil
}

// listOperation lists operation by page and condition.
func (s *Storage) listOperation(nCtx contextx.IContext, page types.Page, conditions ...*types.OperationCondition) (
	[]*workoper.Operation, int64, error) {

	if nCtx == nil {
		return nil, 0, basestorage.ErrNilContent()
	}

	opts, err := convertOperationConditionsToOptions(conditions...)
	if err != nil {
		return nil, 0, err
	}

	return s.daoOperation.List(nCtx, page, opts...)
}

// countOperation counts operation by condition.
func (s *Storage) countOperation(nCtx contextx.IContext, conditions ...*types.OperationCondition) (int64, error) {
	if nCtx == nil {
		return 0, basestorage.ErrNilContent()
	}

	opts, err := convertOperationConditionsToOptions(conditions...)
	if err != nil {
		return 0, err
	}

	return s.daoOperation.Count(nCtx, opts...)
}

// distinctOperation distincts operation fields by conditions.
func (s *Storage) distinctOperation(
	nCtx contextx.IContext,
	selector types.WorkflowOperationDistinctSelector,
	conditions ...*types.OperationCondition) (*types.WorkflowOperationDistinctResult, error) {

	if nCtx == nil {
		return nil, basestorage.ErrNilContent()
	}

	opts, err := convertOperationConditionsToOptions(conditions...)
	if err != nil {
		return nil, err
	}

	result := new(types.WorkflowOperationDistinctResult)

	if selector.State {
		state, err := s.daoOperation.DistinctLatestInstState(nCtx, opts...)
		if err != nil {
			return nil, fmt.Errorf("failed to distinct latest instance state: %w", err)
		}

		result.State = state
	}

	return result, nil
}

func convertOperationConditionsToOptions(conditions ...*types.OperationCondition) ([]operation.OptFn, error) {
	opts := make([]operation.OptFn, 0)

	for _, condition := range conditions {
		if condition == nil {
			continue
		}

		if condition.ExactInclude != nil {
			opts = append(opts, operation.WithTriggerID(condition.ExactInclude.TriggerID...))
			opts = append(opts, operation.WithLatestInstState(conv.SliceToSlice(condition.ExactInclude.State, func(s workoper.State) string {
				return string(s)
			})...))
		}

		if condition.ExactExclude != nil || condition.FuzzyExclude != nil || condition.FuzzyInclude != nil {
			return nil, errors.New("exact exclude, fuzzy exclude, fuzzy include is not supported")
		}
	}

	return opts, nil
}
