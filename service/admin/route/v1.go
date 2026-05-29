package route

import (
	"github.com/labstack/echo/v4"

	"github.com/oldfritter/lucy/lib/middleware"
	"github.com/oldfritter/lucy/service/admin/v1"
)

func SetV1Interface(e *echo.Echo) {

	adminGroup := e.Group("/admin", middleware.Auth())

	captchaGroup := adminGroup.Group("/captcha")
	{
		textimage4Group := captchaGroup.Group("/4textimage")
		{
			textimage4Group.GET("/list", v1.GetCaptchaText4List)
			textimage4Group.GET("/:id", v1.GetCaptchaText4)
			textimage4Group.POST("", v1.CreateCaptchaText4)
			textimage4Group.POST("/", v1.CreateCaptchaText4)
			textimage4Group.PUT("/:id", v1.UpdateCaptchaText4)
			textimage4Group.DELETE("/:id", v1.DeleteCaptchaText4)
		}

		textimage5Group := captchaGroup.Group("/5textimage")
		{
			textimage5Group.GET("/list", v1.GetCaptchaText5List)
			textimage5Group.GET("/:id", v1.GetCaptchaText5)
			textimage5Group.POST("", v1.CreateCaptchaText5)
			textimage5Group.POST("/", v1.CreateCaptchaText5)
			textimage5Group.PUT("/:id", v1.UpdateCaptchaText5)
			textimage5Group.DELETE("/:id", v1.DeleteCaptchaText5)
		}

		textimage6Group := captchaGroup.Group("/6textimage")
		{
			textimage6Group.GET("/list", v1.GetCaptchaText6List)
			textimage6Group.GET("/:id", v1.GetCaptchaText6)
			textimage6Group.POST("", v1.CreateCaptchaText6)
			textimage6Group.POST("/", v1.CreateCaptchaText6)
			textimage6Group.PUT("/:id", v1.UpdateCaptchaText6)
			textimage6Group.DELETE("/:id", v1.DeleteCaptchaText6)
		}

		rotateGroup := captchaGroup.Group("/rotate")
		{
			rotateGroup.GET("/list", v1.GetCaptchaRotateImageList)
			rotateGroup.GET("/:id", v1.GetCaptchaRotateImage)
			rotateGroup.POST("", v1.CreateCaptchaRotateImage)
			rotateGroup.POST("/", v1.CreateCaptchaRotateImage)
			rotateGroup.PUT("/:id", v1.UpdateCaptchaRotateImage)
			rotateGroup.DELETE("/:id", v1.DeleteCaptchaRotateImage)
		}
	}

	userGroup := adminGroup.Group("/user")
	{
		userGroup.GET("/list", v1.GetUserList)
		userGroup.GET("/:id", v1.GetUser)
		userGroup.POST("", v1.CreateUser)
		userGroup.POST("/", v1.CreateUser)
		userGroup.PUT("/:id", v1.UpdateUser)
		userGroup.DELETE("/:id", v1.DeleteUser)
	}

}
