package api

import (
	"encoding/json"
	"fmt"
)

type UserAggregates struct {
	Count    int64 `json:"count"`
	SumPrice int64 `json:"sum_price"`
}

type UserAggregatesCodec struct{}

func (c *UserAggregatesCodec) Encode(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (c *UserAggregatesCodec) Decode(data []byte) (interface{}, error) {
	var ua UserAggregates
	err := json.Unmarshal(data, &ua)
	if err != nil {
		return nil, fmt.Errorf("can't unmarshal data: %w", err)
	}

	return ua, nil
}
