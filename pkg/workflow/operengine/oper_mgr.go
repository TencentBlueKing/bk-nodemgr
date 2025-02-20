/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package operengine ...
package operengine

import (
	"context"
	"errors"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/locker"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
)

// OperationMgr defines the operation operInstMgr.
type OperationMgr interface {
	// RetryOperation an operation.
	RetryOperation(operationID string, param *OperInstParam) error

	// ExecuteOperation an operation.
	ExecuteOperation(operation *Operation, param *OperInstParam) error

	// PauseOperation an operation.
	PauseOperation(operationID string) error

	// ResumeOperation an operation.
	ResumeOperation(operationID string) error
}

// OperMgrOptFn ...
type OperMgrOptFn func(*operMgr)

// OperMgrWithLogger ...
func OperMgrWithLogger(logger logger.Logger) OperMgrOptFn {
	return func(mgr *operMgr) {
		mgr.logger = logger
	}
}

// OperMgrWithLockerFactory ...
func OperMgrWithLockerFactory(mutexFactory locker.MutexFactory) OperMgrOptFn {
	return func(mgr *operMgr) {
		mgr.mutexFactory = mutexFactory
	}
}

// NewOperationMgr creates a new OperationMgr.
func NewOperationMgr(operInstMgr OperInstMgr, storage OperationStorage, opts ...OperMgrOptFn) (OperationMgr, error) {
	mgr := &operMgr{
		operInstMgr:  operInstMgr,
		storage:      storage,
		mutexFactory: locker.MutexFactoryDefault{},
		logger:       logger.LoggerDefault{},
	}

	// TODO: 为 operationMgr 补充 Start 和 TerminateOperInst 方法
	mgr.ctx, mgr.cancel = context.WithCancel(context.Background())

	for _, opt := range opts {
		opt(mgr)
	}

	return mgr, nil
}

// operMgr ...
type operMgr struct {
	operInstMgr  OperInstMgr
	storage      OperationStorage
	mutexFactory locker.MutexFactory
	logger       logger.Logger

	// context
	ctx    context.Context
	cancel context.CancelFunc

	// state
	isRunning bool
}

// buildInst parse operation define snapshot, then use it to create an operation instance.
func (m *operMgr) buildInst(operation *Operation, param *OperInstParam) (
	*OperInst, error) {

	operDef := newOperationDef(operation.DefSnapshot.OperDefName)
	for _, actionName := range operation.DefSnapshot.ActionNames {
		operDef.Next(m.operInstMgr.GetRegisteredAction(actionName))
	}

	operInst, err := operDef.NewInstance(operation.TriggerID, param.Timeout)
	if err != nil {
		return nil, err
	}

	operInst.data.InitContent = param.InitContent
	operInst.data.ParentOperInstID = param.ParentOperInstID

	for idx, actionName := range operation.DefSnapshot.ActionNames {
		actionInstData := &ActionInstData{
			TriggerID:  operInst.data.TriggerID,
			OperInstID: operInst.data.OperInstID,
			Name:       actionName,
			Index:      idx,
			Lifecycle: &ActInstLifeCycle{
				State: ActionInstStatePending,
			},
			Messages: make([]Message, 0),
			Content:  make(map[string]any),
		}

		operInst.data.ActionInstDataMap[actionName] = actionInstData
	}

	return operInst, nil
}

// ExecuteOperation an operation.
func (m *operMgr) ExecuteOperation(operation *Operation, param *OperInstParam) (err error) {
	if operation == nil {
		return errors.New("operation is nil")
	}

	if param == nil {
		return errors.New("param is nil")
	}

	mutex := m.mutexFactory.NewMutex(operation.OperationID)
	if err = mutex.TryLock(); err != nil {
		return err
	}

	defer func() {
		if lockErr := mutex.Unlock(); lockErr != nil {
			err = fmt.Errorf("original-err(%v), lock-err(%v)", err, lockErr)
		}
	}()

	operInst, err := m.buildInst(operation, param)
	if err != nil {
		return err
	}

	operation.OperInstIDs = append(operation.OperInstIDs, operInst.data.OperInstID)
	if err = m.storage.UpsertOperation(m.ctx, operation); err != nil {
		return err
	}

	if err = m.operInstMgr.DispatchOperInst(operInst); err != nil {
		return err
	}

	return nil
}

// RetryOperation an operation.
func (m *operMgr) RetryOperation(operationID string, param *OperInstParam) (err error) {
	operation, err := m.storage.GetOperation(m.ctx, operationID)
	if err != nil {
		return err
	}

	if err = operation.CheckEnforceability(); err != nil {
		return err
	}

	if err = m.ExecuteOperation(operation, param); err != nil {
		return err
	}

	return nil
}

// PauseOperation an operation.
func (m *operMgr) PauseOperation(operationID string) error {
	//TODO implement me
	panic("implement me")
}

// ResumeOperation an operation.
func (m *operMgr) ResumeOperation(operationID string) error {
	//TODO implement me
	panic("implement me")
}
