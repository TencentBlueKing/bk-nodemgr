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

	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameTransferPkgToNode defines the action name.
	ActionNameTransferPkgToNode = "upgrade_node"
)

// NewActionTransferPkgToNode get a new action.
func NewActionTransferPkgToNode() action.Definition {
	return &actionTransferPkgToNode{}
}

// ActionParamTransferPkgToNode defines the action param.
type ActionParamTransferPkgToNode struct {
	Token string `json:"token"`
}

type actionTransferPkgToNode struct {
}

// Name returns the name of the action.
func (act *actionTransferPkgToNode) Name() string {
	return ActionNameTransferPkgToNode
}

// Version returns the version of the action.
func (act *actionTransferPkgToNode) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionTransferPkgToNode) Description() string {
	return "transfer pkg to node"
}

// Timeout returns the timeout of the action.
func (act *actionTransferPkgToNode) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionTransferPkgToNode) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionTransferPkgToNode) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionTransferPkgToNode) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
// nolint: funlen,fnsize
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (act *actionTransferPkgToNode) Do(_ *action.InstanceContext) error {
	return nil
}
