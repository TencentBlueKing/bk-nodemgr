/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package nodeinstall

import (
	"fmt"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/nodedeployment"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operengine"
)

// NewActionUpdateHost ...
func NewActionUpdateHost(
	iDaoHost topoStg.IDaoHost,
	iDaoNodeDeployment nodedeployment.IDaoNodeDeployment,
	logger logger.Logger,
) operengine.ActionDef {

	return &UpdateHost{
		iDaoHost:           iDaoHost,
		iDaoNodeDeployment: iDaoNodeDeployment,
		logger:             logger,
	}
}

// UpdateHostParam ...
type UpdateHostParam struct {
	Token string `json:"token"`
}

// UpdateHost ...
type UpdateHost struct {
	iDaoHost           topoStg.IDaoHost
	iDaoNodeDeployment nodedeployment.IDaoNodeDeployment
	logger             logger.Logger
}

// Name returns the name of the action.
func (action *UpdateHost) Name() string {
	return ActionNameUpdateHost
}

// Version returns the version of the action.
func (action *UpdateHost) Version() string {
	return "1.0.0"
}

// Description returns the description of the action.
func (action *UpdateHost) Description() string {
	return "update host"
}

// Timeout returns the timeout of the action.
func (action *UpdateHost) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (action *UpdateHost) Tags() []operengine.ActionTag {
	return []operengine.ActionTag{}
}

// MaxRetryCount returns the max retry count of the action.
func (action *UpdateHost) MaxRetryCount() uint {
	return 3 //nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (action *UpdateHost) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
func (action *UpdateHost) Do(ctx *operengine.ActionInstContext) error {
	param := new(UpdateHostParam)
	err := conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	info, err := action.iDaoNodeDeployment.GetInfo(ctx.Ctx, param.Token)
	if err != nil {
		return fmt.Errorf("get node deployment info failed, err: %w", err)
	}

	err = action.iDaoHost.UpdateManyHostDynamic(ctx.Ctx, &info.Host)
	if err != nil {
		return fmt.Errorf("update host dynamic failed, err: %w", err)
	}

	return nil
}
