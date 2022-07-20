//go:build integration

package test

import (
	"context"
	"testing"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/dodopizza/stand-schedule-policy-controller/internal/controller"
)

func (f *fixture) AssertKubernetesClient(t *testing.T) {
	_, err := f.kube.CoreClient().
		CoreV1().
		Pods("").
		List(context.Background(), meta.ListOptions{})

	if err != nil {
		t.Fatal(err)
	}

	_, err = f.kube.StandSchedulesClient().
		StandSchedulesV1().
		StandSchedulePolicies().
		List(context.Background(), meta.ListOptions{})

	if err != nil {
		t.Fatal(err)
	}
}

func (f *fixture) AssertControllerStarted(c *controller.Controller, t *testing.T) {
	if c == nil {
		t.Fail()
	}

	c.Start(f.interrupt)

	select {
	case err := <-c.Notify():
		t.Fatal(err)
	default:
	}
}

func Test_StartController(t *testing.T) {
	f := NewFixture(t)

	f.AssertKubernetesClient(t)
	f.AssertControllerStarted(f.CreateController(), t)
}
