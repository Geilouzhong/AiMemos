package v1

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Geilouzhong/AiMemos/internal/profile"
	"github.com/Geilouzhong/AiMemos/server/auth"
)

func TestToolRegistry(t *testing.T) {
	t.Run("工具数量正确", func(t *testing.T) {
		assert.Len(t, ToolRegistry, 5)
	})

	t.Run("工具名称唯一", func(t *testing.T) {
		names := make(map[string]bool)
		for _, tool := range ToolRegistry {
			assert.False(t, names[tool.Name], "duplicate tool name: %s", tool.Name)
			names[tool.Name] = true
		}
	})

	t.Run("验证 create_memo 必填参数", func(t *testing.T) {
		err := ValidateToolCall("create_memo", map[string]interface{}{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing required field")

		err = ValidateToolCall("create_memo", map[string]interface{}{
			"content": "test",
		})
		assert.NoError(t, err)
	})

	t.Run("验证 get_memo 必填参数", func(t *testing.T) {
		err := ValidateToolCall("get_memo", map[string]interface{}{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing required field")

		err = ValidateToolCall("get_memo", map[string]interface{}{
			"id": "123",
		})
		assert.NoError(t, err)
	})

	t.Run("验证 update_memo 必填参数", func(t *testing.T) {
		err := ValidateToolCall("update_memo", map[string]interface{}{
			"id": "123",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing required field")

		err = ValidateToolCall("update_memo", map[string]interface{}{
			"id":      "123",
			"content": "updated content",
		})
		assert.NoError(t, err)
	})

	t.Run("验证 delete_memo 必填参数", func(t *testing.T) {
		err := ValidateToolCall("delete_memo", map[string]interface{}{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing required field")

		err = ValidateToolCall("delete_memo", map[string]interface{}{
			"id": "123",
		})
		assert.NoError(t, err)
	})

	t.Run("验证 list_memos 可选参数", func(t *testing.T) {
		err := ValidateToolCall("list_memos", map[string]interface{}{})
		assert.NoError(t, err) // list_memos 没有必填参数

		err = ValidateToolCall("list_memos", map[string]interface{}{
			"filter": "tag:#work",
			"limit":  10,
		})
		assert.NoError(t, err)
	})

	t.Run("验证未知工具", func(t *testing.T) {
		err := ValidateToolCall("unknown_tool", map[string]interface{}{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown tool")
	})
}

func TestGetToolsList(t *testing.T) {
	t.Run("返回正确的工具列表结构", func(t *testing.T) {
		result := GetToolsList()

		assert.Contains(t, result, "tools")
		tools, ok := result["tools"].([]map[string]interface{})
		require.True(t, ok)
		assert.Len(t, tools, 5)

		// 验证每个工具都有必需的字段
		for _, tool := range tools {
			assert.Contains(t, tool, "name")
			assert.Contains(t, tool, "description")
			assert.Contains(t, tool, "inputSchema")
		}
	})
}

func TestMCPHTTPToolsListResponseShape(t *testing.T) {
	secret := "test-secret"
	token, _, err := auth.GenerateAccessTokenV2(1, "testuser", "USER", "NORMAL", []byte(secret))
	require.NoError(t, err)

	service := &APIV1Service{
		Secret: secret,
		Profile: &profile.Profile{
			MCPPath: "/mcp",
		},
	}
	handler := NewMCPHandler(service, service.Profile)

	req := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"tools/list"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	require.NoError(t, handler.HandleHTTPToolCall(c))
	require.Equal(t, http.StatusOK, rec.Code)

	var body map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, "2.0", body["jsonrpc"])
	require.Equal(t, float64(1), body["id"])

	result, ok := body["result"].(map[string]interface{})
	require.True(t, ok)
	tools, ok := result["tools"].([]interface{})
	require.True(t, ok, "result.tools must be a JSON array, not a nested object")
	require.Len(t, tools, len(ToolRegistry))
	_, nested := result["tools"].(map[string]interface{})
	require.False(t, nested, "tools/list must not return result.tools.tools")
}

func TestMCPHTTPInitializedNotification(t *testing.T) {
	secret := "test-secret"
	token, _, err := auth.GenerateAccessTokenV2(1, "testuser", "USER", "NORMAL", []byte(secret))
	require.NoError(t, err)

	service := &APIV1Service{
		Secret: secret,
		Profile: &profile.Profile{
			MCPPath: "/mcp",
		},
	}
	handler := NewMCPHandler(service, service.Profile)

	req := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(`{"jsonrpc":"2.0","method":"notifications/initialized"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	require.NoError(t, handler.HandleHTTPToolCall(c))
	require.Equal(t, http.StatusAccepted, rec.Code)
	require.Empty(t, strings.TrimSpace(rec.Body.String()))
}

func TestMCPToolSchema(t *testing.T) {
	t.Run("create_memo schema 结构正确", func(t *testing.T) {
		var tool *MCPTool
		for i := range ToolRegistry {
			if ToolRegistry[i].Name == "create_memo" {
				tool = &ToolRegistry[i]
				break
			}
		}

		assert.NotNil(t, tool)
		assert.Equal(t, "object", tool.InputSchema["type"])

		props, ok := tool.InputSchema["properties"].(map[string]interface{})
		require.True(t, ok)
		assert.Contains(t, props, "content")
		assert.Contains(t, props, "visibility")

		content, ok := props["content"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "string", content["type"])
		assert.Equal(t, 1, content["minLength"])
		assert.Equal(t, 10000, content["maxLength"])

		visibility, ok := props["visibility"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "string", visibility["type"])
		assert.Contains(t, visibility["enum"], "PRIVATE")
		assert.Contains(t, visibility["enum"], "PROTECTED")
		assert.Contains(t, visibility["enum"], "PUBLIC")
	})
}
