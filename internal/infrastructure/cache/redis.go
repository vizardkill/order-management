package cache

import (
	"encoding/json"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"golang.org/x/net/context"
)

type RedisClient struct {
	Client *redis.Client
}

func NewRedisClient(host, port, password string, db int) *RedisClient {
	rdb := redis.NewClient(&redis.Options{
		Addr:     host + ":" + port,
		Password: password,
		DB:       db,
	})
	log.Println(`Redis corriendo en ` + host + `:` + port)
	return &RedisClient{Client: rdb}
}

type IdempotencyData struct {
	Status   string `json:"status"`   // IN_PROGRESS o COMPLETED
	Response string `json:"response"` // Respuesta generada
}

func (r *RedisClient) GetIdempotencyKey(key string) (*IdempotencyData, error) {
	ctx := context.Background()
	data, err := r.Client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var idempotencyData IdempotencyData
	if err := json.Unmarshal([]byte(data), &idempotencyData); err != nil {
		return nil, err
	}

	return &idempotencyData, nil
}

func (r *RedisClient) SetIdempotencyKey(key string, data IdempotencyData, expiration time.Duration) error {
	ctx := context.Background()
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return r.Client.Set(ctx, key, jsonData, expiration).Err()
}

func (r *RedisClient) DeleteIdempotencyKey(key string) error {
    ctx := context.Background()
    return r.Client.Del(ctx, key).Err()
}

func (r *RedisClient) AcquireLock(key string, expiration time.Duration) (bool, error) {
    ctx := context.Background()
    result, err := r.Client.SetNX(ctx, key, "locked", expiration).Result()
    return result, err
}

func (r *RedisClient) ReleaseLock(key string) error {
    ctx := context.Background()
    return r.Client.Del(ctx, key).Err()
}