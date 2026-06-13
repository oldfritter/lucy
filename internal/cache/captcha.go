package cache

import (
	"github.com/gomodule/redigo/redis"

	"github.com/oldfritter/lucy/base"
	"github.com/oldfritter/lucy/lib/kv"
)

const (
	cacheKeyPrefix = base.RedisNamespace + ":captcha:"
)

// CaptchaCacher 验证码缓存接口
type CaptchaCacher interface {
	GetCaptcha() string
	Json() string
}

// SetCaptchaCache 将验证码序列化后存入 Redis（持久存储，消费时删除）
func SetCaptchaCache(c CaptchaCacher) error {
	conn := kv.GetRedisConn("data")
	defer conn.Close()

	key := cacheKeyPrefix + c.GetCaptcha()
	_, err := conn.Do("SET", key, c.Json())
	return err
}

// GetCaptchaCache 从 Redis 读取验证码缓存
func GetCaptchaCache(captcha string) (string, error) {
	conn := kv.GetRedisConn("data")
	defer conn.Close()

	key := cacheKeyPrefix + captcha
	return redis.String(conn.Do("GET", key))
}

// DelCaptchaCache 删除 Redis 中的验证码缓存
func DelCaptchaCache(captcha string) error {
	conn := kv.GetRedisConn("data")
	defer conn.Close()

	key := cacheKeyPrefix + captcha
	_, err := conn.Do("DEL", key)
	return err
}

// BuildCacheKey 返回完整 Redis key
func BuildCacheKey(captcha string) string {
	return base.RedisNamespace + ":" + "captcha:" + captcha
}

// ── Fetch 阶段 owner 暂存（uid → user_api_key_id，供 batch 任务批量回写 DB）──

const fetchOwnerHashKey = base.RedisNamespace + ":captcha:fetch-owner"

// SetFetchOwner 记录 uid 被哪个 ApiKey 消费，1 天自动过期
func SetFetchOwner(uid string, apiKeyId int) error {
	conn := kv.GetRedisConn("data")
	defer conn.Close()
	conn.Send("MULTI")
	conn.Send("HSET", fetchOwnerHashKey, uid, apiKeyId)
	conn.Send("EXPIRE", fetchOwnerHashKey, 86400)
	_, err := redis.Ints(conn.Do("EXEC"))
	return err
}

// GetFetchOwner 查询 uid 对应的 ApiKey ID
func GetFetchOwner(uid string) (int, error) {
	conn := kv.GetRedisConn("data")
	defer conn.Close()
	return redis.Int(conn.Do("HGET", fetchOwnerHashKey, uid))
}

// DelFetchOwner 删除 uid 的 owner 记录（batch 处理后调用）
func DelFetchOwner(uid string) error {
	conn := kv.GetRedisConn("data")
	defer conn.Close()
	_, err := conn.Do("HDEL", fetchOwnerHashKey, uid)
	return err
}
