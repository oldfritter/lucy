package wechat

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/oldfritter/lucy/base"
	"github.com/oldfritter/lucy/dom"
	"github.com/oldfritter/lucy/lib/db"
	libWechat "github.com/oldfritter/lucy/lib/payment/wechat"
	"github.com/oldfritter/lucy/model"
	"github.com/oldfritter/lucy/util"
)

// CreateOrder 创建微信支付订单
func CreateOrder(c echo.Context) (err error) {
	claims, _ := base.GetClaim(c)

	var input struct {
		OrderNo string `form:"OrderNo" validate:"required"`
		OpenID  string `form:"OpenID" validate:"required"`
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

	client := libWechat.NewClient()

	// 调用微信支付下单
	prepayParams, err := client.CreateJSAPIOrder(
		"lucy-captcha",       // description
		order.OrderNo,        // out_trade_no
		input.OpenID,         // openid
		order.FinalAmount,    // total amount (分)
	)
	if err != nil {
		return util.BuildError("1007", fmt.Sprintf("微信支付下单失败: %v", err))
	}

	response := util.SuccessResponse()
	response.Body = prepayParams
	return c.JSON(http.StatusOK, response)
}

// QueryOrder 查询微信支付订单状态
func QueryOrder(c echo.Context) (err error) {
	claims, _ := base.GetClaim(c)

	orderNo := c.Param("orderNo")
	if orderNo == "" {
		return util.BuildError("1001", "缺少订单号")
	}

	// 查询订单，验证归属
	var order model.Order
	if err = db.MysqlDB.Where("order_no = ? AND user_id = ?", orderNo, claims.UserId).
		First(&order).Error; err != nil {
		return util.BuildError("1003", "订单不存在")
	}

	client := libWechat.NewClient()

	// 查询微信支付订单状态
	wxOrder, err := client.QueryOrder(order.OrderNo)
	if err != nil {
		return util.BuildError("1007", fmt.Sprintf("查询微信支付订单失败: %v", err))
	}

	response := util.SuccessResponse()
	response.Body = wxOrder
	return c.JSON(http.StatusOK, response)
}

// PayNotify 微信支付回调通知
func PayNotify(c echo.Context) (err error) {
	// 1. 读取 HTTP 头
	timestamp := c.Request().Header.Get("Wechatpay-Timestamp")
	nonce := c.Request().Header.Get("Wechatpay-Nonce")
	signature := c.Request().Header.Get("Wechatpay-Signature")
	serialNo := c.Request().Header.Get("Wechatpay-Serial")

	if timestamp == "" || nonce == "" || signature == "" || serialNo == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"code": "FAIL", "message": "missing required headers"})
	}

	// 2. 读取请求体
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"code": "FAIL", "message": "read body failed"})
	}

	// 3. 验证签名
	if err := libWechat.VerifyNotifySignatureWithSerial(serialNo, timestamp, nonce, signature, body); err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"code": "FAIL", "message": "signature verification failed"})
	}

	// 4. 解析通知
	var notify libWechat.NotifyRequest
	if err := json.Unmarshal(body, &notify); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"code": "FAIL", "message": "invalid notify body"})
	}

	// 5. 仅处理支付成功事件
	if notify.EventType != "TRANSACTION.SUCCESS" {
		return c.JSON(http.StatusOK, map[string]string{"code": "SUCCESS"})
	}

	// 6. 解密资源
	cfg := libWechat.GetConfig()
	decrypted, err := libWechat.DecryptNotifyResource(notify.Resource, cfg.APIv3Key)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"code": "FAIL", "message": "decrypt failed"})
	}

	var txn libWechat.TransactionNotification
	if err := json.Unmarshal(decrypted, &txn); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"code": "FAIL", "message": "invalid transaction data"})
	}

	// 7. 幂等检查：是否已处理过该交易号
	var existingIncome model.Income
	if db.MysqlDB.Where("transaction_id = ?", txn.TransactionID).First(&existingIncome).Error == nil {
		// 已处理，直接返回成功
		return c.JSON(http.StatusOK, map[string]string{"code": "SUCCESS"})
	}

	// 8. 查找对应订单
	var order model.Order
	if err := db.MysqlDB.Where("order_no = ?", txn.OutTradeNo).First(&order).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"code": "FAIL", "message": "order not found"})
	}

	// 9. 验证金额
	if order.FinalAmount != txn.Amount.Total {
		return c.JSON(http.StatusInternalServerError, map[string]string{"code": "FAIL", "message": "amount mismatch"})
	}

	// 10. 更新订单和创建入账记录（事务）
	tx := db.BeginTx()
	defer tx.DbRollback()

	// 更新订单状态为已付款
	if err := tx.Model(&order).Update("status", dom.StatusPaid).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"code": "FAIL", "message": "update order failed"})
	}

	// 支付完成后激活关联的 Reason（如 Campaign）
	if err := model.ActivateReason(tx.DB, order.ReasonType, order.ReasonId); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"code": "FAIL", "message": "activate reason failed: " + err.Error()})
	}

	// 创建入账记录
	income := model.Income{
		Income: dom.Income{
			OrderId:       order.Id,
			Amount:        txn.Amount.Total,
			PaymentSource: "wechat",
			Status:        dom.StatusPaid,
			TransactionId: txn.TransactionID,
			NotifyData:    string(decrypted),
		},
	}
	if err := tx.Create(&income).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"code": "FAIL", "message": "create income failed"})
	}

	tx.DbCommit()

	return c.JSON(http.StatusOK, map[string]string{"code": "SUCCESS"})
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

	// 查询入账记录
	var income model.Income
	if err = db.MysqlDB.Preload("Order").
		Where("id = ? AND payment_source = ?", input.IncomeId, "wechat").
		First(&income).Error; err != nil {
		return util.BuildError("1003", "入账记录不存在或非微信支付")
	}

	// 验证归属
	if income.Order == nil || income.Order.UserId != claims.UserId {
		return util.BuildError("1003", "入账记录不存在")
	}

	// 验证退款金额
	if input.Amount > income.Amount {
		return util.BuildError("1005", "退款金额不能超过入账金额")
	}

	// 生成退款单号
	refundNo := fmt.Sprintf("RF%s%s", income.Order.OrderNo, fmt.Sprintf("%d", income.Id))

	client := libWechat.NewClient()

	// 调用微信退款 API
	refundResp, err := client.ApplyRefund(
		income.Order.OrderNo, // out_trade_no (商户订单号)
		refundNo,              // out_refund_no
		input.Reason,
		input.Amount,  // refund amount
		income.Amount, // total amount
	)
	if err != nil {
		return util.BuildError("1007", fmt.Sprintf("微信退款失败: %v", err))
	}

	// 记录退款
	tx := db.BeginTx()
	defer tx.DbRollback()

	notifyDataJSON, _ := json.Marshal(refundResp)

	refund := model.Refund{
		Refund: dom.Refund{
			IncomeId:      income.Id,
			Amount:        input.Amount,
			Status:        dom.StatusPending, // 等待退款回调确认
			RefundNo:      refundNo,
			TransactionId: refundResp.RefundID,
			NotifyData:    string(notifyDataJSON),
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

// RefundNotify 微信退款回调通知
func RefundNotify(c echo.Context) (err error) {
	// 1. 读取 HTTP 头
	timestamp := c.Request().Header.Get("Wechatpay-Timestamp")
	nonce := c.Request().Header.Get("Wechatpay-Nonce")
	signature := c.Request().Header.Get("Wechatpay-Signature")
	serialNo := c.Request().Header.Get("Wechatpay-Serial")

	if timestamp == "" || nonce == "" || signature == "" || serialNo == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"code": "FAIL", "message": "missing required headers"})
	}

	// 2. 读取请求体
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"code": "FAIL", "message": "read body failed"})
	}

	// 3. 验证签名
	if err := libWechat.VerifyNotifySignatureWithSerial(serialNo, timestamp, nonce, signature, body); err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"code": "FAIL", "message": "signature verification failed"})
	}

	// 4. 解析通知
	var notify libWechat.NotifyRequest
	if err := json.Unmarshal(body, &notify); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"code": "FAIL", "message": "invalid notify body"})
	}

	// 5. 仅处理退款成功事件
	if notify.EventType != "REFUND.SUCCESS" {
		return c.JSON(http.StatusOK, map[string]string{"code": "SUCCESS"})
	}

	// 6. 解密资源
	cfg := libWechat.GetConfig()
	decrypted, err := libWechat.DecryptNotifyResource(notify.Resource, cfg.APIv3Key)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"code": "FAIL", "message": "decrypt failed"})
	}

	var refNotif libWechat.RefundNotification
	if err := json.Unmarshal(decrypted, &refNotif); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"code": "FAIL", "message": "invalid refund data"})
	}

	// 7. 查找退款记录并更新状态
	var refund model.Refund
	if err := db.MysqlDB.Where("refund_no = ?", refNotif.OutRefundNo).First(&refund).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"code": "FAIL", "message": "refund record not found"})
	}

	// 8. 更新退款记录
	updates := map[string]any{
		"status":         dom.StatusFullRefund,
		"transaction_id": refNotif.RefundID,
		"notify_data":    string(decrypted),
	}

	tx := db.BeginTx()
	defer tx.DbRollback()

	if err := tx.Model(&refund).Updates(updates).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"code": "FAIL", "message": "update refund failed"})
	}

	// 9. 同步更新订单状态
	var income model.Income
	if db.MysqlDB.Where("id = ?", refund.IncomeId).First(&income).Error == nil {
		if refund.Amount >= income.Amount {
			// 全额退款
			tx.Model(&model.Order{}).Where("id = ?", income.OrderId).
				Update("status", dom.StatusFullRefund)
		} else {
			// 部分退款
			tx.Model(&model.Order{}).Where("id = ?", income.OrderId).
				Update("status", dom.StatusPartialRefund)
		}
	}

	tx.DbCommit()

	return c.JSON(http.StatusOK, map[string]string{"code": "SUCCESS"})
}
