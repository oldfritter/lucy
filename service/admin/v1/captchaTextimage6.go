package v1

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/oldfritter/lucy/lib/db"
	"github.com/oldfritter/lucy/model"
	"github.com/oldfritter/lucy/util"
)

func GetCaptchaText6List(c echo.Context) (err error) {
	var (
		captcha model.CaptchaText6
		body    = util.ArrayResponse()
	)
	if err := c.Bind(&body.Params); err != nil {
		return util.BuildError("1001")
	}
	if err := c.Bind(body.Pagination); err != nil {
		return util.BuildError("1001")
	}
	captcha.GetWithPaginate(db.MysqlDB, &body)
	return c.JSON(http.StatusOK, &body)
}

func GetCaptchaText6(c echo.Context) (err error) {
	var captcha model.CaptchaText6
	if err = db.MysqlDB.Where("id = ?", c.Param("id")).
		First(&captcha).Error; err != nil {
		return util.BuildError("1003")
	}
	response := util.SuccessResponse()
	response.Body = captcha
	return c.JSON(http.StatusOK, response)
}

func CreateCaptchaText6(c echo.Context) (err error) {
	var captcha model.CaptchaText6
	if err = c.Bind(&captcha); err != nil {
		return util.BuildError("1001")
	}
	if err = c.Validate(&captcha); err != nil {
		return util.BuildError("1002")
	}
	tx := db.BeginTx()
	defer tx.DbRollback()
	if tx.Save(&captcha).Error != nil {
		return util.BuildError("1007")
	}
	tx.DbCommit()
	response := util.SuccessResponse()
	response.Body = captcha
	return c.JSON(http.StatusOK, response)
}

func UpdateCaptchaText6(c echo.Context) (err error) {
	var captcha model.CaptchaText6
	if err = db.MysqlDB.Where("id = ?", c.Param("id")).
		First(&captcha).Error; err != nil {
		return util.BuildError("1003")
	}
	if err = c.Bind(&captcha); err != nil {
		return util.BuildError("1001")
	}
	if err = c.Validate(&captcha); err != nil {
		return util.BuildError("1002")
	}
	db.MysqlDB.Save(&captcha)
	response := util.SuccessResponse()
	response.Body = captcha
	return c.JSON(http.StatusOK, response)
}

func DeleteCaptchaText6(c echo.Context) (err error) {
	var captcha model.CaptchaText6
	if err = db.MysqlDB.Where("id = ?", c.Param("id")).
		First(&captcha).Error; err != nil {
		return util.BuildError("1003")
	}
	tx := db.BeginTx()
	defer tx.DbRollback()
	tx.Delete(&captcha)
	tx.DbCommit()
	response := util.SuccessResponse()
	return c.JSON(http.StatusOK, response)
}
