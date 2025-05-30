package scheduler

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
)

// TestScheduler tests the scheduler.
func TestScheduler(t *testing.T) {
	s := NewScheduler(WithLogger(logger.LoggerDefault{}))

	cnt := 0
	s.RegisterTask(&Task{
		ID:       "normal",
		Interval: time.Second,
		Timeout:  time.Minute,
		Fn: func(ctx context.Context) error {
			t.Logf("cnt: %d\n", cnt)

			cnt++

			return nil
		},
	})

	s.RegisterTask(&Task{
		ID:       "timeout",
		Interval: time.Second,
		Timeout:  2 * time.Second,
		Fn: func(ctx context.Context) error {
			time.Sleep(2 * time.Second)

			return nil
		},
	})

	s.RegisterTask(&Task{
		ID:       "error",
		Interval: time.Second,
		Timeout:  time.Minute,
		Fn: func(ctx context.Context) error {
			return fmt.Errorf("error")
		},
	})

	s.RegisterTask(&Task{
		ID:       "panic",
		Interval: time.Second,
		Timeout:  time.Minute,
		Fn: func(ctx context.Context) error {
			panic("panic")
		},
	})

	s.Start()

	time.Sleep(time.Second * 10)

	s.Terminate()
}
