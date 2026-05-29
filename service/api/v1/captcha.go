package v1

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/oldfritter/lucy/internal/cache"
	"github.com/oldfritter/lucy/lib/db"
	"github.com/oldfritter/lucy/model"
	"github.com/oldfritter/lucy/util"
)

type verifyPoint struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type verifyRequest struct {
	Captcha string        `json:"captcha"`
	Points  []verifyPoint `json:"points,omitempty"`
	Angle   *float64      `json:"angle,omitempty"`
}

// VerifyCaptcha 统一验证入口：优先从 Redis 缓存判断类型，回退到请求字段推断
func VerifyCaptcha(c echo.Context) (err error) {
	var req verifyRequest
	if err = c.Bind(&req); err != nil {
		return util.BuildError("1001")
	}
	if req.Captcha == "" {
		return util.BuildError("1001")
	}

	// 从 Redis 缓存获取验证码类型
	captchaType := lookupCaptchaType(req.Captcha)

	switch captchaType {
	case "rotate":
		return verifyRotate(c, req)
	case "text":
		return verifyTextImage(c, req)
	default:
		// 缓存未命中，按请求字段回退判断
		if req.Angle != nil {
			return verifyRotate(c, req)
		}
		if len(req.Points) >= 4 && len(req.Points) <= 6 {
			return verifyTextImage(c, req)
		}
		return util.BuildError("1001")
	}
}

// lookupCaptchaType 从 Redis 缓存中解析验证码类型（"text" | "rotate" | ""）
func lookupCaptchaType(captcha string) string {
	cached, err := cache.GetCaptchaCache(captcha)
	if err != nil || cached == "" {
		return ""
	}
	var wrapper struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal([]byte(cached), &wrapper); err != nil {
		return ""
	}
	switch {
	case strings.HasPrefix(wrapper.ID, "text"):
		return "text"
	case strings.HasPrefix(wrapper.ID, "rotate"):
		return "rotate"
	}
	return ""
}

func verifyTextImage(c echo.Context, req verifyRequest) error {
	switch len(req.Points) {
	case 4:
		var captcha model.CaptchaText4
		if err := db.MysqlDB.Where("captcha = ?", req.Captcha).First(&captcha).Error; err != nil {
			return util.BuildError("1003")
		}
		if !captcha.Verify(map[string]any{"points": pointsToInts(req.Points)}) {
			return util.BuildError("1008")
		}
	case 5:
		var captcha model.CaptchaText5
		if err := db.MysqlDB.Where("captcha = ?", req.Captcha).First(&captcha).Error; err != nil {
			return util.BuildError("1003")
		}
		if !captcha.Verify(map[string]any{"points": pointsToInts(req.Points)}) {
			return util.BuildError("1008")
		}
	case 6:
		var captcha model.CaptchaText6
		if err := db.MysqlDB.Where("captcha = ?", req.Captcha).First(&captcha).Error; err != nil {
			return util.BuildError("1003")
		}
		if !captcha.Verify(map[string]any{"points": pointsToInts(req.Points)}) {
			return util.BuildError("1008")
		}
	}
	return c.JSON(http.StatusOK, util.SuccessResponse())
}

func verifyRotate(c echo.Context, req verifyRequest) error {
	var captcha model.CaptchaRotateImage
	if err := db.MysqlDB.Where("captcha = ?", req.Captcha).First(&captcha).Error; err != nil {
		return util.BuildError("1003")
	}
	if !captcha.Verify(map[string]any{"angle": *req.Angle}) {
		return util.BuildError("1008")
	}
	return c.JSON(http.StatusOK, util.SuccessResponse())
}

func pointsToInts(input []verifyPoint) [][]int {
	points := make([][]int, len(input))
	for i, p := range input {
		points[i] = []int{p.X, p.Y}
	}
	return points
}
