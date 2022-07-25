package controller

import (
	"fmt"

	"go.uber.org/zap"

	util "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/utils/clock"

	"github.com/dodopizza/stand-schedule-policy-controller/internal/azure"
	"github.com/dodopizza/stand-schedule-policy-controller/internal/kubernetes"
	apis "github.com/dodopizza/stand-schedule-policy-controller/pkg/apis/standschedules/v1"
	"github.com/dodopizza/stand-schedule-policy-controller/pkg/eventsource"
	"github.com/dodopizza/stand-schedule-policy-controller/pkg/worker"
)

type (
	Controller struct {
		notify     chan error
		logger     *zap.Logger
		kube       kubernetes.Interface
		azure      azure.Interface
		clock      clock.WithTicker
		state      *State
		reconciler *worker.Worker
		executor   *worker.Worker
		factory    *FactoryGroup
		lister     *ListerGroup
		events     *eventsource.EventSource[apis.StandSchedulePolicy]
	}
)

func NewController(
	cfg *Config,
	l *zap.Logger,
	clock clock.WithTicker,
	k kubernetes.Interface,
	az azure.Interface,
) *Controller {
	c := &Controller{
		notify: make(chan error, 1),
		logger: l.Named("controller"),
		kube:   k,
		azure:  az,
		clock:  clock,
		state:  NewControllerState(),
	}
	c.factory = NewFactoryGroup(k, cfg)
	c.lister = NewListerGroup(c.factory)
	c.events = eventsource.New[apis.StandSchedulePolicy](
		c.factory.stands.StandSchedules().V1().StandSchedulePolicies(),
		eventsource.Handlers[apis.StandSchedulePolicy]{
			AddFunc:    c.add,
			UpdateFunc: c.update,
			DeleteFunc: c.delete,
		},
	)
	c.reconciler = worker.New(cfg.GetWorkerConfig(), c.logger.Named("reconciler"), c.clock, c.reconcile)
	c.executor = worker.New(cfg.GetWorkerConfig(), c.logger.Named("executor"), c.clock, c.execute)
	return c
}

func (c *Controller) Start(interrupt <-chan struct{}) {
	c.logger.Info("Starting informers")
	c.factory.core.Start(interrupt)
	c.factory.stands.Start(interrupt)
	c.logger.Info("Started informers")

	c.logger.Info("Syncing caches")
	for t, ok := range c.factory.WaitForCacheSync(interrupt) {
		if !ok {
			c.handleCachesDesyncFor(t.Name())
			return
		}
	}
	c.logger.Info("Synced caches")

	c.logger.Info("Starting workers")
	c.reconciler.Start(interrupt)
	c.executor.Start(interrupt)
	c.logger.Info("Started workers")
}

func (c *Controller) Shutdown() error {
	c.reconciler.Shutdown()
	c.executor.Shutdown()
	return nil
}

func (c *Controller) Notify() <-chan error {
	return c.notify
}

func (c *Controller) handleCachesDesyncFor(name string) {
	err := fmt.Errorf("failed to sync informer caches for: %s", name)
	c.logger.Error("Failed to sync informer caches for type", zap.Error(err))

	// invoke high level error handlers
	util.HandleError(err)

	// critical error, notify about it
	c.notify <- err
	close(c.notify)
}
