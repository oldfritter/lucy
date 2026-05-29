package route

import (
	"github.com/labstack/echo/v4"

	"github.com/oldfritter/lucy/service/api/v1"
)

func SetV1Interface(e *echo.Echo) {
	captchaGroup := e.Group("/api/captcha")
	{
		captchaGroup.POST("/verify", v1.VerifyCaptcha)
	}
}
