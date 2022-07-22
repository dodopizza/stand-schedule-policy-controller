package eventsource

import (
	"errors"

	"k8s.io/client-go/tools/cache"
)

type (
	InformerFactory interface {
		Informer() cache.SharedIndexInformer
	}

	EventSource[T any] struct {
		factory  InformerFactory
		handlers Handlers[T]
	}

	Handlers[T any] struct {
		AddFunc    func(obj *T)
		UpdateFunc func(oldObj, newObj *T)
		DeleteFunc func(obj *T)
	}
)

func New[T any](f InformerFactory, h Handlers[T]) *EventSource[T] {
	i := &EventSource[T]{
		factory:  f,
		handlers: h,
	}
	i.factory.Informer().AddEventHandler(i)

	return i
}

func (ig *EventSource[T]) OnAdd(obj interface{}) {
	o, err := ig.toObject(obj)
	if err != nil {
		return
	}
	ig.handlers.AddFunc(o)
}

func (ig *EventSource[T]) OnUpdate(oldObj, newObj interface{}) {
	o, err := ig.toObject(oldObj)
	if err != nil {
		return
	}
	n, err := ig.toObject(newObj)
	if err != nil {
		return
	}
	ig.handlers.UpdateFunc(o, n)
}

func (ig *EventSource[T]) OnDelete(obj interface{}) {
	o, err := ig.toObject(obj)
	if err != nil {
		return
	}
	ig.handlers.DeleteFunc(o)
}

func (ig *EventSource[T]) toObject(obj interface{}) (*T, error) {
	switch o := obj.(type) {
	case *T:
		return o, nil
	case cache.DeletedFinalStateUnknown:
		return ig.toObject(o.Obj)
	default:
		return nil, errors.New("unknown type found")
	}
}
