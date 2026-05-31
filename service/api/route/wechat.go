package route

import (
	"github.com/labstack/echo/v4"

	"github.com/oldfritter/lucy/service/api/wechat"
)

func setWechatRoutes(v1Group *echo.Group) {
	v1Group.POST("/wechat/notify", wechat.PayNotify)
	v1Group.POST("/wechat/refund_notify", wechat.RefundNotify)
}

func setWechatAuthRoutes(authGroup *echo.Group) {
	wechatGroup := authGroup.Group("/wechat")
	{
		wechatGroup.POST("/order", wechat.CreateOrder)
		wechatGroup.GET("/order/:orderNo", wechat.QueryOrder)
		wechatGroup.POST("/refund", wechat.CreateRefund)
	}
}
