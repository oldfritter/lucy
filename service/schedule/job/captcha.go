package job

import (
	"log"

	"github.com/oldfritter/lucy/dom"
	"github.com/oldfritter/lucy/internal/cache"
	"github.com/oldfritter/lucy/internal/pool"
	"github.com/oldfritter/lucy/lib/db"
	"github.com/oldfritter/lucy/model"
)

func init() {
	Register(Job{
		Name: "batch-confirm-captcha-success",
		Spec: "@every 5m",
		Func: batchConfirmCaptchaSuccess,
	})
	Register(Job{
		Name: "sync-captcha-cache",
		Spec: "0 0 1 * * *",
		Func: syncCaptchaCache,
	})
	Register(Job{
		Name: "recycle-failed-captcha",
		Spec: "@every 10m",
		Func: recycleFailedCaptcha,
	})
	Register(Job{
		Name: "recycle-pending-captcha",
		Spec: "@every 5m",
		Func: recyclePendingCaptcha,
	})
}

// captchaTypeTable 验证码类型到 DB 模型的映射
var captchaTypeTable = []struct {
	poolType string
	model    any
}{
	{"text:4", &model.CaptchaText4{}},
	{"text:5", &model.CaptchaText5{}},
	{"text:6", &model.CaptchaText6{}},
	{"image:rotate", &model.CaptchaImageRotate{}},
}

// captchaTypeTableName 验证码类型到 DB 表名的映射（避免 GORM Model 通过 any 类型推导表名）
var captchaTypeTableName = map[string]string{
	"text:4":        "captcha_text_4",
	"text:5":        "captcha_text_5",
	"text:6":        "captcha_text_6",
	"image:rotate":  "captcha_image_rotate",
}

// batchConfirmCaptchaSuccess 定期从 success 池取出已验证通过的 UID，批量写入 DB status=2
func batchConfirmCaptchaSuccess() {
	for _, ct := range captchaTypeTable {
		success, err := pool.DrainVerifiedPool(ct.poolType)
		if err != nil {
			log.Printf("[batch-confirm-captcha-success] drain %s pool failed: %v", ct.poolType, err)
			continue
		}

		if len(success) > 0 {
			tableName, ok := captchaTypeTableName[ct.poolType]
			if !ok {
				log.Printf("[batch-confirm-captcha-success] unknown type: %s", ct.poolType)
				continue
			}
			result := db.MysqlDB.Table(tableName).
				Where("uid IN ?", success).
				Update("status", dom.CaptchaStatusSuccess)
			if result.Error != nil {
				log.Printf("[batch-confirm-captcha-success] %s update failed: %v", ct.poolType, result.Error)
			} else {
				log.Printf("[batch-confirm-captcha-success] %s count=%d", ct.poolType, result.RowsAffected)
			}

			// 回写 fetch 阶段暂存的 user_api_key_id
			for _, uid := range success {
				apiKeyId, err := cache.GetFetchOwner(uid)
				if err != nil || apiKeyId <= 0 {
					continue
				}
				if err := db.MysqlDB.Table(tableName).
					Where("uid = ?", uid).
					Update("user_api_key_id", apiKeyId).Error; err != nil {
					log.Printf("[batch-confirm-captcha-success] %s uid=%s update user_api_key_id failed: %v", ct.poolType, uid, err)
				}
				cache.DelFetchOwner(uid)
			}
		}
	}
}

const (
	maxRecallAttempts        = 10
	maxPendingRecallAttempts = 3
)

// recycleFailedCaptcha 每10分钟将验证失败的验证码重新送入待用池，
// 单个验证码回收次数超过 maxRecallAttempts（10次）则丢弃，防止积压。
func recycleFailedCaptcha() {
	for _, ct := range captchaTypeTable {
		uids, err := pool.DrainFailedPool(ct.poolType)
		if err != nil {
			log.Printf("[recycle-failed-captcha] drain %s failed pool error: %v", ct.poolType, err)
			continue
		}
		if len(uids) == 0 {
			continue
		}

		recalled := 0
		discarded := 0
		for _, uid := range uids {
			count, err := pool.IncrRecallCount(uid)
			if err != nil {
				log.Printf("[recycle-failed-captcha] incr recall count for %s failed: %v", uid, err)
				continue
			}
			if count <= maxRecallAttempts {
				if err := pool.AddToPool(ct.poolType, uid); err != nil {
					log.Printf("[recycle-failed-captcha] add %s to pool %s failed: %v", uid, ct.poolType, err)
					continue
				}
				recalled++
			} else {
				discarded++
				log.Printf("[recycle-failed-captcha] %s recall count=%d exceeds max, discarded", uid, count)
			}
		}
		if recalled > 0 || discarded > 0 {
			log.Printf("[recycle-failed-captcha] %s recalled=%d discarded=%d", ct.poolType, recalled, discarded)
		}
	}
}

// recyclePendingCaptcha 每5分钟扫描待验证池中超时的验证码，
//
//	回收到待用池（上限 maxPendingRecallAttempts 次），超出则放入失败池。
func recyclePendingCaptcha() {
	for _, ct := range captchaTypeTable {
		uids, err := pool.ExpiredFromPendingPool(ct.poolType)
		if err != nil {
			log.Printf("[recycle-pending-captcha] drain %s pending pool error: %v", ct.poolType, err)
			continue
		}
		if len(uids) == 0 {
			continue
		}

		recalled := 0
		failed := 0
		for _, uid := range uids {
			// 如果已经被验证（在成功或失败池中），跳过
			verified, err := pool.IsInVerifiedPool(uid)
			if err != nil {
				log.Printf("[recycle-pending-captcha] check verified for %s: %v", uid, err)
				continue
			}
			if verified {
				continue
			}

			count, err := pool.IncrRecallCount(uid)
			if err != nil {
				log.Printf("[recycle-pending-captcha] incr recall for %s: %v", uid, err)
				continue
			}

			if count <= maxPendingRecallAttempts {
				if err := pool.AddToPool(ct.poolType, uid); err != nil {
					log.Printf("[recycle-pending-captcha] add %s to pool %s: %v", uid, ct.poolType, err)
					continue
				}
				recalled++
			} else {
				// 超过 3 次，视为验证失败
				if err := pool.AddToVerifiedPool(ct.poolType, uid, false); err != nil {
					log.Printf("[recycle-pending-captcha] add %s to failed pool: %v", uid, err)
					continue
				}
				failed++
			}
		}
		if recalled > 0 || failed > 0 {
			log.Printf("[recycle-pending-captcha] %s recalled=%d failed=%d", ct.poolType, recalled, failed)
		}
	}
}

// syncCaptchaCache 在 batchConfirmCaptchaSuccess 执行后 1 小时（凌晨 1 点）将数据库中未被使用（status=1）的验证码重新写入 Redis 缓存，
// 防止因缓存过期或重启导致池中验证码无法被 FetchCaptcha 获取。
func syncCaptchaCache() {
	// 四张验证码表，各自查询 status=1 的记录，跳过已验证的 uid

	// ---- CaptchaText4 ----
	var list4 []model.CaptchaText4
	if err := db.MysqlDB.Where("status = ?", dom.CaptchaStatusActive).Find(&list4).Error; err != nil {
		log.Printf("[sync-captcha-cache] query captcha_text_4 failed: %v", err)
	} else {
		synced := 0
		for i := range list4 {
			if verified, err := pool.IsInVerifiedPool(list4[i].Uid); err != nil {
				log.Printf("[sync-captcha-cache] check verified pool failed for %s: %v", list4[i].Uid, err)
				continue
			} else if verified {
				continue
			}
			if err := cache.SetCaptchaCache(&list4[i]); err != nil {
				log.Printf("[sync-captcha-cache] text:4 uid=%s cache set failed: %v", list4[i].Uid, err)
				continue
			}
			if err := pool.AddToPool("text:4", list4[i].Uid); err != nil {
				log.Printf("[sync-captcha-cache] text:4 uid=%s pool add failed: %v", list4[i].Uid, err)
				continue
			}
			synced++
		}
		if synced > 0 {
			log.Printf("[sync-captcha-cache] text:4 synced %d", synced)
		}
	}

	// ---- CaptchaText5 ----
	var list5 []model.CaptchaText5
	if err := db.MysqlDB.Where("status = ?", dom.CaptchaStatusActive).Find(&list5).Error; err != nil {
		log.Printf("[sync-captcha-cache] query captcha_text_5 failed: %v", err)
	} else {
		synced := 0
		for i := range list5 {
			if verified, err := pool.IsInVerifiedPool(list5[i].Uid); err != nil {
				log.Printf("[sync-captcha-cache] check verified pool failed for %s: %v", list5[i].Uid, err)
				continue
			} else if verified {
				continue
			}
			if err := cache.SetCaptchaCache(&list5[i]); err != nil {
				log.Printf("[sync-captcha-cache] text:5 uid=%s cache set failed: %v", list5[i].Uid, err)
				continue
			}
			if err := pool.AddToPool("text:5", list5[i].Uid); err != nil {
				log.Printf("[sync-captcha-cache] text:5 uid=%s pool add failed: %v", list5[i].Uid, err)
				continue
			}
			synced++
		}
		if synced > 0 {
			log.Printf("[sync-captcha-cache] text:5 synced %d", synced)
		}
	}

	// ---- CaptchaText6 ----
	var list6 []model.CaptchaText6
	if err := db.MysqlDB.Where("status = ?", dom.CaptchaStatusActive).Find(&list6).Error; err != nil {
		log.Printf("[sync-captcha-cache] query captcha_text_6 failed: %v", err)
	} else {
		synced := 0
		for i := range list6 {
			if verified, err := pool.IsInVerifiedPool(list6[i].Uid); err != nil {
				log.Printf("[sync-captcha-cache] check verified pool failed for %s: %v", list6[i].Uid, err)
				continue
			} else if verified {
				continue
			}
			if err := cache.SetCaptchaCache(&list6[i]); err != nil {
				log.Printf("[sync-captcha-cache] text:6 uid=%s cache set failed: %v", list6[i].Uid, err)
				continue
			}
			if err := pool.AddToPool("text:6", list6[i].Uid); err != nil {
				log.Printf("[sync-captcha-cache] text:6 uid=%s pool add failed: %v", list6[i].Uid, err)
				continue
			}
			synced++
		}
		if synced > 0 {
			log.Printf("[sync-captcha-cache] text:6 synced %d", synced)
		}
	}

	// ---- CaptchaImageRotate ----
	var listR []model.CaptchaImageRotate
	if err := db.MysqlDB.Where("status = ?", dom.CaptchaStatusActive).Find(&listR).Error; err != nil {
		log.Printf("[sync-captcha-cache] query captcha_image_rotate failed: %v", err)
	} else {
		synced := 0
		for i := range listR {
			if verified, err := pool.IsInVerifiedPool(listR[i].Uid); err != nil {
				log.Printf("[sync-captcha-cache] check verified pool failed for %s: %v", listR[i].Uid, err)
				continue
			} else if verified {
				continue
			}
			if err := cache.SetCaptchaCache(&listR[i]); err != nil {
				log.Printf("[sync-captcha-cache] image:rotate uid=%s cache set failed: %v", listR[i].Uid, err)
				continue
			}
			if err := pool.AddToPool("image:rotate", listR[i].Uid); err != nil {
				log.Printf("[sync-captcha-cache] image:rotate uid=%s pool add failed: %v", listR[i].Uid, err)
				continue
			}
			synced++
		}
		if synced > 0 {
			log.Printf("[sync-captcha-cache] image:rotate synced %d", synced)
		}
	}
}
