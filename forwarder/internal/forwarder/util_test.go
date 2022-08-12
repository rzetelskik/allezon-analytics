package forwarder

import (
	"github.com/google/go-cmp/cmp"
	"reflect"
	"testing"
)

func TestBacktrack(t *testing.T) {
	ts := []struct {
		name     string
		set      []string
		expected [][]string
	}{
		{
			name: "All subsets are generated",
			set:  []string{"origin", "brand_id", "category_id"},
			expected: [][]string{
				{},
				{"origin"},
				{"origin", "brand_id"},
				{"origin", "brand_id", "category_id"},
				{"origin", "category_id"},
				{"brand_id"},
				{"brand_id", "category_id"},
				{"category_id"},
			},
		},
	}

	t.Parallel()
	for _, test := range ts {
		t.Run(test.name, func(t *testing.T) {
			res := make([][]string, 0)
			Backtrack([]string{}, test.set, &res)
			if !reflect.DeepEqual(test.expected, res) {
				t.Errorf("expected and computed result differ: %s", cmp.Diff(test.expected, res))
			}
		})
	}
}
