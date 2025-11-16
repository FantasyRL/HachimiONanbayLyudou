package constant

import "time"

const (
	MCPTransportStdio          = "stdio"          // MCP基于标准输入输出连接
	MCPTransportSSE            = "sse"            // MCP基于SSE连接
	MCPTransportHTTP           = "http"           // MCP基于http连接
	MCPClientInitTimeout       = 10 * time.Second // MCP客户端初始化超时时间
	MCPDefaultCallTimeout      = 30 * time.Second // MCP调用默认超时时间
	MCPServerHeartbeatInterval = 25 * time.Second // MCP服务器心跳间隔

	AiProviderModeLocal   = "local"  // 本地模型
	AiProviderModeRemote  = "remote" // 远程模型
	FzuHelperServerMCPUrl = "https://fzuhelper.west2.online/mcp"
)
