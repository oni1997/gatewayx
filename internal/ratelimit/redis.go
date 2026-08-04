package ratelimit

import (
	"context"
	"fmt"
	"strconv"
	"time"
)

type RedisStore struct {
	client RedisClient
	prefix string
	config Config
}

type RedisClient interface {
	Eval(ctx context.Context, script string, keys []string, args ...any) (any, error)
	Del(ctx context.Context, keys ...string) error
	Ping(ctx context.Context) error
	Close() error
}

func NewRedisStore(client RedisClient, cfg Config) *RedisStore {
	if cfg.RedisPrefix == "" {
		cfg.RedisPrefix = "gatewayx:ratelimit"
	}
	return &RedisStore{
		client: client,
		prefix: cfg.RedisPrefix,
		config: cfg,
	}
}

func (rs *RedisStore) buildKey(key string) string {
	return fmt.Sprintf("%s:%s", rs.prefix, key)
}

func (rs *RedisStore) Allow(key string) bool {
	return rs.AllowN(key, 1)
}

func (rs *RedisStore) AllowN(key string, n int) bool {
	switch rs.config.Strategy {
	case "sliding_window":
		return rs.slidingWindowAllow(key, n)
	default:
		return rs.tokenBucketAllow(key, n)
	}
}

var tokenBucketScript = `
local key = KEYS[1]
local rate = tonumber(ARGV[1])
local burst = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local requested = tonumber(ARGV[4])

local tokens = tonumber(redis.call("HGET", key, "tokens"))
local last_refill = tonumber(redis.call("HGET", key, "last_refill"))

if tokens == nil then
	tokens = burst
	last_refill = now
end

local elapsed = (now - last_refill) / 1000
tokens = math.min(burst, tokens + elapsed * rate)
last_refill = now

if tokens >= requested then
	tokens = tokens - requested
	redis.call("HMSET", key, "tokens", tokens, "last_refill", last_refill)
	redis.call("EXPIRE", key, math.ceil(burst / rate) + 1)
	return 1
end

redis.call("HMSET", key, "tokens", tokens, "last_refill", last_refill)
redis.call("EXPIRE", key, math.ceil(burst / rate) + 1)
return 0
`

var slidingWindowScript = `
local key = KEYS[1]
local rate = tonumber(ARGV[1])
local burst = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local requested = tonumber(ARGV[4])
local window = 1000

local max_allowed = math.max(rate, burst)

redis.call("ZREMRANGEBYSCORE", key, 0, now - window)

local count = redis.call("ZCARD", key)

if count + requested <= max_allowed then
	for i = 1, requested do
		redis.call("ZADD", key, now, now .. ":" .. i .. ":" .. count)
	end
	redis.call("EXPIRE", key, 2)
	return 1
end

return 0
`

func (rs *RedisStore) tokenBucketAllow(key string, n int) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	now := time.Now().UnixMilli()
	result, err := rs.client.Eval(ctx, tokenBucketScript, []string{rs.buildKey(key)},
		strconv.FormatFloat(rs.config.Rate, 'f', -1, 64),
		strconv.Itoa(rs.config.Burst),
		strconv.FormatInt(now, 10),
		strconv.Itoa(n),
	)
	if err != nil {
		return false
	}

	val, ok := result.(int64)
	return ok && val == 1
}

func (rs *RedisStore) slidingWindowAllow(key string, n int) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	now := time.Now().UnixMilli()
	result, err := rs.client.Eval(ctx, slidingWindowScript, []string{rs.buildKey(key)},
		strconv.FormatFloat(rs.config.Rate, 'f', -1, 64),
		strconv.Itoa(rs.config.Burst),
		strconv.FormatInt(now, 10),
		strconv.Itoa(n),
	)
	if err != nil {
		return false
	}

	val, ok := result.(int64)
	return ok && val == 1
}

func (rs *RedisStore) Reset(key string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return rs.client.Del(ctx, rs.buildKey(key))
}

func (rs *RedisStore) ResetAll() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return rs.client.Del(ctx, rs.prefix+":*")
}
