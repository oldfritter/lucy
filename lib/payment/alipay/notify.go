package alipay

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strings"
)

// --- 回调通知数据结构 ---

// NotifyRequest 支付宝回调通知
type NotifyRequest struct {
	NotifyTime       string `json:"notify_time"`        // 通知时间 yyyy-MM-dd HH:mm:ss
	NotifyType       string `json:"notify_type"`        // 通知类型：trade_status_sync
	NotifyID         string `json:"notify_id"`          // 通知ID
	AppID            string `json:"app_id"`             // 应用ID
	Charset          string `json:"charset"`            // 编码
	SignType         string `json:"sign_type"`          // 签名类型
	Sign             string `json:"sign"`               // 签名
	TradeNo          string `json:"trade_no"`           // 支付宝交易号
	OutTradeNo       string `json:"out_trade_no"`       // 商户订单号
	OutBizNo         string `json:"out_biz_no"`         // 外部业务号
	BuyerLogonID     string `json:"buyer_logon_id"`     // 买家支付宝账号
	TradeStatus      string `json:"trade_status"`       // 交易状态
	TotalAmount      string `json:"total_amount"`       // 订单金额（元）
	ReceiptAmount    string `json:"receipt_amount"`     // 实收金额（元）
	InvoiceAmount    string `json:"invoice_amount"`     // 开票金额（元）
	BuyerPayAmount   string `json:"buyer_pay_amount"`   // 买家实付（元）
	PointAmount      string `json:"point_amount"`       // 集分宝金额（元）
	GmtCreate        string `json:"gmt_create"`         // 交易创建时间
	GmtPayment       string `json:"gmt_payment"`        // 交易付款时间
	Subject          string `json:"subject"`            // 订单标题
	Body             string `json:"body"`               // 订单描述
	PassbackParams   string `json:"passback_params"`    // 回传参数
	FundBillList     string `json:"fund_bill_list"`     // 支付金额信息
	VoucherDetailList string `json:"voucher_detail_list"` // 优惠券信息
}

// VerifyNotifySignature 验证支付宝回调签名
// 1. 将通知参数中除去 sign、sign_type 外的所有参数按 key 字母升序排序
// 2. 拼接为 key=value&key=value 格式
// 3. 用支付宝公钥验证签名
func VerifyNotifySignature(params map[string]string, sign, signType string) error {
	cfg := GetConfig()

	if cfg.AlipayPublicKey == nil {
		return fmt.Errorf("alipay public key not configured")
	}

	// 过滤空值和 sign、sign_type
	keys := make([]string, 0, len(params))
	for k, v := range params {
		if k != "sign" && k != "sign_type" && v != "" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	// 拼接待验证字符串
	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, params[k]))
	}
	signStr := strings.Join(parts, "&")

	// RSA-SHA256 验签
	sigBytes, err := base64.StdEncoding.DecodeString(sign)
	if err != nil {
		return fmt.Errorf("decode signature: %w", err)
	}

	h := sha256.New()
	h.Write([]byte(signStr))
	digest := h.Sum(nil)

	if err := rsa.VerifyPKCS1v15(cfg.AlipayPublicKey, crypto.SHA256, digest, sigBytes); err != nil {
		return fmt.Errorf("signature verification failed: %w", err)
	}

	return nil
}

// ParseNotifyParams 将 URL-encoded 的通知字符串解析为 map[string]string
func ParseNotifyParams(body string) (map[string]string, error) {
	values, err := url.ParseQuery(body)
	if err != nil {
		return nil, fmt.Errorf("parse notify body: %w", err)
	}

	params := make(map[string]string)
	for k, v := range values {
		if len(v) > 0 {
			params[k] = v[0]
		}
	}
	return params, nil
}

// ParseNotify 解析并验证支付宝回调通知
// 返回解析后的 NotifyRequest 和验证错误
func ParseNotify(body string) (*NotifyRequest, error) {
	params, err := ParseNotifyParams(body)
	if err != nil {
		return nil, err
	}

	sign := params["sign"]
	signType := params["sign_type"]

	// 验证签名
	if err := VerifyNotifySignature(params, sign, signType); err != nil {
		return nil, err
	}

	// 将 map 转为 NotifyRequest
	jsonData, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("marshal params: %w", err)
	}

	var notify NotifyRequest
	if err := json.Unmarshal(jsonData, &notify); err != nil {
		return nil, fmt.Errorf("unmarshal notify: %w", err)
	}

	return &notify, nil
}

// TradeStatus 交易状态常量
const (
	TradeStatusWaitBuyerPay  = "WAIT_BUYER_PAY"  // 交易创建，等待买家付款
	TradeStatusClosed        = "TRADE_CLOSED"     // 交易关闭
	TradeStatusSuccess       = "TRADE_SUCCESS"    // 交易成功（支付成功）
	TradeStatusFinished      = "TRADE_FINISHED"   // 交易完结
)
