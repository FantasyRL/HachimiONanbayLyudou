package application

import (
	"context"
	"encoding/base64"

	"github.com/FantasyRL/go-mcp-demo/config"
	"github.com/FantasyRL/go-mcp-demo/pkg/base/ai_provider"
	"github.com/FantasyRL/go-mcp-demo/pkg/constant"
	"github.com/FantasyRL/go-mcp-demo/pkg/errno"
	"github.com/FantasyRL/go-mcp-demo/pkg/logger"
)

func (h *Host) Chat(id int64, msg string, imageData []byte) (string, error) {
	// 如果是远程模式（OpenAI），使用 ChatOpenAI
	if config.AiProvider.Mode == constant.AiProviderModeRemote {
		return h.ChatOpenAI(id, msg, imageData)
	}

	// 本地模式（Ollama）的原有逻辑
	// 获取当前用户的对话历史（如果没有则初始化为空切片）
	userHistory := history[id]
	if userHistory == nil {
		userHistory = []ai_provider.Message{}
	}

	// 构建用户消息
	userMsg := ai_provider.Message{Role: "user", Content: msg}

	// 如果有图片数据，转换为base64并添加到消息中
	if len(imageData) > 0 {
		base64Image := base64.StdEncoding.EncodeToString(imageData)
		userMsg.Images = []string{base64Image}
	}

	// 将当前用户消息加入历史
	userHistory = append(userHistory, userMsg)

	// 转换工具定义
	ollamaTools := h.mcpCli.ConvertToolsToOllama()
	ollamaOptions := ai_provider.BuildOptions()

	// 第一次调用模型（带历史）
	resp, err := h.aiProviderCli.Chat(h.ctx, ai_provider.ChatRequest{
		Model:     config.AiProvider.Model,
		Messages:  userHistory, // 使用完整历史
		Options:   ollamaOptions,
		Tools:     ollamaTools,
		KeepAlive: config.AiProvider.Options.KeepAlive,
	})
	if err != nil {
		return "", err
	}

	// 更新历史：添加模型回复
	userHistory = append(userHistory, ai_provider.Message{Role: "assistant", Content: resp.Message.Content})

	// 如果有工具调用
	if len(resp.Message.ToolCalls) > 0 {
		for _, c := range resp.Message.ToolCalls {
			args, err := ai_provider.ParseToolArguments(c.Function.Arguments)
			if err != nil {
				args = map[string]any{"_error": err.Error()}
			}

			out, err := h.mcpCli.CallTool(context.Background(), c.Function.Name, args)
			if err != nil {
				out = "tool error: " + err.Error()
			}

			// 添加工具执行结果到历史
			userHistory = append(userHistory, ai_provider.Message{
				Role:     "tool",
				ToolName: c.Function.Name,
				Content:  out,
			})
			logger.Infof("[tool] %s executed\n", c.Function.Name)
		}

		// 再次调用模型，传入完整历史（包含工具返回）
		resp2, err := h.aiProviderCli.Chat(h.ctx, ai_provider.ChatRequest{
			Model:    config.AiProvider.Model,
			Messages: userHistory, // 包含工具返回的新历史
			Options:  ollamaOptions,
			Tools:    ollamaTools,
		})
		if err != nil {
			return "", err
		}

		// 更新历史：添加最终模型回复
		userHistory = append(userHistory, ai_provider.Message{Role: "assistant", Content: resp2.Message.Content})

		// 保存回 map
		history[id] = userHistory

		return resp2.Message.Content, nil
	}

	// 无工具调用，直接返回模型回复
	history[id] = userHistory // 保存更新后的历史
	return resp.Message.Content, nil
}


func (h *Host) StreamChat(
	ctx context.Context,
	id int64,
	userMsg string,
	emit func(event string, v any) error, // SSE: event 名 + 任意 JSON 数据
) error {
	// 历史
	hist := history[id]
	if hist == nil {
		hist = []ai_provider.Message{}
	}
	// 加用户消息
	hist = append(hist, ai_provider.Message{Role: "user", Content: userMsg})

	tools := h.mcpCli.ConvertToolsToOllama()
	opts := ai_provider.BuildOptions()
	// 首次流式：边生成边推，遇到 tool_calls 停止
	var assistantBuf string
	var toolCalls []ai_provider.ToolCall

	err := h.aiProviderCli.ChatStream(ctx, ai_provider.ChatRequest{
		Model:     config.AiProvider.Model,
		Messages:  hist,
		Tools:     tools,
		Options:   opts,
		KeepAlive: config.AiProvider.Options.KeepAlive,
	}, func(chunk *ai_provider.ChatResponse) error {
		// 增量文本
		if s := chunk.Message.Content; s != "" {
			assistantBuf += s
			// 推送到handler层
			_ = emit(constant.SSEEventDelta, map[string]any{"text": s})
		}
		// 工具调用（可能在中途出现）
		if len(chunk.Message.ToolCalls) > 0 {
			toolCalls = append(toolCalls, chunk.Message.ToolCalls...)
			_ = emit(constant.SSEEventStartToolCall, map[string]any{"tool_calls": chunk.Message.ToolCalls})
			return errno.OllamaInternalStopStream // 提前结束首次流
		}
		return nil
	})
	if err != nil {
		return err
	}

	// 把模型已生成的片段先落历史
	if assistantBuf != "" {
		hist = append(hist, ai_provider.Message{Role: "assistant", Content: assistantBuf})
	}

	// 没有工具调用：直接完成
	if len(toolCalls) == 0 {
		history[id] = hist
		_ = emit("done", map[string]any{"reason": "no_tool"})
		return nil
	}

	// 执行工具
	for _, tc := range toolCalls {
		args, err := ai_provider.ParseToolArguments(tc.Function.Arguments)
		if err != nil {
			args = map[string]any{"_error": err.Error()}
		}

		_ = emit(constant.SSEEventToolCall, map[string]any{
			"name": tc.Function.Name,
			"args": args,
		})

		out, callErr := h.mcpCli.CallTool(ctx, tc.Function.Name, args)
		if callErr != nil {
			out = "tool error: " + callErr.Error()
		}

		// 工具结果给前端
		_ = emit(constant.SSEEventToolResult, map[string]any{
			"name":   tc.Function.Name,
			"result": out,
		})

		// 工具结果落历史
		hist = append(hist, ai_provider.Message{
			Role:     "tool",
			ToolName: tc.Function.Name,
			Content:  out,
		})
	}

	// 6) 二次流式：带工具结果，让模型给最终回答
	var finalBuf string
	err = h.aiProviderCli.ChatStream(ctx, ai_provider.ChatRequest{
		Model:     config.AiProvider.Model,
		Messages:  hist,
		Tools:     tools,
		Options:   opts,
		KeepAlive: config.AiProvider.Options.KeepAlive,
	}, func(chunk *ai_provider.ChatResponse) error {
		if s := chunk.Message.Content; s != "" {
			finalBuf += s
			_ = emit(constant.SSEEventDelta, map[string]any{"text": s})
		}
		return nil
	})
	if err != nil {
		return err
	}

	// 7) 结束收尾：保存历史、发 done
	if finalBuf != "" {
		hist = append(hist, ai_provider.Message{Role: "assistant", Content: finalBuf})
	}
	history[id] = hist
	_ = emit(constant.SSEEventDone, map[string]any{"reason": "completed"})
	return nil
}
