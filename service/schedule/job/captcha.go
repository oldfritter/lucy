package job

import (
	"log"

	"github.com/oldfritter/lucy/dom"
	"github.com/oldfritter/lucy/internal/pool"
	"github.com/oldfritter/lucy/lib/db"
	"github.com/oldfritter/lucy/model"
)

func init() {
	Register(Job{
		Name: "batch-status",
		Spec: "@daily",
		Func: batchUpdateCaptchaStatus,
	})
}

// captchaTypeTable 验证码类型到 DB 模型的映射
var captchaTypeTable = []struct {
	poolType string
	model    any
}{
	{"text4", &model.CaptchaText4{}},
	{"text5", &model.CaptchaText5{}},
	{"text6", &model.CaptchaText6{}},
	{"rotate", &model.CaptchaImageRotate{}},
}

// batchUpdateCaptchaStatus 凌晨按类型分别取出并批量写入验证状态
func batchUpdateCaptchaStatus() {
	for _, ct := range captchaTypeTable {
		success, failed, err := pool.DrainVerifiedPool(ct.poolType)
		if err != nil {
			log.Printf("[batch-status] drain %s pool failed: %v", ct.poolType, err)
			continue
		}

		if len(success) > 0 {
			result := db.MysqlDB.Model(ct.model).
				Where("uid IN ?", success).
				Update("status", dom.CaptchaStatusSuccess)
			if result.Error != nil {
				log.Printf("[batch-status] %s success update failed: %v", ct.poolType, result.Error)
			} else {
				log.Printf("[batch-status] %s success=%d", ct.poolType, result.RowsAffected)
			}
		}

		if len(failed) > 0 {
			result := db.MysqlDB.Model(ct.model).
				Where("uid IN ?", failed).
				Update("status", dom.CaptchaStatusFailed)
			if result.Error != nil {
				log.Printf("[batch-status] %s failed update failed: %v", ct.poolType, result.Error)
			} else {
				log.Printf("[batch-status] %s failed=%d", ct.poolType, result.RowsAffected)
			}
		}
	}
}
