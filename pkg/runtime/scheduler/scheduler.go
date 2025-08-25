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
	"bytes"
	"context"
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/robfig/cron/v3"
)

// CronParser returns a cron parser that supports seconds, minutes, hours, day, month, day of week, and descriptors.
func CronParser() cron.Parser {
	return cron.NewParser(
		cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
	)
}

// NextActiveTime calculates the next active time for a given cron expression starting from a specified time.
func NextActiveTime(cronExpr string, from time.Time) (time.Time, error) {
	parser := CronParser()
	schedule, err := parser.Parse(cronExpr)
	if err != nil {
		return time.Time{}, fmt.Errorf("parser cron-expr(%s), err: %v", cronExpr, err)
	}

	nextTime := schedule.Next(from)
	if nextTime.IsZero() {
		return time.Time{}, fmt.Errorf("no next active time found for cron-expr(%s)", cronExpr)
	}

	return nextTime, nil
}

// Scheduler ...
type Scheduler interface {
	RegisterTask(task *Task) error
	Start()
	Terminate()
	RemoveTask(id string)
	ListTask() map[string]*Task
}

// Task defines a Task that can be scheduled.
// Interval can be a time.Duration or a cron expression string.
type Task struct {
	ID       string
	Interval string
	Timeout  time.Duration
	Fn       func(context.Context) error
}

// NewTask creates a new Task with the given ID, interval, timeout, and function.
// The interval can be a string representing a cron expression or a time.Duration.
func NewTask[T string | time.Duration](
	id string,
	interval T,
	timeout time.Duration,
	fn func(context.Context) error,
) *Task {

	var cronExpr string
	switch t := any(interval).(type) {
	case time.Duration:
		cronExpr = Every + t.String()
	case string:
		cronExpr = t
	}

	return &Task{
		ID:       id,
		Interval: cronExpr,
		Timeout:  timeout,
		Fn:       fn,
	}
}

// scheduler ...
type scheduler struct {
	tasks  map[string]*scheduledTask
	mu     sync.Mutex
	ctx    context.Context
	cancel context.CancelFunc
	cron   *cron.Cron
	logger logger.ILogger
}

// scheduledTask ...
type scheduledTask struct {
	*Task
	entryID      cron.EntryID
	lastExecuted time.Time
}

// OptionFn ...
type OptionFn func(*scheduler)

// WithLogger this func will set the logger of the scheduler.
func WithLogger(logger logger.ILogger) OptionFn {
	return func(s *scheduler) {
		s.logger = logger
	}
}

// NewScheduler ...
func NewScheduler(opts ...OptionFn) Scheduler {
	ctx, cancel := context.WithCancel(context.Background())

	s := &scheduler{
		tasks:  make(map[string]*scheduledTask),
		ctx:    ctx,
		cancel: cancel,
		logger: logger.LoggerDefault{},
	}

	for _, opt := range opts {
		opt(s)
	}

	s.cron = cron.New(
		cron.WithSeconds(),
		// skips the task if it is still running when the next scheduled time arrives.
		cron.WithChain(
			cron.SkipIfStillRunning(LoggerAdapter{s.logger}),
			cron.Recover(LoggerAdapter{s.logger}),
		),
	)

	return s
}

// RegisterTask register a task
func (s *scheduler) RegisterTask(task *Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.tasks[task.ID] != nil {
		return fmt.Errorf("task already exists, task-id(%s)", task.ID)
	}

	s.tasks[task.ID] = &scheduledTask{
		Task: task,
	}

	entryID, err := s.cron.AddFunc(task.Interval, func() {
		s.executeTask(s.tasks[task.ID])
	})
	if err != nil {
		s.logger.Errorf("failed to add cron task, task-id(%s), err: %v", task.ID, err)
		return err
	}

	s.tasks[task.ID].entryID = entryID
	s.logger.Infof("add task into cron list succeed, task-id(%s), entry-id(%v)", task.ID, entryID)

	return nil
}

// Start ...
func (s *scheduler) Start() {
	s.cron.Start()
}

// executeTask ...
func (s *scheduler) executeTask(task *scheduledTask) {
	ctx, cancel := context.WithTimeout(s.ctx, task.Timeout)
	defer cancel()

	defer func() {
		if r := recover(); r != nil {
			stack := debug.Stack()

			// The first line of the stack trace is of the form "goroutine N [status]:",
			// but by the time the panic reaches Do the goroutine may no longer exist,
			// and its status will have changed. Trim out the misleading line.
			if line := bytes.IndexByte(stack[:], '\n'); line >= 0 { //nolint: gocritic
				stack = stack[line+1:]
			}

			s.logger.Errorf("task execution panic, scheduler-task-id(%s), err: %v, stack: \n%s", task.ID, r, stack)
		}

		task.lastExecuted = time.Now()
	}()

	if err := task.Fn(ctx); err != nil {
		s.logger.Errorf("task execution failed, scheduler-task-id(%s), err: %v", task.ID, err)
		return
	}

	s.logger.Debugf("task execution completed, scheduler-task-id(%s)", task.ID)
}

// Terminate ...
func (s *scheduler) Terminate() {
	s.cron.Stop()
	s.cancel()
}

// RemoveTask removes a task by its ID.
func (s *scheduler) RemoveTask(taskID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if task, ok := s.tasks[taskID]; ok {
		s.cron.Remove(task.entryID)
		delete(s.tasks, taskID)
	}
}

// ListTask returns a map of all registered tasks.
func (s *scheduler) ListTask() map[string]*Task {
	s.mu.Lock()
	defer s.mu.Unlock()

	tasks := make(map[string]*Task, len(s.tasks))
	for id, task := range s.tasks {
		tasks[id] = task.Task
	}

	return tasks
}
