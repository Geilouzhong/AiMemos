package v1

import (
	"context"
	"fmt"
	"strings"

	v1pb "github.com/usememos/memos/proto/gen/api/v1"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// handleCreateMemo implements the create_memo tool.
func (h *MCPHandler) handleCreateMemo(ctx context.Context, userID int32, args map[string]interface{}) (map[string]interface{}, error) {
	content, ok := args["content"].(string)
	if !ok || strings.TrimSpace(content) == "" {
		return nil, fmt.Errorf("content is required")
	}

	visibility := v1pb.Visibility_PRIVATE
	if v, ok := args["visibility"].(string); ok {
		switch v {
		case "PUBLIC":
			visibility = v1pb.Visibility_PUBLIC
		case "PROTECTED":
			visibility = v1pb.Visibility_PROTECTED
		default:
			visibility = v1pb.Visibility_PRIVATE
		}
	}

	// TODO: 自动标签提取 - will be implemented in Task 6
	// tags := ExtractTags(content)

	req := &v1pb.CreateMemoRequest{
		Memo: &v1pb.Memo{
			Content:    content,
			Visibility: visibility,
		},
	}

	memo, err := h.CreateMemo(ctx, req)
	if err != nil {
		return nil, err
	}

	baseURL := h.profile.InstanceURL
	if baseURL == "" {
		baseURL = "http://localhost:5230"
	}

	// Extract UID from name (format: memos/{uid})
	uid := strings.TrimPrefix(memo.Name, "memos/")

	return map[string]interface{}{
		"memo": map[string]interface{}{
			"id":         uid,
			"uid":        uid,
			"name":       memo.Name,
			"content":    memo.Content,
			"visibility": memo.Visibility.String(),
			"createdAt":  memo.CreateTime,
			"updatedAt":  memo.UpdateTime,
			"url":        fmt.Sprintf("%s/m/%s", baseURL, uid),
		},
		"extractedTags": memo.Tags,
	}, nil
}

// handleGetMemo implements the get_memo tool.
func (h *MCPHandler) handleGetMemo(ctx context.Context, userID int32, args map[string]interface{}) (map[string]interface{}, error) {
	id, ok := args["id"].(string)
	if !ok {
		return nil, fmt.Errorf("id is required")
	}

	req := &v1pb.GetMemoRequest{
		Name: fmt.Sprintf("memos/%s", id),
	}

	memo, err := h.GetMemo(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("memo not found: %s", id)
	}

	baseURL := h.profile.InstanceURL
	if baseURL == "" {
		baseURL = "http://localhost:5230"
	}

	// Extract UID from name (format: memos/{uid})
	uid := strings.TrimPrefix(memo.Name, "memos/")

	return map[string]interface{}{
		"memo": map[string]interface{}{
			"id":         uid,
			"uid":        uid,
			"name":       memo.Name,
			"content":    memo.Content,
			"visibility": memo.Visibility.String(),
			"createdAt":  memo.CreateTime,
			"updatedAt":  memo.UpdateTime,
			"tags":       memo.Tags,
			"url":        fmt.Sprintf("%s/m/%s", baseURL, uid),
		},
	}, nil
}

// handleListMemos implements the list_memos tool.
func (h *MCPHandler) handleListMemos(ctx context.Context, userID int32, args map[string]interface{}) (map[string]interface{}, error) {
	filter := ""
	if f, ok := args["filter"].(string); ok {
		filter = f
	}

	pageSize := int32(20)
	if l, ok := args["limit"].(float64); ok {
		pageSize = int32(l)
		if pageSize > 100 {
			pageSize = 100
		}
	}

	req := &v1pb.ListMemosRequest{
		Filter:   filter,
		PageSize: pageSize,
	}

	resp, err := h.ListMemos(ctx, req)
	if err != nil {
		return nil, err
	}

	baseURL := h.profile.InstanceURL
	if baseURL == "" {
		baseURL = "http://localhost:5230"
	}

	memos := make([]map[string]interface{}, len(resp.Memos))
	for i, memo := range resp.Memos {
		// Extract UID from name (format: memos/{uid})
		uid := strings.TrimPrefix(memo.Name, "memos/")
		memos[i] = map[string]interface{}{
			"id":         uid,
			"uid":        uid,
			"name":       memo.Name,
			"content":    memo.Content,
			"visibility": memo.Visibility.String(),
			"createdAt":  memo.CreateTime,
			"tags":       memo.Tags,
			"url":        fmt.Sprintf("%s/m/%s", baseURL, uid),
		}
	}

	return map[string]interface{}{
		"memos": memos,
		"count": len(memos),
	}, nil
}

// handleUpdateMemo implements the update_memo tool.
func (h *MCPHandler) handleUpdateMemo(ctx context.Context, userID int32, args map[string]interface{}) (map[string]interface{}, error) {
	id, ok := args["id"].(string)
	if !ok {
		return nil, fmt.Errorf("id is required")
	}

	content, ok := args["content"].(string)
	if !ok || strings.TrimSpace(content) == "" {
		return nil, fmt.Errorf("content is required")
	}

	req := &v1pb.UpdateMemoRequest{
		Memo: &v1pb.Memo{
			Name:    fmt.Sprintf("memos/%s", id),
			Content: content,
		},
		UpdateMask: &fieldmaskpb.FieldMask{
			Paths: []string{"content"},
		},
	}

	memo, err := h.UpdateMemo(ctx, req)
	if err != nil {
		return nil, err
	}

	baseURL := h.profile.InstanceURL
	if baseURL == "" {
		baseURL = "http://localhost:5230"
	}

	// Extract UID from name (format: memos/{uid})
	uid := strings.TrimPrefix(memo.Name, "memos/")

	return map[string]interface{}{
		"memo": map[string]interface{}{
			"id":         uid,
			"uid":        uid,
			"name":       memo.Name,
			"content":    memo.Content,
			"visibility": memo.Visibility.String(),
			"createdAt":  memo.CreateTime,
			"updatedAt":  memo.UpdateTime,
			"tags":       memo.Tags,
			"url":        fmt.Sprintf("%s/m/%s", baseURL, uid),
		},
	}, nil
}

// handleDeleteMemo implements the delete_memo tool.
func (h *MCPHandler) handleDeleteMemo(ctx context.Context, userID int32, args map[string]interface{}) (map[string]interface{}, error) {
	id, ok := args["id"].(string)
	if !ok {
		return nil, fmt.Errorf("id is required")
	}

	req := &v1pb.DeleteMemoRequest{
		Name: fmt.Sprintf("memos/%s", id),
	}

	_, err := h.DeleteMemo(ctx, req)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"success": true,
		"message": "Memo deleted successfully",
	}, nil
}
