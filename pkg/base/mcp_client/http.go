package mcp_client

import (
	"context"
	"fmt"
	"github.com/FantasyRL/go-mcp-demo/pkg/constant"
	mcpc "github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// newSSEMCPClientWithConn [MCP规范已废弃]通过 SSE 连接指定 URL
func newSSEMCPClientWithConn(url string) (*MCPClient, error) {
	c, err := mcpc.NewStreamableHttpClient(url)
	if err != nil {
		return nil, fmt.Errorf("new sse client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), constant.MCPClientInitTimeout)
	defer cancel()

	if err := c.Start(ctx); err != nil {
		return nil, fmt.Errorf("sse start: %w", err)
	}
	_, err = c.Initialize(ctx, mcp.InitializeRequest{
		Params: mcp.InitializeParams{
			ClientInfo: mcp.Implementation{Name: "mcp-host", Version: "0.1.0"},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("initialize (sse): %w", err)
	}

	resTool, err := c.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		return nil, fmt.Errorf("list tools: %w", err)
	}

	return &MCPClient{Client: c, Tools: resTool.Tools}, nil
}

// newHTTPMCPClientWithConn 通过 Streamable HTTP 连接指定 URL
func newHTTPMCPClientWithConn(url string) (*MCPClient, error) {
	c, err := mcpc.NewStreamableHttpClient(url)
	if err != nil {
		return nil, fmt.Errorf("new http client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), constant.MCPClientInitTimeout)
	defer cancel()

	if err := c.Start(ctx); err != nil {
		return nil, fmt.Errorf("http start: %w", err)
	}
	_, err = c.Initialize(ctx, mcp.InitializeRequest{
		Params: mcp.InitializeParams{
			ClientInfo: mcp.Implementation{Name: "mcp-host", Version: "0.1.0"},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("initialize (http): %w", err)
	}

	resTool, err := c.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		return nil, fmt.Errorf("list tools: %w", err)
	}

	return &MCPClient{Client: c, Tools: resTool.Tools}, nil
}
