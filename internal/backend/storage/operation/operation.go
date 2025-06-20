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
package operation

import (
	"context"
	"errors"
	"fmt"

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
func NewStorage(client *mongo.Client, database string, logger logger.Logger) (*Storage, error) {
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
func (s *Storage) GetOperation(ctx context.Context, operationID string) (*workoper.Operation, error) {
	if ctx == nil {
		return nil, basestorage.ErrNilContent()
	}

	operations, count, err := s.daoOperation.List(ctx, types.SingleItemPage(), operation.WithOperationID(operationID))
	if err != nil {
		return nil, err
	}

	if count != 1 || int64(len(operations)) != count {
		return nil, fmt.Errorf("get operation failed, match num not 1, operationID: %s, count: %d",
			operationID, count)
	}

	return operations[0], nil
}

// UpsertOperation upsert operation.
func (s *Storage) UpsertOperation(ctx context.Context, operation *workoper.Operation) error {
	if ctx == nil {
		return basestorage.ErrNilContent()
	}

	if operation == nil {
		return basestorage.ErrUpsertNilData()
	}

	return s.daoOperation.Upsert(ctx, operation)
}

// ListOperationByTrigger lists operation by triggerid.
func (s *Storage) ListOperationByTrigger(ctx context.Context, page types.Page, triggerID ...string) (
	[]*workoper.Operation, int64, error) {

	if ctx == nil {
		return nil, 0, basestorage.ErrNilContent()
	}

	if err := page.Validate(); err != nil {
		return nil, 0, err
	}

	if len(triggerID) == 0 {
		return nil, 0, basestorage.ErrEmptyTriggerID()
	}

	return s.daoOperation.List(ctx, page, operation.WithTriggerID(triggerID...))
}

// ListEmptyOperation lists empty operation by triggerid.
func (s *Storage) ListEmptyOperation(
	ctx context.Context, page types.Page, triggerID string) ([]*workoper.Operation, int64, error) {

	if ctx == nil {
		return nil, 0, basestorage.ErrNilContent()
	}

	if err := page.Validate(); err != nil {
		return nil, 0, err
	}

	if triggerID == "" {
		return nil, 0, basestorage.ErrEmptyTriggerID()
	}

	return s.daoOperation.List(ctx, page, operation.WithTriggerID(triggerID), operation.WithEmptyOperation())
}
