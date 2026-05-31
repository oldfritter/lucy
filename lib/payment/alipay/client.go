package alipay

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

// Client 支付宝支付 HTTP 客户端
type Client struct {
	httpClient *http.Client
	config     *Config
}

// NewClient 创建支付宝支付客户端
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		config:     GetConfig(),
	}
}

// CommonParams 支付宝公共请求参数
type CommonParams struct {
	AppID     string `json:"app_id"`
	Method    string `json:"method"`
	Format    string `json:"format"`
	Charset   string `json:"charset"`
	SignType  string `json:"sign_type"`
	Timestamp string `json:"timestamp"`
	Version   string `json:"version"`
}

// newCommonParams 创建公共参数
func (c *Client) newCommonParams(method string) CommonParams {
	return CommonParams{
		AppID:     c.config.AppID,
		Method:    method,
		Format:    "JSON",
		Charset:   "utf-8",
		SignType:  c.config.SignType,
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		Version:   "1.0",
	}
}

// sign 生成 RSA2 签名
// 支付宝签名规则：将所有参数（不含 sign）按 key 字母排序，拼接为 key=value&key=value 格式，用私钥签名
func (c *Client) sign(params map[string]string) (string, error) {
	// 过滤空值和 sign 字段
	keys := make([]string, 0, len(params))
	for k, v := range params {
		if k != "sign" && v != "" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	// 拼接待签名字符串
	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, params[k]))
	}
	signStr := strings.Join(parts, "&")

	// RSA-SHA256 签名
	h := sha256.New()
	h.Write([]byte(signStr))
	digest := h.Sum(nil)

	signature, err := rsa.SignPKCS1v15(rand.Reader, c.config.PrivateKey, crypto.SHA256, digest)
	if err != nil {
		return "", fmt.Errorf("sign: %w", err)
	}

	return base64.StdEncoding.EncodeToString(signature), nil
}

// buildFormParams 构建表单参数（含签名）
func (c *Client) buildFormParams(params map[string]string) (url.Values, error) {
	sign, err := c.sign(params)
	if err != nil {
		return nil, err
	}
	params["sign"] = sign

	form := make(url.Values)
	for k, v := range params {
		form.Set(k, v)
	}
	return form, nil
}

// Post 执行 POST 请求到支付宝网关，返回响应字节
func (c *Client) Post(params map[string]string) ([]byte, error) {
	form, err := c.buildFormParams(params)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.config.Gateway, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded;charset=utf-8")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	return respBody, nil
}

// BuildPagePayForm 构建 PC 网页支付的自动提交表单（HTML）
// 返回完整 HTML 字符串，前端可直接输出
func (c *Client) BuildPagePayForm(method string, params map[string]string, bizContent string) (string, error) {
	form, err := c.buildFormParams(params)
	if err != nil {
		return "", err
	}
	// biz_content 是 JSON，需要单独加入表单
	if bizContent != "" {
		form.Set("biz_content", bizContent)
	}

	var buf bytes.Buffer
	buf.WriteString(`<form id="alipay-submit" name="alipay-submit" action="`)
	buf.WriteString(c.config.Gateway)
	buf.WriteString(`" method="POST">`)
	buf.WriteString("\n")

	for key, values := range form {
		for _, v := range values {
			buf.WriteString(fmt.Sprintf(`<input type="hidden" name="%s" value="%s"/>`, key, v))
			buf.WriteString("\n")
		}
	}

	buf.WriteString(`<input type="submit" value="ok" style="display:none;">`)
	buf.WriteString("\n</form>\n")
	buf.WriteString(`<script>document.forms["alipay-submit"].submit();</script>`)

	return buf.String(), nil
}

// BuildPayURL 构建移动网页支付 URL（用于重定向）
func (c *Client) BuildPayURL(method string, params map[string]string, bizContent string) (string, error) {
	form, err := c.buildFormParams(params)
	if err != nil {
		return "", err
	}
	if bizContent != "" {
		form.Set("biz_content", bizContent)
	}

	return c.config.Gateway + "?" + form.Encode(), nil
}
