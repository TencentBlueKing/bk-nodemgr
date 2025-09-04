/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package release provides the release storage interface.
package release

import (
	"errors"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/release"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"go.mongodb.org/mongo-driver/mongo"
)

// IStorage defines the interface of release storage.
// nolint: interfacebloat
type IStorage interface {
	basestorage.Interface

	IAgent
	IProxy
	IBinTool
	ICert
	IPluginBinTool
	IOfficialPlugin
	IExternalPlugin
}

// StorageName defines the storage name.
const StorageName = "release"

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

var _ IStorage = &Storage{}

// Storage implements IStorage.
type Storage struct {
	basestorage.Storage

	daoRelease release.IHandler
}

func (s *Storage) initDao() error {
	s.daoRelease = release.New(s.Database, s.Logger)

	return nil
}

func (s *Storage) check() error {
	if s.daoRelease == nil {
		return errors.New("dao release is nil")
	}

	return nil
}
