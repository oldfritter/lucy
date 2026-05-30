package cache

import (
	"github.com/gomodule/redigo/redis"

	"github.com/oldfritter/lucy/base"
	"github.com/oldfritter/lucy/lib/kv"
)

const (
	cacheKeyPrefix = "lucy:captcha:"
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
