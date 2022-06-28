package wallet

import (
	"encoding/json"
	"github.com/initialed85/uneventful/pkg/models"
	"github.com/segmentio/ksuid"
)

type Caller struct {
	models.Caller
}

func NewCaller(name string, entityID ksuid.KSUID) *Caller {
	c := Caller{}

	c.Caller = models.NewCaller(name, entityID)

	_ = c.Caller.AddHandler("credit", func(entityID ksuid.KSUID, requestBody interface{}) (interface{}, error) {
		return c.call(entityID, requestBody, c.Credit)
	})

	_ = c.Caller.AddHandler("debit", func(entityID ksuid.KSUID, requestBody interface{}) (interface{}, error) {
		return c.call(entityID, requestBody, c.Debit)
	})

	return &c
}

func (c *Caller) call(entityID ksuid.KSUID, requestBody interface{}, method func(ksuid.KSUID, float64) error) (interface{}, error) {
	amount, err := castRequestBodyToAmount(requestBody)
	if err != nil {
		return nil, err
	}

	return nil, method(entityID, amount.Amount)
}

func (c *Caller) Credit(entityID ksuid.KSUID, amount float64) error {
	amountRequest := Amount{Amount: amount}

	data, err := json.Marshal(amountRequest)
	if err != nil {
		return err
	}

	return c.Call(domainName, entityID, credit, data)
}

func (c *Caller) Debit(entityID ksuid.KSUID, amount float64) error {
	amountRequest := Amount{Amount: amount}

	data, err := json.Marshal(amountRequest)
	if err != nil {
		return err
	}

	return c.Call(domainName, entityID, debit, data)
}
