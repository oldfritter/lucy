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

	// 获取验证码（ApiKey 认证：X-Api-Key + X-Api-Secret）
	v1Group.GET("/captcha/fetch", v1.FetchCaptcha, middleware.ApiKeyAuth())

	userGroup := v1Group.Group("/user")
	{
		userGroup.POST("/register", v1.Register)
		userGroup.POST("/login", v1.Login)
	}

	// 支付宝 / 微信支付回调通知（无需认证）
	setAlipayRoutes(v1Group)
	setWechatRoutes(v1Group)

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

		campaignGroup := authGroup.Group("/campaign")
		{
			campaignGroup.GET("/list", v1.GetMyCampaignList)
			campaignGroup.GET("/:id", v1.GetMyCampaign)
			campaignGroup.POST("", v1.CreateMyCampaign)
			campaignGroup.POST("/", v1.CreateMyCampaign)
			campaignGroup.PUT("/:id", v1.UpdateMyCampaign)
			campaignGroup.DELETE("/:id", v1.DeleteMyCampaign)
		}

		orderGroup := authGroup.Group("/order")
		{
			orderGroup.POST("", v1.CreateOrder)
			orderGroup.POST("/", v1.CreateOrder)
			orderGroup.GET("/list", v1.GetMyOrderList)
			orderGroup.GET("/:id", v1.GetMyOrder)
		}

		// 支付宝 / 微信支付（需要认证）
		setAlipayAuthRoutes(authGroup)
		setWechatAuthRoutes(authGroup)
	}
}
