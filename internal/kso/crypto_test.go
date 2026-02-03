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
		{
			name:  "test_secret",
			input: "test_secret",
			want:  "e989d46fdbc1c376c19a43aaf85227a4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Md5(tt.input); got != tt.want {
				t.Errorf("Md5() = %v, want %v", got, tt.want)
			}
		})
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
			message: "test_message",
			secret:  "test_secret",
			want:    "ZaIJF7XWibQHwbbgx6qd5AIh78SB_-WPJIXFHYIqzs4",
		},
		{
			name:    "empty message",
			message: "",
			secret:  "secret",
			want:    "-eZuF5tnR65UEI-C-K3os8Jddv0wr95sOVgixTAZYWk",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HmacSha256(tt.message, tt.secret); got != tt.want {
				t.Errorf("HmacSha256() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVerifySignature(t *testing.T) {
	// 测试数据
	accessKey := "test_access_key"
	secretKey := "test_secret_key"
	topic := "kso.drive.file"
	nonce := "1234567890123456"
	time := int64(1234567890)
	encryptedData := "encrypted_data_base64"

	// 计算正确的签名
	content := accessKey + ":" + topic + ":" + nonce + ":1234567890:" + encryptedData
	correctSignature := HmacSha256(content, secretKey)

	tests := []struct {
		name    string
		params  *VerifySignatureParams
		wantErr bool
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
				Signature:     correctSignature,
			},
			wantErr: false,
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
			wantErr: true,
		},
		{
			name: "wrong secret key",
			params: &VerifySignatureParams{
				AccessKey:     accessKey,
				SecretKey:     "wrong_secret",
				Topic:         topic,
				Nonce:         nonce,
				Time:          time,
				EncryptedData: encryptedData,
				Signature:     correctSignature,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := VerifySignature(tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("VerifySignature() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDecrypt(t *testing.T) {
	secretKey := "test_secret_key_1234"
	nonce := "1234567890123456"
	plainText := "hello world"

	// 计算 cipher key
	cipherKey := Md5(secretKey)

	// 加密数据
	encryptedData, err := EncryptAesCbc([]byte(plainText), cipherKey, nonce)
	if err != nil {
		t.Fatalf("EncryptAesCbc() error = %v", err)
	}

	// 解密
	decrypted, err := Decrypt(&DecryptParams{
		SecretKey:     secretKey,
		EncryptedData: encryptedData,
		Nonce:         nonce,
	})
	if err != nil {
		t.Fatalf("Decrypt() error = %v", err)
	}

	if decrypted != plainText {
		t.Errorf("Decrypt() = %v, want %v", decrypted, plainText)
	}
}

func TestDecrypt_InvalidBase64(t *testing.T) {
	_, err := Decrypt(&DecryptParams{
		SecretKey:     "test_secret",
		EncryptedData: "invalid_base64!!!",
		Nonce:         "1234567890123456",
	})
	if err == nil {
		t.Error("Decrypt() should return error for invalid base64")
	}
}

func TestEncryptDecryptRoundTrip(t *testing.T) {
	tests := []struct {
		name      string
		plainText string
	}{
		{
			name:      "short text",
			plainText: "hello",
		},
		{
			name:      "exact block size",
			plainText: "1234567890123456",
		},
		{
			name:      "longer text",
			plainText: "this is a longer text that spans multiple blocks",
		},
		{
			name:      "json data",
			plainText: `{"user_id":"123","name":"test"}`,
		},
	}

	secretKey := "test_secret_key_1234"
	nonce := "1234567890123456"
	cipherKey := Md5(secretKey)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 加密
			encrypted, err := EncryptAesCbc([]byte(tt.plainText), cipherKey, nonce)
			if err != nil {
				t.Fatalf("EncryptAesCbc() error = %v", err)
			}

			// 解密
			decrypted, err := Decrypt(&DecryptParams{
				SecretKey:     secretKey,
				EncryptedData: encrypted,
				Nonce:         nonce,
			})
			if err != nil {
				t.Fatalf("Decrypt() error = %v", err)
			}

			if decrypted != tt.plainText {
				t.Errorf("Round trip failed: got %v, want %v", decrypted, tt.plainText)
			}
		})
	}
}
