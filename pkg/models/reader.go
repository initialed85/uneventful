package models

import (
	"context"
	"fmt"
	"github.com/initialed85/uneventful/pkg/lifecycles"
	"github.com/initialed85/uneventful/pkg/models/states"
	"github.com/initialed85/uneventful/pkg/workers/redis_worker"
	"github.com/segmentio/ksuid"
)

type Reader interface {
	lifecycles.Worker
	Handlers
	GetState(name string, entityID ksuid.KSUID) (*states.State, error)
}

type ReaderImplementation struct {
	lifecycles.Worker
	Handlers
	redisWorker *redis_worker.Worker
}

func NewReader(name string) *ReaderImplementation {
	name = fmt.Sprintf("reader_%v", name)

	r := ReaderImplementation{Handlers: NewHandlers(), redisWorker: redis_worker.New(name)}

	r.Worker = lifecycles.NewLazyWorker(name, r.setup, r.teardown)

	return &r
}

func (r *ReaderImplementation) setup() (err error) {
	return lifecycles.Setup(r.redisWorker)
}

func (r *ReaderImplementation) teardown() (err error) {
	return lifecycles.Teardown(r.redisWorker)
}

func (r *ReaderImplementation) GetState(name string, entityID ksuid.KSUID) (*states.State, error) {
	redisClient, err := r.redisWorker.GetRedisClient()
	if err != nil {
		return nil, err
	}

	key := fmt.Sprintf("%v.%v", name, entityID.String())

	stringData, err := redisClient.Get(context.Background(), key).Result()
	if err != nil {
		return nil, err
	}

	state, err := states.FromJSON([]byte(stringData))
	if err != nil {
		return nil, err
	}

	return state, nil
}
