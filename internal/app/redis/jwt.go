package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

const jwtPrefix = "jwt."

func getJWTKey(token string) string {
	return servicePrefix + jwtPrefix + token
}

func WriteJWTToBlacklist(c *redis.Client, ctx context.Context, jwtStr string, jwtTTL time.Duration) error {
	return c.Set(ctx, getJWTKey(jwtStr), true, jwtTTL).Err()
}

func CheckJWTInBlacklist(c *redis.Client, ctx context.Context, jwtStr string) error {
	return c.Get(ctx, getJWTKey(jwtStr)).Err()
	// если токена нет, то вернется ошибка not exists
}
