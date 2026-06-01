package model

import (
	"encoding/json"
	"fmt"

	"gorm.io/gorm"

	"github.com/oldfritter/lucy/dom"
	"github.com/oldfritter/lucy/internal/cache"
	captchaImage "github.com/oldfritter/lucy/lib/captcha"
	"github.com/oldfritter/lucy/lib/storage/oss"
	"github.com/oldfritter/lucy/util"
)

type CaptchaText4 struct {
	dom.Captcha
	Prompt1  string `gorm:"size:1"`
	Prompt2  string `gorm:"size:1"`
	Prompt3  string `gorm:"size:1"`
	Prompt4  string `gorm:"size:1"`
	Verify1X int    `gorm:"size:12"`
	Verify1Y int    `gorm:"size:12"`
	Verify2X int    `gorm:"size:12"`
	Verify2Y int    `gorm:"size:12"`
	Verify3X int    `gorm:"size:12"`
	Verify3Y int    `gorm:"size:12"`
	Verify4X int    `gorm:"size:12"`
	Verify4Y int    `gorm:"size:12"`
}

func (*CaptchaText4) TableName() string {
	return "captcha_text_4"
}

func (captcha *CaptchaText4) GetCaptcha() string {
	return fmt.Sprintf("text:4:%s", captcha.Uid)
}

func (captcha *CaptchaText4) AfterCreate(tx *gorm.DB) error {
	return cache.SetCaptchaCache(captcha)
}

func (captcha *CaptchaText4) AfterUpdate(tx *gorm.DB) error {
	return cache.SetCaptchaCache(captcha)
}

func (captcha *CaptchaText4) AfterDelete(tx *gorm.DB) error {
	_ = oss.DeleteObject(captcha.Key)
	return nil
}

func (captcha *CaptchaText4) Json() string {
	b, _ := json.Marshal(map[string]any{
		"uid":        captcha.Uid,
		"valid_code": captcha.ValidCode,
		"key":        captcha.Key,

		"p1": captcha.Prompt1,
		"p2": captcha.Prompt2,
		"p3": captcha.Prompt3,
		"p4": captcha.Prompt4,

		"v1x": captcha.Verify1X,
		"v1y": captcha.Verify1Y,
		"v2x": captcha.Verify2X,
		"v2y": captcha.Verify2Y,
		"v3x": captcha.Verify3X,
		"v3y": captcha.Verify3Y,
		"v4x": captcha.Verify4X,
		"v4y": captcha.Verify4Y,
	})
	return string(b)
}

func (captcha *CaptchaText4) GetWithPaginate(db *gorm.DB, r *util.Response) {
	var results []*CaptchaText4
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

func (captcha *CaptchaText4) Create() {
	captchaImage.GenerateCaptchaImage(
		[]string{captcha.Prompt1, captcha.Prompt2, captcha.Prompt3, captcha.Prompt4},
	)
}

func (captcha *CaptchaText4) Verify(attrs map[string]any) (yes bool) {
	points, ok := attrs["points"].([][]int)
	if !ok || len(points) != 4 {
		return false
	}
	expected := [][]int{
		{captcha.Verify1X, captcha.Verify1Y},
		{captcha.Verify2X, captcha.Verify2Y},
		{captcha.Verify3X, captcha.Verify3Y},
		{captcha.Verify4X, captcha.Verify4Y},
	}
	return matchTextPoints(points, expected)
}

// matchTextPoints 坐标匹配：所有点与期望点的欧几里得距离均 ≤ 30px
func matchTextPoints(input, expected [][]int) bool {
	for i, p := range input {
		dx := float64(p[0] - expected[i][0])
		dy := float64(p[1] - expected[i][1])
		if dx*dx+dy*dy > 900 { // 30²
			return false
		}
	}
	return true
}
