package services

import (
	"context"

	"github.com/go-redis/redis/v8"
)

type CounterService struct {
	redisClient *redis.Client
	ctx         context.Context
}

func NewCounterService(redisClient *redis.Client) *CounterService {
	return &CounterService{
		redisClient: redisClient,
		ctx:         context.Background(),
	}
}

func (cs *CounterService) Increment() (int64, error) {
	val, err := cs.redisClient.Incr(cs.ctx, "counter").Result()
	if err != nil {
		return 0, err
	}
	return val, nil
}

func (cs *CounterService) Decrement() (int64, error) {
	val, err := cs.redisClient.Decr(cs.ctx, "counter").Result()
	if err != nil {
		return 0, err
	}
	return val, nil
}

func (cs *CounterService) GetCurrentStatus() (int64, error) {
	val, err := cs.redisClient.Get(cs.ctx, "counter").Int64()
	if err != nil {
		if err == redis.Nil {
			return 0, nil // Counter not set, return 0
		}
		return 0, err
	}
	return val, nil
}
