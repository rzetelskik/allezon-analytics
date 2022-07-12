package api

import (
	"encoding/json"
	"fmt"
)

type Action int

const (
	VIEW Action = iota
	BUY
)

func (a Action) MarshalJSON() ([]byte, error) {
	var s string

	switch a {
	case VIEW:
		s = `"VIEW"`
	case BUY:
		s = `"BUY"`
	default:
		return nil, fmt.Errorf("invalid action")
	}

	return []byte(s), nil
}

func (a *Action) UnmarshalJSON(data []byte) error {
	var err error
	var s string

	err = json.Unmarshal(data, &s)
	if err != nil {
		return fmt.Errorf("can't unmarshal to string: %w", err)
	}

	switch s {
	case "VIEW":
		*a = VIEW
	case "BUY":
		*a = BUY
	default:
		return fmt.Errorf("action '%s' is not supported", s)
	}

	return nil
}
