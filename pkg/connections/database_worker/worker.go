package database_worker

import (
	"fmt"
	"github.com/initialed85/uneventful/internal/helpers"
	"github.com/initialed85/uneventful/pkg/workers"
	"gorm.io/gorm"
)

type Worker struct {
	workers.Worker
	db *gorm.DB
}

func New(
	name string,
) *Worker {
	w := Worker{}

	w.Worker = workers.NewLazyWorker(
		fmt.Sprintf("database_%v", name),
		w.setup,
		w.teardown,
	)

	return &w
}

func (w *Worker) setup() (err error) {
	w.db, err = helpers.GetDatabase()
	if err != nil {
		return err
	}

	return nil
}

func (w *Worker) teardown() error {
	w.db = nil

	return nil
}

func (w *Worker) GetDB() (*gorm.DB, error) {
	if !w.IsStarted() {
		return nil, fmt.Errorf("not started")
	}

	return w.db, nil
}
