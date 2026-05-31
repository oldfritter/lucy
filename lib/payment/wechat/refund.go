package wechat

import (
	"encoding/json"
	"fmt"
)

type RefundAmount struct {
	Refund   int    `json:"refund"`
	Total    int    `json:"total"`
	Currency string `json:"currency"`
}

type RefundRequest struct {
	OutTradeNo  string       `json:"out_trade_no"`
	OutRefundNo string       `json:"out_refund_no"`
	Reason      string       `json:"reason,omitempty"`
	NotifyURL   string       `json:"notify_url,omitempty"`
	Amount      RefundAmount `json:"amount"`
}

type RefundResponse struct {
	RefundID      string       `json:"refund_id"`
	OutRefundNo   string       `json:"out_refund_no"`
	TransactionID string       `json:"transaction_id"`
	OutTradeNo    string       `json:"out_trade_no"`
	Channel       string       `json:"channel"`
	Status        string       `json:"status"`
	Amount        RefundAmount `json:"amount"`
}

type RefundNotification struct {
	MchID         string       `json:"mchid"`
	OutTradeNo    string       `json:"out_trade_no"`
	TransactionID string       `json:"transaction_id"`
	OutRefundNo   string       `json:"out_refund_no"`
	RefundID      string       `json:"refund_id"`
	RefundStatus  string       `json:"refund_status"`
	SuccessTime   string       `json:"success_time"`
	Amount        RefundAmount `json:"amount"`
}

func (c *Client) ApplyRefund(outTradeNo, outRefundNo, reason string, refundAmount, totalAmount int) (*RefundResponse, error) {
	cfg := c.config
	reqBody := RefundRequest{
		OutTradeNo:  outTradeNo,
		OutRefundNo: outRefundNo,
		Reason:      reason,
		NotifyURL:   cfg.RefundNotifyURL,
		Amount:      RefundAmount{Refund: refundAmount, Total: totalAmount, Currency: "CNY"},
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal refund request: %w", err)
	}
	respBody, err := c.Post("/v3/refund/domestic/refunds", body)
	if err != nil {
		return nil, err
	}
	var result RefundResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("unmarshal refund response: %w", err)
	}
	return &result, nil
}

func (c *Client) QueryRefund(outRefundNo string) (*RefundResponse, error) {
	path := fmt.Sprintf("/v3/refund/domestic/refunds/%s", outRefundNo)
	respBody, err := c.Get(path, nil)
	if err != nil {
		return nil, err
	}
	var result RefundResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("unmarshal query refund response: %w", err)
	}
	return &result, nil
}
