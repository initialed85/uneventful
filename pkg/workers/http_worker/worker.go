package http_worker

import (
	"fmt"
	"github.com/initialed85/uneventful/internal/helpers"
	"github.com/initialed85/uneventful/pkg/lifecycles"
	"net/http"
	"time"
)

type Worker struct {
	lifecycles.Worker
	serveMux *http.ServeMux
	server   *http.Server
}

func New(name string, port int64, handlerByPattern map[string]http.HandlerFunc) *Worker {
	w := Worker{serveMux: http.NewServeMux()}

	w.server = &http.Server{Addr: fmt.Sprintf(":%v", port), Handler: w.serveMux}

	for pattern, handler := range handlerByPattern {
		w.serveMux.HandleFunc(pattern, handler)
	}

	w.Worker = lifecycles.NewLazyWorker(fmt.Sprintf("http_%v", name), w.setup, w.teardown)

	return &w
}

func (w *Worker) setup() error {
	errors := helpers.GetErrorChannel()

	go func() {
		err := w.server.ListenAndServe()
		if err != nil {
			errors <- err
		}
	}()

	err := helpers.WaitForError(errors, time.Millisecond*100)
	if err != nil {
		return err
	}

	return nil
}

func (w *Worker) teardown() error {
	return nil
}
