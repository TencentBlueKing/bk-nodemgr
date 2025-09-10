/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package cptemplate provides the config policy template storage interface.
// nolint: nonamedreturns
package cptemplate

import (
	"context"
	"errors"

	"github.com/TencentBlueKing/bk-nodemgr/internal/file/storage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/cptemplate"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
)

// IStorage defines the config policy template storage interface.
type IStorage interface {
	basestorage.Interface

	// GetConfigPolicyTemplate gets the config policy template.
	GetConfigPolicyTemplate(ctx context.Context, configPolicyID int64) (*types.ConfigPolicyTemplate, error)

	// UpsertManyConfigPolicyTemplate upserts the config policy template.
	UpsertManyConfigPolicyTemplate(ctx context.Context, configPolicyTemplates ...*types.ConfigPolicyTemplate) error

	// DeleteManyConfigPolicyTemplate deletes the config policy template.
	DeleteManyConfigPolicyTemplate(ctx context.Context, configPolicyID ...int64) error
}

// StorageName defines the storage name.
const StorageName = "cptemplate"

// NewStorage creates a new release storage.
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

// Storage implements IStorage.
type Storage struct {
	basestorage.Storage

	daoConfigPolicyTemplate cptemplate.IHandler
}

func (s *Storage) initDao() error {
	s.daoConfigPolicyTemplate = cptemplate.New(s.Database, s.Logger)

	return nil
}

func (s *Storage) check() error {
	if s.daoConfigPolicyTemplate == nil {
		return errors.New("dao cptemplate is nil")
	}

	return nil
}

// GetConfigPolicyTemplate gets the config policy template.
func (s *Storage) GetConfigPolicyTemplate(ctx context.Context, configPolicyID int64) (data *types.ConfigPolicyTemplate, err error) {
	// record metric.
	metric := s.metric().Start("get")
	defer metric.End(err)

	return s.daoConfigPolicyTemplate.Get(ctx, configPolicyID)
}

// UpsertManyConfigPolicyTemplate upserts the config policy template.
func (s *Storage) UpsertManyConfigPolicyTemplate(ctx context.Context, configPolicyTemplates ...*types.ConfigPolicyTemplate) (err error) {
	// record metric.
	metric := s.metric().Start("upsert_many")
	defer metric.End(err)

	return s.daoConfigPolicyTemplate.UpsertMany(ctx, configPolicyTemplates...)
}

// DeleteManyConfigPolicyTemplate deletes the config policy template.
func (s *Storage) DeleteManyConfigPolicyTemplate(ctx context.Context, configPolicyID ...int64) (err error) {
	// record metric.
	metric := s.metric().Start("delete_many")
	defer metric.End(err)

	return s.daoConfigPolicyTemplate.DeleteMany(ctx, configPolicyID...)
}

func (s *Storage) metric() *storage.MetricData {
	return storage.Metric(StorageName)
}
