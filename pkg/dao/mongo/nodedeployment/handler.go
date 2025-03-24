/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package nodedeployment

import (
	"context"
	"errors"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
)

// IHandler node deployment Handler interface.
type IHandler interface {
	Create(ctx context.Context, nodeDeployment *types.NodeDeployment) error
	GetInfo(ctx context.Context, Token string) (*types.DeploymentInfo, error)
	GetNodeConf(ctx context.Context, Token string) (*types.NodeConf, error)
}

// Handler this is a handler to operate node deployment table.
type Handler struct {
	dao    *dao
	logger logger.Logger
}

// New new a handler.
func New(client *mongo.Database, logger logger.Logger) *Handler {
	h := &Handler{
		dao:    newDao(client, logger),
		logger: logger,
	}

	if err := h.dao.ensureIndexes(); err != nil {
		h.logger.Warnf("failed to ensure nodedeloyment indexes, err: %v",
			errors.Join(base.ErrEnsureIndexesFailed(), err))
	}

	return h
}

// GetInfo get a node deployment info.
func (h *Handler) GetInfo(ctx context.Context, Token string) (*types.DeploymentInfo, error) {
	if ctx == nil {
		return nil, base.ErrInvalidContext()
	}

	if Token == "" {
		return nil, base.ErrInvalidID()
	}

	filter := base.AliveFilter()
	opt := base.WithStringValues(FieldKeyToken, Token)
	filter = opt(filter)

	field := fmt.Sprintf(FieldKeyInfo)
	data, err := h.dao.get(ctx, filter, field)
	if err != nil {
		return nil, base.ErrRecordNoFound()
	}

	return convertDeploymentInfoToTypes(data.Info)
}

func convertDeploymentInfoToTypes(info *Info) (*types.DeploymentInfo, error) {
	if info == nil {
		return nil, errors.New("info is nil")
	}

	return &types.DeploymentInfo{
		OperInstID:     info.OperInstID,
		ActionName:     info.ActionName,
		HostID:         info.HostID,
		TenantID:       info.TenantID,
		NodeRole:       types.NodeRole(info.NodeRole),
		NodeStatus:     types.NodeStatus(info.NodeStatus),
		NodeVersion:    info.NodeVersion,
		NodeGeneration: info.NodeGeneration,
		AgentID:        info.AgentID,
		NetworkUnitID:  info.NetworkUnitID,
	}, nil
}

// Create create a new node deployment.
func (h *Handler) Create(ctx context.Context, nodeDeployment *types.NodeDeployment) error {
	if ctx == nil {
		return base.ErrInvalidContext()
	}

	if nodeDeployment == nil {
		return base.ErrEmptyParamData()
	}

	data, err := convertNodeDeploymentFromTypes(nodeDeployment)
	if err != nil {
		return err
	}

	return h.dao.create(ctx, data)
}

func convertNodeDeploymentFromTypes(data *types.NodeDeployment) (*NodeDeployment, error) {
	nodeDeployment := &NodeDeployment{
		Token: data.Token,
		Info: &Info{
			OperInstID:     data.Info.OperInstID,
			ActionName:     data.Info.ActionName,
			HostID:         data.Info.HostID,
			TenantID:       data.Info.TenantID,
			NodeRole:       string(data.Info.NodeRole),
			NodeStatus:     string(data.Info.NodeStatus),
			NodeVersion:    data.Info.NodeVersion,
			NodeGeneration: data.Info.NodeGeneration,
			AgentID:        data.Info.AgentID,
			NetworkUnitID:  data.Info.NetworkUnitID,
		},
		NodeConf: new(NodeConf),
	}

	var err error
	nodeDeployment.NodeConf, err = convertNodeConfFromTypes(data.NodeConf)
	if err != nil {
		return nil, err
	}

	return nodeDeployment, nil
}

func convertNodeConfFromTypes(nodeConf *types.NodeConf) (*NodeConf, error) {
	if nodeConf == nil {
		return nil, errors.New("node conf is nil")
	}

	return &NodeConf{
		PreSetting:    nodeConf.PreSetting,
		CustomSetting: nodeConf.CustomSetting,
	}, nil
}

// GetNodeConf get a node deployment node conf.
func (h *Handler) GetNodeConf(ctx context.Context, Token string) (*types.NodeConf, error) {
	if ctx == nil {
		return nil, base.ErrInvalidContext()
	}

	if Token == "" {
		return nil, base.ErrInvalidID()
	}

	filter := base.AliveFilter()
	opt := base.WithStringValues(FieldKeyToken, Token)
	filter = opt(filter)

	field := fmt.Sprintf(FieldKeyNodeConf)
	data, err := h.dao.get(ctx, filter, field)
	if err != nil {
		return nil, base.ErrRecordNoFound()
	}

	return convertNodeConfToTypes(data.NodeConf)
}

func convertNodeConfToTypes(nodeConf *NodeConf) (*types.NodeConf, error) {
	if nodeConf == nil {
		return nil, errors.New("node conf is nil")
	}

	return &types.NodeConf{
		PreSetting:    nodeConf.PreSetting,
		CustomSetting: nodeConf.CustomSetting,
	}, nil
}
