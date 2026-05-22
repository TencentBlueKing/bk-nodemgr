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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/pageexecutor"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/locker"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/scheduler"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/common"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/trigger"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
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
)

func newTriggerHandler(mgr *manager, globalLocker locker.MutexFactory) (*triggerHandler, error) {
	trigHandler := &triggerHandler{
		mgr:          mgr,
		globalLocker: globalLocker,

		onceTriggers:     newCachedTriggers(),
		orderedTriggers:  newCachedTriggers(),
		periodicTriggers: newCachedTriggers(),

		tracerProvider:            mgr.traceSvc.TracerProvider(),
		triggerExecutionSemaphore: make(chan struct{}, triggerExecutionConcurrency),
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

	triggerExecutionSemaphore chan struct{}
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

	onceTriggersSyncAndCheckIntervalDefault         = 1 * time.Second
	orderedTriggersSyncAndCheckIntervalDefault      = 1 * time.Second
	periodicTriggersSyncAndCheckIntervalDefault     = 1 * time.Second
	checkAccumulateOperationInstanceIntervalDefault = 1 * time.Minute

	// listAliveTriggerBatchSizeDefault is used to avoid overloading the database from a single large query.
	listAliveTriggerBatchSizeDefault = 500

	taskIDSyncAndCheckOnceTrigger          = "sync_and_check_once_trigger"
	taskIDSyncAndCheckOrderedTrigger       = "sync_and_check_ordered_trigger"
	taskIDSyncAndCheckPeriodicTrigger      = "sync_and_check_periodic_trigger"
	taskIDCheckAccumulateOperationInstance = "check_accumulate_operation_instance"

	checkAccumulateOperationInstanceUpdateConcurrency = 100
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
		scheduler.NewTask(
			taskIDCheckAccumulateOperationInstance,
			checkAccumulateOperationInstanceIntervalDefault,
			defaultTimeout,
			func(nCtx contextx.IContext) error {
				if err := handler.checkAccumulateOperationInstance(nCtx); err != nil {
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
	list, err := handler.listActiveTriggers(nCtx, trigger.CategoryOnce)
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
	list, err := handler.listActiveTriggers(nCtx, trigger.CategoryOrdered)
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
	list, err := handler.listActiveTriggers(nCtx, trigger.CategoryPeriodic)
	if err != nil {
		// set cached triggers to empty cause the cache is no longer valid.
		handler.periodicTriggers.set([]*trigger.Trigger{})

		return err
	}

	logger.G.Sys().With("count", len(list)).Debug("synced periodic triggers")

	handler.periodicTriggers.set(list)

	return nil
}

// listActiveTriggers lists active triggers by category in pages.
func (handler *triggerHandler) listActiveTriggers(nCtx contextx.IContext, category trigger.Category) ([]*trigger.Trigger, error) {
	executor := pageexecutor.NewPageExecutor[*trigger.Trigger](listAliveTriggerBatchSizeDefault, defaultTimeout)
	fn := func(nCtx contextx.IContext, p types.Page) ([]*trigger.Trigger, error) {
		return handler.mgr.stgTrigger.ListActiveTrigger(nCtx, p, category)
	}

	result, err := executor.Execute(nCtx, types.UnlimitedPage(), fn)
	if err != nil {
		return nil, fmt.Errorf("failed to list active triggers, category(%s): %w", category, err)
	}

	return result.Items, nil
}

// executeTriggerList executes the trigger list.
func (handler *triggerHandler) executeTriggerList(nCtx contextx.IContext, list []*trigger.Trigger) error {
	logger.G.Sys().With("count", len(list)).Debug("check trigger list")

	for idx := range list {
		trig := list[idx]
		fn := func(nCtx contextx.IContext) error {
			handler.triggerExecutionSemaphore <- struct{}{}
			defer func() {
				<-handler.triggerExecutionSemaphore
			}()
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

// nolint: nonamedreturns
func (handler *triggerHandler) doTrigger(nCtx contextx.IContext, trigCtl ITriggerCtl) (err error) {
	traceCtx, span := handler.tracerProvider.Tracer(scopeNameTrigger).Start(nCtx,
		fmt.Sprintf("%s %s", spanNamePrefixTrigger, trigCtl.GetTriggerCategory()),
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(
			attribute.String(attributeKeyTriggerID, trigCtl.GetTriggerID()),
			attribute.String(attributeKeyTriggerCategory, string(trigCtl.GetTriggerCategory())),
		))
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
			span.RecordError(err)
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

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

const (
	// instantiateOperationTimeoutRatio reserves one quarter of defaultTimeout for operation instantiation queries.
	instantiateOperationTimeoutRatio = 4
	// instantiateOperationListCostLimit is the maximum expected cost of one ListNeedInstantiateOperation page.
	instantiateOperationListCostLimit = 1 * time.Second
	// launchOperationInstanceTimeoutRatio reserves one quarter of defaultTimeout for operation instance launches.
	launchOperationInstanceTimeoutRatio = 4
	// launchOperationInstanceCostLimit is the maximum expected cost of one LaunchOperationInstance DB operation.
	launchOperationInstanceCostLimit = 1 * time.Second

	// triggerExecutionConcurrency limits the number of triggers running doTrigger concurrently.
	triggerExecutionConcurrency = 20

	// instantiateOperationConcurrency limits concurrent MongoDB writes when creating operation instances.
	instantiateOperationConcurrency = 20
)

// onceTriggerBatchSize returns the maximum operations instantiated and launched per once trigger cycle.
func onceTriggerBatchSize() int {
	return instantiateOperationBatchSize()
}

// orderedTriggerBatchSize returns the maximum operations instantiated and launched per ordered trigger cycle.
func orderedTriggerBatchSize() int {
	return instantiateOperationBatchSize()
}

func instantiateOperationBatchSize() int {
	queryBudget := defaultTimeout / instantiateOperationTimeoutRatio
	pageCount := int(queryBudget / instantiateOperationListCostLimit)

	return pageCount * instantiateOperationConcurrency
}

// launchOperationInstanceConcurrency returns the concurrent launches needed to finish one once-trigger batch within the launch budget.
func launchOperationInstanceConcurrency() int {
	launchBudget := defaultTimeout / launchOperationInstanceTimeoutRatio
	operationCountPerWorker := max(1, int(launchBudget/launchOperationInstanceCostLimit))

	return (instantiateOperationBatchSize() + operationCountPerWorker - 1) / operationCountPerWorker
}

// Instantiate operation list queries are expected to finish within one quarter of defaultTimeout.
// Each paged ListNeedInstantiateOperation call should finish within instantiateOperationListCostLimit.

// instantiateOperation instantiates operations for the trigger.
func (handler *triggerHandler) instantiateOperation(nCtx contextx.IContext, trigCtl ITriggerCtl, page types.Page) error {
	executor := pageexecutor.NewPageExecutor[IOperationCtl](instantiateOperationConcurrency, defaultTimeout)
	fn := func(nCtx contextx.IContext, p types.Page) ([]IOperationCtl, error) {
		operList, err := trigCtl.ListNeedInstantiateOperation(nCtx, p)
		if err != nil {
			return nil, fmt.Errorf("failed to list need instantiate operation: %w", err)
		}

		return operList, nil
	}

	pageResult, err := executor.Execute(nCtx, page, fn)
	if err != nil {
		return fmt.Errorf("failed to list need instantiate operation: %w", err)
	}

	logger.G.Sys().With("trigger-id", trigCtl.GetTriggerID(), "operation-count", len(pageResult.Items)).Debug("instantiate operation")

	gp := gopool.NewPool()
	gp.SetLimit(instantiateOperationConcurrency)
	for _, operCtl := range pageResult.Items {
		ctl := operCtl
		gp.Go(func() error {
			operInst, err := ctl.CreateOperationInstance(nCtx)
			if err != nil {
				logger.G.Sys().
					WithErr(err).
					With("trigger-id", trigCtl.GetTriggerID(),
						"operation-id", ctl.GetOperationID()).
					Error("failed to create operation instance")

				return err
			}

			logger.G.Sys().
				With("trigger-id", trigCtl.GetTriggerID(),
					"operation-id", ctl.GetOperationID(),
					"oper-inst-id", operInst.GetOperationInstanceID()).
				Debug("created operation instance")

			return nil
		})
	}

	if err := gp.Wait(); err != nil {
		return err
	}

	logger.G.Sys().With("trigger-id", trigCtl.GetTriggerID()).Debug("instantiate operation done")

	return nil
}

func (handler *triggerHandler) doOnceTrigger(nCtx contextx.IContext, trigCtl ITriggerCtl) ([]IOperationInstanceCtl, error) {
	if err := handler.instantiateOperation(nCtx, trigCtl, types.Page{Limit: onceTriggerBatchSize()}); err != nil {
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
		operation.StateLaunched, operation.StateRunning)
	if err != nil {
		return nil, err
	}

	// not idle concurrent num.
	idleNum := min(orderedTriggerBatchSize(), metadata.MaxConcurrencyNum-int(workingCount))
	instanceList := make([]IOperationInstanceCtl, 0)
	if idleNum > 0 {
		if err := handler.instantiateOperation(nCtx, trigCtl, types.Page{Limit: idleNum}); err != nil {
			logger.G.Sys().
				WithErr(err).
				With("trigger-id", trigCtl.GetTriggerID()).
				Warn("failed to init ordered empty operation")
		}

		instanceList, err = trigCtl.ListOperationInstances(nCtx, types.Page{Limit: idleNum}, operation.StateInit)
		if err != nil {
			return nil, err
		}
	}

	logger.G.Sys().With("trigger-id", trigCtl.GetTriggerID(),
		"working-count", workingCount,
		"idle-num", idleNum,
		"instantiated-count", len(instanceList)).
		Debug("ordered trigger processed")

	return instanceList, nil
}

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
		return make([]IOperationInstanceCtl, 0), nil
	}

	workingCount, err := handler.mgr.stgOperationInstance.CountOperationInstanceByState(nCtx, trigCtl.GetTriggerID(),
		operation.StateLaunched, operation.StateRunning)
	if err != nil {
		return nil, err
	}

	if metadata.AllowedConcurrency || workingCount == 0 {
		logger.G.Sys().With("trigger-id", trigCtl.GetTriggerID()).
			Debug("periodic trigger next activation time reached, no working instance, proceed to create operation instance")

		operList, count, err := handler.mgr.stgOperation.ListOperation(nCtx, types.UnlimitedPage(), &types.OperationCondition{
			ExactInclude: &types.OperationExactFields{
				TriggerID: []string{trigCtl.GetTriggerID()},
			},
		})
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

		if _, err = operCtl.CreateOperationInstance(nCtx); err != nil {
			return nil, err
		}

		logger.G.Sys().With("trigger-id", trigCtl.GetTriggerID(),
			"oper-id", operCtl.GetOperationID()).
			Debug("created periodic operation instance")
	}

	return trigCtl.ListOperationInstances(nCtx, types.UnlimitedPage(), operation.StateInit)
}

func (handler *triggerHandler) launchOperationInstance(nCtx contextx.IContext, trigCtl ITriggerCtl, instanceCtls []IOperationInstanceCtl) error {
	gp := gopool.NewPool()
	gp.SetLimit(launchOperationInstanceConcurrency())
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

func (handler *triggerHandler) checkAccumulateOperationInstance(nCtx contextx.IContext) error {
	condition := types.OperInstDataCondition{
		ExactInclude: &types.OperInstDataExactFields{
			State: []operation.State{operation.StateLaunched, operation.StateRunning},
		},
		LifeCycleStartedAtTimeRange: &types.TimeRange{
			// If it starts before the start time of the previous life cycle, you need to determine whether it is abnormal data.
			EndTime: time.Now().Add(-checkAccumulateOperationInstanceIntervalDefault),
		},
	}

	maxPageSize := 5000

	queryExecutor := pageexecutor.NewPageExecutor[*operation.InstanceBriefData](maxPageSize, time.Minute)
	queryFn := func(nCtx contextx.IContext, p types.Page) ([]*operation.InstanceBriefData, error) {
		operationInstanceBriefData, _, err := handler.mgr.stgOperationInstance.ListOperationInstanceBriefDataWithoutActionInst(nCtx, p, &condition)
		if err != nil {
			return nil, err
		}

		return operationInstanceBriefData, err
	}

	pageResult, err := queryExecutor.Execute(nCtx, types.UnlimitedPage(), queryFn)
	if err != nil {
		return fmt.Errorf("failed to list operation instance, err: %w", err)
	}

	checkPoint := time.Now()
	needEndWithTimeout := make([]*operation.InstanceBriefData, 0)
	for _, item := range pageResult.Items {
		// theoretical end time = started time + timeout
		theoreticalEndAt := item.Lifecycle.StartedAt.Add(item.Metadata.Timeout)

		// This operation instance has crossed the theoretical endpoint and needs to be marked for a timeout
		if checkPoint.After(theoreticalEndAt) {
			item.Lifecycle.End(operation.StateTimeout)
			needEndWithTimeout = append(needEndWithTimeout, item)
		}
	}

	logger.G.Sys().Ctx(nCtx).With(
		"candidate-count", len(pageResult.Items),
		"timeout-count", len(needEndWithTimeout),
		"check-point", checkPoint,
	).Info("checked accumulated operation instances")

	gp := gopool.NewPool()
	gp.SetLimit(checkAccumulateOperationInstanceUpdateConcurrency)
	for idx := range needEndWithTimeout {
		item := needEndWithTimeout[idx]
		gp.Go(func() error {
			err := handler.mgr.stgOperationInstance.UpdateOperationInstanceLifecycle(nCtx, item.Metadata.OperationInstanceID, item.Lifecycle)
			if err != nil {
				logger.G.Sys().Ctx(nCtx).WithErr(err).With(
					"oper-inst-id", item.Metadata.OperationInstanceID,
					"operation-id", item.Metadata.OperationID,
					"trigger-id", item.Metadata.TriggerID,
					"started-at", item.Lifecycle.StartedAt,
				).Error("failed to update the needed timeout operation instance's lifecycle")

				return fmt.Errorf("failed to update the needed timeout operation instance's lifecycle, err: %w", err)
			}

			return nil
		})
	}

	if err = gp.Wait(); err != nil {
		return err
	}

	logger.G.Sys().Ctx(nCtx).With(
		"timeout-count", len(needEndWithTimeout),
		"check-point", checkPoint,
	).Info("updated timeout operation instance lifecycle")

	return nil
}
