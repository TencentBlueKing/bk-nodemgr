/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package nodedeployment ...
package nodedeployment

import (
	"context"
	"errors"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/nodedeployment"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"go.mongodb.org/mongo-driver/mongo"
)

// StorageName the name of storage.
const StorageName = "nodedeployment"

// NewStorage ...
func NewStorage(client *mongo.Client, database string, logger logger.Logger) (*Storage, error) {
	if client == nil {
		return nil, errors.New("mongo client is nil")
	}

	s := &Storage{
		Storage: base.Storage{
			Name:     StorageName,
			Database: client.Database(database),
			Logger:   logger,
		},
	}
	err := base.InitStorage(&s.Storage,
		base.WithStartFunc(s.initDao),
		base.WithCheckFunc(s.check))
	if err != nil {
		s.Logger.Errorf("new storage failed, err: %v", err)
		return nil, err
	}

	return s, nil
}

// Storage this is the storage of nodedeployment.
type Storage struct {
	base.Storage

	// dao
	nodeDeploymentDao nodedeployment.IHandler
}

func (s *Storage) initDao() error {
	s.nodeDeploymentDao = nodedeployment.New(s.Database, s.Logger)

	return nil
}

func (s *Storage) check() error {
	if s.nodeDeploymentDao == nil {
		return errors.New("nodedeployment dao is nil")
	}

	return nil
}

// GetGseAgentSetting get gse agent setting.
func (s *Storage) GetGseAgentSetting(ctx context.Context, token string) (map[string]any, map[string]any, error) {
	if ctx == nil {
		return nil, nil, base.ErrNilContent()
	}

	if token == "" {
		return nil, nil, base.ErrEmptyUniqueKey()
	}

	agentConf, err := s.nodeDeploymentDao.GetAgentConf(ctx, token)
	if err != nil {
		return nil, nil, err
	}

	return agentConf.PreSetting, agentConf.CustomSetting, nil
}

// GetGseFileProxySetting get gse file proxy setting.
func (s *Storage) GetGseFileProxySetting(ctx context.Context, token string) (map[string]any, map[string]any, error) {
	if ctx == nil {
		return nil, nil, base.ErrNilContent()
	}

	if token == "" {
		return nil, nil, base.ErrEmptyUniqueKey()
	}

	fileProxyConf, err := s.nodeDeploymentDao.GetFileProxyConf(ctx, token)
	if err != nil {
		return nil, nil, err
	}

	return fileProxyConf.PreSetting, fileProxyConf.CustomSetting, nil
}

// GetGseDataProxySetting get gse data proxy setting.
func (s *Storage) GetGseDataProxySetting(ctx context.Context, token string) (map[string]any, map[string]any, error) {
	if ctx == nil {
		return nil, nil, base.ErrNilContent()
	}

	if token == "" {
		return nil, nil, base.ErrEmptyUniqueKey()
	}

	dataProxyConf, err := s.nodeDeploymentDao.GetDataProxyConf(ctx, token)
	if err != nil {
		return nil, nil, err
	}

	return dataProxyConf.PreSetting, dataProxyConf.CustomSetting, nil
}

// GetCheckListSetting get check list setting.
func (s *Storage) GetCheckListSetting(ctx context.Context, token string) (map[string]any, map[string]any, error) {
	if ctx == nil {
		return nil, nil, base.ErrNilContent()
	}

	if token == "" {
		return nil, nil, base.ErrEmptyUniqueKey()
	}

	checkList, err := s.nodeDeploymentDao.GetCheckList(ctx, token)
	if err != nil {
		return nil, nil, err
	}

	return checkList.PreSetting, checkList.CustomSetting, nil
}
