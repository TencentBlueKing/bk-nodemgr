/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package goasync provides a handler of goasync.
package goasync

import (
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
)

// RunFn defines the function of goasync.
type RunFn = func(nCtx contextx.IContext) error

// Task defines a task of goasync.
type Task struct {
	_     struct{}
	nCtx  contextx.IContext
	runFn RunFn
	name  string
}

// RunOptions defines the options of goasync.
type RunOptions func(task *Task)

// WithTimeout defines the timeout of goasync.
func WithTimeout(timeout time.Duration) RunOptions {
	return func(task *Task) {
		task.nCtx, _ = contextx.WithTimeout(task.nCtx, timeout)
	}
}

// WithName defines the name of goasync.
func WithName(name string) RunOptions {
	return func(task *Task) {
		task.name = name
	}
}
