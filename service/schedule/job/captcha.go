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
		Name: "batch-update-captcha-status",
		Spec: "@daily",
		Func: batchUpdateCaptchaStatus,
	})
	Register(Job{
		Name: "sync-captcha-cache",
		Spec: "@every 2m",
		Func: syncCaptchaCache,
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

// batchUpdateCaptchaStatus 凌晨按类型分别取出并批量写入验证状态
func batchUpdateCaptchaStatus() {
	for _, ct := range captchaTypeTable {
		success, failed, err := pool.DrainVerifiedPool(ct.poolType)
		if err != nil {
			log.Printf("[batch-update-captcha-status] drain %s pool failed: %v", ct.poolType, err)
			continue
		}

		if len(success) > 0 {
			result := db.MysqlDB.Model(ct.model).
				Where("uid IN ?", success).
				Update("status", dom.CaptchaStatusSuccess)
			if result.Error != nil {
				log.Printf("[batch-update-captcha-status] %s success update failed: %v", ct.poolType, result.Error)
			} else {
				log.Printf("[batch-update-captcha-status] %s success=%d", ct.poolType, result.RowsAffected)
			}
		}

		if len(failed) > 0 {
			result := db.MysqlDB.Model(ct.model).
				Where("uid IN ?", failed).
				Update("status", dom.CaptchaStatusFailed)
			if result.Error != nil {
				log.Printf("[batch-update-captcha-status] %s failed update failed: %v", ct.poolType, result.Error)
			} else {
				log.Printf("[batch-update-captcha-status] %s failed=%d", ct.poolType, result.RowsAffected)
			}
		}
	}
}

// syncCaptchaCache 每 10 分钟将数据库中未被使用（status=1）的验证码重新写入 Redis 缓存，
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
