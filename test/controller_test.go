//go:build integration

package test

import (
	"context"
	"testing"
	"time"

	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
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

func (f *fixture) AssertResourceQuotaNotExists(namespace string) {
	quotaName := "zero-quota"
	_, err := f.kube.CoreClient().
		CoreV1().
		ResourceQuotas(namespace).
		Get(context.Background(), quotaName, meta.GetOptions{})

	if err == nil {
		f.t.Errorf("Resource quota %s exists in namespace %s", quotaName, namespace)
	}

	if !errors.IsNotFound(err) {
		f.t.Error(err)
	}
}

func Test_StartController(t *testing.T) {
	f := NewFixture(t)

	f.AssertKubernetesClient()
	f.AssertControllerStarted(f.CreateController())
}

func Test_PolicyWithShutdown(t *testing.T) {
	f := NewFixture(t).
		WithNamespaces("namespace1").
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
						Startup:  "@yearly",
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

func Test_PolicyWithStartup(t *testing.T) {
	f := NewFixture(t).
		WithNamespaces("namespace1").
		WithZeroQuota("namespace1").
		WithPolicies(
			&apis.StandSchedulePolicy{
				ObjectMeta: meta.ObjectMeta{
					Name: "test-policy",
				},
				Spec: apis.StandSchedulePolicySpec{
					TargetNamespaceFilter: "namespace1",
					Schedule: apis.ScheduleSpec{
						Startup:  "* * * * *",
						Shutdown: "@yearly",
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
	f.AssertResourceQuotaNotExists("namespace1")
}

func Test_PolicyWithShutdownOverride(t *testing.T) {
	f := NewFixture(t).
		WithNamespaces("namespace1").
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
						Startup:  "@yearly",
						Shutdown: "@yearly",
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

func Test_PolicyWithStartupOverride(t *testing.T) {
	f := NewFixture(t).
		WithNamespaces("namespace1").
		WithZeroQuota("namespace1").
		WithPolicies(
			&apis.StandSchedulePolicy{
				ObjectMeta: meta.ObjectMeta{
					Name: "test-policy",
					Annotations: map[string]string{
						apis.AnnotationScheduleStartupTime: _Time.Add(time.Second * 1).Format(time.RFC3339),
					},
				},
				Spec: apis.StandSchedulePolicySpec{
					TargetNamespaceFilter: "namespace1",
					Schedule: apis.ScheduleSpec{
						Startup:  "@yearly",
						Shutdown: "@yearly",
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
	f.AssertResourceQuotaNotExists("namespace1")
}
