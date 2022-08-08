package collector

import (
	"fmt"
	"github.com/rzetelskik/allezon-analytics/collector/internal/aerospike"
	"github.com/rzetelskik/allezon-analytics/shared/pkg/api"
	"sync"
)

type Store struct {
	sync.Mutex
	m map[string]api.UserAggregates
}

func NewStore() *Store {
	return &Store{
		m: make(map[string]api.UserAggregates),
	}
}

func (s *Store) Add(k string, agg api.UserAggregates) {
	s.Lock()
	defer s.Unlock()

	v := s.m[k]
	v.Count += agg.Count
	v.SumPrice += agg.SumPrice
	s.m[k] = v
}

func (s *Store) Dump(uas *aerospike.AerospikeStore) error {
	s.Lock()
	m := s.m
	s.m = make(map[string]api.UserAggregates)
	s.Unlock()

	// TODO append errors instead of breaking early
	for k, v := range m {
		err := uas.Add(k, v)
		if err != nil {
			return fmt.Errorf("can't add to database: %w", err)
		}
	}

	return nil
}
