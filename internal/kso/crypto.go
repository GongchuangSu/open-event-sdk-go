package kso

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
)

var (
	// ErrInvalidSignature 签名验证失败
	ErrInvalidSignature = errors.New("invalid signature")

	// ErrDecryptFailed 解密失败
	ErrDecryptFailed = errors.New("decrypt failed")

	// errCipherTextTooShort 密文太短
	errCipherTextTooShort = errors.New("cipher text too short")

	// errCipherTextInvalidLength 密文长度不是块大小的倍数
	errCipherTextInvalidLength = errors.New("cipher text is not a multiple of the block size")
)

// VerifySignatureParams 签名验证参数
type VerifySignatureParams struct {
	AccessKey     string // 应用 AccessKey (AppId)
	SecretKey     string // 应用 SecretKey (AppSecret)
	Topic         string // 消息主题
	Nonce         string // 随机数/iv向量
	Time          int64  // 时间戳（秒）
	EncryptedData string // 加密数据
	Signature     string // 待验证的签名
}

// VerifySignature 验证消息签名
//
// 签名算法：
//  1. 计算签名原文 content = access_key:topic:nonce:time:encrypted_data
//  2. 使用 HMAC-SHA256(content, secret_key) 计算签名
//  3. 签名使用 URL 安全的无填充 base64 编码
func VerifySignature(p *VerifySignatureParams) error {
	// 构建签名原文
	content := fmt.Sprintf("%s:%s:%s:%d:%s",
		p.AccessKey, p.Topic, p.Nonce, p.Time, p.EncryptedData)

	// 计算签名
	expectedSignature := hmacSha256(content, p.SecretKey)

	// 比较签名
	if !hmac.Equal([]byte(expectedSignature), []byte(p.Signature)) {
		return ErrInvalidSignature
	}

	return nil
}

// DecryptParams 解密参数
type DecryptParams struct {
	SecretKey     string // 应用 SecretKey
	EncryptedData string // 加密数据（标准 base64 编码）
	Nonce         string // iv 向量
}

// Decrypt 解密事件数据
//
// 解密算法：
//  1. encrypted_data 使用标准的有填充 base64 编码，先进行 base64 解码
//  2. cipher = md5(secretKey)
//  3. 使用 AES-CBC 解密，iv 为 nonce 的前 16 字节
//  4. 解密后的数据经过 PKCS7 填充，需要移除填充
func Decrypt(p *DecryptParams) (string, error) {
	// 计算 cipher = md5(secretKey)
	cipherKey := md5Hash(p.SecretKey)

	// 解密数据
	decrypted, err := decryptAesCbc(p.EncryptedData, cipherKey, p.Nonce)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrDecryptFailed, err)
	}

	return decrypted, nil
}

// decryptAesCbc 使用 AES-CBC 模式解密数据
func decryptAesCbc(encryptedData, cipherKey, nonce string) (string, error) {
	// base64 解码
	data, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return "", fmt.Errorf("base64 decode failed: %w", err)
	}

	// AES-CBC 解密
	rawData, err := aesCbcPkcs7Decrypt(data, []byte(cipherKey), []byte(nonce))
	if err != nil {
		return "", err
	}

	return string(rawData), nil
}

// aesCbcPkcs7Decrypt AES-CBC 解密（PKCS7 填充）
func aesCbcPkcs7Decrypt(encryptData, key, nonce []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	blockSize := block.BlockSize()
	if len(encryptData) < blockSize {
		return nil, errCipherTextTooShort
	}
	if len(encryptData)%blockSize != 0 {
		return nil, errCipherTextInvalidLength
	}

	iv := nonce[:blockSize]
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(encryptData, encryptData)

	encryptData = pkcs7UnPadding(encryptData)
	return encryptData, nil
}

// pkcs7UnPadding PKCS7 去填充
func pkcs7UnPadding(origData []byte) []byte {
	length := len(origData)
	if length == 0 {
		return origData
	}
	unPadding := int(origData[length-1])
	if unPadding > length {
		return origData
	}
	return origData[:(length - unPadding)]
}

// md5Hash 计算字符串的 MD5 值（返回 32 字符的 hex 字符串）
func md5Hash(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

// hmacSha256 计算 HMAC-SHA256 签名
// 返回 URL 安全的无填充 base64 编码
func hmacSha256(message, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}

// ----- 以下函数用于测试 -----

// Md5 计算字符串的 MD5 值（导出用于测试）
func Md5(s string) string {
	return md5Hash(s)
}

// HmacSha256 计算 HMAC-SHA256 签名（导出用于测试）
func HmacSha256(message, secret string) string {
	return hmacSha256(message, secret)
}

// EncryptAesCbc 使用 AES-CBC 模式加密数据（导出用于测试）
func EncryptAesCbc(rawData []byte, cipherKey, nonce string) (string, error) {
	data, err := aesCbcEncrypt(rawData, []byte(cipherKey), []byte(nonce))
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

// aesCbcEncrypt AES-CBC 加密
func aesCbcEncrypt(rawData, key, nonce []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	blockSize := block.BlockSize()
	rawData = pkcs7Padding(rawData, blockSize)
	cipherText := make([]byte, len(rawData))

	iv := nonce[:blockSize]
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(cipherText, rawData)

	return cipherText, nil
}

// pkcs7Padding PKCS7 填充
func pkcs7Padding(cipherText []byte, blockSize int) []byte {
	padding := blockSize - len(cipherText)%blockSize
	paddingText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(cipherText, paddingText...)
}
