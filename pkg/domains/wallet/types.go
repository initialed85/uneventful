package wallet

import "time"

type Balance struct {
	Timestamp time.Time `json:"timestamp"`
	Balance   float64   `json:"balance"`
}

type Transactions struct {
	Timestamp    time.Time     `json:"timestamp"`
	Transactions []Transaction `json:"transactions"`
}

type Amount struct {
	Amount float64 `json:"amount"`
}
