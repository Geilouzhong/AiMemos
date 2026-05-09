package v1

import (
	"fmt"
	"net"
	"os"
	"sync"
	"time"
)

var (
	activityLogFile     *os.File
	activityLogMutex     sync.Mutex
	activityLogPath      = "logs/memo_activity.log"
	activityLoggerOnce   sync.Once
	activityLoggerInitErr error
)

// initActivityLogger 初始化日志文件（线程安全的单例模式）
func initActivityLogger() error {
	activityLoggerOnce.Do(func() {
		if activityLogFile != nil {
			return
		}

		// 创建 logs 目录
		if err := os.MkdirAll("logs", 0755); err != nil {
			activityLoggerInitErr = fmt.Errorf("failed to create logs directory: %w", err)
			return
		}

		// 打开日志文件（追加模式）
		file, err := os.OpenFile(activityLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0640)
		if err != nil {
			activityLoggerInitErr = fmt.Errorf("failed to open activity log file: %w", err)
			return
		}

		activityLogFile = file
	})

	return activityLoggerInitErr
}

// logMemoActivity 记录 memo 查看活动到日志文件
// 线程安全，可从多个 goroutine 并发调用
func logMemoActivity(userID int32, memoID string, username string, ip string, port int) error {
	if err := initActivityLogger(); err != nil {
		return err
	}

	activityLogMutex.Lock()
	defer activityLogMutex.Unlock()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logLine := fmt.Sprintf("[%s] user_id=%d memo_id=%s ip=%s:%d username=%s\n",
		timestamp, userID, memoID, ip, port, username)

	if _, err := activityLogFile.WriteString(logLine); err != nil {
		return fmt.Errorf("failed to write activity log: %w", err)
	}

	return nil
}

// extractIPAndPort 从 RemoteAddr 字符串中提取 IP 和端口
// RemoteAddr 格式通常为 "IP:port" 或 "IP"
func extractIPAndPort(remoteAddr string) (string, int) {
	host, portStr, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		// 如果没有端口，返回整个字符串作为 IP，端口为 0
		return remoteAddr, 0
	}

	port := 0
	if portStr != "" {
		fmt.Sscanf(portStr, "%d", &port)
	}

	return host, port
}
