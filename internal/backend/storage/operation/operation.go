/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package operation ...
// nolint: nonamedreturns
package operation

import (
	"context"
	"errors"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/operation"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	workoper "github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
	"go.mongodb.org/mongo-driver/mongo"
)

// StorageName ...
const StorageName = "operation"

// NewStorage ...
func NewStorage(client *mongo.Client, database string, logger logger.ILogger) (*Storage, error) {
	if client == nil {
		return nil, errors.New("mongo client is nil")
	}

	s := &Storage{
		Storage: basestorage.Storage{
			Name:     StorageName,
			Database: client.Database(database),
			Logger:   logger,
		},
	}
	err := basestorage.InitStorage(&s.Storage,
		basestorage.WithStartFunc(s.initDao),
		basestorage.WithCheckFunc(s.check))
	if err != nil {
		s.Logger.Errorf("new storage failed, err: %v", err)
		return nil, err
	}

	return s, nil
}

// Storage implements the Storage interface.
type Storage struct {
	basestorage.Storage

	daoOperation operation.IHandler
}

func (s *Storage) initDao() error {
	s.daoOperation = operation.New(s.Database, s.Logger)

	return nil
}

func (s *Storage) check() error {
	if s.daoOperation == nil {
		return errors.New("dao operation is nil")
	}

	return nil
}

// GetOperation get operation by operationID.
func (s *Storage) GetOperation(ctx context.Context, operationID string) (oper *workoper.Operation, err error) {
	// record metric.
	metric := s.metric().Start("get")
	defer metric.End(err)

	if ctx == nil {
		return nil, basestorage.ErrNilContent()
	}

	var opers []*workoper.Operation
	var count int64
	if opers, count, err = s.daoOperation.List(ctx, types.SingleItemPage(), operation.WithOperationID(operationID)); err != nil {
		return nil, err
	}

	if count != 1 || int64(len(opers)) != count {
		return nil, fmt.Errorf("get operation failed, match num not 1, operationID: %s, count: %d",
			operationID, count)
	}

	return opers[0], nil
}

// UpsertOperation upsert operation.
func (s *Storage) UpsertOperation(ctx context.Context, operation *workoper.Operation) (err error) {
	// record metric.
	metric := s.metric().Start("upsert")
	defer metric.End(err)

	if ctx == nil {
		return basestorage.ErrNilContent()
	}

	if operation == nil {
		return basestorage.ErrUpsertNilData()
	}

	if err = s.daoOperation.Upsert(ctx, operation); err != nil {
		return err
	}

	return nil
}

// ListOperationByTrigger lists operation by triggerid.
func (s *Storage) ListOperationByTrigger(ctx context.Context, page types.Page, triggerID ...string) (
	opers []*workoper.Operation, num int64, err error) {

	// record metric.
	metric := s.metric().Start("list_by_trigger")
	defer metric.End(err)

	if ctx == nil {
		return nil, 0, basestorage.ErrNilContent()
	}

	if err = page.Validate(); err != nil {
		return nil, 0, err
	}

	if len(triggerID) == 0 {
		return nil, 0, basestorage.ErrEmptyTriggerID()
	}

	if opers, num, err = s.daoOperation.List(ctx, page, operation.WithTriggerID(triggerID...)); err != nil {
		return nil, 0, err
	}

	return opers, num, nil
}

// ListOperation lists operation by operation id.
func (s *Storage) ListOperation(ctx context.Context, operationID ...string) (
	opers []*workoper.Operation, num int64, err error) {

	// record metric.
	metric := s.metric().Start("list")
	defer metric.End(err)

	if ctx == nil {
		return nil, 0, basestorage.ErrNilContent()
	}

	if len(operationID) == 0 {
		return nil, 0, basestorage.ErrEmptyOperationID()
	}

	if opers, num, err = s.daoOperation.List(ctx, types.UnlimitedPage(), operation.WithOperationID(operationID...)); err != nil {
		return nil, 0, err
	}

	return opers, num, nil
}

// ListOperationByCondition lists operation by condition.
func (s *Storage) ListOperationByCondition(ctx context.Context, page types.Page,
	condition ...*types.NodeWorkflowOperationCondition) (opers []*workoper.Operation, num int64, err error) {

	// record metric.
	metric := s.metric().Start("list_by_condition")
	defer metric.End(err)

	opts := make([]operation.OptFn, 0)
	for _, condition := range condition {
		if condition == nil {
			continue
		}

		if condition.ExactInclude != nil {
			opts = append(opts,
				operation.WithTriggerID(condition.ExactInclude.TriggerID),
				operation.WithBizID(condition.ExactInclude.BizID...),
				operation.WithNetworkAreaID(condition.ExactInclude.NetworkAreaID...),
				operation.WithIPv4(condition.ExactInclude.InnerIP...),
				operation.WithIPv6(condition.ExactInclude.InnerIPv6...),
			)
		}
	}

	if opers, num, err = s.daoOperation.List(ctx, page, opts...); err != nil {
		return nil, 0, err
	}

	return opers, num, nil
}

// ListEmptyOperation lists empty operation by triggerid.
func (s *Storage) ListEmptyOperation(
	ctx context.Context, page types.Page, triggerID string) (opers []*workoper.Operation, num int64, err error) {

	// record metric.
	metric := s.metric().Start("list_empty")
	defer metric.End(err)

	if ctx == nil {
		return nil, 0, basestorage.ErrNilContent()
	}

	if err := page.Validate(); err != nil {
		return nil, 0, err
	}

	if triggerID == "" {
		return nil, 0, basestorage.ErrEmptyTriggerID()
	}

	if opers, num, err = s.daoOperation.List(ctx, page, operation.WithTriggerID(triggerID), operation.WithEmptyOperation()); err != nil {
		return nil, 0, err
	}

	return opers, num, nil
}

// DeleteOperations deletes operations by operation ids.
func (s *Storage) DeleteOperations(ctx context.Context, operationID ...string) (err error) {
	// record metric.
	metric := s.metric().Start("delete")
	defer metric.End(err)

	if ctx == nil {
		return basestorage.ErrNilContent()
	}

	if len(operationID) == 0 {
		return basestorage.ErrEmptyOperationID()
	}

	if err = s.daoOperation.Delete(ctx, operationID...); err != nil {
		return err
	}

	return nil
}

// PullOperationInstanceIDs pulls operation instance IDs from operation.
func (s *Storage) PullOperationInstanceIDs(ctx context.Context, operationID string, operInstIDs ...string) (err error) {
	// record metric.
	metric := s.metric().Start("pull_instance_ids")
	defer metric.End(err)

	if ctx == nil {
		return basestorage.ErrNilContent()
	}

	if len(operationID) == 0 {
		return basestorage.ErrEmptyOperationID()
	}

	if err = s.daoOperation.PullOperInstIDs(ctx, operationID, operInstIDs...); err != nil {
		return err
	}

	return nil
}

func (s *Storage) metric() *storage.MetricData {
	return storage.Metric(StorageName)
}
