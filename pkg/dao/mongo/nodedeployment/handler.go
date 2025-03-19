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
	GetCheckList(ctx context.Context, Token string) (*types.NodeConf, error)
	GetAgentConf(ctx context.Context, Token string) (*types.NodeConf, error)
	GetDataProxyConf(ctx context.Context, Token string) (*types.NodeConf, error)
	GetFileProxyConf(ctx context.Context, Token string) (*types.NodeConf, error)
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
		OperInstID: info.OperInstID,
		ActionName: info.ActionName,
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
			OperInstID: data.Info.OperInstID,
			ActionName: data.Info.ActionName,
		},
		CheckList:     new(NodeConf),
		AgentConf:     new(NodeConf),
		DataProxyConf: new(NodeConf),
		FileProxyConf: new(NodeConf),
	}

	var err error
	nodeDeployment.CheckList, err = convertNodeConfFromTypes(data.CheckList)
	if err != nil {
		return nil, err
	}

	nodeDeployment.AgentConf, err = convertNodeConfFromTypes(data.AgentConf)
	if err != nil {
		return nil, err
	}

	nodeDeployment.DataProxyConf, err = convertNodeConfFromTypes(data.DataProxyConf)
	if err != nil {
		return nil, err
	}

	nodeDeployment.FileProxyConf, err = convertNodeConfFromTypes(data.FileProxyConf)
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

// GetCheckList get a node deployment check list.
func (h *Handler) GetCheckList(ctx context.Context, Token string) (*types.NodeConf, error) {
	if ctx == nil {
		return nil, base.ErrInvalidContext()
	}

	if Token == "" {
		return nil, base.ErrInvalidID()
	}

	filter := base.AliveFilter()
	opt := base.WithStringValues(FieldKeyToken, Token)
	filter = opt(filter)

	field := fmt.Sprintf(FieldKeyCheckList)
	data, err := h.dao.get(ctx, filter, field)
	if err != nil {
		return nil, base.ErrRecordNoFound()
	}

	return convertNodeConfToTypes(data.CheckList)
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

// GetAgentConf get a node deployment agent conf.
func (h *Handler) GetAgentConf(ctx context.Context, Token string) (*types.NodeConf, error) {
	if ctx == nil {
		return nil, base.ErrInvalidContext()
	}

	if Token == "" {
		return nil, base.ErrInvalidID()
	}

	filter := base.AliveFilter()
	opt := base.WithStringValues(FieldKeyToken, Token)
	filter = opt(filter)

	field := fmt.Sprintf(FieldKeyAgentConf)
	data, err := h.dao.get(ctx, filter, field)
	if err != nil {
		return nil, base.ErrRecordNoFound()
	}

	return convertNodeConfToTypes(data.AgentConf)
}

// GetDataProxyConf get a node deployment data proxy conf.
func (h *Handler) GetDataProxyConf(ctx context.Context, Token string) (*types.NodeConf, error) {
	if ctx == nil {
		return nil, base.ErrInvalidContext()
	}

	if Token == "" {
		return nil, base.ErrInvalidID()
	}

	filter := base.AliveFilter()
	opt := base.WithStringValues(FieldKeyToken, Token)
	filter = opt(filter)

	field := fmt.Sprintf(FieldKeyDataProxyConf)
	data, err := h.dao.get(ctx, filter, field)
	if err != nil {
		return nil, base.ErrRecordNoFound()
	}

	return convertNodeConfToTypes(data.DataProxyConf)
}

// GetFileProxyConf get a node deployment file proxy conf.
func (h *Handler) GetFileProxyConf(ctx context.Context, Token string) (*types.NodeConf, error) {
	if ctx == nil {
		return nil, base.ErrInvalidContext()
	}

	if Token == "" {
		return nil, base.ErrInvalidID()
	}

	filter := base.AliveFilter()
	opt := base.WithStringValues(FieldKeyToken, Token)
	filter = opt(filter)

	field := fmt.Sprintf(FieldKeyFileProxyConf)
	data, err := h.dao.get(ctx, filter, field)
	if err != nil {
		return nil, base.ErrRecordNoFound()
	}

	return convertNodeConfToTypes(data.FileProxyConf)
}
