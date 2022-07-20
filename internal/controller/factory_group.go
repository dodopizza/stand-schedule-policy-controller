package controller

import (
	"reflect"

	core "k8s.io/client-go/informers"

	"github.com/dodopizza/stand-schedule-policy-controller/internal/kubernetes"
	stands "github.com/dodopizza/stand-schedule-policy-controller/pkg/client/informers/externalversions"
)

type (
	FactoryGroup struct {
		core   core.SharedInformerFactory
		stands stands.SharedInformerFactory
	}
)

func NewFactoryGroup(k kubernetes.Interface, cfg *Config) *FactoryGroup {
	return &FactoryGroup{
		core:   core.NewSharedInformerFactory(k.CoreClient(), cfg.GetResyncDuration()),
		stands: stands.NewSharedInformerFactory(k.StandSchedulesClient(), cfg.GetResyncDuration()),
	}
}

func (fg *FactoryGroup) WaitForCacheSync(interrupt <-chan struct{}) map[reflect.Type]bool {
	result := map[reflect.Type]bool{}
	for k, v := range fg.core.WaitForCacheSync(interrupt) {
		result[k] = v
	}
	for k, v := range fg.stands.WaitForCacheSync(interrupt) {
		result[k] = v
	}
	return result
}
