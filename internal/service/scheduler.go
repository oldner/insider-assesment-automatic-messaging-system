package service

import (
	"insider-assessment/internal/config"
	"log/slog"
	"sync"
	"time"
)

// Scheduler handles the background ticker.
type Scheduler struct {
	Sender  *WorkerService
	Config  *config.Config
	ticker  *time.Ticker
	quit    chan struct{}
	running bool
	mu      sync.Mutex
}

func NewScheduler(sender *WorkerService, cfg *config.Config) *Scheduler {
	return &Scheduler{
		Sender: sender,
		Config: cfg,
		quit:   make(chan struct{}),
	}
}

// Start initiates the ticker if it's not already running.
func (s *Scheduler) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		slog.Warn("scheduler is already running.")
		return
	}

	slog.Info("starting scheduler", "interval", s.Config.WorkerInterval)

	s.ticker = time.NewTicker(s.Config.WorkerInterval)
	s.running = true
	s.quit = make(chan struct{})

	// run on start
	go s.Sender.ProcessMessages()

	go func() {
		for {
			select {
			case <-s.ticker.C:
				s.Sender.ProcessMessages()
			case <-s.quit:
				s.ticker.Stop()
				slog.Info("Scheduler stopped.")
				return
			}
		}
	}()
}

// Stop stops the ticker.
func (s *Scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		slog.Warn("Scheduler is not running.")
		return
	}

	slog.Info("Stopping Scheduler...")
	close(s.quit)
	s.running = false
}
