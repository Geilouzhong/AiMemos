package v1

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	v1pb "github.com/Geilouzhong/AiMemos/proto/gen/api/v1"
)

// handleCreateMemo implements the create_memo tool.
func (h *MCPHandler) handleCreateMemo(ctx context.Context, _ int32, args map[string]interface{}) (map[string]interface{}, error) {
	content, ok := args["content"].(string)
	if !ok || strings.TrimSpace(content) == "" {
		return nil, errors.New("content is required")
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

// handlePullMemo implements the pull_memo tool.
func (h *MCPHandler) handlePullMemo(ctx context.Context, _ int32, args map[string]interface{}) (map[string]interface{}, error) {
	id, ok := args["id"].(string)
	if !ok {
		return nil, errors.New("id is required")
	}
	path, ok := args["path"].(string)
	if !ok || strings.TrimSpace(path) == "" {
		return nil, errors.New("path is required")
	}

	req := &v1pb.GetMemoRequest{
		Name: fmt.Sprintf("memos/%s", id),
	}

	memo, err := h.GetMemo(ctx, req)
	if err != nil {
		return nil, errors.Errorf("memo not found: %s", id)
	}
	resolvedPath, err := writeMemoToLocalFile(path, memo)
	if err != nil {
		return nil, errors.Wrap(err, "failed to write memo to local file")
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
		"file": map[string]interface{}{
			"path":    resolvedPath,
			"written": true,
		},
	}, nil
}

// handleListMemos implements the list_memos tool.
func (h *MCPHandler) handleListMemos(ctx context.Context, _ int32, args map[string]interface{}) (map[string]interface{}, error) {
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


// handlePushMemo implements the push_memo tool.
func (h *MCPHandler) handlePushMemo(ctx context.Context, _ int32, args map[string]interface{}) (map[string]interface{}, error) {
	path, ok := args["path"].(string)
	if !ok || strings.TrimSpace(path) == "" {
		return nil, errors.New("path is required")
	}
	checkConflict := true
	if raw, ok := args["check_conflict"].(bool); ok {
		checkConflict = raw
	}

	content, meta, err := readMemoFromLocalFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read local memo files")
	}

	baseURL := h.profile.InstanceURL
	if baseURL == "" {
		baseURL = "http://localhost:5230"
	}

	if meta == nil || strings.TrimSpace(meta.MemoName) == "" {
		createReq := &v1pb.CreateMemoRequest{
			Memo: &v1pb.Memo{
				Title:   strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)),
				Content: content,
				Visibility: v1pb.Visibility_PRIVATE,
			},
		}
		memo, err := h.CreateMemo(ctx, createReq)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create memo from local file")
		}
		if err := updateMemoMetadataInLocalFile(path, memo); err != nil {
			return nil, errors.Wrap(err, "failed to write created memo metadata back to file")
		}
		uid := strings.TrimPrefix(memo.Name, "memos/")
		return map[string]interface{}{
			"memo": map[string]interface{}{
				"id":         uid,
				"uid":        uid,
				"name":       memo.Name,
				"title":      memo.Title,
				"content":    memo.Content,
				"visibility": memo.Visibility.String(),
				"createdAt":  memo.CreateTime,
				"updatedAt":  memo.UpdateTime,
				"url":        fmt.Sprintf("%s/m/%s", baseURL, uid),
			},
			"created": true,
			"file": map[string]interface{}{
				"path": path,
			},
		}, nil
	}

	remoteMemo, err := h.GetMemo(ctx, &v1pb.GetMemoRequest{Name: meta.MemoName})
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch remote memo before push")
	}
	if conflict, message := checkMemoPushConflict(meta, remoteMemo); conflict && checkConflict {
		return map[string]interface{}{
			"conflict": true,
			"message":  message,
			"memo": map[string]interface{}{
				"name":      remoteMemo.Name,
				"updatedAt": remoteMemo.UpdateTime,
			},
		}, nil
	}

	req := &v1pb.UpdateMemoRequest{
		Memo: &v1pb.Memo{
			Name:    meta.MemoName,
			Title:   meta.Title,
			Content: content,
		},
		UpdateMask: &fieldmaskpb.FieldMask{
			Paths: []string{"title", "content"},
		},
	}

	memo, err := h.UpdateMemo(ctx, req)
	if err != nil {
		return nil, err
	}

	// Extract UID from name (format: memos/{uid})
	uid := strings.TrimPrefix(memo.Name, "memos/")
	if err := updateMemoMetadataInLocalFile(path, memo); err != nil {
		return nil, errors.Wrap(err, "failed to refresh memo metadata after push")
	}

	return map[string]interface{}{
		"memo": map[string]interface{}{
			"id":         uid,
			"uid":        uid,
			"name":       memo.Name,
			"title":      memo.Title,
			"content":    memo.Content,
			"visibility": memo.Visibility.String(),
			"createdAt":  memo.CreateTime,
			"updatedAt":  memo.UpdateTime,
			"tags":       memo.Tags,
			"url":        fmt.Sprintf("%s/m/%s", baseURL, uid),
		},
		"conflict": false,
		"updated":  true,
	}, nil
}

// handleDeleteMemo implements the delete_memo tool.
func (h *MCPHandler) handleDeleteMemo(ctx context.Context, _ int32, args map[string]interface{}) (map[string]interface{}, error) {
	id, ok := args["id"].(string)
	if !ok {
		return nil, errors.New("id is required")
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
