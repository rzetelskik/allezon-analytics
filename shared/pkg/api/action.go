package api

import (
	"encoding/json"
	"fmt"
)

type Action int

const (
	VIEW Action = iota + 1
	BUY
)

var actionToString = map[Action]string{
	VIEW: "VIEW",
	BUY:  "BUY",
}

var actionFromString = map[string]Action{
	"VIEW": VIEW,
	"BUY":  BUY,
}

func (a Action) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.String())
}

func (a *Action) UnmarshalJSON(data []byte) error {
	var err error
	var s string

	err = json.Unmarshal(data, &s)
	if err != nil {
		return fmt.Errorf("can't unmarshal to string: %w", err)
	}

	*a, err = ParseAction(s)
	if err != nil {
		return fmt.Errorf("can't parse action: %w", err)
	}

	return nil
}

func ParseAction(s string) (Action, error) {
	a, ok := actionFromString[s]
	if !ok {
		return Action(0), fmt.Errorf("%q is not a valid action", s)
	}

	return a, nil
}

func (a Action) String() string {
	return actionToString[a]
}
