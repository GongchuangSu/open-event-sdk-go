package kso

import (
	"testing"
)

func TestMd5(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "empty string",
			input: "",
			want:  "d41d8cd98f00b204e9800998ecf8427e",
		},
		{
			name:  "hello",
			input: "hello",
			want:  "5d41402abc4b2a76b9719d911017c592",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Md5(tt.input)
			if got != tt.want {
				t.Errorf("Md5() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMd5_Length(t *testing.T) {
	// MD5 输出应该是 32 字符的 hex 字符串
	result := Md5("any_input")
	if len(result) != 32 {
		t.Errorf("Md5() length = %v, want 32", len(result))
	}
}

func TestHmacSha256(t *testing.T) {
	tests := []struct {
		name    string
		message string
		secret  string
		want    string
	}{
		{
			name:    "basic test",
			message: "hello",
			secret:  "secret",
			want:    "iKqz7ejTrflNJquQ07r9SiCDBww7zOnAFO4EpEOEfAs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HmacSha256(tt.message, tt.secret)
			if got != tt.want {
				t.Errorf("HmacSha256() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHmacSha256_Consistency(t *testing.T) {
	// 验证相同输入产生相同输出
	message := "test message"
	secret := "test secret"

	result1 := HmacSha256(message, secret)
	result2 := HmacSha256(message, secret)

	if result1 != result2 {
		t.Errorf("HmacSha256 is not consistent: %v != %v", result1, result2)
	}

	// 验证不同输入产生不同输出
	result3 := HmacSha256(message, "different_secret")
	if result1 == result3 {
		t.Error("HmacSha256 should produce different results for different secrets")
	}
}

func TestVerifySignature(t *testing.T) {
	// 使用已知的签名进行测试
	accessKey := "test_access_key"
	secretKey := "test_secret_key"
	topic := "user.created"
	nonce := "abc123def456ghij"
	time := int64(1234567890)
	encryptedData := "encrypted_data_here"

	// 计算预期签名
	content := "test_access_key:user.created:abc123def456ghij:1234567890:encrypted_data_here"
	expectedSignature := HmacSha256(content, secretKey)

	tests := []struct {
		name      string
		params    *VerifySignatureParams
		wantError bool
	}{
		{
			name: "valid signature",
			params: &VerifySignatureParams{
				AccessKey:     accessKey,
				SecretKey:     secretKey,
				Topic:         topic,
				Nonce:         nonce,
				Time:          time,
				EncryptedData: encryptedData,
				Signature:     expectedSignature,
			},
			wantError: false,
		},
		{
			name: "invalid signature",
			params: &VerifySignatureParams{
				AccessKey:     accessKey,
				SecretKey:     secretKey,
				Topic:         topic,
				Nonce:         nonce,
				Time:          time,
				EncryptedData: encryptedData,
				Signature:     "invalid_signature",
			},
			wantError: true,
		},
		{
			name: "wrong secret key",
			params: &VerifySignatureParams{
				AccessKey:     accessKey,
				SecretKey:     "wrong_secret_key",
				Topic:         topic,
				Nonce:         nonce,
				Time:          time,
				EncryptedData: encryptedData,
				Signature:     expectedSignature,
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := VerifySignature(tt.params)
			if (err != nil) != tt.wantError {
				t.Errorf("VerifySignature() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestEncryptDecrypt(t *testing.T) {
	secretKey := "my_secret_key_12"
	nonce := "1234567890123456" // 16 bytes
	plaintext := `{"user_id":"123","name":"test"}`

	// 计算 cipher
	cipher := Md5(secretKey)

	// 加密
	encrypted, err := EncryptAesCbc([]byte(plaintext), cipher, nonce)
	if err != nil {
		t.Fatalf("EncryptAesCbc() error = %v", err)
	}

	// 解密
	decrypted, err := DecryptAesCbc(encrypted, cipher, nonce)
	if err != nil {
		t.Fatalf("DecryptAesCbc() error = %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("DecryptAesCbc() = %v, want %v", decrypted, plaintext)
	}
}

func TestDecrypt(t *testing.T) {
	secretKey := "my_secret_key_12"
	nonce := "1234567890123456"
	plaintext := `{"user_id":"123"}`

	// 先加密
	cipher := Md5(secretKey)
	encrypted, err := EncryptAesCbc([]byte(plaintext), cipher, nonce)
	if err != nil {
		t.Fatalf("EncryptAesCbc() error = %v", err)
	}

	// 使用 Decrypt 函数解密
	decrypted, err := Decrypt(&DecryptParams{
		SecretKey:     secretKey,
		EncryptedData: encrypted,
		Nonce:         nonce,
	})
	if err != nil {
		t.Fatalf("Decrypt() error = %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("Decrypt() = %v, want %v", decrypted, plaintext)
	}
}

func TestDecrypt_InvalidBase64(t *testing.T) {
	_, err := Decrypt(&DecryptParams{
		SecretKey:     "secret",
		EncryptedData: "invalid_base64!!!",
		Nonce:         "1234567890123456",
	})

	if err == nil {
		t.Error("expected error for invalid base64")
	}
}

func TestPkcs7Padding(t *testing.T) {
	tests := []struct {
		name      string
		input     []byte
		blockSize int
		wantLen   int
	}{
		{
			name:      "empty input",
			input:     []byte{},
			blockSize: 16,
			wantLen:   16,
		},
		{
			name:      "input length equals block size",
			input:     make([]byte, 16),
			blockSize: 16,
			wantLen:   32, // 需要额外一个块的填充
		},
		{
			name:      "input length less than block size",
			input:     []byte("hello"),
			blockSize: 16,
			wantLen:   16,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pkcs7Padding(tt.input, tt.blockSize)
			if len(got) != tt.wantLen {
				t.Errorf("pkcs7Padding() len = %v, want %v", len(got), tt.wantLen)
			}
		})
	}
}

func TestPkcs7UnPadding(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		wantLen int
	}{
		{
			name:    "empty input",
			input:   []byte{},
			wantLen: 0,
		},
		{
			name:    "padded data",
			input:   []byte{'h', 'e', 'l', 'l', 'o', 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11},
			wantLen: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pkcs7UnPadding(tt.input)
			if len(got) != tt.wantLen {
				t.Errorf("pkcs7UnPadding() len = %v, want %v", len(got), tt.wantLen)
			}
		})
	}
}
