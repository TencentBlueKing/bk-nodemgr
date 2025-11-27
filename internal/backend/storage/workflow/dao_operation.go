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
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/operation"
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

// listOperationByTriggerID lists operation by triggerID.
func (s *Storage) listOperationByTriggerID(nCtx contextx.IContext, page types.Page, triggerID ...string) ([]*workoper.Operation, int64, error) {
	if nCtx == nil {
		return nil, 0, basestorage.ErrNilContent()
	}

	if len(triggerID) == 0 {
		return nil, 0, basestorage.ErrEmptyOperationID()
	}

	return s.daoOperation.List(nCtx, page, operation.WithTriggerID(triggerID...))
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

// listOperationByParentOperationID lists operation by parent operation id.
func (s *Storage) listOperationByParentOperationID(
	nCtx contextx.IContext, page types.Page, parentOperationID ...string) (
	[]*workoper.Operation, int64, error) {

	if nCtx == nil {
		return nil, 0, basestorage.ErrNilContent()
	}

	if len(parentOperationID) == 0 {
		return nil, 0, basestorage.ErrEmptyOperationID()
	}

	return s.daoOperation.List(nCtx, page, operation.WithParentOperationID(parentOperationID...))
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
