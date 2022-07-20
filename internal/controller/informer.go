package controller

import (
	"errors"

	"k8s.io/client-go/tools/cache"

	api "github.com/dodopizza/stand-schedule-policy-controller/pkg/apis/standschedules/v1"
	stands "github.com/dodopizza/stand-schedule-policy-controller/pkg/client/informers/externalversions/standschedules/v1"
)

type (
	Informer struct {
		stands   stands.StandSchedulePolicyInformer
		handlers InformerHandlers
	}

	InformerHandlers struct {
		AddFunc    func(obj *api.StandSchedulePolicy)
		UpdateFunc func(oldObj, newObj *api.StandSchedulePolicy)
		DeleteFunc func(obj *api.StandSchedulePolicy)
	}
)

func NewInformer(f *FactoryGroup, h InformerHandlers) *Informer {
	i := &Informer{
		stands:   f.stands.StandSchedules().V1().StandSchedulePolicies(),
		handlers: h,
	}
	i.stands.Informer().AddEventHandler(i)

	return i
}

func (ig *Informer) OnAdd(obj interface{}) {
	pto, err := toStandSchedulePolicyObject(obj)
	if err != nil {
		return
	}
	ig.handlers.AddFunc(pto)
}

func (ig *Informer) OnUpdate(oldObj, newObj interface{}) {
	oldPTO, err := toStandSchedulePolicyObject(oldObj)
	if err != nil {
		return
	}
	newPTO, err := toStandSchedulePolicyObject(newObj)
	if err != nil {
		return
	}
	ig.handlers.UpdateFunc(oldPTO, newPTO)
}

func (ig *Informer) OnDelete(obj interface{}) {
	pto, err := toStandSchedulePolicyObject(obj)
	if err != nil {
		return
	}
	ig.handlers.DeleteFunc(pto)
}

func toStandSchedulePolicyObject(obj interface{}) (*api.StandSchedulePolicy, error) {
	switch o := obj.(type) {
	case *api.StandSchedulePolicy:
		return o, nil
	case cache.DeletedFinalStateUnknown:
		return toStandSchedulePolicyObject(o.Obj)
	default:
		return nil, errors.New("unknown type found")
	}
}
