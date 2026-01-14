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
	"fmt"
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

// lastResortOnceCleanPolicy is the clean policy for the lastResort cleanup of once triggers.
func lastResortOnceCleanPolicy() trigger.MetadataCleanPolicy {
	return trigger.MetadataCleanPolicy{
		MaxDays: 90, // nolint: mnd
	}
}

// nolint: gocognit
func (pt *PeriodicTask) cleanOnceTrigger(nCtx contextx.IContext) error {
	executor := pageexecutor.NewPageExecutor[string](triggerListMaxPage, 1*time.Hour)

	fn := func(nCtx contextx.IContext, p types.Page) ([]string, error) {
		triggers, _, err := pt.conf.StgWorkflow.ListTrigger(nCtx, p, trigger.CategoryOnce)
		if err != nil {
			return nil, fmt.Errorf("failed to list triggers: %w", err)
		}

		needDeletingTriggerIDs := make([]string, 0)
		for _, trig := range triggers {
			if trig.Active {
				continue
			}

			meta, ok := trig.Metadata.(*trigger.MetadataOnce)
			if !ok || meta == nil {
				continue
			}

			cleanPolicy := lastResortOnceCleanPolicy()

			if meta.CleanPolicy.MaxDays > 0 {
				cleanPolicy = meta.CleanPolicy
			}

			// if the trigger's last updated time is too long ago, delete it.
			// nolint: mnd
			if time.Since(trig.UpdatedAt).Hours()/24 > cleanPolicy.MaxDays {
				needDeletingTriggerIDs = append(needDeletingTriggerIDs, trig.TriggerID)
			}
		}

		if len(needDeletingTriggerIDs) > 0 {
			if err = pt.conf.StgWorkflow.DeleteOperationInstancesByTriggerID(nCtx, needDeletingTriggerIDs...); err != nil {
				logger.G.Sys().WithErr(err).With("trigger-count", len(needDeletingTriggerIDs)).
					Warn("failed to delete operation instances of once triggers")

				// Log error but continue processing subsequent pages, return empty list to let PageExecutor continue
				return []string{}, nil
			}

			if err = pt.conf.StgWorkflow.DeleteOperationsByTriggerID(nCtx, needDeletingTriggerIDs...); err != nil {
				logger.G.Sys().WithErr(err).With("trigger-count", len(needDeletingTriggerIDs)).
					Warn("failed to delete operations of once triggers")

				// Log error but continue processing subsequent pages, return empty list to let PageExecutor continue
				return []string{}, nil
			}

			if err = pt.conf.StgWorkflow.DeleteTriggers(nCtx, needDeletingTriggerIDs...); err != nil {
				logger.G.Sys().WithErr(err).With("trigger-count", len(needDeletingTriggerIDs)).
					Warn("failed to delete once triggers")

				// Log error but continue processing subsequent pages, return empty list to let PageExecutor continue
				return []string{}, nil
			}

			logger.G.Sys().
				With("trigger-count", len(needDeletingTriggerIDs)).
				Info("successfully deleted once triggers, operations and operation instances")
		}

		return needDeletingTriggerIDs, nil
	}

	result, err := executor.Execute(nCtx, types.UnlimitedPage(), fn)
	if err != nil {
		return fmt.Errorf("failed to clean once trigger: %w", err)
	}

	logger.G.Sys().With("total-deleted-trigger-count", result.Total).Info("successfully cleaned once triggers")

	return nil
}

// lastResortOrderedCleanPolicy is the clean policy for the lastResort cleanup of ordered triggers.
func lastResortOrderedCleanPolicy() trigger.MetadataCleanPolicy {
	return trigger.MetadataCleanPolicy{
		MaxDays: 90, // nolint: mnd
	}
}

// nolint: gocognit
func (pt *PeriodicTask) cleanOrderedTrigger(nCtx contextx.IContext) error {
	executor := pageexecutor.NewPageExecutor[string](triggerListMaxPage, 1*time.Hour)

	fn := func(nCtx contextx.IContext, p types.Page) ([]string, error) {
		triggers, _, err := pt.conf.StgWorkflow.ListTrigger(nCtx, p, trigger.CategoryOrdered)
		if err != nil {
			return nil, fmt.Errorf("failed to list triggers: %w", err)
		}

		needDeletingTriggerIDs := make([]string, 0)
		for _, trig := range triggers {
			if trig.Active {
				continue
			}

			meta, ok := trig.Metadata.(*trigger.MetadataOrdered)
			if !ok || meta == nil {
				continue
			}

			cleanPolicy := lastResortOrderedCleanPolicy()

			if meta.CleanPolicy.MaxDays > 0 {
				cleanPolicy = meta.CleanPolicy
			}

			// if the trigger's last updated time is too long ago, delete it.
			// nolint: mnd
			if time.Since(trig.UpdatedAt).Hours()/24 > cleanPolicy.MaxDays {
				needDeletingTriggerIDs = append(needDeletingTriggerIDs, trig.TriggerID)
			}
		}

		if len(needDeletingTriggerIDs) > 0 {
			if err = pt.conf.StgWorkflow.DeleteOperationInstancesByTriggerID(nCtx, needDeletingTriggerIDs...); err != nil {
				logger.G.Sys().WithErr(err).With("trigger-count", len(needDeletingTriggerIDs)).
					Warn("failed to delete operation instances of ordered triggers")

				// Log error but continue processing subsequent pages, return empty list to let PageExecutor continue
				return []string{}, nil
			}

			if err = pt.conf.StgWorkflow.DeleteOperationsByTriggerID(nCtx, needDeletingTriggerIDs...); err != nil {
				logger.G.Sys().WithErr(err).With("trigger-count", len(needDeletingTriggerIDs)).
					Warn("failed to delete operations of ordered triggers")

				// Log error but continue processing subsequent pages, return empty list to let PageExecutor continue
				return []string{}, nil
			}

			if err = pt.conf.StgWorkflow.DeleteTriggers(nCtx, needDeletingTriggerIDs...); err != nil {
				logger.G.Sys().WithErr(err).With("trigger-count", len(needDeletingTriggerIDs)).
					Warn("failed to delete ordered triggers")

				// Log error but continue processing subsequent pages, return empty list to let PageExecutor continue
				return []string{}, nil
			}

			logger.G.Sys().
				With("trigger-count", len(needDeletingTriggerIDs)).
				Info("successfully deleted ordered triggers, operations and operation instances")
		}

		return needDeletingTriggerIDs, nil
	}

	result, err := executor.Execute(nCtx, types.UnlimitedPage(), fn)
	if err != nil {
		return fmt.Errorf("failed to clean ordered trigger: %w", err)
	}

	logger.G.Sys().With("total-deleted-trigger-count", result.Total).Info("successfully cleaned ordered triggers")

	return nil
}

// nolint: gocognit
func (pt *PeriodicTask) cleanPeriodicTrigger(nCtx contextx.IContext) error {
	executor := pageexecutor.NewPageExecutor[string](triggerListMaxPage, 1*time.Hour)

	fn := func(nCtx contextx.IContext, p types.Page) ([]string, error) {
		triggers, _, err := pt.conf.StgWorkflow.ListTrigger(nCtx, p, trigger.CategoryPeriodic)
		if err != nil {
			return nil, fmt.Errorf("failed to list triggers: %w", err)
		}

		needDeletingOperInstIDs := make([]string, 0)
		for _, trig := range triggers {
			meta, ok := trig.Metadata.(*trigger.MetadataPeriodic)
			if !ok || meta == nil {
				continue
			}

			operInsts, _, err := pt.conf.StgWorkflow.ListOperInstanceBriefWithoutActionInstByTriggerID(nCtx, types.UnlimitedPage(), trig.TriggerID)
			if err != nil {
				logger.G.Sys().WithErr(err).With("trigger_id", trig.TriggerID).
					Warn("failed to list oper instance")

				continue
			}

			if meta.CleanPolicy.MaxOperInstNum <= 0 || len(operInsts) <= meta.CleanPolicy.MaxOperInstNum {
				continue
			}

			// sorted by updated_at desc.
			sort.Sort(sort.Reverse(operInstList(operInsts)))
			for i := meta.CleanPolicy.MaxOperInstNum; i < len(operInsts); i++ {
				needDeletingOperInstIDs = append(needDeletingOperInstIDs, operInsts[i].Metadata.OperationInstanceID)
			}
		}

		if len(needDeletingOperInstIDs) > 0 {
			if err = pt.conf.StgWorkflow.DeleteOperationInstances(nCtx, needDeletingOperInstIDs...); err != nil {
				logger.G.Sys().
					WithErr(err).
					With("oper-inst-count", len(needDeletingOperInstIDs)).
					Warn("failed to delete overflow operation instances in periodic trigger")

				// Log error but continue processing subsequent pages, return empty list to let PageExecutor continue
				return []string{}, nil
			}

			logger.G.Sys().
				With("oper-inst-count", len(needDeletingOperInstIDs)).
				Info("successfully deleted overflow operation instances in periodic trigger")
		}

		return needDeletingOperInstIDs, nil
	}

	result, err := executor.Execute(nCtx, types.UnlimitedPage(), fn)
	if err != nil {
		return fmt.Errorf("failed to clean periodic trigger: %w", err)
	}

	logger.G.Sys().With("total-deleted-oper-inst-count", result.Total).
		Info("successfully cleaned periodic triggers")

	return nil
}
