package dao

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
)

type RedisClient struct {
	Addr     string
	Password string
	DB       int
	Client   *redis.Client
}

var Global_redis *RedisClient

func RedisInit(url, pass string, db int, dialTimeout time.Duration) error {
	r := &RedisClient{
		Addr:     url,
		Password: pass,
		DB:       db,
	}

	r.Client = redis.NewClient(&redis.Options{
		Addr:     r.Addr,
		Password: r.Password,
		DB:       r.DB,
	})

	// 使用超时上下文，验证redis
	timeoutCtx, cancelFunc := context.WithTimeout(context.Background(), dialTimeout)
	defer cancelFunc()
	_, err := r.Client.Ping(timeoutCtx).Result()
	if err != nil {
		err = errors.Errorf("redis connection timeout: %s", err.Error())
		return err
	}

	return nil
}

func (r *RedisClient) Set(key string, value interface{}) error {
	var ctx = context.Background()

	bytes, _ := json.Marshal(value)
	err := r.Client.Set(ctx, key, string(bytes), 0).Err()
	if err != nil {
		err = errors.Errorf("failed to set key-value: %s", err.Error())
		return err
	}

	return nil
}

func (r *RedisClient) Get(key string, obj interface{}) (interface{}, error) {
	var ctx = context.Background()

	data, err := r.Client.Get(ctx, key).Result()
	if err != nil {
		err = errors.Errorf("failed to get value: %s", err.Error())
		return nil, err
	}
	json.Unmarshal([]byte(data), obj)
	return obj, nil
}
func (r *RedisClient) Scan(key string) []string {
	var ctx = context.Background()
	keys := []string{}

	iterator := r.Client.Scan(ctx, 0, key, 0).Iterator()
	for iterator.Next(ctx) {
		key := iterator.Val()
		keys = append(keys, key)
	}

	return keys
}

func (r *RedisClient) Delete(key string) error {
	var ctx = context.Background()

	err := r.Client.Del(ctx, key).Err()
	if err != nil {
		err = errors.Errorf("failed to del key-value: %s", err.Error())
		return err
	}
	return nil
}