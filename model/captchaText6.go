package model

import (
	"encoding/json"
	"fmt"

	"gorm.io/gorm"

	"github.com/oldfritter/lucy/dom"
	"github.com/oldfritter/lucy/internal/cache"
	captchaImage "github.com/oldfritter/lucy/lib/captcha"
	"github.com/oldfritter/lucy/util"
)

type CaptchaText6 struct {
	dom.Captcha
	Prompt1  string `gorm:"size:1"`
	Prompt2  string `gorm:"size:1"`
	Prompt3  string `gorm:"size:1"`
	Prompt4  string `gorm:"size:1"`
	Prompt5  string `gorm:"size:1"`
	Prompt6  string `gorm:"size:1"`
	Verify1X int    `gorm:"size:12"`
	Verify1Y int    `gorm:"size:12"`
	Verify2X int    `gorm:"size:12"`
	Verify2Y int    `gorm:"size:12"`
	Verify3X int    `gorm:"size:12"`
	Verify3Y int    `gorm:"size:12"`
	Verify4X int    `gorm:"size:12"`
	Verify4Y int    `gorm:"size:12"`
	Verify5X int    `gorm:"size:12"`
	Verify5Y int    `gorm:"size:12"`
	Verify6X int    `gorm:"size:12"`
	Verify6Y int    `gorm:"size:12"`
}

func (*CaptchaText6) TableName() string {
	return "captcha_text6"
}

func (captcha *CaptchaText6) Json() string {
	b, _ := json.Marshal(map[string]any{
		"id": fmt.Sprintf("text6-%d", captcha.Id),

		"p1": captcha.Prompt1,
		"p2": captcha.Prompt2,
		"p3": captcha.Prompt3,
		"p4": captcha.Prompt4,
		"p5": captcha.Prompt5,
		"p6": captcha.Prompt6,

		"v1x": captcha.Verify1X,
		"v1y": captcha.Verify1Y,
		"v2x": captcha.Verify2X,
		"v2y": captcha.Verify2Y,
		"v3x": captcha.Verify3X,
		"v3y": captcha.Verify3Y,
		"v4x": captcha.Verify4X,
		"v4y": captcha.Verify4Y,
		"v5x": captcha.Verify5X,
		"v5y": captcha.Verify5Y,
		"v6x": captcha.Verify6X,
		"v6y": captcha.Verify6Y,
	})
	return string(b)
}

func (captcha *CaptchaText6) GetWithPaginate(db *gorm.DB, r *util.Response) {
	var results []*CaptchaText6
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

func (captcha *CaptchaText6) Create() {
	captchaImage.GenerateCaptchaImage(
		[]string{captcha.Prompt1, captcha.Prompt2, captcha.Prompt3, captcha.Prompt4, captcha.Prompt5, captcha.Prompt6},
		"config/background/c71eda17095e9a92e300ca207f09c778.jpg",
	)
	cache.SetCaptchaCache(captcha)
}

func (captcha *CaptchaText6) Verify(attrs map[string]any) (yes bool) {
	return
}
