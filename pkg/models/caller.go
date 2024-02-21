package models

import (
	"fmt"
	"time"

	"github.com/initialed85/uneventful/pkg/lifecycles"
	"github.com/initialed85/uneventful/pkg/models/calls"
	"github.com/initialed85/uneventful/pkg/models/events"
	"github.com/initialed85/uneventful/pkg/workers/nats_worker"
	"github.com/segmentio/ksuid"
)

type Caller interface {
	lifecycles.Worker
	Handlers
	Call(name string, entityID ksuid.KSUID, endpoint string, data []byte) error
}

type CallerImplementation struct {
	lifecycles.Worker
	Handlers
	natsWorker *nats_worker.Worker
	name       string
	entityID   ksuid.KSUID
}

func NewCaller(name string, entityID ksuid.KSUID) *CallerImplementation {
	c := CallerImplementation{Handlers: NewHandlers(), name: name, entityID: entityID}

	workerName := fmt.Sprintf("caller_%v.%v", name, entityID)

	c.Worker = lifecycles.NewLazyWorker(workerName, c.setup, c.teardown)

	c.natsWorker = nats_worker.New(workerName)

	return &c
}

func (c *CallerImplementation) setup() error {
	return lifecycles.Setup(c.natsWorker)
}

func (c *CallerImplementation) teardown() error {
	return lifecycles.Teardown(c.natsWorker)
}

func (c *CallerImplementation) Call(name string, entityID ksuid.KSUID, endpoint string, data []byte) error {
	natsConn, err := c.natsWorker.GetNatsConn()
	if err != nil {
		return err
	}

	request := &calls.Request{Endpoint: endpoint, Data: data}

	requestJSON, err := request.ToJSON()
	if err != nil {
		return err
	}

	address := fmt.Sprintf("%v.%v.%v", name, entityID, endpoint)

	event := events.NewWithoutCorrelation(address, requestJSON)

	event.SetSource(c.name, c.entityID)

	eventJSON, err := event.ToJSON()
	if err != nil {
		return err
	}

	msg, err := natsConn.Request(fmt.Sprintf("event.%v", address), eventJSON, time.Second*5)
	if err != nil {
		return err
	}

	responseEvent, err := events.FromJSON(msg.Data)
	if err != nil {
		return err
	}

	response, err := calls.ResponseFromJSON(responseEvent.Data)
	if err != nil {
		return err
	}

	if response.Error != "" {
		return fmt.Errorf(response.Error)
	}

	if !response.Success {
		return fmt.Errorf("unknown error")
	}

	return nil
}
