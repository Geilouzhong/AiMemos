package v1

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lithammer/shortuuid/v4"
	"github.com/stretchr/testify/require"

	"github.com/Geilouzhong/AiMemos/internal/profile"
	"github.com/Geilouzhong/AiMemos/plugin/markdown"
	"github.com/Geilouzhong/AiMemos/server/auth"
	"github.com/Geilouzhong/AiMemos/store"
	teststore "github.com/Geilouzhong/AiMemos/store/test"
)

func newTestMCPHandler(t *testing.T) (*MCPHandler, context.Context, int32) {
	t.Helper()
	ctx := context.Background()
	ts := teststore.NewTestingStore(ctx, t)
	t.Cleanup(func() { _ = ts.Close() })

	adminUser, err := ts.CreateUser(ctx, &store.User{
		Username: "mcpadmin",
		Role:     store.RoleAdmin,
		Email:    "mcpadmin@example.com",
	})
	require.NoError(t, err)

	service := &APIV1Service{
		Secret: "test-secret",
		Profile: &profile.Profile{
			Demo:        true,
			Version:     "test-1.0.0",
			InstanceURL: "http://localhost:8081",
			Driver:      "sqlite",
			DSN:         ":memory:",
		},
		Store:           ts,
		MarkdownService: markdown.NewService(markdown.WithTagExtension()),
	}

	handler := NewMCPHandler(service, service.Profile)
	userCtx := context.WithValue(ctx, auth.UserIDContextKey, adminUser.ID)
	return handler, userCtx, adminUser.ID
}

func TestPushMemoCreatesNewMemoAndWritesMetadata(t *testing.T) {
	handler, ctx, userID := newTestMCPHandler(t)

	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "新建MCP测试memo.md")
	err := os.WriteFile(path, []byte("# 新建MCP测试memo\n\n这是一篇通过 push_memo 自动创建的测试 memo。\n"), 0644)
	require.NoError(t, err)

	result, err := handler.handlePushMemo(ctx, userID, map[string]interface{}{
		"path":           path,
		"check_conflict": true,
	})
	require.NoError(t, err)
	require.Equal(t, true, result["created"])

	content, err := os.ReadFile(path)
	require.NoError(t, err)
	contentStr := string(content)
	require.Contains(t, contentStr, "<!-- aimemos-meta: ")
	require.Contains(t, contentStr, "\"memo_name\":\"memos/")

	result, err = handler.handlePushMemo(ctx, userID, map[string]interface{}{
		"path":           path,
		"check_conflict": true,
	})
	require.NoError(t, err)
	require.Equal(t, true, result["updated"])
	require.Equal(t, false, result["conflict"])
	_, hasCreated := result["created"]
	require.False(t, hasCreated)
	}

func TestPullMemoUsesTitleForFileName(t *testing.T) {
	handler, ctx, _ := newTestMCPHandler(t)

	memoID := shortuuid.New()
	createdMemo, err := handler.Store.CreateMemo(ctx, &store.Memo{
		UID:        memoID,
		CreatorID:  1,
		Title:      "示例标题",
		Content:    "# 示例标题\n\n正文",
		Visibility: store.Private,
	})
	require.NoError(t, err)

	tempDir := t.TempDir()
	result, err := handler.handlePullMemo(ctx, 1, map[string]interface{}{
		"id":   createdMemo.UID,
		"path": tempDir + string(os.PathSeparator),
	})
	require.NoError(t, err)

	fileResult := result["file"].(map[string]interface{})
	path := fileResult["path"].(string)
	require.True(t, strings.HasSuffix(path, "示例标题.md"))
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Contains(t, string(content), "<!-- aimemos-meta: ")
}
