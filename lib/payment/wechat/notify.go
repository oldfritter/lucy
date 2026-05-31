package wechat

import (
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"sync"
	"time"
)

type NotifyResource struct {
	Algorithm      string `json:"algorithm"`
	Ciphertext     string `json:"ciphertext"`
	AssociatedData string `json:"associated_data"`
	Nonce          string `json:"nonce"`
	OriginalType   string `json:"original_type"`
}

type NotifyRequest struct {
	ID           string         `json:"id"`
	CreateTime   string         `json:"create_time"`
	ResourceType string         `json:"resource_type"`
	EventType    string         `json:"event_type"`
	Resource     NotifyResource `json:"resource"`
}

type TransactionNotification struct {
	AppID         string `json:"appid"`
	MchID         string `json:"mchid"`
	OutTradeNo    string `json:"out_trade_no"`
	TransactionID string `json:"transaction_id"`
	TradeState    string `json:"trade_state"`
	TradeType     string `json:"trade_type"`
	SuccessTime   string `json:"success_time"`
	Payer         struct {
		OpenID string `json:"openid"`
	} `json:"payer"`
	Amount struct {
		Total    int    `json:"total"`
		Currency string `json:"currency"`
	} `json:"amount"`
}

type platformCertCache struct {
	mu        sync.RWMutex
	certs     map[string]*x509.Certificate
	expiresAt time.Time
}

var certCache = &platformCertCache{
	certs: make(map[string]*x509.Certificate),
}

func (c *Client) RefreshCertificates() error {
	body, err := c.Get("/v3/certificates", nil)
	if err != nil {
		return fmt.Errorf("get certificates: %w", err)
	}
	var result struct {
		Data []struct {
			SerialNo           string `json:"serial_no"`
			EffectiveTime      string `json:"effective_time"`
			ExpireTime         string `json:"expire_time"`
			EncryptCertificate struct {
				Algorithm      string `json:"algorithm"`
				Nonce          string `json:"nonce"`
				AssociatedData string `json:"associated_data"`
				Ciphertext     string `json:"ciphertext"`
			} `json:"encrypt_certificate"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("unmarshal certificates: %w", err)
	}
	certCache.mu.Lock()
	defer certCache.mu.Unlock()
	for _, item := range result.Data {
		certPEM, err := decryptAES256GCM(c.config.APIv3Key, item.EncryptCertificate.Nonce, item.EncryptCertificate.AssociatedData, item.EncryptCertificate.Ciphertext)
		if err != nil {
			return fmt.Errorf("decrypt certificate %s: %w", item.SerialNo, err)
		}
		block, _ := pem.Decode([]byte(certPEM))
		if block == nil {
			return fmt.Errorf("failed to parse certificate PEM for serial %s", item.SerialNo)
		}
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return fmt.Errorf("parse certificate %s: %w", item.SerialNo, err)
		}
		certCache.certs[item.SerialNo] = cert
	}
	certCache.expiresAt = time.Now().Add(12 * time.Hour)
	return nil
}

func getPlatformCert(serialNo string) (*x509.Certificate, error) {
	certCache.mu.RLock()
	defer certCache.mu.RUnlock()
	cert, ok := certCache.certs[serialNo]
	if !ok {
		return nil, fmt.Errorf("certificate not found for serial %s", serialNo)
	}
	return cert, nil
}

func VerifyNotifySignature(timestamp, nonce, signature string, body []byte) error {
	certCache.mu.RLock()
	certs := make(map[string]*x509.Certificate, len(certCache.certs))
	for k, v := range certCache.certs {
		certs[k] = v
	}
	certCache.mu.RUnlock()
	signStr := fmt.Sprintf("%s\n%s\n%s\n", timestamp, nonce, string(body))
	sigBytes, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return fmt.Errorf("decode signature: %w", err)
	}
	h := sha256.New()
	h.Write([]byte(signStr))
	digest := h.Sum(nil)
	for serial, cert := range certs {
		pubKey, ok := cert.PublicKey.(*rsa.PublicKey)
		if !ok {
			continue
		}
		if err := rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, digest, sigBytes); err == nil {
			_ = serial
			return nil
		}
	}
	return fmt.Errorf("signature verification failed")
}

func VerifyNotifySignatureWithSerial(serialNo, timestamp, nonce, signature string, body []byte) error {
	cert, err := getPlatformCert(serialNo)
	if err != nil {
		return err
	}
	signStr := fmt.Sprintf("%s\n%s\n%s\n", timestamp, nonce, string(body))
	sigBytes, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return fmt.Errorf("decode signature: %w", err)
	}
	h := sha256.New()
	h.Write([]byte(signStr))
	digest := h.Sum(nil)
	pubKey, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		return fmt.Errorf("certificate public key is not RSA")
	}
	return rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, digest, sigBytes)
}

func DecryptNotifyResource(resource NotifyResource, apiV3Key string) ([]byte, error) {
	return decryptAES256GCMBytes(apiV3Key, resource.Nonce, resource.AssociatedData, resource.Ciphertext)
}

func decryptAES256GCM(key, nonce, associatedData, ciphertext string) (string, error) {
	data, err := decryptAES256GCMBytes(key, nonce, associatedData, ciphertext)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func decryptAES256GCMBytes(keyStr, nonce, associatedData, ciphertext string) ([]byte, error) {
	key := []byte(keyStr)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("new cipher: %w", err)
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("new gcm: %w", err)
	}
	cipherBytes, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return nil, fmt.Errorf("decode ciphertext: %w", err)
	}
	plaintext, err := aesGCM.Open(nil, []byte(nonce), cipherBytes, []byte(associatedData))
	if err != nil {
		return nil, fmt.Errorf("decrypt: %w", err)
	}
	return plaintext, nil
}
