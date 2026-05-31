package alipay

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

// --- 公共响应结构 ---

// AlipayResponse 支付宝通用 API 响应体
type AlipayResponse struct {
	AlipayTradePagePayResponse  *PagePayResponse  `json:"alipay_trade_page_pay_response,omitempty"`
	AlipayTradeWapPayResponse   *WapPayResponse   `json:"alipay_trade_wap_pay_response,omitempty"`
	AlipayTradeQueryResponse    *QueryResponse    `json:"alipay_trade_query_response,omitempty"`
	AlipayTradeRefundResponse   *RefundResponse   `json:"alipay_trade_refund_response,omitempty"`
	AlipayTradeFastpayRefundQueryResponse *RefundQueryResponse `json:"alipay_trade_fastpay_refund_query_response,omitempty"`
	Sign                        string           `json:"sign"`
}

// 支付宝响应错误子结构
type responseMsg struct {
	Code    string `json:"code"`
	Msg     string `json:"msg"`
	SubCode string `json:"sub_code"`
	SubMsg  string `json:"sub_msg"`
}

// checkResponse 检查响应是否为成功（code == "10000"）
func checkResponse(code, msg, subCode, subMsg string) error {
	if code == "10000" {
		return nil
	}
	errMsg := fmt.Sprintf("alipay error: code=%s, msg=%s", code, msg)
	if subCode != "" {
		errMsg += fmt.Sprintf(", sub_code=%s, sub_msg=%s", subCode, subMsg)
	}
	return errors.New(errMsg)
}

// --- 统一下单请求 ---

// BizContentTradePay 统一下单业务参数（biz_content）
type BizContentTradePay struct {
	OutTradeNo  string `json:"out_trade_no"`
	ProductCode string `json:"product_code"`
	TotalAmount string `json:"total_amount"` // 元，字符串精度控制
	Subject     string `json:"subject"`
	Body        string `json:"body,omitempty"`
	PassbackParams string `json:"passback_params,omitempty"` // 回传参数（URL编码），回调时原样返回
	TimeExpire  string `json:"time_expire,omitempty"`        // 绝对超时时间，格式 yyyy-MM-dd HH:mm:ss
}

// PagePayResponse 电脑网站支付响应
type PagePayResponse struct {
	responseMsg
	OutTradeNo string `json:"out_trade_no"`
}

// WapPayResponse 手机网站支付响应
type WapPayResponse struct {
	responseMsg
	OutTradeNo string `json:"out_trade_no"`
}

// CreatePagePayOrder 创建 PC 网站支付订单，返回自动提交的 HTML 表单
func (c *Client) CreatePagePayOrder(biz BizContentTradePay) (string, error) {
	common := c.newCommonParams("alipay.trade.page.pay")
	biz.ProductCode = "FAST_INSTANT_TRADE_PAY"

	bizJSON, err := json.Marshal(biz)
	if err != nil {
		return "", fmt.Errorf("marshal biz_content: %w", err)
	}

	params := map[string]string{
		"app_id":      common.AppID,
		"method":      common.Method,
		"format":      common.Format,
		"charset":     common.Charset,
		"sign_type":   common.SignType,
		"timestamp":   common.Timestamp,
		"version":     common.Version,
		"notify_url":  c.config.NotifyURL,
		"return_url":  c.config.ReturnURL,
		"biz_content": string(bizJSON),
	}

	return c.BuildPagePayForm(common.Method, params, string(bizJSON))
}

// CreateWapPayOrder 创建手机网站支付订单，返回重定向 URL
func (c *Client) CreateWapPayOrder(biz BizContentTradePay) (string, error) {
	common := c.newCommonParams("alipay.trade.wap.pay")
	biz.ProductCode = "QUICK_WAP_WAY"

	bizJSON, err := json.Marshal(biz)
	if err != nil {
		return "", fmt.Errorf("marshal biz_content: %w", err)
	}

	params := map[string]string{
		"app_id":      common.AppID,
		"method":      common.Method,
		"format":      common.Format,
		"charset":     common.Charset,
		"sign_type":   common.SignType,
		"timestamp":   common.Timestamp,
		"version":     common.Version,
		"notify_url":  c.config.NotifyURL,
		"return_url":  c.config.ReturnURL,
		"biz_content": string(bizJSON),
	}

	return c.BuildPayURL(common.Method, params, string(bizJSON))
}

// --- 查询订单 ---

// BizContentTradeQuery 查询订单业务参数
type BizContentTradeQuery struct {
	OutTradeNo string `json:"out_trade_no,omitempty"`
	TradeNo    string `json:"trade_no,omitempty"` // 支付宝交易号，与 OutTradeNo 二选一
}

// QueryResponse 查询订单响应
type TradeQueryResponse struct {
	responseMsg
	TradeNo       string `json:"trade_no"`
	OutTradeNo    string `json:"out_trade_no"`
	TotalAmount   string `json:"total_amount"`
	TradeStatus   string `json:"trade_status"`
	BuyerLogonID  string `json:"buyer_logon_id"`
	BuyerPayAmount string `json:"buyer_pay_amount"`
	PointAmount   string `json:"point_amount"`
	InvoiceAmount string `json:"invoice_amount"`
	ReceiptAmount string `json:"receipt_amount"`
	SendPayDate   string `json:"send_pay_date"`
	GmtPayment    string `json:"gmt_payment"`
}

type QueryResponse struct {
	TradeQueryResponse
}

// QueryOrder 查询订单状态（通过商户订单号或支付宝交易号）
func (c *Client) QueryOrder(biz BizContentTradeQuery) (*TradeQueryResponse, error) {
	common := c.newCommonParams("alipay.trade.query")

	bizJSON, err := json.Marshal(biz)
	if err != nil {
		return nil, fmt.Errorf("marshal biz_content: %w", err)
	}

	params := map[string]string{
		"app_id":      common.AppID,
		"method":      common.Method,
		"format":      common.Format,
		"charset":     common.Charset,
		"sign_type":   common.SignType,
		"timestamp":   common.Timestamp,
		"version":     common.Version,
		"biz_content": string(bizJSON),
	}

	respBody, err := c.Post(params)
	if err != nil {
		return nil, err
	}

	var alipayResp AlipayResponse
	if err := json.Unmarshal(respBody, &alipayResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if alipayResp.AlipayTradeQueryResponse == nil {
		return nil, fmt.Errorf("empty query response")
	}

	qr := alipayResp.AlipayTradeQueryResponse
	if err := checkResponse(qr.Code, qr.Msg, qr.SubCode, qr.SubMsg); err != nil {
		return nil, err
	}

	return &qr.TradeQueryResponse, nil
}

// ConvertYuanToFen 将元（字符串）转为分（整数）
func ConvertYuanToFen(yuan string) (int, error) {
	f, err := strconv.ParseFloat(yuan, 64)
	if err != nil {
		return 0, fmt.Errorf("parse amount: %w", err)
	}
	return int(f * 100), nil
}

// ConvertFenToYuan 将分（整数）转为元（字符串）
func ConvertFenToYuan(fen int) string {
	return fmt.Sprintf("%.2f", float64(fen)/100.0)
}
