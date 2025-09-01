package worker_pool

import (
	"context"
	"log/slog"
	"sync"
)

const (
	MaxWorkers       = 10000
	MinWorkers       = 1
	DefaultQueueSize = 256
)

type WorkerPool struct {
	tasks  chan func()
	mx     sync.Mutex
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
	closed bool
}

func NewWorkerPool(numberOfWorkers int) *WorkerPool {
	if numberOfWorkers > MaxWorkers {
		numberOfWorkers = MaxWorkers
	}
	if numberOfWorkers < MinWorkers {
		numberOfWorkers = MinWorkers
	}
	ctx, cancel := context.WithCancel(context.Background())
	wp := &WorkerPool{
		tasks:  make(chan func(), DefaultQueueSize),
		ctx:    ctx,
		cancel: cancel,
		closed: false,
	}

	for i := 0; i < numberOfWorkers; i++ {
		wp.wg.Add(1)
		go wp.worker()
	}
	return wp
}

func (wp *WorkerPool) worker() {
	defer wp.wg.Done()
	for {
		select {
		case <-wp.ctx.Done():
			return
		case task, ok := <-wp.tasks:
			if !ok {
				return
			}
			if task != nil {
				wp.execute(task)
			}
		}
	}
}

func (wp *WorkerPool) execute(task func()) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("task panicked, recovering.", slog.Any("error", r))
		}
	}()

	task()

}

func (wp *WorkerPool) Submit(task func()) {
	wp.mx.Lock()
	defer wp.mx.Unlock()
	if wp.closed {
		return
	}

	wp.tasks <- task

}

func (wp *WorkerPool) SubmitWait(task func()) {
	wp.mx.Lock()
	if wp.closed {
		wp.mx.Unlock()
		return
	}
	wp.mx.Unlock()

	done := make(chan struct{})
	wrappedTask := func() {
		defer close(done)
		task()
	}

	wp.Submit(wrappedTask)
	<-done
}

func (wp *WorkerPool) Stop() {
	wp.mx.Lock()
	if wp.closed {
		wp.mx.Unlock()
		return
	}
	wp.closed = true
	wp.cancel()
	wp.mx.Unlock()

	go func() {
		for range wp.tasks {
		}
	}()
	wp.wg.Wait()

	close(wp.tasks)

}

func (wp *WorkerPool) StopWait() {
	wp.mx.Lock()
	if wp.closed {
		wp.mx.Unlock()
		return
	}
	wp.closed = true
	wp.mx.Unlock()

	close(wp.tasks)

	wp.wg.Wait()
	wp.cancel()

}
