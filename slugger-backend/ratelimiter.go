package main

import (
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type RateLimiter struct {
	rdb *redis.Client
}

func NewRateLimiter(rdb *redis.Client) *RateLimiter {
	return &RateLimiter{rdb: rdb}
}

// func (rl *RateLimiter) IsAllowed(ip string, limit int, window time.Duration) (bool, error) {
// 	key := fmt.Sprintf("rate_limit:%s", ip)

// 	pipe := rl.rdb.Pipeline()
// 	pipe.Incr(ctx, key)
// 	pipe.Expire(ctx, key, window)

// 	cmds, err := pipe.Exec(ctx)
// 	if err != nil {
// 		return false, err
// 	}

// 	val := cmds[0].(*redis.IntCmd).Val()
// 	return val <= int64(limit), nil
// }

func (rl *RateLimiter) IsAllowed(ip string, limit int, window time.Duration) (bool, error) {
	key := fmt.Sprintf("rate_limit:%s", ip)

	pipe := rl.rdb.Pipeline()
	pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, window)

	cmds, err := pipe.Exec(ctx)
	if err != nil {
		return false, err
	}

	val := cmds[0].(*redis.IntCmd).Val()
	return val <= int64(limit), nil
}