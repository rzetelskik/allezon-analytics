package aerospike

import (
	"encoding/json"
	"errors"
	"fmt"
	as "github.com/aerospike/aerospike-client-go/v6"
	ast "github.com/aerospike/aerospike-client-go/v6/types"
	"github.com/golang/snappy"
	"k8s.io/klog/v2"
)

type AerospikeStore[T interface{}] struct {
	Client *as.Client

	Policy *as.BasePolicy

	Namespace string
	Set       string
	Bin       string
}

func (s *AerospikeStore[T]) getKey(key string) (*as.Key, error) {
	return as.NewKey(s.Namespace, s.Set, key)
}

func (s *AerospikeStore[T]) convertRecord(record *as.Record, i *T) error {
	raw, ok := record.Bins[s.Bin]
	if !ok {
		return errors.New("can't get bin")
	}

	encoded, ok := raw.([]byte)
	if !ok {
		return errors.New("can't get byte data")
	}

	decoded, err := snappy.Decode(nil, encoded)
	if err != nil {
		return fmt.Errorf("can't decode data: %w", err)
	}

	err = json.Unmarshal(decoded, i)
	if err != nil {
		return fmt.Errorf("can't unmarshal data: %w", err)
	}

	return nil
}

func (s *AerospikeStore[T]) RMWWithGenCheck(key string, maxRetries int, i *T, modify func(*T) error) error {
	var err error

	asKey, err := s.getKey(key)
	if err != nil {
		return fmt.Errorf("can't get key: %w", err)
	}

	var retryCount int
	for retryCount = 0; retryCount < maxRetries; retryCount++ {
		var record *as.Record
		record, err = s.Client.Get(s.Policy, asKey)
		if err != nil {
			if !errors.Is(err, as.ErrKeyNotFound) {
				return fmt.Errorf("can't get record: %w", err)
			}
		} else {
			err = s.convertRecord(record, i)
			if err != nil {
				return fmt.Errorf("can't convert record: %w", err)
			}
		}

		err = modify(i)
		if err != nil {
			return fmt.Errorf("can't modify record")
		}

		writePolicy := as.NewWritePolicy(0, 0)
		writePolicy.GenerationPolicy = as.EXPECT_GEN_EQUAL
		if record == nil {
			writePolicy.Generation = 0
		} else {
			writePolicy.Generation = record.Generation
		}

		data, err := json.Marshal(*i)
		if err != nil {
			return fmt.Errorf("can't marshal data: %w", err)
		}

		encoded := snappy.Encode(nil, data)

		err = s.Client.PutBins(writePolicy, asKey, as.NewBin(s.Bin, encoded))
		if errors.Is(err, &as.AerospikeError{ResultCode: ast.GENERATION_ERROR}) {
			klog.V(3).InfoS("can't modify record due to generation mismatch", "key", key, "attempt", retryCount)
			continue
		}
		if err != nil {
			return fmt.Errorf("can't put bins: %w", err)
		}

		break
	}

	if retryCount == maxRetries {
		return errors.New("max retries exceeded")
	}

	return nil
}

func (s *AerospikeStore[T]) Get(key string, i *T) error {
	asKey, err := s.getKey(key)
	if err != nil {
		return fmt.Errorf("can't get key: %w", err)
	}

	record, err := s.Client.Get(s.Policy, asKey)
	if err != nil {
		return fmt.Errorf("can't get record: %w", err)
	}

	err = s.convertRecord(record, i)
	if err != nil {
		return fmt.Errorf("can't convert record: %w", err)
	}

	return nil
}
