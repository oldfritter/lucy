package v1

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/oldfritter/lucy/lib/db"
	"github.com/oldfritter/lucy/model"
	"github.com/oldfritter/lucy/util"
)

func GetProductList(c echo.Context) (err error) {
	var (
		product model.Product
		body    = util.ArrayResponse()
	)
	if err := c.Bind(&body.Params); err != nil {
		return util.BuildError("1001")
	}
	if err := c.Bind(body.Pagination); err != nil {
		return util.BuildError("1001")
	}
	product.GetWithPaginate(db.MysqlDB, &body)
	return c.JSON(http.StatusOK, &body)
}

func GetProduct(c echo.Context) (err error) {
	var product model.Product
	if err = db.MysqlDB.Where("id = ?", c.Param("id")).
		Preload("Currency").
		First(&product).Error; err != nil {
		return util.BuildError("1003")
	}
	response := util.SuccessResponse()
	response.Body = product
	return c.JSON(http.StatusOK, response)
}

func CreateProduct(c echo.Context) (err error) {
	var product model.Product
	if err = c.Bind(&product); err != nil {
		return util.BuildError("1001")
	}
	if err = c.Validate(&product); err != nil {
		return util.BuildError("1002")
	}
	tx := db.BeginTx()
	defer tx.DbRollback()
	if tx.Create(&product).Error != nil {
		return util.BuildError("1007")
	}
	tx.DbCommit()
	response := util.SuccessResponse()
	response.Body = product
	return c.JSON(http.StatusOK, response)
}

func UpdateProduct(c echo.Context) (err error) {
	var product model.Product
	if err = db.MysqlDB.Where("id = ?", c.Param("id")).
		First(&product).Error; err != nil {
		return util.BuildError("1003")
	}
	if err = c.Bind(&product); err != nil {
		return util.BuildError("1001")
	}
	if err = c.Validate(&product); err != nil {
		return util.BuildError("1002")
	}
	db.MysqlDB.Save(&product)
	response := util.SuccessResponse()
	response.Body = product
	return c.JSON(http.StatusOK, response)
}

func DeleteProduct(c echo.Context) (err error) {
	var product model.Product
	if err = db.MysqlDB.Where("id = ?", c.Param("id")).
		First(&product).Error; err != nil {
		return util.BuildError("1003")
	}
	tx := db.BeginTx()
	defer tx.DbRollback()
	tx.Delete(&product)
	tx.DbCommit()
	response := util.SuccessResponse()
	return c.JSON(http.StatusOK, response)
}
