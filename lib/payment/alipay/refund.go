package alipay

import (
	"encoding/json"
	"fmt"
)

// --- 退款请求 & 响应 ---

// BizContentTradeRefund 退款业务参数
type BizContentTradeRefund struct {
	OutTradeNo   string `json:"out_trade_no,omitempty"`
	TradeNo      string `json:"trade_no,omitempty"`      // 支付宝交易号，与 OutTradeNo 二选一
	OutRequestNo string `json:"out_request_no"`          // 商户退款单号（必填）
	RefundAmount string `json:"refund_amount"`           // 退款金额（元），字符串
	RefundReason string `json:"refund_reason,omitempty"` // 退款原因
}

// RefundResponse 退款响应
type RefundResponse struct {
	responseMsg
	TradeNo      string `json:"trade_no"`
	OutTradeNo   string `json:"out_trade_no"`
	OutRequestNo string `json:"out_request_no"`
	RefundFee    string `json:"refund_fee"`
	FundChange   string `json:"fund_change"`
	GmtRefundPay string `json:"gmt_refund_pay"`
}

// ApplyRefund 申请退款
func (c *Client) ApplyRefund(biz BizContentTradeRefund) (*RefundResponse, error) {
	common := c.newCommonParams("alipay.trade.refund")

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

	if alipayResp.AlipayTradeRefundResponse == nil {
		return nil, fmt.Errorf("empty refund response")
	}

	rr := alipayResp.AlipayTradeRefundResponse
	if err := checkResponse(rr.Code, rr.Msg, rr.SubCode, rr.SubMsg); err != nil {
		return nil, err
	}

	return rr, nil
}

// --- 退款查询 ---

// BizContentRefundQuery 退款查询业务参数
type BizContentRefundQuery struct {
	OutTradeNo   string `json:"out_trade_no,omitempty"`
	TradeNo      string `json:"trade_no,omitempty"`
	OutRequestNo string `json:"out_request_no"` // 商户退款单号（必填）
}

// RefundQueryResponse 退款查询响应
type RefundQueryResponse struct {
	responseMsg
	TradeNo       string `json:"trade_no"`
	OutTradeNo    string `json:"out_trade_no"`
	OutRequestNo  string `json:"out_request_no"`
	RefundAmount  string `json:"refund_amount"`
	TotalAmount   string `json:"total_amount"`
	GmtRefundPay  string `json:"gmt_refund_pay"`
	RefundStatus  string `json:"refund_status"` // REFUND_SUCCESS
}

// QueryRefund 查询退款状态（通过商户退款单号）
func (c *Client) QueryRefund(biz BizContentRefundQuery) (*RefundQueryResponse, error) {
	common := c.newCommonParams("alipay.trade.fastpay.refund.query")

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

	if alipayResp.AlipayTradeFastpayRefundQueryResponse == nil {
		return nil, fmt.Errorf("empty refund query response")
	}

	rqr := alipayResp.AlipayTradeFastpayRefundQueryResponse
	if err := checkResponse(rqr.Code, rqr.Msg, rqr.SubCode, rqr.SubMsg); err != nil {
		return nil, err
	}

	return rqr, nil
}
