package controller

import (
	"fmt"

	"go.uber.org/zap"

	util "k8s.io/apimachinery/pkg/util/runtime"

	"github.com/dodopizza/stand-schedule-policy-controller/internal/azure"
	"github.com/dodopizza/stand-schedule-policy-controller/internal/kubernetes"
	apis "github.com/dodopizza/stand-schedule-policy-controller/pkg/apis/standschedules/v1"
	"github.com/dodopizza/stand-schedule-policy-controller/pkg/clock"
)

type (
	Controller struct {
		notify   chan error
		logger   *zap.Logger
		clock    clock.Interface
		kube     kubernetes.Interface
		azure    azure.Interface
		factory  *FactoryGroup
		informer *Informer
	}
)

func NewController(
	k kubernetes.Interface,
	az azure.Interface,
	l *zap.Logger,
	cfg *Config,
	clock clock.Interface,
) *Controller {
	c := &Controller{
		notify: make(chan error, 1),
		logger: l.Named("controller"),
		clock:  clock,
		kube:   k,
		azure:  az,
	}
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
}

func (c *Controller) Shutdown() error {
	return nil
}

func (c *Controller) Notify() <-chan error {
	return c.notify
}

func (c *Controller) add(obj *apis.StandSchedulePolicy) {
	c.logger.Debug("Added object with name", zap.String("object", obj.Name))
}

func (c *Controller) update(oldObj, newObj *apis.StandSchedulePolicy) {
	c.logger.Debug("Updated object with name", zap.String("object", newObj.Name))
}

func (c *Controller) delete(obj *apis.StandSchedulePolicy) {
	c.logger.Debug("Deleted object with name", zap.String("object", obj.Name))
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
