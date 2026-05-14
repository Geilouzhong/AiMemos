package v1

import (
	"context"
	"net/http"

	"connectrpc.com/connect"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/usememos/memos/proto/gen/api/v1/apiv1connect"
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
}

// NewConnectServiceHandler creates a new Connect service handler.
func NewConnectServiceHandler(svc *APIV1Service) *ConnectServiceHandler {
	return &ConnectServiceHandler{
		APIV1Service: svc,
	}
}

// RegisterConnectHandlers registers all Connect service handlers on the given mux.
func (s *ConnectServiceHandler) RegisterConnectHandlers(mux *http.ServeMux, opts ...connect.HandlerOption) {
	// Register all service handlers
	handlers := []struct {
		path    string
		handler http.Handler
	}{
		wrapConnectHandler(apiv1connect.NewInstanceServiceHandler(s, opts...)),
		wrapConnectHandler(apiv1connect.NewAuthServiceHandler(s, opts...)),
		wrapConnectHandler(apiv1connect.NewUserServiceHandler(s, opts...)),
		wrapConnectHandler(apiv1connect.NewMemoServiceHandler(s, opts...)),
		wrapConnectHandler(apiv1connect.NewAttachmentServiceHandler(s, opts...)),
		wrapConnectHandler(apiv1connect.NewShortcutServiceHandler(s, opts...)),
		wrapConnectHandler(apiv1connect.NewActivityServiceHandler(s, opts...)),
		wrapConnectHandler(apiv1connect.NewIdentityProviderServiceHandler(s, opts...)),
	}

	for _, h := range handlers {
		mux.Handle(h.path, h.handler)
	}
}

// wrapConnectHandler converts (path, handler) return value to a struct for cleaner iteration.
// 同时创建一个中间件，将 HTTP Request 存入 context，供后续的 activity logging 使用。
func wrapConnectHandler(path string, handler http.Handler) struct {
	path    string
	handler http.Handler
} {
	// 创建中间件，将 HTTP Request 存入 context
	wrappedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
