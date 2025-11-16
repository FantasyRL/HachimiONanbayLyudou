package application

import (
	"context"
	"github.com/FantasyRL/go-mcp-demo/pkg/base/tool_set"
	"github.com/mark3labs/mcp-go/mcp"
	"time"
)

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

//func WithLongRunningOperationTool() tool_set.Option {
//	return func(toolSet *tool_set.ToolSet) {
//		newTool := mcp.NewTool("long_running_tool",
//			mcp.WithDescription("A long running tool that reports progress"),
//			mcp.WithNumber("duration",
//				mcp.Description("Total duration of the operation in seconds"),
//				mcp.Required(),
//			),
//			mcp.WithNumber("steps",
//				mcp.Description("Number of steps to complete the operation"),
//				mcp.Required(),
//			),
//		)
//		// handleLongRunningOperationTool 示例长时间运行的工具，支持进度汇报
//		// https://github.com/mark3labs/mcp-go/blob/main/examples/everything/main.go 413
//		handleLongRunningOperationTool := func(
//			ctx context.Context,
//			request mcp.CallToolRequest,
//		) (*mcp.CallToolResult, error) {
//			// 从请求中提取工具参数
//			arguments := request.GetArguments()
//			// 从请求元数据中提取进度标识符
//			progressToken := request.Params.Meta.ProgressToken
//
//			// 获取任务总持续时间和步骤数
//			duration, _ := arguments["duration"].(float64) // 任务总持续时间（秒）
//			steps, _ := arguments["steps"].(float64)       // 任务步骤数
//
//			// 计算每一步的持续时间（每步的时长）
//			stepDuration := duration / steps
//
//			// 获取服务器上下文
//			server := server.ServerFromContext(ctx)
//
//			// 执行任务：模拟长时间操作，并在每一步发送进度通知
//			for i := 1; i < int(steps)+1; i++ {
//				// 每步执行完成后，等待相应的时间（模拟耗时操作）
//				time.Sleep(time.Duration(stepDuration * float64(time.Second)))
//
//				// 如果有进度令牌（progressToken），则发送进度通知
//				if progressToken != nil {
//					// 构造进度通知消息
//					err := server.SendNotificationToClient(
//						ctx,
//						"notifications/progress", // 通知类型
//						map[string]any{
//							"progress":      i,                                                              // 当前进度
//							"total":         int(steps),                                                     // 总步骤数
//							"progressToken": progressToken,                                                  // 进度令牌，标识该操作
//							"message":       fmt.Sprintf("Server progress %v%%", int(float64(i)*100/steps)), // 进度消息
//						},
//					)
//					// 错误处理：如果通知发送失败，返回错误
//					if err != nil {
//						logger.Errorf("Failed to send progress notification: %v", err)
//						return nil, fmt.Errorf("failed to send notification: %w", err)
//					}
//				}
//			}
//			time.Sleep(time.Second)
//
//			// 返回工具执行的最终结果（任务完成）
//			return &mcp.CallToolResult{
//				Content: []mcp.Content{
//					mcp.TextContent{
//						Type: "text", // 内容类型：文本
//						Text: fmt.Sprintf(
//							"Long running operation completed. Duration: %f seconds, Steps: %d.",
//							duration,   // 任务总持续时间
//							int(steps), // 总步骤数
//						),
//					},
//				},
//			}, nil
//		}
//
//		toolSet.Tools = append(toolSet.Tools, &newTool)
//		toolSet.HandlerFunc[newTool.Name] = handleLongRunningOperationTool
//	}
//}
