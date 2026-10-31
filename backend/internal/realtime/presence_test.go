package realtime

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestPresenceSurvivesOtherTabClosing(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379"})
	ctx := context.Background()
	if rdb.Ping(ctx).Err() != nil {
		t.Skip("no redis")
	}
	k := "pres:test:" + t.Name()
	defer rdb.Del(ctx, k)
	Touch(ctx, rdb, k, "a")
	Touch(ctx, rdb, k, "b")
	if Leave(ctx, rdb, k, "a") || !Online(ctx, rdb, k) {
		t.Fatal("closing one connection must not mark the person offline")
	}
	if !Leave(ctx, rdb, k, "b") || Online(ctx, rdb, k) {
		t.Fatal("closing the last connection must mark the person offline")
	}
}

func TestPresenceInMemoryRedisAndExpiry(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	ctx := context.Background()
	k := "pres:o:p"
	Touch(ctx, rdb, k, "a")
	Touch(ctx, rdb, k, "b")
	if Leave(ctx, rdb, k, "a") || !Online(ctx, rdb, k) {
		t.Fatal("one connection left must keep the person online")
	}
	Touch(ctx, rdb, k, "b") // a ping keeps it alive
	if !Online(ctx, rdb, k) {
		t.Fatal("pinged connection must be online")
	}
	if !Leave(ctx, rdb, k, "b") {
		t.Fatal("last connection leaving must be offline")
	}
	Touch(ctx, rdb, k, "c")
	if Online(ctx, rdb, "pres:o:nobody") {
		t.Fatal("unknown person must be offline")
	}
	// a connection that stops pinging expires (score is in the past)
	rdb.ZAdd(ctx, k, redis.Z{Score: 1, Member: "dead"})
	rdb.ZRem(ctx, k, "c")
	if Online(ctx, rdb, k) {
		t.Fatal("expired connection must not count")
	}
}
