package main

import (
	"flag"
	"fmt"
	"math/rand"
	"time"

	worker_pool "github.com/zhavkk/corp-security/workerpool"
)

func main() {
	workersCnt := flag.Int("workers", 10, "number of workers")
	tasks := flag.Int("tasks", 100, "number of tasks")
	flag.Parse()

	wp := worker_pool.NewWorkerPool(*workersCnt)
	for i := 0; i < *tasks; i++ {
		taskID := i + 1
		wp.Submit(func() {
			err := SimulateWork()
			if err != nil {
				fmt.Printf("task %d: error: %v\n", taskID, err)
			} else {
				fmt.Printf("task %d: completed\n", taskID)
			}
		})
	}
	time.Sleep(50 * time.Millisecond)

	wp.StopWait()

	fmt.Println("All tasks completed.")
}

func SimulateWork() error {
	duration := time.Duration(100+rand.Intn(400)) * time.Millisecond
	time.Sleep(duration)

	if rand.Float64() < 0.2 {
		panic("simulated panic")
	}

	return nil
}
