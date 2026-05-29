package v1

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/oldfritter/lucy/lib/db"
	"github.com/oldfritter/lucy/model"
	"github.com/oldfritter/lucy/util"
)

func GetUserList(c echo.Context) (err error) {
	var (
		user model.User
		body = util.ArrayResponse()
	)
	if err := c.Bind(&body.Params); err != nil {
		return util.BuildError("1001")
	}
	if err := c.Bind(body.Pagination); err != nil {
		return util.BuildError("1001")
	}
	user.GetWithPaginate(db.MysqlDB, &body)
	return c.JSON(http.StatusOK, &body)
}

func GetUser(c echo.Context) (err error) {
	var user model.User
	if err = db.MysqlDB.Where("id = ?", c.Param("id")).
		First(&user).Error; err != nil {
		return util.BuildError("1003")
	}
	response := util.SuccessResponse()
	response.Body = user
	return c.JSON(http.StatusOK, response)
}

func CreateUser(c echo.Context) (err error) {
	var user model.User
	if err = c.Bind(&user); err != nil {
		return util.BuildError("1001")
	}
	if err = c.Validate(&user); err != nil {
		return util.BuildError("1002")
	}
	tx := db.BeginTx()
	defer tx.DbRollback()
	if tx.Save(&user).Error != nil {
		return util.BuildError("1007")
	}
	tx.DbCommit()
	response := util.SuccessResponse()
	response.Body = user
	return c.JSON(http.StatusOK, response)
}

func UpdateUser(c echo.Context) (err error) {
	var user model.User
	if err = db.MysqlDB.Where("id = ?", c.Param("id")).
		First(&user).Error; err != nil {
		return util.BuildError("1003")
	}
	if err = c.Bind(&user); err != nil {
		return util.BuildError("1001")
	}
	if err = c.Validate(&user); err != nil {
		return util.BuildError("1002")
	}
	db.MysqlDB.Save(&user)
	response := util.SuccessResponse()
	response.Body = user
	return c.JSON(http.StatusOK, response)
}

func DeleteUser(c echo.Context) (err error) {
	var user model.User
	if err = db.MysqlDB.Where("id = ?", c.Param("id")).
		First(&user).Error; err != nil {
		return util.BuildError("1003")
	}
	tx := db.BeginTx()
	defer tx.DbRollback()
	tx.Delete(&user)
	tx.DbCommit()
	response := util.SuccessResponse()
	return c.JSON(http.StatusOK, response)
}
