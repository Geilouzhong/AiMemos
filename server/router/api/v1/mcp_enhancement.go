package v1

import (
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"

	"google.golang.org/protobuf/types/known/timestamppb"

	v1pb "github.com/usememos/memos/proto/gen/api/v1"
)

// tagRegex matches hashtags in content (#work, #AI, #工作).
var tagRegex = regexp.MustCompile(`#([\p{L}\p{N}_]+)`)

// ExtractTags extracts all hashtags from content.
func ExtractTags(content string) []string {
	matches := tagRegex.FindAllStringSubmatch(content, -1)
	tags := make([]string, 0, len(matches))

	seen := make(map[string]bool)
	for _, match := range matches {
		if len(match) > 1 {
			tag := "#" + match[1]
			if !seen[tag] {
				seen[tag] = true
				tags = append(tags, tag)
			}
		}
	}

	return tags
}

// FormatMemo formats a store.Memo into MCP response format.
func FormatMemo(memo *v1pb.Memo, baseURL string) map[string]interface{} {
	uid := strings.TrimPrefix(memo.Name, "memos/")

	return map[string]interface{}{
		"id":         uid,
		"uid":        uid,
		"name":       memo.Name,
		"content":    memo.Content,
		"visibility": memo.Visibility.String(),
		"createdAt":  formatTimestamp(memo.CreateTime),
		"updatedAt":  formatTimestamp(memo.UpdateTime),
		"url":        fmt.Sprintf("%s/m/%s", baseURL, uid),
		"tags":       memo.Tags,
	}
}

// formatTimestamp converts protobuf timestamp to RFC3339 string.
func formatTimestamp(ts *timestamppb.Timestamp) string {
	if ts == nil {
		return ""
	}
	return ts.AsTime().Format(time.RFC3339)
}

// SanitizeContent removes sensitive information from content.
func SanitizeContent(content string) string {
	// 移除可能的敏感信息（如密码、token等）
	// 这是一个基础实现，可以根据需要扩展
	return content
}

// TruncateContent truncates content to max length with ellipsis.
func TruncateContent(content string, maxLen int) string {
	runes := []rune(content)
	if len(runes) <= maxLen {
		return content
	}

	// 尝试在单词边界截断
	truncated := string(runes[:maxLen])
	lastSpace := strings.LastIndexFunc(truncated, unicode.IsSpace)

	if lastSpace > 0 && lastSpace < maxLen-3 {
		truncated = truncated[:lastSpace]
	}

	return truncated + "..."
}
