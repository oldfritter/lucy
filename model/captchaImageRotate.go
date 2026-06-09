package model

import (
	"encoding/json"
	"fmt"
	"math"

	"gorm.io/gorm"

	"github.com/oldfritter/lucy/dom"
	"github.com/oldfritter/lucy/internal/cache"
	captchaImage "github.com/oldfritter/lucy/lib/captcha"
	"github.com/oldfritter/lucy/lib/storage/oss"
	"github.com/oldfritter/lucy/util"
)

// CaptchaImageRotate 旋转验证码
type CaptchaImageRotate struct {
	dom.Captcha
	Indicator string  `gorm:"size:64"`    // 方向指示文字，如 "▲"
	Angle     float64 `gorm:"type:float"` // 随机旋转角度
	Tolerance float64 `gorm:"default:15"` // 验证容差角度，默认 15°
}

func (*CaptchaImageRotate) TableName() string {
	return "captcha_image_rotate"
}

func (captcha *CaptchaImageRotate) GetCaptcha() string {
	return fmt.Sprintf("image:rotate:%s", captcha.Uid)
}

func (captcha *CaptchaImageRotate) AfterCreate(tx *gorm.DB) error {
	return cache.SetCaptchaCache(captcha)
}

func (captcha *CaptchaImageRotate) AfterUpdate(tx *gorm.DB) error {
	return cache.SetCaptchaCache(captcha)
}

func (captcha *CaptchaImageRotate) AfterDelete(tx *gorm.DB) error {
	_ = oss.DeleteObject(captcha.Key)
	return nil
}

func (captcha *CaptchaImageRotate) Json() string {
	b, _ := json.Marshal(map[string]any{
		"valid_code": captcha.ValidCode,
		"key":        captcha.Key,
		"campaign_id": captcha.CampaignId,
		"width": captcha.Width,
		"height": captcha.Height,
		"angle":      captcha.Angle,
	})
	return string(b)
}

func (captcha *CaptchaImageRotate) GetWithPaginate(db *gorm.DB, r *util.Response) {
	var results []*CaptchaImageRotate
	where, values := captcha.WhereBuild(captcha.QueryParams(r.Params))
	condition := db.Model(captcha).Where(where, values...)
	condition.Count(&r.Pagination.Count)
	r.Pagination.Init()
	if err := condition.
		Order(captcha.TableName() + "." + r.Pagination.Order).
		Offset((int(r.Pagination.CurrentPage) - 1) * int(r.Pagination.PerPage)).
		Limit(int(r.Pagination.PerPage)).
		Find(&results).Error; err != nil {
		return
	}
	r.Body = results
}

// Create 生成旋转验证码图片并写入缓存，同时更新图片宽高
func (captcha *CaptchaImageRotate) Create() {
	if captcha.Tolerance == 0 {
		captcha.Tolerance = 15
	}

	// 先生成图片，获取实际尺寸
	_, captcha.Angle = captchaImage.GenerateRotateCaptcha()
	// 随机取一张背景图来确定尺寸
	if len(captchaImage.BackgroundImgs) > 0 {
		bg := captchaImage.BackgroundImgs[0]
		bounds := bg.Bounds()
		captcha.Width = bounds.Dx()
		captcha.Height = bounds.Dy()
	} else {
		captcha.Width = 300
		captcha.Height = 300
	}
}

// Verify 验证用户提交的旋转角度
func (captcha *CaptchaImageRotate) Verify(attrs map[string]any) (yes bool) {
	angle, ok := attrs["angle"].(float64)
	if !ok {
		return false
	}
	diff := math.Abs(captcha.Angle - angle)
	return diff <= captcha.Tolerance
}
