package dev_runner

import (
	"context"
	"fmt"
	"github.com/FantasyRL/go-mcp-demo/pkg/utils"
	"github.com/mark3labs/mcp-go/mcp"
)

func HandleFsCat(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	p, _ := args["path"].(string)
	if p == "" {
		return mcp.NewToolResultError("missing required arg: path"), nil
	}
	maxF, _ := args["max_bytes"].(float64)
	maxBytes := 64 * 1024
	if maxF > 0 {
		maxBytes = int(maxF)
	}
	content, truncated, err := utils.ReadFileMax(p, maxBytes)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	header := fmt.Sprintf("### fs_cat: %s (max_bytes=%d, truncated=%v)\n\n", p, maxBytes, truncated)
	return mcp.NewToolResultText(header + content), nil
}
