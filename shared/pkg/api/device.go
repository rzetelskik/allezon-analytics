package api

import (
	"encoding/json"
	"fmt"
)

type Device int

const (
	PC Device = iota
	MOBILE
	TV
)

func (d Device) MarshalJSON() ([]byte, error) {
	var s string

	switch d {
	case PC:
		s = `"PC"`
	case MOBILE:
		s = `"MOBILE"`
	case TV:
		s = `"TV"`
	default:
		return nil, fmt.Errorf("invalid device")
	}

	return []byte(s), nil
}

func (d *Device) UnmarshalJSON(data []byte) error {
	var err error
	var s string

	err = json.Unmarshal(data, &s)
	if err != nil {
		return fmt.Errorf("can't unmarshal to string: %w", err)
	}

	switch s {
	case "PC":
		*d = PC
	case "MOBILE":
		*d = MOBILE
	case "TV":
		*d = TV
	default:
		return fmt.Errorf("device '%s' is not supported", s)
	}

	return nil
}
