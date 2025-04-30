package handlers

import (
	"context"
	"log"
	"net/http"
	"strconv"

	"github.com/go-redis/redis/v8"
)

type CounterService struct {
	redisClient *redis.Client
}

func NewCounterService(redisClient *redis.Client) *CounterService {
	return &CounterService{redisClient: redisClient}
}

func (cs *CounterService) Increment(ctx context.Context) (int, error) {
	count, err := cs.redisClient.Incr(ctx, "counter").Result()
	if err != nil {
		log.Printf("Error incrementing counter: %v", err)
		return 0, err
	}
	return int(count), nil
}

func (cs *CounterService) Decrement(ctx context.Context) (int, error) {
	count, err := cs.redisClient.Decr(ctx, "counter").Result()
	if err != nil {
		log.Printf("Error decrementing counter: %v", err)
		return 0, err
	}
	return int(count), nil
}

func (cs *CounterService) GetCurrentCount(ctx context.Context) (int, error) {
	count, err := cs.redisClient.Get(ctx, "counter").Result()
	if err != nil {
		if err == redis.Nil {
			// Key doesn't exist yet, not an error
			log.Println("Counter key not found in Redis, initializing to 0")
			return 0, nil
		}
		log.Printf("Error retrieving counter from Redis: %v", err)
		return 0, err
	}

	intValue, err := strconv.Atoi(count)
	if err != nil {
		log.Printf("Error converting counter value '%s' to integer: %v", count, err)
		return 0, err
	}

	return intValue, nil
}

func HelloHandler(counterService *CounterService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log.Printf("Handling hello request from %s", r.RemoteAddr)

		currentCount, err := counterService.GetCurrentCount(ctx)
		if err != nil {
			log.Printf("HelloHandler error: %v", err)
			http.Error(w, "Error retrieving counter", http.StatusInternalServerError)
			return
		}

		message := "Hello, World! Current counter value: " + strconv.Itoa(currentCount)
		w.Write([]byte(message))
		log.Printf("Hello request served successfully, counter: %d", currentCount)
	}
}
