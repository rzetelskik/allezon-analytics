package server

import (
	"github.com/google/go-cmp/cmp"
	"reflect"
	"testing"
)

type insertIntoSortedSliceTestCase[T any] struct {
	name     string
	x        T
	xs       []T
	f        func(T, []T) func(int) bool
	expected []T
}

func runInsertIntoSortedSliceTestCase[T any](test insertIntoSortedSliceTestCase[T]) func(t *testing.T) {
	return func(t *testing.T) {
		res := InsertIntoSortedSlice(test.x, test.xs, test.f)

		if !reflect.DeepEqual(test.expected, res) {
			t.Errorf("expected and computed results differ: %s", cmp.Diff(test.expected, res))
		}
	}
}

func TestInsertIntoSortedSlice(t *testing.T) {
	testInt := func(t *testing.T) {
		ts := []insertIntoSortedSliceTestCase[int]{
			{
				name: "Insert int into non-empty slice",
				x:    3,
				xs:   []int{1, 2, 4, 5},
				f: func(x int, xs []int) func(int) bool {
					return func(i int) bool {
						return x < xs[i]
					}
				},
				expected: []int{1, 2, 3, 4, 5},
			},
			{
				name: "Insert int into empty slice",
				x:    3,
				xs:   []int{},
				f: func(x int, xs []int) func(int) bool {
					return func(i int) bool {
						return xs[i] < x
					}
				},
				expected: []int{3},
			},
		}

		t.Parallel()
		for _, test := range ts {
			t.Run(test.name, runInsertIntoSortedSliceTestCase(test))
		}
	}

	t.Parallel()
	t.Run("int test", testInt)
}

type headSliceTestCase[T any] struct {
	name     string
	xs       []T
	n        int
	expected []T
}

func runHeadSliceTestCase[T any](test headSliceTestCase[T]) func(t *testing.T) {
	return func(t *testing.T) {
		res := HeadSlice(test.xs, test.n)

		if !reflect.DeepEqual(test.expected, res) {
			t.Errorf("expected and computed results differ: %s", cmp.Diff(test.expected, res))
		}
	}
}

func TestHeadSlice(t *testing.T) {
	testInt := func(t *testing.T) {
		ts := []headSliceTestCase[int]{
			{
				name:     "Get less than len",
				xs:       []int{1, 2, 3, 4},
				n:        1,
				expected: []int{1},
			},
			{
				name:     "Get more than len",
				xs:       []int{1, 2, 3, 4},
				n:        5,
				expected: []int{1, 2, 3, 4},
			},
		}

		t.Parallel()
		for _, test := range ts {
			t.Run(test.name, runHeadSliceTestCase(test))
		}
	}

	t.Parallel()
	t.Run("int test", testInt)
}
