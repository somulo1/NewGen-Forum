package realtime

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Limiter is a fixed-window rate limiter shared by all nodes through Redis.
type Limiter struct{ rdb *redis.Client }

func NewLimiter(rdb *redis.Client) *Limiter { return &Limiter{rdb: rdb} }

// Allow counts one hit for (scope,id) and reports whether it is still within limit per window.
// If Redis is unreachable it fails open for ordinary traffic; callers protecting logins pass strict=true to fail closed.
func (l *Limiter) Allow(ctx context.Context, scope, id string, limit int, window time.Duration, strict bool) (bool, time.Duration) {
	slot := time.Now().Unix() / int64(window.Seconds())
	key := fmt.Sprintf("rl:%s:%s:%d", scope, id, slot)
	pipe := l.rdb.TxPipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, window+time.Second)
	if _, err := pipe.Exec(ctx); err != nil {
		return !strict, 0
	}
	n := incr.Val()
	if n > int64(limit) {
		return false, time.Duration(window.Seconds()-float64(time.Now().Unix()%int64(window.Seconds()))) * time.Second
	}
	return true, 0
}
