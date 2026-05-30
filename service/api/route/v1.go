package route

import (
	"github.com/labstack/echo/v4"

	"github.com/oldfritter/lucy/lib/middleware"
	"github.com/oldfritter/lucy/service/api/v1"
)

func SetV1Interface(e *echo.Echo) {
	v1Group := e.Group("/api/v1")

	captchaGroup := v1Group.Group("/captcha")
	{
		captchaGroup.POST("/verify", v1.VerifyCaptcha)
	}

	userGroup := v1Group.Group("/user")
	{
		userGroup.POST("/register", v1.Register)
		userGroup.POST("/login", v1.Login)
	}

	authGroup := v1Group.Group("", middleware.Auth())
	{
		userGroup := authGroup.Group("/user")
		{
			userGroup.GET("/profile", v1.GetMyProfile)
			userGroup.PUT("/profile", v1.UpdateMyProfile)
		}

		apikeyGroup := authGroup.Group("/apikey")
		{
			apikeyGroup.GET("/list", v1.GetMyApiKeyList)
			apikeyGroup.GET("/:id", v1.GetMyApiKey)
			apikeyGroup.POST("", v1.CreateMyApiKey)
			apikeyGroup.POST("/", v1.CreateMyApiKey)
			apikeyGroup.PUT("/:id", v1.UpdateMyApiKey)
			apikeyGroup.DELETE("/:id", v1.DeleteMyApiKey)
		}

		imageGroup := authGroup.Group("/image")
		{
			imageGroup.POST("/upload", v1.UploadImage)
			imageGroup.GET("/list", v1.GetMyImageList)
			imageGroup.GET("/:id", v1.GetMyImage)
			imageGroup.DELETE("/:id", v1.DeleteMyImage)
		}
	}
}
