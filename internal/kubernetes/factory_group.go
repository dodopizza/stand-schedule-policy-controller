package kubernetes

import (
	"reflect"
	"time"

	core "k8s.io/client-go/informers"

	stands "github.com/dodopizza/stand-schedule-policy-controller/pkg/client/informers/externalversions"
)

type (
	FactoryGroup struct {
		Core   core.SharedInformerFactory
		Stands stands.SharedInformerFactory
	}
)

func NewFactoryGroup(k Interface, resyncInterval time.Duration) *FactoryGroup {
	return &FactoryGroup{
		Core:   core.NewSharedInformerFactory(k.CoreClient(), resyncInterval),
		Stands: stands.NewSharedInformerFactory(k.StandSchedulesClient(), resyncInterval),
	}
}

func (fg *FactoryGroup) Start(interrupt <-chan struct{}) {
	fg.Core.Start(interrupt)
	fg.Stands.Start(interrupt)
}

func (fg *FactoryGroup) WaitForCacheSync(interrupt <-chan struct{}) map[reflect.Type]bool {
	result := map[reflect.Type]bool{}
	for k, v := range fg.Core.WaitForCacheSync(interrupt) {
		result[k] = v
	}
	for k, v := range fg.Stands.WaitForCacheSync(interrupt) {
		result[k] = v
	}
	return result
}
