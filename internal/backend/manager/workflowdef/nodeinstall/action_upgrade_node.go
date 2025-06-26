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
	"time"

	nodedeployment "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node-deployment"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/gse"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameUpgradeNode defines the action name.
	ActionNameUpgradeNode = "upgrade_node"
)

// NewActionUpgradeNode get a new action.
func NewActionUpgradeNode(storageNodeDeployment nodedeployment.IStorageNodeDeployment,
	gseHandler gse.IHandler,
	logger logger.Logger) action.Definition {

	return &actionUpgradeNode{
		storageNodeDeployment: storageNodeDeployment,
		gseHandler:            gseHandler,
		logger:                logger,
	}
}

// ActionParamUpgradeNode defines the action param.
type ActionParamUpgradeNode struct {
	Token string `json:"token"`
}

type actionUpgradeNode struct {
	storageNodeDeployment nodedeployment.IStorageNodeDeployment
	gseHandler            gse.IHandler
	logger                logger.Logger
}

// Name returns the name of the action.
func (act *actionUpgradeNode) Name() string {
	return ActionNameUpgradeNode
}

// Version returns the version of the action.
func (act *actionUpgradeNode) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionUpgradeNode) Description() string {
	return "upgrade node"
}

// Timeout returns the timeout of the action.
func (act *actionUpgradeNode) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionUpgradeNode) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionUpgradeNode) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionUpgradeNode) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
// nolint: funlen,fnsize,nonamedreturns
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (act *actionUpgradeNode) Do(_ *action.InstanceContext) (err error) {
	return nil
}
