package v1

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/oldfritter/lucy/internal/cache"
	"github.com/oldfritter/lucy/internal/pool"
	"github.com/oldfritter/lucy/lib/db"
	"github.com/oldfritter/lucy/lib/storage/oss"
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

// VerifyCaptcha 从池中捞取验证码并验证，Uid 作为 captcha 参数传入
func VerifyCaptcha(c echo.Context) (err error) {
	var req verifyRequest
	if err = c.Bind(&req); err != nil {
		return util.BuildError("1001")
	}
	uid := req.Captcha
	if uid == "" {
		return util.BuildError("1001")
	}

	// 从池中捞取，确保只能消费一次
	removed, err := pool.RemoveFromPool(uid)
	if err != nil || !removed {
		return util.BuildError("1008")
	}

	// 根据请求字段判断类型
	if req.Angle != nil {
		return verifyRotateByUid(c, uid, req)
	}
	if len(req.Points) >= 4 && len(req.Points) <= 6 {
		return verifyTextImageByUid(c, uid, req)
	}
	return util.BuildError("1001")
}

func verifyTextImageByUid(c echo.Context, uid string, req verifyRequest) error {
	switch len(req.Points) {
	case 4:
		var captcha model.CaptchaText4
		if err := db.MysqlDB.Where("uid = ?", uid).First(&captcha).Error; err != nil {
			return util.BuildError("1003")
		}
		if !captcha.Verify(map[string]any{"points": pointsToInts(req.Points)}) {
			pool.AddToVerifiedPool("text:4", uid, false)
			return util.BuildError("1008")
		}
		pool.AddToVerifiedPool("text:4", uid, true)
		cleanupCaptcha(captcha.Key, captcha.GetCaptcha())
		resp := util.SuccessResponse()
		resp.Body = map[string]string{"valid_code": captcha.ValidCode}
		return c.JSON(http.StatusOK, resp)
	case 5:
		var captcha model.CaptchaText5
		if err := db.MysqlDB.Where("uid = ?", uid).First(&captcha).Error; err != nil {
			return util.BuildError("1003")
		}
		if !captcha.Verify(map[string]any{"points": pointsToInts(req.Points)}) {
			pool.AddToVerifiedPool("text:5", uid, false)
			return util.BuildError("1008")
		}
		pool.AddToVerifiedPool("text:5", uid, true)
		cleanupCaptcha(captcha.Key, captcha.GetCaptcha())
		resp := util.SuccessResponse()
		resp.Body = map[string]string{"valid_code": captcha.ValidCode}
		return c.JSON(http.StatusOK, resp)
	case 6:
		var captcha model.CaptchaText6
		if err := db.MysqlDB.Where("uid = ?", uid).First(&captcha).Error; err != nil {
			return util.BuildError("1003")
		}
		if !captcha.Verify(map[string]any{"points": pointsToInts(req.Points)}) {
			pool.AddToVerifiedPool("text:6", uid, false)
			return util.BuildError("1008")
		}
		pool.AddToVerifiedPool("text:6", uid, true)
		cleanupCaptcha(captcha.Key, captcha.GetCaptcha())
		resp := util.SuccessResponse()
		resp.Body = map[string]string{"valid_code": captcha.ValidCode}
		return c.JSON(http.StatusOK, resp)
	}
	return util.BuildError("1001")
}

func verifyRotateByUid(c echo.Context, uid string, req verifyRequest) error {
	var captcha model.CaptchaImageRotate
	if err := db.MysqlDB.Where("uid = ?", uid).First(&captcha).Error; err != nil {
		return util.BuildError("1003")
	}
	if !captcha.Verify(map[string]any{"angle": *req.Angle}) {
		pool.AddToVerifiedPool("image:rotate", uid, false)
		return util.BuildError("1008")
	}
	pool.AddToVerifiedPool("image:rotate", uid, true)
	cleanupCaptcha(captcha.Key, captcha.GetCaptcha())
	resp := util.SuccessResponse()
	resp.Body = map[string]string{"valid_code": captcha.ValidCode}
	return c.JSON(http.StatusOK, resp)
}

func pointsToInts(input []verifyPoint) [][]int {
	points := make([][]int, len(input))
	for i, p := range input {
		points[i] = []int{p.X, p.Y}
	}
	return points
}

func cleanupCaptcha(ossKey, captchaId string) {
	_ = oss.DeleteObject(ossKey)
	_ = cache.DelCaptchaCache(captchaId)
}

func cacheIdForType(captchaType, uid string) string {
	return captchaType + ":" + uid
}

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

// FetchCaptcha 使用 ApiKey 中间件认证，从对应类型的池中捞取验证码
func FetchCaptcha(c echo.Context) (err error) {
	apiKey := c.Get("ApiKey").(*cache.ApiKeyCache)

	// 从对应类型的池中捞取一个 uid
	uid, err := pool.PopFromPool(apiKey.CaptchaType)
	if err != nil {
		return util.BuildError("1007", "捞取验证码失败")
	}
	if uid == "" {
		return util.BuildError("1008", "池中无可用验证码")
	}

	// 从缓存获取验证码数据
	cacheId := cacheIdForType(apiKey.CaptchaType, uid)
	cached, err := cache.GetCaptchaCache(cacheId)
	if err != nil || cached == "" {
		pool.AddToPool(apiKey.CaptchaType, uid) // 回放
		return util.BuildError("1003")
	}

	var capData struct {
		Uid       string `json:"uid"`
		ValidCode string `json:"valid_code"`
		Key       string `json:"key"`
	}
	if err = json.Unmarshal([]byte(cached), &capData); err != nil {
		pool.AddToPool(apiKey.CaptchaType, uid)
		return util.BuildError("1003")
	}

	respBody := map[string]string{
		"uid":        capData.Uid,
		"valid_code": capData.ValidCode,
		"key":        oss.OssAsset() + "/" + capData.Key,
	}

	response := util.SuccessResponse()
	response.Body = respBody
	return c.JSON(http.StatusOK, response)
}
