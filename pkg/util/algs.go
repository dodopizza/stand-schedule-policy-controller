package util

func Reverse[T any](source []T) []T {
	var rev []T
	for _, n := range source {
		rev = append([]T{n}, rev...)
	}
	return rev
}
