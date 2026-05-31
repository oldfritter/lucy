package cache

import (
	"encoding/json"

	"github.com/gomodule/redigo/redis"

	"github.com/oldfritter/lucy/lib/kv"
)

const apiKeyPrefix = "lucy:apikey:"

// ApiKeyCache API Key 的缓存结构
type ApiKeyCache struct {
	Secret      string `json:"secret"`
	IsActive    bool   `json:"is_active"`
	UserId      int    `json:"user_id"`
	CaptchaType string `json:"captcha_type"`
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
