package pool

import (
	"github.com/gomodule/redigo/redis"

	"github.com/oldfritter/lucy/base"
	"github.com/oldfritter/lucy/lib/kv"
)

func poolKey(captchaType string) string {
	return base.RedisNamespace + ":captcha:pool:" + captchaType
}

func AddToPool(captchaType, uid string) error {
	conn := kv.GetRedisConn("data")
	defer conn.Close()
	_, err := conn.Do("SADD", poolKey(captchaType), uid)
	return err
}

func PopFromPool(captchaType string) (string, error) {
	conn := kv.GetRedisConn("data")
	defer conn.Close()
	uid, err := redis.String(conn.Do("SPOP", poolKey(captchaType)))
	if err == redis.ErrNil {
		return "", nil
	}
	return uid, err
}

func RemoveFromPool(uid string) (bool, error) {
	conn := kv.GetRedisConn("data")
	defer conn.Close()
	types := []string{"text4", "text5", "text6", "rotate"}
	for _, t := range types {
		n, err := redis.Int(conn.Do("SREM", poolKey(t), uid))
		if err != nil {
			return false, err
		}
		if n > 0 {
			return true, nil
		}
	}
	return false, nil
}

func PoolSize(captchaType string) (int, error) {
	conn := kv.GetRedisConn("data")
	defer conn.Close()
	return redis.Int(conn.Do("SCARD", poolKey(captchaType)))
}
