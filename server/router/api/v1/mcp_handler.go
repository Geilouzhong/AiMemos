package v1

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"

	"github.com/Geilouzhong/AiMemos/internal/profile"
	"github.com/Geilouzhong/AiMemos/server/auth"
)

// MCPHandler handles MCP (Model Context Protocol) requests.
type MCPHandler struct {
	*APIV1Service
	profile *profile.Profile
}

// NewMCPHandler creates a new MCP handler.
func NewMCPHandler(svc *APIV1Service, profile *profile.Profile) *MCPHandler {
	return &MCPHandler{
		APIV1Service: svc,
		profile:      profile,
	}
}

// RegisterMCPEndpoint registers the MCP endpoint with Echo server.
func (h *MCPHandler) RegisterMCPEndpoint(echoServer *echo.Echo) error {
	echoServer.GET(h.profile.MCPPath, h.HandleSSEConnection)
	echoServer.POST(h.profile.MCPPath, h.HandleHTTPToolCall)
	return nil
}

// HandleSSEConnection handles SSE connections from AI agents.
func (h *MCPHandler) HandleSSEConnection(c echo.Context) error {
	// 1. 设置 SSE headers
	c.Response().Header().Set("Content-Type", "text/event-stream")
	c.Response().Header().Set("Cache-Control", "no-cache")
	c.Response().Header().Set("Connection", "keep-alive")
	c.Response().Header().Set("Access-Control-Allow-Origin", "*")
	c.Response().Header().Set("X-Accel-Buffering", "no")

	// 2. 获取用户认证
	ctx := c.Request().Context()
	userID, ok := ctx.Value(auth.UserIDContextKey).(int32)
	if !ok {
		// 尝试从 Authorization header 获取 token（用于 Claude Desktop MCP）
		authHeader := c.Request().Header.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			// 使用 Authenticator 同时支持 JWT 和 PAT
			authenticator := auth.NewAuthenticator(h.Store, h.Secret)
			result := authenticator.Authenticate(ctx, authHeader)
			if result == nil {
				return sendSSEError(c, "AUTHENTICATION_REQUIRED", "")
			}
			// 提取 userID（支持 JWT claims 和 PAT user）
			if result.Claims != nil {
				// JWT Access Token V2
				userID = result.Claims.UserID
				ctx = auth.SetUserClaimsInContext(ctx, result.Claims)
				ctx = context.WithValue(ctx, auth.UserIDContextKey, result.Claims.UserID)
			} else if result.User != nil {
				// Personal Access Token
				userID = result.User.ID
				ctx = auth.SetUserInContext(ctx, result.User, result.AccessToken)
			} else {
				return sendSSEError(c, "AUTHENTICATION_REQUIRED", "")
			}
		} else {
			return sendSSEError(c, "AUTHENTICATION_REQUIRED", "")
		}
	}

	// 3. 发送初始化事件（工具列表）
	if err := sendSSEEvent(c, "tools/list", GetToolsList()); err != nil {
		return errors.Wrap(err, "failed to send tools list")
	}

	// 4. 保持连接，处理消息
	flusher, ok := c.Response().Writer.(http.Flusher)
	if !ok {
		return errors.New("streaming not supported")
	}
	flusher.Flush()

	// 5. 启动心跳协程
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go heartbeatLoop(ctx, c, flusher)

	// 6. 监听客户端消息
	scanner := bufio.NewScanner(c.Request().Body)
	for scanner.Scan() {
		line := scanner.Text()

		// 跳过空行和注释
		if line == "" || strings.HasPrefix(line, ":") {
			continue
		}

		// 解析 JSON-RPC 请求
		var req map[string]interface{}
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			sendSSEError(c, "INVALID_REQUEST", "")
			continue
		}

		// 处理请求
		if err := h.handleMCPRequest(ctx, c, req, userID); err != nil {
			logMCPError("handleMCPRequest", err)
		}

		flusher.Flush()
	}

	return nil
}

// sendSSEEvent sends an SSE event to the client.
func sendSSEEvent(c echo.Context, event string, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return errors.Wrap(err, "failed to marshal data")
	}

	w := c.Response()
	fmt.Fprintf(w, "event: %s\n", event)
	fmt.Fprintf(w, "data: %s\n\n", jsonData)

	return nil
}

// sendSSEError sends a localized error event to the client.
func sendSSEError(c echo.Context, code string, message string) error {
	// 如果提供了自定义消息，使用自定义消息
	// 否则使用本地化消息
	finalMessage := message
	if message == "" {
		finalMessage = getLocalizedMessage(code, c.Request())
	}

	return sendSSEEvent(c, "error", map[string]interface{}{
		"code":    code,
		"message": finalMessage,
	})
}

// heartbeatLoop sends periodic ping events to keep the connection alive.
func heartbeatLoop(ctx context.Context, c echo.Context, flusher http.Flusher) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := sendSSEEvent(c, "ping", map[string]interface{}{
				"timestamp": time.Now().Unix(),
			}); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

// logError logs errors with context.
func logMCPError(context string, err error) {
	fmt.Printf("[MCP Error] %s: %v\n", context, err)
}

// toolHandlerMap maps tool names to their handler functions.
var toolHandlerMap = map[string]func(*MCPHandler, context.Context, int32, map[string]interface{}) (map[string]interface{}, error){
	"list_memos":  (*MCPHandler).handleListMemos,
	"pull_memo":   (*MCPHandler).handlePullMemo,
	"push_memo":   (*MCPHandler).handlePushMemo,
}

// handleMCPRequest processes an incoming MCP request.
func (h *MCPHandler) handleMCPRequest(ctx context.Context, c echo.Context, req map[string]interface{}, userID int32) error {
	method, ok := req["method"].(string)
	if !ok {
		return sendSSEError(c, "INVALID_REQUEST", "")
	}

	if method == "tools/call" {
		return h.handleToolCall(ctx, c, req, userID)
	}

	return sendSSEError(c, "UNKNOWN_METHOD", fmt.Sprintf("unknown method: %s", method))
}

// handleToolCall processes a tool/call request.
func (h *MCPHandler) handleToolCall(ctx context.Context, c echo.Context, req map[string]interface{}, userID int32) error {
	params, ok := req["params"].(map[string]interface{})
	if !ok {
		return sendSSEError(c, "INVALID_REQUEST", "")
	}

	name, ok := params["name"].(string)
	if !ok {
		return sendSSEError(c, "INVALID_REQUEST", "")
	}

	arguments, ok := params["arguments"].(map[string]interface{})
	if !ok {
		arguments = make(map[string]interface{})
	}

	// 验证工具调用
	if err := ValidateToolCall(name, arguments); err != nil {
		return sendSSEError(c, "INVALID_ARGUMENTS", err.Error())
	}

	// 查找并执行处理器
	handler, ok := toolHandlerMap[name]
	if !ok {
		return sendSSEError(c, "UNKNOWN_TOOL", fmt.Sprintf("unknown tool: %s", name))
	}

	result, err := handler(h, ctx, userID, arguments)
	if err != nil {
		return sendServiceError(c, err)
	}

	return sendSSEEvent(c, "tools/call/response", map[string]interface{}{
		"result": result,
	})
}

// sendServiceError converts service errors to MCP errors.
func sendServiceError(c echo.Context, err error) error {
	// 简化版本：直接返回错误
	// 下一任务会实现完整的 gRPC 错误转换
	return sendSSEError(c, "INTERNAL_ERROR", err.Error())
}

// HandleHTTPToolCall handles standard MCP JSON-RPC over HTTP requests.
func (h *MCPHandler) HandleHTTPToolCall(c echo.Context) error {
	// 1. 认证
	ctx := c.Request().Context()
	authHeader := c.Request().Header.Get("Authorization")

	authenticator := auth.NewAuthenticator(h.Store, h.Secret)
	result := authenticator.Authenticate(ctx, authHeader)
	if result == nil {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"jsonrpc": "2.0",
			"error": map[string]interface{}{
				"code":    -32600,
				"message": "Unauthorized",
			},
		})
	}

	// 设置用户上下文
	if result.Claims != nil {
		ctx = auth.SetUserClaimsInContext(ctx, result.Claims)
		ctx = context.WithValue(ctx, auth.UserIDContextKey, result.Claims.UserID)
		c.SetRequest(c.Request().Clone(ctx))
	} else if result.User != nil {
		ctx = auth.SetUserInContext(ctx, result.User, result.AccessToken)
		c.SetRequest(c.Request().Clone(ctx))
	}

	// 2. 解析 JSON-RPC 请求
	var req map[string]interface{}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"jsonrpc": "2.0",
			"error": map[string]interface{}{
				"code":    -32700,
				"message": "Parse error",
			},
		})
	}

	// 3. 验证 JSON-RPC 格式
	if req["jsonrpc"] != "2.0" {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"jsonrpc": "2.0",
			"error": map[string]interface{}{
				"code":    -32600,
				"message": "Invalid Request",
			},
			"id": req["id"],
		})
	}

	// 4. 提取 userID
	var userID int32
	if result.Claims != nil {
		userID = result.Claims.UserID
	} else if result.User != nil {
		userID = result.User.ID
	}

	// 5. 处理不同的方法
	methodValue, ok := req["method"]
	method, methodOK := methodValue.(string)
	if !ok || !methodOK {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"jsonrpc": "2.0",
			"error": map[string]interface{}{
				"code":    -32600,
				"message": "Invalid Request",
			},
			"id": req["id"],
		})
	}
	id := req["id"]
	params := map[string]interface{}{}
	if paramsValue, ok := req["params"]; ok {
		if castParams, ok := paramsValue.(map[string]interface{}); ok {
			params = castParams
		}
	}

	switch method {
	case "initialize":
		return c.JSON(http.StatusOK, map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      id,
			"result": map[string]interface{}{
				"protocolVersion": "2024-11-05",
				"capabilities": map[string]interface{}{
					"tools": map[string]interface{}{},
				},
				"serverInfo": map[string]interface{}{
					"name":    "memos-mcp-server",
					"version": "0.1.0",
				},
			},
		})

	case "notifications/initialized":
		// MCP notifications do not have an id and must not receive a JSON-RPC response.
		// Streamable HTTP clients expect a 202 to acknowledge the notification.
		return c.NoContent(http.StatusAccepted)

	case "tools/list":
		return c.JSON(http.StatusOK, map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      id,
			"result":  GetToolsList(),
		})

	case "tools/call":
		// 调用工具
		nameValue, ok := params["name"]
		name, nameOK := nameValue.(string)
		if !ok || !nameOK {
			return c.JSON(http.StatusOK, map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      id,
				"error": map[string]interface{}{
					"code":    -32602,
					"message": "missing tool name",
				},
			})
		}

		arguments := map[string]interface{}{}
		if argumentsValue, ok := params["arguments"]; ok {
			if castArguments, ok := argumentsValue.(map[string]interface{}); ok {
				arguments = castArguments
			}
		}

		if err := ValidateToolCall(name, arguments); err != nil {
			return c.JSON(http.StatusOK, map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      id,
				"error": map[string]interface{}{
					"code":    -32602,
					"message": err.Error(),
				},
			})
		}

		handler, ok := toolHandlerMap[name]
		if !ok {
			return c.JSON(http.StatusOK, map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      id,
				"error": map[string]interface{}{
					"code":    -32601,
					"message": fmt.Sprintf("unknown tool: %s", name),
				},
			})
		}

		result, err := handler(h, ctx, userID, arguments)
		if err != nil {
			return c.JSON(http.StatusOK, map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      id,
				"error": map[string]interface{}{
					"code":    -32603,
					"message": err.Error(),
				},
			})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      id,
			"result":  result,
		})

	default:
		return c.JSON(http.StatusOK, map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      id,
			"error": map[string]interface{}{
				"code":    -32601,
				"message": fmt.Sprintf("method not found: %s", method),
			},
		})
	}
}
