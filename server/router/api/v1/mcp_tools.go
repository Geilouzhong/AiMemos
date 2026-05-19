package v1

import (
	"github.com/pkg/errors"
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
		Name:        "pull_memo",
		Description: "将单个笔记拉取为本地 Markdown 文件，供本地编辑后再推送",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"id": map[string]interface{}{
					"type":        "string",
					"description": "笔记 ID 或 UID，例如：123 或 abc123",
				},
				"path": map[string]interface{}{
					"type":        "string",
					"description": "本地 Markdown 文件路径，例如：/workspace/memos/weekly-review.md",
				},
			},
			"required": []string{"id", "path"},
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
		Name:        "push_memo",
		Description: "将本地 Markdown 文件内容推送回远端笔记，并进行冲突检查",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path": map[string]interface{}{
					"type":        "string",
					"description": "本地 Markdown 文件路径",
				},
				"check_conflict": map[string]interface{}{
					"type":        "boolean",
					"description": "是否在推送前检查远端更新时间冲突，默认 true",
					"default":     true,
				},
			},
			"required": []string{"path"},
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
		return errors.Errorf("unknown tool: %s", name)
	}

	// 验证必填参数
	required, ok := tool.InputSchema["required"].([]string)
	if !ok {
		return nil
	}

	for _, field := range required {
		if _, exists := args[field]; !exists {
			return errors.Errorf("missing required field: %s", field)
		}
	}

	return nil
}
