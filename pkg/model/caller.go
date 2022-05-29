package model

import (
	"fmt"
	"github.com/initialed85/uneventful/pkg/connections/nats_worker"
	"github.com/initialed85/uneventful/pkg/model/calls"
	"github.com/initialed85/uneventful/pkg/model/events"
	"github.com/initialed85/uneventful/pkg/workers"
	"github.com/segmentio/ksuid"
	"log"
	"time"
)

type Caller struct {
	workers.Worker
	natsWorker *nats_worker.Worker
	name       string
	entityID   ksuid.KSUID
}

func NewCaller(
	name string,
	entityID ksuid.KSUID,
) *Caller {
	c := Caller{
		name:     name,
		entityID: entityID,
	}

	workerName := fmt.Sprintf("caller_%v.%v", name, entityID)

	c.Worker = workers.NewLazyWorker(
		workerName,
		c.setup,
		c.teardown,
	)

	c.natsWorker = nats_worker.New(workerName)

	return &c
}

func (c *Caller) setup() error {
	return workers.Setup(c.natsWorker)
}

func (c *Caller) teardown() error {
	return workers.Teardown(c.natsWorker)
}

func (c *Caller) Call(
	name string,
	entityID ksuid.KSUID,
	endpoint string,
	data []byte,
) error {
	natsConn, err := c.natsWorker.GetNatsConn()
	if err != nil {
		return err
	}

	request := &calls.Request{
		Endpoint: endpoint,
		Data:     data,
	}

	requestJSON, err := request.ToJSON()
	if err != nil {
		return err
	}

	address := fmt.Sprintf("%v.%v.%v", name, entityID, endpoint)

	event := events.NewWithoutCorrelation(
		address,
		requestJSON,
	)

	event.SetSource(c.name, c.entityID)

	eventJSON, err := event.ToJSON()
	if err != nil {
		return err
	}

	log.Printf("%v - calling %#+v", c.name, address)

	msg, err := natsConn.Request(fmt.Sprintf("event.%v", address), eventJSON, time.Second*1)
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
