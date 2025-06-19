/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package upload provides the upload storage interface.
package upload

import (
	"context"
	"errors"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/upload"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/identifier"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
)

// IStorage defines the interface of upload storage.
type IStorage interface {
	basestorage.Interface

	// GetUpload gets a upload by upload-id.
	GetUpload(ctx context.Context, uploadID string) (*types.Upload, error)

	// CreateUpload creates a upload.
	CreateUpload(ctx context.Context, up *types.Upload) (string, error)

	// DeleteUpload deletes a upload by upload-id.
	DeleteUpload(ctx context.Context, uploadID string) error
}

// StorageName defines the storage name.
const StorageName = "upload"

// NewStorage creates a new upload storage.
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

// Storage implements IStorage.
type Storage struct {
	basestorage.Storage

	daoUpload upload.IHandler
}

func (s *Storage) initDao() error {
	s.daoUpload = upload.New(s.Database, s.Logger)

	return nil
}

func (s *Storage) check() error {
	if s.daoUpload == nil {
		return errors.New("dao upload is nil")
	}

	return nil
}

// GetUpload gets a upload by upload-id.
func (s *Storage) GetUpload(ctx context.Context, uploadID string) (*types.Upload, error) {
	return s.daoUpload.Get(ctx, uploadID)
}

// CreateUpload creates a upload.
func (s *Storage) CreateUpload(ctx context.Context, up *types.Upload) (string, error) {
	up.UploadID = identifier.GenUploadID()

	return up.UploadID, s.daoUpload.Create(ctx, up)
}

// DeleteUpload deletes a upload by upload-id.
func (s *Storage) DeleteUpload(ctx context.Context, uploadID string) error {
	return s.daoUpload.DeleteMany(ctx, uploadID)
}
