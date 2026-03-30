package identity

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
)

// Encryptor handles AES-256-GCM encryption of PII fields.
type Encryptor struct {
	gcm cipher.AEAD
}

// NewEncryptor creates an AES-256-GCM encryptor from a hex-encoded 32-byte key.
func NewEncryptor(hexKey string) (*Encryptor, error) {
	key, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, fmt.Errorf("invalid hex key: %w", err)
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("key must be 32 bytes, got %d", len(key))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("aes cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("gcm: %w", err)
	}

	return &Encryptor{gcm: gcm}, nil
}

// Encrypt encrypts plaintext and returns (ciphertext, nonce).
func (e *Encryptor) Encrypt(plaintext []byte) (ciphertext []byte, nonce []byte, err error) {
	nonce = make([]byte, e.gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, fmt.Errorf("nonce generation: %w", err)
	}

	ciphertext = e.gcm.Seal(nil, nonce, plaintext, nil)
	return ciphertext, nonce, nil
}

// Decrypt decrypts ciphertext using the provided nonce.
func (e *Encryptor) Decrypt(ciphertext, nonce []byte) ([]byte, error) {
	plaintext, err := e.gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypt: %w", err)
	}
	return plaintext, nil
}

// EncryptString is a convenience wrapper for string encryption.
func (e *Encryptor) EncryptString(s string) (ciphertext []byte, nonce []byte, err error) {
	return e.Encrypt([]byte(s))
}

// DecryptString is a convenience wrapper for string decryption.
func (e *Encryptor) DecryptString(ciphertext, nonce []byte) (string, error) {
	b, err := e.Decrypt(ciphertext, nonce)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// SealWithNonce encrypts and prepends the nonce to the ciphertext.
// Output format: [nonce || ciphertext]. Each field is self-contained.
func (e *Encryptor) SealWithNonce(plaintext []byte) ([]byte, error) {
	nonce := make([]byte, e.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("nonce generation: %w", err)
	}
	sealed := e.gcm.Seal(nonce, nonce, plaintext, nil)
	return sealed, nil
}

// OpenWithNonce extracts the nonce from the front of the sealed blob and decrypts.
func (e *Encryptor) OpenWithNonce(sealed []byte) ([]byte, error) {
	nonceSize := e.gcm.NonceSize()
	if len(sealed) < nonceSize {
		return nil, fmt.Errorf("sealed data too short")
	}
	nonce, ciphertext := sealed[:nonceSize], sealed[nonceSize:]
	return e.gcm.Open(nil, nonce, ciphertext, nil)
}

// SealStringWithNonce encrypts a string with an embedded nonce.
func (e *Encryptor) SealStringWithNonce(s string) ([]byte, error) {
	return e.SealWithNonce([]byte(s))
}

// OpenStringWithNonce decrypts a nonce-prefixed blob to a string.
func (e *Encryptor) OpenStringWithNonce(sealed []byte) (string, error) {
	b, err := e.OpenWithNonce(sealed)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
