package allezon_analytics

import (
	"encoding/json"
	"fmt"
	"time"
)

const RFC3339Milli = "2006-01-02T15:04:05.000Z"

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

	for _, f := range []string{RFC3339Milli, "2006-01-02T15:04:05Z"} { // FIXME: remove this after email
		ut.Time, err = time.Parse(f, aux.Time)
		if err == nil {
			break
		}
	}
	if err != nil {
		return fmt.Errorf("can't parse time: %w", err)
	}

	return nil
}
