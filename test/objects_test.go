//go:build integration

package test

import (
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
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
