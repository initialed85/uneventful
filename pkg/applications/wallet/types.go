package wallet

import (
	"encoding/json"
	"github.com/initialed85/uneventful/internal/helpers"
	"github.com/segmentio/ksuid"
	"time"
)

type Transaction struct {
	Timestamp      time.Time   `json:"timestamp"`
	SourceEntityID ksuid.KSUID `json:"source_entity_id"`
	Amount         float64     `json:"amount"`
}

func NewTransaction(sourceEntityID ksuid.KSUID, amount float64) Transaction {
	t := Transaction{Timestamp: helpers.GetNow(), SourceEntityID: sourceEntityID, Amount: amount}

	return t
}

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

type Amount struct {
	Amount float64 `json:"amount"`
}

type Balance struct {
	Timestamp time.Time `json:"timestamp"`
	Balance   float64   `json:"balance"`
}

type Transactions struct {
	Timestamp    time.Time     `json:"timestamp"`
	Transactions []Transaction `json:"transactions"`
}
