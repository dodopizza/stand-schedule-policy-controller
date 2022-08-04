//go:build integration

package test

import (
	"context"
	"os"
	"testing"
	"time"

	"go.uber.org/zap"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	clock "k8s.io/utils/clock/testing"

	"github.com/dodopizza/stand-schedule-policy-controller/internal/controller"
	"github.com/dodopizza/stand-schedule-policy-controller/internal/kubernetes"
	apis "github.com/dodopizza/stand-schedule-policy-controller/pkg/apis/standschedules/v1"
	"github.com/dodopizza/stand-schedule-policy-controller/pkg/util"
)

var (
	_WorkingDir, _      = os.Getwd()
	_DefaultKubeCfgPath = _WorkingDir + "/../bin/kubeconfig.yaml"
	_KubeCfgEnvVar      = "TEST_KUBECONFIG_PATH"
	_Time               = time.Now()
)

type (
	fixture struct {
		cfg       *controller.Config
		kube      kubernetes.Interface
		azure     *azure
		clock     *clock.FakeClock
		interrupt chan struct{}
		t         *testing.T
		cleanup   *fixtureCleanup
	}
	fixtureCleanup struct {
		t          *testing.T
		kube       kubernetes.Interface
		policies   map[string]struct{}
		namespaces map[string]struct{}
		cleanup    bool
		interrupt  chan struct{}
		controller *controller.Controller
	}
	azure struct{}
)

func NewFixture(t *testing.T) *fixture {
	cfgPath := os.Getenv(_KubeCfgEnvVar)
	if cfgPath == "" {
		t.Setenv(_KubeCfgEnvVar, _DefaultKubeCfgPath)
	}

	k, err := kubernetes.NewForTest(_KubeCfgEnvVar)
	if err != nil {
		t.Fatal(err)
	}

	cleanup := NewFixtureCleanup(t, k)

	return &fixture{
		cfg: &controller.Config{
			ObjectsResyncSeconds:   10,
			PoliciesResyncSeconds:  10,
			WorkerQueueThreadiness: 1,
			WorkerQueueRetries:     5,
		},
		kube:      k,
		azure:     &azure{},
		clock:     clock.NewFakeClock(_Time),
		interrupt: cleanup.interrupt,
		t:         t,
		cleanup:   cleanup,
	}
}

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

func (f *fixture) WithClockTime(ts time.Time) *fixture {
	_Time = ts
	f.clock = clock.NewFakeClock(_Time)
	return f
}

func (f *fixture) WithNamespaces(namespaces ...string) *fixture {
	for _, ns := range namespaces {
		namespace := &core.Namespace{
			ObjectMeta: meta.ObjectMeta{
				Name: ns,
			},
		}
		_, err := f.kube.CoreClient().
			CoreV1().
			Namespaces().
			Create(context.Background(), namespace, meta.CreateOptions{})
		if err != nil {
			f.t.Fatal(err)
		}
		f.cleanup.AddNamespace(namespace)
	}
	return f
}

func (f *fixture) WithPods(pods ...*core.Pod) *fixture {
	for _, pod := range pods {
		_, err := f.kube.CoreClient().
			CoreV1().
			Pods(pod.Namespace).
			Create(context.Background(), pod, meta.CreateOptions{})
		if err != nil {
			f.t.Error(err)
		}
	}
	return f
}

func (f *fixture) WithDeployments(deployments ...*apps.Deployment) *fixture {
	for _, deployment := range deployments {
		_, err := f.kube.CoreClient().
			AppsV1().
			Deployments(deployment.Namespace).
			Create(context.Background(), deployment, meta.CreateOptions{})
		if err != nil {
			f.t.Error(err)
		}
	}
	return f
}

func (f *fixture) WithZeroQuota(namespace string) *fixture {
	quota := &core.ResourceQuota{
		ObjectMeta: meta.ObjectMeta{
			Name:      "zero-quota",
			Namespace: namespace,
		},
		Spec: core.ResourceQuotaSpec{
			Hard: core.ResourceList{
				core.ResourcePods: resource.MustParse("0"),
			},
		},
	}

	_, err := f.kube.CoreClient().
		CoreV1().
		ResourceQuotas(quota.Namespace).
		Create(context.Background(), quota, meta.CreateOptions{})

	if err != nil {
		f.t.Fatal(err)
	}
	return f
}

func (f *fixture) WithPolicies(policies ...*apis.StandSchedulePolicy) *fixture {
	for _, policy := range policies {
		_, err := f.kube.StandSchedulesClient().
			StandSchedulesV1().
			StandSchedulePolicies().
			Create(context.Background(), policy, meta.CreateOptions{})
		if err != nil {
			f.t.Fatal(err)
		}
		f.cleanup.AddPolicy(policy)
	}
	return f
}

func (f *fixture) WithoutCleanup() *fixture {
	f.cleanup.cleanup = false
	return f
}

func (f *fixture) CreateController() *controller.Controller {
	l, err := zap.NewDevelopment()
	if err != nil {
		f.t.Fatal(err)
	}

	f.cleanup.controller = controller.NewController(f.cfg, l, f.clock, f.kube, f.azure)

	return f.cleanup.controller
}

func (f *fixture) IncreaseTime(d time.Duration) {
	nextTime := _Time.Add(d)
	f.t.Logf("Increase controller time from %s to %s", _Time, nextTime)
	f.clock.SetTime(nextTime)
	_Time = nextTime
}

func (f *fixture) WaitUntilPolicyStatus(name string, ct apis.ConditionType, sht apis.ConditionScheduleType) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*3)
	defer cancel()

	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			f.t.Errorf("Deadline waiting status exceeded")
			return

		case <-ticker.C:
			f.t.Logf("Waiting policy (%s) status for %s to %s", name, sht, ct)
			policy, err := f.kube.StandSchedulesClient().
				StandSchedulesV1().
				StandSchedulePolicies().
				Get(ctx, name, meta.GetOptions{})

			if err != nil {
				f.t.Error(err)
			}

			for _, status := range policy.Status.Conditions {
				if status.Type == ct && status.Status == sht {
					return
				}
			}
		}
	}
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
			Delete(context.Background(), policy, meta.DeleteOptions{
				PropagationPolicy: util.Pointer(meta.DeletePropagationForeground),
			})
		if err != nil {
			f.t.Fatal(err)
		}
	}

	for namespace := range f.namespaces {
		err := f.kube.CoreClient().
			CoreV1().
			Namespaces().
			Delete(context.Background(), namespace, meta.DeleteOptions{
				PropagationPolicy: util.Pointer(meta.DeletePropagationForeground),
			})
		if err != nil {
			f.t.Fatal(err)
		}
	}
}
