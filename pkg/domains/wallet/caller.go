package wallet

import (
	"encoding/json"
	"github.com/initialed85/uneventful/pkg/model"
	"github.com/segmentio/ksuid"
)

type Caller struct {
	*model.Caller
}

func NewCaller(
	name string,
	entityID ksuid.KSUID,
) *Caller {
	c := Caller{}

	c.Caller = model.NewCaller(name, entityID)

	return &c
}

func (c *Caller) Credit(entityID ksuid.KSUID, amount float64) error {
	amountRequest := Amount{
		Amount: amount,
	}

	data, err := json.Marshal(amountRequest)
	if err != nil {
		return err
	}

	return c.Call(
		domainName,
		entityID,
		credit,
		data,
	)
}

func (c *Caller) Debit(entityID ksuid.KSUID, amount float64) error {
	amountRequest := Amount{
		Amount: amount,
	}

	data, err := json.Marshal(amountRequest)
	if err != nil {
		return err
	}

	return c.Call(
		domainName,
		entityID,
		debit,
		data,
	)
}
