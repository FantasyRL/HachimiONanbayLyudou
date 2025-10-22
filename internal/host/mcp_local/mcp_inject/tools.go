package mcp_inject

import (
	"context"
	"fmt"
	"github.com/FantasyRL/go-mcp-demo/internal/mcp_local/internal/ai_se_solver"
	"github.com/FantasyRL/go-mcp-demo/internal/mcp_local/internal/dev_runner"
	"github.com/FantasyRL/go-mcp-demo/pkg/base/tool_set"
	"github.com/FantasyRL/go-mcp-demo/pkg/logger"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"time"
)

// WithDevRunnerTools 本地开发辅助工具
// 这组工具让 AI 能像本地助手一样：查看项目目录树(fs_tree)、读取文件(fs_cat)、运行项目/脚本(code_run)。
// - fs_tree：列出指定目录的树形结构（可控制深度/忽略模式），帮助 AI 感知项目布局。
// - fs_cat ：读取指定文件的内容（可限制最大字节），帮助 AI 查看未直接提供的代码。
// - code_run：在给定根目录下自动/按命令运行项目或单文件，返回 stdout/stderr/exit code，并给出基于错误输出的建议。
func WithDevRunnerTools() tool_set.Option {
	return func(toolSet *tool_set.ToolSet) {

		// fs_tree 目录树查看，让AI感知在哪个目录下运行代码
		toolTree := mcp.NewTool("fs_tree",
			mcp.WithDescription("List a directory as a plain text tree to understand project layout."),
			mcp.WithString("path", mcp.Required(), mcp.Description("Directory path to list")),
			// depth 最大遍历深度
			mcp.WithNumber("depth", mcp.Description("Max depth to traverse (default 4)")),
			// ignore 如 node_modules, *.log
			mcp.WithString("ignore", mcp.Description("Comma-separated glob patterns to ignore (optional)")),
		)
		toolSet.Tools = append(toolSet.Tools, &toolTree)
		toolSet.HandlerFunc[toolTree.Name] = dev_runner.HandleFsTree

		// fs_cat 读取文件里的内容
		toolCat := mcp.NewTool("fs_cat",
			mcp.WithDescription("Read a file content to inspect code that was not provided in the prompt."),
			// 文件路径
			mcp.WithString("path", mcp.Required(), mcp.Description("File path to read")),
			// 最大读取字节数
			mcp.WithNumber("max_bytes", mcp.Description("Max bytes to read (default 65536)")),
		)
		toolSet.Tools = append(toolSet.Tools, &toolCat)
		toolSet.HandlerFunc[toolCat.Name] = dev_runner.HandleFsCat

		// code_run 运行命令行
		toolRun := mcp.NewTool("code_run",
			// 工具用途：在本地命令行运行项目/脚本，返回 stdout/stderr/exit code，并基于错误输出给建议
			mcp.WithDescription("Run a code file/project locally in the given root directory with the EXACT command provided by the AI,return stdout"),
			// required ：工作目录（项目根目录）
			mcp.WithString("root", mcp.Required(), mcp.Description("Working directory of the project")),
			// 可选参数：运行前将 content 写入到 root 下的 file（相对路径）
			//mcp.WithString("file", mcp.Description("Optional file path relative to root to write/update before running")),
			//mcp.WithString("content", mcp.Description("Optional content to write to `file` before running")),
			// required ：显式运行命令
			mcp.WithString("command", mcp.Description("Explicit shell command to run under the root directory(eg `python main.py`,`go run cmd/host`,`npm run dev`)")),
			// optional ：超时（秒），默认 120s
			mcp.WithNumber("timeout_sec", mcp.Description("Timeout in seconds (default 120)")),
			// optional ：传给程序的标准输入
			mcp.WithString("stdin", mcp.Description("Optional STDIN to pass to the program")),
		)
		toolSet.Tools = append(toolSet.Tools, &toolRun)
		toolSet.HandlerFunc[toolRun.Name] = dev_runner.HandleCodeRun
	}
}

func WithAIScienceAndEngineeringBuildHtmlTool() tool_set.Option {
	return func(toolSet *tool_set.ToolSet) {
		newTool := mcp.NewTool(
			"build_html_to_solve_science_and_engineering_problem",
			mcp.WithDescription("当用户遇到学习问题上的困难时，通过 <htmath> 特殊标签的使用规范来帮助用户理解数学概念和绘制图像，你可以在图像的后面加上对问题的辅助解析"),
			mcp.WithString("question", mcp.Required(), mcp.Description("用户提出的科学或工程相关的问题")),
		)
		toolSet.Tools = append(toolSet.Tools, &newTool)
		toolSet.HandlerFunc[newTool.Name] = ai_se_solver.AIScienceAndEngineeringBuildHtml
	}
}

func WithTimeTool() tool_set.Option {
	return func(toolSet *tool_set.ToolSet) {
		newTool := mcp.NewTool(
			"time_now",
			mcp.WithDescription("返回当前时间（RFC3339）"),
		)
		toolFunc := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			now := time.Now().Format(time.RFC3339)
			return mcp.NewToolResultText(now), nil
		}
		toolSet.Tools = append(toolSet.Tools, &newTool)
		toolSet.HandlerFunc[newTool.Name] = toolFunc
	}
}

func WithLongRunningOperationTool() tool_set.Option {
	return func(toolSet *tool_set.ToolSet) {
		newTool := mcp.NewTool("long_running_tool",
			mcp.WithDescription("A long running tool that reports progress"),
			mcp.WithNumber("duration",
				mcp.Description("Total duration of the operation in seconds"),
				mcp.Required(),
			),
			mcp.WithNumber("steps",
				mcp.Description("Number of steps to complete the operation"),
				mcp.Required(),
			),
		)
		// handleLongRunningOperationTool 示例长时间运行的工具，支持进度汇报
		// https://github.com/mark3labs/mcp-go/blob/main/examples/everything/main.go 413
		handleLongRunningOperationTool := func(
			ctx context.Context,
			request mcp.CallToolRequest,
		) (*mcp.CallToolResult, error) {
			// 从请求中提取工具参数
			arguments := request.GetArguments()
			// 从请求元数据中提取进度标识符
			progressToken := request.Params.Meta.ProgressToken

			// 获取任务总持续时间和步骤数
			duration, _ := arguments["duration"].(float64) // 任务总持续时间（秒）
			steps, _ := arguments["steps"].(float64)       // 任务步骤数

			// 计算每一步的持续时间（每步的时长）
			stepDuration := duration / steps

			// 获取服务器上下文
			server := server.ServerFromContext(ctx)

			// 执行任务：模拟长时间操作，并在每一步发送进度通知
			for i := 1; i < int(steps)+1; i++ {
				// 每步执行完成后，等待相应的时间（模拟耗时操作）
				time.Sleep(time.Duration(stepDuration * float64(time.Second)))

				// 如果有进度令牌（progressToken），则发送进度通知
				if progressToken != nil {
					// 构造进度通知消息
					err := server.SendNotificationToClient(
						ctx,
						"notifications/progress", // 通知类型
						map[string]any{
							"progress":      i,                                                              // 当前进度
							"total":         int(steps),                                                     // 总步骤数
							"progressToken": progressToken,                                                  // 进度令牌，标识该操作
							"message":       fmt.Sprintf("Server progress %v%%", int(float64(i)*100/steps)), // 进度消息
						},
					)
					// 错误处理：如果通知发送失败，返回错误
					if err != nil {
						logger.Errorf("Failed to send progress notification: %v", err)
						return nil, fmt.Errorf("failed to send notification: %w", err)
					}
				}
			}
			time.Sleep(time.Second)

			// 返回工具执行的最终结果（任务完成）
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.TextContent{
						Type: "text", // 内容类型：文本
						Text: fmt.Sprintf(
							"Long running operation completed. Duration: %f seconds, Steps: %d.",
							duration,   // 任务总持续时间
							int(steps), // 总步骤数
						),
					},
				},
			}, nil
		}

		toolSet.Tools = append(toolSet.Tools, &newTool)
		toolSet.HandlerFunc[newTool.Name] = handleLongRunningOperationTool
	}
}
