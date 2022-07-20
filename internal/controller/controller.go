package controller

import (
	"fmt"

	"go.uber.org/zap"

	util "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/utils/clock"

	"github.com/dodopizza/stand-schedule-policy-controller/internal/azure"
	"github.com/dodopizza/stand-schedule-policy-controller/internal/kubernetes"
	apis "github.com/dodopizza/stand-schedule-policy-controller/pkg/apis/standschedules/v1"
	"github.com/dodopizza/stand-schedule-policy-controller/pkg/worker"
)

type (
	Controller struct {
		notify   chan error
		logger   *zap.Logger
		kube     kubernetes.Interface
		azure    azure.Interface
		worker   *worker.Worker
		factory  *FactoryGroup
		informer *Informer
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
	}
	c.worker = worker.New(cfg.GetWorkerConfig(), c.logger, clock, c.reconcile)
	c.factory = NewFactoryGroup(k, cfg)
	c.informer = NewInformer(c.factory, InformerHandlers{
		AddFunc:    c.add,
		UpdateFunc: c.update,
		DeleteFunc: c.delete,
	})
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
	c.worker.Start(interrupt)
	c.logger.Info("Started workers")
}

func (c *Controller) Shutdown() error {
	c.worker.Shutdown()
	return nil
}

func (c *Controller) Notify() <-chan error {
	return c.notify
}

func (c *Controller) add(obj *apis.StandSchedulePolicy) {
	c.logger.Debug("Added policy object with name", zap.String("policy_name", obj.Name))
}

func (c *Controller) update(oldObj, newObj *apis.StandSchedulePolicy) {
	c.logger.Debug("Sync policy object with name", zap.String("policy_name", newObj.Name))
}

func (c *Controller) delete(obj *apis.StandSchedulePolicy) {
	c.logger.Debug("Deleted policy object with name", zap.String("policy_name", obj.Name))
}

func (c *Controller) reconcile(key string) error {
	return nil
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
