package v1

import (
	"fmt"
)

// MCPTool represents an MCP tool with its schema.
type MCPTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// ToolRegistry holds all available MCP tools.
var ToolRegistry = []MCPTool{
	{
		Name:        "create_memo",
		Description: "创建新笔记，支持 Markdown 格式、标签（#work #AI）和提及（@username）",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"content": map[string]interface{}{
					"type":        "string",
					"description": "笔记内容，支持 Markdown 格式、标签（#work #AI）和提及（@username）",
					"minLength":   1,
					"maxLength":   10000,
				},
				"visibility": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"PRIVATE", "PROTECTED", "PUBLIC"},
					"description": "可见性：PRIVATE（仅自己）、PROTECTED（登录用户）、PUBLIC（所有人）",
				},
			},
			"required": []string{"content"},
		},
	},
	{
		Name:        "get_memo",
		Description: "获取单个笔记的详细内容",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"id": map[string]interface{}{
					"type":        "string",
					"description": "笔记 ID 或 UID，例如：123 或 abc123",
				},
			},
			"required": []string{"id"},
		},
	},
	{
		Name:        "list_memos",
		Description: "列出笔记，支持过滤条件（tag:#work, creator:me, created:today）",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"filter": map[string]interface{}{
					"type":        "string",
					"description": "过滤条件，例如：tag:#work, creator:me, created:today",
				},
				"limit": map[string]interface{}{
					"type":        "integer",
					"description": "返回数量限制，默认 20，最大 100",
					"default":     20,
					"minimum":     1,
					"maximum":     100,
				},
			},
		},
	},
	{
		Name:        "update_memo",
		Description: "更新已有笔记的内容",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"id": map[string]interface{}{
					"type":        "string",
					"description": "笔记 ID 或 UID",
				},
				"content": map[string]interface{}{
					"type":        "string",
					"description": "新的笔记内容",
					"minLength":   1,
					"maxLength":   10000,
				},
			},
			"required": []string{"id", "content"},
		},
	},
	{
		Name:        "delete_memo",
		Description: "删除笔记（不可撤销）",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"id": map[string]interface{}{
					"type":        "string",
					"description": "要删除的笔记 ID 或 UID",
				},
			},
			"required": []string{"id"},
		},
	},
}

// GetToolsList returns the JSON-RPC tools/list response.
func GetToolsList() map[string]interface{} {
	tools := make([]map[string]interface{}, len(ToolRegistry))
	for i, tool := range ToolRegistry {
		tools[i] = map[string]interface{}{
			"name":        tool.Name,
			"description": tool.Description,
			"inputSchema": tool.InputSchema,
		}
	}
	return map[string]interface{}{
		"tools": tools,
	}
}

// ValidateToolCall validates if a tool name exists and arguments match schema.
func ValidateToolCall(name string, args map[string]interface{}) error {
	var tool *MCPTool
	for i := range ToolRegistry {
		if ToolRegistry[i].Name == name {
			tool = &ToolRegistry[i]
			break
		}
	}

	if tool == nil {
		return fmt.Errorf("unknown tool: %s", name)
	}

	// 验证必填参数
	required, ok := tool.InputSchema["required"].([]string)
	if !ok {
		return nil
	}

	for _, field := range required {
		if _, exists := args[field]; !exists {
			return fmt.Errorf("missing required field: %s", field)
		}
	}

	return nil
}
