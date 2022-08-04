package kubernetes

import (
	apps "k8s.io/client-go/listers/apps/v1"
	core "k8s.io/client-go/listers/core/v1"

	stands "github.com/dodopizza/stand-schedule-policy-controller/pkg/client/listers/standschedules/v1"
)

type (
	ListerGroup struct {
		Namespaces   core.NamespaceLister
		Deployments  apps.DeploymentLister
		StatefulSets apps.StatefulSetLister
		Stands       stands.StandSchedulePolicyLister
	}
)

func NewListerGroup(f *FactoryGroup) *ListerGroup {
	return &ListerGroup{
		Namespaces:   f.Core.Core().V1().Namespaces().Lister(),
		Deployments:  f.Core.Apps().V1().Deployments().Lister(),
		StatefulSets: f.Core.Apps().V1().StatefulSets().Lister(),
		Stands:       f.Stands.StandSchedules().V1().StandSchedulePolicies().Lister(),
	}
}
