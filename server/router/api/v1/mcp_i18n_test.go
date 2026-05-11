package v1

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetLocalizedMessage(t *testing.T) {
	tests := []struct {
		name           string
		code           string
		acceptLanguage string
		expected       string
	}{
		{
			name:           "中文错误消息",
			code:           "MEMO_NOT_FOUND",
			acceptLanguage: "zh-CN",
			expected:       "笔记不存在，请检查 ID 是否正确",
		},
		{
			name:           "英文错误消息",
			code:           "MEMO_NOT_FOUND",
			acceptLanguage: "en-US",
			expected:       "Memo not found, please check the ID",
		},
		{
			name:           "不支持的语言降级到英文",
			code:           "MEMO_NOT_FOUND",
			acceptLanguage: "fr-FR",
			expected:       "Memo not found, please check the ID",
		},
		{
			name:           "默认英文",
			code:           "MEMO_NOT_FOUND",
			acceptLanguage: "",
			expected:       "Memo not found, please check the ID",
		},
		{
			name:           "认证错误中文",
			code:           "AUTHENTICATION_REQUIRED",
			acceptLanguage: "zh",
			expected:       "需要登录才能使用此功能",
		},
		{
			name:           "权限错误英文",
			code:           "PERMISSION_DENIED",
			acceptLanguage: "en",
			expected:       "Permission denied to access this memo",
		},
		{
			name:           "未知错误代码",
			code:           "UNKNOWN_ERROR_CODE",
			acceptLanguage: "zh-CN",
			expected:       "An error occurred",
		},
		{
			name:           "内部错误中文",
			code:           "INTERNAL_ERROR",
			acceptLanguage: "zh-TW",
			expected:       "服务器内部错误，请稍后重试",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &http.Request{
				Header: http.Header{"Accept-Language": []string{tt.acceptLanguage}},
			}
			result := getLocalizedMessage(tt.code, req)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseAcceptLanguage(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "单个语言",
			input:    "zh-CN",
			expected: []string{"zh-CN"},
		},
		{
			name:     "多个语言",
			input:    "zh-CN,en-US;q=0.9,en;q=0.8",
			expected: []string{"zh-CN", "en-US", "en"},
		},
		{
			name:     "带空格",
			input:    "zh-CN, en-US",
			expected: []string{"zh-CN", "en-US"},
		},
		{
			name:     "复杂优先级",
			input:    "zh-CN;q=0.9,en-US;q=0.8,en;q=0.7",
			expected: []string{"zh-CN", "en-US", "en"},
		},
		{
			name:     "单个英文",
			input:    "en",
			expected: []string{"en"},
		},
		{
			name:     "空字符串",
			input:    "",
			expected: []string{},
		},
		{
			name:     "只有逗号",
			input:    ",",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseAcceptLanguage(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetLocalizedErrorResponse(t *testing.T) {
	tests := []struct {
		name           string
		code           string
		acceptLanguage string
		expectedCode   string
		expectMessage  bool
	}{
		{
			name:           "中文错误响应",
			code:           "MEMO_NOT_FOUND",
			acceptLanguage: "zh-CN",
			expectedCode:   "MEMO_NOT_FOUND",
			expectMessage:  true,
		},
		{
			name:           "英文错误响应",
			code:           "AUTHENTICATION_REQUIRED",
			acceptLanguage: "en-US",
			expectedCode:   "AUTHENTICATION_REQUIRED",
			expectMessage:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &http.Request{
				Header: http.Header{"Accept-Language": []string{tt.acceptLanguage}},
			}
			result := getLocalizedErrorResponse(tt.code, req)

			assert.Equal(t, tt.expectedCode, result["code"])
			assert.NotNil(t, result["message"])
			if tt.expectMessage {
				assert.NotEmpty(t, result["message"])
			}
		})
	}
}
