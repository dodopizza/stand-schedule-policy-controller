package controller

import (
	"time"

	"github.com/dodopizza/stand-schedule-policy-controller/pkg/worker"
)

type (
	Config struct {
		ResyncSeconds          int `json:"resync_seconds" env:"CONTROLLER_RESYNC_SECONDS"`
		WorkerQueueThreadiness int `json:"worker_queue_threadiness" env:"CONTROLLER_WORKER_QUEUE_THREADINESS"`
		WorkerQueueRetries     int `json:"worker_queue_retries" env:"CONTROLLER_WORKER_QUEUE_RETRIES"`
	}
)

const (
	_DefaultResyncSeconds      = 300 // 5 min
	_DefaultWorkerQueueRetries = 5
	_DefaultWorkerQueueName    = "stand-schedule-policy-controller"
)

func (c *Config) GetResyncDuration() time.Duration {
	if c.ResyncSeconds < _DefaultResyncSeconds {
		return time.Duration(_DefaultResyncSeconds) * time.Second
	}
	return time.Duration(c.ResyncSeconds) * time.Second
}

func (c *Config) GetThreadiness() int {
	if c.WorkerQueueThreadiness < 1 {
		return 1
	}
	return c.WorkerQueueThreadiness
}

func (c *Config) GetWorkerQueueRetries() int {
	if c.WorkerQueueRetries < 1 {
		return _DefaultWorkerQueueRetries
	}
	return c.WorkerQueueRetries
}

func (c *Config) GetWorkerConfig() *worker.Config {
	return &worker.Config{
		Name:        _DefaultWorkerQueueName,
		Retries:     c.GetWorkerQueueRetries(),
		Threadiness: c.GetThreadiness(),
	}
}
