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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameEnableReleaseTransfer defines the action name.
	ActionNameEnableReleaseTransfer = "enable_release_transfer"
)

// NewActionEnableReleaseTransfer get a new action.
func NewActionEnableReleaseTransfer(capability *Capability) action.Definition {
	return &actionEnableReleaseTransfer{
		storageNodeDeployment: capability.StorageNode,
	}
}

// ActionParamEnableReleaseTransfer defines the action param.
type ActionParamEnableReleaseTransfer struct {
	nodeUtils.NodeActionStandardParam `json:",inline"`
}

type actionEnableReleaseTransfer struct {
	storageNodeDeployment nodeStg.IDaoNodeDeployment
}

// Name returns the name of the action.
func (act *actionEnableReleaseTransfer) Name() string {
	return ActionNameEnableReleaseTransfer
}

// Version returns the version of the action.
func (act *actionEnableReleaseTransfer) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionEnableReleaseTransfer) Description() string {
	return "enable release transfer option"
}

// Timeout returns the timeout of the action.
func (act *actionEnableReleaseTransfer) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionEnableReleaseTransfer) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionEnableReleaseTransfer) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionEnableReleaseTransfer) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
// nolint: funlen,nonamedreturns
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (act *actionEnableReleaseTransfer) Do(ctx *action.InstanceContext) error {
	param := new(ActParamDetectInfoBySSH)
	err := conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	// initialize standard data.
	std := nodeUtils.NewNodeActionStandarder(act.storageNodeDeployment)
	if err = std.Initialize(ctx, param.NodeActionStandardParam); err != nil {
		return err
	}
	defer func() {
		if storeErr := std.Save(); storeErr != nil {
			err = errors.Join(storeErr, err)
		}
	}()

	// enable release transfer. close installer transfer.
	// so that first action will transfer installer package, than next action will transfer release package.
	std.DeployInfo().TransferOptions.EnableInstaller = false
	std.DeployInfo().TransferOptions.EnableReleasePackage = true

	std.InstanceData().LogI(fmt.Sprintf("enable release transfer success. installer(%v), release(%v)",
		std.DeployInfo().TransferOptions.EnableInstaller, std.DeployInfo().TransferOptions.EnableReleasePackage))

	return nil
}
