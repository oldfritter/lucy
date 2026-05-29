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
			textimage4Group.GET("/list", v1.GetCaptchaTextimage4List)
			textimage4Group.GET("/:id", v1.GetCaptchaTextimage4)
			textimage4Group.POST("", v1.CreateCaptchaTextimage4)
			textimage4Group.POST("/", v1.CreateCaptchaTextimage4)
			textimage4Group.PUT("/:id", v1.UpdateCaptchaTextimage4)
			textimage4Group.DELETE("/:id", v1.DeleteCaptchaTextimage4)
		}

		textimage5Group := captchaGroup.Group("/5textimage")
		{
			textimage5Group.GET("/list", v1.GetCaptchaTextimage5List)
			textimage5Group.GET("/:id", v1.GetCaptchaTextimage5)
			textimage5Group.POST("", v1.CreateCaptchaTextimage5)
			textimage5Group.POST("/", v1.CreateCaptchaTextimage5)
			textimage5Group.PUT("/:id", v1.UpdateCaptchaTextimage5)
			textimage5Group.DELETE("/:id", v1.DeleteCaptchaTextimage5)
		}

		textimage6Group := captchaGroup.Group("/6textimage")
		{
			textimage6Group.GET("/list", v1.GetCaptchaTextimage6List)
			textimage6Group.GET("/:id", v1.GetCaptchaTextimage6)
			textimage6Group.POST("", v1.CreateCaptchaTextimage6)
			textimage6Group.POST("/", v1.CreateCaptchaTextimage6)
			textimage6Group.PUT("/:id", v1.UpdateCaptchaTextimage6)
			textimage6Group.DELETE("/:id", v1.DeleteCaptchaTextimage6)
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
