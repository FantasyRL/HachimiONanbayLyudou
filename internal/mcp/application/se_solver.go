package application

import (
	"context"

	"github.com/FantasyRL/go-mcp-demo/config"
	"github.com/FantasyRL/go-mcp-demo/pkg/base/ai_provider"
	"github.com/FantasyRL/go-mcp-demo/pkg/base/tool_set"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/openai/openai-go/v2"
)

var instance *AISESolver

type AISESolver struct {
	aiProviderCli *ai_provider.Client
}

func NewAISESolver(aiProviderCli *ai_provider.Client) *AISESolver {
	instance = &AISESolver{
		aiProviderCli: aiProviderCli,
	}
	return instance
}

func WithAIScienceAndEngineeringBuildHtmlTool() tool_set.Option {
	return func(toolSet *tool_set.ToolSet) {
		newTool := mcp.NewTool(
			"build_html_to_solve_science_and_engineering_problem",
			mcp.WithDescription("当用户遇到学习问题上的困难时，通过 <htmath> 特殊标签的使用规范来帮助用户理解数学概念和绘制图像, 可以在图像后面加上对问题的辅助解析"),
			mcp.WithString("question", mcp.Required(), mcp.Description("用户提出的科学或工程相关的问题")),
		)
		toolSet.Tools = append(toolSet.Tools, &newTool)
		toolSet.HandlerFunc[newTool.Name] = AIScienceAndEngineeringBuildHtml
	}
}

func AIScienceAndEngineeringBuildHtml(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	question := args["question"].(string)
	if question == "" {
		return mcp.NewToolResultError("missing required arg: question"), nil
	}
	resp, err := instance.aiProviderCli.ChatOpenAI(ctx, openai.ChatCompletionNewParams{
		Model: openai.ChatModel(config.AiProvider.Model),
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(systemPromptHTMLPrinter),
			openai.UserMessage(question),
		},
	})
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(resp.Choices[0].Message.Content), nil
}

const systemPromptHTMLPrinter = `
你是一名叫ssibal的html生成小助手：
1. 你只能使用<htmath>标签来展示内容，除此之外不输出其它。
`

// 新增工具集合构造，用于在注入阶段集中注册
func BuildToolSet() *tool_set.ToolSet {
	return tool_set.NewToolSet(
		WithAIScienceAndEngineeringBuildHtmlTool(),
		WithWebSearchTool(), // 注册 web.search
	)
}
