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

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/operation"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/operinstdata"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/scheduleworkflow"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/trigger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/identifier"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/locker"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/scheduler"
)

// Config defines the configuration of watcher.
type Config struct {
	Locker locker.MutexFactory
	Logger logger.ILogger

	StgTrigger          trigger.IStorage
	StgOperation        operation.IStorage
	StgOperInst         operinstdata.IStorage
	StgScheduleWorkflow scheduleworkflow.IStorage
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
		scheduler: scheduler.NewScheduler(scheduler.WithLogger(conf.Logger)),
	}
}

// Start starts the watcher manager.
func (pt *PeriodicTask) Start() error {
	pt.conf.Logger.Info("started backend periodic task manager")

	if err := pt.registerTasks(); err != nil {
		pt.conf.Logger.Errorf("register periodic-tasks failed: %v", err)
		return err
	}

	pt.scheduler.Start()

	return nil
}

// Terminate terminates the watcher manager.
func (pt *PeriodicTask) Terminate() {
	pt.conf.Logger.Info("terminating backend periodic task manager")

	pt.scheduler.Terminate()
}

// registerTasks register periodic tasks.
func (pt *PeriodicTask) registerTasks() error {
	periodicTasks := []*scheduler.Task{
		scheduler.NewTask(
			deleteNonLatestScheduleWorkflowOperInstRecordsTaskName,
			deleteNonLatestScheduleWorkflowOperInstRecordsInterval,
			deleteNonLatestScheduleWorkflowOperInstRecordsTimeout,
			pt.DeleteNonLatestWorkflowScheduleOperInstRecords,
		),
	}

	for _, task := range periodicTasks {
		if err := pt.scheduler.RegisterTask(task); err != nil {
			pt.conf.Logger.Errorf("register periodic-task(%s) failed: %v", task.ID, err)
			return err
		}
	}

	return nil
}
