package alipay

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"sync"

	"github.com/oldfritter/lucy/base"
)

// Config 支付宝支付配置
type Config struct {
	AppID              string          // 应用AppID
	PrivateKey         *rsa.PrivateKey // 商户私钥
	AlipayPublicKey    *rsa.PublicKey  // 支付宝公钥（用于验证通知签名）
	NotifyURL          string          // 支付回调通知地址
	ReturnURL          string          // 同步跳转地址
	Gateway            string          // 网关地址
	SignType           string          // 签名算法
}

var (
	alipayConfig *Config
	once         sync.Once
)

// GetConfig 获取支付宝支付配置（单例）
func GetConfig() *Config {
	once.Do(func() {
		cfg := base.GetConfig()
		appID := cfg.Get("alipay.app_id", "")
		privateKeyPath := cfg.Get("alipay.private_key_path", "")
		alipayPublicKeyPath := cfg.Get("alipay.alipay_public_key_path", "")
		notifyURL := cfg.Get("alipay.notify_url", "")
		returnURL := cfg.Get("alipay.return_url", "")
		gateway := cfg.Get("alipay.gateway", "https://openapi.alipay.com/gateway.do")
		signType := cfg.Get("alipay.sign_type", "RSA2")

		if appID == "" {
			panic("alipay: app_id is required")
		}

		var privateKey *rsa.PrivateKey
		if privateKeyPath != "" {
			keyData, err := os.ReadFile(privateKeyPath)
			if err != nil {
				panic("alipay: failed to read private key: " + err.Error())
			}
			privateKey, err = parsePrivateKey(keyData)
			if err != nil {
				panic("alipay: failed to parse private key: " + err.Error())
			}
		}

		var alipayPublicKey *rsa.PublicKey
		if alipayPublicKeyPath != "" {
			keyData, err := os.ReadFile(alipayPublicKeyPath)
			if err != nil {
				panic("alipay: failed to read alipay public key: " + err.Error())
			}
			alipayPublicKey, err = parsePublicKey(keyData)
			if err != nil {
				panic("alipay: failed to parse alipay public key: " + err.Error())
			}
		}

		alipayConfig = &Config{
			AppID:           appID,
			PrivateKey:      privateKey,
			AlipayPublicKey: alipayPublicKey,
			NotifyURL:       notifyURL,
			ReturnURL:       returnURL,
			Gateway:         gateway,
			SignType:        signType,
		}
	})
	return alipayConfig
}

// parsePrivateKey 解析 PEM 格式的 RSA 私钥（PKCS1 或 PKCS8）
func parsePrivateKey(data []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, nil
	}

	// 尝试 PKCS1
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err == nil {
		return key, nil
	}

	// 尝试 PKCS8
	pkcs8Key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err == nil {
		if rsaKey, ok := pkcs8Key.(*rsa.PrivateKey); ok {
			return rsaKey, nil
		}
	}

	return nil, err
}

// parsePublicKey 解析 PEM 格式的 RSA 公钥
func parsePublicKey(data []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, nil
	}

	// 尝试 PKIX
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err == nil {
		if rsaPub, ok := pub.(*rsa.PublicKey); ok {
			return rsaPub, nil
		}
	}

	// 尝试 PKCS1
	return x509.ParsePKCS1PublicKey(block.Bytes)
}
