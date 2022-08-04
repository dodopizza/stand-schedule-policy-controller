package util

import (
	"go.uber.org/multierr"
)

func Reverse[T any](source []T) (ret []T) {
	for _, n := range source {
		ret = append([]T{n}, ret...)
	}
	return ret
}

func Where[T any](source []T, f func(i int, t T) bool) (ret []T) {
	for i, t := range source {
		if f(i, t) {
			ret = append(ret, t)
		}
	}
	return ret
}

func ForEachE[T any](source []T, f func(i int, t T) error) (err error) {
	for i, t := range source {
		err = multierr.Append(err, f(i, t))
	}
	return err
}
