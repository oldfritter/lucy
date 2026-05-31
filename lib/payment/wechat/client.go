package wechat

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strconv"
	"time"
)

const WechatPayHost = "https://api.mch.weixin.qq.com"

type Client struct {
	httpClient *http.Client
	config     *Config
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		config:     GetConfig(),
	}
}

func (c *Client) sign(method, path string, timestamp int64, nonceStr string, body []byte) (string, error) {
	signStr := fmt.Sprintf("%s\n%s\n%d\n%s\n%s\n", method, path, timestamp, nonceStr, string(body))
	h := sha256.New()
	h.Write([]byte(signStr))
	digest := h.Sum(nil)
	signature, err := rsa.SignPKCS1v15(rand.Reader, c.config.PrivateKey, crypto.SHA256, digest)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(signature), nil
}

func (c *Client) buildAuthHeader(method, path string, body []byte) (string, error) {
	timestamp := time.Now().Unix()
	nonceStr := strconv.FormatInt(timestamp, 10) + randomString(16)
	signature, err := c.sign(method, path, timestamp, nonceStr, body)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(
		`WECHATPAY2-SHA256-RSA2048 mchid="%s",nonce_str="%s",signature="%s",timestamp="%d",serial_no="%s"`,
		c.config.MchID, nonceStr, signature, timestamp, c.config.SerialNo,
	), nil
}

func (c *Client) Do(method, path string, body []byte) ([]byte, int, error) {
	url := WechatPayHost + path
	authHeader, err := c.buildAuthHeader(method, path, body)
	if err != nil {
		return nil, 0, fmt.Errorf("build auth header: %w", err)
	}
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return nil, 0, fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "lucy-wechat/1.0")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("read response: %w", err)
	}
	return respBody, resp.StatusCode, nil
}

func (c *Client) Post(path string, body []byte) ([]byte, error) {
	respBody, statusCode, err := c.Do("POST", path, body)
	if err != nil {
		return nil, err
	}
	if statusCode >= 300 {
		return nil, fmt.Errorf("wechat api error (status=%d): %s", statusCode, string(respBody))
	}
	return respBody, nil
}

func (c *Client) Get(path string, body []byte) ([]byte, error) {
	respBody, statusCode, err := c.Do("GET", path, body)
	if err != nil {
		return nil, err
	}
	if statusCode >= 300 {
		return nil, fmt.Errorf("wechat api error (status=%d): %s", statusCode, string(respBody))
	}
	return respBody, nil
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	letterLen := big.NewInt(int64(len(letters)))
	for i := range b {
		idx, _ := rand.Int(rand.Reader, letterLen)
		b[i] = letters[idx.Int64()]
	}
	return string(b)
}
