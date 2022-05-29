package wallet

import (
	"encoding/json"
	"fmt"
	"github.com/initialed85/uneventful/pkg/model"
	"github.com/initialed85/uneventful/pkg/model/calls"
	"github.com/initialed85/uneventful/pkg/model/events"
	"github.com/segmentio/ksuid"
)

type Writer struct {
	*model.Writer
	name           string
	creditTypeName string
	debitTypeName  string
	wallet         *Wallet
}

func NewWriter(
	entityID ksuid.KSUID,
) *Writer {
	name := domainName

	w := Writer{
		name:           name,
		creditTypeName: fmt.Sprintf("%v.%v.credit", name, entityID.String()),
		debitTypeName:  fmt.Sprintf("%v.%v.debit", name, entityID.String()),
		wallet:         NewWallet(entityID),
	}

	w.Writer = model.NewWriter(
		name,
		entityID,
		w.eventHandler,
	)

	return &w
}

func (w *Writer) eventHandler(event *events.Event, request *calls.Request) error {
	if event.TypeName != w.creditTypeName && event.TypeName != w.debitTypeName {
		return fmt.Errorf("unknown endpoint in typeName=%#+v (not %#+v or %#+v)", event.TypeName, w.creditTypeName, w.debitTypeName)
	}

	if request.Endpoint != credit && request.Endpoint != debit {
		return fmt.Errorf("unknown request endpoint=%#+v (not %#+v or %#+v)", request.Endpoint, credit, debit)
	}

	amount := Amount{}

	err := json.Unmarshal(request.Data, &amount)
	if err != nil {
		return err
	}

	if request.Endpoint == credit {
		err = w.wallet.Credit(event.SourceID, amount.Amount)
		if err != nil {
			return err
		}

		stateJSON, err := json.Marshal(w.wallet.state)
		if err != nil {
			return err
		}

		return w.SetState(stateJSON)
	} else if request.Endpoint == debit {
		err = w.wallet.Debit(event.SourceID, amount.Amount)
		if err != nil {
			return err
		}

		stateJSON, err := json.Marshal(w.wallet.state)
		if err != nil {
			return err
		}

		return w.SetState(stateJSON)
	}

	return fmt.Errorf("insane state reached")
}
