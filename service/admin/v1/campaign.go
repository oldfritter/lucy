package v1

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/oldfritter/lucy/lib/db"
	"github.com/oldfritter/lucy/model"
	"github.com/oldfritter/lucy/util"
)

// GetCampaignList 投放列表（分页）
func GetCampaignList(c echo.Context) (err error) {
	var (
		campaign model.Campaign
		body     = util.ArrayResponse()
	)
	if err := c.Bind(&body.Params); err != nil {
		return util.BuildError("1001")
	}
	if err := c.Bind(body.Pagination); err != nil {
		return util.BuildError("1001")
	}
	campaign.GetWithPaginate(db.MysqlDB, &body)
	return c.JSON(http.StatusOK, &body)
}

// GetCampaign 获取单条投放详情
func GetCampaign(c echo.Context) (err error) {
	var campaign model.Campaign
	if err = db.MysqlDB.Where("id = ?", c.Param("id")).
		Preload("User").
		Preload("Product").
		First(&campaign).Error; err != nil {
		return util.BuildError("1003")
	}
	response := util.SuccessResponse()
	response.Body = campaign
	return c.JSON(http.StatusOK, response)
}

// CreateCampaign 创建投放
func CreateCampaign(c echo.Context) (err error) {
	var campaign model.Campaign
	if err = c.Bind(&campaign); err != nil {
		return util.BuildError("1001")
	}
	if err = c.Validate(&campaign); err != nil {
		return util.BuildError("1002")
	}
	tx := db.BeginTx()
	defer tx.DbRollback()
	if tx.Create(&campaign).Error != nil {
		return util.BuildError("1005")
	}
	tx.DbCommit()
	db.MysqlDB.Preload("Product").First(&campaign, campaign.Id)
	response := util.SuccessResponse()
	response.Body = campaign
	return c.JSON(http.StatusOK, response)
}

// UpdateCampaign 更新投放
func UpdateCampaign(c echo.Context) (err error) {
	var campaign model.Campaign
	if err = db.MysqlDB.Where("id = ?", c.Param("id")).
		First(&campaign).Error; err != nil {
		return util.BuildError("1003")
	}
	if err = c.Bind(&campaign); err != nil {
		return util.BuildError("1001")
	}
	if err = c.Validate(&campaign); err != nil {
		return util.BuildError("1002")
	}
	db.MysqlDB.Save(&campaign)
	db.MysqlDB.Preload("Product").First(&campaign, campaign.Id)
	response := util.SuccessResponse()
	response.Body = campaign
	return c.JSON(http.StatusOK, response)
}

// DeleteCampaign 删除投放
func DeleteCampaign(c echo.Context) (err error) {
	var campaign model.Campaign
	if err = db.MysqlDB.Where("id = ?", c.Param("id")).
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
