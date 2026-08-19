/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 节点管理 (BlueKing - Node Management) available.
 * Copyright (C) Tencent. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.

 * We undertake not to change the open source license (MIT license) applicable

 * to the current version of the project delivered to anyone in the future.
 */

package scheduler

import (
	"fmt"
	"testing"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
)

// TestScheduler tests the scheduler.
func TestScheduler(t *testing.T) {
	s := NewScheduler(WithLogger())

	cnt := 0
	s.RegisterTask(NewTask(
		"normal",
		time.Second,
		time.Minute,
		func(ctx contextx.IContext) error {
			t.Logf("cnt: %d\n", cnt)
			cnt++
			return nil
		},
	))
	s.RegisterTask(NewTask(
		"normal",
		"*/5 * * * * *",
		time.Minute,
		func(ctx contextx.IContext) error {
			t.Logf("cnt: %d\n", cnt)
			cnt++
			return nil
		},
	))

	s.RegisterTask(NewTask(
		"timeout",
		time.Second,
		2*time.Second,
		func(ctx contextx.IContext) error {
			time.Sleep(2 * time.Second)
			return nil
		},
	))

	s.RegisterTask(NewTask(
		"error",
		time.Second,
		time.Minute,
		func(ctx contextx.IContext) error {
			return fmt.Errorf("error")
		},
	))

	s.RegisterTask(NewTask(
		"panic",
		time.Second,
		time.Minute,
		func(ctx contextx.IContext) error {
			panic("panic")
		},
	))

	s.Start()

	time.Sleep(time.Second * 10)

	tasks := s.ListTask()
	taskNum := len(tasks)

	s.RemoveTask("panic")

	tasks = s.ListTask()
	if len(tasks) != taskNum-1 {
		t.Errorf("expected task number %d, got %d", taskNum-1, len(tasks))
	}

	s.Terminate()
}
