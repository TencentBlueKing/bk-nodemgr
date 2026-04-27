/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package periodictask implements non business periodic tasks.
package periodictask

import (
	"sync"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/globalsettings"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/identifier"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/locker"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/scheduler"
)

// IPeriodicTask defines the interface for periodic task manager.
type IPeriodicTask interface {
	Start(ctx contextx.IContext) error
	Terminate()
}

// Config defines the configuration of watcher.
type Config struct {
	Locker locker.MutexFactory

	StgWorkflow workflow.IStorage
}

// PeriodicTask defines a watcher manager.
type PeriodicTask struct {
	id        string
	mu        sync.Mutex
	conf      Config
	scheduler scheduler.Scheduler
}

// NewPeriodicTask creates a new watcher manager.
func NewPeriodicTask(conf Config) *PeriodicTask {
	return &PeriodicTask{
		id:        identifier.GenServiceID(),
		mu:        sync.Mutex{},
		conf:      conf,
		scheduler: scheduler.NewScheduler(),
	}
}

// Start starts the watcher manager.
func (pt *PeriodicTask) Start(nCtx contextx.IContext) error {
	logger.G.Sys().Info("started backend periodic task manager")

	if err := pt.registerTasks(nCtx); err != nil {
		return err
	}

	pt.scheduler.Start()

	return nil
}

// Terminate terminates the watcher manager.
func (pt *PeriodicTask) Terminate() {
	logger.G.Sys().Info("terminating backend periodic task manager")

	pt.scheduler.Terminate()
}

// registerTasks register periodic tasks.
func (pt *PeriodicTask) registerTasks(ctx contextx.IContext) error {
	periodicTasks := []*scheduler.Task{
		scheduler.NewTask(
			periodicTaskNameCleanTrigger,
			globalsettings.Get(ctx, globalsettings.CleanTriggerIntervalSecond, scheduler.Every1m),
			periodicTaskTimeoutCleanTrigger,
			pt.CleanTrigger,
		),
	}

	for _, task := range periodicTasks {
		if err := pt.scheduler.RegisterTask(task); err != nil {
			logger.G.Sys().WithErr(err).With("task-id", task.ID).Error("failed to register periodic task")

			return err
		}
	}

	return nil
}
