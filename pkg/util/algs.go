package util

import (
	"sync"

	"go.uber.org/multierr"
)

func Reverse[T any](source []T) (ret []T) {
	for _, n := range source {
		ret = append([]T{n}, ret...)
	}
	return ret
}

func MapKeys[K comparable, V any](m map[K]V) []K {
	r := make([]K, 0, len(m))
	for k := range m {
		r = append(r, k)
	}
	return r
}

func Where[T any](source []T, f func(i int, t T) bool) (ret []T) {
	for i, t := range source {
		if f(i, t) {
			ret = append(ret, t)
		}
	}
	return ret
}

func Project[T any, D any](source []T, f func(i int, t T) D) (ret []D) {
	for i, t := range source {
		ret = append(ret, f(i, t))
	}
	return ret
}

func ForEachE[T any](source []T, f func(i int, t T) error) error {
	var err error

	for i, t := range source {
		err = multierr.Append(err, f(i, t))
	}
	return err
}

func ForEachParallelE[T any](source []T, f func(i int, t T) error) error {
	errors := make([]error, len(source))
	wg := &sync.WaitGroup{}
	wg.Add(len(source))

	for i, resource := range source {
		i := i
		resource := resource

		go func() {
			errors[i] = f(i, resource)
			wg.Done()
		}()
	}

	wg.Wait()

	return multierr.Combine(errors...)
}
