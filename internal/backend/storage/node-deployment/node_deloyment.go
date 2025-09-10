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
// nolint: nonamedreturns
package nodedeployment

import (
	"context"
	"errors"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/nodedeployment"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
)

// StorageName the name of storage.
const StorageName = "nodedeployment"

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

// Storage this is a storage to operate node deployment table.
type Storage struct {
	basestorage.Storage

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

// GetNodeConf get gse node conf.
func (s *Storage) GetNodeConf(ctx context.Context, token string) (nodeConf *types.NodeConf, err error) {
	// record metric.
	metric := s.metric().Start("get_node_conf")
	defer metric.End(err)

	if ctx == nil {
		return nil, basestorage.ErrNilContent()
	}

	if token == "" {
		return nil, basestorage.ErrEmptyUniqueKey()
	}

	if nodeConf, err = s.nodeDeploymentDao.GetNodeConf(ctx, token); err != nil {
		return nil, err
	}

	return nodeConf, nil
}

// GetInfo get node deployment info.
func (s *Storage) GetInfo(ctx context.Context, token string) (deployInfo *types.DeploymentInfo, err error) {
	// record metric.
	metric := s.metric().Start("get_deploy_info")
	defer metric.End(err)

	if ctx == nil {
		return nil, basestorage.ErrNilContent()
	}

	if token == "" {
		return nil, basestorage.ErrEmptyUniqueKey()
	}

	if deployInfo, err = s.nodeDeploymentDao.GetInfo(ctx, token); err != nil {
		return nil, err
	}

	return deployInfo, nil
}

// SetNodeConf set gse node conf.
func (s *Storage) SetNodeConf(ctx context.Context, token string, conf *types.NodeConf) (err error) {
	// record metric.
	metric := s.metric().Start("set_node_conf")
	defer metric.End(err)

	if ctx == nil {
		return basestorage.ErrNilContent()
	}

	if token == "" {
		return basestorage.ErrEmptyUniqueKey()
	}

	if conf == nil {
		return errors.New("node conf is nil")
	}

	if err = s.nodeDeploymentDao.SetNodeConf(ctx, token, conf); err != nil {
		return err
	}

	return nil
}

// UpdateInfo update node deployment info.
func (s *Storage) UpdateInfo(ctx context.Context, token string, info *types.DeploymentInfo) (err error) {
	// record metric.
	metric := s.metric().Start("update_deploy_info")
	defer metric.End(err)

	if ctx == nil {
		return basestorage.ErrNilContent()
	}

	if token == "" {
		return basestorage.ErrEmptyUniqueKey()
	}

	if info == nil {
		return errors.New("node deployment info is nil")
	}

	if err = s.nodeDeploymentDao.UpdateInfo(ctx, token, info); err != nil {
		return err
	}

	return nil
}

// Create create a node deployment.
func (s *Storage) Create(ctx context.Context, nodeDeployment *types.NodeDeployment) (err error) {
	// record metric.
	metric := s.metric().Start("create")
	defer metric.End(err)

	if ctx == nil {
		return basestorage.ErrNilContent()
	}

	if nodeDeployment == nil {
		return errors.New("node deployment is nil")
	}

	if err = s.nodeDeploymentDao.Create(ctx, nodeDeployment); err != nil {
		return err
	}

	return nil
}

func (s *Storage) metric() *storage.MetricData {
	return storage.Metric(StorageName)
}
