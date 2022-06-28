package models

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/initialed85/uneventful/pkg/lifecycles"
	"github.com/initialed85/uneventful/pkg/models/calls"
	"github.com/initialed85/uneventful/pkg/models/events"
	"github.com/initialed85/uneventful/pkg/models/states"
	"github.com/initialed85/uneventful/pkg/workers/database_worker"
	"github.com/initialed85/uneventful/pkg/workers/nats_worker"
	"github.com/initialed85/uneventful/pkg/workers/redis_worker"
	"github.com/nats-io/nats.go"
	"github.com/segmentio/ksuid"
	"gorm.io/gorm"
	"log"
	"strings"
	"sync"
	"time"
)

type Writer interface {
	lifecycles.Worker
	Handlers
	SetState(data json.RawMessage) (err error)
}

type WriterImplementation struct {
	lifecycles.Worker
	Handlers
	databaseWorker       *database_worker.Worker
	redisWorker          *redis_worker.Worker
	natsWorker           *nats_worker.Worker
	subject              string
	queue                string
	ignoreResponseNeeded bool
	ignoreEventTypeName  bool
	handleEvents         bool
	subscription         *nats.Subscription
	mu, dbMu             sync.Mutex
	name                 string
	entityID             ksuid.KSUID
}

func NewWriterWithOverrides(name string, entityID ksuid.KSUID, subject string, queue string, ignoreResponseNeeded bool, ignoreEventTypeName bool, handleEvents bool) *WriterImplementation {
	workerName := fmt.Sprintf("writer_%v", name)

	w := WriterImplementation{Handlers: NewHandlers(), databaseWorker: database_worker.New(workerName), redisWorker: redis_worker.New(workerName), natsWorker: nats_worker.New(workerName), subject: subject, queue: queue, ignoreResponseNeeded: ignoreResponseNeeded, ignoreEventTypeName: ignoreEventTypeName, handleEvents: handleEvents, name: name, entityID: entityID}

	w.Worker = lifecycles.NewLazyWorker(workerName, w.setup, w.teardown)

	return &w
}

func NewWriter(name string, entityID ksuid.KSUID) *WriterImplementation {
	name = fmt.Sprintf("%v.%v", name, entityID.String())

	return NewWriterWithOverrides(name, entityID, fmt.Sprintf("event.%v.*", name), name, false, false, true)
}

func (w *WriterImplementation) setup() (err error) {
	err = lifecycles.Setup(w.databaseWorker, w.redisWorker, w.natsWorker)
	if err != nil {
		return err
	}

	db, err := w.databaseWorker.GetDB()
	if err != nil {
		_ = lifecycles.Teardown(w.natsWorker, w.redisWorker, w.databaseWorker)
		return err
	}

	err = events.Migrate(db)
	if err != nil {
		_ = lifecycles.Teardown(w.natsWorker, w.redisWorker, w.databaseWorker)
		return err
	}

	err = states.Migrate(db)
	if err != nil {
		_ = lifecycles.Teardown(w.natsWorker, w.redisWorker, w.databaseWorker)
		return err
	}

	natsConn, err := w.natsWorker.GetNatsConn()
	if err != nil {
		_ = lifecycles.Teardown(w.natsWorker, w.redisWorker, w.databaseWorker)
		return err
	}

	if w.handleEvents {
		databaseEvents, err := events.GetAll(db)
		if err != nil {
			_ = lifecycles.Teardown(w.natsWorker, w.redisWorker, w.databaseWorker)
			return err
		}

		err = w.handleRequestfromDatabasEvents(db, databaseEvents)
		if err != nil {
			_ = lifecycles.Teardown(w.natsWorker, w.redisWorker, w.databaseWorker)
			return err
		}
	}

	log.Printf("%v - subscribing to %#+v", w.name, w.subject)

	if w.queue != "" {
		w.subscription, err = natsConn.QueueSubscribe(w.subject, w.queue, w.handler)
	} else {
		w.subscription, err = natsConn.Subscribe(w.subject, w.handler)
	}

	if err != nil {
		_ = lifecycles.Teardown(w.natsWorker, w.redisWorker, w.databaseWorker)

		return err
	}

	return
}

func (w *WriterImplementation) teardown() (err error) {
	return lifecycles.Teardown(w.natsWorker, w.redisWorker, w.databaseWorker)
}

func (w *WriterImplementation) responder(msg *nats.Msg, event *events.Event, err error) {
	response := calls.NewResponseFromError(err)

	responseData, err := response.ToJSON()
	if err != nil {
		log.Printf("%v - warning: %v", w.name, err)
		return
	}

	responseEvent := events.NewWithCorrelation(event.EventID, fmt.Sprintf("%v_response", event.TypeName), responseData)

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

func (w *WriterImplementation) handleRequestfromDatabasEvents(db *gorm.DB, databaseEvents []*events.DatabaseEvent) error {
	if !w.handleEvents {
		return nil
	}

	log.Printf("%v - replaying %v events to achieve state", w.name, len(databaseEvents))

	var err error
	var request *calls.Request
	var requestData interface{}
	var handler Handler
	var state interface{}
	var stateJSON []byte

	for _, databaseEvent := range databaseEvents {
		request, err = calls.RequestFromJSON(databaseEvent.Data.Bytes)
		if err != nil {
			return err
		}

		err = json.Unmarshal(request.Data, &requestData)
		if err != nil {
			return err
		}

		handler, err = w.GetHandler(request.Endpoint)
		if err != nil {
			return err
		}

		sourceEntityID, err := ksuid.Parse(databaseEvent.SourceID)
		if err != nil {
			return err
		}

		state, err = handler(sourceEntityID, requestData)
		if err != nil {
			return err
		}
	}

	if len(databaseEvents) == 0 {
		return nil
	}

	// TODO: let's hope this never happens, because we've already handled all the events
	stateJSON, err = json.Marshal(state)
	if err != nil {
		return err
	}

	err = w.SetState(stateJSON)
	if err != nil {
		return err
	}

	return nil
}

func (w *WriterImplementation) handler(msg *nats.Msg) {
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

	var request *calls.Request

	if !w.ignoreResponseNeeded && responseNeeded {
		request, err = calls.RequestFromJSON(event.Data)
		if err != nil {
			log.Printf("%v - warning: %v", w.name, err)
			return
		}
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	w.dbMu.Lock()
	_, err = databaseEvent.Create(db)
	w.dbMu.Unlock()
	if err != nil {
		log.Printf("%v - warning: %v", w.name, err)
		return
	}

	if !w.handleEvents {
		return
	}

	var requestData interface{}
	var handler Handler
	var state interface{}
	var stateJSON []byte

	err = json.Unmarshal(request.Data, &requestData)

	if err == nil {
		handler, err = w.GetHandler(request.Endpoint)
	}

	if err == nil {
		state, err = handler(event.SourceID, requestData)
	}

	// TODO: let's hope this never happens, because we've already handled the event
	if err == nil {
		stateJSON, err = json.Marshal(state)
	}

	if err == nil {
		err = w.SetState(stateJSON)
	}

	if err != nil {
		w.dbMu.Lock()
		_, deleteErr := databaseEvent.Delete(db)
		w.dbMu.Unlock()
		if deleteErr != nil {
			err = fmt.Errorf("event handler caused %v requiring event deletion which caused %v", err, deleteErr)
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
	w.dbMu.Lock()
	_, err = databaseEvent.Update(db)
	w.dbMu.Unlock()
	if err != nil {
		log.Printf("%v - warning: %v", w.name, err)
		return
	}
}

func (w *WriterImplementation) SetState(data json.RawMessage) (err error) {
	db, err := w.databaseWorker.GetDB()
	if err != nil {
		return err
	}

	redisClient, err := w.redisWorker.GetRedisClient()
	if err != nil {
		return err
	}

	state := states.New(w.name, w.entityID, data)

	stateJSON, err := state.ToJSON()
	if err != nil {
		return err
	}

	databaseState, err := state.ToDatabaseState()
	if err != nil {
		return err
	}

	w.dbMu.Lock()
	_, err = databaseState.Create(db)
	w.dbMu.Unlock()
	if err != nil {
		return err
	}

	log.Printf("%v - wrote %v to %v", w.name, string(stateJSON), w.name)

	err = redisClient.Set(context.Background(), w.name, stateJSON, time.Duration(0)).Err()
	if err != nil {
		w.dbMu.Lock()
		_, deleteErr := databaseState.Delete(db)
		w.dbMu.Unlock()
		return fmt.Errorf("redis client caused %v required state deletion which caused %v", err, deleteErr)
	}

	return nil
}
