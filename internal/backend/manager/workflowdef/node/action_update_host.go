/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package node

import (
	"fmt"
	"time"

	nodedeployment "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node-deployment"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameUpdateHost defines the action name.
	ActionNameUpdateHost = "update_host"
)

// NewActionUpdateHost get a new action.
func NewActionUpdateHost(
	storageHost topo.IStorageHost,
	storageNodeDeployment nodedeployment.IStorageNodeDeployment,
	logger logger.ILogger,
) action.Definition {

	return &actionUpdateHost{
		storageHost:           storageHost,
		storageNodeDeployment: storageNodeDeployment,
		logger:                logger,
	}
}

// ActParamUpdateHost ...
type ActParamUpdateHost struct {
	Token string `json:"token"`
}

// UpdateHost ...
type actionUpdateHost struct {
	storageHost           topo.IStorageHost
	storageNodeDeployment nodedeployment.IStorageNodeDeployment
	logger                logger.ILogger
}

// Name returns the name of the action.
func (act *actionUpdateHost) Name() string {
	return ActionNameUpdateHost
}

// Version returns the version of the action.
func (act *actionUpdateHost) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionUpdateHost) Description() string {
	return "update host to storage"
}

// Timeout returns the timeout of the action.
func (act *actionUpdateHost) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionUpdateHost) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionUpdateHost) MaxRetryCount() uint {
	return 3 //nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionUpdateHost) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
func (act *actionUpdateHost) Do(ctx *action.InstanceContext) error {
	param := new(ActParamUpdateHost)
	err := conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	info, err := act.storageNodeDeployment.GetInfo(ctx.Ctx, param.Token)
	if err != nil {
		return fmt.Errorf("get node deployment info failed, err: %w", err)
	}

	err = act.storageHost.UpdateManyHostDynamic(ctx.Ctx, &info.Host)
	if err != nil {
		return fmt.Errorf("update host dynamic failed, err: %w", err)
	}

	return nil
}
