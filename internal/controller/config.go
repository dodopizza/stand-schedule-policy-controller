package controller

import (
	"time"

	"github.com/dodopizza/stand-schedule-policy-controller/pkg/worker"
)

type (
	Config struct {
		ObjectsResyncSeconds  int `json:"core_resync_seconds" env:"CONTROLLER_OBJECTS_RESYNC_SECONDS"`
		PoliciesResyncSeconds int `json:"policies_resync_seconds" env:"CONTROLLER_POLICIES_RESYNC_SECONDS"`
		ReconcilerThreadiness int `json:"reconciler_threadiness" env:"CONTROLLER_RECONCILER_THREADINESS"`
		ExecutorThreadiness   int `json:"executor_threadiness" env:"CONTROLLER_EXECUTOR_THREADINESS"`
		WorkerQueueRetries    int `json:"worker_queue_retries" env:"CONTROLLER_WORKER_QUEUE_RETRIES"`
	}
)

const (
	_DefaultPoliciesResyncSeconds = 300 // 5 min
	_DefaultObjectsResyncSeconds  = 60  // 1 min
	_MinResyncSeconds             = 10
	_DefaultWorkerQueueRetries    = 5
	_MinWorkerQueueRetries        = 1
	_DefaultReconcilerThreadiness = 1
	_DefaultExecutorThreadiness   = 1
	_MinThreadiness               = 1
)

func (c *Config) GetObjectsResyncInterval() time.Duration {
	return getResyncInterval(c.ObjectsResyncSeconds, _MinResyncSeconds, _DefaultObjectsResyncSeconds)
}

func (c *Config) GetPoliciesResyncInterval() time.Duration {
	return getResyncInterval(c.PoliciesResyncSeconds, _MinResyncSeconds, _DefaultPoliciesResyncSeconds)
}

func (c *Config) GetWorkerQueueRetries() int {
	if c.WorkerQueueRetries < _MinWorkerQueueRetries {
		return _DefaultWorkerQueueRetries
	}
	return c.WorkerQueueRetries
}

func (c *Config) GetReconcilerConfig() *worker.Config {
	return &worker.Config{
		Name:        "reconciler",
		Retries:     c.GetWorkerQueueRetries(),
		Threadiness: getThreadiness(c.ReconcilerThreadiness, _MinThreadiness, _DefaultReconcilerThreadiness),
	}
}

func (c *Config) GetExecutorConfig() *worker.Config {
	return &worker.Config{
		Name:        "executor",
		Retries:     c.GetWorkerQueueRetries(),
		Threadiness: getThreadiness(c.ExecutorThreadiness, _MinThreadiness, _DefaultExecutorThreadiness),
	}
}

func getThreadiness(actual, min, def int) int {
	if actual < min {
		return def
	}
	return actual
}

func getResyncInterval(actual, min, def int) time.Duration {
	if actual < min {
		return time.Duration(def) * time.Second
	}
	return time.Duration(actual) * time.Second
}
