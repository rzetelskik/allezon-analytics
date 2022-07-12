package api

import (
	"encoding/json"
	"fmt"
	"time"
)

type UserTag struct {
	Time    time.Time `json:"time"`
	Cookie  string    `json:"cookie"`
	Country string    `json:"country"`
	Device  Device    `json:"device"`
	Action  Action    `json:"action"`
	Origin  string    `json:"origin"`
	Product Product   `json:"product_info"`
}

func (ut *UserTag) UnmarshalJSON(data []byte) error {
	var err error

	type Alias UserTag
	aux := &struct {
		Time string `json:"time"`
		*Alias
	}{
		Alias: (*Alias)(ut),
	}

	err = json.Unmarshal(data, &aux)
	if err != nil {
		return err
	}

	ut.Time, err = ParseDatetimeWithZone(aux.Time)
	if err != nil {
		return fmt.Errorf("can't parse time: %w", err)
	}

	return nil
}
