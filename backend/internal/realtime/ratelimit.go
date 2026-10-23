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
	_, ok, retry := l.Take(ctx, scope, id, limit, window, strict)
	return ok, retry
}

// Take is Allow that also says how many hits this window holds, this one included. The count and the decision come from one
// atomic increment, so concurrent callers cannot all see "still room": the first `limit` get in, the rest are refused.
func (l *Limiter) Take(ctx context.Context, scope, id string, limit int, window time.Duration, strict bool) (int64, bool, time.Duration) {
	slot := time.Now().Unix() / int64(window.Seconds())
	key := fmt.Sprintf("rl:%s:%s:%d", scope, id, slot)
	pipe := l.rdb.TxPipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, window+time.Second)
	if _, err := pipe.Exec(ctx); err != nil {
		return 0, !strict, 0
	}
	n := incr.Val()
	if n > int64(limit) {
		return n, false, time.Duration(window.Seconds()-float64(time.Now().Unix()%int64(window.Seconds()))) * time.Second
	}
	return n, true, 0
}

// Refund gives one hit back (a sign-in that worked should not count against the address).
func (l *Limiter) Refund(ctx context.Context, scope, id string, window time.Duration) {
	slot := time.Now().Unix() / int64(window.Seconds())
	key := fmt.Sprintf("rl:%s:%s:%d", scope, id, slot)
	if n, err := l.rdb.Decr(ctx, key).Result(); err == nil && n < 0 {
		l.rdb.Set(ctx, key, 0, window+time.Second)
	}
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

// Cap is one limit for SlidingAll: a counter name, who it counts, how many hits are allowed in the window.
type Cap struct {
	Scope, ID string
	Limit     int
	Window    time.Duration
}

// SlidingAll counts a hit against every cap in a single round trip to Redis (the caps used to be asked one after another, and
// each ask costs a trip to Redis). It says which cap, if any, was passed first, and how long to wait.
func (l *Limiter) SlidingAll(ctx context.Context, caps []Cap) (ok bool, over int, retry time.Duration) {
	if len(caps) == 0 {
		return true, -1, 0
	}
	now := time.Now()
	pipe := l.rdb.Pipeline()
	incr := make([]*redis.IntCmd, len(caps))
	prev := make([]*redis.StringCmd, len(caps))
	for i, c := range caps {
		w := int64(c.Window.Seconds())
		slot := now.Unix() / w
		cur := fmt.Sprintf("rl:%s:%s:%d", c.Scope, c.ID, slot)
		incr[i] = pipe.Incr(ctx, cur)
		pipe.Expire(ctx, cur, 2*c.Window+time.Second)
		prev[i] = pipe.Get(ctx, fmt.Sprintf("rl:%s:%s:%d", c.Scope, c.ID, slot-1))
	}
	if _, err := pipe.Exec(ctx); err != nil && err != redis.Nil {
		return true, -1, 0 // ordinary traffic fails open when Redis is unreachable
	}
	for i, c := range caps {
		w := int64(c.Window.Seconds())
		p, _ := prev[i].Int64()
		frac := float64(now.Unix()%w) / float64(w)
		used := float64(incr[i].Val()) + float64(p)*(1-frac)
		if used > float64(c.Limit) {
			return false, i, time.Duration(float64(w)*(1-frac)+1) * time.Second
		}
	}
	return true, -1, 0
}
