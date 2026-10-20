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

// Sliding counts a hit like Allow, but weighs the previous window in, so a burst split across a window boundary cannot get
// twice the limit. It also reports how many hits are left.
func (l *Limiter) Sliding(ctx context.Context, scope, id string, limit int, window time.Duration, strict bool) (ok bool, retry time.Duration, remaining int) {
	w := int64(window.Seconds())
	now := time.Now()
	slot := now.Unix() / w
	cur := fmt.Sprintf("rl:%s:%s:%d", scope, id, slot)
	prev := fmt.Sprintf("rl:%s:%s:%d", scope, id, slot-1)
	pipe := l.rdb.TxPipeline()
	incr := pipe.Incr(ctx, cur)
	pipe.Expire(ctx, cur, 2*window+time.Second)
	pv := pipe.Get(ctx, prev)
	if _, err := pipe.Exec(ctx); err != nil && err != redis.Nil {
		return !strict, 0, limit
	}
	p, _ := pv.Int64()
	frac := float64(now.Unix()%w) / float64(w)
	used := float64(incr.Val()) + float64(p)*(1-frac)
	left := int(float64(limit) - used)
	if used > float64(limit) {
		return false, time.Duration(float64(w)*(1-frac)+1) * time.Second, 0
	}
	return true, 0, left
}
