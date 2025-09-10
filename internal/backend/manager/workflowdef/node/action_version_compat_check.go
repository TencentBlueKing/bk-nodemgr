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

	nodedeployment "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node-deployment"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameVersionCompatCheck defines the action name.
	ActionNameVersionCompatCheck = "version_compat_check"

	// since v2.1.6-beta.35, GSE supports operate agent restart with idle check.
	agentOperateRestartLowestVersion = "v2.1.6-any.35"
)

// NewActionVersionCompatCheck get a new action.
func NewActionVersionCompatCheck(storageNodeDeployment nodedeployment.IStorageNodeDeployment,
	logger logger.ILogger) action.Definition {

	return &actionVersionCompatCheck{
		storageNodeDeployment: storageNodeDeployment,
		logger:                logger,

		operateAgentSupportedLowestVersionFmt: types.NewGSEVersionFormatter(agentOperateRestartLowestVersion),
	}
}

// ActionParamVersionCompatCheck defines the action param.
type ActionParamVersionCompatCheck struct {
	Token string `json:"token"`
}

type actionVersionCompatCheck struct {
	storageNodeDeployment nodedeployment.IStorageNodeDeployment
	logger                logger.ILogger

	operateAgentSupportedLowestVersionFmt types.GSEVersionFormatter
}

// Name returns the name of the action.
func (act *actionVersionCompatCheck) Name() string {
	return ActionNameVersionCompatCheck
}

// Version returns the version of the action.
func (act *actionVersionCompatCheck) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionVersionCompatCheck) Description() string {
	return "version compat check"
}

// Timeout returns the timeout of the action.
func (act *actionVersionCompatCheck) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionVersionCompatCheck) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionVersionCompatCheck) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionVersionCompatCheck) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
// nolint: funlen,nonamedreturns
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (act *actionVersionCompatCheck) Do(ctx *action.InstanceContext) (err error) {
	param := new(ActionParamVersionCompatCheck)
	err = conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		err = fmt.Errorf("failed to convert param, err: %w", err)

		return err
	}

	info, err := act.storageNodeDeployment.GetInfo(ctx.Ctx, param.Token)
	if err != nil {
		return err
	}

	defer func() {
		if storeErr := act.storageNodeDeployment.UpdateInfo(ctx.Ctx, param.Token, info); storeErr != nil {
			err = errors.Join(storeErr, err)
		}
	}()

	// check if this node version is >= lowest version which supports the soft restart through cluster.
	versionFormatter := types.NewGSEVersionFormatter(info.Host.Dynamic.NodeVersion)
	if versionFormatter.Valid() && act.operateAgentSupportedLowestVersionFmt.Valid() &&
		versionFormatter.GreaterEqualThan(act.operateAgentSupportedLowestVersionFmt) {

		info.CurrentVersionSupports.OperateAgentRestart = true

		return nil
	}

	return nil
}
