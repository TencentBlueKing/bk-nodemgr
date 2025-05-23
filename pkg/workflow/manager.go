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

	"github.com/RichardKnop/machinery/v2"
	backendIface "github.com/RichardKnop/machinery/v2/backends/iface"
	backendRedis "github.com/RichardKnop/machinery/v2/backends/redis"
	brokerIface "github.com/RichardKnop/machinery/v2/brokers/iface"
	brokerRedis "github.com/RichardKnop/machinery/v2/brokers/redis"
	machineryConfig "github.com/RichardKnop/machinery/v2/config"
	machinerylog "github.com/RichardKnop/machinery/v2/log"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/locker"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

// IManager defines the manager interface.
type IManager interface {
	IController
	IWorker

	// Start starts the worker.
	Start(ctx context.Context) error

	// CheckHealth checks the health of the worker.
	CheckHealth() error

	// GracefulShutdown gracefully shutdowns the worker.
	GracefulShutdown() error

	// RegisterActions registers a list of actions.
	RegisterActions(actionDefs ...action.Definition) error
}

const (
	queueNameDefault       = "operation_inst_engine_queue"
	resultsExpireInDefault = 3600
)

// NewManager creates a new manager.
func NewManager(workerNum int, opts ...OptionsFunc) IManager {
	mgr := &manager{
		mConfig: &machineryConfig.Config{
			DefaultQueue:    queueNameDefault,
			ResultsExpireIn: resultsExpireInDefault,
			NoUnixSignals:   true,
		},
		isRunning:            false,
		registeredActionDefs: make(map[string]action.Definition),
		logger:               logger.LoggerDefault{},
		WorkerNum:            workerNum,
		launchWorkerErr:      make(chan error, 1),
	}

	for _, opt := range opts {
		opt(mgr)
	}

	mgr.triggerHandler = newTriggerHandler(mgr, mgr.globalLocker)
	machinerylog.Set(newLoggerAdaptor(mgr.logger))

	return mgr
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
func WithRedis(address, password string, db int) OptionsFunc {
	return func(mgr *manager) {
		mgr.mConfig.Redis = &machineryConfig.RedisConfig{
			MaxIdle:                redisMaxIdle,
			IdleTimeout:            redisTimeout,
			ReadTimeout:            redisReadTimeout,
			WriteTimeout:           redisWriteTimeout,
			ConnectTimeout:         redisConnectTimeout,
			NormalTasksPollPeriod:  redisNormalTasksPollPeriod,
			DelayedTasksPollPeriod: redisDelayedTasksPollPeriod,
		}

		mgr.broker = brokerRedis.New(mgr.mConfig, address, password, "", db)
		mgr.backend = backendRedis.New(mgr.mConfig, address, password, "", db)
	}
}

// WithStorageTrigger sets the storage trigger for the manager.
func WithStorageTrigger(storageTrigger IStorageTrigger) OptionsFunc {
	return func(mgr *manager) {
		mgr.storageTrigger = storageTrigger
	}
}

// WithStorageOperation sets the storage operation for the manager.
func WithStorageOperation(storageOperation IStorageOperation) OptionsFunc {
	return func(mgr *manager) {
		mgr.storageOperation = storageOperation
	}
}

// WithStorageOperationInstance sets the storage operation instance for the manager.
func WithStorageOperationInstance(storageOperationInstance IStorageOperationInstance) OptionsFunc {
	return func(mgr *manager) {
		mgr.storageOperationInstance = storageOperationInstance
	}
}

// WithStorageActionInstance sets the storage action instance for the manager.
func WithStorageActionInstance(storageActionInst IStorageActionInstance) OptionsFunc {
	return func(mgr *manager) {
		mgr.storageActionInstance = storageActionInst
	}
}

// WithLocker sets the locker for the manager.
func WithLocker(lock locker.MutexFactory) OptionsFunc {
	return func(mgr *manager) {
		mgr.globalLocker = lock
	}
}

// WithLogger sets the logger for the manager.
func WithLogger(logger logger.Logger) OptionsFunc {
	return func(mgr *manager) {
		mgr.logger = logger
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

	// state
	isRunning   bool
	isConsuming bool

	// context
	ctx    context.Context
	cancel context.CancelFunc

	server *machinery.Server
	worker *machinery.Worker

	storageTrigger           IStorageTrigger
	storageOperation         IStorageOperation
	storageOperationInstance IStorageOperationInstance
	storageActionInstance    IStorageActionInstance

	logger logger.Logger

	globalLocker locker.MutexFactory

	registeredActionDefs map[string]action.Definition

	launchWorkerErr chan error

	// trigger handler.
	triggerHandler *triggerHandler
}

// Start starts the manager.
func (mgr *manager) Start(ctx context.Context) error {
	if mgr.isRunning {
		return errors.New("manager already started")
	}

	mgr.ctx, mgr.cancel = context.WithCancel(ctx)

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

	if mgr.isConsuming {
		go mgr.broker.StopConsuming()
	}

	mgr.isRunning = false
	mgr.isConsuming = false

	defer mgr.cancel()

	err := <-mgr.launchWorkerErr
	if !errors.Is(err, machinery.ErrWorkerQuitGracefully) {
		return err
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
		return errors.New("action already registered")
	}

	mgr.registeredActionDefs[actionDef.Name()] = actionDef

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

	if mgr.storageTrigger == nil {
		return errors.New("storage trigger is nil")
	}

	if mgr.storageOperation == nil {
		return errors.New("storage operation is nil")
	}

	if mgr.storageOperationInstance == nil {
		return errors.New("storage operation instance is nil")
	}

	if mgr.storageActionInstance == nil {
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
