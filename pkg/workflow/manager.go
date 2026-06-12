/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package workflow ...
package workflow

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/RichardKnop/machinery/v2"
	backendIface "github.com/RichardKnop/machinery/v2/backends/iface"
	backendRedis "github.com/RichardKnop/machinery/v2/backends/redis"
	brokerIface "github.com/RichardKnop/machinery/v2/brokers/iface"
	brokerRedis "github.com/RichardKnop/machinery/v2/brokers/redis"
	machineryConfig "github.com/RichardKnop/machinery/v2/config"
	machineryLog "github.com/RichardKnop/machinery/v2/log"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/config"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/locker"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/tracing"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
)

// IManager defines the manager interface.
type IManager interface {
	IController
	IWorker

	// Start starts the worker.
	Start(ctx contextx.IContext) error

	// CheckHealth checks the health of the worker.
	CheckHealth() error

	// GracefulShutdown gracefully shutdowns the worker.
	GracefulShutdown() error

	// RegisterActions registers a list of actions.
	RegisterActions(actionDefs ...action.Definition) error

	// RegisterOperExtraExecutions registers a list of extra execution definitions.
	RegisterOperExtraExecutions(extraDefs ...operation.ExtraExecution) error
}

const (
	queueNameDefault                       = "operation_inst_engine_queue"
	resultsExpireInDefault                 = 3600
	defaultGracefulShutdownTimeoutDuration = 60 * time.Second
)

// NewManager creates a new manager.
func NewManager(workerNum int, opts ...OptionsFunc) (IManager, error) {
	mgr := &manager{
		mConfig: &machineryConfig.Config{
			DefaultQueue:    queueNameDefault,
			ResultsExpireIn: resultsExpireInDefault,
			// notice: this is used to prevent the worker from being terminated by the SIGTERM signal.
			// we will handle the graceful shutdown in the manager.
			NoUnixSignals: true,
		},
		isRunning:                        false,
		registeredActionDefs:             make(map[string]action.Definition),
		registeredOperExtraExecutionDefs: make(map[string]operation.ExtraExecution),
		WorkerNum:                        workerNum,
		gracefulShutdownTimeout:          defaultGracefulShutdownTimeoutDuration,
		launchWorkerErr:                  make(chan error, 1),
		handlerTail:                      make(chan struct{}),
	}

	for _, opt := range opts {
		opt(mgr)
	}

	var err error
	mgr.triggerHandler, err = newTriggerHandler(mgr, mgr.globalLocker)
	if err != nil {
		return nil, fmt.Errorf("failed to create trigger handler: %w", err)
	}

	machineryLog.Set(newLoggerAdaptor())

	return mgr, nil
}

// OptionsFunc is a function that configures manager.
type OptionsFunc func(mgr *manager)

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
func WithRedis(redisConf config.Redis) OptionsFunc {
	return func(mgr *manager) {
		mgr.mConfig.Redis = &machineryConfig.RedisConfig{
			MaxIdle:                redisMaxIdle,
			IdleTimeout:            redisTimeout,
			ReadTimeout:            redisReadTimeout,
			WriteTimeout:           redisWriteTimeout,
			ConnectTimeout:         redisConnectTimeout,
			NormalTasksPollPeriod:  redisNormalTasksPollPeriod,
			DelayedTasksPollPeriod: redisDelayedTasksPollPeriod,
			MasterName:             redisConf.MasterName,
		}

		// NOTE: NewGR modifies the addrs slice in-place (extracts and removes password from first address).
		// We must create separate copies for broker and backend to avoid the bug where
		// broker removes the password, then backend receives addrs without password.
		// See: https://github.com/RichardKnop/machinery/issues/815
		brokerAddrs := buildAddrsWithPassword(redisConf)
		backendAddrs := buildAddrsWithPassword(redisConf)
		mgr.broker = brokerRedis.NewGR(mgr.mConfig, brokerAddrs, redisConf.DB)
		mgr.backend = backendRedis.NewGR(mgr.mConfig, backendAddrs, redisConf.DB)
	}
}

// buildAddrsWithPassword prepends password to the first address for NewGR.
// NewGR parses password from first address in format "password@host:port".
// For Redis Cluster mode, ensures at least 2 addresses so that go-redis
// NewUniversalClient creates a ClusterClient instead of a regular Client.
func buildAddrsWithPassword(redisConf config.Redis) []string {
	if len(redisConf.Addrs) == 0 {
		return nil
	}

	// Create a copy of the addresses to avoid modifying the original slice.
	result := make([]string, len(redisConf.Addrs))
	copy(result, redisConf.Addrs)

	// For Redis Cluster mode, go-redis NewUniversalClient requires >1 address
	// to create a ClusterClient. If only one address is provided, duplicate it.
	// This allows the ClusterClient to discover other nodes via CLUSTER SLOTS.
	if redisConf.Type == config.RedisTypeCluster && len(result) == 1 {
		result = append(result, result[0])
	}

	// Prepend password to the first address if present.
	// NewGR extracts password from "password@host:port" format.
	if redisConf.Password != "" {
		result[0] = redisConf.Password + "@" + result[0]
	}

	return result
}

// WithStorageTrigger sets the storage trigger for the manager.
func WithStorageTrigger(storageTrigger IStorageTrigger) OptionsFunc {
	return func(mgr *manager) {
		mgr.stgTrigger = storageTrigger
	}
}

// WithStorageOperation sets the storage operation for the manager.
func WithStorageOperation(storageOperation IStorageOperation) OptionsFunc {
	return func(mgr *manager) {
		mgr.stgOperation = storageOperation
	}
}

// WithStorageOperationInstance sets the storage operation instance for the manager.
func WithStorageOperationInstance(storageOperationInstance IStorageOperationInstance) OptionsFunc {
	return func(mgr *manager) {
		mgr.stgOperationInstance = storageOperationInstance
	}
}

// WithStorageActionInstance sets the storage action instance for the manager.
func WithStorageActionInstance(storageActionInst IStorageActionInstance) OptionsFunc {
	return func(mgr *manager) {
		mgr.stgActionInstance = storageActionInst
	}
}

// WithTraceService sets the trace service for the manager.
func WithTraceService(traceSvc tracing.IService) OptionsFunc {
	return func(mgr *manager) {
		mgr.traceSvc = traceSvc
	}
}

// WithLocker sets the locker for the manager.
func WithLocker(lock locker.MutexFactory) OptionsFunc {
	return func(mgr *manager) {
		mgr.globalLocker = lock
	}
}

// WithGracefulShutdownTimeout sets the graceful shutdown timeout for the manager.
func WithGracefulShutdownTimeout(timeout time.Duration) OptionsFunc {
	return func(mgr *manager) {
		if timeout <= 0 {
			return
		}

		mgr.gracefulShutdownTimeout = timeout
	}
}

type manager struct {
	mConfig *machineryConfig.Config

	// broker
	broker brokerIface.Broker

	// backend
	backend backendIface.Backend

	// WorkerNum defines the number of workers.
	WorkerNum int

	// gracefulShutdownTimeout defines how long running tasks can drain before manager context is cancelled.
	gracefulShutdownTimeout time.Duration

	// state
	isRunning   bool
	isConsuming bool
	handlerTail chan struct{}

	// context
	ctx    contextx.IContext
	cancel context.CancelFunc

	server *machinery.Server
	worker *machinery.Worker

	stgTrigger           IStorageTrigger
	stgOperation         IStorageOperation
	stgOperationInstance IStorageOperationInstance
	stgActionInstance    IStorageActionInstance

	globalLocker locker.MutexFactory

	registeredActionDefs             map[string]action.Definition
	registeredOperExtraExecutionDefs map[string]operation.ExtraExecution

	launchWorkerErr chan error

	// trigger handler.
	triggerHandler *triggerHandler

	// tracing service
	traceSvc tracing.IService
}

// Start starts the manager.
func (mgr *manager) Start(nCtx contextx.IContext) error {
	if mgr.isRunning {
		return errors.New("manager already started")
	}

	mgr.ctx, mgr.cancel = contextx.WithCancel(contextx.From(nCtx))

	if err := mgr.initialize(); err != nil {
		return err
	}

	// starts trigger handler.
	mgr.triggerHandler.Start()

	if mgr.WorkerNum > 0 {
		if err := mgr.launchWorker(); err != nil {
			return err
		}
	}

	return nil
}

// GracefulShutdown shuts down the manager gracefully.
func (mgr *manager) GracefulShutdown() error {
	if !mgr.isRunning {
		return errors.New("manager is not running")
	}

	mgr.isRunning = false
	mgr.isConsuming = false

	// Cancel is idempotent; defer covers the normal worker-exit path.
	defer mgr.cancel()

	timer := time.NewTimer(mgr.gracefulShutdownTimeout)
	defer timer.Stop()

	go func() {
		select {
		case <-timer.C:
			close(mgr.handlerTail)
		}
	}()

	mgr.worker.Quit()
	workerErr := <-mgr.launchWorkerErr
	if workerErr != nil {
		// notice: we don't use machinery.ErrWorkerQuitGracefully to check the error,
		// because we set NoUnixSignals to true in the manager config, so the worker will not return this error.
		return workerErr
	}

	return nil
}

// CheckHealth checks the health of the manager.
func (mgr *manager) CheckHealth() error {
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

// RegisterActions registers multiple OperInst operInstMgr actions.
func (mgr *manager) RegisterActions(actionDefs ...action.Definition) error {
	for _, actionDef := range actionDefs {
		if err := mgr.RegisterAction(actionDef); err != nil {
			return err
		}
	}

	return nil
}

// RegisterAction registers a OperationInst operInstMgr action.
func (mgr *manager) RegisterAction(actionDef action.Definition) error {
	if mgr.isRunning {
		return errors.New("operation instance manager already started, can not register action")
	}

	if _, ok := mgr.registeredActionDefs[actionDef.Name()]; ok {
		return fmt.Errorf("action already registered, action-name(%s)", actionDef.Name())
	}

	mgr.registeredActionDefs[actionDef.Name()] = actionDef

	return nil
}

// RegisterOperExtraExecutions registers a list of extra execution definitions.
func (mgr *manager) RegisterOperExtraExecutions(execDefs ...operation.ExtraExecution) error {
	if mgr.isRunning {
		return errors.New("operation instance manager already started, can not register operation execution")
	}

	for _, execDef := range execDefs {
		if _, ok := mgr.registeredOperExtraExecutionDefs[execDef.Name()]; ok {
			return errors.New("operation execution already registered")
		}

		mgr.registeredOperExtraExecutionDefs[execDef.Name()] = execDef
	}

	return nil
}

// initialize initializes the manager.
func (mgr *manager) initialize() error {
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

	if mgr.stgTrigger == nil {
		return errors.New("storage trigger is nil")
	}

	if mgr.stgOperation == nil {
		return errors.New("storage operation is nil")
	}

	if mgr.stgOperationInstance == nil {
		return errors.New("storage operation instance is nil")
	}

	if mgr.stgActionInstance == nil {
		return errors.New("storage action instance is nil")
	}

	if mgr.globalLocker == nil {
		return errors.New("global locker is nil")
	}

	// We don't need to use the periodic task of machinery, so there is no need to access the lock.
	mgr.server = machinery.NewServer(mgr.mConfig, mgr.broker, mgr.backend, nil)

	mgr.isRunning = true

	return nil
}
