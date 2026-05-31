package wechat

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

type JSAPIOrderRequest struct {
	AppID       string      `json:"appid"`
	MchID       string      `json:"mchid"`
	Description string      `json:"description"`
	OutTradeNo  string      `json:"out_trade_no"`
	NotifyURL   string      `json:"notify_url"`
	Amount      OrderAmount `json:"amount"`
	Payer       OrderPayer  `json:"payer"`
}

type OrderAmount struct {
	Total    int    `json:"total"`
	Currency string `json:"currency"`
}

type OrderPayer struct {
	OpenID string `json:"openid"`
}

type JSAPIOrderResponse struct {
	PrepayID string `json:"prepay_id"`
}

type QueryOrderResponse struct {
	AppID          string      `json:"appid"`
	MchID          string      `json:"mchid"`
	OutTradeNo     string      `json:"out_trade_no"`
	TransactionID  string      `json:"transaction_id"`
	TradeState     string      `json:"trade_state"`
	TradeStateDesc string      `json:"trade_state_desc"`
	Amount         OrderAmount `json:"amount"`
	Payer          OrderPayer  `json:"payer"`
	SuccessTime    string      `json:"success_time"`
}

type PrepayParams struct {
	AppID     string `json:"appId"`
	TimeStamp string `json:"timeStamp"`
	NonceStr  string `json:"nonceStr"`
	Package   string `json:"package"`
	SignType  string `json:"signType"`
	PaySign   string `json:"paySign"`
}

func (c *Client) CreateJSAPIOrder(description, outTradeNo, openID string, totalAmount int) (*PrepayParams, error) {
	cfg := c.config
	reqBody := JSAPIOrderRequest{
		AppID:       cfg.AppID,
		MchID:       cfg.MchID,
		Description: description,
		OutTradeNo:  outTradeNo,
		NotifyURL:   cfg.NotifyURL,
		Amount:      OrderAmount{Total: totalAmount, Currency: "CNY"},
		Payer:       OrderPayer{OpenID: openID},
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}
	respBody, err := c.Post("/v3/pay/transactions/jsapi", body)
	if err != nil {
		return nil, err
	}
	var orderResp JSAPIOrderResponse
	if err := json.Unmarshal(respBody, &orderResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}
	prepayID := orderResp.PrepayID
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	nonceStr := randomString(32)
	pkg := "prepay_id=" + prepayID
	signStr := fmt.Sprintf("%s\n%s\n%s\n%s\n", cfg.AppID, timestamp, nonceStr, pkg)
	h := sha256.New()
	h.Write([]byte(signStr))
	digest := h.Sum(nil)
	signature, err := rsa.SignPKCS1v15(rand.Reader, cfg.PrivateKey, crypto.SHA256, digest)
	if err != nil {
		return nil, fmt.Errorf("sign prepay params: %w", err)
	}
	return &PrepayParams{
		AppID:     cfg.AppID,
		TimeStamp: timestamp,
		NonceStr:  nonceStr,
		Package:   pkg,
		SignType:  "RSA",
		PaySign:   base64.StdEncoding.EncodeToString(signature),
	}, nil
}

func (c *Client) QueryOrder(outTradeNo string) (*QueryOrderResponse, error) {
	path := fmt.Sprintf("/v3/pay/transactions/out-trade-no/%s?mchid=%s", outTradeNo, c.config.MchID)
	respBody, err := c.Get(path, nil)
	if err != nil {
		return nil, err
	}
	var result QueryOrderResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}
	return &result, nil
}
