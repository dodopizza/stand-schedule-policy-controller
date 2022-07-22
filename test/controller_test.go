//go:build integration

package test

import (
	"context"
	"testing"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/dodopizza/stand-schedule-policy-controller/internal/controller"
	apis "github.com/dodopizza/stand-schedule-policy-controller/pkg/apis/standschedules/v1"
)

func (f *fixture) AssertKubernetesClient() {
	_, err := f.kube.CoreClient().
		CoreV1().
		Pods("").
		List(context.Background(), meta.ListOptions{})

	if err != nil {
		f.t.Fatal(err)
	}

	_, err = f.kube.StandSchedulesClient().
		StandSchedulesV1().
		StandSchedulePolicies().
		List(context.Background(), meta.ListOptions{})

	if err != nil {
		f.t.Fatal(err)
	}
}

func (f *fixture) AssertControllerStarted(c *controller.Controller) {
	if c == nil {
		f.t.Fail()
	}

	c.Start(f.interrupt)

	select {
	case err := <-c.Notify():
		f.t.Fatal(err)
	default:
	}
}

func Test_StartController(t *testing.T) {
	f := NewFixture(t)

	f.AssertKubernetesClient()
	f.AssertControllerStarted(f.CreateController())
}

func Test_CreateStandSchedulePolicy(t *testing.T) {
	f := NewFixture(t).
		WithPolicies(&apis.StandSchedulePolicy{
			ObjectMeta: meta.ObjectMeta{
				Name: "test-policy",
			},
			Spec: apis.StandSchedulePolicySpec{
				TargetNamespaceFilter: "namespace1",
				Schedule: apis.ScheduleSpec{
					Startup:  "* * * * *",
					Shutdown: "* * * * *",
				},
				Resources: apis.ResourcesSpec{
					Azure: []apis.AzureResource{},
				},
			},
		})
	c := f.CreateController()

	f.AssertControllerStarted(c)
}
