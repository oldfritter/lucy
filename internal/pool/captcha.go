package pool

import (
	"github.com/gomodule/redigo/redis"

	"github.com/oldfritter/lucy/base"
	"github.com/oldfritter/lucy/lib/kv"
)

func poolKey(captchaType string) string {
	return base.RedisNamespace + ":captcha:pool:" + captchaType
}

// ── 验证结果批量更新池 ──

func verifiedSuccessKey(captchaType string) string {
	return base.RedisNamespace + ":captcha:verified:success:" + captchaType
}
func verifiedFailedKey(captchaType string) string {
	return base.RedisNamespace + ":captcha:verified:failed:" + captchaType
}

// AddToVerifiedPool 将 UID 放入对应类型的验证结果池（success / failed）
func AddToVerifiedPool(captchaType, uid string, success bool) error {
	conn := kv.GetRedisConn("data")
	defer conn.Close()
	key := verifiedFailedKey(captchaType)
	if success {
		key = verifiedSuccessKey(captchaType)
	}
	_, err := conn.Do("SADD", key, uid)
	return err
}

// DrainVerifiedPool 取出并清空某类型验证结果池，返回两组 UID
func DrainVerifiedPool(captchaType string) (success []string, failed []string, err error) {
	conn := kv.GetRedisConn("data")
	defer conn.Close()

	sk := verifiedSuccessKey(captchaType)
	fk := verifiedFailedKey(captchaType)

	success, err = redis.Strings(conn.Do("SMEMBERS", sk))
	if err != nil && err != redis.ErrNil {
		return nil, nil, err
	}
	if len(success) > 0 {
		conn.Do("DEL", sk)
	}

	failed, err = redis.Strings(conn.Do("SMEMBERS", fk))
	if err != nil && err != redis.ErrNil {
		return nil, nil, err
	}
	if len(failed) > 0 {
		conn.Do("DEL", fk)
	}

	return success, failed, nil
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
	types := []string{"text:4", "text:5", "text:6", "image:rotate"}
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

// IsInVerifiedPool 检查 uid 是否已经被验证过（存在于任一类型的 success 或 failed 池中）
func IsInVerifiedPool(uid string) (bool, error) {
	conn := kv.GetRedisConn("data")
	defer conn.Close()
	types := []string{"text:4", "text:5", "text:6", "image:rotate"}
	for _, t := range types {
		for _, kind := range []string{"success", "failed"} {
			key := base.RedisNamespace + ":captcha:verified:" + kind + ":" + t
			n, err := redis.Int(conn.Do("SISMEMBER", key, uid))
			if err != nil {
				return false, err
			}
			if n > 0 {
				return true, nil
			}
		}
	}
	return false, nil
}
