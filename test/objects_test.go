//go:build integration

package test

import (
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	apis "github.com/dodopizza/stand-schedule-policy-controller/pkg/apis/standschedules/v1"
	"github.com/dodopizza/stand-schedule-policy-controller/pkg/util"
)

func podObject(namespace, name string) *core.Pod {
	return &core.Pod{
		ObjectMeta: meta.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: core.PodSpec{
			Containers: []core.Container{
				{
					Name:  "test",
					Image: "nginx",
				},
			},
		},
	}
}

func deploymentObject(namespace, name string) *apps.Deployment {
	return &apps.Deployment{
		ObjectMeta: meta.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: apps.DeploymentSpec{
			Replicas: util.Pointer(int32(3)),
			Selector: &meta.LabelSelector{
				MatchLabels: map[string]string{
					"app": name,
				},
			},
			Template: core.PodTemplateSpec{
				ObjectMeta: meta.ObjectMeta{
					Labels: map[string]string{
						"app": name,
					},
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
		},
	}
}

func disabledDeploymentObject(namespace, name string) *apps.Deployment {
	return &apps.Deployment{
		ObjectMeta: meta.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Annotations: map[string]string{
				apis.AnnotationPrefix + "/restore-replicas": "3",
			},
		},
		Spec: apps.DeploymentSpec{
			Replicas: util.Pointer(int32(0)),
			Selector: &meta.LabelSelector{
				MatchLabels: map[string]string{
					"app": name,
				},
			},
			Template: core.PodTemplateSpec{
				ObjectMeta: meta.ObjectMeta{
					Labels: map[string]string{
						"app": name,
					},
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
		},
	}
}
