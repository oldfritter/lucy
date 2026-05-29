package model

import (
	"encoding/json"
	"fmt"
	"math"

	"gorm.io/gorm"

	"github.com/oldfritter/lucy/dom"
	"github.com/oldfritter/lucy/internal/cache"
	captchaImage "github.com/oldfritter/lucy/lib/captcha"
	"github.com/oldfritter/lucy/util"
)

// CaptchaRotateImage 旋转验证码
type CaptchaRotateImage struct {
	dom.Captcha
	Indicator string  `gorm:"size:64"`       // 方向指示文字，如 "▲"
	Angle     float64 `gorm:"type:float"`    // 随机旋转角度
	Tolerance float64 `gorm:"default:15"`    // 验证容差角度，默认 15°
}

func (*CaptchaRotateImage) TableName() string {
	return "captcha_rotate_image"
}

func (captcha *CaptchaRotateImage) Json() string {
	b, _ := json.Marshal(map[string]any{
		"id":    fmt.Sprintf("rotate-%d", captcha.Id),
		"key":   captcha.Key,
		"angle": captcha.Angle,
	})
	return string(b)
}

func (captcha *CaptchaRotateImage) GetWithPaginate(db *gorm.DB, r *util.Response) {
	var results []*CaptchaRotateImage
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

// Create 生成旋转验证码图片并写入缓存
func (captcha *CaptchaRotateImage) Create() {
	if captcha.Indicator == "" {
		captcha.Indicator = "▲"
	}
	if captcha.Tolerance == 0 {
		captcha.Tolerance = 15
	}

	_, captcha.Angle = captchaImage.GenerateRotateCaptcha(
		"config/background/c71eda17095e9a92e300ca207f09c778.jpg",
		captcha.Indicator,
	)
	cache.SetCaptchaCache(captcha)
}

// Verify 验证用户提交的旋转角度
func (captcha *CaptchaRotateImage) Verify(attrs map[string]any) (yes bool) {
	angle, ok := attrs["angle"].(float64)
	if !ok {
		return false
	}
	diff := math.Abs(captcha.Angle - angle)
	return diff <= captcha.Tolerance
}
