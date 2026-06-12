package dom

import (
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/oldfritter/lucy/util"
)

// Captcha 状态常量
const (
	CaptchaStatusActive  = 1  // 待验证
	CaptchaStatusSuccess = 2  // 验证成功
	CaptchaStatusFailed  = -1 // 验证失败
)

type Captcha struct {
	CommonModel
	UserApiKeyId *int   `gorm:"index"`                                           // 消费此验证码的 ApiKey ID（可为空）
	CampaignId   *int   `gorm:"index"`                                           // 投放此验证码的 Campaign ID（可为空）
	Status       int    `gorm:"size:8;default:1" query:"Status"`                 // 状态
	Uid          string `gorm:"size:32;uniqueIndex" query:"Uid"`                 // 唯一识别名
	ValidCode    string `gorm:"size:64"`                                         // 验证令牌，客户端持有 Uid+ValidCode 即可验证
	Key          string `gorm:"size:64" query:"Key"`                             // 存储图片的 OSS 路径
	Width        int    `gorm:"default:0"`                                       // 图片宽度（px）
	Height       int    `gorm:"default:0"`                                       // 图片高度（px）
	Suffix       string `gorm:"size:8;default:png" form:"Suffix" query:"Suffix"` // 后缀
}

func (c *Captcha) BeforeCreate(db *gorm.DB) (err error) {
	if c.Suffix == "" {
		c.Suffix = "png"
	}
	if c.Uid == "" {
		c.Uid = util.RandStringRunes(32)
	}
	if c.ValidCode == "" {
		c.ValidCode = util.RandStringRunes(32)
	}
	if c.Key == "" {
		c.Key = fmt.Sprintf("captcha/%s/%s.%s", time.Now().Format("2006/01/02"), util.RandStringRunes(32), c.Suffix)
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
