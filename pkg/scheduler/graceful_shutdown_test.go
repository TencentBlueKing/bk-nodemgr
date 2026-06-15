/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package scheduler

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScheduler_GracefulShutdown_StopsNewScheduling(t *testing.T) {
	t.Parallel()

	s := NewScheduler().(*scheduler)
	var runCount atomic.Int32

	require.NoError(t, s.RegisterTask(NewTask(
		"tick",
		100*time.Millisecond,
		time.Minute,
		func(ctx contextx.IContext) error {
			runCount.Add(1)
			return nil
		},
	)))

	s.Start()
	time.Sleep(250 * time.Millisecond)
	before := runCount.Load()

	require.NoError(t, s.GracefulShutdown(2*time.Second))

	time.Sleep(300 * time.Millisecond)
	assert.Equal(t, before, runCount.Load())
}

func TestScheduler_GracefulShutdown_WaitsForRunningTask(t *testing.T) {
	t.Parallel()

	s := NewScheduler().(*scheduler)
	started := make(chan struct{})
	release := make(chan struct{})
	done := make(chan struct{})

	require.NoError(t, s.RegisterTask(NewTask(
		"slow",
		time.Hour,
		time.Minute,
		func(ctx contextx.IContext) error {
			close(started)
			<-release
			close(done)

			return nil
		},
	)))

	s.Start()
	go s.executeTask(s.tasks["slow"])

	<-started

	errCh := make(chan error, 1)
	go func() {
		errCh <- s.GracefulShutdown(2 * time.Second)
	}()

	select {
	case <-done:
		t.Fatal("task finished before graceful shutdown returned")
	case <-time.After(100 * time.Millisecond):
	}

	close(release)

	select {
	case err := <-errCh:
		require.NoError(t, err)
	case <-time.After(2 * time.Second):
		t.Fatal("graceful shutdown did not return after task finished")
	}

	<-done
}

func TestScheduler_Terminate_DoesNotCancelBackgroundTask(t *testing.T) {
	t.Parallel()

	s := NewScheduler().(*scheduler)
	finished := make(chan struct{})

	require.NoError(t, s.RegisterTask(NewTask(
		"long",
		time.Hour,
		time.Minute,
		func(ctx contextx.IContext) error {
			time.Sleep(200 * time.Millisecond)
			close(finished)

			return nil
		},
	)))

	s.Start()
	go s.executeTask(s.tasks["long"])

	time.Sleep(20 * time.Millisecond)
	s.Terminate()

	select {
	case <-finished:
	case <-time.After(time.Second):
		t.Fatal("terminate should not cancel tasks that use background context")
	}
}

func TestScheduler_GracefulShutdown_FallbackCancelsShuttingDownTask(t *testing.T) {
	t.Parallel()

	s := NewScheduler().(*scheduler)
	started := make(chan struct{})
	cancelled := make(chan struct{}, 1)

	require.NoError(t, s.RegisterTask(NewTask(
		"blocking",
		time.Hour,
		time.Minute,
		func(ctx contextx.IContext) error {
			close(started)
			<-ctx.Done()
			cancelled <- struct{}{}

			return ctx.Err()
		},
	)))

	s.Start()
	s.shuttingDown.Store(true)
	go s.executeTask(s.tasks["blocking"])
	<-started

	s.cancel()

	select {
	case <-cancelled:
	case <-time.After(time.Second):
		t.Fatal("expected shutting-down task to be cancelled during fallback phase")
	}
}

func TestScheduler_GracefulShutdown_ZeroTimeoutDelegatesToTerminate(t *testing.T) {
	t.Parallel()

	s := NewScheduler().(*scheduler)

	require.NoError(t, s.RegisterTask(NewTask(
		"noop",
		time.Hour,
		time.Minute,
		func(ctx contextx.IContext) error {
			return nil
		},
	)))

	s.Start()
	require.NoError(t, s.GracefulShutdown(0))
}
