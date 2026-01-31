package redis

import (
	"context"
	"encoding/json"
	"time"

	"backend/pkgs/logger"

	goredis "github.com/redis/go-redis/v9"
)

const Timeout = time.Second * 2

type IRedis interface {
	IsConnected() bool
	Get(key string, value interface{}) error
	Set(key string, value interface{}) error
	SetWithExpiration(key string, value interface{}, expiration time.Duration) error
	Remove(keys ...string) error
	Keys(pattern string) ([]string, error)
	RemovePattern(pattern string) error
}

type Config struct {
	Address  string
	Password string
	Database int
}

type redis struct {
	cmd goredis.Cmdable
}

func NewRedis(config Config) IRedis {
	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()

	rdb := goredis.NewClient(&goredis.Options{
		Addr:     config.Address,
		Password: config.Password,
		DB:       config.Database,
	})

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		logger.Fatal("Redis connection failed:", err)
		return nil
	}

	logger.Info("Redis connection established")

	return &redis{cmd: rdb}
}

func (r *redis) IsConnected() bool {
	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()

	if r.cmd == nil {
		return false
	}

	_, err := r.cmd.Ping(ctx).Result()
	return err == nil
}

func (r *redis) Get(key string, value interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()

	strValue, err := r.cmd.Get(ctx, key).Result()
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(strValue), value)
}

func (r *redis) Set(key string, value interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()

	bData, _ := json.Marshal(value)
	return r.cmd.Set(ctx, key, bData, 0).Err()
}

func (r *redis) SetWithExpiration(key string, value interface{}, expiration time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()

	bData, _ := json.Marshal(value)
	return r.cmd.Set(ctx, key, bData, expiration).Err()
}

func (r *redis) Remove(keys ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()

	return r.cmd.Del(ctx, keys...).Err()
}

func (r *redis) Keys(pattern string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()

	return r.cmd.Keys(ctx, pattern).Result()
}

func (r *redis) RemovePattern(pattern string) error {
	keys, err := r.Keys(pattern)
	if err != nil {
		return err
	}
	if len(keys) == 0 {
		return nil
	}
	return r.Remove(keys...)
}
