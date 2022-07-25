package controller

import (
	core "k8s.io/client-go/listers/core/v1"

	stands "github.com/dodopizza/stand-schedule-policy-controller/pkg/client/listers/standschedules/v1"
)

type (
	ListerGroup struct {
		ns     core.NamespaceLister
		stands stands.StandSchedulePolicyLister
	}
)

func NewListerGroup(f *FactoryGroup) *ListerGroup {
	return &ListerGroup{
		ns:     f.core.Core().V1().Namespaces().Lister(),
		stands: f.stands.StandSchedules().V1().StandSchedulePolicies().Lister(),
	}
}
