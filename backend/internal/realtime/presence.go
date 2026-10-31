package realtime

import (
	"context"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// A person is online while at least one of their connections (tabs, devices) is alive. Each connection is a member of a sorted set
// scored by the time it expires; pings push that time forward and closing a connection removes only its own member, so closing one
// tab or a quick reconnect never marks someone offline while another connection of theirs is still open. The time allowed between
// pings is generous because browsers slow timers in background tabs.
const PresenceTTL = 150 * time.Second

func PresenceKey(org uuid.UUID, person string) string { return "pres:" + org.String() + ":" + person }

// Touch marks the connection alive for another PresenceTTL.
func Touch(ctx context.Context, rdb *redis.Client, key, conn string) {
	now := time.Now()
	p := rdb.Pipeline()
	p.ZAdd(ctx, key, redis.Z{Score: float64(now.Add(PresenceTTL).UnixMilli()), Member: conn})
	p.ZRemRangeByScore(ctx, key, "-inf", itoa(now.UnixMilli()))
	p.Expire(ctx, key, 2*PresenceTTL)
	_, _ = p.Exec(ctx)
}

// Leave removes one connection and says whether the person has no other live connection.
func Leave(ctx context.Context, rdb *redis.Client, key, conn string) (offline bool) {
	rdb.ZRem(ctx, key, conn)
	return !Online(ctx, rdb, key)
}

// Online reports whether any connection of the person is alive.
func Online(ctx context.Context, rdb *redis.Client, key string) bool {
	n, err := rdb.ZCount(ctx, key, itoa(time.Now().UnixMilli()), "+inf").Result()
	return err == nil && n > 0
}

func itoa(n int64) string { return strconv.FormatInt(n, 10) }
