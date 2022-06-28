package models

import (
	"fmt"
	"github.com/segmentio/ksuid"
	"sync"
)

type Handler func(ksuid.KSUID, interface{}) (interface{}, error)

type Handlers interface {
	GetHandler(string) (Handler, error)
	AddHandler(string, Handler) error
	RemoveHandler(string) error
}

type HandlersImplementation struct {
	mu       sync.Mutex
	handlers map[string]Handler
}

func NewHandlers() *HandlersImplementation {
	h := HandlersImplementation{handlers: make(map[string]Handler)}

	return &h
}

func (h *HandlersImplementation) getHandler(endpoint string) (Handler, error) {
	handler, ok := h.handlers[endpoint]
	if !ok {
		return nil, fmt.Errorf("handler for endpoint=%#+v does not exist", endpoint)
	}

	return handler, nil
}

func (h *HandlersImplementation) GetHandler(endpoint string) (Handler, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	return h.getHandler(endpoint)
}

func (h *HandlersImplementation) AddHandler(endpoint string, handler Handler) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	_, err := h.getHandler(endpoint)
	if err == nil {
		return fmt.Errorf("handler for endpoint=%#+v already exists", endpoint)
	}

	h.handlers[endpoint] = handler

	return nil
}

func (h *HandlersImplementation) RemoveHandler(endpoint string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	_, err := h.getHandler(endpoint)
	if err != nil {
		return err
	}

	delete(h.handlers, endpoint)

	return nil
}
