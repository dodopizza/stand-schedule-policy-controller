package kubernetes

import (
	core "k8s.io/client-go/listers/core/v1"

	stands "github.com/dodopizza/stand-schedule-policy-controller/pkg/client/listers/standschedules/v1"
)

type (
	ListerGroup struct {
		Namespaces core.NamespaceLister
		Stands     stands.StandSchedulePolicyLister
	}
)

func NewListerGroup(f *FactoryGroup) *ListerGroup {
	return &ListerGroup{
		Namespaces: f.Core.Core().V1().Namespaces().Lister(),
		Stands:     f.Stands.StandSchedules().V1().StandSchedulePolicies().Lister(),
	}
}
