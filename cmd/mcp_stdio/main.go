package main

// 旧代码，留档
//import (
//	"context"
//	"flag"
//	"github.com/FantasyRL/go-mcp-demo/config"
//	"github.com/FantasyRL/go-mcp-demo/pkg/logger"
//	"time"
//
//	"github.com/mark3labs/mcp-go/mcp"
//	"github.com/mark3labs/mcp-go/server"
//)
//
//var (
//	serviceName = "mcp_server"
//	configPath  = flag.String("cfg", "config/config.yaml", "config file path")
//)
//
//func init() {
//	flag.Parse()
//	config.Load(*configPath, serviceName)
//	logger.Init(serviceName, config.GetLoggerLevel())
//}
//func main() {
//	logger.Infof("[mcp] server=%s transport=%s", config.MCP.ServerName, config.MCP.Transport)
//
//	s := server.NewMCPServer(
//		config.MCP.ServerName,
//		"1.0.0",
//		server.WithToolCapabilities(false),
//		server.WithRecovery(),
//	)
//
//	// time_now 工具
//	tool := mcp.NewTool("time_now",
//		mcp.WithDescription("返回当前时间（RFC3339）"),
//	)
//	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
//		now := time.Now().Format(time.RFC3339)
//		return mcp.NewToolResultText(now), nil
//	})
//
//	// stdio 客户端通过启动 Server 子进程，通过标准输入输出进行消息交换
//	if err := server.ServeStdio(s); err != nil {
//		logger.Fatalf("serve stdio: %v", err)
//	}
//}
