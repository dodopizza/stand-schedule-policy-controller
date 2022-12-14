/*
Copyright Dodo Engineering

Authored by The Infrastructure Platform Team.
*/

// Code generated by lister-gen. DO NOT EDIT.

package v1

import (
	v1 "github.com/dodopizza/stand-schedule-policy-controller/pkg/apis/standschedules/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// StandSchedulePolicyLister helps list StandSchedulePolicies.
// All objects returned here must be treated as read-only.
type StandSchedulePolicyLister interface {
	// List lists all StandSchedulePolicies in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1.StandSchedulePolicy, err error)
	// Get retrieves the StandSchedulePolicy from the index for a given name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1.StandSchedulePolicy, error)
	StandSchedulePolicyListerExpansion
}

// standSchedulePolicyLister implements the StandSchedulePolicyLister interface.
type standSchedulePolicyLister struct {
	indexer cache.Indexer
}

// NewStandSchedulePolicyLister returns a new StandSchedulePolicyLister.
func NewStandSchedulePolicyLister(indexer cache.Indexer) StandSchedulePolicyLister {
	return &standSchedulePolicyLister{indexer: indexer}
}

// List lists all StandSchedulePolicies in the indexer.
func (s *standSchedulePolicyLister) List(selector labels.Selector) (ret []*v1.StandSchedulePolicy, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1.StandSchedulePolicy))
	})
	return ret, err
}

// Get retrieves the StandSchedulePolicy from the index for a given name.
func (s *standSchedulePolicyLister) Get(name string) (*v1.StandSchedulePolicy, error) {
	obj, exists, err := s.indexer.GetByKey(name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1.Resource("standschedulepolicy"), name)
	}
	return obj.(*v1.StandSchedulePolicy), nil
}
