package wallet

import (
	"github.com/initialed85/uneventful/internal/helpers"
	"github.com/segmentio/ksuid"
	"time"
)

type Transaction struct {
	Timestamp      time.Time   `json:"timestamp"`
	SourceEntityID ksuid.KSUID `json:"source_entity_id"`
	Amount         float64     `json:"amount"`
}

func NewTransaction(
	sourceEntityID ksuid.KSUID,
	amount float64,
) Transaction {
	t := Transaction{
		Timestamp:      helpers.GetNow(),
		SourceEntityID: sourceEntityID,
		Amount:         amount,
	}

	return t
}
