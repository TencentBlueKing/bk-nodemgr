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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/operation"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	workoper "github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
)

// upsertOperation upsert operation.
func (s *Storage) upsertOperation(ctx contextx.IContext, operation *workoper.Operation) error {
	return s.daoOperation.Upsert(ctx, operation)
}

// listOperation lists operation.
func (s *Storage) listOperation(
	ctx contextx.IContext, page types.Page, conditions ...*types.OperationCondition) (
	[]*workoper.Operation, int64, error) {

	opts := make([]operation.OptFn, 0)
	for _, condition := range conditions {
		if condition == nil {
			continue
		}

		if condition.ExactInclude != nil {
			opts = append(opts,
				operation.WithTriggerID(condition.ExactInclude.TriggerID...),
				operation.WithOperationID(condition.ExactInclude.OperationID...),
				operation.WithEmptyOperation(condition.ExactInclude.OperInstEmpty...),
			)
		}
	}

	return s.daoOperation.List(ctx, page, opts...)
}

// deleteOperations deletes operations by operation ids.
func (s *Storage) deleteOperations(ctx contextx.IContext, operationID ...string) error {
	return s.daoOperation.Delete(ctx, operationID...)
}

// pullOperationInstanceIDs pulls operation instance IDs from operation.
func (s *Storage) pullOperationInstanceIDs(ctx contextx.IContext, operationID string, operInstIDs ...string) error {
	return s.daoOperation.PullOperInstIDs(ctx, operationID, operInstIDs...)
}
