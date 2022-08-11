package kubernetes

import (
	"k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func IgnoreNotFound(err error) error {
	return IgnoreError(err, errors.IsNotFound)
}

func IgnoreAlreadyExists(err error) error {
	return IgnoreError(err, errors.IsAlreadyExists)
}

func IgnoreTimeout(err error) error {
	return IgnoreError(err, errors.IsTimeout)
}

func IgnoreError(err error, handler func(error) bool) error {
	if handler(err) {
		return nil
	}
	return err
}

func SetAnnotation(m meta.ObjectMeta, name, val string) {
	if m.Annotations == nil {
		m.Annotations = make(map[string]string)
	}
	m.Annotations[name] = val
}

func GetAnnotation(m meta.ObjectMeta, name string) (string, bool) {
	if m.Annotations == nil {
		m.Annotations = make(map[string]string)
	}
	val, ok := m.Annotations[name]
	return val, ok
}
