package alipay

import (
	"fmt"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/oldfritter/lucy/base"
	"github.com/oldfritter/lucy/dom"
	libAlipay "github.com/oldfritter/lucy/lib/payment/alipay"
	"github.com/oldfritter/lucy/lib/db"
	"github.com/oldfritter/lucy/model"
	"github.com/oldfritter/lucy/util"
)

// CreatePagePayOrder 创建 PC 网站支付订单，返回自动提交的 HTML 表单
func CreatePagePayOrder(c echo.Context) (err error) {
	claims, _ := base.GetClaim(c)

	var input struct {
		OrderNo string `form:"OrderNo" validate:"required"`
	}
	if err = c.Bind(&input); err != nil {
		return util.BuildError("1001")
	}
	if err = c.Validate(&input); err != nil {
		return util.BuildError("1002", err.Error())
	}

	// 查询订单，验证归属
	var order model.Order
	if err = db.MysqlDB.Preload("Currency").
		Where("order_no = ? AND user_id = ?", input.OrderNo, claims.UserId).
		First(&order).Error; err != nil {
		return util.BuildError("1003", "订单不存在")
	}

	if order.Status != dom.StatusPending {
		return util.BuildError("1005", "订单状态不允许支付")
	}

	if order.OrderNo == "" {
		return util.BuildError("1007", "订单号缺失")
	}

	client := libAlipay.NewClient()

	subject := "lucy-captcha"
	biz := libAlipay.BizContentTradePay{
		OutTradeNo:  order.OrderNo,
		TotalAmount: libAlipay.ConvertFenToYuan(order.FinalAmount),
		Subject:     subject,
	}

	htmlForm, err := client.CreatePagePayOrder(biz)
	if err != nil {
		return util.BuildError("1007", fmt.Sprintf("支付宝下单失败: %v", err))
	}

	response := util.SuccessResponse()
	response.Body = map[string]string{"form": htmlForm}
	return c.JSON(http.StatusOK, response)
}

// CreateWapPayOrder 创建手机网站支付订单，返回重定向 URL
func CreateWapPayOrder(c echo.Context) (err error) {
	claims, _ := base.GetClaim(c)

	var input struct {
		OrderNo string `form:"OrderNo" validate:"required"`
	}
	if err = c.Bind(&input); err != nil {
		return util.BuildError("1001")
	}
	if err = c.Validate(&input); err != nil {
		return util.BuildError("1002", err.Error())
	}

	var order model.Order
	if err = db.MysqlDB.Preload("Currency").
		Where("order_no = ? AND user_id = ?", input.OrderNo, claims.UserId).
		First(&order).Error; err != nil {
		return util.BuildError("1003", "订单不存在")
	}

	if order.Status != dom.StatusPending {
		return util.BuildError("1005", "订单状态不允许支付")
	}

	if order.OrderNo == "" {
		return util.BuildError("1007", "订单号缺失")
	}

	client := libAlipay.NewClient()

	subject := "lucy-captcha"
	biz := libAlipay.BizContentTradePay{
		OutTradeNo:  order.OrderNo,
		TotalAmount: libAlipay.ConvertFenToYuan(order.FinalAmount),
		Subject:     subject,
	}

	payURL, err := client.CreateWapPayOrder(biz)
	if err != nil {
		return util.BuildError("1007", fmt.Sprintf("支付宝下单失败: %v", err))
	}

	response := util.SuccessResponse()
	response.Body = map[string]string{"pay_url": payURL}
	return c.JSON(http.StatusOK, response)
}

// QueryOrder 查询支付宝订单状态
func QueryOrder(c echo.Context) (err error) {
	claims, _ := base.GetClaim(c)

	orderNo := c.Param("orderNo")
	if orderNo == "" {
		return util.BuildError("1001", "缺少订单号")
	}

	var order model.Order
	if err = db.MysqlDB.Where("order_no = ? AND user_id = ?", orderNo, claims.UserId).
		First(&order).Error; err != nil {
		return util.BuildError("1003", "订单不存在")
	}

	client := libAlipay.NewClient()

	aliOrder, err := client.QueryOrder(libAlipay.BizContentTradeQuery{
		OutTradeNo: order.OrderNo,
	})
	if err != nil {
		return util.BuildError("1007", fmt.Sprintf("查询支付宝订单失败: %v", err))
	}

	response := util.SuccessResponse()
	response.Body = aliOrder
	return c.JSON(http.StatusOK, response)
}

// PayNotify 支付宝支付回调通知
func PayNotify(c echo.Context) (err error) {
	// 1. 读取请求体（URL-encoded）
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "read body failed")
	}

	// 2. 解析并验证签名
	notify, err := libAlipay.ParseNotify(string(body))
	if err != nil {
		return c.JSON(http.StatusUnauthorized, "signature verification failed")
	}

	// 3. 仅处理交易成功事件
	if notify.TradeStatus != libAlipay.TradeStatusSuccess {
		return c.String(http.StatusOK, "success")
	}

	// 4. 幂等检查
	var existingIncome model.Income
	if db.MysqlDB.Where("transaction_id = ?", notify.TradeNo).First(&existingIncome).Error == nil {
		return c.String(http.StatusOK, "success")
	}

	// 5. 查找对应订单
	var order model.Order
	if err := db.MysqlDB.Where("order_no = ?", notify.OutTradeNo).First(&order).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, "order not found")
	}

	// 6. 验证金额
	notifyAmount, err := libAlipay.ConvertYuanToFen(notify.TotalAmount)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "invalid amount")
	}
	if order.FinalAmount != notifyAmount {
		return c.JSON(http.StatusInternalServerError, "amount mismatch")
	}

	// 7. 更新订单和创建入账记录（事务）
	tx := db.BeginTx()
	defer tx.DbRollback()

	if err := tx.Model(&order).Update("status", dom.StatusPaid).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, "update order failed")
	}

	// 支付完成后激活关联的 Reason（如 Campaign）
	if err := model.ActivateReason(tx.DB, order.ReasonType, order.ReasonId); err != nil {
		return c.JSON(http.StatusInternalServerError, "activate reason failed: "+err.Error())
	}

	income := model.Income{
		Income: dom.Income{
			OrderId:       order.Id,
			Amount:        order.FinalAmount,
			PaymentSource: "alipay",
			Status:        dom.StatusPaid,
			TransactionId: notify.TradeNo,
			NotifyData:    string(body),
		},
	}
	if err := tx.Create(&income).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, "create income failed")
	}

	tx.DbCommit()

	return c.String(http.StatusOK, "success")
}

// CreateRefund 申请退款
func CreateRefund(c echo.Context) (err error) {
	claims, _ := base.GetClaim(c)

	var input struct {
		IncomeId int    `form:"IncomeId" validate:"required"`
		Amount   int    `form:"Amount" validate:"required,min=1"`
		Reason   string `form:"Reason"`
	}
	if err = c.Bind(&input); err != nil {
		return util.BuildError("1001")
	}
	if err = c.Validate(&input); err != nil {
		return util.BuildError("1002", err.Error())
	}

	var income model.Income
	if err = db.MysqlDB.Preload("Order").
		Where("id = ? AND payment_source = ?", input.IncomeId, "alipay").
		First(&income).Error; err != nil {
		return util.BuildError("1003", "入账记录不存在或非支付宝支付")
	}

	if income.Order == nil || income.Order.UserId != claims.UserId {
		return util.BuildError("1003", "入账记录不存在")
	}

	if input.Amount > income.Amount {
		return util.BuildError("1005", "退款金额不能超过入账金额")
	}

	refundNo := fmt.Sprintf("RF%s%s", income.Order.OrderNo, fmt.Sprintf("%d", income.Id))

	client := libAlipay.NewClient()

	refundResp, err := client.ApplyRefund(libAlipay.BizContentTradeRefund{
		OutTradeNo:   income.Order.OrderNo,
		OutRequestNo: refundNo,
		RefundAmount: libAlipay.ConvertFenToYuan(input.Amount),
		RefundReason: input.Reason,
	})
	if err != nil {
		return util.BuildError("1007", fmt.Sprintf("支付宝退款失败: %v", err))
	}

	tx := db.BeginTx()
	defer tx.DbRollback()

	refund := model.Refund{
		Refund: dom.Refund{
			IncomeId:      income.Id,
			Amount:        input.Amount,
			Status:        dom.StatusPending,
			RefundNo:      refundNo,
			TransactionId: refundResp.TradeNo,
			NotifyData:    fmt.Sprintf(`{"trade_no":"%s","out_trade_no":"%s","refund_fee":"%s"}`, refundResp.TradeNo, refundResp.OutTradeNo, refundResp.RefundFee),
		},
	}
	if err := tx.Create(&refund).Error; err != nil {
		return util.BuildError("1007", "创建退款记录失败")
	}

	tx.DbCommit()

	response := util.SuccessResponse()
	response.Body = refundResp
	return c.JSON(http.StatusOK, response)
}
