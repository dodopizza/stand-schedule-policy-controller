//go:build integration

package test

import (
	"context"
	"testing"
	"time"

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

func (f *fixture) AssertNamespaceEmptyOrPodsTerminated(namespace string) {
	pods, err := f.kube.CoreClient().
		CoreV1().
		Pods(namespace).
		List(context.Background(), meta.ListOptions{})
	if err != nil {
		f.t.Error(err)
	}

	if len(pods.Items) == 0 {
		return
	}

	for _, pod := range pods.Items {
		for _, cs := range pod.Status.ContainerStatuses {
			if cs.State.Terminated == nil {
				f.t.Errorf("pod %s/%s container %s not in terminated state", pod.Namespace, pod.Name, cs.Name)
			}
		}
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
		WithPods(podObject("namespace1", "test-pod-1")).
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

	f.WaitUntilPolicyStatus("test-policy", apis.ConditionScheduled, apis.StatusShutdown)
	f.IncreaseTime(time.Minute * 2)
	f.WaitUntilPolicyStatus("test-policy", apis.ConditionCompleted, apis.StatusShutdown)
	f.AssertNamespaceEmptyOrPodsTerminated("namespace1")
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

	f.WaitUntilPolicyStatus("test-policy", apis.ConditionScheduled, apis.StatusStartup)
	f.IncreaseTime(time.Minute * 2)
	f.WaitUntilPolicyStatus("test-policy", apis.ConditionCompleted, apis.StatusStartup)
	f.AssertResourceQuotaNotExists("namespace1")
}

func Test_PolicyWithShutdownOverride(t *testing.T) {
	f := NewFixture(t).
		WithNamespaces("namespace1").
		WithPods(podObject("namespace1", "test-pod-1")).
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

	f.WaitUntilPolicyStatus("test-policy", apis.ConditionScheduled, apis.StatusShutdown)
	f.IncreaseTime(time.Minute * 2)
	f.WaitUntilPolicyStatus("test-policy", apis.ConditionCompleted, apis.StatusShutdown)
	f.AssertNamespaceEmptyOrPodsTerminated("namespace1")
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

	f.WaitUntilPolicyStatus("test-policy", apis.ConditionScheduled, apis.StatusStartup)
	f.IncreaseTime(time.Minute * 2)
	f.WaitUntilPolicyStatus("test-policy", apis.ConditionCompleted, apis.StatusStartup)
	f.AssertResourceQuotaNotExists("namespace1")
}

func Test_PolicyWithShutdownStartup(t *testing.T) {
	f := NewFixture(t).
		WithClockTime(_Time.Round(time.Minute * 10)).
		WithNamespaces("namespace1").
		WithPods(podObject("namespace1", "test-pod-1")).
		WithPolicies(
			&apis.StandSchedulePolicy{
				ObjectMeta: meta.ObjectMeta{
					Name: "test-policy",
				},
				Spec: apis.StandSchedulePolicySpec{
					TargetNamespaceFilter: "namespace1",
					Schedule: apis.ScheduleSpec{
						Startup:  "5 * * * *",
						Shutdown: "3 * * * *",
					},
					Resources: apis.ResourcesSpec{
						Azure: []apis.AzureResource{},
					},
				},
			},
		)

	c := f.CreateController()
	f.AssertControllerStarted(c)

	// wait to policies scheduled
	f.WaitUntilPolicyStatus("test-policy", apis.ConditionScheduled, apis.StatusShutdown)
	f.WaitUntilPolicyStatus("test-policy", apis.ConditionScheduled, apis.StatusStartup)

	// increase time to trigger shutdown policy & assert
	f.IncreaseTime(time.Minute * 3)
	f.WaitUntilPolicyStatus("test-policy", apis.ConditionCompleted, apis.StatusShutdown)
	f.AssertNamespaceEmptyOrPodsTerminated("namespace1")

	// increase time to trigger startup policy & assert
	f.IncreaseTime(time.Minute * 2)
	f.WaitUntilPolicyStatus("test-policy", apis.ConditionCompleted, apis.StatusStartup)
	f.AssertResourceQuotaNotExists("namespace-1")

	// increase time to > half of shutdown interval & assert shutdown scheduled
	f.IncreaseTime(time.Minute * 1)
	f.WaitUntilPolicyStatus("test-policy", apis.ConditionScheduled, apis.StatusShutdown)

	// increase time to > half of startup interval & assert startup scheduled
	f.IncreaseTime(time.Minute * 3)
	f.WaitUntilPolicyStatus("test-policy", apis.ConditionScheduled, apis.StatusStartup)
}
