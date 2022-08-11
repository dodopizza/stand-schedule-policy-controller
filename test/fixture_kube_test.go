//go:build integration

package test

import (
	"context"
	"time"

	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	apis "github.com/dodopizza/stand-schedule-policy-controller/pkg/apis/standschedules/v1"
)

const (
	_WaitPolicyStatusInterval = time.Second * 5
	_WaitPolicyStatusTimeout  = time.Minute * 3
)

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

func (f *fixture) WaitUntilPolicyStatus(name string, ct apis.ConditionType, sht apis.ConditionScheduleType) {
	err := wait.PollImmediate(_WaitPolicyStatusInterval, _WaitPolicyStatusTimeout, func() (bool, error) {
		f.t.Logf("Waiting policy (%s) status for %s to %s", name, sht, ct)
		policy, err := f.kube.StandSchedulesClient().
			StandSchedulesV1().
			StandSchedulePolicies().
			Get(context.Background(), name, meta.GetOptions{})

		if err != nil {
			return false, err
		}

		for _, status := range policy.Status.Conditions {
			if status.Type == ct && status.Status == sht {
				return true, nil
			}
		}

		return false, nil
	})

	if err != nil {
		f.t.Error(err)
	}
}
