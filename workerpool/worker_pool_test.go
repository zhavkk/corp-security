package worker_pool

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestWorkerPoolBasic(t *testing.T) {
	wp := NewWorkerPool(2)
	defer wp.Stop()

	var counter int64
	wp.Submit(func() {
		atomic.AddInt64(&counter, 1)
	})

	time.Sleep(100 * time.Millisecond)
	if atomic.LoadInt64(&counter) != 1 {
		t.Errorf("Expected 1, got %d", counter)
	}
}

func TestStopVsStopWait(t *testing.T) {
	t.Run("Stop ignores queued tasks", func(t *testing.T) {
		var completed int64
		wp := NewWorkerPool(1)

		for i := 0; i < 10; i++ {
			wp.Submit(func() {
				time.Sleep(50 * time.Millisecond)
				atomic.AddInt64(&completed, 1)
			})
		}

		time.Sleep(25 * time.Millisecond)
		wp.Stop()

		if count := atomic.LoadInt64(&completed); count > 2 {
			t.Errorf("Stop() completed too many tasks: %d", count)
		}
	})

	t.Run("StopWait processes all queued tasks", func(t *testing.T) {
		var completed int64
		wp := NewWorkerPool(1)

		for i := 0; i < 5; i++ {
			wp.Submit(func() {
				time.Sleep(10 * time.Millisecond)
				atomic.AddInt64(&completed, 1)
			})
		}

		wp.StopWait()

		if count := atomic.LoadInt64(&completed); count != 5 {
			t.Errorf("StopWait() should complete all tasks, got %d/5", count)
		}
	})
}

func TestSubmitWait(t *testing.T) {
	wp := NewWorkerPool(2)
	defer wp.Stop()

	var completed bool
	wp.SubmitWait(func() {
		time.Sleep(50 * time.Millisecond)
		completed = true
	})

	if !completed {
		t.Error("SubmitWait should wait for task completion")
	}
}

func TestWorkerPoolClosedAfterStop(t *testing.T) {
	wp := NewWorkerPool(2)
	wp.Stop()

	wp.Submit(func() {})
}

func TestStopWaitProcessesAllTasks(t *testing.T) {
	var completed int64
	wp := NewWorkerPool(2)

	for i := 0; i < 10; i++ {
		wp.Submit(func() {
			time.Sleep(10 * time.Millisecond)
			atomic.AddInt64(&completed, 1)
		})
	}

	wp.StopWait()

	if atomic.LoadInt64(&completed) != 10 {
		t.Errorf("Expected all tasks to be completed, got %d/10", completed)
	}
}

func TestWorkerPoolHandlesPanic(t *testing.T) {
	wp := NewWorkerPool(2)
	defer wp.Stop()

	var completed int64
	wp.Submit(func() {
		panic("test panic")
	})
	wp.Submit(func() {
		atomic.AddInt64(&completed, 1)
	})

	time.Sleep(50 * time.Millisecond)
	if atomic.LoadInt64(&completed) != 1 {
		t.Errorf("Expected 1 task to complete after panic, got %d", completed)
	}
}

func TestSubmitWaitMultipleTasks(t *testing.T) {
	wp := NewWorkerPool(2)
	defer wp.Stop()

	var completed int64
	for i := 0; i < 5; i++ {
		wp.SubmitWait(func() {
			time.Sleep(10 * time.Millisecond)
			atomic.AddInt64(&completed, 1)
		})
	}

	if atomic.LoadInt64(&completed) != 5 {
		t.Errorf("Expected 5 tasks to complete, got %d", completed)
	}
}

func TestWorkerPoolQueueOverflow(t *testing.T) {
	wp := NewWorkerPool(2)
	defer wp.Stop()

	for i := 0; i < DefaultQueueSize+10; i++ {
		wp.Submit(func() {
			time.Sleep(10 * time.Millisecond)
		})
	}

	time.Sleep(500 * time.Millisecond)
}

func TestWorkerPoolWorkersTerminate(t *testing.T) {
	wp := NewWorkerPool(5)
	wp.Stop()

	done := make(chan struct{})
	go func() {
		wp.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Error("Expected all workers to terminate after Stop, but they did not")
	}
}
