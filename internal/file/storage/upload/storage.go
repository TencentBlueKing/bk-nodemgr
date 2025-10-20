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
	"errors"

	"github.com/TencentBlueKing/bk-nodemgr/internal/file/storage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/upload"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"go.mongodb.org/mongo-driver/mongo"
)

// IStorage defines the interface of upload storage.
type IStorage interface {
	basestorage.Interface

	IAgent
	IServer
	IBinTool
	ICert
	IPluginBinTool
	IPlugin
	IExternalPlugin
}

// StorageName defines the storage name.
const StorageName = "upload"

// NewStorage creates a new upload storage.
func NewStorage(client *mongo.Client, database string) (*Storage, error) {
	if client == nil {
		return nil, errors.New("mongo client is nil")
	}

	s := &Storage{
		Storage: basestorage.Storage{
			Name:     StorageName,
			Database: client.Database(database),
		},
	}
	err := basestorage.InitStorage(&s.Storage,
		basestorage.WithStartFunc(s.initDao),
		basestorage.WithCheckFunc(s.check))
	if err != nil {
		logger.G.Sys().WithErr(err).Error("failed to new storage")

		return nil, err
	}

	return s, nil
}

var _ IStorage = &Storage{}

// Storage implements IStorage.
type Storage struct {
	basestorage.Storage

	daoUpload upload.IHandler
}

func (s *Storage) initDao() error {
	s.daoUpload = upload.New(s.Database)

	return nil
}

func (s *Storage) check() error {
	if s.daoUpload == nil {
		return errors.New("dao upload is nil")
	}

	return nil
}

func (s *Storage) metric() *storage.MetricData {
	return storage.Metric(StorageName)
}
