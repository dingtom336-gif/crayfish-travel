package identity

import (
	"testing"
)

const testKey = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

func TestEncryptDecryptRoundtrip(t *testing.T) {
	enc, err := NewEncryptor(testKey)
	if err != nil {
		t.Fatalf("new encryptor: %v", err)
	}

	tests := []string{
		"Zhang San",
		"110101199001011234",
		"13800138000",
		"",
		"Unicode test: 张三丰 李四",
	}

	for _, plaintext := range tests {
		ct, nonce, err := enc.EncryptString(plaintext)
		if err != nil {
			t.Errorf("encrypt %q: %v", plaintext, err)
			continue
		}

		got, err := enc.DecryptString(ct, nonce)
		if err != nil {
			t.Errorf("decrypt %q: %v", plaintext, err)
			continue
		}

		if got != plaintext {
			t.Errorf("roundtrip mismatch: got %q, want %q", got, plaintext)
		}
	}
}

func TestEncryptProducesDifferentCiphertext(t *testing.T) {
	enc, err := NewEncryptor(testKey)
	if err != nil {
		t.Fatal(err)
	}

	ct1, _, _ := enc.EncryptString("same input")
	ct2, _, _ := enc.EncryptString("same input")

	if string(ct1) == string(ct2) {
		t.Error("two encryptions of same plaintext produced identical ciphertext (nonce reuse)")
	}
}

func TestDecryptWithWrongNonce(t *testing.T) {
	enc, err := NewEncryptor(testKey)
	if err != nil {
		t.Fatal(err)
	}

	ct, _, err := enc.EncryptString("secret")
	if err != nil {
		t.Fatal(err)
	}

	wrongNonce := make([]byte, 12)
	_, err = enc.DecryptString(ct, wrongNonce)
	if err == nil {
		t.Error("expected error decrypting with wrong nonce")
	}
}

func TestInvalidKeyLength(t *testing.T) {
	_, err := NewEncryptor("short")
	if err == nil {
		t.Error("expected error for short key")
	}

	_, err = NewEncryptor("zzzz")
	if err == nil {
		t.Error("expected error for non-hex key")
	}
}

func TestSealWithNonce_Roundtrip(t *testing.T) {
	enc, err := NewEncryptor(testKey)
	if err != nil {
		t.Fatal(err)
	}

	tests := []string{
		"Zhang San",
		"110101199001011234",
		"13800138000",
		"Unicode: 张三丰",
	}

	for _, plaintext := range tests {
		sealed, err := enc.SealStringWithNonce(plaintext)
		if err != nil {
			t.Errorf("seal %q: %v", plaintext, err)
			continue
		}

		got, err := enc.OpenStringWithNonce(sealed)
		if err != nil {
			t.Errorf("open %q: %v", plaintext, err)
			continue
		}

		if got != plaintext {
			t.Errorf("roundtrip mismatch: got %q, want %q", got, plaintext)
		}
	}
}

func TestSealWithNonce_EachFieldIndependent(t *testing.T) {
	enc, err := NewEncryptor(testKey)
	if err != nil {
		t.Fatal(err)
	}

	// Simulate the repository pattern: each field sealed independently
	name, _ := enc.SealStringWithNonce("Zhang San")
	idNum, _ := enc.SealStringWithNonce("110101199001011234")
	phone, _ := enc.SealStringWithNonce("13800138000")

	// Each can be decrypted independently
	gotName, err := enc.OpenStringWithNonce(name)
	if err != nil || gotName != "Zhang San" {
		t.Errorf("name decrypt failed: %v", err)
	}
	gotID, err := enc.OpenStringWithNonce(idNum)
	if err != nil || gotID != "110101199001011234" {
		t.Errorf("id decrypt failed: %v", err)
	}
	gotPhone, err := enc.OpenStringWithNonce(phone)
	if err != nil || gotPhone != "13800138000" {
		t.Errorf("phone decrypt failed: %v", err)
	}
}
