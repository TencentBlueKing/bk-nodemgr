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
	"errors"
	"fmt"
	"time"

	nodeUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/node/utils"
	nodeStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameUpdateHost defines the action name.
	ActionNameUpdateHost = "update_host"
)

// NewActionUpdateHost get a new action.
func NewActionUpdateHost(capability *Capability) action.Definition {
	return &actionUpdateHost{
		storageHost:           capability.StorageTopo,
		storageNodeDeployment: capability.StorageNode,
	}
}

// ActParamUpdateHost ...
type ActParamUpdateHost struct {
	nodeUtils.NodeActionStandardParam `json:",inline"`
}

// UpdateHost ...
type actionUpdateHost struct {
	storageHost           topoStg.IStorageHost
	storageNodeDeployment nodeStg.IDaoNodeDeployment
}

// Name returns the name of the action.
func (act *actionUpdateHost) Name() string {
	return ActionNameUpdateHost
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionUpdateHost) DisplayNameZh() string {
	return "更新主机信息"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionUpdateHost) DisplayNameEn() string {
	return "Update Host"
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

	// initialize standard data.
	std := nodeUtils.NewNodeActionStandarder(act.storageNodeDeployment, act.storageHost)
	if err = std.Initialize(ctx, param.NodeActionStandardParam); err != nil {
		return err
	}
	defer func() {
		if storeErr := std.Save(); storeErr != nil {
			err = errors.Join(storeErr, err)
		}
	}()

	err = act.storageHost.UpdateManyHostDynamic(std.Context(), &std.DeployInfo().Host)
	if err != nil {
		return fmt.Errorf("update host dynamic failed: %w", err)
	}

	if touchErr := act.storageHost.TouchHostOperationTime(std.Context(), std.DeployInfo().Host.HostID); touchErr != nil {
		logger.G.Sys().
			WithErr(touchErr).
			With("host-id", std.DeployInfo().Host.HostID).
			Warn("failed to touch host operation time after update host dynamic")
	}

	return nil
}
