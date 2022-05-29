package wallet

import (
	"github.com/initialed85/uneventful/pkg/model"
	"github.com/segmentio/ksuid"
)

type Reader struct {
	*model.Reader
}

func NewReader(name string) *Reader {
	r := Reader{}

	r.Reader = model.NewReader(name)

	return &r
}

func (r *Reader) GetWalletState(entityID ksuid.KSUID) (*State, error) {
	state, err := r.GetState(domainName, entityID)
	if err != nil {
		return nil, err
	}

	walletState, err := FromJSON(state.Data)
	if err != nil {
		return nil, err
	}

	return walletState, nil
}

func (r *Reader) GetBalance(entityID ksuid.KSUID) (*Balance, error) {
	walletState, err := r.GetWalletState(entityID)
	if err != nil {
		return nil, err
	}

	return &Balance{
		Timestamp: walletState.Timestamp,
		Balance:   walletState.Balance,
	}, nil
}

func (r *Reader) GetTransactions(entityID ksuid.KSUID) (*Transactions, error) {
	walletState, err := r.GetWalletState(entityID)
	if err != nil {
		return nil, err
	}

	return &Transactions{
		Timestamp:    walletState.Timestamp,
		Transactions: walletState.Transactions,
	}, nil
}
