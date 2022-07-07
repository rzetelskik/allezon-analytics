package allezon_analytics

import (
	"golang.org/x/exp/constraints"
	"sync"
)

type MapStore struct {
	sync.RWMutex
	m map[string][]UserTag
}

func max[T constraints.Ordered](a, b T) T {
	if a < b {
		return b
	}
	return a
}

func (s *MapStore) Append(key string, value UserTag) error {
	s.Lock()
	arr := s.m[key]
	s.m[key] = append(arr[max(len(arr)-199, 0):], value)
	s.Unlock()

	return nil
}

func (s *MapStore) Get(key string) ([]UserTag, error) {
	s.RLock()
	v, ok := s.m[key]
	s.RUnlock()

	if !ok {
		v = make([]UserTag, 0)
	}

	return v, nil
}
