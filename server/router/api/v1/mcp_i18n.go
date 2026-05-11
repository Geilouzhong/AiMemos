package v1

import (
	"net/http"
	"strings"
)

// errorCodeMessages holds error messages for all supported languages.
var errorCodeMessages = map[string]map[string]string{
	"zh": {
		"AUTHENTICATION_REQUIRED": "需要登录才能使用此功能",
		"PERMISSION_DENIED":       "您没有权限访问此笔记",
		"MEMO_NOT_FOUND":          "笔记不存在，请检查 ID 是否正确",
		"INVALID_CONTENT":         "笔记内容不能为空",
		"INVALID_ARGUMENTS":       "参数无效",
		"INTERNAL_ERROR":          "服务器内部错误，请稍后重试",
		"UNKNOWN_TOOL":            "未知的工具",
		"UNKNOWN_METHOD":          "未知的方法",
		"INVALID_REQUEST":         "无效的请求",
	},
	"en": {
		"AUTHENTICATION_REQUIRED": "Authentication required",
		"PERMISSION_DENIED":       "Permission denied to access this memo",
		"MEMO_NOT_FOUND":          "Memo not found, please check the ID",
		"INVALID_CONTENT":         "Memo content cannot be empty",
		"INVALID_ARGUMENTS":       "Invalid arguments",
		"INTERNAL_ERROR":          "Internal server error, please try again later",
		"UNKNOWN_TOOL":            "Unknown tool",
		"UNKNOWN_METHOD":          "Unknown method",
		"INVALID_REQUEST":         "Invalid request",
	},
}

// getLocalizedMessage returns the localized error message based on Accept-Language header.
func getLocalizedMessage(code string, r *http.Request) string {
	// 从 Accept-Language header 获取语言偏好
	acceptLang := r.Header.Get("Accept-Language")
	if acceptLang == "" {
		acceptLang = "en" // 默认英文
	}

	// 解析语言列表，按优先级排序
	languages := parseAcceptLanguage(acceptLang)

	// 按优先级查找支持的语言
	for _, lang := range languages {
		// 中文（简体/繁体）
		if strings.HasPrefix(lang, "zh") {
			if msg, ok := errorCodeMessages["zh"][code]; ok {
				return msg
			}
		}

		// 英文
		if strings.HasPrefix(lang, "en") {
			if msg, ok := errorCodeMessages["en"][code]; ok {
				return msg
			}
		}
	}

	// 默认返回英文
	if msg, ok := errorCodeMessages["en"][code]; ok {
		return msg
	}

	// 如果代码不存在，返回通用错误
	return "An error occurred"
}

// parseAcceptLanguage parses Accept-Language header and returns sorted language list.
func parseAcceptLanguage(acceptLang string) []string {
	// 简化实现：按逗号分割，去除空格
	// 完整实现应该处理 q 值（优先级）
	languages := strings.Split(acceptLang, ",")
	result := make([]string, 0, len(languages))

	for _, lang := range languages {
		lang = strings.TrimSpace(lang)
		// 移除 q 值（如 "zh-CN;q=0.9" -> "zh-CN"）
		if idx := strings.Index(lang, ";"); idx != -1 {
			lang = lang[:idx]
		}
		if lang != "" {
			result = append(result, lang)
		}
	}

	return result
}

// getLocalizedErrorResponse returns a complete localized error response.
func getLocalizedErrorResponse(code string, r *http.Request) map[string]interface{} {
	message := getLocalizedMessage(code, r)

	return map[string]interface{}{
		"code":    code,
		"message": message,
	}
}
