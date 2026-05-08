package v1

import (
	"regexp"
	"strings"
)

// PublicMethods defines API endpoints that don't require authentication.
// All other endpoints require a valid session or access token.
//
// This is the SINGLE SOURCE OF TRUTH for public endpoints.
// Both Connect interceptor and gRPC-Gateway interceptor use this map.
//
// Format: Full gRPC procedure path as returned by req.Spec().Procedure (Connect)
// or info.FullMethod (gRPC interceptor).
var PublicMethods = map[string]struct{}{
	// Auth Service - login/token endpoints must be accessible without auth
	"/memos.api.v1.AuthService/SignIn":       {},
	"/memos.api.v1.AuthService/RefreshToken": {}, // Token refresh uses cookie, must be accessible when access token expired

	// Instance Service - needed before login to show instance info
	"/memos.api.v1.InstanceService/GetInstanceProfile": {},
	"/memos.api.v1.InstanceService/GetInstanceSetting": {},

	// User Service - public user profiles and stats
	"/memos.api.v1.UserService/CreateUser":       {}, // Allow first user registration
	"/memos.api.v1.UserService/GetUser":          {},
	"/memos.api.v1.UserService/GetUserAvatar":    {},
	"/memos.api.v1.UserService/GetUserStats":     {},
	"/memos.api.v1.UserService/ListAllUserStats": {},
	"/memos.api.v1.UserService/SearchUsers":      {},

	// Identity Provider Service - SSO buttons on login page
	"/memos.api.v1.IdentityProviderService/ListIdentityProviders": {},

	// Memo Service - public memos (visibility filtering done in service layer)
	"/memos.api.v1.MemoService/GetMemo":          {},
	"/memos.api.v1.MemoService/ListMemos":        {},
	"/memos.api.v1.MemoService/ListMemoComments": {},
}

// GuestBlockedWriteMethods defines mutation RPCs that guest users are not allowed to call.
var GuestBlockedWriteMethods = map[string]struct{}{
	"/memos.api.v1.MemoService/CreateMemo":         {},
	"/memos.api.v1.MemoService/UpdateMemo":         {},
	"/memos.api.v1.MemoService/DeleteMemo":         {},
	"/memos.api.v1.MemoService/SetMemoAttachments": {},
	"/memos.api.v1.MemoService/SetMemoRelations":   {},
	"/memos.api.v1.MemoService/CreateMemoComment":  {},
	"/memos.api.v1.MemoService/UpsertMemoReaction": {},
	"/memos.api.v1.MemoService/DeleteMemoReaction": {},

	"/memos.api.v1.AttachmentService/CreateAttachment": {},
	"/memos.api.v1.AttachmentService/UpdateAttachment": {},
	"/memos.api.v1.AttachmentService/DeleteAttachment": {},

	"/memos.api.v1.ShortcutService/CreateShortcut": {},
	"/memos.api.v1.ShortcutService/UpdateShortcut": {},
	"/memos.api.v1.ShortcutService/DeleteShortcut": {},

	"/memos.api.v1.UserService/UpdateUser":                {},
	"/memos.api.v1.UserService/DeleteUser":                {},
	"/memos.api.v1.UserService/UpdateUserSetting":         {},
	"/memos.api.v1.UserService/CreatePersonalAccessToken": {},
	"/memos.api.v1.UserService/DeletePersonalAccessToken": {},
	"/memos.api.v1.UserService/CreateUserWebhook":         {},
	"/memos.api.v1.UserService/UpdateUserWebhook":         {},
	"/memos.api.v1.UserService/DeleteUserWebhook":         {},
	"/memos.api.v1.UserService/UpdateUserNotification":    {},
	"/memos.api.v1.UserService/DeleteUserNotification":    {},

	"/memos.api.v1.InstanceService/UpdateInstanceSetting": {},

	"/memos.api.v1.IdentityProviderService/CreateIdentityProvider": {},
	"/memos.api.v1.IdentityProviderService/UpdateIdentityProvider": {},
	"/memos.api.v1.IdentityProviderService/DeleteIdentityProvider": {},
}

var guestBlockedGatewayRoutePatterns = []struct {
	method  string
	pattern *regexp.Regexp
}{
	{method: "POST", pattern: regexp.MustCompile(`^/api/v1/memos/?$`)},
	{method: "PATCH", pattern: regexp.MustCompile(`^/api/v1/memos/[^/]+/?$`)},
	{method: "DELETE", pattern: regexp.MustCompile(`^/api/v1/memos/[^/]+/?$`)},
	{method: "POST", pattern: regexp.MustCompile(`^/api/v1/memos/[^/]+/attachments/?$`)},
	{method: "POST", pattern: regexp.MustCompile(`^/api/v1/memos/[^/]+/relations/?$`)},
	{method: "POST", pattern: regexp.MustCompile(`^/api/v1/memos/[^/]+/comments/?$`)},
	{method: "POST", pattern: regexp.MustCompile(`^/api/v1/memos/[^/]+/reactions/?$`)},
	{method: "DELETE", pattern: regexp.MustCompile(`^/api/v1/memos/[^/]+/reactions/[^/]+/?$`)},
	{method: "POST", pattern: regexp.MustCompile(`^/api/v1/attachments/?$`)},
	{method: "PATCH", pattern: regexp.MustCompile(`^/api/v1/attachments/[^/]+/?$`)},
	{method: "DELETE", pattern: regexp.MustCompile(`^/api/v1/attachments/[^/]+/?$`)},
	{method: "POST", pattern: regexp.MustCompile(`^/api/v1/shortcuts/?$`)},
	{method: "PATCH", pattern: regexp.MustCompile(`^/api/v1/shortcuts/[^/]+/?$`)},
	{method: "DELETE", pattern: regexp.MustCompile(`^/api/v1/shortcuts/[^/]+/?$`)},
	{method: "PATCH", pattern: regexp.MustCompile(`^/api/v1/users/[^/]+/?$`)},
	{method: "DELETE", pattern: regexp.MustCompile(`^/api/v1/users/[^/]+/?$`)},
	{method: "PATCH", pattern: regexp.MustCompile(`^/api/v1/users/[^/]+/setting/?$`)},
	{method: "POST", pattern: regexp.MustCompile(`^/api/v1/users/[^/]+/access_tokens/?$`)},
	{method: "DELETE", pattern: regexp.MustCompile(`^/api/v1/users/[^/]+/access_tokens/[^/]+/?$`)},
	{method: "POST", pattern: regexp.MustCompile(`^/api/v1/users/[^/]+/webhooks/?$`)},
	{method: "PATCH", pattern: regexp.MustCompile(`^/api/v1/users/[^/]+/webhooks/[^/]+/?$`)},
	{method: "DELETE", pattern: regexp.MustCompile(`^/api/v1/users/[^/]+/webhooks/[^/]+/?$`)},
	{method: "PATCH", pattern: regexp.MustCompile(`^/api/v1/users/[^/]+/notifications/[^/]+/?$`)},
	{method: "DELETE", pattern: regexp.MustCompile(`^/api/v1/users/[^/]+/notifications/[^/]+/?$`)},
	{method: "PATCH", pattern: regexp.MustCompile(`^/api/v1/instance/setting/?$`)},
	{method: "POST", pattern: regexp.MustCompile(`^/api/v1/identity_providers/?$`)},
	{method: "PATCH", pattern: regexp.MustCompile(`^/api/v1/identity_providers/[^/]+/?$`)},
	{method: "DELETE", pattern: regexp.MustCompile(`^/api/v1/identity_providers/[^/]+/?$`)},
}

// IsPublicMethod checks if a procedure path is public (no authentication required).
// Returns true for public methods, false for protected methods.
func IsPublicMethod(procedure string) bool {
	_, ok := PublicMethods[procedure]
	return ok
}

// IsGuestWriteBlockedMethod checks if the RPC is a write method blocked for guest users.
func IsGuestWriteBlockedMethod(procedure string) bool {
	_, ok := GuestBlockedWriteMethods[procedure]
	return ok
}

func IsGuestWriteBlockedGatewayRequest(method, path string) bool {
	normalizedMethod := strings.ToUpper(method)
	for _, route := range guestBlockedGatewayRoutePatterns {
		if route.method == normalizedMethod && route.pattern.MatchString(path) {
			return true
		}
	}
	return false
}
