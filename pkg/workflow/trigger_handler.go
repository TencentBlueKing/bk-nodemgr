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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/locker"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/scheduler"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/common"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/trigger"
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

func newTriggerHandler(mgr *manager, globalLocker locker.MutexFactory) *triggerHandler {
	return &triggerHandler{
		mgr:          mgr,
		globalLocker: globalLocker,

		onceTriggers:     newCachedTriggers(),
		orderedTriggers:  newCachedTriggers(),
		periodicTriggers: newCachedTriggers(),
	}
}

type triggerHandler struct {
	mgr          *manager
	globalLocker locker.MutexFactory

	scheduler scheduler.Scheduler

	onceTriggers     *cachedTriggers
	orderedTriggers  *cachedTriggers
	periodicTriggers *cachedTriggers
}

// Start starts the manager.
func (handler *triggerHandler) Start() {
	if handler.scheduler != nil {
		handler.scheduler.Terminate()
	}

	handler.scheduler = scheduler.NewScheduler(scheduler.WithLogger(handler.mgr.logger))
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

	maxOnceTriggerProcessLimit = 500

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
			func(ctx contextx.IContext) error {
				if err := handler.syncOnceTrigger(ctx); err != nil {
					return err
				}

				if err := handler.checkTriggerList(ctx, handler.onceTriggers.get()); err != nil {
					return err
				}

				return nil
			},
		),
		scheduler.NewTask(
			taskIDSyncAndCheckOrderedTrigger,
			orderedTriggersSyncAndCheckIntervalDefault,
			defaultTimeout,
			func(ctx contextx.IContext) error {
				if err := handler.syncOrderedTrigger(ctx); err != nil {
					return err
				}

				if err := handler.checkTriggerList(ctx, handler.orderedTriggers.get()); err != nil {
					return err
				}

				return nil
			},
		),
		scheduler.NewTask(
			taskIDSyncAndCheckPeriodicTrigger,
			periodicTriggersSyncAndCheckIntervalDefault,
			defaultTimeout,
			func(ctx contextx.IContext) error {
				if err := handler.syncPeriodicTrigger(ctx); err != nil {
					return err
				}

				if err := handler.checkTriggerList(ctx, handler.periodicTriggers.get()); err != nil {
					return err
				}

				return nil
			},
		),
	}
	for _, task := range scheduleTasks {
		err := handler.scheduler.RegisterTask(task)
		if err != nil {
			handler.mgr.logger.Errorf("failed to register task, task-id(%s), err: %v", task.ID, err)
			continue
		}
	}
}

// syncOnceTrigger syncs once triggers from storage.
func (handler *triggerHandler) syncOnceTrigger(ctx contextx.IContext) error {
	list, err := handler.mgr.stgTrigger.ListAliveTrigger(ctx, trigger.CategoryOnce)
	if err != nil {
		// set cached triggers to empty cause the cache is no longer valid.
		handler.onceTriggers.set([]*trigger.Trigger{})

		return err
	}

	handler.mgr.logger.Debugf("synced once triggers. count: %d", len(list))

	handler.onceTriggers.set(list)

	return nil
}

// syncOrderedTrigger syncs ordered triggers from storage.
func (handler *triggerHandler) syncOrderedTrigger(ctx contextx.IContext) error {
	list, err := handler.mgr.stgTrigger.ListAliveTrigger(ctx, trigger.CategoryOrdered)
	if err != nil {
		// set cached triggers to empty cause the cache is no longer valid.
		handler.orderedTriggers.set([]*trigger.Trigger{})

		return err
	}

	handler.mgr.logger.Debugf("synced ordered triggers. count: %d", len(list))

	handler.orderedTriggers.set(list)

	return nil
}

// syncPeriodicTrigger syncs periodic triggers from storage.
func (handler *triggerHandler) syncPeriodicTrigger(ctx contextx.IContext) error {
	list, err := handler.mgr.stgTrigger.ListAliveTrigger(ctx, trigger.CategoryPeriodic)
	if err != nil {
		// set cached triggers to empty cause the cache is no longer valid.
		handler.periodicTriggers.set([]*trigger.Trigger{})

		return err
	}

	handler.mgr.logger.Debugf("synced periodic triggers. count: %d", len(list))

	handler.periodicTriggers.set(list)

	return nil
}

// defaultCheckConcurrency defines the default check concurrency.
const defaultCheckConcurrency = 100

// checkTriggerList checks trigger list and executes triggers.
func (handler *triggerHandler) checkTriggerList(ctx contextx.IContext, list []*trigger.Trigger) error {
	handler.mgr.logger.DebugCtxf(ctx, "check trigger list. count: %d", len(list))

	gp := gopool.NewPool()
	gp.SetLimit(defaultCheckConcurrency)

	for idx := range list {
		trig := list[idx]
		fn := func() error {
			handler.mgr.logger.DebugCtxf(ctx, "try lock trigger. trigger-id:(%s)", trig.TriggerID)

			mutex := handler.tryLockTrigger(ctx, trig)
			if mutex == nil {
				return nil
			}
			defer func() {
				_ = mutex.Unlock()
			}()

			handler.mgr.logger.DebugCtxf(ctx, "check trigger. trigger-id:(%s)", trig.TriggerID)

			// get trigger from storage after get lock.
			// make sure the trigger data is fresh.
			trigCtl, err := handler.mgr.GetTrigger(ctx, trig.TriggerID)
			if err != nil {
				handler.mgr.logger.ErrorCtxf(ctx, "failed to get trigger. trigger-id:(%s), err: %v", trig.TriggerID, err)

				return nil
			}

			if err := handler.doTrigger(ctx, trigCtl); err != nil {
				handler.mgr.logger.ErrorCtxf(ctx, "failed to do trigger. trigger-id:(%s), err: %v", trig.TriggerID, err)

				return nil
			}

			return nil
		}

		gp.Go(fn)
	}

	if err := gp.Wait(); err != nil {
		handler.mgr.logger.Errorf("check periodic trigger failed, err: %v", err)
		return err
	}

	return nil
}

func (handler *triggerHandler) tryLockTrigger(ctx contextx.IContext, trig *trigger.Trigger) locker.Mutex {
	mutex := handler.globalLocker.NewMutex(trig.TriggerID)

	handler.mgr.logger.DebugCtxf(ctx, "try lock trigger. trigger-id:(%s)", trig.TriggerID)

	err := mutex.TryLock()
	if err != nil {
		handler.mgr.logger.ErrorCtxf(ctx, "failed to lock trigger. trigger-id:(%s), err: %v", trig.TriggerID, err)

		return nil
	}

	if handler.checkFeasibility(ctx, trig) != nil {
		_ = mutex.Unlock()

		handler.mgr.logger.DebugCtxf(ctx, "trigger is not feasible. trigger-id:(%s), state(%s)",
			trig.TriggerID, trig.State)

		return nil
	}

	return mutex
}

func (handler *triggerHandler) checkFeasibility(_ contextx.IContext, trig *trigger.Trigger) error {
	if trig.State != trigger.StateRunning {
		return common.ErrTriggerNotRunning()
	}

	switch trig.Category {
	case trigger.CategoryOnce:
		return nil
	case trigger.CategoryOrdered:
		return nil

	case trigger.CategoryPeriodic:
		metadata, ok := trig.Metadata.(*trigger.MetadataPeriodic)
		if !ok {
			return errors.Join(common.ErrInvalidTriggerMetadata(),
				fmt.Errorf("trigger metadata is not periodic type. trigger-id(%s)", trig.TriggerID))
		}

		// not enough interval yet.
		nextTime, err := scheduler.NextActiveTime(metadata.Interval, trig.LastTriggeredAt)
		if err != nil {
			return err
		}

		if nextTime.After(time.Now()) {
			return common.ErrTriggerNotReady()
		}

		return nil

	default:
		return nil
	}
}

func (handler *triggerHandler) doTrigger(ctx contextx.IContext, trigCtl ITriggerCtl) error {
	var instanceCtls []IOperationInstanceCtl
	var err error

	switch trigCtl.GetTriggerCategory() {
	case trigger.CategoryOnce:
		instanceCtls, err = handler.doOnceTrigger(ctx, trigCtl)
		if err != nil {
			return err
		}

	case trigger.CategoryOrdered:
		instanceCtls, err = handler.doOrderedTrigger(ctx, trigCtl)
		if err != nil {
			return err
		}

	case trigger.CategoryPeriodic:
		instanceCtls, err = handler.doPeriodicTrigger(ctx, trigCtl)
		if err != nil {
			return err
		}

	default:
		return common.ErrUnknownTriggerCategory()
	}

	gp := gopool.NewPool()
	for _, instanceCtl := range instanceCtls {
		ctl := instanceCtl
		gp.Go(func() error {
			if err := ctl.LaunchOperationInstance(ctx); err != nil {
				handler.mgr.logger.ErrorCtxf(ctx, "failed to launch operation instance. trigger-id(%s), oper-inst-id(%s), err(%s)",
					trigCtl.GetTriggerID(), ctl.GetOperationInstanceID(), err.Error())

				return err
			}

			handler.mgr.logger.InfoCtxf(ctx, "launched operation instance. trigger-id(%s), oper-inst-id(%s)",
				trigCtl.GetTriggerID(), ctl.GetOperationInstanceID())

			return nil
		})
	}

	// update triggered time if there is any instance launched.
	if len(instanceCtls) > 0 {
		if err = trigCtl.UpdateLastTriggeredTime(ctx); err != nil {
			handler.mgr.logger.WarnCtxf(ctx, "failed to update last triggered time. trigger-id:(%s), err: %v",
				trigCtl.GetTriggerID(), err)
		}
	}

	return gp.Wait()
}

func (handler *triggerHandler) initEmptyOperation(ctx contextx.IContext, trigCtl ITriggerCtl, limit int) error {
	operList, err := trigCtl.ListEmptyOperation(ctx, types.Page{Limit: limit})
	if err != nil {
		return err
	}

	handler.mgr.logger.DebugCtxf(ctx,
		"init empty operation, list empty operation(%d). trigger-id(%s), operation-count(%d)",
		len(operList), trigCtl.GetTriggerID(), len(operList))

	gp := gopool.NewPool()
	for _, operCtl := range operList {
		ctl := operCtl
		gp.Go(func() error {
			if _, err := ctl.CreateOperationInstance(ctx); err != nil {
				handler.mgr.logger.ErrorCtxf(ctx, "failed to create operation instance. trigger-id(%s), operation-id(%s), err(%v)",
					trigCtl.GetTriggerID(), ctl.GetOperationID(), err)

				return err
			}

			return nil
		})
	}

	return gp.Wait()
}

func (handler *triggerHandler) doOnceTrigger(
	ctx contextx.IContext, trigCtl ITriggerCtl) ([]IOperationInstanceCtl, error) {

	if err := handler.initEmptyOperation(ctx, trigCtl, maxOnceTriggerProcessLimit); err != nil {
		handler.mgr.logger.WarnCtxf(ctx, "failed to init empty operation. trigger-id(%s), err(%v)",
			trigCtl.GetTriggerID(), err)
	}

	instanceList, err := trigCtl.ListOperationInstances(
		ctx, types.Page{Limit: maxOnceTriggerProcessLimit}, operation.StateInit)
	if err != nil {
		return nil, err
	}

	return instanceList, nil
}

func (handler *triggerHandler) doOrderedTrigger(
	ctx contextx.IContext, trigCtl ITriggerCtl) ([]IOperationInstanceCtl, error) {

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

	workingCount, err := handler.mgr.stgOperationInstance.CountOperationInstance(ctx, trigCtl.GetTriggerID(),
		operation.StateLaunched, operation.StateRunning)
	if err != nil {
		return nil, err
	}

	// not idle concurrent num.
	idleNum := metadata.MaxConcurrencyNum - int(workingCount)
	if idleNum <= 0 {
		return nil, nil
	}

	if err := handler.initEmptyOperation(ctx, trigCtl, idleNum); err != nil {
		handler.mgr.logger.WarnCtxf(ctx, "failed to init empty operation. trigger-id(%s), err(%v)",
			trigCtl.GetTriggerID(), err)
	}

	instanceList, err := trigCtl.ListOperationInstances(ctx, types.Page{Limit: idleNum}, operation.StateInit)
	if err != nil {
		return nil, err
	}

	return instanceList, nil
}

func (handler *triggerHandler) doPeriodicTrigger(
	ctx contextx.IContext, trigCtl ITriggerCtl) ([]IOperationInstanceCtl, error) {

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
		return nil, nil
	}

	if !metadata.AllowedConcurrency {
		workingCount, err := handler.mgr.stgOperationInstance.CountOperationInstance(ctx, trigCtl.GetTriggerID(),
			operation.StateLaunched, operation.StateRunning)
		if err != nil {
			return nil, err
		}

		// do not allow concurrency.
		if workingCount > 0 {
			return nil, nil
		}
	}

	operList, count, err := handler.mgr.stgOperation.ListOperationByTrigger(
		ctx, types.UnlimitedPage(), trigCtl.GetTriggerID())
	if err != nil {
		return nil, err
	}

	if count != 1 || len(operList) != 1 {
		return nil, errors.Join(common.ErrInvalidPeriodicOperationNum(),
			fmt.Errorf("periodic trigger should only have one operation. trigger-id(%s), operation-count(%d)",
				trigCtl.GetTriggerID(), count))
	}

	oper := operList[0]
	operCtl, err := trigCtl.GetOperation(ctx, oper.OperationID)
	if err != nil {
		return nil, err
	}

	operInstCtl, err := operCtl.CreateOperationInstance(ctx)
	if err != nil {
		return nil, err
	}

	return []IOperationInstanceCtl{operInstCtl}, nil
}
