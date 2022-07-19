package controller

import (
	"go.uber.org/zap"

	"github.com/dodopizza/stand-schedule-policy-controller/internal/azure"
	"github.com/dodopizza/stand-schedule-policy-controller/internal/kubernetes"
)

type (
	Config     struct{}
	Controller struct {
		notify chan error
		logger *zap.Logger
		kube   kubernetes.Interface
		azure  azure.Interface
	}
)

func NewController(k kubernetes.Interface, az azure.Interface, l *zap.Logger, _ *Config) *Controller {
	return &Controller{
		notify: make(chan error, 1),
		logger: l.Named("controller"),
		kube:   k,
		azure:  az,
	}
}

func (c *Controller) Start(interrupt <-chan struct{}) {

}

func (c *Controller) Shutdown() error {
	return nil
}

func (c *Controller) Notify() <-chan error {
	return c.notify
}
