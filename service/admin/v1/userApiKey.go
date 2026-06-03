package v1

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/oldfritter/lucy/dom"
	"github.com/oldfritter/lucy/lib/db"
	"github.com/oldfritter/lucy/model"
	"github.com/oldfritter/lucy/util"
)

// GetUserApiKeyList 获取用户的 ApiKey 列表（分页）
func GetUserApiKeyList(c echo.Context) (err error) {
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
	// 从路径参数中获取 UserId 优先
	if userId := c.Param("userId"); userId != "" {
		body.Params["UserId"] = userId
	}
	uak.GetWithPaginate(db.MysqlDB, &body)
	return c.JSON(http.StatusOK, &body)
}

// GetUserApiKey 获取单个 ApiKey 详情
func GetUserApiKey(c echo.Context) (err error) {
	var uak model.UserApiKey
	if err = db.MysqlDB.Where("id = ?", c.Param("id")).
		Preload("User").
		First(&uak).Error; err != nil {
		return util.BuildError("1003")
	}
	response := util.SuccessResponse()
	response.Body = uak
	return c.JSON(http.StatusOK, response)
}

// CreateUserApiKey 创建 ApiKey（Key、Secret 由系统自动生成）
func CreateUserApiKey(c echo.Context) (err error) {
	var input struct {
		UserId    int    `form:"UserId" validate:"required"`
		ProductId int    `form:"ProductId" validate:"required"`
		Name      string `form:"Name" validate:"max=64"`
	}
	if err = c.Bind(&input); err != nil {
		return util.BuildError("1001")
	}
	if err = c.Validate(&input); err != nil {
		return util.BuildError("1002")
	}

	// 验证商品存在
	var product model.Product
	if err = db.MysqlDB.First(&product, input.ProductId).Error; err != nil {
		return util.BuildError("1003", "商品不存在")
	}

	uak := model.UserApiKey{
		UserApiKey: dom.UserApiKey{
			UserId:      input.UserId,
			ProductId:   input.ProductId,
			CaptchaType: product.CaptchaType,
			Name:        input.Name,
		},
		Product: &product,
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

// UpdateUserApiKey 更新 ApiKey
func UpdateUserApiKey(c echo.Context) (err error) {
	var uak model.UserApiKey
	if err = db.MysqlDB.Where("id = ?", c.Param("id")).
		First(&uak).Error; err != nil {
		return util.BuildError("1003")
	}
	if err = c.Bind(&uak); err != nil {
		return util.BuildError("1001")
	}
	if err = c.Validate(&uak); err != nil {
		return util.BuildError("1002")
	}
	db.MysqlDB.Save(&uak)
	response := util.SuccessResponse()
	response.Body = uak
	return c.JSON(http.StatusOK, response)
}

// DeleteUserApiKey 删除 ApiKey
func DeleteUserApiKey(c echo.Context) (err error) {
	var uak model.UserApiKey
	if err = db.MysqlDB.Where("id = ?", c.Param("id")).
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
