package kubernetes

import (
	"k8s.io/apimachinery/pkg/api/errors"
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
