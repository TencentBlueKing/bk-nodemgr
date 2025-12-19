/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package periodictask implements non business and concurrency-safe periodic tasks.
package periodictask

import (
	"sort"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/pageexecutor"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/trigger"
)

const (
	periodicTaskTimeoutCleanTrigger = 5 * time.Minute
	periodicTaskNameCleanTrigger    = "clean_trigger"

	triggerListMaxPage = 1000
)

// CleanTrigger clean triggers according to the clean policy.
func (pt *PeriodicTask) CleanTrigger(nCtx contextx.IContext) error {
	gp := gopool.NewPool()
	gp.Go(func() error {
		_ = pt.cleanOnceTrigger(nCtx)

		return nil
	})
	gp.Go(func() error {
		_ = pt.cleanOrderedTrigger(nCtx)

		return nil
	})
	gp.Go(func() error {
		_ = pt.cleanPeriodicTrigger(nCtx)

		return nil
	})
	_ = gp.Wait()

	return nil
}

// triggerList defines the trigger list.
// make it sortable and sorted with updated time from old to new.
type triggerList []*trigger.Trigger

// Len returns the length of the trigger list.
func (l triggerList) Len() int {
	return len(l)
}

// Less compares the updated time of the trigger.
func (l triggerList) Less(i, j int) bool {
	return l[i].UpdatedAt.Before(l[j].UpdatedAt)
}

// Swap swaps the elements with indexes i and j.
func (l triggerList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

// operInstList defines the operation instance list.
// make it sortable and sorted with creation time from old to new.
type operInstList []*operation.InstanceBriefData

// Len returns the length of the operation instance list.
func (l operInstList) Len() int {
	return len(l)
}

// Less compares the creation time of the operation instance.
func (l operInstList) Less(i, j int) bool {
	return l[i].Lifecycle.CreatedAt.Before(l[j].Lifecycle.CreatedAt)
}

// Swap swaps the elements with indexes i and j.
func (l operInstList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

// nolint: gocognit
func (pt *PeriodicTask) cleanOnceTrigger(nCtx contextx.IContext) error {
	executor := pageexecutor.NewPageExecutor[*trigger.Trigger](triggerListMaxPage, 1*time.Hour)
	fn := func(nCtx contextx.IContext, p types.Page) ([]*trigger.Trigger, error) {
		triggers, _, err := pt.conf.StgWorkflow.ListTrigger(nCtx, p, trigger.CategoryOnce)
		return triggers, err
	}

	results, err := executor.Execute(nCtx, types.UnlimitedPage(), fn)
	if err != nil {
		return err
	}

	cleanQueue := make(map[string]triggerList)

	deletingTriggerIDs := make([]string, 0)
	for _, trig := range results.Items {
		if trig.Active {
			continue
		}

		meta, ok := trig.Metadata.(*trigger.MetadataOnce)
		if !ok || meta == nil {
			continue
		}

		// if the trigger's last updated time is too long ago, delete it.
		// nolint: mnd
		if meta.CleanPolicy.MaxDays > 0 && int(time.Since(trig.UpdatedAt).Hours()/24) > meta.CleanPolicy.MaxDays {
			deletingTriggerIDs = append(deletingTriggerIDs, trig.TriggerID)

			continue
		}

		// add trigger into the clean queue.
		if _, ok := cleanQueue[meta.CleanPolicy.Namespace]; !ok {
			cleanQueue[meta.CleanPolicy.Namespace] = make(triggerList, 0)
		}
		cleanQueue[meta.CleanPolicy.Namespace] = append(cleanQueue[meta.CleanPolicy.Namespace], trig)
	}

	for _, triggerList := range cleanQueue {
		sort.Sort(sort.Reverse(triggerList))

		for i, trig := range triggerList {
			meta, ok := trig.Metadata.(*trigger.MetadataOnce)
			if !ok || meta == nil {
				continue
			}

			if meta.CleanPolicy.MaxNum > 0 && i >= meta.CleanPolicy.MaxNum {
				deletingTriggerIDs = append(deletingTriggerIDs, trig.TriggerID)
			}
		}
	}

	if len(deletingTriggerIDs) > 0 {
		if err = pt.conf.StgWorkflow.DeleteTriggers(nCtx, deletingTriggerIDs...); err != nil {
			logger.G.Sys().WithErr(err).With("count", len(deletingTriggerIDs)).Warn("failed to delete once triggers")

			return err
		}

		if err = pt.conf.StgWorkflow.DeleteOperationsByTriggerID(nCtx, deletingTriggerIDs...); err != nil {
			logger.G.Sys().WithErr(err).With("count", len(deletingTriggerIDs)).Warn("failed to delete operations of once triggers")

			return err
		}

		if err = pt.conf.StgWorkflow.DeleteOperationInstancesByTriggerID(nCtx, deletingTriggerIDs...); err != nil {
			logger.G.Sys().WithErr(err).With("count", len(deletingTriggerIDs)).Warn("failed to delete operations of once triggers")

			return err
		}

		logger.G.Sys().
			With("trigger-count", len(deletingTriggerIDs)).
			Info("successfully deleted once triggers, operations and operation instances")
	}

	return nil
}

// nolint: gocognit
func (pt *PeriodicTask) cleanOrderedTrigger(nCtx contextx.IContext) error {
	executor := pageexecutor.NewPageExecutor[*trigger.Trigger](triggerListMaxPage, 1*time.Hour)
	fn := func(nCtx contextx.IContext, p types.Page) ([]*trigger.Trigger, error) {
		triggers, _, err := pt.conf.StgWorkflow.ListTrigger(nCtx, p, trigger.CategoryOrdered)
		return triggers, err
	}

	results, err := executor.Execute(nCtx, types.UnlimitedPage(), fn)
	if err != nil {
		return err
	}

	cleanQueue := make(map[string]triggerList)

	deletingTriggerIDs := make([]string, 0)
	for _, trig := range results.Items {
		if trig.Active {
			continue
		}

		meta, ok := trig.Metadata.(*trigger.MetadataOrdered)
		if !ok || meta == nil {
			continue
		}

		// if the trigger's last updated time is too long ago, delete it.
		// nolint: mnd
		if meta.CleanPolicy.MaxDays > 0 && int(time.Since(trig.UpdatedAt).Hours()/24) > meta.CleanPolicy.MaxDays {
			deletingTriggerIDs = append(deletingTriggerIDs, trig.TriggerID)

			continue
		}

		// add trigger into the clean queue.
		if _, ok := cleanQueue[meta.CleanPolicy.Namespace]; !ok {
			cleanQueue[meta.CleanPolicy.Namespace] = make(triggerList, 0)
		}
		cleanQueue[meta.CleanPolicy.Namespace] = append(cleanQueue[meta.CleanPolicy.Namespace], trig)
	}

	for _, triggerList := range cleanQueue {
		sort.Sort(sort.Reverse(triggerList))

		for i, trig := range triggerList {
			meta, ok := trig.Metadata.(*trigger.MetadataOrdered)
			if !ok || meta == nil {
				continue
			}

			if meta.CleanPolicy.MaxNum > 0 && i >= meta.CleanPolicy.MaxNum {
				deletingTriggerIDs = append(deletingTriggerIDs, trig.TriggerID)
			}
		}
	}

	if len(deletingTriggerIDs) > 0 {
		if err = pt.conf.StgWorkflow.DeleteTriggers(nCtx, deletingTriggerIDs...); err != nil {
			logger.G.Sys().WithErr(err).With("trigger-count", len(deletingTriggerIDs)).Warn("failed to delete ordered triggers")

			return err
		}

		if err = pt.conf.StgWorkflow.DeleteOperationsByTriggerID(nCtx, deletingTriggerIDs...); err != nil {
			logger.G.Sys().WithErr(err).With("trigger-count", len(deletingTriggerIDs)).Warn("failed to delete operations of ordered triggers")

			return err
		}

		if err = pt.conf.StgWorkflow.DeleteOperationInstancesByTriggerID(nCtx, deletingTriggerIDs...); err != nil {
			logger.G.Sys().WithErr(err).With("trigger-count", len(deletingTriggerIDs)).Warn("failed to delete oper-instances of ordered triggers")

			return err
		}

		logger.G.Sys().
			With("trigger-count", len(deletingTriggerIDs)).
			Info("successfully deleted ordered triggers, operations and operation instances")
	}

	return nil
}

func (pt *PeriodicTask) cleanPeriodicTrigger(nCtx contextx.IContext) error {
	executor := pageexecutor.NewPageExecutor[*trigger.Trigger](triggerListMaxPage, 1*time.Hour)
	fn := func(nCtx contextx.IContext, p types.Page) ([]*trigger.Trigger, error) {
		triggers, _, err := pt.conf.StgWorkflow.ListTrigger(nCtx, p, trigger.CategoryPeriodic)
		return triggers, err
	}

	results, err := executor.Execute(nCtx, types.UnlimitedPage(), fn)
	if err != nil {
		return err
	}

	deletingOperInstIDs := make([]string, 0)
	for _, trig := range results.Items {
		meta, ok := trig.Metadata.(*trigger.MetadataPeriodic)
		if !ok || meta == nil {
			continue
		}

		operInsts, _, err := pt.conf.StgWorkflow.ListOperInstanceBriefWithoutActionInstByTriggerID(nCtx, types.UnlimitedPage(), trig.TriggerID)
		if err != nil {
			logger.G.Sys().WithErr(err).With("trigger_id", trig.TriggerID).Info("failed to list oper instance")

			continue
		}

		if meta.CleanPolicy.MaxOperInstNum <= 0 || len(operInsts) <= meta.CleanPolicy.MaxOperInstNum {
			continue
		}

		// sorted by updated_at desc.
		sort.Sort(sort.Reverse(operInstList(operInsts)))
		for i := meta.CleanPolicy.MaxOperInstNum; i < len(operInsts); i++ {
			deletingOperInstIDs = append(deletingOperInstIDs, operInsts[i].Metadata.OperationInstanceID)
		}
	}

	if len(deletingOperInstIDs) > 0 {
		if err = pt.conf.StgWorkflow.DeleteOperationInstances(nCtx, deletingOperInstIDs...); err != nil {
			logger.G.Sys().
				WithErr(err).
				With("oper-inst-count", len(deletingOperInstIDs)).
				Info("failed to delete overflow operation instances in periodic trigger")

			return err
		}

		logger.G.Sys().
			With("oper-inst-count", len(deletingOperInstIDs)).
			Info("successfully deleted overflow operation instances in periodic trigger")
	}

	return nil
}
