package wallet

import (
	"fmt"
	"github.com/initialed85/uneventful/pkg/models"
	"github.com/segmentio/ksuid"
)

type Writer struct {
	models.Writer
	name           string
	creditTypeName string
	debitTypeName  string
	wallet         *Wallet
}

func NewWriter(entityID ksuid.KSUID) *Writer {
	name := domainName

	w := Writer{name: name, creditTypeName: fmt.Sprintf("%v.%v.credit", name, entityID.String()), debitTypeName: fmt.Sprintf("%v.%v.debit", name, entityID.String()), wallet: NewWallet(entityID)}

	w.Writer = models.NewWriter(name, entityID)

	_ = w.Writer.AddHandler(credit, func(entityID ksuid.KSUID, requestBody interface{}) (interface{}, error) {
		return w.call(entityID, requestBody, w.wallet.Credit)
	})

	_ = w.Writer.AddHandler(debit, func(entityID ksuid.KSUID, requestBody interface{}) (interface{}, error) {
		return w.call(entityID, requestBody, w.wallet.Debit)
	})

	return &w
}

func (w *Writer) call(entityID ksuid.KSUID, requestBody interface{}, method func(ksuid.KSUID, float64) error) (State, error) {
	amount, err := castRequestBodyToAmount(requestBody)
	if err != nil {
		return State{}, err
	}

	err = method(entityID, amount.Amount)
	if err != nil {
		return State{}, err
	}

	return w.wallet.state, nil
}
