// Package kso 提供 KSO-1 签名相关工具
package kso

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"
)

const (
	// Kso1Type 签名类型标识
	Kso1Type = "KSO-1"

	// HeaderKsoDate 日期请求头
	HeaderKsoDate = "X-Kso-Date"

	// HeaderKsoAuthorization 授权请求头
	HeaderKsoAuthorization = "X-Kso-Authorization"

	// HeaderContentType Content-Type 请求头
	HeaderContentType = "Content-Type"
)

// SignParams 签名参数
type SignParams struct {
	AppId       string // 应用 ID
	AppSecret   string // 应用密钥
	Method      string // HTTP 方法
	RequestURI  string // 请求 URI（包含 path 和 query）
	ContentType string // Content-Type（可选）
	Body        []byte // 请求体（可选）
	Date        string // 请求时间（可选，为空时自动生成）
}

// SignResult 签名结果
type SignResult struct {
	Date          string // X-Kso-Date 头的值
	Authorization string // X-Kso-Authorization 头的值
}

// Sign 计算 KSO-1 签名
//
// 签名算法：
//  1. 计算 body 的 SHA256（body 为空时为空字符串）
//  2. 构建待签名字符串：KSO-1 + method + uri + contentType + date + sha256(body)
//  3. 使用 HMAC-SHA256(appSecret, stringToSign) 计算签名
//  4. 返回 Authorization 格式：KSO-1 {app_id}:{hex_signature}
func Sign(p *SignParams) (*SignResult, error) {
	if p.AppId == "" {
		return nil, fmt.Errorf("app_id is required")
	}
	if p.AppSecret == "" {
		return nil, fmt.Errorf("app_secret is required")
	}
	if p.Method == "" {
		return nil, fmt.Errorf("method is required")
	}
	if p.RequestURI == "" {
		return nil, fmt.Errorf("request_uri is required")
	}

	// 处理日期：如果为空，使用当前 UTC 时间
	date := p.Date
	if date == "" {
		date = time.Now().UTC().Format(http.TimeFormat)
	}

	// 1. 计算 body 的 SHA256 值
	// body 为空时，sha256Hex 为空字符串
	sha256Hex := ""
	if len(p.Body) > 0 {
		h := sha256.New()
		h.Write(p.Body)
		sha256Hex = hex.EncodeToString(h.Sum(nil))
	}

	// 2. 构建待签名字符串
	// 格式：KSO-1 + method + uri + contentType + date + sha256(body)
	stringToSign := Kso1Type + p.Method + p.RequestURI + p.ContentType + date + sha256Hex

	// 3. 使用 HMAC-SHA256 计算签名
	mac := hmac.New(sha256.New, []byte(p.AppSecret))
	mac.Write([]byte(stringToSign))
	signature := hex.EncodeToString(mac.Sum(nil))

	// 4. 构建 Authorization 头
	authorization := fmt.Sprintf("%s %s:%s", Kso1Type, p.AppId, signature)

	return &SignResult{
		Date:          date,
		Authorization: authorization,
	}, nil
}

// NewHeaders 生成包含 KSO-1 签名的 HTTP 请求头
// 这是 Sign 的便捷封装，直接返回 http.Header
func NewHeaders(appId, appSecret, method, uri, contentType string, body []byte) (http.Header, error) {
	result, err := Sign(&SignParams{
		AppId:       appId,
		AppSecret:   appSecret,
		Method:      method,
		RequestURI:  uri,
		ContentType: contentType,
		Body:        body,
	})
	if err != nil {
		return nil, err
	}

	headers := make(http.Header)
	headers.Set(HeaderKsoDate, result.Date)
	headers.Set(HeaderKsoAuthorization, result.Authorization)
	if contentType != "" {
		headers.Set(HeaderContentType, contentType)
	}

	return headers, nil
}

// SignForWebSocket 为 WebSocket 连接生成签名头
// WebSocket 握手没有 body，Content-Type 通常为空
func SignForWebSocket(appId, appSecret, uri string) (http.Header, error) {
	return NewHeaders(appId, appSecret, http.MethodGet, uri, "", nil)
}
