package model

import (
	"context"
	"fmt"
	"github.com/initialed85/uneventful/pkg/connections/redis_worker"
	"github.com/initialed85/uneventful/pkg/model/states"
	"github.com/initialed85/uneventful/pkg/workers"
	"github.com/segmentio/ksuid"
)

type Reader struct {
	workers.Worker
	redisWorker *redis_worker.Worker
}

func NewReader(
	name string,
) *Reader {
	name = fmt.Sprintf("reader_%v", name)

	r := Reader{
		redisWorker: redis_worker.New(name),
	}

	r.Worker = workers.NewLazyWorker(
		name,
		r.setup,
		r.teardown,
	)

	return &r
}

func (r *Reader) setup() (err error) {
	return workers.Setup(r.redisWorker)
}

func (r *Reader) teardown() (err error) {
	return workers.Teardown(r.redisWorker)
}

func (r *Reader) GetState(
	name string,
	entityID ksuid.KSUID,
) (*states.State, error) {
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
