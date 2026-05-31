package route

import (
	"github.com/labstack/echo/v4"

	"github.com/oldfritter/lucy/service/api/alipay"
)

func setAlipayRoutes(v1Group *echo.Group) {
	v1Group.POST("/alipay/notify", alipay.PayNotify)
}

func setAlipayAuthRoutes(authGroup *echo.Group) {
	alipayGroup := authGroup.Group("/alipay")
	{
		alipayGroup.POST("/page-pay", alipay.CreatePagePayOrder)
		alipayGroup.POST("/wap-pay", alipay.CreateWapPayOrder)
		alipayGroup.GET("/order/:orderNo", alipay.QueryOrder)
		alipayGroup.POST("/refund", alipay.CreateRefund)
	}
}
