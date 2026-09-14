package ratelimiter

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

var (
	ErrRateLimitExceeded = errors.New("rate limit exceeded")
)

type RateLimitTier string

const (
	TierAPIKey       RateLimitTier = "api_key"
	TierProject      RateLimitTier = "project"
	TierOrganization RateLimitTier = "organization"
)

type RateLimitRequest struct {
	Tier   RateLimitTier
	Key    string
	Limit  int
	Window time.Duration
	Cost   int
}

type RateLimitResult struct {
	Allowed    bool
	Limit      int
	Remaining  int
	ResetTime  time.Time
	RetryAfter time.Duration
}

type RateLimiter interface {
	CheckRateLimit(ctx context.Context, req RateLimitRequest) (*RateLimitResult, error)
}

// Sliding window Lua script executed atomically in Redis
const slidingWindowLuaScript = `
local key = KEYS[1]
local now = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local limit = tonumber(ARGV[3])
local cost = tonumber(ARGV[4])
local member_id = ARGV[5]

local clear_before = now - window
redis.call('ZREMRANGEBYSCORE', key, '-inf', '(' .. clear_before)

local current_count = redis.call('ZCARD', key)

if (current_count + cost) <= limit then
    for i = 1, cost do
        local seq_member = member_id .. ':' .. i
        redis.call('ZADD', key, now, seq_member)
    end
    local ttl_seconds = math.ceil(window / 1000)
    redis.call('EXPIRE', key, ttl_seconds)
    local remaining = limit - (current_count + cost)
    return {1, limit, remaining, now + window}
else
    local remaining = 0
    if current_count < limit then
        remaining = limit - current_count
    end
    return {0, limit, remaining, now + window}
end
`

type RedisRateLimiter struct {
	client *redis.Client
	script *redis.Script
}

func NewRedisRateLimiter(client *redis.Client) *RedisRateLimiter {
	return &RedisRateLimiter{
		client: client,
		script: redis.NewScript(slidingWindowLuaScript),
	}
}

func (r *RedisRateLimiter) CheckRateLimit(ctx context.Context, req RateLimitRequest) (*RateLimitResult, error) {
	if r.client == nil {
		return nil, fmt.Errorf("redis client is uninitialized")
	}

	if req.Limit <= 0 {
		req.Limit = 600
	}
	if req.Window <= 0 {
		req.Window = time.Minute
	}
	if req.Cost <= 0 {
		req.Cost = 1
	}

	nowMs := time.Now().UTC().UnixMilli()
	windowMs := req.Window.Milliseconds()
	memberID := fmt.Sprintf("%d:%s", nowMs, uuid.New().String()[:8])

	keys := []string{req.Key}
	args := []interface{}{nowMs, windowMs, req.Limit, req.Cost, memberID}

	res, err := r.script.Run(ctx, r.client, keys, args...).Slice()
	if err != nil {
		return nil, fmt.Errorf("failed to execute rate limit script: %w", err)
	}

	if len(res) < 4 {
		return nil, fmt.Errorf("invalid rate limit script response length: %d", len(res))
	}

	allowedInt, _ := res[0].(int64)
	limitInt, _ := res[1].(int64)
	remainingInt, _ := res[2].(int64)
	resetMs, _ := res[3].(int64)

	allowed := allowedInt == 1
	resetTime := time.UnixMilli(resetMs)
	retryAfter := time.Duration(0)

	if !allowed {
		retryAfter = time.Until(resetTime)
		if retryAfter < time.Second {
			retryAfter = time.Second
		}
	}

	return &RateLimitResult{
		Allowed:    allowed,
		Limit:      int(limitInt),
		Remaining:  int(remainingInt),
		ResetTime:  resetTime,
		RetryAfter: retryAfter,
	}, nil
}

// MemoryRateLimiter provides an in-memory sliding window rate limiter for testing or fallback
type MemoryRateLimiter struct {
	mu     sync.Mutex
	stores map[string][]time.Time
}

func NewMemoryRateLimiter() *MemoryRateLimiter {
	return &MemoryRateLimiter{
		stores: make(map[string][]time.Time),
	}
}

func (m *MemoryRateLimiter) CheckRateLimit(ctx context.Context, req RateLimitRequest) (*RateLimitResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if req.Limit <= 0 {
		req.Limit = 600
	}
	if req.Window <= 0 {
		req.Window = time.Minute
	}
	if req.Cost <= 0 {
		req.Cost = 1
	}

	now := time.Now().UTC()
	cutoff := now.Add(-req.Window)

	// Filter expired entries
	timestamps := m.stores[req.Key]
	valid := make([]time.Time, 0, len(timestamps))
	for _, t := range timestamps {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}

	currentCount := len(valid)
	if currentCount+req.Cost <= req.Limit {
		for i := 0; i < req.Cost; i++ {
			valid = append(valid, now)
		}
		m.stores[req.Key] = valid
		remaining := req.Limit - len(valid)
		return &RateLimitResult{
			Allowed:    true,
			Limit:      req.Limit,
			Remaining:  remaining,
			ResetTime:  now.Add(req.Window),
			RetryAfter: 0,
		}, nil
	}

	m.stores[req.Key] = valid
	resetTime := now.Add(req.Window)
	return &RateLimitResult{
		Allowed:    false,
		Limit:      req.Limit,
		Remaining:  0,
		ResetTime:  resetTime,
		RetryAfter: time.Until(resetTime),
	}, nil
}
