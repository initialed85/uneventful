package nats_worker

import (
	"fmt"

	"github.com/initialed85/uneventful/internal/helpers"
	"github.com/initialed85/uneventful/pkg/lifecycles"
	"github.com/nats-io/nats.go"
)

type Worker struct {
	lifecycles.Worker
	natsConn *nats.Conn
}

func New(name string) *Worker {
	w := Worker{}

	w.Worker = lifecycles.NewLazyWorker(fmt.Sprintf("nats_%v", name), w.setup, w.teardown)

	return &w
}

func (w *Worker) setup() (err error) {
	w.natsConn, err = helpers.GetNatsConn()
	if err != nil {
		return err
	}

	return nil
}

func (w *Worker) teardown() error {
	w.natsConn.Close()
	w.natsConn = nil

	return nil
}

func (w *Worker) GetNatsConn() (*nats.Conn, error) {
	if !w.IsStarted() {
		return nil, fmt.Errorf("not started")
	}

	return w.natsConn, nil
}
