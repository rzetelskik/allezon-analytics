package aerospike

import (
	"errors"
	"fmt"
	as "github.com/aerospike/aerospike-client-go/v6"
	"github.com/rzetelskik/allezon-analytics/shared/pkg/api"
)

type UserAggregatesStore struct {
	Client *as.Client

	Policy *as.BasePolicy

	Namespace string
	Set       string
}

func (s *UserAggregatesStore) getKey(key string) (*as.Key, error) {
	return as.NewKey(s.Namespace, s.Set, key)
}

func (s *UserAggregatesStore) Get(key string, i *api.UserAggregates) error {
	asKey, err := s.getKey(key)

	if err != nil {
		return fmt.Errorf("can't get key: %w", err)
	}

	err = s.Client.GetObject(s.Policy, asKey, i)
	if err != nil {
		if errors.Is(err, as.ErrKeyNotFound) {
			return nil
		} else {
			return fmt.Errorf("can't get object: %w", err)
		}
	}

	//err = s.convertRecord(record, i)
	//if err != nil {
	//	return fmt.Errorf("can't convert record: %w", err)
	//}

	return nil
}
