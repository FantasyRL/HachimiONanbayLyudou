package base

import (
	"github.com/FantasyRL/go-mcp-demo/config"
	"github.com/FantasyRL/go-mcp-demo/pkg/base/ai_provider"
	"github.com/FantasyRL/go-mcp-demo/pkg/base/db"
	"github.com/FantasyRL/go-mcp-demo/pkg/base/mcp_client"
	"github.com/FantasyRL/go-mcp-demo/pkg/base/registry/consul"
	"github.com/FantasyRL/go-mcp-demo/pkg/constant"
	"log"
)

// WithMCPClient 通过配置手动注入初始化 ClientSet.MCPCli。
// - stdio: 直接创建单连接客户端（本地进程/stdio）
// - none(单点): 使用 config.MCP.HTTP.BaseURL 创建单连接客户端
// - consul: 创建聚合客户端（基于 Consul Resolver 定时刷新，自动发现多实例）
func WithMCPClient(services []string) Option {
	return func(clientSet *ClientSet) {
		switch {
		// stdio 启动（本地）
		case config.MCP.Transport == constant.MCPTransportStdio:
			mcpCli, err := mcp_client.NewMCPClient("")
			if err != nil {
				log.Fatalf("failed to create stdio mcp client: %s", err)
			}
			clientSet.MCPCli = mcpCli

		// 单点通信：直接使用配置的 BaseURL
		case config.Registry.Provider == constant.RegistryProviderNone:
			if config.MCP.HTTP.BaseURL == "" {
				log.Fatalf("missing MCP HTTP BaseURL while registry provider is 'none'")
			}
			mcpCli, err := mcp_client.NewMCPClient(config.MCP.HTTP.BaseURL)
			if err != nil {
				log.Fatalf("failed to create http mcp client: %s", err)
			}
			clientSet.MCPCli = mcpCli

		// 服务发现（Consul）：使用聚合客户端，多路连接 + 定时刷新
		case config.Registry.Provider == constant.RegistryProviderConsul:
			resolver := consul.NewResolver()
			if resolver == nil {
				log.Fatalf("consul config invalid, can't create MCP client")
			}
			ac := mcp_client.NewAggregatedClient(resolver, services) // 可按需调整刷新周期
			clientSet.RegistryResolver = resolver
			clientSet.MCPCli = ac
			clientSet.cleanups = append(clientSet.cleanups, ac.Close)

		default:
			log.Fatalf("unknown registry provider: %s, can't create MCP client", config.Registry.Provider)
		}
	}
}

func WithAiProviderClient() Option {
	return func(clientSet *ClientSet) {
		cli := ai_provider.NewAiProviderClient()
		clientSet.AiProviderCli = cli
	}
}

func WithDB() Option {
	return func(clientSet *ClientSet) {
		actualDB, err := db.InitDBClient()
		if err != nil {
			log.Fatalf("failed to initialize gorm db: %s", err)
		}
		clientSet.ActualDB = actualDB
	}
}
