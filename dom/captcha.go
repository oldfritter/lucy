package dom

import (
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/oldfritter/lucy/util"
)

type Captcha struct {
	CommonModel
	UserId    int    `form:"UserId" query:"UserId"`                           // 所属用户
	Kind      int    `gorm:"size:8;default:1" query:"Kind"`                   // 1: 点击图片上的文字；2: 拼图滑块
	Status    int    `gorm:"size:8;default:1" query:"Status"`                 // 状态
	Captcha   string `gorm:"size:32;uniqueIndex:idx1" query:"Captcha"`        // 唯一识别号
	Key       string `gorm:"size:64" query:"Key"`                             // 存储图片的key
	Suffix    string `gorm:"size:8;default:png" form:"Suffix" query:"Suffix"` // 后缀
	ExpiredAt time.Time
}

func (captcha *Captcha) BeforeCreate(db *gorm.DB) (err error) {
	if captcha.Key == "" {
		captcha.Key = fmt.Sprintf("captcha/%s/%s.%s", time.Now().Format("2006/01/02"), util.RandStringRunes(32), captcha.Suffix)
	}
	if captcha.Captcha == "" {
		captcha.Captcha = util.RandStringRunes(32)
	}
	return
}

func (captcha *Captcha) QueryParams(p map[string]string) map[string][]any {
	params := make(map[string][]any)
	if p["Key"] != "" {
		params["key"] = []any{"like", p["Key"]}
	}
	if p["Suffix"] != "" {
		params["suffix"] = []any{"suffix", p["Suffix"]}
	}
	if p["Captcha"] != "" {
		params["captcha_id"] = []any{"=", p["Captcha"]}
	}
	if p["Status"] != "" {
		params["status"] = []any{"=", p["Status"]}
	}
	if p["Kind"] != "" {
		params["kind"] = []any{"=", p["Kind"]}
	}
	return params
}
