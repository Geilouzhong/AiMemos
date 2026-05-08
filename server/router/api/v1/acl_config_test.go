package v1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestPublicMethodsArePublic verifies that methods in PublicMethods are recognized as public.
func TestPublicMethodsArePublic(t *testing.T) {
	publicMethods := []string{
		// Auth Service
		"/memos.api.v1.AuthService/SignIn",
		"/memos.api.v1.AuthService/RefreshToken",
		// Instance Service
		"/memos.api.v1.InstanceService/GetInstanceProfile",
		"/memos.api.v1.InstanceService/GetInstanceSetting",
		// User Service
		"/memos.api.v1.UserService/CreateUser",
		"/memos.api.v1.UserService/GetUser",
		"/memos.api.v1.UserService/GetUserAvatar",
		"/memos.api.v1.UserService/GetUserStats",
		"/memos.api.v1.UserService/ListAllUserStats",
		"/memos.api.v1.UserService/SearchUsers",
		// Identity Provider Service
		"/memos.api.v1.IdentityProviderService/ListIdentityProviders",
		// Memo Service
		"/memos.api.v1.MemoService/GetMemo",
		"/memos.api.v1.MemoService/ListMemos",
	}

	for _, method := range publicMethods {
		t.Run(method, func(t *testing.T) {
			assert.True(t, IsPublicMethod(method), "Expected %s to be public", method)
		})
	}
}

// TestProtectedMethodsRequireAuth verifies that non-public methods are recognized as protected.
func TestProtectedMethodsRequireAuth(t *testing.T) {
	protectedMethods := []string{
		// Auth Service - logout and get current user require auth
		"/memos.api.v1.AuthService/SignOut",
		"/memos.api.v1.AuthService/GetCurrentUser",
		// Instance Service - admin operations
		"/memos.api.v1.InstanceService/UpdateInstanceSetting",
		// User Service - modification operations
		"/memos.api.v1.UserService/ListUsers",
		"/memos.api.v1.UserService/UpdateUser",
		"/memos.api.v1.UserService/DeleteUser",
		// Memo Service - write operations
		"/memos.api.v1.MemoService/CreateMemo",
		"/memos.api.v1.MemoService/UpdateMemo",
		"/memos.api.v1.MemoService/DeleteMemo",
		// Attachment Service - write operations
		"/memos.api.v1.AttachmentService/CreateAttachment",
		"/memos.api.v1.AttachmentService/DeleteAttachment",
		// Shortcut Service
		"/memos.api.v1.ShortcutService/CreateShortcut",
		"/memos.api.v1.ShortcutService/ListShortcuts",
		"/memos.api.v1.ShortcutService/UpdateShortcut",
		"/memos.api.v1.ShortcutService/DeleteShortcut",
		// Activity Service
		"/memos.api.v1.ActivityService/GetActivity",
	}

	for _, method := range protectedMethods {
		t.Run(method, func(t *testing.T) {
			assert.False(t, IsPublicMethod(method), "Expected %s to require auth", method)
		})
	}
}

// TestUnknownMethodsRequireAuth verifies that unknown methods default to requiring auth.
func TestUnknownMethodsRequireAuth(t *testing.T) {
	unknownMethods := []string{
		"/unknown.Service/Method",
		"/memos.api.v1.UnknownService/Method",
		"",
		"invalid",
	}

	for _, method := range unknownMethods {
		t.Run(method, func(t *testing.T) {
			assert.False(t, IsPublicMethod(method), "Unknown method %q should require auth", method)
		})
	}
}

func TestGuestBlockedWriteMethods(t *testing.T) {
	blockedMethods := []string{
		"/memos.api.v1.MemoService/CreateMemo",
		"/memos.api.v1.MemoService/UpdateMemo",
		"/memos.api.v1.MemoService/DeleteMemo",
		"/memos.api.v1.MemoService/CreateMemoComment",
		"/memos.api.v1.AttachmentService/CreateAttachment",
		"/memos.api.v1.ShortcutService/DeleteShortcut",
		"/memos.api.v1.UserService/UpdateUser",
		"/memos.api.v1.IdentityProviderService/DeleteIdentityProvider",
	}

	for _, method := range blockedMethods {
		t.Run(method, func(t *testing.T) {
			assert.True(t, IsGuestWriteBlockedMethod(method), "Expected %s to be blocked for guest", method)
		})
	}

	notBlockedMethods := []string{
		"/memos.api.v1.AuthService/SignIn",
		"/memos.api.v1.MemoService/ListMemos",
		"/memos.api.v1.MemoService/GetMemo",
		"/memos.api.v1.UserService/CreateUser",
	}

	for _, method := range notBlockedMethods {
		t.Run(method, func(t *testing.T) {
			assert.False(t, IsGuestWriteBlockedMethod(method), "Expected %s to be allowed for guest", method)
		})
	}
}

func TestIsGuestWriteBlockedGatewayRequest(t *testing.T) {
	tests := []struct {
		name     string
		method   string
		path     string
		expected bool
	}{
		{name: "block create memo", method: "POST", path: "/api/v1/memos", expected: true},
		{name: "block update memo", method: "PATCH", path: "/api/v1/memos/abc123", expected: true},
		{name: "block delete memo", method: "DELETE", path: "/api/v1/memos/abc123", expected: true},
		{name: "block create comment", method: "POST", path: "/api/v1/memos/abc123/comments", expected: true},
		{name: "block create attachment", method: "POST", path: "/api/v1/attachments", expected: true},
		{name: "normalize lowercase method", method: "post", path: "/api/v1/memos", expected: true},
		{name: "allow read memos", method: "GET", path: "/api/v1/memos", expected: false},
		{name: "allow sign in", method: "POST", path: "/api/v1/auth/signin", expected: false},
		{name: "allow user registration", method: "POST", path: "/api/v1/users", expected: false},
		{name: "allow unmatched mutation path", method: "PATCH", path: "/api/v1/memos", expected: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := IsGuestWriteBlockedGatewayRequest(tc.method, tc.path)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
