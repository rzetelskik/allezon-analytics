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

const compressedBinName = "data"

type AerospikeStore[T any] struct {
	Client *as.Client

	Policy *as.BasePolicy

	Namespace string
	Set       string

	Compress bool
}

func (s *AerospikeStore[T]) getKey(key string) (*as.Key, error) {
	return as.NewKey(s.Namespace, s.Set, key)
}

func (s *AerospikeStore[T]) convertRecord(record *as.Record, i *T) error {
	raw, ok := record.Bins[compressedBinName]
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
		writePolicy := as.NewWritePolicy(0, 0)
		writePolicy.GenerationPolicy = as.EXPECT_GEN_EQUAL

		if s.Compress {
			var record *as.Record
			record, err = s.Client.Get(s.Policy, asKey)
			if err != nil {
				if !errors.Is(err, as.ErrKeyNotFound) {
					return fmt.Errorf("can't get record: %w", err)
				}
				writePolicy.Generation = 0
			} else {
				writePolicy.Generation = record.Generation

				err = s.convertRecord(record, i)
				if err != nil {
					return fmt.Errorf("can't convert record: %w", err)
				}
			}
		} else {
			obj := &struct {
				Internal T
				Gen      uint32 `asm:"gen" json:"-"`
			}{
				Internal: *i,
			}
			err = s.Client.GetObject(s.Policy, asKey, obj)
			if err != nil {
				if !errors.Is(err, as.ErrKeyNotFound) {
					return fmt.Errorf("can't get object: %w", err)
				}
				writePolicy.Generation = 0
			} else {
				i = &obj.Internal
				writePolicy.Generation = obj.Gen
			}
		}

		err = modify(i)
		if err != nil {
			return fmt.Errorf("can't modify record: %w", err)
		}

		if s.Compress {
			var data []byte
			data, err = json.Marshal(*i)
			if err != nil {
				return fmt.Errorf("can't marshal data: %w", err)
			}

			encoded := snappy.Encode(nil, data)

			err = s.Client.PutBins(writePolicy, asKey, as.NewBin(compressedBinName, encoded))
		} else {
			err = s.Client.PutObject(writePolicy, asKey, i)
		}

		if errors.Is(err, &as.AerospikeError{ResultCode: ast.GENERATION_ERROR}) {
			klog.V(3).InfoS("can't modify record due to generation mismatch", "key", key, "attempt", retryCount)
			continue
		}
		if err != nil {
			return fmt.Errorf("can't put data: %w", err)
		}

		break
	}

	if retryCount == maxRetries {
		return errors.New("max retries exceeded")
	}

	return nil
}

func (s *AerospikeStore[T]) getObject(key *as.Key, i *T, errorOnNotFound bool) error {
	var err error

	err = s.Client.GetObject(s.Policy, key, i)

	if err != nil && (!errors.Is(err, as.ErrKeyNotFound) || errorOnNotFound) {
		return fmt.Errorf("can't get object: %w", err)
	}

	return nil
}

func (s *AerospikeStore[T]) getRecord(key *as.Key, i *T, errorOnNotFound bool) error {
	var err error

	var record *as.Record
	record, err = s.Client.Get(s.Policy, key)
	if err != nil && (!errors.Is(err, as.ErrKeyNotFound) || errorOnNotFound) {
		return fmt.Errorf("can't get record: %w", err)
	}

	err = s.convertRecord(record, i)
	if err != nil {
		return fmt.Errorf("can't convert record: %w", err)
	}

	return nil
}

func (s *AerospikeStore[T]) Get(key string, i *T, errorOnNotFound bool) error {
	asKey, err := s.getKey(key)
	if err != nil {
		return fmt.Errorf("can't get key: %w", err)
	}

	if s.Compress {
		return s.getRecord(asKey, i, errorOnNotFound)
	} else {
		return s.getObject(asKey, i, errorOnNotFound)
	}
}
