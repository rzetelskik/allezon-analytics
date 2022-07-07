package server

import (
	"sort"
)

func InsertIntoSortedSlice[T any](x T, xs []T, f func(int) bool) []T {
	i := sort.Search(len(xs), f)
	xs = append(xs, x)
	copy(xs[i+1:], xs[i:])
	xs[i] = x
	return xs
}

func LimitSlice[T any](xs []T, limit int) []T {
	if len(xs) > limit {
		xs = xs[len(xs)-limit:]
	}
	return xs
}

func FilterSlice[T any](xs []T, f func(x T) bool) []T {
	tmp := make([]T, 0)
	for _, x := range xs {
		if f(x) {
			tmp = append(tmp, x)
		}
	}

	return tmp
}
