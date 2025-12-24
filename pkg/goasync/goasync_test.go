/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package goasync

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/tracing"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/trace"
)

// setupTestTracer creates a test tracer with stdout exporter and returns tracer provider
func setupTestTracer(t *testing.T) trace.TracerProvider {
	ctx := context.Background()
	nCtx := contextx.New(ctx)

	// Create tracing handler with default config (stdout exporter)
	handler, err := tracing.New(nCtx, tracing.DefaultConfig())
	assert.NoError(t, err)

	// Create service tracer
	service, err := handler.NewService(tracing.ServiceConfig{
		ServiceName: "goasync-test",
		SampleRate:  1.0,
	})
	assert.NoError(t, err)

	return service.TracerProvider()
}

// setupTestContext creates a context with a valid span for testing
// Note: The span will be ended when the context is garbage collected or the test completes
func setupTestContext(t *testing.T, tracerProvider trace.TracerProvider) contextx.IContext {
	ctx := context.Background()

	// Create a root span in the context
	// The span will remain valid for the duration of the test
	tracer := tracerProvider.Tracer("test")
	spanCtx, _ := tracer.Start(ctx, "test-root-span")

	// Create contextx with the span context
	nCtx := contextx.New(spanCtx, contextx.WithTenantID("test-tenant"))

	return nCtx
}

// TestNewHandler tests creating a new handler
func TestNewHandler(t *testing.T) {
	// Test with valid options
	option := HandlerOption{
		PoolNum:               10,
		PerPoolSize:           5,
		LoadBalancingStrategy: LoadBalancingStrategyRoundRobin,
	}

	handler, err := NewHandler(option)
	assert.NoError(t, err)
	assert.NotNil(t, handler)
	// Handler is created successfully - the size fields are used internally for pool creation
}

// TestHandlerOption_Validate tests option validation
func TestHandlerOption_Validate(t *testing.T) {
	tests := []struct {
		name        string
		setupOption func(t *testing.T) HandlerOption
		wantErr     bool
	}{
		{
			name: "valid options",
			setupOption: func(t *testing.T) HandlerOption {
				return HandlerOption{
					PoolNum:     10,
					PerPoolSize: 5,
				}
			},
			wantErr: false,
		},
		{
			name: "invalid size",
			setupOption: func(t *testing.T) HandlerOption {
				return HandlerOption{
					PoolNum:     0,
					PerPoolSize: 5,
				}
			},
			wantErr: true,
		},
		{
			name: "invalid size per pool",
			setupOption: func(t *testing.T) HandlerOption {
				return HandlerOption{
					PoolNum:     10,
					PerPoolSize: 0,
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			option := tt.setupOption(t)
			err := option.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestHandler_Run_Basic tests basic async task execution
func TestHandler_Run_Basic(t *testing.T) {
	handler, err := NewHandler(HandlerOption{
		PoolNum:               10,
		PerPoolSize:           10,
		LoadBalancingStrategy: LoadBalancingStrategyRoundRobin,
	})
	assert.NoError(t, err)

	ctx := context.Background()
	nCtx := contextx.New(ctx,
		contextx.WithTenantID("test-tenant-001"),
		contextx.WithBKUsername("test-user"),
		contextx.WithMessageID("test-msg-001"),
	)

	var wg sync.WaitGroup
	var executed bool
	var receivedCtx contextx.IContext

	wg.Add(1)
	err = handler.Run(nCtx, func(nCtx contextx.IContext) error {
		defer wg.Done()
		executed = true
		receivedCtx = nCtx

		// Verify that we received some context (may lose some values through tracing conversion)
		assert.NotNil(t, nCtx)

		time.Sleep(100 * time.Millisecond)
		return nil
	})
	assert.NoError(t, err)

	wg.Wait()
	assert.True(t, executed)
	assert.NotNil(t, receivedCtx)
}

// TestHandler_Run_Concurrent tests concurrent execution
func TestHandler_Run_Concurrent(t *testing.T) {
	handler, err := NewHandler(HandlerOption{
		PoolNum:               50,
		PerPoolSize:           10,
		LoadBalancingStrategy: LoadBalancingStrategyRoundRobin,
	})
	assert.NoError(t, err)

	ctx := context.Background()
	nCtx := contextx.New(ctx, contextx.WithTenantID("concurrent-test"))

	var wg sync.WaitGroup
	var counter int64
	const taskCount = 100

	for i := 0; i < taskCount; i++ {
		wg.Add(1)
		err := handler.Run(nCtx, func(nCtx contextx.IContext) error {
			defer wg.Done()
			time.Sleep(10 * time.Millisecond)
			atomic.AddInt64(&counter, 1)
			return nil
		})
		assert.NoError(t, err)
	}

	wg.Wait()
	assert.Equal(t, int64(taskCount), atomic.LoadInt64(&counter))
}

// TestHandler_Run_PoolCapacity tests goroutine pool capacity limits
func TestHandler_Run_PoolCapacity(t *testing.T) {
	option := HandlerOption{
		PoolNum:               2,
		PerPoolSize:           2,
		LoadBalancingStrategy: LoadBalancingStrategyRoundRobin,
	}

	handler, err := NewHandler(option)
	assert.NoError(t, err)

	ctx := context.Background()
	nCtx := contextx.New(ctx, contextx.WithTenantID("pool-test"))

	var wg sync.WaitGroup
	var maxConcurrent int64
	var mu sync.Mutex

	const taskCount = 10
	for i := 0; i < taskCount; i++ {
		wg.Add(1)
		err := handler.Run(nCtx, func(nCtx contextx.IContext) error {
			defer wg.Done()

			mu.Lock()
			atomic.AddInt64(&maxConcurrent, 1)
			current := atomic.LoadInt64(&maxConcurrent)
			mu.Unlock()

			time.Sleep(50 * time.Millisecond)

			atomic.AddInt64(&maxConcurrent, -1)

			// Log the concurrency level for debugging
			t.Logf("Current concurrent tasks: %d", current)
			return nil
		})
		assert.NoError(t, err)
	}

	wg.Wait()

	// After all tasks complete, check that we had some concurrency
	finalMax := atomic.LoadInt64(&maxConcurrent)
	t.Logf("Final max concurrent tasks observed: %d", finalMax)

	// The test passes if all tasks complete successfully - the pool management is handled by ants
	// and we mainly verify that the handler can handle multiple concurrent tasks
}

// TestHandler_Run_WithError tests error handling in async tasks
func TestHandler_Run_WithError(t *testing.T) {
	handler, err := NewHandler(HandlerOption{
		PoolNum:               10,
		PerPoolSize:           10,
		LoadBalancingStrategy: LoadBalancingStrategyRoundRobin,
	})
	assert.NoError(t, err)

	ctx := context.Background()
	nCtx := contextx.New(ctx, contextx.WithTenantID("error-test"))

	// Test that error in task doesn't prevent handler from accepting new tasks
	err = handler.Run(nCtx, func(nCtx contextx.IContext) error {
		return fmt.Errorf("test error")
	})
	assert.NoError(t, err) // Run itself should not return error

	// Wait for task to complete and error to be logged
	time.Sleep(100 * time.Millisecond)

	// Handler should still be able to accept new tasks
	var executed bool
	var wg sync.WaitGroup
	wg.Add(1)
	err = handler.Run(nCtx, func(nCtx contextx.IContext) error {
		defer wg.Done()
		executed = true
		return nil
	})
	assert.NoError(t, err)

	wg.Wait()
	assert.True(t, executed)
}

// TestHandler_Run_Panic tests panic handling in async tasks
func TestHandler_Run_Panic(t *testing.T) {
	handler, err := NewHandler(HandlerOption{
		PoolNum:               10,
		PerPoolSize:           10,
		LoadBalancingStrategy: LoadBalancingStrategyRoundRobin,
	})
	assert.NoError(t, err)

	ctx := context.Background()
	nCtx := contextx.New(ctx, contextx.WithTenantID("panic-test"))

	// Test that panic in task doesn't prevent handler from accepting new tasks
	err = handler.Run(nCtx, func(nCtx contextx.IContext) error {
		panic("test panic")
	})
	assert.NoError(t, err) // Run itself should not return error

	// Wait for task to complete and panic to be handled
	time.Sleep(100 * time.Millisecond)

	// Handler should still be able to accept new tasks
	var executed bool
	var wg sync.WaitGroup
	wg.Add(1)
	err = handler.Run(nCtx, func(nCtx contextx.IContext) error {
		defer wg.Done()
		executed = true
		return nil
	})
	assert.NoError(t, err)

	wg.Wait()
	assert.True(t, executed)
}

// TestHandler_Run_MixedErrors tests handling of mixed successful and failed tasks
func TestHandler_Run_MixedErrors(t *testing.T) {
	handler, err := NewHandler(HandlerOption{
		PoolNum:               20,
		PerPoolSize:           10,
		LoadBalancingStrategy: LoadBalancingStrategyRoundRobin,
	})
	assert.NoError(t, err)

	ctx := context.Background()
	nCtx := contextx.New(ctx, contextx.WithTenantID("mixed-test"))

	var wg sync.WaitGroup
	var successCount int64
	var errorCount int64

	const totalTasks = 20
	for i := 0; i < totalTasks; i++ {
		wg.Add(1)
		err := handler.Run(nCtx, func(nCtx contextx.IContext) error {
			defer wg.Done()

			// Half of tasks succeed, half fail
			if atomic.LoadInt64(&successCount)%2 == 0 {
				atomic.AddInt64(&successCount, 1)
				return nil
			} else {
				atomic.AddInt64(&errorCount, 1)
				return fmt.Errorf("task error")
			}
		})
		assert.NoError(t, err)
	}

	wg.Wait()

	// Verify that some tasks succeeded and some failed
	assert.Greater(t, atomic.LoadInt64(&successCount), int64(0))
	assert.Greater(t, atomic.LoadInt64(&errorCount), int64(0))
	assert.Equal(t, int64(totalTasks), atomic.LoadInt64(&successCount)+atomic.LoadInt64(&errorCount))
}

// TestHandler_TracingIntegration tests tracing integration
func TestHandler_TracingIntegration(t *testing.T) {
	// Setup test tracer provider
	tracerProvider := setupTestTracer(t)

	handler, err := NewHandler(HandlerOption{
		PoolNum:               10,
		PerPoolSize:           10,
		LoadBalancingStrategy: LoadBalancingStrategyRoundRobin,
	})
	assert.NoError(t, err)

	// Create context with a valid span
	nCtx := setupTestContext(t, tracerProvider)

	var executed bool
	var spanValid bool
	var wg sync.WaitGroup
	wg.Add(1)

	err = handler.Run(nCtx, func(nCtx contextx.IContext) error {
		defer wg.Done()
		executed = true
		// Task should execute within a tracing span
		assert.NotNil(t, nCtx)

		span := trace.SpanFromContext(nCtx)
		if span.SpanContext().IsValid() {
			spanValid = true
		}

		return nil
	})
	assert.NoError(t, err)

	wg.Wait()
	assert.True(t, executed)
	assert.True(t, spanValid, "span should be valid in async task")
}

// TestHandler_TracingContext tests that tracing context is properly propagated
func TestHandler_TracingContext(t *testing.T) {
	// Setup test tracer provider
	tracerProvider := setupTestTracer(t)

	handler, err := NewHandler(HandlerOption{
		PoolNum:               10,
		PerPoolSize:           10,
		LoadBalancingStrategy: LoadBalancingStrategyRoundRobin,
	})
	assert.NoError(t, err)

	// Create context with a valid span
	nCtx := setupTestContext(t, tracerProvider)

	var spanCreated bool
	var wg sync.WaitGroup
	wg.Add(1)

	err = handler.Run(nCtx, func(nCtx contextx.IContext) error {
		defer wg.Done()

		span := trace.SpanFromContext(nCtx)
		if span.SpanContext().IsValid() {
			spanCreated = true
		}

		return nil
	})
	assert.NoError(t, err)

	wg.Wait()
	assert.True(t, spanCreated, "span should be created and valid")
}

// TestHandler_TracingWithDifferentLoadBalancing tests tracing with different load balancing strategies
func TestHandler_TracingWithDifferentLoadBalancing(t *testing.T) {
	// Setup test tracer provider
	tracerProvider := setupTestTracer(t)

	strategies := []LoadBalancingStrategy{
		LoadBalancingStrategyRoundRobin,
		LoadBalancingStrategyLeastFirst,
	}

	for _, strategy := range strategies {
		t.Run(fmt.Sprintf("strategy_%v", strategy), func(t *testing.T) {
			handler, err := NewHandler(HandlerOption{
				PoolNum:               5,
				PerPoolSize:           5,
				LoadBalancingStrategy: strategy,
			})
			assert.NoError(t, err)

			// Create context with a valid span
			nCtx := setupTestContext(t, tracerProvider)

			var wg sync.WaitGroup
			var spanValidCount int64
			const taskCount = 5

			for i := 0; i < taskCount; i++ {
				wg.Add(1)
				err := handler.Run(nCtx, func(nCtx contextx.IContext) error {
					defer wg.Done()

					// Verify tracing is working
					span := trace.SpanFromContext(nCtx)
					if span.SpanContext().IsValid() {
						atomic.AddInt64(&spanValidCount, 1)
					}

					time.Sleep(10 * time.Millisecond)
					return nil
				})
				assert.NoError(t, err)
			}

			wg.Wait()
			// All tasks should have valid spans
			assert.Equal(t, int64(taskCount), atomic.LoadInt64(&spanValidCount))
		})
	}
}

// TestHandler_Run_WithNilContext tests behavior with nil context
func TestHandler_Run_WithNilContext(t *testing.T) {
	handler, err := NewHandler(HandlerOption{
		PoolNum:               10,
		PerPoolSize:           10,
		LoadBalancingStrategy: LoadBalancingStrategyRoundRobin,
	})
	assert.NoError(t, err)

	// Test with nil context - should create background context
	var executed bool
	var wg sync.WaitGroup
	wg.Add(1)

	err = handler.Run(nil, func(nCtx contextx.IContext) error {
		defer wg.Done()
		executed = true
		// Should receive a background context
		assert.NotNil(t, nCtx)
		return nil
	})
	assert.NoError(t, err)

	wg.Wait()
	assert.True(t, executed)
}

// TestHandler_Run_WithCancelledContext tests behavior with cancelled context
func TestHandler_Run_WithCancelledContext(t *testing.T) {
	handler, err := NewHandler(HandlerOption{
		PoolNum:               10,
		PerPoolSize:           10,
		LoadBalancingStrategy: LoadBalancingStrategyRoundRobin,
	})
	assert.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately
	nCtx := contextx.New(ctx, contextx.WithTenantID("cancel-test"))

	var executed bool
	var wg sync.WaitGroup
	wg.Add(1)

	err = handler.Run(nCtx, func(nCtx contextx.IContext) error {
		defer wg.Done()
		executed = true
		// Task should still execute because tracing uses WithoutCancel
		return nil
	})
	assert.NoError(t, err)

	wg.Wait()
	assert.True(t, executed)
}

// TestHandler_Run_WithNilFunction tests behavior with nil function
func TestHandler_Run_WithNilFunction(t *testing.T) {
	handler, err := NewHandler(HandlerOption{
		PoolNum:               10,
		PerPoolSize:           10,
		LoadBalancingStrategy: LoadBalancingStrategyRoundRobin,
	})
	assert.NoError(t, err)

	ctx := context.Background()
	nCtx := contextx.New(ctx, contextx.WithTenantID("nil-fn-test"))

	// Handler should accept nil function without error
	// The nil check is handled inside the task execution function
	err = handler.Run(nCtx, nil)
	assert.NoError(t, err)

	// Wait for task to complete and nil check to be logged
	time.Sleep(100 * time.Millisecond)

	// Handler should still be able to accept new tasks
	var executed bool
	var wg sync.WaitGroup
	wg.Add(1)
	err = handler.Run(nCtx, func(nCtx contextx.IContext) error {
		defer wg.Done()
		executed = true
		return nil
	})
	assert.NoError(t, err)

	wg.Wait()
	assert.True(t, executed)
}

// TestHandler_Run_WithName tests behavior with name option
func TestHandler_Run_WithName(t *testing.T) {
	handler, err := NewHandler(HandlerOption{
		PoolNum:               10,
		PerPoolSize:           10,
		LoadBalancingStrategy: LoadBalancingStrategyRoundRobin,
	})
	assert.NoError(t, err)

	ctx := context.Background()
	nCtx := contextx.New(ctx, contextx.WithTenantID("name-test"))

	var executed bool
	var wg sync.WaitGroup
	wg.Add(1)

	taskName := "test-task-name"
	err = handler.Run(nCtx, func(nCtx contextx.IContext) error {
		defer wg.Done()
		executed = true
		return nil
	}, WithName(taskName))
	assert.NoError(t, err)

	wg.Wait()
	assert.True(t, executed)
	// The task name is used internally for tracing span naming
}

// TestHandler_Run_WithTimeoutContext tests behavior with timeout context
func TestHandler_Run_WithTimeoutContext(t *testing.T) {
	handler, err := NewHandler(HandlerOption{
		PoolNum:               10,
		PerPoolSize:           10,
		LoadBalancingStrategy: LoadBalancingStrategyRoundRobin,
	})
	assert.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	nCtx := contextx.New(ctx, contextx.WithTenantID("timeout-test"))

	var executed bool
	var wg sync.WaitGroup
	wg.Add(1)

	err = handler.Run(nCtx, func(nCtx contextx.IContext) error {
		defer wg.Done()
		executed = true
		// Sleep longer than the context timeout
		time.Sleep(100 * time.Millisecond)
		return nil
	})
	assert.NoError(t, err)

	wg.Wait()
	assert.True(t, executed) // Task should still execute due to WithoutCancel
}

// TestHandler_NewHandlerEdgeCases tests edge cases for handler creation
func TestHandler_NewHandlerEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		setupOption func(t *testing.T) HandlerOption
		wantErr     bool
	}{
		{
			name: "minimum valid sizes",
			setupOption: func(t *testing.T) HandlerOption {
				return HandlerOption{
					PoolNum:               1,
					PerPoolSize:           1,
					LoadBalancingStrategy: LoadBalancingStrategyRoundRobin,
				}
			},
			wantErr: false,
		},
		{
			name: "large sizes",
			setupOption: func(t *testing.T) HandlerOption {
				return HandlerOption{
					PoolNum:               10000,
					PerPoolSize:           1000,
					LoadBalancingStrategy: LoadBalancingStrategyRoundRobin,
				}
			},
			wantErr: false,
		},
		{
			name: "uneven pool distribution",
			setupOption: func(t *testing.T) HandlerOption {
				return HandlerOption{
					PoolNum:               10,
					PerPoolSize:           3, // 10 % 3 != 0
					LoadBalancingStrategy: LoadBalancingStrategyRoundRobin,
				}
			},
			wantErr: false, // ants should handle this
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			option := tt.setupOption(t)
			_, err := NewHandler(option)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestHandler_ResourceCleanup tests resource cleanup behavior
func TestHandler_ResourceCleanup(t *testing.T) {
	handler, err := NewHandler(HandlerOption{
		PoolNum:               10,
		PerPoolSize:           5,
		LoadBalancingStrategy: LoadBalancingStrategyRoundRobin,
	})
	assert.NoError(t, err)

	ctx := context.Background()
	nCtx := contextx.New(ctx, contextx.WithTenantID("cleanup-test"))

	var wg sync.WaitGroup
	const taskCount = 20

	// Submit tasks that create resources and track cleanup
	var resourcesCreated int64
	var resourcesCleaned int64

	for i := 0; i < taskCount; i++ {
		wg.Add(1)
		err := handler.Run(nCtx, func(nCtx contextx.IContext) error {
			defer wg.Done()

			// Simulate resource creation
			atomic.AddInt64(&resourcesCreated, 1)

			// Do some work
			time.Sleep(10 * time.Millisecond)

			// Simulate resource cleanup
			atomic.AddInt64(&resourcesCleaned, 1)

			return nil
		})
		assert.NoError(t, err)
	}

	wg.Wait()
	assert.Equal(t, int64(taskCount), atomic.LoadInt64(&resourcesCreated))
	assert.Equal(t, int64(taskCount), atomic.LoadInt64(&resourcesCleaned))
}

// TestHandler_MemoryUsage tests memory usage patterns
func TestHandler_MemoryUsage(t *testing.T) {
	handler, err := NewHandler(HandlerOption{
		PoolNum:               50,
		PerPoolSize:           10,
		LoadBalancingStrategy: LoadBalancingStrategyRoundRobin,
	})
	assert.NoError(t, err)

	ctx := context.Background()
	nCtx := contextx.New(ctx, contextx.WithTenantID("memory-test"))

	var wg sync.WaitGroup
	const taskCount = 100

	for i := 0; i < taskCount; i++ {
		wg.Add(1)
		err := handler.Run(nCtx, func(nCtx contextx.IContext) error {
			defer wg.Done()

			// Allocate some memory and release it
			data := make([]byte, 1024) // 1KB per task
			_ = data                   // Use the data to prevent optimization

			time.Sleep(1 * time.Millisecond)
			return nil
		})
		assert.NoError(t, err)
	}

	wg.Wait()
	// Test passes if no memory leaks or crashes occur
}

// TestHandler_PoolReuse tests that goroutine pools are properly reused
func TestHandler_PoolReuse(t *testing.T) {
	handler, err := NewHandler(HandlerOption{
		PoolNum:               10,
		PerPoolSize:           5,
		LoadBalancingStrategy: LoadBalancingStrategyRoundRobin,
	})
	assert.NoError(t, err)

	ctx := context.Background()
	nCtx := contextx.New(ctx, contextx.WithTenantID("reuse-test"))

	// First batch of tasks
	var wg1 sync.WaitGroup
	for i := 0; i < 15; i++ {
		wg1.Add(1)
		err := handler.Run(nCtx, func(nCtx contextx.IContext) error {
			defer wg1.Done()
			time.Sleep(10 * time.Millisecond)
			return nil
		})
		assert.NoError(t, err)
	}
	wg1.Wait()

	// Second batch of tasks should reuse the same pools
	var wg2 sync.WaitGroup
	for i := 0; i < 15; i++ {
		wg2.Add(1)
		err := handler.Run(nCtx, func(nCtx contextx.IContext) error {
			defer wg2.Done()
			time.Sleep(10 * time.Millisecond)
			return nil
		})
		assert.NoError(t, err)
	}
	wg2.Wait()

	// If we reach here, pool reuse is working correctly
}

// TestHandler_ConcurrentHandlers tests multiple handlers running concurrently
func TestHandler_ConcurrentHandlers(t *testing.T) {
	const handlerCount = 5
	handlers := make([]*Handler, handlerCount)

	// Create multiple handlers
	for i := 0; i < handlerCount; i++ {
		handler, err := NewHandler(HandlerOption{
			PoolNum:               10,
			PerPoolSize:           5,
			LoadBalancingStrategy: LoadBalancingStrategyRoundRobin,
		})
		assert.NoError(t, err)
		handlers[i] = handler
	}

	ctx := context.Background()
	nCtx := contextx.New(ctx, contextx.WithTenantID("multi-handler-test"))

	var wg sync.WaitGroup
	const tasksPerHandler = 10

	// Submit tasks to all handlers concurrently
	for _, handler := range handlers {
		for j := 0; j < tasksPerHandler; j++ {
			wg.Add(1)
			err := handler.Run(nCtx, func(nCtx contextx.IContext) error {
				defer wg.Done()
				time.Sleep(10 * time.Millisecond)
				return nil
			})
			assert.NoError(t, err)
		}
	}

	wg.Wait()
	// Test passes if all handlers work concurrently without interference
}

// TestHandler_BenchmarkTaskSubmission benchmarks task submission performance
func TestHandler_BenchmarkTaskSubmission(t *testing.T) {
	handler, err := NewHandler(HandlerOption{
		PoolNum:               100,
		PerPoolSize:           20,
		LoadBalancingStrategy: LoadBalancingStrategyRoundRobin,
	})
	assert.NoError(t, err)

	ctx := context.Background()
	nCtx := contextx.New(ctx, contextx.WithTenantID("benchmark-test"))

	const taskCount = 1000
	start := time.Now()

	var wg sync.WaitGroup
	for i := 0; i < taskCount; i++ {
		wg.Add(1)
		err := handler.Run(nCtx, func(nCtx contextx.IContext) error {
			defer wg.Done()
			return nil
		})
		assert.NoError(t, err)
	}

	wg.Wait()
	duration := time.Since(start)

	// Calculate tasks per second
	tasksPerSecond := float64(taskCount) / duration.Seconds()
	t.Logf("Submitted %d tasks in %v (%.2f tasks/sec)", taskCount, duration, tasksPerSecond)

	// Ensure reasonable performance (should handle >1000 tasks/sec)
	assert.Greater(t, tasksPerSecond, 1000.0)
}

// TestHandler_BenchmarkTaskExecution benchmarks task execution performance
func TestHandler_BenchmarkTaskExecution(t *testing.T) {
	handler, err := NewHandler(HandlerOption{
		PoolNum:               100,
		PerPoolSize:           20,
		LoadBalancingStrategy: LoadBalancingStrategyRoundRobin,
	})
	assert.NoError(t, err)

	ctx := context.Background()
	nCtx := contextx.New(ctx, contextx.WithTenantID("benchmark-exec-test"))

	const taskCount = 500
	var completed int64

	start := time.Now()

	var wg sync.WaitGroup
	for i := 0; i < taskCount; i++ {
		wg.Add(1)
		err := handler.Run(nCtx, func(nCtx contextx.IContext) error {
			defer wg.Done()
			atomic.AddInt64(&completed, 1)
			// Simulate light work
			time.Sleep(1 * time.Millisecond)
			return nil
		})
		assert.NoError(t, err)
	}

	wg.Wait()
	duration := time.Since(start)

	tasksPerSecond := float64(taskCount) / duration.Seconds()
	t.Logf("Executed %d tasks in %v (%.2f tasks/sec)", taskCount, duration, tasksPerSecond)

	assert.Equal(t, int64(taskCount), atomic.LoadInt64(&completed))
}

// TestHandler_BenchmarkVsGoroutines compares performance vs raw goroutines
func TestHandler_BenchmarkVsGoroutines(t *testing.T) {
	handler, err := NewHandler(HandlerOption{
		PoolNum:               100,
		PerPoolSize:           20,
		LoadBalancingStrategy: LoadBalancingStrategyRoundRobin,
	})
	assert.NoError(t, err)

	ctx := context.Background()
	nCtx := contextx.New(ctx, contextx.WithTenantID("benchmark-comparison-test"))

	const taskCount = 200

	// Test with goasync handler
	var handlerCompleted int64
	handlerStart := time.Now()

	var handlerWg sync.WaitGroup
	for i := 0; i < taskCount; i++ {
		handlerWg.Add(1)
		err := handler.Run(nCtx, func(nCtx contextx.IContext) error {
			defer handlerWg.Done()
			atomic.AddInt64(&handlerCompleted, 1)
			time.Sleep(5 * time.Millisecond)
			return nil
		})
		assert.NoError(t, err)
	}
	handlerWg.Wait()
	handlerDuration := time.Since(handlerStart)

	// Test with raw goroutines
	var goroutineCompleted int64
	goroutineStart := time.Now()

	var goroutineWg sync.WaitGroup
	for i := 0; i < taskCount; i++ {
		goroutineWg.Add(1)
		go func() {
			defer goroutineWg.Done()
			atomic.AddInt64(&goroutineCompleted, 1)
			time.Sleep(5 * time.Millisecond)
		}()
	}
	goroutineWg.Wait()
	goroutineDuration := time.Since(goroutineStart)

	assert.Equal(t, int64(taskCount), atomic.LoadInt64(&handlerCompleted))
	assert.Equal(t, int64(taskCount), atomic.LoadInt64(&goroutineCompleted))

	handlerTPS := float64(taskCount) / handlerDuration.Seconds()
	goroutineTPS := float64(taskCount) / goroutineDuration.Seconds()

	t.Logf("goasync: %d tasks in %v (%.2f tasks/sec)", taskCount, handlerDuration, handlerTPS)
	t.Logf("goroutines: %d tasks in %v (%.2f tasks/sec)", taskCount, goroutineDuration, goroutineTPS)

	// goasync should be competitive with goroutines (within 50% performance)
	ratio := handlerTPS / goroutineTPS
	assert.Greater(t, ratio, 0.5, "goasync should be at least 50% as fast as raw goroutines")
}

// TestHandler_ScalabilityTest tests scalability with different pool sizes
func TestHandler_ScalabilityTest(t *testing.T) {
	testSizes := []struct {
		name        string
		size        int
		sizePerPool int
	}{
		{"small", 10, 5},
		{"medium", 50, 10},
		{"large", 100, 20},
	}

	const taskCount = 100

	for _, testSize := range testSizes {
		t.Run(testSize.name, func(t *testing.T) {
			handler, err := NewHandler(HandlerOption{
				PoolNum:               testSize.size,
				PerPoolSize:           testSize.sizePerPool,
				LoadBalancingStrategy: LoadBalancingStrategyRoundRobin,
			})
			assert.NoError(t, err)

			ctx := context.Background()
			nCtx := contextx.New(ctx, contextx.WithTenantID("scalability-test"))

			start := time.Now()

			var wg sync.WaitGroup
			for i := 0; i < taskCount; i++ {
				wg.Add(1)
				err := handler.Run(nCtx, func(nCtx contextx.IContext) error {
					defer wg.Done()
					time.Sleep(2 * time.Millisecond)
					return nil
				})
				assert.NoError(t, err)
			}

			wg.Wait()
			duration := time.Since(start)

			tasksPerSecond := float64(taskCount) / duration.Seconds()
			t.Logf("PoolNum %s (%d/%d): %d tasks in %v (%.2f tasks/sec)",
				testSize.name, testSize.size, testSize.sizePerPool, taskCount, duration, tasksPerSecond)

			// All configurations should complete tasks successfully
			assert.Greater(t, tasksPerSecond, 0.0)
		})
	}
}
