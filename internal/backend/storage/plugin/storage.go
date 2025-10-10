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
// nolint: nonamedreturns
package plugin

import (
	"errors"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	plugindeployment "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/plugin-deployment"
	pluginworkflow "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/plugin-workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
)

// StorageName the name of storage.
const StorageName = "plugin"

// NewStorage ...
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

// Storage this is a storage to operate node deployment table.
type Storage struct {
	basestorage.Storage

	// dao
	daoPluginDeployment plugindeployment.IHandler
	daoPluginWorkflow   pluginworkflow.IHandler
}

func (s *Storage) initDao() error {
	s.daoPluginDeployment = plugindeployment.New(s.Database)
	s.daoPluginWorkflow = pluginworkflow.New(s.Database)

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

// GetPluginDeploymentInfo get plugin deployment info.
func (s *Storage) GetPluginDeploymentInfo(nCtx contextx.IContext, token string) (*types.PluginDeploymentInfo, error) {
	var (
		info *types.PluginDeploymentInfo
		err  error
	)

	// record metric.
	metric := s.metric().Start("get_info")
	defer metric.End(err)

	info, err = s.getPluginDeploymentInfo(nCtx, token)

	return info, err
}

// CreatePluginDeployment plugin deployment.
func (s *Storage) CreatePluginDeployment(nCtx contextx.IContext, pluginDeployment *types.PluginDeployment) error {
	var (
		err error
	)

	// record metric.
	metric := s.metric().Start("createPluginDeployment")
	defer metric.End(err)

	err = s.createPluginDeployment(nCtx, pluginDeployment)

	return err
}

// UpdatePluginDeploymentInfo update a node deployment info.
func (s *Storage) UpdatePluginDeploymentInfo(nCtx contextx.IContext, token string, pluginDeploymentInfo *types.PluginDeploymentInfo) error {
	var (
		err error
	)

	// record metric.
	metric := s.metric().Start("update_info")
	defer metric.End(err)

	err = s.updatePluginDeploymentInfo(nCtx, token, pluginDeploymentInfo)

	return err
}

// GetPluginDeploymentMainConfig

// GetPluginWorkflow get plugin workflow.
func (s *Storage) GetPluginWorkflow(nCtx contextx.IContext, workflowID string) (*types.PluginWorkflow, error) {
	var (
		pluginWorkflow *types.PluginWorkflow
		err            error
	)

	// record metric.
	metric := s.metric().Start("create_plugin_workflow")
	defer metric.End(err)

	pluginWorkflow, err = s.getPluginWorkflow(nCtx, workflowID)

	return pluginWorkflow, err
}

// CreatePluginWorkflow createPluginDeployment plugin workflow.
func (s *Storage) CreatePluginWorkflow(nCtx contextx.IContext, workflow *types.PluginWorkflow) error {
	var (
		err error
	)

	// record metric.
	metric := s.metric().Start("create_plugin_workflow")
	defer metric.End(err)

	err = s.createPluginWorkflow(nCtx, workflow)

	return err
}

// UpdatePluginWorkflowStatus update plugin workflow status.
func (s *Storage) UpdatePluginWorkflowStatus(nCtx contextx.IContext, workflowID string, status types.PluginWorkflowStatus) error {
	var (
		err error
	)

	// record metric.
	metric := s.metric().Start("update_plugin_workflow_status")
	defer metric.End(err)

	err = s.updatePluginWorkflowStatus(nCtx, workflowID, status)

	return err
}

// GetPluginDeploymentMainConfig get plugin deployment main config.
func (s *Storage) GetPluginDeploymentMainConfig(ctx contextx.IContext, token string) (mainConfig []byte, err error) {
	// record metric.
	metric := s.metric().Start("get_plugin_deployment_main_config")
	defer metric.End(err)

	mainConfig, err = s.getPluginDeploymentMainConfig(ctx, token)

	return mainConfig, err
}

// UpdatePluginDeploymentMainConfig set plugin deployment main config.
func (s *Storage) UpdatePluginDeploymentMainConfig(ctx contextx.IContext, token string, mainConfig []byte) (err error) {
	// record metric.
	metric := s.metric().Start("update_plugin_deployment_main_config")
	defer metric.End(err)

	err = s.updatePluginDeploymentMainConfig(ctx, token, mainConfig)

	return err
}
