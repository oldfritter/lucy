package middleware

import (
	"github.com/labstack/echo/v4"

	"github.com/oldfritter/lucy/lib/db"
	"github.com/oldfritter/lucy/model"
	"github.com/oldfritter/lucy/util"
)

func ApiKeyAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			key := c.Request().Header.Get("X-Api-Key")
			secret := c.Request().Header.Get("X-Api-Secret")
			if key == "" || secret == "" {
				return util.BuildError("1005")
			}
			var uak model.UserApiKey
			if err := db.MysqlDB.Where("key = ?", key).First(&uak).Error; err != nil {
				return util.BuildError("1003")
			}
			if uak.Secret != secret {
				return util.BuildError("1005")
			}
			if !uak.IsActive {
				return util.BuildError("1005")
			}
			c.Set("ApiKey", &uak)
			return next(c)
		}
	}
}
