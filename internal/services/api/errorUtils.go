package api

import (
	"fmt"
	"net"
	"reflect"
	"strings"
)

// SSL_ERROR_CODES SSL/TLS错误代码集合
// 用于识别证书验证失败等SSL相关错误
var SSL_ERROR_CODES = map[string]bool{
	"UNABLE_TO_VERIFY_LEAF_SIGNATURE":             true,
	"UNABLE_TO_GET_ISSUER_CERT":                   true,
	"UNABLE_TO_GET_ISSUER_CERT_LOCALLY":           true,
	"CERT_SIGNATURE_FAILURE":                      true,
	"CERT_NOT_YET_VALID":                          true,
	"CERT_HAS_EXPIRED":                            true,
	"CERT_REVOKED":                                true,
	"CERT_REJECTED":                               true,
	"CERT_UNTRUSTED":                              true,
	"DEPTH_ZERO_SELF_SIGNED_CERT":                 true,
	"SELF_SIGNED_CERT_IN_CHAIN":                   true,
	"CERT_CHAIN_TOO_LONG":                         true,
	"PATH_LENGTH_EXCEEDED":                        true,
	"ERR_TLS_CERT_ALTNAME_INVALID":                true,
	"HOSTNAME_MISMATCH":                           true,
	"ERR_TLS_HANDSHAKE_TIMEOUT":                   true,
	"ERR_SSL_WRONG_VERSION_NUMBER":                true,
	"ERR_SSL_DECRYPTION_FAILED_OR_BAD_RECORD_MAC": true,
}

// ConnectionErrorDetails 连接错误详情
type ConnectionErrorDetails struct {
	Code       string
	Message    string
	IsSSLError bool
}

// ExtractConnectionErrorDetails 从错误链中提取连接错误详情
// 遍历错误的cause链来查找根错误代码和消息
func ExtractConnectionErrorDetails(err error) *ConnectionErrorDetails {
	if err == nil {
		return nil
	}

	current := err
	maxDepth := 5
	depth := 0

	for current != nil && depth < maxDepth {
		if code := getErrorCode(current); code != "" {
			isSSL := SSL_ERROR_CODES[code]
			return &ConnectionErrorDetails{
				Code:       code,
				Message:    current.Error(),
				IsSSLError: isSSL,
			}
		}

		// 移动到下一个cause
		if ne, ok := current.(interface{ Cause() error }); ok {
			cause := ne.Cause()
			if cause == current {
				break
			}
			current = cause
			depth++
		} else {
			break
		}
	}

	return nil
}

// getErrorCode 获取错误的代码
func getErrorCode(err error) string {
	if err == nil {
		return ""
	}

	// 尝试从各种错误类型中提取code
	switch e := err.(type) {
	case interface{ Code() string }:
		return e.Code()
	case interface{ Unwrap() error }:
		return getErrorCode(e.Unwrap())
	}

	// 使用反射检查code字段
	if code := extractErrorCode(err); code != "" {
		return code
	}

	// 使用net.Error检查
	if ne, ok := err.(net.Error); ok {
		if ne.Timeout() {
			return "ETIMEDOUT"
		}
		if ne.Temporary() {
			return "ECONNRESET"
		}
	}

	return ""
}

// extractErrorCode 使用反射提取错误的code字段
func extractErrorCode(err error) string {
	if err == nil {
		return ""
	}

	// 使用反射获取code字段
	val := reflect.ValueOf(err)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return ""
	}

	codeField := val.FieldByName("Code")
	if codeField.IsValid() && codeField.Kind() == reflect.String {
		return codeField.String()
	}

	// 检查code字段(小写)
	codeField = val.FieldByName("code")
	if codeField.IsValid() && codeField.Kind() == reflect.String {
		return codeField.String()
	}

	return ""
}

// GetSSLErrorHint 获取SSL错误的处理建议
// 针对TLS拦截代理(Zscaler等)导致的问题
func GetSSLErrorHint(err error) string {
	details := ExtractConnectionErrorDetails(err)
	if details == nil || !details.IsSSLError {
		return ""
	}
	return fmt.Sprintf("SSL certificate error (%s). If you are behind a corporate proxy or TLS-intercepting firewall, set NODE_EXTRA_CA_CERTS to your CA bundle path, or ask IT to allowlist *.anthropic.com. Run /doctor for details.", details.Code)
}

// sanitizeMessageHTML 清理HTML内容
// 从消息字符串中剥离HTML内容(如CloudFlare错误页面)
func SanitizeMessageHTML(message string) string {
	if strings.Contains(message, "<!DOCTYPE html") || strings.Contains(message, "<html") {
		// 尝试提取title
		titleMatch := findTitleInHTML(message)
		if titleMatch != "" {
			return titleMatch
		}
		return ""
	}
	return message
}

// findTitleInHTML 从HTML中提取title
func findTitleInHTML(html string) string {
	start := strings.Index(html, "<title>")
	if start == -1 {
		return ""
	}
	start += 7
	end := strings.Index(html[start:], "</title>")
	if end == -1 {
		return ""
	}
	return strings.TrimSpace(html[start : start+end])
}

// FormatAPIError 格式化API错误消息
func FormatAPIError(statusCode int, message string) string {
	if message == "" {
		message = "Unknown error"
	}

	switch statusCode {
	case 400:
		return fmt.Sprintf("Bad request: %s", message)
	case 401:
		return "Authentication failed. Please check your API key."
	case 403:
		return "Access forbidden. You may not have permission for this operation."
	case 404:
		return fmt.Sprintf("Resource not found: %s", message)
	case 429:
		return "Rate limit exceeded. Please wait and try again."
	case 500:
		return "Internal server error. Please try again later."
	case 503:
		return "Service unavailable. Please try again later."
	default:
		if statusCode >= 500 {
			return fmt.Sprintf("Server error (%d): %s", statusCode, message)
		}
		if statusCode >= 400 {
			return fmt.Sprintf("Request error (%d): %s", statusCode, message)
		}
		return message
	}
}

// IsRetryableError 判断错误是否可重试
func IsRetryableError(statusCode int) bool {
	// 5xx错误和429(速率限制)可重试
	return statusCode >= 500 || statusCode == 429
}

// GetErrorHint 获取错误的处理建议
func GetErrorHint(err error) string {
	details := ExtractConnectionErrorDetails(err)
	if details == nil {
		return ""
	}

	if details.Code == "ETIMEDOUT" {
		return "Request timed out. Check your internet connection and proxy settings."
	}

	if details.IsSSLError {
		switch details.Code {
		case "UNABLE_TO_VERIFY_LEAF_SIGNATURE", "UNABLE_TO_GET_ISSUER_CERT", "UNABLE_TO_GET_ISSUER_CERT_LOCALLY":
			return "Unable to connect to API: SSL certificate verification failed. Check your proxy or corporate SSL certificates."
		case "CERT_HAS_EXPIRED":
			return "Unable to connect to API: SSL certificate has expired."
		case "CERT_REVOKED":
			return "Unable to connect to API: SSL certificate has been revoked."
		case "DEPTH_ZERO_SELF_SIGNED_CERT", "SELF_SIGNED_CERT_IN_CHAIN":
			return "Unable to connect to API: Self-signed certificate detected. Check your proxy or corporate SSL certificates."
		case "ERR_TLS_CERT_ALTNAME_INVALID", "HOSTNAME_MISMATCH":
			return "Unable to connect to API: SSL certificate hostname mismatch."
		case "CERT_NOT_YET_VALID":
			return "Unable to connect to API: SSL certificate is not yet valid."
		default:
			return fmt.Sprintf("Unable to connect to API: SSL error (%s)", details.Code)
		}
	}

	if details.Message == "Connection error." {
		if details.Code != "" {
			return fmt.Sprintf("Unable to connect to API (%s)", details.Code)
		}
		return "Unable to connect to API. Check your internet connection."
	}

	return ""
}
