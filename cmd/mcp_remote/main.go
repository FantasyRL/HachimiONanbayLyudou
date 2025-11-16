package main

import (
	"flag"
	"github.com/FantasyRL/go-mcp-demo/config"
	"github.com/FantasyRL/go-mcp-demo/internal/mcp/application"
	"github.com/FantasyRL/go-mcp-demo/pkg/base/mcp_server"
	"github.com/FantasyRL/go-mcp-demo/pkg/base/tool_set"
	"github.com/FantasyRL/go-mcp-demo/pkg/constant"
	"github.com/FantasyRL/go-mcp-demo/pkg/logger"
	"github.com/FantasyRL/go-mcp-demo/pkg/utils"
)

var (
	serviceName = constant.ServiceNameMCPRemote
	configPath  = flag.String("cfg", "config/config.yaml", "config file path")
	toolSet     = new(tool_set.ToolSet)
)

func init() {
	flag.Parse()
	config.Load(*configPath, serviceName)
	logger.Init(serviceName, config.GetLoggerLevel())
	toolSet = tool_set.NewToolSet(application.WithTimeTool())
}

func main() {
	logger.Infof("starting mcp server, transport = %s", config.MCP.Transport)
	coreServer := mcp_server.NewCoreServer(config.MCP.ServerName, config.MCP.Transport, toolSet, nil)
	switch config.MCP.Transport {
	case constant.MCPTransportStdio:
		if err := mcp_server.ServeStdio(coreServer); err != nil {
			logger.Errorf("serve stdio: %v", err)
			return
		}
	// streamable HTTP 启动
	case constant.MCPTransportHTTP:
		addr, err := utils.GetAvailablePort()
		if err != nil {
			logger.Errorf("mcp_server: get available port failed, err: %v", err)
			return
		}
		logger.Infof("mcp_server: http server listening at %s", addr)
		if err := mcp_server.NewStreamableHTTPServer(coreServer, serviceName, addr).Start(addr); err != nil {
			logger.Errorf("serve http: %v", err)
			return
		}
	default:
		logger.Errorf("mcp_server: unknown transport type: %s", config.MCP.Transport)
		return
	}
}
