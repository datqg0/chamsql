package cronjob

import (
	"context"
	"time"

	"backend/pkgs/logger"
)

// Task represents a scheduled task that runs periodically
type Task interface {
	Name() string
	Execute(ctx context.Context) error
}

// Scheduler manages periodic task execution
type Scheduler struct {
	tasks     []Task
	intervals []time.Duration
	stopCh    chan struct{}
}

// NewScheduler creates a new scheduler
func NewScheduler() *Scheduler {
	return &Scheduler{
		tasks:     make([]Task, 0),
		intervals: make([]time.Duration, 0),
		stopCh:    make(chan struct{}),
	}
}

// Register adds a task to be executed at a given interval
func (s *Scheduler) Register(task Task, interval time.Duration) {
	s.tasks = append(s.tasks, task)
	s.intervals = append(s.intervals, interval)
}

// Start begins executing all registered tasks
func (s *Scheduler) Start(ctx context.Context) {
	for i, task := range s.tasks {
		interval := s.intervals[i]
		go s.runTask(ctx, task, interval)
	}
	logger.Info("Scheduler started with %d tasks", len(s.tasks))
}

// Stop stops the scheduler gracefully
func (s *Scheduler) Stop() {
	close(s.stopCh)
	logger.Info("Scheduler stopped")
}

// runTask executes a task repeatedly at the given interval
func (s *Scheduler) runTask(ctx context.Context, task Task, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	logger.Info("Cronjob started: %s (interval: %v)", task.Name(), interval)

	for {
		select {
		case <-ctx.Done():
			logger.Info("Cronjob stopped (context cancelled): %s", task.Name())
			return
		case <-s.stopCh:
			logger.Info("Cronjob stopped (scheduler shutdown): %s", task.Name())
			return
		case <-ticker.C:
			if err := task.Execute(ctx); err != nil {
				logger.Error("Cronjob %s failed: %v", task.Name(), err)
			}
		}
	}
}
