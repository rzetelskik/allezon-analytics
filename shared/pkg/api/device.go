package api

import (
	"encoding/json"
	"fmt"
)

type Device int

const (
	PC Device = iota + 1
	MOBILE
	TV
)

var deviceToString = map[Device]string{
	PC:     "PC",
	MOBILE: "MOBILE",
	TV:     "TV",
}

var deviceFromString = map[string]Device{
	"PC":     PC,
	"MOBILE": MOBILE,
	"TV":     TV,
}

func (d Device) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d *Device) UnmarshalJSON(data []byte) error {
	var err error
	var s string

	err = json.Unmarshal(data, &s)
	if err != nil {
		return fmt.Errorf("can't unmarshal to string: %w", err)
	}

	*d, err = ParseDevice(s)
	if err != nil {
		return fmt.Errorf("can't parse device: %w", err)
	}

	return nil
}

func ParseDevice(s string) (Device, error) {
	d, ok := deviceFromString[s]
	if !ok {
		return Device(0), fmt.Errorf("%q is not a valid device", s)
	}

	return d, nil
}

func (d Device) String() string {
	return deviceToString[d]
}
