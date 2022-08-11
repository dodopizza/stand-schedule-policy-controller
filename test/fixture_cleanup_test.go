//go:build integration

package test

import (
	"context"
	"testing"
	"time"

	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/dodopizza/stand-schedule-policy-controller/internal/controller"
	"github.com/dodopizza/stand-schedule-policy-controller/internal/kubernetes"
	apis "github.com/dodopizza/stand-schedule-policy-controller/pkg/apis/standschedules/v1"
)

type (
	fixtureCleanup struct {
		t          *testing.T
		kube       kubernetes.Interface
		policies   map[string]struct{}
		namespaces map[string]struct{}
		cleanup    bool
		interrupt  chan struct{}
		controller *controller.Controller
	}
)

const (
	_WaitNamespaceDeletionInterval = time.Second * 5
	_WaitNamespaceDeletionTimeout  = time.Minute * 10
)

func NewFixtureCleanup(t *testing.T, k kubernetes.Interface) *fixtureCleanup {
	f := &fixtureCleanup{
		t:          t,
		kube:       k,
		interrupt:  make(chan struct{}),
		policies:   make(map[string]struct{}, 0),
		namespaces: make(map[string]struct{}, 0),
		cleanup:    true,
	}
	t.Cleanup(f.Handler)

	return f
}

func (f *fixtureCleanup) AddPolicy(policy *apis.StandSchedulePolicy) {
	f.policies[policy.Name] = struct{}{}
}

func (f *fixtureCleanup) AddNamespace(namespace *core.Namespace) {
	f.namespaces[namespace.Name] = struct{}{}
}

func (f *fixtureCleanup) Handler() {
	f.t.Log("Invoke cleanup handler")
	close(f.interrupt)
	time.Sleep(time.Second * 1)

	if !f.cleanup {
		return
	}

	f.t.Log("Cleanup cluster objects")
	for policy := range f.policies {
		err := f.kube.StandSchedulesClient().
			StandSchedulesV1().
			StandSchedulePolicies().
			Delete(context.Background(), policy, meta.DeleteOptions{})
		if err != nil {
			f.t.Fatal(err)
		}
	}

	for namespace := range f.namespaces {
		f.t.Logf("Delete namespace %s", namespace)
		err := f.kube.CoreClient().
			CoreV1().
			Namespaces().
			Delete(context.Background(), namespace, meta.DeleteOptions{})
		if err != nil {
			f.t.Fatal(err)
		}

		f.t.Logf("Wait until namespace %s deleted", namespace)
		if err := f.WaitUntilNamespaceDeleted(namespace); err != nil {
			f.t.Fatal(err)
		}
	}
}

func (f *fixtureCleanup) WaitUntilNamespaceDeleted(namespace string) error {
	return wait.PollImmediate(_WaitNamespaceDeletionInterval, _WaitNamespaceDeletionTimeout, func() (bool, error) {
		f.t.Log("Wait interval..")

		_, err := f.kube.CoreClient().
			CoreV1().
			Namespaces().
			Get(context.Background(), namespace, meta.GetOptions{})

		if errors.IsNotFound(err) {
			return true, nil
		}

		return false, err
	})
}
