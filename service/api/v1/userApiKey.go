package v1

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/oldfritter/lucy/base"
	"github.com/oldfritter/lucy/dom"
	"github.com/oldfritter/lucy/internal/cache"
	"github.com/oldfritter/lucy/lib/db"
	"github.com/oldfritter/lucy/model"
	"github.com/oldfritter/lucy/util"
)

// GetMyApiKeyList 当前用户查看自己的 ApiKey 列表
func GetMyApiKeyList(c echo.Context) (err error) {
	claims, _ := base.GetClaim(c)
	var (
		uak  model.UserApiKey
		body = util.ArrayResponse()
	)
	if err := c.Bind(&body.Params); err != nil {
		return util.BuildError("1001")
	}
	if err := c.Bind(body.Pagination); err != nil {
		return util.BuildError("1001")
	}
	body.Params["UserId"] = ""
	// 强制只看自己的
	db.MysqlDB.Model(&uak).
		Where("user_id = ?", claims.UserId).
		Count(&body.Pagination.Count)
	body.Pagination.Init()
	var results []model.UserApiKey
	if err := db.MysqlDB.Where("user_id = ?", claims.UserId).
		Order("id DESC").
		Offset((int(body.Pagination.CurrentPage) - 1) * int(body.Pagination.PerPage)).
		Limit(int(body.Pagination.PerPage)).
		Find(&results).Error; err != nil {
		return util.BuildError("1005")
	}
	for i := range results {
		results[i].Mask()
	}
	body.Body = results
	return c.JSON(http.StatusOK, &body)
}

// GetMyApiKey 当前用户查看自己的单个 ApiKey 详情
func GetMyApiKey(c echo.Context) (err error) {
	claims, _ := base.GetClaim(c)
	var uak model.UserApiKey
	if err = db.MysqlDB.Where("id = ? AND user_id = ?", c.Param("id"), claims.UserId).
		First(&uak).Error; err != nil {
		return util.BuildError("1003")
	}
	response := util.SuccessResponse()
	response.Body = uak
	return c.JSON(http.StatusOK, response)
}

// CreateMyApiKey 当前用户创建自己的 ApiKey（Key、Secret 由系统自动生成）
func CreateMyApiKey(c echo.Context) (err error) {
	claims, _ := base.GetClaim(c)

	// 限制单个用户最多创建 5 个 API Key
	var count int64
	if db.MysqlDB.Model(&model.UserApiKey{}).Where("user_id = ?", claims.UserId).Count(&count); count >= 5 {
		return util.BuildError("1200")
	}

	var input struct {
		ProductId int    `form:"ProductId" validate:"required"`
		Name      string `form:"Name" validate:"max=64"`
	}
	if err = c.Bind(&input); err != nil {
		return util.BuildError("1001")
	}
	if err = c.Validate(&input); err != nil {
		return util.BuildError("1002", err.Error())
	}

	// 验证商品存在，获取 CaptchaType 和 PerMinuteLimit
	var product model.Product
	if err = db.MysqlDB.First(&product, input.ProductId).Error; err != nil {
		return util.BuildError("1502")
	}

	uak := model.UserApiKey{
		UserApiKey: dom.UserApiKey{
			UserId:      claims.UserId,
			ProductId:   input.ProductId,
			CaptchaType: product.CaptchaType,
			Name:        input.Name,
		},
	}
	uak.GenerateKeys()

	tx := db.BeginTx()
	defer tx.DbRollback()
	if tx.Create(&uak).Error != nil {
		return util.BuildError("1005")
	}
	tx.DbCommit()
	response := util.SuccessResponse()
	response.Body = uak
	return c.JSON(http.StatusOK, response)
}

// UpdateMyApiKey 当前用户修改自己的 ApiKey（仅允许 Name、IsActive）
func UpdateMyApiKey(c echo.Context) (err error) {
	claims, _ := base.GetClaim(c)
	var uak model.UserApiKey
	if err = db.MysqlDB.Where("id = ? AND user_id = ?", c.Param("id"), claims.UserId).
		First(&uak).Error; err != nil {
		return util.BuildError("1003")
	}
	var input struct {
		Name     string `form:"Name"`
		IsActive *bool  `form:"IsActive"`
	}
	if err = c.Bind(&input); err != nil {
		return util.BuildError("1001")
	}
	if input.Name != "" {
		uak.Name = input.Name
	}
	if input.IsActive != nil {
		uak.IsActive = *input.IsActive
	}
	if input.Name == "" && input.IsActive == nil {
		return util.BuildError("1001")
	}
	db.MysqlDB.Save(&uak)
	response := util.SuccessResponse()
	response.Body = uak
	return c.JSON(http.StatusOK, response)
}

// ── API Key 分钟级访问统计 ──

// GetMyApiKeyStats 当前用户查看自己 ApiKey 的分钟级调用统计
//
//	查询参数（可选）：
//	  start — 起始时间，格式 "2006-01-02T15:04:05" 或 Unix 秒时间戳（默认 24 小时前）
//	  end   — 截止时间，同上（默认当前时间）
func GetMyApiKeyStats(c echo.Context) (err error) {
	claims, _ := base.GetClaim(c)

	// 验证 ApiKey 属于当前用户
	var uak model.UserApiKey
	if err = db.MysqlDB.Where("id = ? AND user_id = ?", c.Param("id"), claims.UserId).
		First(&uak).Error; err != nil {
		return util.BuildError("1003")
	}

	// 解析时间范围
	now := time.Now()
	endTs := now.Unix()
	startTs := now.Add(-24 * time.Hour).Unix()

	if s := c.QueryParam("start"); s != "" {
		if t, err := parseTimeParam(s); err == nil {
			startTs = t.Unix()
		}
	}
	if e := c.QueryParam("end"); e != "" {
		if t, err := parseTimeParam(e); err == nil {
			endTs = t.Unix()
		}
	}

	// 同时获取总请求统计和验证结果统计（成功/失败）
	totalStats, err := cache.GetApiKeyStats(uak.Id, startTs, endTs)
	if err != nil {
		return util.BuildError("1006")
	}

	verifyStats, err := cache.GetVerifyStats(uak.Id, startTs, endTs)
	if err != nil {
		return util.BuildError("1006")
	}

	// 按分钟构建验证结果 lookup
	verifyMap := make(map[string]*cache.VerifyStatsEntry, len(verifyStats))
	for i := range verifyStats {
		verifyMap[verifyStats[i].Minute] = &verifyStats[i]
	}

	// 合并：每个分钟的总请求数 + 成功/失败数
	type mergedEntry struct {
		Minute  string `json:"minute"`
		Count   int    `json:"count"`
		Success int    `json:"success"`
		Failed  int    `json:"failed"`
	}
	result := make([]mergedEntry, 0, len(totalStats))
	for _, ts := range totalStats {
		entry := mergedEntry{
			Minute:  ts.Minute,
			Count:   ts.Count,
			Success: 0,
			Failed:  0,
		}
		if vs, ok := verifyMap[ts.Minute]; ok {
			entry.Success = vs.Success
			entry.Failed = vs.Failed
		}
		result = append(result, entry)
	}

	resp := util.SuccessResponse()
	resp.Body = result
	return c.JSON(http.StatusOK, resp)
}

// GetMyApiKeyVerifyStats 当前用户查看自己 ApiKey 的分钟级验证结果（成功/失败）统计
//
//	查询参数（可选）：
//	  start — 起始时间，格式 "2006-01-02T15:04:05" 或 Unix 秒时间戳（默认 24 小时前）
//	  end   — 截止时间，同上（默认当前时间）
func GetMyApiKeyVerifyStats(c echo.Context) (err error) {
	claims, _ := base.GetClaim(c)

	var uak model.UserApiKey
	if err = db.MysqlDB.Where("id = ? AND user_id = ?", c.Param("id"), claims.UserId).
		First(&uak).Error; err != nil {
		return util.BuildError("1003")
	}

	now := time.Now()
	endTs := now.Unix()
	startTs := now.Add(-24 * time.Hour).Unix()

	if s := c.QueryParam("start"); s != "" {
		if t, err := parseTimeParam(s); err == nil {
			startTs = t.Unix()
		}
	}
	if e := c.QueryParam("end"); e != "" {
		if t, err := parseTimeParam(e); err == nil {
			endTs = t.Unix()
		}
	}

	body, err := cache.GetVerifyStats(uak.Id, startTs, endTs)
	if err != nil {
		return util.BuildError("1006")
	}

	resp := util.SuccessResponse()
	resp.Body = body
	return c.JSON(http.StatusOK, resp)
}

// parseTimeParam 尝试将字符串按多种格式解析为 time.Time
func parseTimeParam(s string) (time.Time, error) {
	if t, err := time.Parse("2006-01-02T15:04:05", s); err == nil {
		return t, nil
	}
	if t, err := strconv.ParseInt(s, 10, 64); err == nil {
		return time.Unix(t, 0), nil
	}
	return time.Time{}, fmt.Errorf("无法解析时间: %s", s)
}

// DeleteMyApiKey 当前用户删除自己的 ApiKey
func DeleteMyApiKey(c echo.Context) (err error) {
	claims, _ := base.GetClaim(c)
	var uak model.UserApiKey
	if err = db.MysqlDB.Where("id = ? AND user_id = ?", c.Param("id"), claims.UserId).
		First(&uak).Error; err != nil {
		return util.BuildError("1003")
	}
	tx := db.BeginTx()
	defer tx.DbRollback()
	tx.Delete(&uak)
	tx.DbCommit()
	response := util.SuccessResponse()
	return c.JSON(http.StatusOK, response)
}
