//go:build integration

package test

import (
	"context"
	"testing"
	"time"

	core "k8s.io/api/core/v1"
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

func (f *fixture) AssertNamespaceEmpty(namespace string) {
	pods, err := f.kube.CoreClient().
		CoreV1().
		Pods(namespace).
		List(context.Background(), meta.ListOptions{})
	if err != nil {
		f.t.Error(err)
	}

	if len(pods.Items) != 0 {
		f.t.Errorf("In namespace %s exists %d pods", namespace, len(pods.Items))
	}
}

func Test_StartController(t *testing.T) {
	f := NewFixture(t)

	f.AssertKubernetesClient()
	f.AssertControllerStarted(f.CreateController())
}

func Test_ShutdownPolicy(t *testing.T) {
	f := NewFixture(t).
		WithNamespaces(
			"namespace1",
		).
		WithPods(
			&core.Pod{
				ObjectMeta: meta.ObjectMeta{
					Name:      "test-pod-1",
					Namespace: "namespace1",
				},
				Spec: core.PodSpec{
					Containers: []core.Container{
						{
							Name:  "test",
							Image: "nginx",
						},
					},
				},
			},
		).
		WithPolicies(
			&apis.StandSchedulePolicy{
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
			},
		)
	c := f.CreateController()
	f.AssertControllerStarted(c)

	f.DelayForWorkers(time.Second * 5)
	f.IncreaseTime(time.Minute * 2)
	f.DelayForWorkers(time.Second * 10)
	f.AssertNamespaceEmpty("namespace1")
}

func Test_ShutdownPolicyWithOverride(t *testing.T) {
	f := NewFixture(t).
		WithNamespaces(
			"namespace1",
		).
		WithPods(
			&core.Pod{
				ObjectMeta: meta.ObjectMeta{
					Name:      "test-pod-1",
					Namespace: "namespace1",
				},
				Spec: core.PodSpec{
					Containers: []core.Container{
						{
							Name:  "test",
							Image: "nginx",
						},
					},
				},
			},
		).
		WithPolicies(
			&apis.StandSchedulePolicy{
				ObjectMeta: meta.ObjectMeta{
					Name: "test-policy",
					Annotations: map[string]string{
						apis.AnnotationScheduleShutdownTime: _Time.Add(time.Second * 1).Format(time.RFC3339),
					},
				},
				Spec: apis.StandSchedulePolicySpec{
					TargetNamespaceFilter: "namespace1",
					Schedule: apis.ScheduleSpec{
						Startup:  "* * * * *",
						Shutdown: "0 23 * * *",
					},
					Resources: apis.ResourcesSpec{
						Azure: []apis.AzureResource{},
					},
				},
			},
		)
	c := f.CreateController()
	f.AssertControllerStarted(c)

	f.DelayForWorkers(time.Second * 5)
	f.IncreaseTime(time.Minute * 2)
	f.DelayForWorkers(time.Second * 10)
	f.AssertNamespaceEmpty("namespace1")
}
