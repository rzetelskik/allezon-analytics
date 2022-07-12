package server

import (
	"sort"
)

func InsertIntoSortedSlice[T any](x T, xs []T, f func([]T) func(int) bool) []T {
	i := sort.Search(len(xs), f(xs))
	xs = append(xs, x)
	copy(xs[i+1:], xs[i:])
	xs[i] = x
	return xs
}

func HeadSlice[T any](xs []T, n int) []T {
	if len(xs) > n {
		return xs[:n]
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
