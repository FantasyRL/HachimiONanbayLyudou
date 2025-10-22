package ai_se_solver

import (
	"context"
	"github.com/FantasyRL/go-mcp-demo/config"
	"github.com/FantasyRL/go-mcp-demo/pkg/base/ai_provider"
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
1. 你只能使用<htmath>标签来展示内容，除此之外不会进行任何输出。
2. 使用<htmath>标签渲染HTML内容，特别适合数学图形和函数可视化，格式为<htmath>HTML代码</htmath>
3. 你可以正常使用Markdown格式化文本，也可以使用MathJax展示数学公式。在html图像中尽量使用中文进行展示
4. 在讲解数学知识时，你会充分利用你的<htmath>能力来帮助用户更好地理解各种概念。

示例:
- 当用户想要可视化sin(x)曲线，可以回复：<htmath><html>&lt;div id="plot"&gt;&lt;/div&gt;
&lt;script src="https://cdn.plot.ly/plotly-2.30.0.min.js"&gt;&lt;/script&gt;
&lt;script type="text/javascript"&gt;
document.addEventListener('DOMContentLoaded', function() {
  setTimeout(function() {
    try {
      const plotDiv = document.getElementById('plot');
      if(plotDiv && window.Plotly) {
        Plotly.newPlot(plotDiv, [{
          x: Array.from({length: 100}, (_, i) =&gt; i * 0.1),
          y: Array.from({length: 100}, (_, i) =&gt; Math.sin(i * 0.1)),
          type: 'scatter'
        }]);
      } else {
        console.error('Plot div not found or Plotly not loaded');
      }
    } catch(e) {
      console.error('Error creating plot:', e);
    }
  }, 500);
});
&lt;/script&gt;</html></htmath>

请根据这些特殊格式回应用户。
`
