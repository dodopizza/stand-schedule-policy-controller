package controller

import (
	"time"

	"github.com/dodopizza/stand-schedule-policy-controller/pkg/worker"
)

type (
	Config struct {
		ObjectsResyncSeconds   int `json:"core_resync_seconds" env:"CONTROLLER_OBJECTS_RESYNC_SECONDS"`
		PoliciesResyncSeconds  int `json:"policies_resync_seconds" env:"CONTROLLER_POLICIES_RESYNC_SECONDS"`
		WorkerQueueThreadiness int `json:"worker_queue_threadiness" env:"CONTROLLER_WORKER_QUEUE_THREADINESS"`
		WorkerQueueRetries     int `json:"worker_queue_retries" env:"CONTROLLER_WORKER_QUEUE_RETRIES"`
	}
)

const (
	_DefaultPoliciesResyncSeconds = 300 // 5 min
	_DefaultObjectsResyncSeconds  = 60  // 1 min
	_MinResyncSeconds             = 10
	_DefaultWorkerQueueRetries    = 5
	_MinWorkerQueueRetries        = 1
)

func (c *Config) GetObjectsResyncInterval() time.Duration {
	return c.getResyncInterval(c.ObjectsResyncSeconds, _MinResyncSeconds, _DefaultObjectsResyncSeconds)
}

func (c *Config) GetPoliciesResyncInterval() time.Duration {
	return c.getResyncInterval(c.PoliciesResyncSeconds, _MinResyncSeconds, _DefaultPoliciesResyncSeconds)
}

func (c *Config) GetThreadiness() int {
	if c.WorkerQueueThreadiness < 1 {
		return 1
	}
	return c.WorkerQueueThreadiness
}

func (c *Config) GetWorkerQueueRetries() int {
	if c.WorkerQueueRetries < _MinWorkerQueueRetries {
		return _DefaultWorkerQueueRetries
	}
	return c.WorkerQueueRetries
}

func (c *Config) GetWorkerConfig(name string) *worker.Config {
	return &worker.Config{
		Name:        name,
		Retries:     c.GetWorkerQueueRetries(),
		Threadiness: c.GetThreadiness(),
	}
}

func (c *Config) getResyncInterval(actual, min, def int) time.Duration {
	if actual < min {
		return time.Duration(def) * time.Second
	}
	return time.Duration(actual) * time.Second
}
