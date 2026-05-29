package dom

import (
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/oldfritter/lucy/util"
)

type Captcha struct {
	CommonModel
	UserId       int    `form:"UserId" query:"UserId"`                           // 所属用户
	UserApiKeyId *int   `gorm:"index"`                                           // 消费此验证码的 ApiKey ID（可为空）
	CampaignId   *int   `gorm:"index"`                                           // 投放此验证码的 Campaign ID（可为空）
	Status       int    `gorm:"size:8;default:1" query:"Status"`                 // 状态
	Captcha      string `gorm:"size:32;uniqueIndex:idx1" query:"Captcha"`        // 唯一识别号
	Key          string `gorm:"size:64" query:"Key"`                             // 存储图片的key
	Suffix       string `gorm:"size:8;default:png" form:"Suffix" query:"Suffix"` // 后缀
	ExpiredAt    time.Time
}

func (c *Captcha) GetCaptcha() string {
	return c.Captcha
}

func (c *Captcha) GetExpiredAt() time.Time {
	return c.ExpiredAt
}

func (c *Captcha) BeforeCreate(db *gorm.DB) (err error) {
	if c.Key == "" {
		c.Key = fmt.Sprintf("captcha/%s/%s.%s", time.Now().Format("2006/01/02"), util.RandStringRunes(32), c.Suffix)
	}
	if c.Captcha == "" {
		c.Captcha = util.RandStringRunes(32)
	}
	return
}

func (*Captcha) QueryParams(p map[string]string) map[string][]any {
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
	if p["UserApiKeyId"] != "" {
		params["user_api_key_id"] = []any{"=", p["UserApiKeyId"]}
	}
	if p["CampaignId"] != "" {
		params["campaign_id"] = []any{"=", p["CampaignId"]}
	}
	return params
}
