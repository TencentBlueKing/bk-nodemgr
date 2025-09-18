package scheduler

import (
	"fmt"
	"testing"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
)

// TestScheduler tests the scheduler.
func TestScheduler(t *testing.T) {
	s := NewScheduler(WithLogger(logger.LoggerDefault{}))

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
