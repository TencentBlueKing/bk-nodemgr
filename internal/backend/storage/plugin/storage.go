/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package plugin provide plugin storage.
package plugin

import (
	"errors"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/plugin-deployment"
	pluginworkflow "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/plugin-workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
)

// StorageName the name of storage.
const StorageName = "plugin"

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

var _ IStorage = &Storage{}

// Storage this is a storage to operate node deployment table.
type Storage struct {
	basestorage.Storage

	// dao
	daoPluginDeployment plugindeployment.IHandler
	daoPluginWorkflow   pluginworkflow.IHandler
}

func (s *Storage) initDao() error {
	s.daoPluginDeployment = plugindeployment.New(s.Database, s.Logger)
	s.daoPluginWorkflow = pluginworkflow.New(s.Database, s.Logger)

	return nil
}

func (s *Storage) check() error {
	if s.daoPluginDeployment == nil {
		return errors.New("dao plugin deployment is nil")
	}

	if s.daoPluginWorkflow == nil {
		return errors.New("dao plugin workflow is nil")
	}

	return nil
}

func (s *Storage) metric() *storage.MetricData {
	return storage.Metric(StorageName)
}

// GetInfo get plugin deployment info.
func (s *Storage) GetInfo(ctx contextx.IContext, token string) (*types.PluginDeploymentInfo, error) {
	var (
		info *types.PluginDeploymentInfo
		err  error
	)

	// record metric.
	metric := s.metric().Start("get_info")
	defer metric.End(err)

	info, err = s.getInfo(ctx, token)

	return info, err
}

// Create plugin deployment.
func (s *Storage) Create(ctx contextx.IContext, pluginDeployment *types.PluginDeployment) error {
	var (
		err error
	)

	// record metric.
	metric := s.metric().Start("create")
	defer metric.End(err)

	err = s.create(ctx, pluginDeployment)

	return err
}

// UpdatePluginDeploymentInfo update a node deployment info.
func (s *Storage) UpdatePluginDeploymentInfo(ctx contextx.IContext, token string, pluginDeploymentInfo *types.PluginDeploymentInfo) error {
	var (
		err error
	)

	// record metric.
	metric := s.metric().Start("update_info")
	defer metric.End(err)

	err = s.updateInfo(ctx, token, pluginDeploymentInfo)

	return err
}

// GetPluginWorkflow get plugin workflow.
func (s *Storage) GetPluginWorkflow(ctx contextx.ITenantContext, workflowID string) (*types.PluginWorkflow, error) {
	var (
		pluginWorkflow *types.PluginWorkflow
		err            error
	)

	// record metric.
	metric := s.metric().Start("create_plugin_workflow")
	defer metric.End(err)

	pluginWorkflow, err = s.getPluginWorkflow(ctx, workflowID)

	return pluginWorkflow, err
}

// CreatePluginWorkflow create plugin workflow.
func (s *Storage) CreatePluginWorkflow(ctx contextx.ITenantContext, workflow *types.PluginWorkflow) error {
	var (
		err error
	)

	// record metric.
	metric := s.metric().Start("create_plugin_workflow")
	defer metric.End(err)

	err = s.createPluginWorkflow(ctx, workflow)

	return err
}

// UpdatePluginWorkflowStatus update plugin workflow status.
func (s *Storage) UpdatePluginWorkflowStatus(ctx contextx.ITenantContext, workflowID string, status types.PluginWorkflowStatus) error {
	var (
		err error
	)

	// record metric.
	metric := s.metric().Start("update_plugin_workflow_status")
	defer metric.End(err)

	err = s.updatePluginWorkflowStatus(ctx, workflowID, status)

	return err
}
