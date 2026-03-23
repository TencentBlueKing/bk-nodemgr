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
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/goasync"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/locker"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/scheduler"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/common"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/trigger"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func newCachedTriggers() *cachedTriggers {
	return &cachedTriggers{
		triggers: make([]*trigger.Trigger, 0),
	}
}

// cachedTriggers manages local trigger synced from storage.
type cachedTriggers struct {
	mutex    sync.RWMutex
	triggers []*trigger.Trigger
}

func (ct *cachedTriggers) set(triggers []*trigger.Trigger) {
	ct.mutex.Lock()
	defer ct.mutex.Unlock()

	ct.triggers = make([]*trigger.Trigger, len(triggers))
	copy(ct.triggers, triggers)
}

func (ct *cachedTriggers) get() []*trigger.Trigger {
	ct.mutex.RLock()
	defer ct.mutex.RUnlock()

	result := make([]*trigger.Trigger, len(ct.triggers))
	copy(result, ct.triggers)

	return result
}

const (
	triggerHandlerGoAsyncPoolNum         = 10000
	triggerHandlerGoAsyncPoolPerPoolSize = 10000

	// instantiateOperationConcurrency limits concurrent MongoDB writes when creating operation instances.
	instantiateOperationConcurrency = 20
	// launchOperationInstanceConcurrency limits concurrent MongoDB writes when launching operation instances.
	launchOperationInstanceConcurrency = 20
)

func newTriggerHandler(mgr *manager, globalLocker locker.MutexFactory) (*triggerHandler, error) {
	trigHandler := &triggerHandler{
		mgr:          mgr,
		globalLocker: globalLocker,

		onceTriggers:     newCachedTriggers(),
		orderedTriggers:  newCachedTriggers(),
		periodicTriggers: newCachedTriggers(),

		tracerProvider: mgr.traceSvc.TracerProvider(),
	}

	var err error
	trigHandler.goAsyncPool, err = goasync.NewHandler(goasync.HandlerOption{
		PoolNum:               triggerHandlerGoAsyncPoolNum,
		PerPoolSize:           triggerHandlerGoAsyncPoolPerPoolSize,
		LoadBalancingStrategy: goasync.LoadBalancingStrategyLeastFirst,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create goasync handler: %w", err)
	}

	return trigHandler, nil
}

type triggerHandler struct {
	mgr          *manager
	globalLocker locker.MutexFactory

	scheduler scheduler.Scheduler

	onceTriggers     *cachedTriggers
	orderedTriggers  *cachedTriggers
	periodicTriggers *cachedTriggers

	goAsyncPool goasync.IHandler

	tracerProvider trace.TracerProvider
}

// Start starts the manager.
func (handler *triggerHandler) Start() {
	if handler.scheduler != nil {
		handler.scheduler.Terminate()
	}

	handler.scheduler = scheduler.NewScheduler()
	handler.initSchedulerTasks()
	handler.scheduler.Start()
}

// Stop stops the manager.
func (handler *triggerHandler) Stop() {
	if handler.scheduler != nil {
		handler.scheduler.Terminate()
	}
}

const (
	defaultTimeout = 1 * time.Minute

	onceTriggersSyncAndCheckIntervalDefault     = 1 * time.Second
	orderedTriggersSyncAndCheckIntervalDefault  = 1 * time.Second
	periodicTriggersSyncAndCheckIntervalDefault = 1 * time.Second

	taskIDSyncAndCheckOnceTrigger     = "sync_and_check_once_trigger"
	taskIDSyncAndCheckOrderedTrigger  = "sync_and_check_ordered_trigger"
	taskIDSyncAndCheckPeriodicTrigger = "sync_and_check_periodic_trigger"
)

func (handler *triggerHandler) initSchedulerTasks() {
	// init schedule tasks.
	// sync first and then check to avoid the situation where deleted triggers are still being checked
	// nolint: contextcheck
	scheduleTasks := []*scheduler.Task{
		scheduler.NewTask(
			taskIDSyncAndCheckOnceTrigger,
			onceTriggersSyncAndCheckIntervalDefault,
			defaultTimeout,
			func(nCtx contextx.IContext) error {
				if err := handler.syncOnceTrigger(nCtx); err != nil {
					return err
				}

				if err := handler.executeTriggerList(nCtx, handler.onceTriggers.get()); err != nil {
					return err
				}

				return nil
			},
		),
		scheduler.NewTask(
			taskIDSyncAndCheckOrderedTrigger,
			orderedTriggersSyncAndCheckIntervalDefault,
			defaultTimeout,
			func(nCtx contextx.IContext) error {
				if err := handler.syncOrderedTrigger(nCtx); err != nil {
					return err
				}

				if err := handler.executeTriggerList(nCtx, handler.orderedTriggers.get()); err != nil {
					return err
				}

				return nil
			},
		),
		scheduler.NewTask(
			taskIDSyncAndCheckPeriodicTrigger,
			periodicTriggersSyncAndCheckIntervalDefault,
			defaultTimeout,
			func(nCtx contextx.IContext) error {
				if err := handler.syncPeriodicTrigger(nCtx); err != nil {
					return err
				}

				if err := handler.executeTriggerList(nCtx, handler.periodicTriggers.get()); err != nil {
					return err
				}

				return nil
			},
		),
	}
	for _, task := range scheduleTasks {
		err := handler.scheduler.RegisterTask(task)
		if err != nil {
			logger.G.Sys().WithErr(err).With("task-id", task.ID).Error("failed to register task")

			continue
		}
	}
}

// syncOnceTrigger syncs once triggers from storage.
func (handler *triggerHandler) syncOnceTrigger(nCtx contextx.IContext) error {
	list, err := handler.mgr.stgTrigger.ListActiveTrigger(nCtx, trigger.CategoryOnce)
	if err != nil {
		// set cached triggers to empty cause the cache is no longer valid.
		handler.onceTriggers.set([]*trigger.Trigger{})

		return err
	}

	logger.G.Sys().With("count", len(list)).Debug("synced once triggers")

	handler.onceTriggers.set(list)

	return nil
}

// syncOrderedTrigger syncs ordered triggers from storage.
func (handler *triggerHandler) syncOrderedTrigger(nCtx contextx.IContext) error {
	list, err := handler.mgr.stgTrigger.ListActiveTrigger(nCtx, trigger.CategoryOrdered)
	if err != nil {
		// set cached triggers to empty cause the cache is no longer valid.
		handler.orderedTriggers.set([]*trigger.Trigger{})

		return err
	}

	logger.G.Sys().With("count", len(list)).Debug("synced ordered triggers")

	handler.orderedTriggers.set(list)

	return nil
}

// syncPeriodicTrigger syncs periodic triggers from storage.
func (handler *triggerHandler) syncPeriodicTrigger(nCtx contextx.IContext) error {
	list, err := handler.mgr.stgTrigger.ListActiveTrigger(nCtx, trigger.CategoryPeriodic)
	if err != nil {
		// set cached triggers to empty cause the cache is no longer valid.
		handler.periodicTriggers.set([]*trigger.Trigger{})

		return err
	}

	logger.G.Sys().With("count", len(list)).Debug("synced periodic triggers")

	handler.periodicTriggers.set(list)

	return nil
}

// executeTriggerList executes the trigger list.
func (handler *triggerHandler) executeTriggerList(nCtx contextx.IContext, list []*trigger.Trigger) error {
	logger.G.Sys().With("count", len(list)).Debug("check trigger list")

	for idx := range list {
		trig := list[idx]
		fn := func(nCtx contextx.IContext) error {
			mutex := handler.globalLocker.NewMutex(trig.TriggerID)
			if err := mutex.TryLock(); err != nil {
				logger.G.Sys().WithErr(err).With("trigger-id", trig.TriggerID).Debug("failed to lock trigger")

				return nil
			}

			logger.G.Sys().With("trigger-id", trig.TriggerID).Debug("got trigger lock")

			defer func() {
				_ = mutex.Unlock()
			}()

			// get trigger from storage after get lock.
			// make sure the trigger data is fresh.
			trigCtl, err := handler.mgr.GetTrigger(nCtx, trig.TriggerID)
			if err != nil {
				logger.G.Sys().WithErr(err).With("trigger-id", trig.TriggerID).Error("failed to get trigger")

				return nil
			}

			if err := handler.doTrigger(nCtx, trigCtl); err != nil {
				logger.G.Sys().WithErr(err).With("trigger-id", trig.TriggerID).Error("failed to do trigger")

				return nil
			}

			return nil
		}

		if err := handler.goAsyncPool.Run(nCtx, fn,
			goasync.WithName("executeTriggerList"),
		); err != nil {
			return fmt.Errorf("failed to push trigger fn to async pool: %w", err)
		}
	}

	return nil
}

func (handler *triggerHandler) doTrigger(nCtx contextx.IContext, trigCtl ITriggerCtl) error {
	traceCtx, span := handler.tracerProvider.Tracer(scopeNameTrigger).Start(nCtx,
		fmt.Sprintf("%s %s", spanNamePrefixTrigger, trigCtl.GetTriggerCategory()),
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(
			attribute.String(attributeKeyTriggerID, trigCtl.GetTriggerID()),
			attribute.String(attributeKeyTriggerCategory, string(trigCtl.GetTriggerCategory())),
		))
	defer span.End()

	// set trace context
	nCtx = contextx.FromContext(traceCtx)

	logger.G.Sys().With("trigger-id", trigCtl.GetTriggerID(), "category", trigCtl.GetTriggerCategory()).Debug("do trigger")

	switch trigCtl.GetTriggerCategory() {
	case trigger.CategoryOnce:
		instanceCtls, err := handler.doOnceTrigger(nCtx, trigCtl)
		if err != nil {
			return err
		}

		if err := handler.launchOperationInstance(nCtx, trigCtl, instanceCtls); err != nil {
			return err
		}

		// inactivate once trigger after processing
		if err := trigCtl.TryInactivateTrigger(nCtx); err != nil {
			return err
		}

		logger.G.Sys().With("trigger-id", trigCtl.GetTriggerID()).Debug("launched operation instance list, inactivated once trigger")

		return nil

	case trigger.CategoryOrdered:
		instanceCtls, err := handler.doOrderedTrigger(nCtx, trigCtl)
		if err != nil {
			return err
		}

		if err := handler.launchOperationInstance(nCtx, trigCtl, instanceCtls); err != nil {
			return err
		}

		// inactivate once trigger after processing
		if err := trigCtl.TryInactivateTrigger(nCtx); err != nil {
			return err
		}

		logger.G.Sys().With("trigger-id", trigCtl.GetTriggerID()).Debug("launched operation instance list, inactivated ordered trigger")

		return nil

	case trigger.CategoryPeriodic:
		instanceCtls, err := handler.doPeriodicTrigger(nCtx, trigCtl)
		if err != nil {
			return err
		}

		if err := handler.launchOperationInstance(nCtx, trigCtl, instanceCtls); err != nil {
			return err
		}

		return nil

	default:
		return common.ErrUnknownTriggerCategory()
	}
}

// instantiateOperation instantiates operations for the trigger.
func (handler *triggerHandler) instantiateOperation(nCtx contextx.IContext, trigCtl ITriggerCtl, page types.Page) error {
	operList, err := trigCtl.ListNeedInstantiateOperation(nCtx, page)
	if err != nil {
		return err
	}

	logger.G.Sys().With("trigger-id", trigCtl.GetTriggerID(), "operation-count", len(operList)).Debug("instantiate operation")

	gp := gopool.NewPool()
	gp.SetLimit(instantiateOperationConcurrency)
	for _, operCtl := range operList {
		ctl := operCtl
		gp.Go(func() error {
			operInst, err := ctl.CreateOperationInstance(nCtx)
			if err != nil {
				logger.G.Sys().
					WithErr(err).
					With("trigger-id", trigCtl.GetTriggerID(), "operation-id", ctl.GetOperationID()).
					Error("failed to create operation instance")

				return err
			}

			logger.G.Sys().
				With("trigger-id", trigCtl.GetTriggerID(), "operation-id", ctl.GetOperationID(), "oper-inst-id", operInst.GetOperationInstanceID()).
				Debug("created operation instance")

			return nil
		})
	}

	return gp.Wait()
}

const (
	// onceTriggerBatchSize limits the number of operations instantiated and launched per cycle
	// to avoid overwhelming MongoDB with too many concurrent writes when a trigger has a large backlog.
	onceTriggerBatchSize = 200
)

func (handler *triggerHandler) doOnceTrigger(nCtx contextx.IContext, trigCtl ITriggerCtl) ([]IOperationInstanceCtl, error) {
	if err := handler.instantiateOperation(nCtx, trigCtl, types.Page{Limit: onceTriggerBatchSize}); err != nil {
		logger.G.Sys().
			WithErr(err).
			With("trigger-id", trigCtl.GetTriggerID()).
			Warn("failed to init once empty operation")
	}

	instanceList, err := trigCtl.ListOperationInstances(nCtx, types.UnlimitedPage(), operation.StateInit)
	if err != nil {
		return nil, err
	}

	logger.G.Sys().With("trigger-id", trigCtl.GetTriggerID(), "instance-count", len(instanceList)).
		Debug("once trigger processed")

	return instanceList, nil
}

func (handler *triggerHandler) doOrderedTrigger(nCtx contextx.IContext, trigCtl ITriggerCtl) ([]IOperationInstanceCtl, error) {
	metadata, ok := trigCtl.GetTriggerMetadata().(*trigger.MetadataOrdered)
	if !ok {
		return nil, errors.Join(common.ErrInvalidTriggerMetadata(),
			fmt.Errorf("trigger metadata is not ordered type. trigger-id(%s)", trigCtl.GetTriggerID()))
	}

	if metadata.MaxConcurrencyNum <= 0 {
		return nil, errors.Join(common.ErrInvalidTriggerMetadata(),
			fmt.Errorf("max concurrency num is invalid. trigger-id(%s), max-concurrency-num(%d)",
				trigCtl.GetTriggerID(), metadata.MaxConcurrencyNum))
	}

	workingCount, err := handler.mgr.stgOperationInstance.CountOperationInstanceByState(nCtx, trigCtl.GetTriggerID(),
		operation.StateInit, operation.StateLaunched, operation.StateRunning)
	if err != nil {
		return nil, err
	}

	// not idle concurrent num.
	idleNum := metadata.MaxConcurrencyNum - int(workingCount)
	if idleNum <= 0 {
		return nil, nil
	}

	if err := handler.instantiateOperation(nCtx, trigCtl, types.Page{Limit: idleNum}); err != nil {
		logger.G.Sys().
			WithErr(err).
			With("trigger-id", trigCtl.GetTriggerID()).
			Warn("failed to init ordered empty operation")
	}

	instanceList, err := trigCtl.ListOperationInstances(nCtx, types.Page{Limit: idleNum}, operation.StateInit)
	if err != nil {
		return nil, err
	}

	logger.G.Sys().With("trigger-id", trigCtl.GetTriggerID(),
		"working-count", workingCount,
		"idle-num", idleNum,
		"instantiated-count", len(instanceList)).
		Debug("ordered trigger processed")

	return instanceList, nil
}

const (
	periodicTriggerMaxSleepTime = 1 * time.Minute
)

func (handler *triggerHandler) doPeriodicTrigger(nCtx contextx.IContext, trigCtl ITriggerCtl) ([]IOperationInstanceCtl, error) {
	metadata, ok := trigCtl.GetTriggerMetadata().(*trigger.MetadataPeriodic)
	if !ok {
		return nil, errors.Join(common.ErrInvalidTriggerMetadata(),
			fmt.Errorf("trigger metadata is not periodic type. trigger-id(%s)", trigCtl.GetTriggerID()))
	}

	// not enough interval yet.
	nextTime, err := scheduler.NextActiveTime(metadata.Interval, trigCtl.GetLastTriggeredAt())
	if err != nil {
		return nil, err
	}

	if nextTime.After(time.Now()) {
		sleepTime := time.Until(nextTime)
		sleepTime = min(sleepTime, periodicTriggerMaxSleepTime)
		time.Sleep(sleepTime)

		return nil, nil
	}

	if !metadata.AllowedConcurrency {
		workingCount, err := handler.mgr.stgOperationInstance.CountOperationInstanceByState(nCtx, trigCtl.GetTriggerID(),
			operation.StateInit, operation.StateLaunched, operation.StateRunning)
		if err != nil {
			return nil, err
		}

		// do not allow concurrency.
		if workingCount > 0 {
			return nil, nil
		}
	}

	logger.G.Sys().With("trigger-id", trigCtl.GetTriggerID()).
		Debug("periodic trigger next activation time reached, no working instance, proceed to create operation instance")

	operList, count, err := handler.mgr.stgOperation.ListOperationByTriggerID(nCtx, types.UnlimitedPage(), trigCtl.GetTriggerID())
	if err != nil {
		return nil, err
	}

	if count != 1 || len(operList) != 1 {
		return nil, errors.Join(common.ErrInvalidPeriodicOperationNum(),
			fmt.Errorf("periodic trigger should only have one operation. trigger-id(%s), operation-count(%d)",
				trigCtl.GetTriggerID(), count))
	}

	oper := operList[0]
	operCtl, err := trigCtl.GetOperation(nCtx, oper.OperationID)
	if err != nil {
		return nil, err
	}

	operInstCtl, err := operCtl.CreateOperationInstance(nCtx)
	if err != nil {
		return nil, err
	}

	logger.G.Sys().With("trigger-id", trigCtl.GetTriggerID(),
		"oper-id", operCtl.GetOperationID(),
		"oper-inst-id", operInstCtl.GetOperationInstanceID()).
		Debug("created periodic operation instance")

	return []IOperationInstanceCtl{operInstCtl}, nil
}

func (handler *triggerHandler) launchOperationInstance(nCtx contextx.IContext, trigCtl ITriggerCtl, instanceCtls []IOperationInstanceCtl) error {
	gp := gopool.NewPool()
	gp.SetLimit(launchOperationInstanceConcurrency)
	for _, instanceCtl := range instanceCtls {
		ctl := instanceCtl
		gp.Go(func() error {
			if err := ctl.LaunchOperationInstance(nCtx); err != nil {
				logger.G.Sys().
					WithErr(err).
					With("trigger-id", trigCtl.GetTriggerID(), "oper-inst-id", ctl.GetOperationInstanceID()).
					Error("failed to launch operation instance")

				return err
			}

			logger.G.Sys().
				With("trigger-id", trigCtl.GetTriggerID(), "oper-inst-id", ctl.GetOperationInstanceID()).
				Info("launched operation instance")

			return nil
		})
	}

	// update triggered time if there is any instance launched.
	if len(instanceCtls) > 0 {
		if err := trigCtl.UpdateLastTriggeredTime(nCtx); err != nil {
			logger.G.Sys().WithErr(err).With("trigger-id", trigCtl.GetTriggerID()).Error("failed to update last triggered time")
		}
	}

	return gp.Wait()
}
