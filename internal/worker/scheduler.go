package worker

import (
	"context"
	"log/slog"
	"time"
)

type Job struct {
	Name     string
	Interval time.Duration
	Run      func(context.Context) error
}

type Scheduler struct {
	logger *slog.Logger
	jobs   []Job
}

func NewScheduler(logger *slog.Logger, jobs ...Job) *Scheduler {
	if logger == nil {
		logger = slog.Default()
	}
	return &Scheduler{
		logger: logger,
		jobs:   jobs,
	}
}

func (s *Scheduler) Start(ctx context.Context) {
	s.logger.Info("starting background workers", "jobs", len(s.jobs))
	for _, job := range s.jobs {
		go runLoop(ctx, job, s.logger.With("job", job.Name))
	}
}

func runLoop(ctx context.Context, job Job, logger *slog.Logger) {
	timer := time.NewTimer(0)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			logger.Debug("worker stopped")
			return
		case <-timer.C:
			startedAt := time.Now()
			logger.Debug("worker started", "interval_seconds", int(job.Interval.Seconds()))
			if err := job.Run(ctx); err != nil {
				logger.Error("worker failed", "duration_ms", time.Since(startedAt).Milliseconds(), "err", err)
			} else {
				logger.Debug("worker completed", "duration_ms", time.Since(startedAt).Milliseconds())
			}
			timer.Reset(job.Interval)
		}
	}
}
