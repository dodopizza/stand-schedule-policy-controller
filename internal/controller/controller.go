package controller

import (
	"fmt"
	"time"

	"go.uber.org/zap"

	util "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/utils/clock"

	"github.com/dodopizza/stand-schedule-policy-controller/internal/azure"
	"github.com/dodopizza/stand-schedule-policy-controller/internal/executor"
	"github.com/dodopizza/stand-schedule-policy-controller/internal/kubernetes"
	"github.com/dodopizza/stand-schedule-policy-controller/internal/state"
	apis "github.com/dodopizza/stand-schedule-policy-controller/pkg/apis/standschedules/v1"
	"github.com/dodopizza/stand-schedule-policy-controller/pkg/eventsource"
	"github.com/dodopizza/stand-schedule-policy-controller/pkg/worker"
)

type (
	Controller struct {
		notify   chan error
		logger   *zap.Logger
		clock    clock.WithTicker
		state    *state.State
		kube     kubernetes.Interface
		factory  *kubernetes.FactoryGroup
		lister   *kubernetes.ListerGroup
		events   *eventsource.EventSource[apis.StandSchedulePolicy]
		workers  []*worker.Worker
		executor *executor.Executor
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
		clock:  clock,
		state:  state.New(),
	}
	c.factory = kubernetes.NewFactoryGroup(k, cfg.GetObjectsResyncInterval(), cfg.GetPoliciesResyncInterval())
	c.lister = kubernetes.NewListerGroup(c.factory)
	c.events = eventsource.New[apis.StandSchedulePolicy](
		c.factory.Stands.StandSchedules().V1().StandSchedulePolicies(),
		eventsource.Handlers[apis.StandSchedulePolicy]{
			AddFunc:    c.add,
			UpdateFunc: c.update,
			DeleteFunc: c.delete,
		},
	)
	c.workers = []*worker.Worker{
		worker.New(cfg.GetWorkerConfig("reconciler"), c.logger.Named("reconciler"), c.clock, c.reconcile),
		worker.New(cfg.GetWorkerConfig("executor"), c.logger.Named("executor"), c.clock, c.execute),
	}
	c.executor = executor.New(c.logger, az, c.kube, c.lister)
	return c
}

func (c *Controller) Start(interrupt <-chan struct{}) {
	c.logger.Info("Starting informers")
	c.factory.Start(interrupt)
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
	for _, w := range c.workers {
		w.Start(interrupt)
	}
	c.logger.Info("Started workers")
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

func (c *Controller) enqueueReconcile(key string) {
	c.workers[0].Enqueue(key)
}

func (c *Controller) enqueueExecute(item WorkItem, ts time.Duration) {
	c.workers[1].EnqueueAfter(item, ts)
}
