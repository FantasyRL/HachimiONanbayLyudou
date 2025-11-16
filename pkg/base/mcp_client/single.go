package mcp_client

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/FantasyRL/go-mcp-demo/config"
	"github.com/FantasyRL/go-mcp-demo/pkg/logger"
	mcpc "github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/packages/param"
)

type MCPClient struct {
	Client *mcpc.Client
	Tools  []mcp.Tool
}

// NewMCPClient 启动 MCP Server 并建立连接
func NewMCPClient(url string) (*MCPClient, error) {
	switch config.MCP.Transport {
	case "stdio", "":
		return newStdioMCPClient()
	case "sse":
		return newSSEMCPClientWithConn(url)
	case "http":
		return newHTTPMCPClientWithConn(url)
	default:
		return nil, fmt.Errorf("unknown MCP transport: %s", config.MCP.Transport)
	}
}

// ConvertToolsToOllama 转换 MCP 工具定义到 AiProvider 工具格式
func (m *MCPClient) ConvertToolsToOllama() []map[string]any {
	var out []map[string]any
	for _, t := range m.Tools {
		var params map[string]any
		b, _ := json.Marshal(t.InputSchema)
		_ = json.Unmarshal(b, &params)

		out = append(out, map[string]any{
			"type": "function",
			"function": map[string]any{
				"name":        t.Name,
				"description": t.Description,
				"parameters":  params,
			},
		})
	}
	return out
}

// ConvertToolsToOpenAI 将 MCP 工具定义转换为 OpenAI Chat Completions 的 tools 参数
func (m *MCPClient) ConvertToolsToOpenAI() []openai.ChatCompletionToolUnionParam {
	out := make([]openai.ChatCompletionToolUnionParam, 0, len(m.Tools))
	for _, t := range m.Tools {
		var paramsMap map[string]any
		if b, _ := json.Marshal(t.InputSchema); len(b) != 0 {
			_ = json.Unmarshal(b, &paramsMap)
		}

		var fp openai.FunctionParameters
		if b, err := json.Marshal(paramsMap); err == nil {
			_ = json.Unmarshal(b, &fp)
		}
		fn := openai.FunctionDefinitionParam{
			Name:        t.Name,
			Description: param.Opt[string]{Value: t.Description},
			Parameters:  fp,
		}
		tool := &openai.ChatCompletionFunctionToolParam{
			Type:     "function",
			Function: fn,
		}
		out = append(out, openai.ChatCompletionToolUnionParam{
			OfFunction: tool,
		})
	}
	return out
}

// CallTool 调用 MCP 工具
func (m *MCPClient) CallTool(ctx context.Context, name string, args any) (string, error) {
	// 设置进度通知处理（这里应该用不上，是streamable HTTP的特性，太高级了）
	//m.Client.OnNotification(func(notification mcp.JSONRPCNotification) {
	//	logger.Infof("Received notification: %v", notification)
	//	if notification.Method == "notifications/progress" {
	//		fmt.Println(notification)
	//		params := notification.Params.AdditionalFields
	//		progress := params["progress"].(float64)
	//		total := params["total"].(float64)
	//		message := params["message"].(string)
	//
	//		// 打印进度信息
	//		fmt.Printf("Progress: %.2f%% - %s\n", (progress/total)*100, message)
	//	}
	//})

	res, err := m.Client.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      name,
			Arguments: args,
			Meta: &mcp.Meta{
				ProgressToken: time.Now().Unix(),
			},
		},
	})
	if err != nil {
		logger.Errorf("call tool %s: %v", name, err)
		return "", fmt.Errorf("call tool %s: %w", name, err)
	}

	// 提取返回文本
	var text string
	for _, c := range res.Content {
		if tc, ok := mcp.AsTextContent(c); ok {
			text += tc.Text + "\n"
		}
	}
	if text == "" && res.StructuredContent != nil {
		b, _ := json.Marshal(res.StructuredContent)
		text = string(b)
	}
	if text == "" {
		text = "(no content)"
	}
	return text, nil
}

// Close 关闭连接
func (m *MCPClient) Close() {
	if m.Client != nil {
		_ = m.Client.Close()
	}
}
