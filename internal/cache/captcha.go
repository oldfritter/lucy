package cache

import (
	"time"

	"github.com/gomodule/redigo/redis"

	"github.com/oldfritter/lucy/base"
	"github.com/oldfritter/lucy/lib/kv"
)

const (
	cacheKeyPrefix    = "lucy:captcha:"
	cacheExtraMinutes = 5 // 比验证码过期时间多存 5 分钟
)

// CaptchaCacher 验证码缓存接口
type CaptchaCacher interface {
	GetCaptcha() string
	GetExpiredAt() time.Time
	Json() string
}

// SetCaptchaCache 将验证码序列化后存入 Redis，TTL = 过期时间 + 5 分钟
func SetCaptchaCache(c CaptchaCacher) error {
	conn := kv.GetRedisConn("data")
	defer conn.Close()

	ttl := time.Until(c.GetExpiredAt()) + cacheExtraMinutes*time.Minute
	if ttl <= 0 {
		ttl = cacheExtraMinutes * time.Minute
	}

	key := cacheKeyPrefix + c.GetCaptcha()
	_, err := conn.Do("SETEX", key, int64(ttl.Seconds()), c.Json())
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
