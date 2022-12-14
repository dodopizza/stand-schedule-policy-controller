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

func (f *fixture) AssertDeploymentScaled(namespace string) {
	list, err := f.kube.CoreClient().
		AppsV1().
		Deployments(namespace).
		List(context.Background(), meta.ListOptions{})
	if err != nil {
		f.t.Errorf("Failed to list deployments in namespace %s", namespace)
	}

	for _, deployment := range list.Items {
		_, exists := deployment.Annotations[apis.AnnotationPrefix+"/restore-replicas"]
		if exists {
			f.t.Errorf("Restore annotation exists in namespace %s for deployment %s", namespace, deployment.Name)
		}

		if deployment.Spec.Replicas != nil && *deployment.Spec.Replicas == 0 {
			f.t.Errorf("Deployment %s replcias equal to zero in namespace %s", deployment.Name, namespace)
		}
	}
}

func Test_StartController(t *testing.T) {
	f := NewFixture(t)

	c := f.CreateController()

	f.AssertKubernetesClient()
	f.AssertControllerStarted(c)
}

func Test_PolicyWithShutdown(t *testing.T) {
	f := NewFixture(t).
		WithNamespaces("namespace1").
		WithPods(podObject("namespace1", "test-pod-1")).
		WithDeployments(deploymentObject("namespace1", "test-deployment-1")).
		WithPolicies(
			&apis.StandSchedulePolicy{
				ObjectMeta: meta.ObjectMeta{
					Name: "test-policy1",
				},
				Spec: apis.StandSchedulePolicySpec{
					TargetNamespaceFilter: "namespace1",
					Schedules: apis.SchedulesSpec{
						Startup:  apis.CronSchedule{Cron: "@yearly"},
						Shutdown: apis.CronSchedule{Cron: "* * * * *"},
					},
					Resources: apis.ResourcesSpec{},
				},
			},
		)
	c := f.CreateController()
	f.AssertControllerStarted(c)

	f.WaitUntilPolicyStatus("test-policy1", apis.ConditionScheduled, apis.StatusShutdown)
	f.IncreaseTime(time.Minute * 2)
	f.WaitUntilPolicyStatus("test-policy1", apis.ConditionCompleted, apis.StatusShutdown)
	f.AssertNamespaceEmptyOrPodsTerminated("namespace1")
}

func Test_PolicyWithStartup(t *testing.T) {
	f := NewFixture(t).
		WithNamespaces("namespace2").
		WithZeroQuota("namespace2").
		WithDeployments(disabledDeploymentObject("namespace2", "test-deployment-1")).
		WithPolicies(
			&apis.StandSchedulePolicy{
				ObjectMeta: meta.ObjectMeta{
					Name: "test-policy2",
				},
				Spec: apis.StandSchedulePolicySpec{
					TargetNamespaceFilter: "namespace2",
					Schedules: apis.SchedulesSpec{
						Startup:  apis.CronSchedule{Cron: "* * * * *"},
						Shutdown: apis.CronSchedule{Cron: "@yearly"},
					},
					Resources: apis.ResourcesSpec{},
				},
			},
		)

	c := f.CreateController()
	f.AssertControllerStarted(c)

	f.WaitUntilPolicyStatus("test-policy2", apis.ConditionScheduled, apis.StatusStartup)
	f.IncreaseTime(time.Minute * 2)
	f.WaitUntilPolicyStatus("test-policy2", apis.ConditionCompleted, apis.StatusStartup)
	f.AssertResourceQuotaNotExists("namespace2")
	f.AssertDeploymentScaled("namespace2")
}

func Test_PolicyWithShutdownOverride(t *testing.T) {
	f := NewFixture(t).
		WithNamespaces("namespace3").
		WithPods(podObject("namespace3", "test-pod-1")).
		WithDeployments(deploymentObject("namespace3", "test-deployment-1")).
		WithPolicies(
			&apis.StandSchedulePolicy{
				ObjectMeta: meta.ObjectMeta{
					Name: "test-policy3",
				},
				Spec: apis.StandSchedulePolicySpec{
					TargetNamespaceFilter: "namespace3",
					Schedules: apis.SchedulesSpec{
						Startup: apis.CronSchedule{Cron: "@yearly"},
						Shutdown: apis.CronSchedule{
							Cron:     "@yearly",
							Override: _Time.Add(time.Second * 1).Format(time.RFC3339),
						},
					},
					Resources: apis.ResourcesSpec{},
				},
			},
		)
	c := f.CreateController()
	f.AssertControllerStarted(c)

	f.WaitUntilPolicyStatus("test-policy3", apis.ConditionScheduled, apis.StatusShutdown)
	f.IncreaseTime(time.Minute * 2)
	f.WaitUntilPolicyStatus("test-policy3", apis.ConditionCompleted, apis.StatusShutdown)
	f.AssertNamespaceEmptyOrPodsTerminated("namespace3")
}

func Test_PolicyWithStartupOverride(t *testing.T) {
	f := NewFixture(t).
		WithNamespaces("namespace4").
		WithZeroQuota("namespace4").
		WithDeployments(disabledDeploymentObject("namespace4", "test-deployment-1")).
		WithPolicies(
			&apis.StandSchedulePolicy{
				ObjectMeta: meta.ObjectMeta{
					Name: "test-policy4",
				},
				Spec: apis.StandSchedulePolicySpec{
					TargetNamespaceFilter: "namespace4",
					Schedules: apis.SchedulesSpec{
						Startup: apis.CronSchedule{
							Cron:     "@yearly",
							Override: _Time.Add(time.Second * 1).Format(time.RFC3339),
						},
						Shutdown: apis.CronSchedule{Cron: "@yearly"},
					},
					Resources: apis.ResourcesSpec{},
				},
			},
		)
	c := f.CreateController()
	f.AssertControllerStarted(c)

	f.WaitUntilPolicyStatus("test-policy4", apis.ConditionScheduled, apis.StatusStartup)
	f.IncreaseTime(time.Minute * 2)
	f.WaitUntilPolicyStatus("test-policy4", apis.ConditionCompleted, apis.StatusStartup)
	f.AssertResourceQuotaNotExists("namespace4")
	f.AssertDeploymentScaled("namespace4")
}

func Test_PolicyWithShutdownStartup(t *testing.T) {
	f := NewFixture(t).
		WithClockTime(_Time.Round(time.Minute * 10)).
		WithNamespaces("namespace5").
		WithPods(podObject("namespace5", "test-pod-1")).
		WithDeployments(deploymentObject("namespace5", "test-deployment-1")).
		WithPolicies(
			&apis.StandSchedulePolicy{
				ObjectMeta: meta.ObjectMeta{
					Name: "test-policy5",
				},
				Spec: apis.StandSchedulePolicySpec{
					TargetNamespaceFilter: "namespace5",
					Schedules: apis.SchedulesSpec{
						Startup:  apis.CronSchedule{Cron: "0/5 * * * *"},
						Shutdown: apis.CronSchedule{Cron: "0/3 * * * *"},
					},
					Resources: apis.ResourcesSpec{},
				},
			},
		)

	c := f.CreateController()
	f.AssertControllerStarted(c)

	// wait to policies scheduled
	f.WaitUntilPolicyStatus("test-policy5", apis.ConditionScheduled, apis.StatusShutdown)
	f.WaitUntilPolicyStatus("test-policy5", apis.ConditionScheduled, apis.StatusStartup)

	// increase time to trigger shutdown policy & assert
	f.IncreaseTime(time.Minute * 3)
	f.WaitUntilPolicyStatus("test-policy5", apis.ConditionCompleted, apis.StatusShutdown)
	f.AssertNamespaceEmptyOrPodsTerminated("namespace5")

	// increase time to trigger startup policy & assert
	f.IncreaseTime(time.Minute * 2)
	f.WaitUntilPolicyStatus("test-policy5", apis.ConditionCompleted, apis.StatusStartup)
	f.AssertResourceQuotaNotExists("namespace5")

	// increase time to > half of shutdown interval & assert shutdown scheduled
	f.IncreaseTime(time.Minute * 1)
	f.WaitUntilPolicyStatus("test-policy5", apis.ConditionScheduled, apis.StatusShutdown)

	// increase time to > half of startup interval & assert startup scheduled
	f.IncreaseTime(time.Minute * 3)
	f.WaitUntilPolicyStatus("test-policy5", apis.ConditionScheduled, apis.StatusStartup)
}

func Test_PolicyWithOverrides(t *testing.T) {
	f := NewFixture(t).
		WithClockTime(_Time.Round(time.Minute * 10)).
		WithNamespaces("namespace6").
		WithPods(podObject("namespace6", "test-pod-1")).
		WithDeployments(deploymentObject("namespace6", "test-deployment-1")).
		WithPolicies(
			&apis.StandSchedulePolicy{
				ObjectMeta: meta.ObjectMeta{
					Name: "test-policy6",
				},
				Spec: apis.StandSchedulePolicySpec{
					TargetNamespaceFilter: "namespace6",
					Schedules: apis.SchedulesSpec{
						Startup: apis.CronSchedule{
							Override: _Time.Add(time.Minute * 5).Format(time.RFC3339),
						},
						Shutdown: apis.CronSchedule{
							Override: _Time.Add(time.Minute * 2).Format(time.RFC3339),
						},
					},
					Resources: apis.ResourcesSpec{},
				},
			},
		)

	c := f.CreateController()
	f.AssertControllerStarted(c)

	// wait to policies scheduled
	f.WaitUntilPolicyStatus("test-policy6", apis.ConditionScheduled, apis.StatusShutdown)
	f.WaitUntilPolicyStatus("test-policy6", apis.ConditionScheduled, apis.StatusStartup)

	// increase time to trigger shutdown policy & assert
	f.IncreaseTime(time.Minute * 2)
	f.WaitUntilPolicyStatus("test-policy6", apis.ConditionCompleted, apis.StatusShutdown)
	f.AssertNamespaceEmptyOrPodsTerminated("namespace6")

	// increase time to trigger startup policy & assert
	f.IncreaseTime(time.Minute * 3)
	f.WaitUntilPolicyStatus("test-policy6", apis.ConditionCompleted, apis.StatusStartup)
	f.AssertResourceQuotaNotExists("namespace6")

	// increase time and verify policy remains completed
	f.IncreaseTime(time.Minute * 5)
	f.WaitUntilPolicyStatus("test-policy6", apis.ConditionCompleted, apis.StatusStartup)
	f.WaitUntilPolicyStatus("test-policy6", apis.ConditionCompleted, apis.StatusShutdown)
}

func Test_PolicyWithExternalResources(t *testing.T) {
	f := NewFixture(t).
		WithNamespaces("namespace7").
		WithPods(podObject("namespace7", "test-pod-1")).
		WithDeployments(deploymentObject("namespace7", "test-deployment-1")).
		WithAzureResources(
			azureMySQL("test-1-rg", "test-mysql-1"),
			azureMySQL("test-1-rg", "test-mysql-2"),
			azureVM("test-2-rg", "test-vm-1"),
			azureVM("test-2-rg", "test-vm-2"),
		).
		WithPolicies(
			&apis.StandSchedulePolicy{
				ObjectMeta: meta.ObjectMeta{
					Name: "test-policy7",
				},
				Spec: apis.StandSchedulePolicySpec{
					TargetNamespaceFilter: "namespace7",
					Schedules: apis.SchedulesSpec{
						Startup:  apis.CronSchedule{Cron: "@yearly"},
						Shutdown: apis.CronSchedule{Cron: "* * * * *"},
					},
					Resources: apis.ResourcesSpec{
						Azure: apis.AzureResourceList{
							{
								Type:               apis.AzureResourceVirtualMachine,
								ResourceGroupName:  "test-2-rg",
								ResourceNameFilter: "test-vm",
								Priority:           0,
							},
							{
								Type:               apis.AzureResourceManagedMySQL,
								ResourceGroupName:  "test-1-rg",
								ResourceNameFilter: "test-mysql",
								Priority:           1,
							},
						},
					},
				},
			},
		)

	c := f.CreateController()
	f.AssertControllerStarted(c)

	f.WaitUntilPolicyStatus("test-policy7", apis.ConditionScheduled, apis.StatusShutdown)
	f.IncreaseTime(time.Minute * 1)
	f.WaitUntilPolicyStatus("test-policy7", apis.ConditionCompleted, apis.StatusShutdown)
	f.AssertNamespaceEmptyOrPodsTerminated("namespace7")
}
