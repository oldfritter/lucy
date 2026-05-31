package middleware

import (
	"github.com/labstack/echo/v4"

	"github.com/oldfritter/lucy/internal/cache"
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

			cached, err := cache.GetApiKeyCache(key)
			if err != nil {
				return util.BuildError("1003")
			}
			if cached.Secret != secret {
				return util.BuildError("1005")
			}
			if !cached.IsActive {
				return util.BuildError("1005")
			}

			c.Set("ApiKey", &cached)
			return next(c)
		}
	}
}
