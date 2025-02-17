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
	"runtime/debug"
	"time"

	"github.com/RichardKnop/machinery/v2"
	backendIface "github.com/RichardKnop/machinery/v2/backends/iface"
	redisBackend "github.com/RichardKnop/machinery/v2/backends/redis"
	brokerIface "github.com/RichardKnop/machinery/v2/brokers/iface"
	redisBroker "github.com/RichardKnop/machinery/v2/brokers/redis"
	machineryConfig "github.com/RichardKnop/machinery/v2/config"
	"github.com/RichardKnop/machinery/v2/tasks"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
)

const (
	// DefaultQueueName defines the default queue name.
	DefaultQueueName = "operation_inst_engine_queue"

	// ResultsExpireInDefault defines the default results expire in.
	ResultsExpireInDefault = 3600
)

// OperInstMgr defines the operation instance manager.
type OperInstMgr interface {
	// Start starts the operation instance manager.
	Start(ctx context.Context) error

	// CheckHealth checks the health of the operation instance manager.
	CheckHealth() error

	// GracefulShutdown gracefully shuts down the operation instance manager.
	GracefulShutdown() error

	// DispatchOperInst dispatches an operation instance.
	DispatchOperInst(operation *OperInst) error

	// GetRegisteredAction returns the registered action.
	GetRegisteredAction(name string) ActionDef

	// RegisterActions registers a list of actions.
	RegisterActions(actionDefs ...ActionDef) error

	// TerminateOperInst an operation inst.
	TerminateOperInst(operationInstID string) error
}

// operInstMgr ...
type operInstMgr struct {
	// machinery config
	mConfig *machineryConfig.Config

	// broker
	broker brokerIface.Broker

	// backend
	backend backendIface.Backend

	// WorkerNum defines the number of workers.
	WorkerNum int

	// state
	isRunning   bool
	isComsuming bool

	// context
	ctx    context.Context
	cancel context.CancelFunc

	server *machinery.Server
	worker *machinery.Worker

	storage OperInstStorage

	logger logger.Logger

	registeredActionDefs map[string]ActionDef

	launchWorkerErr chan error
}

// OptionsFunc is a function that configures an operInstMgr.
type OptionsFunc func(m *operInstMgr)

// WithLogger sets the logger for the operInstMgr.
func WithLogger(logger logger.Logger) OptionsFunc {
	return func(m *operInstMgr) {
		m.logger = logger
	}
}

// ServerOptionFn will be used to set the server's backend and broker.
type ServerOptionFn func(m *operInstMgr)

const (
	redisMaxIdle                = 10
	redisTimeout                = 240
	redisReadTimeout            = 15
	redisWriteTimeout           = 15
	redisConnectTimeout         = 15
	redisNormalTasksPollPeriod  = 1000
	redisDelayedTasksPollPeriod = 500
)

// WithRedis sets the redis broker for the operInstMgr.
func WithRedis(address, password string, db int) ServerOptionFn {
	return func(e *operInstMgr) {
		e.mConfig.Redis = &machineryConfig.RedisConfig{
			MaxIdle:                redisMaxIdle,
			IdleTimeout:            redisTimeout,
			ReadTimeout:            redisReadTimeout,
			WriteTimeout:           redisWriteTimeout,
			ConnectTimeout:         redisConnectTimeout,
			NormalTasksPollPeriod:  redisNormalTasksPollPeriod,
			DelayedTasksPollPeriod: redisDelayedTasksPollPeriod,
		}

		e.broker = redisBroker.New(e.mConfig, address, password, "", db)
		e.backend = redisBackend.New(e.mConfig, address, password, "", db)
	}
}

// NewOperInstMgr creates a new OperInst operInstMgr.
func NewOperInstMgr(workerNum int, envFunc ServerOptionFn, storage OperInstStorage, opts ...OptionsFunc) (
	OperInstMgr, error) {

	e := &operInstMgr{
		mConfig: &machineryConfig.Config{
			DefaultQueue:    DefaultQueueName,
			ResultsExpireIn: ResultsExpireInDefault,
			NoUnixSignals:   true,
		},
		isRunning:            false,
		registeredActionDefs: make(map[string]ActionDef),
		storage:              storage,
		logger:               logger.LoggerDefault{},
		WorkerNum:            workerNum,
		launchWorkerErr:      make(chan error, 1),
	}

	envFunc(e)

	for _, opt := range opts {
		opt(e)
	}

	return e, nil
}

// Start starts the OperInst operInstMgr.
func (mgr *operInstMgr) Start(ctx context.Context) error {
	if mgr.isRunning {
		return errors.New("operInstMgr already started")
	}

	mgr.ctx, mgr.cancel = context.WithCancel(ctx)

	if err := mgr.initialize(); err != nil {
		return err
	}

	if mgr.WorkerNum > 0 {
		if err := mgr.launchWorker(); err != nil {
			return err
		}
	}

	return nil
}

// GracefulShutdown shuts down the operInstMgr gracefully.
func (mgr *operInstMgr) GracefulShutdown() error {
	if !mgr.isRunning {
		return errors.New("operInstMgr is not running")
	}

	if mgr.isComsuming {
		go mgr.broker.StopConsuming()
	}

	mgr.isRunning = false
	mgr.isComsuming = false

	defer mgr.cancel()

	err := <-mgr.launchWorkerErr
	if !errors.Is(err, machinery.ErrWorkerQuitGracefully) {
		return err
	}

	return nil
}

// CheckHealth checks the health of the operInstMgr.
func (mgr *operInstMgr) CheckHealth() error {
	if !mgr.isRunning {
		return errors.New("operation instance manager is not running")
	}

	if mgr.server == nil {
		return errors.New("machinery server is not initialized")
	}

	if mgr.worker == nil {
		return errors.New("worker is not initialized")
	}

	return nil
}

// RegisterAction registers a OperationInst operInstMgr action.
func (mgr *operInstMgr) RegisterAction(actionDef ActionDef) error {
	if mgr.isRunning {
		return errors.New("operation instance manager already started, can not register action")
	}

	if _, ok := mgr.registeredActionDefs[actionDef.Name()]; ok {
		return errors.New("action already registered")
	}

	mgr.registeredActionDefs[actionDef.Name()] = actionDef

	return nil
}

// RegisterActions registers multiple OperInst operInstMgr actions.
func (mgr *operInstMgr) RegisterActions(actionDefs ...ActionDef) error {
	for _, actionDef := range actionDefs {
		if err := mgr.RegisterAction(actionDef); err != nil {
			return err
		}
	}

	return nil
}

// GetRegisteredAction returns a registered OperInst operInstMgr action.
func (mgr *operInstMgr) GetRegisteredAction(name string) ActionDef {
	return mgr.registeredActionDefs[name]
}

// DispatchOperInst dispatches an operation inst to the operInstMgr.
func (mgr *operInstMgr) DispatchOperInst(inst *OperInst) error {
	if inst == nil {
		return errors.New("operation instance is nil")
	}

	if !mgr.isRunning {
		return errors.New("operation instance manager is not running")
	}

	if err := inst.Validate(); err != nil {
		return err
	}

	if inst.data.Lifecycle.State != OperInstStateInit {
		return fmt.Errorf("operation instance state is not init, state: %s", inst.data.Lifecycle.State)
	}

	if err := mgr.storage.UpsertOperInstData(mgr.ctx, inst.data); err != nil {
		return err
	}

	mgr.logger.Infof("successfully store operation inst, oper-inst-id(%s)", inst.data.OperInstID)

	return mgr.dispatchOperInst(inst)
}

// StopOperInst stops an operation inst.
func (mgr *operInstMgr) StopOperInst(operationInstID string) error {
	return mgr.storage.MarkOperInstStopping(mgr.ctx, operationInstID)
}

// initialize initializes the OperInst operInstMgr.
func (mgr *operInstMgr) initialize() error {
	if mgr.WorkerNum <= 0 {
		return errors.New("worker num should be greater than 0")
	}

	if mgr.mConfig == nil {
		return errors.New("operation instance manager config is nil")
	}

	if mgr.broker == nil {
		return errors.New("broker is nil")
	}

	if mgr.backend == nil {
		return errors.New("backend is nil")
	}

	if mgr.storage == nil {
		return errors.New("storage is nil")
	}

	// We don't need to use the periodic task of machinery, so there is no need to access the lock.
	mgr.server = machinery.NewServer(mgr.mConfig, mgr.broker, mgr.backend, nil)

	mgr.isRunning = true

	return nil
}

const consumerTag = ""

func (mgr *operInstMgr) launchWorker() error {
	if !mgr.isRunning {
		return errors.New("operation instance manager is not running")
	}

	for _, actionDef := range mgr.registeredActionDefs {
		if err := mgr.server.RegisterTask(actionDef.Name(), mgr.do); err != nil {
			return err
		}
	}

	mgr.worker = mgr.server.NewWorker(consumerTag, mgr.WorkerNum)

	mgr.worker.SetErrorHandler(func(err error) {
		mgr.logger.Errorf("worker error, err: %v", err)
	})
	mgr.worker.SetPreTaskHandler(func(signature *tasks.Signature) {})
	mgr.worker.SetPostTaskHandler(func(signature *tasks.Signature) {})

	mgr.isComsuming = true
	go func() {
		mgr.launchWorkerErr <- mgr.worker.Launch()
		mgr.isComsuming = false
	}()

	return nil
}

// dispatchOperInst dispatch OperInst to machinery chain.
func (mgr *operInstMgr) dispatchOperInst(inst *OperInst) error {
	var signatures []*tasks.Signature

	for _, actionDef := range inst.operationDef.actionDefs {
		signature := &tasks.Signature{
			UUID: fmt.Sprintf("%s-%s", inst.data.OperInstID, actionDef.Name()),
			Name: actionDef.Name(),
			Args: []tasks.Arg{
				{
					Name:  "action",
					Type:  "string",
					Value: actionDef.Name(),
				},
				{
					Name:  "operation-instance-id",
					Type:  "string",
					Value: inst.data.OperInstID,
				},
			},
		}

		signatures = append(signatures, signature)
	}

	if len(signatures) == 0 {
		return fmt.Errorf("no action to do")
	}

	chain, err := tasks.NewChain(signatures...)
	if err != nil {
		return err
	}

	_, err = mgr.server.SendChainWithContext(mgr.ctx, chain)
	if err != nil {
		return fmt.Errorf("send chain to machinery failed, err: %v", err)
	}

	return nil
}

// EngineMaxRetryLimit is the max retry limit of an action.
const EngineMaxRetryLimit = uint(10)

// do will dispatch the action of OperInst to machinery.
func (mgr *operInstMgr) do(ctx context.Context, actionName string, operInstID string) (err error) {
	actionDef, ok := mgr.registeredActionDefs[actionName]
	if !ok {
		return fmt.Errorf("action not registered, name(%s)", actionName)
	}

	// operation operInstData context is used to control the timeout of the operation operInstData.
	operInstData, err := mgr.storage.GetOperInstDataWithoutActionData(ctx, operInstID)
	if err != nil {
		return fmt.Errorf("failed to get operation instance actionInstData from db. "+
			"oper-operInstData-id(%s), err: %v", operInstID, err)
	}

	actionInstData, err := mgr.storage.GetActionInstData(ctx, operInstID, actionName)
	if err != nil {
		return fmt.Errorf("failed to get action instance actionInstData from operation operInstData. "+
			"oper-operInstData-id(%s), action-name(%s), err: %v", operInstID, actionName, err)
	}

	skip, err := checkActionInstState(actionInstData)
	if err != nil {
		return err
	}
	if skip {
		return nil
	}

	if err := mgr.markActionRunning(actionInstData, operInstData); err != nil {
		return err
	}

	// action context is used to control the timeout of the action.
	actionCtx, actionCancel := context.WithTimeout(ctx, actionDef.Timeout())
	defer actionCancel()

	startAt := operInstData.Lifecycle.StartedAt
	operInstCtx, cancel := context.WithDeadline(ctx, startAt.Add(operInstData.Timeout))
	defer cancel()

	// watch storage for stopping event.
	terminatingC := mgr.storage.WatchOperInstStopping(actionCtx, operInstData.OperInstID)

	doResult := make(chan error, 1)
	actionInstCtx := &ActionInstContext{
		Ctx:  actionCtx,
		Data: actionInstData,
	}

	go mgr.executeAction(doResult, actionInstCtx, actionDef)

	defer func() {
		// when all action done or error happens, we need to update the state of the operation operInstData.
		if actionInstCtx.Data.Index == len(operInstData.ActionNames)-1 || err != nil {
			switch actionInstCtx.Data.State {
			case ActionInstStateSuccess:
				operInstData.Lifecycle.State = OperInstStateSuccess
			case ActionInstStateFailed:
				operInstData.Lifecycle.State = OperInstStateFailed
			case ActionInstStateTimeout:
				operInstData.Lifecycle.State = OperInstStateTimeout
			case ActionInstStateTerminated:
				operInstData.Lifecycle.State = OperInstStateTerminated
			case ActionInstStateRunning:
				operInstData.Lifecycle.State = OperInstStateRunning
			case ActionInstStatePending:
				// TODO: implement me
				operInstData.Lifecycle.State = OperInstStateFailed
			case ActionInstStateSkipped:
				// last action shouldn't be skipped
				operInstData.Lifecycle.State = OperInstStateFailed
			case ActionInstStateUnknown:
				// last action shouldn't be unknown
				operInstData.Lifecycle.State = OperInstStateFailed
			}
		}

		storeErr := mgr.storage.UpdateLifecycle(ctx, operInstData.OperInstID, operInstData.Lifecycle)
		if storeErr != nil {
			err = fmt.Errorf("failed to update operation inst lifecycle, "+
				"oper-inst-id(%s), action-name(%s), store-err(%v), original-err(%v)",
				operInstID, actionName, storeErr, err)
		}
	}()

	select {
	case err = <-doResult:
		actionInstData.EndedAt = time.Now()

		if err == nil {
			actionInstData.State = ActionInstStateSuccess
		} else {
			actionInstData.State = ActionInstStateFailed
		}

		return err
	case <-actionCtx.Done():
		return fmt.Errorf("action timeout, operation-operInstData-id(%s), action-name(%s)", operInstID, actionName)
	case <-operInstCtx.Done():
		return fmt.Errorf("operation operInstData timeout, operation-operInstData-id(%s), action-name(%s)",
			operInstID, actionName)
	case <-terminatingC:
		return fmt.Errorf("operation operInstData has been terminated, operation-operInstData-id(%s), action-name(%s)",
			operInstID, actionName)
	case <-ctx.Done():
		return fmt.Errorf("operation operInstMgr context done, operation-operInstData-id(%s), action-name(%s)",
			operInstID, actionName)
	}
}

// markActionRunning mark action running.
func (mgr *operInstMgr) markActionRunning(actionInstData *ActionInstData, operInstData *OperInstData) error {
	nowTime := time.Now()
	actionInstData.StartedAt = nowTime
	actionInstData.State = ActionInstStateRunning

	// first action
	if actionInstData.Index == 0 {
		actionInstData.Content = operInstData.InitContent
		operInstData.Lifecycle.StartedAt = actionInstData.StartedAt

		if err := mgr.storage.UpdateLifecycle(mgr.ctx, operInstData.OperInstID, operInstData.Lifecycle); err != nil {
			return fmt.Errorf("failed to store operation inst data. oper-inst-id(%s), action-name(%s), err: %v",
				actionInstData.OperInstID, actionInstData.Name, err)
		}
	} else {
		preActionName := operInstData.ActionNames[actionInstData.Index-1]
		preActionInstContent, err := mgr.storage.GetActionInstData(mgr.ctx, actionInstData.OperInstID, preActionName)
		if err != nil {
			return fmt.Errorf("failed to get action inst data. operation-operInst-id(%s), action-name(%s), err: %v",
				actionInstData.OperInstID, actionInstData.Name, err)
		}

		actionInstData.Content = preActionInstContent.Content
	}

	if err := mgr.storage.UpdateActionInstData(mgr.ctx, actionInstData); err != nil {
		return fmt.Errorf("failed to store operation operInst param. operation-operInst-id(%s), action-name(%s), err: %v",
			actionInstData.OperInstID, actionInstData.Name, err)
	}

	return nil
}

// checkActionInstState evaluate action State is ready to run.
func checkActionInstState(data *ActionInstData) (skip bool, err error) {
	if data == nil {
		return false, fmt.Errorf("action instance data is nil")
	}

	switch data.State {
	case ActionInstStateSuccess, ActionInstStateSkipped:
		return true, nil
	case ActionInstStateFailed, ActionInstStateTimeout,
		ActionInstStateTerminated:
		return false, fmt.Errorf("operation instance has completed. oper-inst-id(%s), action-name(%s), State(%s)",
			data.OperInstID, data.Name, data.State)
	case ActionInstStateRunning:
		return false, fmt.Errorf("action is running. oper-inst-id(%s), action-name(%s), State(%s)",
			data.OperInstID, data.Name, data.State)
	case ActionInstStatePending:
		return false, nil
	default:
		return false, fmt.Errorf("unexpected action State. oper-inst-id(%s), action-name(%s), State(%s)",
			data.OperInstID, data.Name, data.State)
	}
}

// executeAction execute action.
func (mgr *operInstMgr) executeAction(doResult chan error, actionInstCtx *ActionInstContext, actionDef ActionDef) {
	var err error

	defer func() {
		if r := recover(); r != nil {
			mgr.logger.Errorf("action panic, info(%v), revoer(%v), stack(%s)",
				actionInstCtx.Data.Info(), r, debug.Stack())
			err = fmt.Errorf("action panic, info(%v), revoer(%v), stack(%s)",
				actionInstCtx.Data.Info(), r, debug.Stack())
		}

		doResult <- err
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go mgr.autoRefreshActionDataMsg(ctx, actionInstCtx.Data)

	err = mgr.callActionDefWithRetry(actionInstCtx, actionDef)
}

const msgRefreshInterval = 1 * time.Second

func (mgr *operInstMgr) autoRefreshActionDataMsg(ctx context.Context, data *ActionInstData) {
	ticker := time.NewTicker(msgRefreshInterval)
	defer ticker.Stop()

	// refresh action inst data messages to storage
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := mgr.storage.UpdateActInstMsg(ctx, data.OperInstID, data.Name, data.Messages); err != nil {
				mgr.logger.Errorf("failed to refresh action inst data messages, action-name(%s), err: %v",
					data.Name, err)
			}
			continue
		}
	}
}

// callActionDefWithRetry do action with retry.
func (mgr *operInstMgr) callActionDefWithRetry(actionInstCtx *ActionInstContext, actionDef ActionDef) error {
	var doErr error

	for retryNum := uint(0); retryNum <= actionDef.MaxRetryCount() && retryNum < EngineMaxRetryLimit; retryNum++ {
		mgr.logger.Infof("successfully started action, action-name(%s), retry-num(%d)",
			actionInstCtx.Data.Name, retryNum)
		actionInstCtx.Data.Log(fmt.Sprintf("successfully started action, action-name(%s), retry-num(%d)",
			actionInstCtx.Data.Name, retryNum))

		doErr = actionDef.Do(actionInstCtx)

		if doErr != nil {
			mgr.logger.Errorf("failed to do action, action-name(%s), retry-num(%d), err: %v",
				actionInstCtx.Data.Name, retryNum, doErr)
			actionInstCtx.Data.Log(fmt.Sprintf("failed to do action, action-name(%s), retry-num(%d), err: %v",
				actionInstCtx.Data.Name, retryNum, doErr))

			actionDef.DelayFn()
			continue
		}

		mgr.logger.Infof("successfully done action, action-name(%s), retry-num(%d)",
			actionInstCtx.Data.Name, retryNum)
		actionInstCtx.Data.Log(fmt.Sprintf("successfully done action, action-name(%s), retry-num(%d)",
			actionInstCtx.Data.Name, retryNum))

		break
	}

	return doErr
}

// TerminateOperInst an operation instance.
func (mgr *operInstMgr) TerminateOperInst(operInstID string) error {
	// TODO: implement me
	panic("implement me")
}
