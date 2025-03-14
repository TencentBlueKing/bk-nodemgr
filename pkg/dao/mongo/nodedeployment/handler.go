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
	GetPreSetting(ctx context.Context, Token string) (*types.NodePreSetting, error)
}

// Handler this is a handler to operate node deployment table.
type Handler struct {
	dao *dao
}

// New new a handler.
func New(client *mongo.Database, logger logger.Logger) *Handler {
	return &Handler{
		dao: newDao(client, logger),
	}
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
	opt := base.WithStringValues("data.token", Token)
	filter = opt(filter)

	field := fmt.Sprintf("data.deployment_info")
	data, err := h.dao.get(ctx, filter, field)
	if err != nil {
		return nil, base.ErrRecordNoFound()
	}

	return convertDeploymentInfoToTypes(data.DeploymentInfo)
}

// GetPreSetting get the node deployment presetting.
func (h *Handler) GetPreSetting(ctx context.Context, Token string) (*types.NodePreSetting, error) {
	if ctx == nil {
		return nil, base.ErrInvalidContext()
	}

	if Token == "" {
		return nil, base.ErrInvalidID()
	}

	filter := base.AliveFilter()
	opt := base.WithStringValues("data.token", Token)
	filter = opt(filter)

	field := fmt.Sprintf("data.pre_setting")
	data, err := h.dao.get(ctx, filter, field)
	if err != nil {
		return nil, base.ErrRecordNoFound()
	}

	return convertPreSettingToTypes(data.PreSetting)
}

func convertDeploymentInfoToTypes(info *DeploymentInfo) (*types.DeploymentInfo, error) {
	if info == nil {
		return nil, errors.New("info is nil")
	}

	return &types.DeploymentInfo{
		OperInstID: info.OperInstID,
		ActionName: info.ActionName,
	}, nil
}

func convertPreSettingToTypes(data *PreSetting) (*types.NodePreSetting, error) {
	if data == nil {
		return nil, errors.New("pre setting is empty")
	}

	preSetting := &types.NodePreSetting{
		CheckList:     data.CheckList,
		AgentConf:     data.AgentConf,
		DataProxyConf: data.DataProxyConf,
		FileProxyConf: data.FileProxyConf,
	}

	return preSetting, nil
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
		DeploymentInfo: &DeploymentInfo{
			OperInstID: data.DeploymentInfo.OperInstID,
			ActionName: data.DeploymentInfo.ActionName,
		},
		PreSetting: &PreSetting{},
	}

	var err error
	if nodeDeployment.PreSetting, err = convertPreSettingFromTypes(data.NodePreSetting); err != nil {
		return nil, fmt.Errorf("conver pre setting from types failed, err: %w", err)
	}

	return nodeDeployment, nil
}

func convertPreSettingFromTypes(data *types.NodePreSetting) (*PreSetting, error) {
	preSetting := &PreSetting{
		CheckList:     data.CheckList,
		AgentConf:     data.AgentConf,
		DataProxyConf: data.DataProxyConf,
		FileProxyConf: data.FileProxyConf,
	}

	return preSetting, nil
}
