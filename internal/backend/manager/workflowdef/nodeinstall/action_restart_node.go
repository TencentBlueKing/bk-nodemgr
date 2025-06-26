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
	ActionNameRestartNode = "restart_node"
)

// NewActionStartNode get a new action.
func NewActionStartNode(storageNodeDeployment nodedeployment.IStorageNodeDeployment,
	gseHandler gse.IHandler,
	logger logger.Logger) action.Definition {
	return &actionRestartNode{
		storageNodeDeployment: storageNodeDeployment,
		gseHandler:            gseHandler,
		logger:                logger,
	}
}

// ActionParamRestartNode defines the action param.
type ActionParamRestartNode struct {
	Token string `json:"token"`
}

type actionRestartNode struct {
	storageNodeDeployment nodedeployment.IStorageNodeDeployment
	gseHandler            gse.IHandler
	logger                logger.Logger
}

// Name returns the name of the action.
func (act *actionRestartNode) Name() string {
	return ActionNameUpgradeNode
}

// Version returns the version of the action.
func (act *actionRestartNode) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionRestartNode) Description() string {
	return "restart node"
}

// Timeout returns the timeout of the action.
func (act *actionRestartNode) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionRestartNode) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionRestartNode) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionRestartNode) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
// nolint: funlen,fnsize,nonamedreturns
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (act *actionRestartNode) Do(_ *action.InstanceContext) (err error) {
	return nil
}
