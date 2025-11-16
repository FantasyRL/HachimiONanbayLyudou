package mcp_client

import (
	"context"
	"fmt"
	"github.com/FantasyRL/go-mcp-demo/config"
	"github.com/FantasyRL/go-mcp-demo/pkg/constant"
	mcpc "github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// newStdioMCPClient 通过 stdio 连接
func newStdioMCPClient() (*MCPClient, error) {
	cmd := config.MCP.Stdio.ServerCmd
	if cmd == "" {
		cmd = "./bin/mcp-server"
	}
	client, err := mcpc.NewStdioMCPClient(cmd, nil, config.MCP.Stdio.ServerArgs...)
	if err != nil {
		return nil, fmt.Errorf("start stdio client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), constant.MCPClientInitTimeout)
	defer cancel()

	_, err = client.Initialize(ctx, mcp.InitializeRequest{
		Params: mcp.InitializeParams{
			ClientInfo: mcp.Implementation{
				Name:    "mcp-host",
				Version: "0.1.0",
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("initialize mcp (stdio): %w", err)
	}

	res, err := client.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		return nil, fmt.Errorf("list tools: %w", err)
	}
	return &MCPClient{Client: client, Tools: res.Tools}, nil
}
