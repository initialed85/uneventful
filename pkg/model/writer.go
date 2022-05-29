package model

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/initialed85/uneventful/pkg/connections/database_worker"
	"github.com/initialed85/uneventful/pkg/connections/nats_worker"
	"github.com/initialed85/uneventful/pkg/connections/redis_worker"
	"github.com/initialed85/uneventful/pkg/model/calls"
	"github.com/initialed85/uneventful/pkg/model/events"
	"github.com/initialed85/uneventful/pkg/model/states"
	"github.com/initialed85/uneventful/pkg/workers"
	"github.com/nats-io/nats.go"
	"github.com/segmentio/ksuid"
	"log"
	"strings"
	"sync"
	"time"
)

type Writer struct {
	workers.Worker
	databaseWorker       *database_worker.Worker
	redisWorker          *redis_worker.Worker
	natsWorker           *nats_worker.Worker
	subject              string
	queue                string
	ignoreResponseNeeded bool
	ignoreEventTypeName  bool
	handleEvents         bool
	subscription         *nats.Subscription
	mu                   sync.Mutex
	name                 string
	entityID             ksuid.KSUID
	eventHandler         func(event *events.Event, request *calls.Request) error
}

func NewWriterWithOverrides(
	name string,
	entityID ksuid.KSUID,
	eventHandler func(event *events.Event, request *calls.Request) error,
	subject string,
	queue string,
	ignoreResponseNeeded bool,
	ignoreEventTypeName bool,
	handleEvents bool,
) *Writer {
	workerName := fmt.Sprintf("writer_%v", name)

	w := Writer{
		databaseWorker:       database_worker.New(workerName),
		redisWorker:          redis_worker.New(workerName),
		natsWorker:           nats_worker.New(workerName),
		subject:              subject,
		queue:                queue,
		ignoreResponseNeeded: ignoreResponseNeeded,
		ignoreEventTypeName:  ignoreEventTypeName,
		handleEvents:         handleEvents,
		name:                 name,
		entityID:             entityID,
		eventHandler:         eventHandler,
	}

	w.Worker = workers.NewLazyWorker(
		workerName,
		w.setup,
		w.teardown,
	)

	return &w
}

func NewWriter(
	name string,
	entityID ksuid.KSUID,
	eventHandler func(event *events.Event, request *calls.Request) error,
) *Writer {
	name = fmt.Sprintf("%v.%v", name, entityID.String())

	return NewWriterWithOverrides(
		name,
		entityID,
		eventHandler,
		fmt.Sprintf("event.%v.*", name),
		name,
		false,
		false,
		true,
	)
}

func (w *Writer) setup() (err error) {
	err = workers.Setup(w.databaseWorker, w.redisWorker, w.natsWorker)
	if err != nil {
		return err
	}

	db, err := w.databaseWorker.GetDB()
	if err != nil {
		_ = workers.Teardown(w.natsWorker, w.redisWorker, w.databaseWorker)
		return err
	}

	err = events.Migrate(db)
	if err != nil {
		_ = workers.Teardown(w.natsWorker, w.redisWorker, w.databaseWorker)
		return err
	}

	err = states.Migrate(db)
	if err != nil {
		_ = workers.Teardown(w.natsWorker, w.redisWorker, w.databaseWorker)
		return err
	}

	natsConn, err := w.natsWorker.GetNatsConn()
	if err != nil {
		_ = workers.Teardown(w.natsWorker, w.redisWorker, w.databaseWorker)
		return err
	}

	log.Printf("%v - subscribing to %#+v", w.name, w.subject)

	if w.queue != "" {
		w.subscription, err = natsConn.QueueSubscribe(
			w.subject,
			w.queue,
			w.handler,
		)
	} else {
		w.subscription, err = natsConn.Subscribe(
			w.subject,
			w.handler,
		)
	}

	if err != nil {
		_ = workers.Teardown(w.natsWorker, w.redisWorker, w.databaseWorker)

		return err
	}

	return
}

func (w *Writer) teardown() (err error) {
	return workers.Teardown(w.natsWorker, w.redisWorker, w.databaseWorker)
}

func (w *Writer) responder(msg *nats.Msg, event *events.Event, err error) {
	response := calls.NewResponseFromError(err)

	responseData, err := response.ToJSON()
	if err != nil {
		log.Printf("%v - warning: %v", w.name, err)
		return
	}

	responseEvent := events.NewWithCorrelation(
		event.EventID,
		fmt.Sprintf("%v_response", event.TypeName),
		responseData,
	)

	responseEvent.SetSource(w.name, w.entityID)

	responseEventData, err := responseEvent.ToJSON()
	if err != nil {
		log.Printf("%v - warning: %v", w.name, err)
		return
	}

	err = msg.Respond(responseEventData)
	if err != nil {
		log.Printf("%v - warning: %v", w.name, err)
		return
	}
}

func (w *Writer) handler(msg *nats.Msg) {
	var err error

	db, err := w.databaseWorker.GetDB()
	if err != nil {
		log.Printf("%v - warning: %v", w.name, err)
		return
	}

	event, err := events.FromJSON(msg.Data)
	if err != nil {
		log.Printf("%v - warning: %v", w.name, err)
		return
	}

	databaseEvent, err := event.ToDatabaseEvent()
	if err != nil {
		log.Printf("%v - warning: %v", w.name, err)
		return
	}

	var request *calls.Request

	responseNeeded := msg.Reply != ""

	if !w.ignoreResponseNeeded && responseNeeded {
		defer func() {
			w.responder(msg, event, err)
		}()
	}

	if !w.ignoreEventTypeName && !strings.HasPrefix(event.TypeName, w.name) {
		err = fmt.Errorf("unknown domain and / or entity ID in typeName=%#+v", event.TypeName)
		log.Printf("%v - warning: %v", w.name, err)
		return
	}

	if !w.ignoreResponseNeeded && responseNeeded {
		request, err = calls.RequestFromJSON(event.Data)
		if err != nil {
			log.Printf("%v - warning: %v", w.name, err)
			return
		}
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	_, err = databaseEvent.Create(db)
	if err != nil {
		log.Printf("%v - warning: %v", w.name, err)
		return
	}

	if !w.handleEvents {
		return
	}

	err = w.eventHandler(event, request)
	if err != nil {
		_, deleteErr := databaseEvent.Delete(db)
		if deleteErr != nil {
			err = fmt.Errorf(
				"event handler caused %v requiring event deletion which caused %v",
				err,
				deleteErr,
			)
			log.Printf("%v - warning: %v", w.name, err)
			return
		}

		log.Printf("%v - warning: %v", w.name, err)
		return
	}

	databaseEvent.IsHandled = true
	databaseEvent.HandledByName = w.name
	databaseEvent.HandledByID = w.entityID.String()

	// if this happens we're goosed- we've already handled the event yet for some reason we can't
	// update that status in the database
	_, err = databaseEvent.Update(db)
	if err != nil {
		log.Printf("%v - warning: %v", w.name, err)
		return
	}
}

func (w *Writer) SetState(
	data json.RawMessage,
) (err error) {
	db, err := w.databaseWorker.GetDB()
	if err != nil {
		return err
	}

	redisClient, err := w.redisWorker.GetRedisClient()
	if err != nil {
		return err
	}

	state := states.New(
		w.name,
		w.entityID,
		data,
	)

	stateJSON, err := state.ToJSON()
	if err != nil {
		return err
	}

	databaseState, err := state.ToDatabaseState()
	if err != nil {
		return err
	}

	_, err = databaseState.Create(db)
	if err != nil {
		return err
	}

	log.Printf("%v - wrote %v to %v", w.name, string(stateJSON), w.name)

	err = redisClient.Set(
		context.Background(),
		w.name,
		stateJSON,
		time.Duration(0),
	).Err()
	if err != nil {
		_, deleteErr := databaseState.Delete(db)
		return fmt.Errorf(
			"redis client caused %v required state deletion which caused %v",
			err,
			deleteErr,
		)
	}

	return nil
}
