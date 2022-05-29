package redis_worker

import (
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/initialed85/uneventful/internal/helpers"
	"github.com/initialed85/uneventful/pkg/workers"
)

type Worker struct {
	workers.Worker
	redisClient *redis.Client
}

func New(
	name string,
) *Worker {
	w := Worker{}

	w.Worker = workers.NewLazyWorker(
		fmt.Sprintf("redis_%v", name),
		w.setup,
		w.teardown,
	)

	return &w
}

func (w *Worker) setup() (err error) {
	w.redisClient, err = helpers.GetRedisClient()
	if err != nil {
		return err
	}

	return nil
}

func (w *Worker) teardown() error {
	_ = w.redisClient.Close()
	w.redisClient = nil

	return nil
}

func (w *Worker) GetRedisClient() (*redis.Client, error) {
	if !w.IsStarted() {
		return nil, fmt.Errorf("not started")
	}

	return w.redisClient, nil
}
