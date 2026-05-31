package wechat

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"sync"

	"github.com/oldfritter/lucy/base"
)

// Config 微信支付配置
type Config struct {
	AppID           string
	MchID           string
	APIv3Key        string
	SerialNo        string
	PrivateKey      *rsa.PrivateKey
	NotifyURL       string
	RefundNotifyURL string
}

var (
	wechatConfig *Config
	once         sync.Once
)

// GetConfig 获取微信支付配置（单例）
func GetConfig() *Config {
	once.Do(func() {
		cfg := base.GetConfig()
		appID := cfg.Get("wechat.app_id", "")
		mchID := cfg.Get("wechat.mch_id", "")
		apiV3Key := cfg.Get("wechat.api_v3_key", "")
		serialNo := cfg.Get("wechat.serial_no", "")
		privateKeyPath := cfg.Get("wechat.private_key_path", "")
		notifyURL := cfg.Get("wechat.notify_url", "")
		refundNotifyURL := cfg.Get("wechat.refund_notify_url", "")

		var privateKey *rsa.PrivateKey
		if privateKeyPath != "" {
			keyData, err := os.ReadFile(privateKeyPath)
			if err != nil {
				panic("wechat: failed to read private key: " + err.Error())
			}
			privateKey, err = parsePrivateKey(keyData)
			if err != nil {
				panic("wechat: failed to parse private key: " + err.Error())
			}
		}

		wechatConfig = &Config{
			AppID:           appID,
			MchID:           mchID,
			APIv3Key:        apiV3Key,
			SerialNo:        serialNo,
			PrivateKey:      privateKey,
			NotifyURL:       notifyURL,
			RefundNotifyURL: refundNotifyURL,
		}
	})
	return wechatConfig
}

func parsePrivateKey(data []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return x509.ParsePKCS1PrivateKey(data)
	}
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}
