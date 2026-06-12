package cache

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"

	"github.com/oldfritter/lucy/base"
	"github.com/oldfritter/lucy/lib/kv"
)

const rateLimitKeyPrefix = base.RedisNamespace + ":apikey:rate:"

const apiKeyPrefix = base.RedisNamespace + ":apikey:"

// ApiKeyCache API Key 的缓存结构
type ApiKeyCache struct {
	Id             int    `json:"id"`
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

// ── API Key 分钟级访问量统计（Hash + Sorted Set）──

const apiKeyStatsHashPrefix = base.RedisNamespace + ":apikey:stats:"
const apiKeyStatsIdxPrefix = base.RedisNamespace + ":apikey:stats-idx:"

func apiKeyStatsHashKey(apiKeyId int) string {
	return fmt.Sprintf("%s%d", apiKeyStatsHashPrefix, apiKeyId)
}
func apiKeyStatsIdxKey(apiKeyId int) string {
	return fmt.Sprintf("%s%d", apiKeyStatsIdxPrefix, apiKeyId)
}

// StatsEntry 单分钟统计数据
type StatsEntry struct {
	Minute string `json:"minute"`
	Count  int    `json:"count"`
}

// RecordApiKeyRequest 记录 API Key 当前分钟的一次请求
// Hash 存储计数，Sorted Set（score=分钟时间戳）维护时间索引，25 小时后自动过期
func RecordApiKeyRequest(apiKeyId int) error {
	conn := kv.GetRedisConn("data")
	defer conn.Close()

	now := time.Now()
	minute := now.Format("200601021504")
	ts := now.Truncate(time.Minute).Unix()

	hashKey := apiKeyStatsHashKey(apiKeyId)
	idxKey := apiKeyStatsIdxKey(apiKeyId)

	conn.Send("MULTI")
	conn.Send("HINCRBY", hashKey, minute, 1)
	// NX：仅首次插入，避免重复覆盖 score
	conn.Send("ZADD", idxKey, "NX", ts, minute)
	conn.Send("EXPIRE", hashKey, 25*3600)
	conn.Send("EXPIRE", idxKey, 25*3600)
	_, err := redis.Ints(conn.Do("EXEC"))
	return err
}

// GetApiKeyStats 获取 API Key 在指定时间戳范围内的分钟级请求计数
// 返回按时间升序排列的列表，可直接作为 JSON body 返回
func GetApiKeyStats(apiKeyId int, startTs, endTs int64) ([]StatsEntry, error) {
	conn := kv.GetRedisConn("data")
	defer conn.Close()

	idxKey := apiKeyStatsIdxKey(apiKeyId)
	hashKey := apiKeyStatsHashKey(apiKeyId)

	// 从 Sorted Set 中按时间范围取出分钟标识（已按 score 升序）
	members, err := redis.Strings(conn.Do("ZRANGEBYSCORE", idxKey, startTs, endTs))
	if err != nil {
		return nil, err
	}
	if len(members) == 0 {
		return nil, nil
	}

	// 从 Hash 批量拉取计数
	args := make([]interface{}, len(members))
	for i, m := range members {
		args[i] = m
	}
	hmgetArgs := append([]interface{}{hashKey}, args...)
	replies, err := redis.Values(conn.Do("HMGET", hmgetArgs...))
	if err != nil {
		return nil, err
	}

	result := make([]StatsEntry, 0, len(members))
	for i, reply := range replies {
		if reply == nil {
			continue
		}
		count, err := redis.Int(reply, nil)
		if err == nil && count > 0 {
			result = append(result, StatsEntry{Minute: members[i], Count: count})
		}
	}
	return result, nil
}
