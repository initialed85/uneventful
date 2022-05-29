package wallet

import (
	"encoding/json"
	"time"
)

type State struct {
	Timestamp    time.Time     `json:"timestamp"`
	Balance      float64       `json:"balance"`
	Transactions []Transaction `json:"transactions"`
}

func FromJSON(data []byte) (*State, error) {
	state := State{}

	err := json.Unmarshal(data, &state)
	if err != nil {
		return nil, err
	}

	return &state, nil
}

func (s *State) ToJSON() ([]byte, error) {
	data, err := json.Marshal(s)
	if err != nil {
		return []byte{}, err
	}

	return data, nil
}
