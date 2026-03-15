package worker

import (
	"context"
	"log"
	"time"
)

type Job struct {
	Name     string
	Interval time.Duration
	Run      func(context.Context) error
}

type Scheduler struct {
	jobs []Job
}

func NewScheduler(jobs ...Job) *Scheduler {
	return &Scheduler{jobs: jobs}
}

func (s *Scheduler) Start(ctx context.Context) {
	for _, job := range s.jobs {
		go runLoop(ctx, job)
	}
}

func runLoop(ctx context.Context, job Job) {
	timer := time.NewTimer(0)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			if err := job.Run(ctx); err != nil {
				log.Printf("worker %s failed: %v", job.Name, err)
			}
			timer.Reset(job.Interval)
		}
	}
}
