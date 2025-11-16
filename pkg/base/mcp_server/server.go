package mcp_server

import (
	"github.com/FantasyRL/go-mcp-demo/config"
	"github.com/FantasyRL/go-mcp-demo/pkg/base/prompt_set"
	"github.com/FantasyRL/go-mcp-demo/pkg/base/registry"
	"github.com/FantasyRL/go-mcp-demo/pkg/base/registry/consul"
	"github.com/FantasyRL/go-mcp-demo/pkg/base/tool_set"
	"github.com/FantasyRL/go-mcp-demo/pkg/constant"
	"github.com/FantasyRL/go-mcp-demo/pkg/logger"
	"github.com/FantasyRL/go-mcp-demo/pkg/utils"
	"github.com/google/uuid"
	"time"

	"github.com/mark3labs/mcp-go/server"
)

// HTTPOpts：Streamable HTTP(含 SSE) 选项
type HTTPOpts struct {
	// EndpointPath 仅对 shttp.Start(":8080") 的一行启动生效；
	// 若作为 http.Handler 挂到 mux，路由由 mux 决定，该字段不生效。
	EndpointPath      string
	HeartbeatInterval time.Duration // 建议 20~30s，降低中间件 idle 断开
}

// NewCoreServer 在此注册 tools/prompts/resources
func NewCoreServer(name, version string, toolSet *tool_set.ToolSet, promptSet *prompt_set.PromptSet) *server.MCPServer {
	s := server.NewMCPServer(
		name,
		version,
		server.WithRecovery(),
		server.WithToolCapabilities(false),
	)

	if toolSet != nil {
		for _, t := range toolSet.Tools {
			s.AddTool(*t, toolSet.HandlerFunc[t.Name])
		}
	}
	if promptSet != nil {
		for _, p := range promptSet.Prompts {
			s.AddPrompt(*p, promptSet.HandlerFunc[p.Name])
		}
	}

	return s
}

// NewStreamableHTTPServer 基于核心 Server 创建StreamableHTTP服务器组件
func NewStreamableHTTPServer(core *server.MCPServer, serviceName string, addr string) *server.StreamableHTTPServer {
	switch config.Registry.Provider {
	case constant.RegistryProviderConsul:
		registrar := consul.NewRegistrar(serviceName)
		id, _ := uuid.NewV7()
		_, err := registrar.Register(&registry.Registration{
			Service: serviceName,
			ID:      id.String(),
			Address: addr,
			Port:    utils.AddrGetPort(addr),
			Tags:    []string{constant.RegistryMCPTag},
			Meta:    map[string]string{"addr": addr},
			Path:    constant.RegistryMCPDefaultPath,
		})
		if err != nil {
			panic("mcp_server: consul register failed, err: " + err.Error())
		}
		logger.Infof("%s : registered to consul successfully on %s", serviceName, addr)
	default:
	}
	var httpOpts []server.StreamableHTTPOption
	httpOpts = append(httpOpts, server.WithHeartbeatInterval(constant.MCPServerHeartbeatInterval))
	return server.NewStreamableHTTPServer(core, httpOpts...)
}

// ServeStdio stdio
func ServeStdio(core *server.MCPServer) error {
	return server.ServeStdio(core)
}

// NewHTTPSSEServer [MCP规范已废弃]基于核心 Server 创建 SSE 服务器组件
func NewHTTPSSEServer(core *server.MCPServer) *server.SSEServer {
	var sseOpts []server.SSEOption
	sseOpts = append(sseOpts, server.WithKeepAliveInterval(constant.MCPServerHeartbeatInterval))
	return server.NewSSEServer(core, sseOpts...)
}
