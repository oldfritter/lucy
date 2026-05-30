package v1

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/oldfritter/lucy/base"
	"github.com/oldfritter/lucy/dom"
	"github.com/oldfritter/lucy/lib/db"
	"github.com/oldfritter/lucy/model"
	"github.com/oldfritter/lucy/util"
)

// GetMyCampaignList 当前用户的投放列表（分页）
func GetMyCampaignList(c echo.Context) (err error) {
	claims, _ := base.GetClaim(c)
	var campaign model.Campaign
	body := util.ArrayResponse()
	if err = c.Bind(&body.Params); err != nil {
		return util.BuildError("1001")
	}
	if err = c.Bind(body.Pagination); err != nil {
		return util.BuildError("1001")
	}
	// 强制限定只查当前用户
	body.Params["UserId"] = strconv.Itoa(claims.UserId)
	campaign.GetWithPaginate(db.MysqlDB, &body)
	return c.JSON(http.StatusOK, &body)
}

// GetMyCampaign 获取当前用户的单条投放详情
func GetMyCampaign(c echo.Context) (err error) {
	claims, _ := base.GetClaim(c)
	var campaign model.Campaign
	if err = db.MysqlDB.Where("id = ? AND user_id = ?", c.Param("id"), claims.UserId).
		First(&campaign).Error; err != nil {
		return util.BuildError("1003")
	}
	response := util.SuccessResponse()
	response.Body = campaign
	return c.JSON(http.StatusOK, response)
}

// CreateMyCampaign 创建投放（Status 由系统控制，固定为 0-待处理）
func CreateMyCampaign(c echo.Context) (err error) {
	claims, _ := base.GetClaim(c)
	var campaign model.Campaign
	if err = c.Bind(&campaign); err != nil {
		return util.BuildError("1001")
	}

	// 系统控制：UserId 取自当前登录用户，Status 固定为待处理
	campaign.UserId = claims.UserId
	campaign.Status = dom.StatusPending

	if err = c.Validate(&campaign); err != nil {
		return util.BuildError("1002", err.Error())
	}

	tx := db.BeginTx()
	defer tx.DbRollback()
	if tx.Create(&campaign).Error != nil {
		return util.BuildError("1007")
	}
	tx.DbCommit()
	response := util.SuccessResponse()
	response.Body = campaign
	return c.JSON(http.StatusOK, response)
}

// UpdateMyCampaign 更新投放（Status 字段禁止用户修改）
func UpdateMyCampaign(c echo.Context) (err error) {
	claims, _ := base.GetClaim(c)
	var campaign model.Campaign
	if err = db.MysqlDB.Where("id = ? AND user_id = ?", c.Param("id"), claims.UserId).
		First(&campaign).Error; err != nil {
		return util.BuildError("1003")
	}

	// 只绑定用户可修改的字段，不绑定 Status
	var input struct {
		Name              string `form:"Name" validate:"max=64"`
		CaptchaType       string `form:"CaptchaType"`
		BackgroundImages  string `form:"BackgroundImages"`
		WordBank          string `form:"WordBank"`
		UseSystemWordBank bool   `form:"UseSystemWordBank"`
		CaptchaCount      int    `form:"CaptchaCount" validate:"omitempty,min=1"`
	}
	if err = c.Bind(&input); err != nil {
		return util.BuildError("1001")
	}

	updates := map[string]any{}
	if input.Name != "" {
		updates["name"] = input.Name
	}
	if input.CaptchaType != "" {
		updates["captcha_type"] = input.CaptchaType
	}
	if input.BackgroundImages != "" {
		updates["background_images"] = input.BackgroundImages
	}
	if input.WordBank != "" {
		updates["word_bank"] = input.WordBank
	}
	// UseSystemWordBank 是 bool，允许 false 值，用 c.FormValue 直接读取
	if c.FormValue("UseSystemWordBank") != "" {
		updates["use_system_word_bank"] = input.UseSystemWordBank
	}
	if input.CaptchaCount > 0 {
		updates["captcha_count"] = input.CaptchaCount
	}
	if len(updates) > 0 {
		db.MysqlDB.Model(&campaign).Updates(updates)
		db.MysqlDB.First(&campaign, campaign.Id)
	}

	response := util.SuccessResponse()
	response.Body = campaign
	return c.JSON(http.StatusOK, response)
}

// DeleteMyCampaign 删除投放
func DeleteMyCampaign(c echo.Context) (err error) {
	claims, _ := base.GetClaim(c)
	var campaign model.Campaign
	if err = db.MysqlDB.Where("id = ? AND user_id = ?", c.Param("id"), claims.UserId).
		First(&campaign).Error; err != nil {
		return util.BuildError("1003")
	}
	tx := db.BeginTx()
	defer tx.DbRollback()
	tx.Delete(&campaign)
	tx.DbCommit()
	response := util.SuccessResponse()
	return c.JSON(http.StatusOK, response)
}
