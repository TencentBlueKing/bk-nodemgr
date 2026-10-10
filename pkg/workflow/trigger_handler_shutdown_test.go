/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package workflow

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	shutdownWaitAssertDelay = 50 * time.Millisecond
	shutdownTimeout         = 10 * time.Millisecond
)

func TestTriggerHandlerGracefulShutdownReturnsWithoutInFlightTrigger(t *testing.T) {
	handler := &triggerHandler{}

	require.NoError(t, handler.GracefulShutdown(time.Second))
}

func TestTriggerHandlerGracefulShutdownWaitsForInFlightTrigger(t *testing.T) {
	handler := &triggerHandler{}
	done := trackTriggerExecutionForTest(handler)

	shutdownErrCh := make(chan error, 1)
	go func() {
		shutdownErrCh <- handler.GracefulShutdown(time.Second)
	}()

	select {
	case err := <-shutdownErrCh:
		require.NoError(t, err)
		t.Fatal("graceful shutdown returned before in-flight trigger finished")
	case <-time.After(shutdownWaitAssertDelay):
	}

	done()

	select {
	case err := <-shutdownErrCh:
		require.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("graceful shutdown did not return after in-flight trigger finished")
	}
}

func TestTriggerHandlerGracefulShutdownTimesOutWaitingForInFlightTrigger(t *testing.T) {
	handler := &triggerHandler{}
	done := trackTriggerExecutionForTest(handler)
	defer done()

	err := handler.GracefulShutdown(shutdownTimeout)
	require.Error(t, err)
	require.True(t, errors.Is(err, context.DeadlineExceeded), "expected deadline exceeded, got %v", err)
}

func TestManagerGracefulShutdownReturnsTriggerShutdownError(t *testing.T) {
	handler := &triggerHandler{}
	done := trackTriggerExecutionForTest(handler)
	defer done()

	mgr := &manager{
		isRunning:               true,
		cancel:                  func() {},
		gracefulShutdownTimeout: shutdownTimeout,
		handlerTail:             make(chan struct{}),
		triggerHandler:          handler,
	}

	err := mgr.GracefulShutdown()
	require.Error(t, err)
	require.ErrorContains(t, err, "trigger handler shutdown")
	require.True(t, errors.Is(err, context.DeadlineExceeded), "expected deadline exceeded, got %v", err)
}

func trackTriggerExecutionForTest(handler *triggerHandler) func() {
	handler.triggerExecutionMutex.Lock()
	defer handler.triggerExecutionMutex.Unlock()

	return handler.trackTriggerExecutionLocked()
}
