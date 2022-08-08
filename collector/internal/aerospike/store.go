package aerospike

import (
	"fmt"
	as "github.com/aerospike/aerospike-client-go/v6"
	"github.com/rzetelskik/allezon-analytics/shared/pkg/api"
)

type AerospikeStore struct {
	Client *as.Client

	Policy *as.BasePolicy

	Namespace string
	Set       string
}

func (s *AerospikeStore) getKey(key string) (*as.Key, error) {
	return as.NewKey(s.Namespace, s.Set, key)
}

func (s *AerospikeStore) Add(key string, agg api.UserAggregates) error {
	var err error

	asKey, err := s.getKey(key)
	if err != nil {
		return fmt.Errorf("can't get key: %w", err)
	}

	_, err = s.Client.Operate(s.Client.DefaultWritePolicy, asKey, as.AddOp(as.NewBin("count", agg.Count)), as.AddOp(as.NewBin("sum_price", agg.SumPrice)))
	if err != nil {
		return fmt.Errorf("can't perform operation: %w", err)
	}

	return nil
}
