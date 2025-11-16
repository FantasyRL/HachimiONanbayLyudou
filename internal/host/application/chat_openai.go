package application

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"github.com/FantasyRL/go-mcp-demo/pkg/logger"

	"github.com/FantasyRL/go-mcp-demo/config"
	"github.com/FantasyRL/go-mcp-demo/pkg/constant"
	"github.com/FantasyRL/go-mcp-demo/pkg/errno"
	openai "github.com/openai/openai-go/v2"
)

// 将 OpenAI 的 tool_calls[].function.arguments (string) 解成 map[string]any（与原逻辑一致）
func parseOpenAIToolArgs(argStr string) map[string]any {
	if argStr == "" {
		return map[string]any{}
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(argStr), &m); err == nil {
		return m
	}
	// 如果不是 JSON，就当成纯字符串包裹
	return map[string]any{"_": argStr}
}

const maxToolRounds = 10 // 防御性上限，避免死循环

func (h *Host) StreamChatOpenAI(
	ctx context.Context,
	id int64,
	userMsg string,
	imageData []byte,
	emit func(event string, v any) error, // SSE: event 名 + 任意 JSON 数据
) error {
	// 历史（OpenAI）
	hist := historyOpenAI[id]
	if hist == nil {
		hist = []openai.ChatCompletionMessageParamUnion{}
	}

	// 构建用户消息
	if len(imageData) > 0 {
		// 如果有图片，使用多模态消息格式
		base64Image := base64.StdEncoding.EncodeToString(imageData)

		// 创建文本部分
		textPart := openai.ChatCompletionContentPartUnionParam{
			OfText: &openai.ChatCompletionContentPartTextParam{
				Type: "text",
				Text: userMsg,
			},
		}

		// 创建图片部分
		imagePart := openai.ImageContentPart(openai.ChatCompletionContentPartImageImageURLParam{
			URL: "data:image/jpeg;base64," + base64Image,
		})

		// 创建用户消息
		hist = append(hist, openai.ChatCompletionMessageParamUnion{
			OfUser: &openai.ChatCompletionUserMessageParam{
				Role: "user",
				Content: openai.ChatCompletionUserMessageParamContentUnion{
					OfArrayOfContentParts: []openai.ChatCompletionContentPartUnionParam{
						textPart,
						imagePart,
					},
				},
			},
		})
	} else {
		// 纯文本消息
		hist = append(hist, openai.UserMessage(userMsg))
	}

	// 工具（OpenAI 版）
	tools := h.mcpCli.ConvertToolsToOpenAI()

	round := 0
	for {
		round++
		if round > maxToolRounds {
			historyOpenAI[id] = hist
			_ = emit(constant.SSEEventDone, map[string]any{"reason": "tool_round_limit"})
			return nil
		}

		// 一轮生成：边流边推，若需要工具则中断本轮
		var assistantBuf string
		var acc openai.ChatCompletionAccumulator
		var needTools bool

		err := h.aiProviderCli.ChatStreamOpenAI(ctx, openai.ChatCompletionNewParams{
			Model:    openai.ChatModel(config.AiProvider.Model),
			Messages: hist,
			Tools:    tools,
		}, func(chunk *openai.ChatCompletionChunk) error {
			acc.AddChunk(*chunk)
			if len(chunk.Choices) > 0 {
				if s := chunk.Choices[0].Delta.Content; s != "" {
					assistantBuf += s
					_ = emit(constant.SSEEventDelta, map[string]any{"text": s})
				}
				// 工具调用结束标志（OpenAI：最后一帧 finish_reason = "tool_calls"）
				if chunk.Choices[0].FinishReason == "tool_calls" {
					needTools = true
					if len(acc.Choices) > 0 && len(acc.Choices[0].Message.ToolCalls) > 0 {
						_ = emit(constant.SSEEventStartToolCall, map[string]any{
							"tool_calls": acc.Choices[0].Message.ToolCalls,
							"round":      round,
						})
					}
					return errno.OllamaInternalStopStream
				}
			}
			return nil
		})
		if err != nil {
			return err
		}

		// 把已产生的 assistant 文本落历史
		if assistantBuf != "" && !needTools {
			hist = append(hist, openai.AssistantMessage(assistantBuf))
		}

		// 如果本轮不需要工具，说明模型已经给出最终答案
		if !needTools {
			historyOpenAI[id] = hist
			_ = emit(constant.SSEEventDone, map[string]any{"reason": "completed"})
			return nil
		}

		// 执行（可能多个）工具调用，然后将每个工具结果以 ToolMessage 落历史
		if len(acc.Choices) == 0 || len(acc.Choices[0].Message.ToolCalls) == 0 {
			// 偶发兜底：标记需要工具但没聚合到（理论上不会发生）
			historyOpenAI[id] = hist
			_ = emit(constant.SSEEventDone, map[string]any{"reason": "no_tool_details"})
			return nil
		}

		toolCallsParam := make([]openai.ChatCompletionMessageToolCallUnionParam, 0, len(acc.Choices[0].Message.ToolCalls))
		for _, tc := range acc.Choices[0].Message.ToolCalls {
			toolCallsParam = append(toolCallsParam, openai.ChatCompletionMessageToolCallUnionParam{
				OfFunction: &openai.ChatCompletionMessageFunctionToolCallParam{
					ID:   tc.ID,
					Type: "function",
					Function: openai.ChatCompletionMessageFunctionToolCallFunctionParam{
						Name:      tc.Function.Name,
						Arguments: tc.Function.Arguments, // 注意：这里是字符串
					},
				},
			})
		}
		// 根据openAI规范，tool_call前需要一条assistantMsg
		assistantWithCalls := openai.ChatCompletionAssistantMessageParam{
			Role:      "assistant",
			ToolCalls: toolCallsParam,
		}
		hist = append(hist, openai.ChatCompletionMessageParamUnion{OfAssistant: &assistantWithCalls})

		for _, tc := range acc.Choices[0].Message.ToolCalls {
			name := tc.Function.Name

			// OpenAI 的 arguments 是字符串，需要解成 map[string]any
			var args map[string]any
			if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
				args = map[string]any{"_parse_error": err.Error(), "_raw": tc.Function.Arguments}
			}

			_ = emit(constant.SSEEventToolCall, map[string]any{
				"round": round,
				"name":  name,
				"args":  args,
			})

			out, callErr := h.mcpCli.CallTool(ctx, name, args)
			if callErr != nil {
				out = "tool error: " + callErr.Error()
			}

			_ = emit(constant.SSEEventToolResult, map[string]any{
				"round":  round,
				"name":   name,
				"result": out,
			})

			// 工具结果回模型（重要）：OpenAI 规范用 ToolMessage，必须带 tool_call_id
			hist = append(hist, openai.ToolMessage(out, tc.ID))
			//logger.Infof("[tool round %d] %s executed", round, name)
		}

		// 循环进入下一轮：模型会在新的上下文（含工具结果）上继续生成
	}
}

// ChatOpenAI 非流式OpenAI聊天，支持图片和工具调用
func (h *Host) ChatOpenAI(id int64, msg string, imageData []byte) (string, error) {
	// 历史（OpenAI）
	hist := historyOpenAI[id]
	if hist == nil {
		hist = []openai.ChatCompletionMessageParamUnion{}
	}

	// 构建用户消息
	if len(imageData) > 0 {
		// 如果有图片，使用多模态消息格式
		base64Image := base64.StdEncoding.EncodeToString(imageData)

		// 创建文本部分
		textPart := openai.ChatCompletionContentPartUnionParam{
			OfText: &openai.ChatCompletionContentPartTextParam{
				Type: "text",
				Text: msg,
			},
		}

		// 创建图片部分
		imagePart := openai.ImageContentPart(openai.ChatCompletionContentPartImageImageURLParam{
			URL: "data:image/jpeg;base64," + base64Image,
		})

		// 创建用户消息
		hist = append(hist, openai.ChatCompletionMessageParamUnion{
			OfUser: &openai.ChatCompletionUserMessageParam{
				Role: "user",
				Content: openai.ChatCompletionUserMessageParamContentUnion{
					OfArrayOfContentParts: []openai.ChatCompletionContentPartUnionParam{
						textPart,
						imagePart,
					},
				},
			},
		})
	} else {
		// 纯文本消息
		hist = append(hist, openai.UserMessage(msg))
	}

	// 工具（OpenAI 版）- 如果有图片则不使用工具（vision模型可能不支持）
	var tools []openai.ChatCompletionToolUnionParam
	if len(imageData) == 0 {
		tools = h.mcpCli.ConvertToolsToOpenAI()
	}

	round := 0
	for {
		round++
		if round > maxToolRounds {
			historyOpenAI[id] = hist
			return "已达到工具调用轮次上限", nil
		}

		// 调用OpenAI API
		params := openai.ChatCompletionNewParams{
			Model:    openai.ChatModel(config.AiProvider.Model),
			Messages: hist,
			Tools:    tools,
		}
		if config.AiProvider.Options.MaxTokens != nil {
			params.MaxTokens = openai.Int(int64(*config.AiProvider.Options.MaxTokens))
		}
		if config.AiProvider.Options.Temperature != nil {
			params.Temperature = openai.Float(*config.AiProvider.Options.Temperature)
		}
		if config.AiProvider.Options.TopP != nil {
			params.TopP = openai.Float(*config.AiProvider.Options.TopP)
		}

		resp, err := h.aiProviderCli.ChatOpenAI(h.ctx, params)
		if err != nil {
			logger.Errorf("ChatOpenAI API error: %v", err)
			return "", err
		}

		logger.Infof("ChatOpenAI response: choices=%d", len(resp.Choices))
		if len(resp.Choices) == 0 {
			logger.Errorf("ChatOpenAI: no choices in response")
			return "模型返回为空", nil
		}

		logger.Infof("ChatOpenAI finish_reason=%s, content=%s, tool_calls=%d",
			resp.Choices[0].FinishReason, resp.Choices[0].Message.Content, len(resp.Choices[0].Message.ToolCalls))

		// 检查是否需要工具调用
		if resp.Choices[0].FinishReason != "tool_calls" || len(resp.Choices[0].Message.ToolCalls) == 0 {
			// 无工具调用，返回模型回复
			content := resp.Choices[0].Message.Content
			hist = append(hist, openai.AssistantMessage(content))
			historyOpenAI[id] = hist
			return content, nil
		}

		// 有工具调用，处理工具
		toolCallsParam := make([]openai.ChatCompletionMessageToolCallUnionParam, 0, len(resp.Choices[0].Message.ToolCalls))
		for _, tc := range resp.Choices[0].Message.ToolCalls {
			toolCallsParam = append(toolCallsParam, openai.ChatCompletionMessageToolCallUnionParam{
				OfFunction: &openai.ChatCompletionMessageFunctionToolCallParam{
					ID:   tc.ID,
					Type: "function",
					Function: openai.ChatCompletionMessageFunctionToolCallFunctionParam{
						Name:      tc.Function.Name,
						Arguments: tc.Function.Arguments,
					},
				},
			})
		}

		// 添加assistant消息（带工具调用）
		assistantWithCalls := openai.ChatCompletionAssistantMessageParam{
			Role:      "assistant",
			ToolCalls: toolCallsParam,
		}
		hist = append(hist, openai.ChatCompletionMessageParamUnion{OfAssistant: &assistantWithCalls})

		// 执行所有工具调用
		for _, tc := range resp.Choices[0].Message.ToolCalls {
			name := tc.Function.Name

			// OpenAI 的 arguments 是字符串，需要解成 map[string]any
			var args map[string]any
			if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
				args = map[string]any{"_parse_error": err.Error(), "_raw": tc.Function.Arguments}
			}

			out, callErr := h.mcpCli.CallTool(h.ctx, name, args)
			if callErr != nil {
				out = "tool error: " + callErr.Error()
			}

			// 工具结果回模型（重要）：OpenAI 规范用 ToolMessage，必须带 tool_call_id
			hist = append(hist, openai.ToolMessage(out, tc.ID))
		}

		// 循环进入下一轮：模型会在新的上下文（含工具结果）上继续生成
	}
}
