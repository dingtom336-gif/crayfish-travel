package payment

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

// PaymentResult holds the result of creating a payment.
type PaymentResult struct {
	OutTradeNo string `json:"out_trade_no"`
	QRCodeURL  string `json:"qr_code_url,omitempty"`
	VoiceToken string `json:"voice_token,omitempty"`
	Method     string `json:"method"`
}

// AlipayClient is the interface for Alipay payment operations.
type AlipayClient interface {
	// CreateQRPayment generates a QR code URL for mini program payment.
	CreateQRPayment(outTradeNo string, amountCents int64, subject string) (*PaymentResult, error)
	// CreateVoiceToken generates a "zhi kouling" (voice token) for sharing.
	CreateVoiceToken(outTradeNo string, amountCents int64, subject string) (*PaymentResult, error)
	// VerifyCallback validates the signature of a payment callback.
	VerifyCallback(params map[string]string) (bool, error)
}

// MockAlipayClient provides mock Alipay operations for development.
type MockAlipayClient struct{}

// NewMockAlipayClient creates a mock Alipay client.
func NewMockAlipayClient() *MockAlipayClient {
	return &MockAlipayClient{}
}

// CreateQRPayment returns a fake QR code URL.
func (m *MockAlipayClient) CreateQRPayment(outTradeNo string, amountCents int64, subject string) (*PaymentResult, error) {
	return &PaymentResult{
		OutTradeNo: outTradeNo,
		QRCodeURL:  fmt.Sprintf("https://qr.alipay.com/mock/%s?amount=%d", outTradeNo, amountCents),
		Method:     "qr",
	}, nil
}

// CreateVoiceToken returns a fake voice token (zhi kouling).
func (m *MockAlipayClient) CreateVoiceToken(outTradeNo string, amountCents int64, subject string) (*PaymentResult, error) {
	token := fmt.Sprintf("$crayfish_%s_%d$", outTradeNo[:8], rand.Intn(9999))
	return &PaymentResult{
		OutTradeNo: outTradeNo,
		VoiceToken: token,
		Method:     "voice_token",
	}, nil
}

// VerifyCallback always returns true in mock mode.
func (m *MockAlipayClient) VerifyCallback(_ map[string]string) (bool, error) {
	return true, nil
}

// GenerateOutTradeNo creates a unique trade number.
func GenerateOutTradeNo() string {
	ts := time.Now().Format("20060102150405")
	suffix := uuid.New().String()[:8]
	return fmt.Sprintf("CT%s%s", ts, suffix)
}
