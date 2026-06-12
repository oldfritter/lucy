package pool

import (
	"fmt"
	"time"

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


// ── 各 system 类型 campaign 的独立待用验证码池 ──

func campaignPoolKey(campaignId int) string {
	return base.RedisNamespace + ":captcha:campaign:pool:" + fmt.Sprint(campaignId)
}

// AddToCampaignPool 将 uid 加入 campaign 的待用验证码池
func AddToCampaignPool(campaignId int, uid string) error {
	conn := kv.GetRedisConn("data")
	defer conn.Close()
	_, err := conn.Do("SADD", campaignPoolKey(campaignId), uid)
	return err
}

// RemoveFromCampaignPool 将 uid 从 campaign 的待用验证码池中移除
func RemoveFromCampaignPool(campaignId int, uid string) error {
	conn := kv.GetRedisConn("data")
	defer conn.Close()
	_, err := conn.Do("SREM", campaignPoolKey(campaignId), uid)
	return err
}

// CampaignPoolSize 返回 campaign 待用验证码池中的存量
func CampaignPoolSize(campaignId int) (int64, error) {
	conn := kv.GetRedisConn("data")
	defer conn.Close()
	return redis.Int64(conn.Do("SCARD", campaignPoolKey(campaignId)))
}

// ── 验证失败验证码回收 ──

// DrainFailedPool 取出并清空某类型验证失败池，返回 UID 列表
func DrainFailedPool(captchaType string) ([]string, error) {
	conn := kv.GetRedisConn("data")
	defer conn.Close()

	fk := verifiedFailedKey(captchaType)
	uids, err := redis.Strings(conn.Do("SMEMBERS", fk))
	if err != nil && err != redis.ErrNil {
		return nil, err
	}
	if len(uids) > 0 {
		conn.Do("DEL", fk)
	}
	return uids, nil
}

// recallCountKey 验证码回收次数的 Redis key
func recallCountKey(uid string) string {
	return base.RedisNamespace + ":captcha:recall-count:" + uid
}

// GetRecallCount 获取当前回收次数
func GetRecallCount(uid string) (int, error) {
	conn := kv.GetRedisConn("data")
	defer conn.Close()
	n, err := redis.Int(conn.Do("GET", recallCountKey(uid)))
	if err == redis.ErrNil {
		return 0, nil
	}
	return n, err
}

// IncrRecallCount 递增回收次数并刷新 7 天 TTL，返回递增后的值
func IncrRecallCount(uid string) (int, error) {
	conn := kv.GetRedisConn("data")
	defer conn.Close()
	key := recallCountKey(uid)
	n, err := redis.Int(conn.Do("INCR", key))
	if err != nil {
		return 0, err
	}
	conn.Do("EXPIRE", key, 7*24*3600)
	return n, nil
}

// ResetRecallCount 删除回收次数记录（验证码成功消费后调用）
func ResetRecallCount(uid string) error {
	conn := kv.GetRedisConn("data")
	defer conn.Close()
	_, err := conn.Do("DEL", recallCountKey(uid))
	return err
}

// ── 捞取后未验证回收 ──

const (
	maxRecall  = 3
	pendingTTL = 10 * 60 // 10 分钟
)

func pendingKey(captchaType string) string {
	return base.RedisNamespace + ":captcha:pending:" + captchaType
}

// AddToPendingPool 将 uid 记录为已捞取、等待验证状态
func AddToPendingPool(captchaType, uid string) error {
	conn := kv.GetRedisConn("data")
	defer conn.Close()
	_, err := conn.Do("ZADD", pendingKey(captchaType), time.Now().Unix()+pendingTTL, uid)
	return err
}

// RemoveFromPendingPool 从所有类型的 pending 池中移除 uid
func RemoveFromPendingPool(uid string) error {
	conn := kv.GetRedisConn("data")
	defer conn.Close()
	types := []string{"text:4", "text:5", "text:6", "image:rotate"}
	for _, t := range types {
		conn.Do("ZREM", pendingKey(t), uid)
	}
	return nil
}

// ExpiredFromPendingPool 返回某类型中已超时的 uid 并移除
func ExpiredFromPendingPool(captchaType string) ([]string, error) {
	conn := kv.GetRedisConn("data")
	defer conn.Close()
	now := time.Now().Unix()
	key := pendingKey(captchaType)
	uids, err := redis.Strings(conn.Do("ZRANGEBYSCORE", key, "-inf", now))
	if err != nil {
		return nil, err
	}
	if len(uids) > 0 {
		args := make([]any, 0, len(uids)+1)
		args = append(args, key)
		for _, u := range uids {
			args = append(args, u)
		}
		conn.Do("ZREM", args...)
	}
	return uids, nil
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
