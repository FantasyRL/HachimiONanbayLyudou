package mcp_client

import (
	"context"
	"github.com/openai/openai-go/v2"
)

// ToolClient 定义了调用 MCP 工具客户端的接口
type ToolClient interface {
	// ConvertToolsToOllama 将MCP工具定义转换为 Ollama 工具格式
	ConvertToolsToOllama() []map[string]any
	// ConvertToolsToOpenAI 将MCP工具定义转换为 OpenAI Chat Completions 的 tools 参数
	ConvertToolsToOpenAI() []openai.ChatCompletionToolUnionParam
	// CallTool 调用工具
	CallTool(ctx context.Context, name string, args any) (string, error)
	// Close 关闭客户端连接
	Close()
}
