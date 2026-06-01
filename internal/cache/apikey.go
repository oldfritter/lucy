package cache

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"

	"github.com/oldfritter/lucy/lib/kv"
)

const rateLimitKeyPrefix = "rate:apikey:"

const apiKeyPrefix = "lucy:apikey:"

// ApiKeyCache API Key 的缓存结构
type ApiKeyCache struct {
	Secret         string `json:"secret"`
	IsActive       bool   `json:"is_active"`
	UserId         int    `json:"user_id"`
	CaptchaType    string `json:"captcha_type"`
	PerMinuteLimit int    `json:"per_minute_limit"`
}

// GetApiKeyCache 从 Redis 获取 API Key 缓存
func GetApiKeyCache(key string) (*ApiKeyCache, error) {
	conn := kv.GetRedisConn("data")
	defer conn.Close()

	data, err := redis.String(conn.Do("GET", apiKeyPrefix+key))
	if err != nil {
		return nil, err
	}

	var cache ApiKeyCache
	if err := json.Unmarshal([]byte(data), &cache); err != nil {
		return nil, err
	}
	return &cache, nil
}

// SetApiKeyCache 写入/更新单个 API Key 到 Redis
func SetApiKeyCache(key string, cache *ApiKeyCache) error {
	conn := kv.GetRedisConn("data")
	defer conn.Close()

	data, _ := json.Marshal(cache)
	_, err := conn.Do("SET", apiKeyPrefix+key, string(data))
	return err
}

// DelApiKeyCache 从 Redis 删除单个 API Key
func DelApiKeyCache(key string) error {
	conn := kv.GetRedisConn("data")
	defer conn.Close()

	_, err := conn.Do("DEL", apiKeyPrefix+key)
	return err
}

// rateLimitKey 生成每分钟限速的 Redis key
func rateLimitKey(apiKey string) string {
	minute := time.Now().Format("200601021504")
	return fmt.Sprintf("%s%s:%s", rateLimitKeyPrefix, apiKey, minute)
}

// CheckRateLimit 检查 API Key 是否超出每分钟验证次数限制
//
//	返回 true 表示未超限（可继续请求），false 表示已超限。
//	每次调用原子递增计数器，计数器每分钟自动过期。
func CheckRateLimit(apiKey string, perMinuteLimit int) (bool, error) {
	if perMinuteLimit <= 0 {
		return true, nil
	}

	conn := kv.GetRedisConn("data")
	defer conn.Close()

	key := rateLimitKey(apiKey)

	// INCR + 设置 60 秒过期（确保即使没有显式 EXPIRE 也会自动清理）
	conn.Send("MULTI")
	conn.Send("INCR", key)
	conn.Send("EXPIRE", key, 600)
	values, err := redis.Ints(conn.Do("EXEC"))
	if err != nil {
		return false, fmt.Errorf("速率检查失败: %w", err)
	}
	if len(values) < 1 {
		return true, nil
	}

	count := values[0]
	if count > perMinuteLimit {
		return false, nil
	}
	return true, nil
}
