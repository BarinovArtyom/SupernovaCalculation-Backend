package redis

import (
	"lab1/internal/app/config"
	"strconv"

	"github.com/go-redis/redis/v8"
)

const servicePrefix = "energycalculation_service." // наш префикс сервиса

type Client struct {
	cfg    config.RedisConfig
	client *redis.Client
}

func New(cfg config.RedisConfig) (*redis.Client, error) {
	client := &Client{}

	client.cfg = cfg

	redisClient := redis.NewClient(&redis.Options{
		Password:    cfg.Password,
		Username:    cfg.User,
		Addr:        cfg.Host + ":" + strconv.Itoa(cfg.Port),
		DB:          0,
		DialTimeout: cfg.DialTimeout,
		ReadTimeout: cfg.ReadTimeout,
	})

	return redisClient, nil
}

func (c *Client) Close() error {
	return c.client.Close()
}
