package v1

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/oldfritter/lucy/base"
	"github.com/oldfritter/lucy/dom"
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
		return util.BuildError("1000")
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
		return util.BuildError("1009")
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
		return util.BuildError("1003", "商品不存在")
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
		return util.BuildError("1007")
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
