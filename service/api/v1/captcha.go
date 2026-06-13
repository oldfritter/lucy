package v1

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/oldfritter/lucy/dom"
	"github.com/oldfritter/lucy/internal/cache"
	"github.com/oldfritter/lucy/internal/pool"
	"github.com/oldfritter/lucy/lib/db"
	"github.com/oldfritter/lucy/lib/storage/oss"
	"github.com/oldfritter/lucy/model"
	"github.com/oldfritter/lucy/util"
)

type verifyPoint struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
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

	// // 从池中捞取，确保只能消费一次
	// removed, err := pool.RemoveFromPool(uid)
	// if err != nil || !removed {
	//   return util.BuildError("1008")
	// }

	// 根据请求字段判断类型
	if req.Angle != nil {
		return verifyRotateByUid(c, uid, req)
	}
	if len(req.Points) >= 4 && len(req.Points) <= 6 {
		return verifyTextByUid(c, uid, req)
	}
	return util.BuildError("1001")
}

func verifyTextByUid(c echo.Context, uid string, req verifyRequest) error {
	switch len(req.Points) {
	case 4:
		var captcha model.CaptchaText4
		if err := db.MysqlDB.Where("uid = ?", uid).First(&captcha).Error; err != nil {
			return util.BuildError("1301")
		}
		if ok, _ := pool.IsInVerifiedPool(uid); ok {
			return util.BuildError("1008")
		}
		if !captcha.Verify(map[string]any{"points": pointsToInts(req.Points)}) {
			pool.AddToVerifiedPool("text:4", uid, false)
			pool.RemoveFromPendingPool(uid)
			recordVerifyResult(captcha.UserApiKeyId, false)
			return util.BuildError("1300")
		}
		captcha.Status = dom.CaptchaStatusSuccess
		cache.SetCaptchaCache(&captcha)
		pool.AddToVerifiedPool("text:4", uid, true)
		pool.ResetRecallCount(uid)
		pool.RemoveFromPendingPool(uid)
		cleanupCaptcha(captcha.Key, captcha.GetCaptcha())
		recordVerifyResult(captcha.UserApiKeyId, true)
		resp := util.SuccessResponse()
		resp.Body = map[string]string{"valid_code": captcha.ValidCode}
		return c.JSON(http.StatusOK, resp)
	case 5:
		var captcha model.CaptchaText5
		if err := db.MysqlDB.Where("uid = ?", uid).First(&captcha).Error; err != nil {
			return util.BuildError("1301")
		}
		if ok, _ := pool.IsInVerifiedPool(uid); ok {
			return util.BuildError("1008")
		}
		if !captcha.Verify(map[string]any{"points": pointsToInts(req.Points)}) {
			pool.AddToVerifiedPool("text:5", uid, false)
			pool.RemoveFromPendingPool(uid)
			recordVerifyResult(captcha.UserApiKeyId, false)
			return util.BuildError("1300")
		}
		captcha.Status = dom.CaptchaStatusSuccess
		cache.SetCaptchaCache(&captcha)
		pool.AddToVerifiedPool("text:5", uid, true)
		pool.ResetRecallCount(uid)
		pool.RemoveFromPendingPool(uid)
		cleanupCaptcha(captcha.Key, captcha.GetCaptcha())
		recordVerifyResult(captcha.UserApiKeyId, true)
		resp := util.SuccessResponse()
		resp.Body = map[string]string{"valid_code": captcha.ValidCode}
		return c.JSON(http.StatusOK, resp)
	case 6:
		var captcha model.CaptchaText6
		if err := db.MysqlDB.Where("uid = ?", uid).First(&captcha).Error; err != nil {
			return util.BuildError("1301")
		}
		if ok, _ := pool.IsInVerifiedPool(uid); ok {
			return util.BuildError("1008")
		}
		if !captcha.Verify(map[string]any{"points": pointsToInts(req.Points)}) {
			pool.AddToVerifiedPool("text:6", uid, false)
			pool.RemoveFromPendingPool(uid)
			recordVerifyResult(captcha.UserApiKeyId, false)
			return util.BuildError("1300")
		}
		captcha.Status = dom.CaptchaStatusSuccess
		cache.SetCaptchaCache(&captcha)
		pool.AddToVerifiedPool("text:6", uid, true)
		pool.ResetRecallCount(uid)
		pool.RemoveFromPendingPool(uid)
		cleanupCaptcha(captcha.Key, captcha.GetCaptcha())
		recordVerifyResult(captcha.UserApiKeyId, true)
		resp := util.SuccessResponse()
		resp.Body = map[string]string{"valid_code": captcha.ValidCode}
		return c.JSON(http.StatusOK, resp)
	}
	return util.BuildError("1001")
}

func verifyRotateByUid(c echo.Context, uid string, req verifyRequest) error {
	var captcha model.CaptchaImageRotate
	if err := db.MysqlDB.Where("uid = ?", uid).First(&captcha).Error; err != nil {
		return util.BuildError("1301")
	}
	if ok, _ := pool.IsInVerifiedPool(uid); ok {
		return util.BuildError("1008")
	}
	if !captcha.Verify(map[string]any{"angle": *req.Angle}) {
		pool.AddToVerifiedPool("image:rotate", uid, false)
		pool.RemoveFromPendingPool(uid)
		recordVerifyResult(captcha.UserApiKeyId, false)
		return util.BuildError("1300")
	}
	captcha.Status = dom.CaptchaStatusSuccess
	cache.SetCaptchaCache(&captcha)
	pool.AddToVerifiedPool("image:rotate", uid, true)
	pool.ResetRecallCount(uid)
	pool.RemoveFromPendingPool(uid)
	cleanupCaptcha(captcha.Key, captcha.GetCaptcha())
	recordVerifyResult(captcha.UserApiKeyId, true)
	resp := util.SuccessResponse()
	resp.Body = map[string]string{"valid_code": captcha.ValidCode}
	return c.JSON(http.StatusOK, resp)
}

func recordVerifyResult(captchaUserApiKeyId *int, isSuccess bool) {
	if captchaUserApiKeyId != nil {
		if err := cache.RecordVerifyResult(*captchaUserApiKeyId, isSuccess); err != nil {
			log.Printf("[verify] record verify result (apiKeyId=%d, success=%v) failed: %v", *captchaUserApiKeyId, isSuccess, err)
		}
	}
}

func pointsToInts(input []verifyPoint) [][]int {
	points := make([][]int, len(input))
	for i, p := range input {
		points[i] = []int{int(math.Round(p.X)), int(math.Round(p.Y))}
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

// FetchCaptcha 使用 ApiKey 中间件认证，从对应类型的池中捞取验证码
func FetchCaptcha(c echo.Context) (err error) {
	apiKey := c.Get("ApiKey").(*cache.ApiKeyCache)

	// 每分钟限速检查
	allowed, err := cache.CheckRateLimit(apiKey.Secret, apiKey.PerMinuteLimit)
	if err != nil {
		return util.BuildError("1006")
	}
	if !allowed {
		return c.JSON(http.StatusTooManyRequests, util.BuildError("1303"))
	}

	// 记录本次请求到分钟级访问统计
	if err := cache.RecordApiKeyRequest(apiKey.Id); err != nil {
		log.Printf("[fetch-captcha] record api key %d request failed: %v", apiKey.Id, err)
	}

	// 从对应类型的池中捞取一个 uid
	uid, err := pool.PopFromPool(apiKey.CaptchaType)
	if err != nil {
		return util.BuildError("1006")
	}
	if uid == "" {
		return util.BuildError("1302")
	}

	// 从缓存获取验证码数据
	cacheId := cacheIdForType(apiKey.CaptchaType, uid)
	cached, err := cache.GetCaptchaCache(cacheId)
	if err != nil || cached == "" {
		pool.AddToPool(apiKey.CaptchaType, uid) // 回放
		return util.BuildError("1301")
	}

	var capData struct {
		ValidCode string `json:"valid_code"`
		Key       string `json:"key"`
	}
	if err = json.Unmarshal([]byte(cached), &capData); err != nil {
		pool.AddToPool(apiKey.CaptchaType, uid)
		return util.BuildError("1301")
	}

	// 解析文字内容（文字类验证码才包含 p1-p6 字段）
	var rawData map[string]any
	json.Unmarshal([]byte(cached), &rawData)
	var texts []string
	for i := 1; i <= 6; i++ {
		key := fmt.Sprintf("p%d", i)
		if p, ok := rawData[key].(string); ok && p != "" {
			texts = append(texts, p)
		} else {
			break
		}
	}

	// 如果验证码属于某个 system campaign，从 campaign 待用池中移除
	if campaignID, ok := rawData["campaign_id"].(float64); ok {
		pool.RemoveFromCampaignPool(int(campaignID), uid)
	}

	// 将 uid 记录到待验证池，超时未验证将由定时任务回收
	pool.AddToPendingPool(apiKey.CaptchaType, uid)

	// 将消费此验证码的 ApiKey ID 暂存 Redis，由 batchConfirmCaptchaSuccess 批量回写 DB
	if err := cache.SetFetchOwner(uid, apiKey.Id); err != nil {
		log.Printf("[fetch-captcha] set fetch owner for %s failed: %v", uid, err)
	}

	// 生成带签名的临时下载链接
	imageUrl, err := oss.GetObjectURL(capData.Key, 300)
	if err != nil {
		pool.AddToPool(apiKey.CaptchaType, uid)
		return util.BuildError("1505")
	}

	respBody := map[string]any{
		"uid":        uid,
		"valid_code": capData.ValidCode,
		"url":        imageUrl,
	}
	// 图片尺寸（用于前端坐标换算）
	if w, ok := rawData["width"].(float64); ok {
		respBody["width"] = int(w)
	}
	if h, ok := rawData["height"].(float64); ok {
		respBody["height"] = int(h)
	}
	if len(texts) > 0 {
		respBody["texts"] = texts
	}

	response := util.SuccessResponse()
	response.Body = respBody
	return c.JSON(http.StatusOK, response)
}
