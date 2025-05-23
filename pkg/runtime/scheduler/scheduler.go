/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package scheduler ...
package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

// Scheduler ...
type Scheduler interface {
	RegisterTask(task *Task)
	Start()
	Terminate()
}

// Task ...
type Task struct {
	ID       string
	Interval time.Duration
	Timeout  time.Duration
	Fn       func(context.Context) error
}

// scheduler ...
type scheduler struct {
	tasks    map[string]*scheduledTask
	mu       sync.Mutex
	ctx      context.Context
	cancel   context.CancelFunc
	group    singleflight.Group
	interval time.Duration
	logger   Logger
}

// scheduledTask ...
type scheduledTask struct {
	*Task
	lastExecuted time.Time
}

// OptionFn ...
type OptionFn func(*scheduler)

// WithLogger this func will set the logger of the scheduler.
func WithLogger(logger Logger) OptionFn {
	return func(s *scheduler) {
		s.logger = logger
	}
}

// WithInterval this func will set the interval of the scheduler.
func WithInterval(interval time.Duration) OptionFn {
	return func(s *scheduler) {
		s.interval = interval
	}
}

// NewScheduler ...
func NewScheduler(opts ...OptionFn) Scheduler {
	ctx, cancel := context.WithCancel(context.Background())

	s := &scheduler{
		tasks:    make(map[string]*scheduledTask),
		ctx:      ctx,
		cancel:   cancel,
		logger:   &defaultLogger{},
		interval: time.Second,
		group:    singleflight.Group{},
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// RegisterTask register a task
func (s *scheduler) RegisterTask(task *Task) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.tasks[task.ID] = &scheduledTask{
		Task: task,
	}
}

// Start ...
func (s *scheduler) Start() {
	go func() {
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		for {
			select {
			case <-s.ctx.Done():
				return
			case <-ticker.C:
				s.runTasks()
			}
		}
	}()
}

// runTasks ...
func (s *scheduler) runTasks() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for _, task := range s.tasks {
		if now.Sub(task.lastExecuted) >= task.Interval {
			go s.executeTask(task)
		}
	}
}

// executeTask ...
func (s *scheduler) executeTask(task *scheduledTask) {
	ctx, cancel := context.WithTimeout(s.ctx, task.Timeout)
	defer cancel()

	// use singleflight to prevent repeated calls.
	// nolint: dogsled,nonamedreturns
	_, _, _ = s.group.Do(task.ID, func() (result interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				result = false
				err = fmt.Errorf("task execution panic, scheduler-task-id(%s), err: %v", task.ID, r)
			}

			task.lastExecuted = time.Now()
		}()

		err = task.Fn(ctx)
		if err != nil {
			s.logger.Errorf("task execution failed, scheduler-task-id(%s), err: %v", task.ID, err)

			return false, err
		}

		s.logger.Debugf("task execution completed, scheduler-task-id(%s)", task.ID)

		return true, nil
	})
}

// Terminate ...
func (s *scheduler) Terminate() {
	s.cancel()
}
