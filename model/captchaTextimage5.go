package model

import (
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg"
	"log"
	"os"

	"gorm.io/gorm"

	"github.com/oldfritter/lucy/dom"
	"github.com/oldfritter/lucy/internal/cache"
	captchaImage "github.com/oldfritter/lucy/lib/captcha"
	"github.com/oldfritter/lucy/util"
)

type CaptchaText5Image struct {
	dom.Captcha
	Prompt1  string `gorm:"size:1"`
	Prompt2  string `gorm:"size:1"`
	Prompt3  string `gorm:"size:1"`
	Prompt4  string `gorm:"size:1"`
	Prompt5  string `gorm:"size:1"`
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
}

func (*CaptchaText5Image) TableName() string {
	return "captcha_text5image"
}

func (captcha *CaptchaText5Image) Json() string {
	b, _ := json.Marshal(map[string]any{
		"id": fmt.Sprintf("text5-%d", captcha.Id),

		"p1": captcha.Prompt1,
		"p2": captcha.Prompt2,
		"p3": captcha.Prompt3,
		"p4": captcha.Prompt4,
		"p5": captcha.Prompt5,

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
	})
	return string(b)
}

func (captcha *CaptchaText5Image) GetWithPaginate(db *gorm.DB, r *util.Response) {
	var results []*CaptchaText5Image
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

func (captcha *CaptchaText5Image) Create() {
	var texts []image.Image
	texts = append(texts, captchaImage.CreateTextImage(captcha.Prompt1))
	texts = append(texts, captchaImage.CreateTextImage(captcha.Prompt2))
	texts = append(texts, captchaImage.CreateTextImage(captcha.Prompt3))
	texts = append(texts, captchaImage.CreateTextImage(captcha.Prompt4))
	texts = append(texts, captchaImage.CreateTextImage(captcha.Prompt5))

	img := captcha.loadBackground("config/background/c71eda17095e9a92e300ca207f09c778.jpg")
	captchaImage.AddText(img, "请依次点击以下文字：", 0, 30)
	for i, t := range texts {
		captchaImage.AddImage(img, t, 100+30*i, 30, 0)
	}

	cache.SetCaptchaCache(captcha)
}

func (captcha *CaptchaText5Image) Verify(attrs map[string]any) (yes bool) {
	return
}

func (captcha *CaptchaText5Image) loadBackground(inputPath string) image.Image {
	reader, err := os.Open(inputPath)
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()
	m, _, err := image.Decode(reader)
	return m
}
