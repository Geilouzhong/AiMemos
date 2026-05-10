package v1

import (
	"context"
	"net/http"
	"strings"

	"connectrpc.com/connect"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/usememos/memos/proto/gen/api/v1/apiv1connect"
	"github.com/usememos/memos/server/auth"
)

// ConnectServiceHandler wraps APIV1Service to implement Connect handler interfaces.
// It adapts the existing gRPC service implementations to work with Connect's
// request/response wrapper types.
//
// This wrapper pattern allows us to:
// - Reuse existing gRPC service implementations
// - Support both native gRPC and Connect protocols
// - Maintain a single source of truth for business logic.
type ConnectServiceHandler struct {
	*APIV1Service
	authenticator *auth.Authenticator
}

// NewConnectServiceHandler creates a new Connect service handler.
func NewConnectServiceHandler(svc *APIV1Service) *ConnectServiceHandler {
	return &ConnectServiceHandler{
		APIV1Service: svc,
		authenticator: auth.NewAuthenticator(svc.Store, svc.Secret),
	}
}

// RegisterConnectHandlers registers all Connect service handlers on the given mux.
func (s *ConnectServiceHandler) RegisterConnectHandlers(mux *http.ServeMux, opts ...connect.HandlerOption) {
	// Register all service handlers
	handlers := []struct {
		path    string
		handler http.Handler
	}{
		s.wrap(apiv1connect.NewInstanceServiceHandler(s, opts...)),
		s.wrap(apiv1connect.NewAuthServiceHandler(s, opts...)),
		s.wrap(apiv1connect.NewUserServiceHandler(s, opts...)),
		s.wrap(apiv1connect.NewMemoServiceHandler(s, opts...)),
		s.wrap(apiv1connect.NewAttachmentServiceHandler(s, opts...)),
		s.wrap(apiv1connect.NewShortcutServiceHandler(s, opts...)),
		s.wrap(apiv1connect.NewActivityServiceHandler(s, opts...)),
		s.wrap(apiv1connect.NewIdentityProviderServiceHandler(s, opts...)),
	}

	for _, h := range handlers {
		mux.Handle(h.path, h.handler)
	}
}

// wrap converts (path, handler) return value to a struct for cleaner iteration.
// 同时创建一个中间件，将 HTTP Request 存入 context，供后续的 activity logging 使用
func (s *ConnectServiceHandler) wrap(path string, handler http.Handler) struct {
	path    string
	handler http.Handler
} {
	// 创建中间件，将 HTTP Request 存入 context
	wrappedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 检查认证状态，对未认证的浏览器请求重定向到登录页
		authHeader := r.Header.Get("Authorization")
		result := s.authenticator.Authenticate(r.Context(), authHeader)

		// 检查是否为公开方法
		isPublic := false
		for _, prefix := range []string{
			"/memos.api.v1.AuthService/",
			"/memos.api.v1.InstanceService/",
		} {
			if strings.HasPrefix(path, prefix) {
				// 进一步检查具体方法
				if path == "/memos.api.v1.AuthService/SignIn" ||
					path == "/memos.api.v1.AuthService/RefreshToken" ||
					path == "/memos.api.v1.InstanceService/GetInstanceProfile" ||
					path == "/memos.api.v1.InstanceService/GetInstanceSetting" ||
					path == "/memos.api.v1.UserService/CreateUser" ||
					path == "/memos.api.v1.IdentityProviderService/ListIdentityProviders" {
					isPublic = true
					break
				}
			}
		}

		// 如果未认证且不是公开方法，检查是否为浏览器请求
		if result == nil && !isPublic {
			acceptHeader := r.Header.Get("Accept")
			userAgent := r.Header.Get("User-Agent")
			isBrowserRequest := (acceptHeader != "" &&
				(strings.Contains(acceptHeader, "text/html") || strings.Contains(acceptHeader, "*/*"))) &&
				!strings.Contains(userAgent, "compatible") &&
				r.URL.Query().Get("f") != "json"

			if isBrowserRequest {
				// 重定向浏览器请求到登录页
				returnURL := r.URL.String()
				loginURL := "/auth/sign-in?redirect=" + returnURL
				http.Redirect(w, r, loginURL, http.StatusFound)
				return
			}
			// API 请求继续，由 interceptor 返回错误
		}

		// 将 HTTP Request 存入 context，使用在 memo_service.go 中定义的 context key
		ctx := context.WithValue(r.Context(), httpRequestContextKey{}, r)
		handler.ServeHTTP(w, r.WithContext(ctx))
	})

	return struct {
		path    string
		handler http.Handler
	}{path, wrappedHandler}
}

// convertGRPCError converts gRPC status errors to Connect errors.
// This preserves the error code semantics between the two protocols.
func convertGRPCError(err error) error {
	if err == nil {
		return nil
	}
	if st, ok := status.FromError(err); ok {
		return connect.NewError(grpcCodeToConnectCode(st.Code()), err)
	}
	return connect.NewError(connect.CodeInternal, err)
}

// grpcCodeToConnectCode converts gRPC status codes to Connect error codes.
// gRPC and Connect use the same error code semantics, so this is a direct cast.
// See: https://connectrpc.com/docs/protocol/#error-codes
func grpcCodeToConnectCode(code codes.Code) connect.Code {
	return connect.Code(code)
}
