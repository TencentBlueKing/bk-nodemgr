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
	"context"
	"fmt"
	"time"

	nodeStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/cache"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
)

const (
	// OperExtraExecutionName defines the operation instance extra execution name.
	OperExtraExecutionName = "node_operation_extra_execution"
)

// NewOperationExtraExecution creates a new operation extra execution.
func NewOperationExtraExecution(
	locker cache.ICache,
	storageNodeDeployment nodeStg.IDaoNodeDeployment,
) operation.ExtraExecution {

	return &extraExecution{
		locker:                locker,
		storageNodeDeployment: storageNodeDeployment,
	}
}

// ExtraExecutionParam ...
type ExtraExecutionParam struct {
	Token string `json:"token"`
}

type extraExecution struct {
	locker                cache.ICache
	storageNodeDeployment nodeStg.IDaoNodeDeployment
}

// Name returns the name of the action.
func (exec *extraExecution) Name() string {
	return OperExtraExecutionName
}

// Do this func define what the action will do.
func (exec *extraExecution) Do(ctx context.Context, instance *operation.InstanceBriefData) error {
	param := new(ExtraExecutionParam)
	err := conv.MapToStruct(instance.Metadata.InitContent, param)
	if err != nil {
		return err
	}

	info, err := exec.storageNodeDeployment.GetNodeDeploymentInfo(ctx, param.Token)
	if err != nil {
		instance.LogE(fmt.Sprintf("get deployment info by token(%s) failed: %v", param.Token, err))
		return fmt.Errorf("get node deployment info by token(%s) failed: %w", param.Token, err)
	}

	lockerName := GenLockerName(
		info.Host.Static.NetworkAreaID, info.Host.Static.InnerIP, string(info.Host.Static.Addressing),
	)

	switch instance.Lifecycle.State {
	case operation.StateLaunched:
		err := exec.Lock(ctx, lockerName, instance.Metadata.OperationInstanceID, instance.Metadata.Timeout)
		if err != nil {
			instance.LogE(err.Error())
			instance.Lifecycle.End(action.StateFailed)

			return err
		}

		instance.LogI(fmt.Sprintf("lock host by locker(%s) success", lockerName))

		return nil

	case operation.StateSuccess, operation.StateFailed, operation.StateTimeout, operation.StateTerminated:
		err := exec.Unlock(ctx, lockerName, instance.Metadata.OperationInstanceID)
		if err != nil {
			instance.LogE(err.Error())
			instance.Lifecycle.End(action.StateFailed)

			return fmt.Errorf("unlock host by locker(%s) failed: %w", lockerName, err)
		}

		instance.LogI(fmt.Sprintf("unlock host by locker(%s) success", lockerName))

		return nil
	default:
		instance.LogE(fmt.Sprintf("unexpected operation instance state: %s", instance.Lifecycle.State))
		instance.Lifecycle.End(action.StateFailed)

		return fmt.Errorf("unexpected operation instance state: %s", instance.Lifecycle.State)
	}
}

// Lock this func define what the action will do when locking.
func (exec *extraExecution) Lock(
	ctx context.Context, lockerName string, operationInstanceID string, expireTime time.Duration) error {

	locked, err := exec.locker.SetNXWithExpiration(
		ctx,
		lockerName,
		[]byte(operationInstanceID),
		expireTime,
	)
	if err != nil {
		return fmt.Errorf("lock host by locker(%s) failed: %w", lockerName, err)
	}

	if !locked {
		otherOperInstID, err := exec.locker.Get(ctx, lockerName)
		if err != nil {
			return fmt.Errorf("get lock by locker(%s) failed: %w", lockerName, err)
		}

		return fmt.Errorf("locker(%s) already locked by other oper-inst(%s)", lockerName, otherOperInstID)
	}

	return nil
}

// Unlock this func define what the action will do when unlocking.
func (exec *extraExecution) Unlock(ctx context.Context, lockerName string, operationInstanceID string) error {
	operInstID, err := exec.locker.Get(ctx, lockerName)
	if err != nil {
		return fmt.Errorf("get lock by locker(%s) failed: %w", lockerName, err)
	}

	if string(operInstID) != operationInstanceID {
		return fmt.Errorf("locker(%s) not match, oper-inst-id(%s), expected(%s)",
			lockerName, operInstID, operationInstanceID)
	}

	unlocked, err := exec.locker.Delete(ctx, lockerName)
	if err != nil {
		return fmt.Errorf("unlock host by locker(%s) failed: %w", lockerName, err)
	}

	if !unlocked {
		return fmt.Errorf("locker(%s) already unlocked", lockerName)
	}

	return nil
}

// GenLockerName generates a unique locker name based on the node deployment info.
func GenLockerName(networkAreaID int64, ip, addressing string) string {
	return fmt.Sprintf("bknm:backend:node:%d:%s:%s", networkAreaID, ip, addressing)
}
