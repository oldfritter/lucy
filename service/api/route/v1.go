package route

import (
	"github.com/labstack/echo/v4"

	"github.com/oldfritter/lucy/lib/middleware"
	"github.com/oldfritter/lucy/service/api/v1"
)

func SetV1Interface(e *echo.Echo) {
	captchaGroup := e.Group("/api/captcha")
	{
		captchaGroup.POST("/verify", v1.VerifyCaptcha)
	}

	apikeyGroup := e.Group("/api/apikey", middleware.Auth())
	{
		apikeyGroup.GET("/list", v1.GetMyApiKeyList)
		apikeyGroup.GET("/:id", v1.GetMyApiKey)
		apikeyGroup.POST("", v1.CreateMyApiKey)
		apikeyGroup.POST("/", v1.CreateMyApiKey)
		apikeyGroup.PUT("/:id", v1.UpdateMyApiKey)
		apikeyGroup.DELETE("/:id", v1.DeleteMyApiKey)
	}
}
