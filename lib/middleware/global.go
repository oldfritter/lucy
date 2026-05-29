package middleware

import (
	"strings"

	"github.com/labstack/echo/v4"
)

func Language() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			lang := []string{"zh-CN"}
			if len(c.Request().Header["Accept-Language"]) > 0 {
				if c.Request().Header["Accept-Language"][0] != "" {
					lang = strings.Split(c.Request().Header["Accept-Language"][0], ",")
				}
			} else if len(c.Request().Header["Language"]) > 0 {
				if c.Request().Header["Language"][0] != "" {
					lang = strings.Split(c.Request().Header["Language"][0], ",")
				}
			}
			if lang[0] == "zh-CN" || lang[0] == "zh-TW" || lang[0] == "en-US" {
				c.Set("Language", lang[0])
			} else {
				c.Set("Language", "zh-CN")
			}
			return next(c)
		}
	}
}
