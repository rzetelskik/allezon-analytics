package util

import (
	"github.com/rzetelskik/allezon-analytics/shared/pkg/api"
	"reflect"
	"testing"
	"time"
)

func TestGetAggregateHash(t *testing.T) {
	bucket := time.Now()

	ts := []struct {
		name     string
		bucket   time.Time
		action   api.Action
		filters  []string
		expected string
	}{
		{
			name:    "Providing empty strings in filters does not affect the copmputed value",
			bucket:  bucket,
			action:  api.VIEW,
			filters: []string{"Knitwear", "", "", ""},
			expected: func() string {
				return GetAggregateHash(bucket, api.VIEW, []string{"Knitwear"}...)
			}(),
		},
	}

	t.Parallel()
	for _, test := range ts {
		t.Run(test.name, func(t *testing.T) {
			res := GetAggregateHash(test.bucket, test.action, test.filters...)
			if !reflect.DeepEqual(test.expected, res) {
				t.Errorf("expected and computed values differ")
			}
		})
	}
}
