package wallet

import (
	"fmt"
	"github.com/segmentio/ksuid"
	"sync"
	"time"
)

type Wallet struct {
	mu       sync.Mutex
	state    State
	entityID ksuid.KSUID
}

func NewWallet(entityID ksuid.KSUID) *Wallet {
	w := Wallet{entityID: entityID, state: State{Timestamp: time.Now(), Balance: 0, Transactions: make([]Transaction, 0)}}

	return &w
}

func (w *Wallet) applyTransaction(transaction Transaction) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	proposedBalance := w.state.Balance + transaction.Amount

	if w.state.Balance+transaction.Amount < 0 {
		return fmt.Errorf("%#+v rejected; would cause balance of %#+v (overdrawn)", transaction, proposedBalance)
	}

	w.state.Balance = proposedBalance
	w.state.Transactions = append(w.state.Transactions, transaction)

	return nil
}

func (w *Wallet) Credit(sourceEntityID ksuid.KSUID, amount float64) error {
	if amount <= 0 {
		return fmt.Errorf("credit amount must be greater than 0")
	}

	return w.applyTransaction(NewTransaction(sourceEntityID, amount))
}

func (w *Wallet) Debit(sourceEntityID ksuid.KSUID, amount float64) error {
	if amount <= 0 {
		return fmt.Errorf("debit amount must be greater than 0")
	}

	return w.applyTransaction(NewTransaction(sourceEntityID, -amount))
}
