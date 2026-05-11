package v1

import (
	"testing"
	"time"

	v1pb "github.com/usememos/memos/proto/gen/api/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestExtractTags(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name:     "英文标签",
			content:  "今天学习了 #AI 和 #MachineLearning",
			expected: []string{"#AI", "#MachineLearning"},
		},
		{
			name:     "中文标签",
			content:  "关于 #人工智能 的思考",
			expected: []string{"#人工智能"},
		},
		{
			name:     "混合标签",
			content:  "学习了 #AI 和 #深度学习",
			expected: []string{"#AI", "#深度学习"},
		},
		{
			name:     "无标签",
			content:  "普通笔记内容",
			expected: []string{},
		},
		{
			name:     "重复标签去重",
			content:  "#AI #AI #机器学习 #AI",
			expected: []string{"#AI", "#机器学习"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractTags(tt.content)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTruncateContent(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		maxLen   int
		expected string
	}{
		{
			name:     "不需要截断",
			content:  "短文本",
			maxLen:   20,
			expected: "短文本",
		},
		{
			name:     "需要截断",
			content:  "这是一段很长的文本需要被截断",
			maxLen:   10,
			expected: "这是一段很长的文本需...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TruncateContent(tt.content, tt.maxLen)
			assert.Equal(t, tt.expected, result)
			assert.LessOrEqual(t, len([]rune(result)), tt.maxLen+3) // +3 for "..."
		})
	}
}

func TestFormatTimestamp(t *testing.T) {
	t.Run("nil 时间戳", func(t *testing.T) {
		result := formatTimestamp(nil)
		assert.Equal(t, "", result)
	})

	t.Run("有效时间戳", func(t *testing.T) {
		ts := timestamppb.New(time.Unix(1715328000, 0)) // 2024-05-10 00:00:00 UTC
		result := formatTimestamp(ts)
		assert.Contains(t, result, "2024")
	})
}

func TestFormatMemo(t *testing.T) {
	t.Run("格式化基本 Memo", func(t *testing.T) {
		createTime := timestamppb.New(time.Unix(1715328000, 0)) // 2024-05-10 00:00:00 UTC
		updateTime := timestamppb.New(time.Unix(1715414400, 0)) // 2024-05-11 00:00:00 UTC

		memo := &v1pb.Memo{
			Name:       "memos/123",
			Content:    "测试内容",
			Visibility: v1pb.Visibility_PUBLIC,
			CreateTime: createTime,
			UpdateTime: updateTime,
			Tags:       []string{"#test", "#测试"},
		}

		result := FormatMemo(memo, "https://example.com")

		assert.Equal(t, "123", result["id"])
		assert.Equal(t, "123", result["uid"])
		assert.Equal(t, "memos/123", result["name"])
		assert.Equal(t, "测试内容", result["content"])
		assert.Equal(t, "PUBLIC", result["visibility"])
		assert.Contains(t, result["createdAt"], "2024")
		assert.Contains(t, result["updatedAt"], "2024")
		assert.Equal(t, "https://example.com/m/123", result["url"])
		assert.Equal(t, []string{"#test", "#测试"}, result["tags"])
	})

	t.Run("nil 时间戳处理", func(t *testing.T) {
		memo := &v1pb.Memo{
			Name:       "memos/1",
			Content:    "内容",
			Visibility: v1pb.Visibility_PRIVATE,
			CreateTime: nil,
			UpdateTime: nil,
			Tags:       []string{},
		}

		result := FormatMemo(memo, "https://example.com")

		assert.Equal(t, "", result["createdAt"])
		assert.Equal(t, "", result["updatedAt"])
	})
}

func TestSanitizeContent(t *testing.T) {
	t.Run("基础实现", func(t *testing.T) {
		content := "这是测试内容"
		result := SanitizeContent(content)
		assert.Equal(t, content, result)
	})
}
