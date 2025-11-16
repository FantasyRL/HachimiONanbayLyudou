package ai_provider

import (
	"encoding/json"
	"github.com/FantasyRL/go-mcp-demo/config"
)

// ToolFunction 工具调用函数
type ToolFunction struct {
	Name      string          `json:"name"`      // 工具名
	Arguments json.RawMessage `json:"arguments"` // 这里可能是JSON，也可能是字符串里套JSON
}

// ToolCall 工具调用
type ToolCall struct {
	ID       string       `json:"id,omitempty"`
	Type     string       `json:"type,omitempty"` // 工具调用，比如类型"function"
	Function ToolFunction `json:"function"`       // 具体要调用哪个函数
}

// Message 对话消息
type Message struct {
	Role      string     `json:"role"`                 // "system"(初始化AI风格) | "user" | "assistant" | "tool"(工具结果回填给模型时使用)
	Content   string     `json:"content,omitempty"`    // 对话文本 | 工具结果
	Images    []string   `json:"images,omitempty"`     // 多模态... 后面再说
	ToolCalls []ToolCall `json:"tool_calls,omitempty"` // 模型(assistant)需要调用的工具列表
	ToolName  string     `json:"tool_name,omitempty"`  // 回填工具执行结果时带上,对应 ToolCall.Function.Name,声明这是哪个工具的结果
}

type ChatRequest struct {
	Model     string           `json:"model"`                // AiProvider 模型名 如 qwen3:4b
	Messages  []Message        `json:"messages"`             // 对话历史
	Tools     []map[string]any `json:"tools,omitempty"`      // 声明给模型能用的工具
	Options   map[string]any   `json:"options,omitempty"`    // 温度等，这些AI模型参数我不熟
	Stream    bool             `json:"stream"`               // 是否流式返回
	KeepAlive string           `json:"keep_alive,omitempty"` // 让模型权重留在内存的时长，这个我也不知道
	Format    any              `json:"format,omitempty"`     // 需要结构化输出时，可以给一个 JSON Schema；模型会尽量按这个结构生成
}

type ChatResponse struct {
	Model         string  `json:"model"`          // AiProvider 模型名 如 qwen3:4b
	CreatedAt     string  `json:"created_at"`     // 响应时间
	Message       Message `json:"message"`        // toolCalls不为空时则去执行工具
	Done          bool    `json:"done"`           // 非流式时总是true，流式时表示是否结束
	TotalDuration int64   `json:"total_duration"` // 整体耗时
}

// ParseToolArguments 解析ToolFunction
// 1) arguments 是 JSON 对象（e.g. {} / {"k":"v"}）
// 2) arguments 是 JSON 字符串，里面包着对象（e.g. "{\"k\":\"v\"}"）
func ParseToolArguments(raw json.RawMessage) (any, error) {
	if len(raw) == 0 {
		return map[string]any{}, nil
	}

	// 先尝试按对象/任意类型直接解
	var v any
	if err := json.Unmarshal(raw, &v); err == nil {
		// 如果是字符串，再尝试把字符串内容当 JSON 再解一层
		if s, ok := v.(string); ok {
			var inner any
			if err2 := json.Unmarshal([]byte(s), &inner); err2 == nil {
				return inner, nil
			}
			// 不是 JSON，就当成普通字符串参数
			return s, nil
		}
		return v, nil
	}

	// 兜底：按字符串解
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		var inner any
		if err2 := json.Unmarshal([]byte(s), &inner); err2 == nil {
			return inner, nil
		}
		return s, nil
	}

	// 再兜底：原样返回
	return string(raw), nil
}

func BuildOptions() map[string]any {
	opt := map[string]any{}
	if config.AiProvider.Options.Temperature != nil {
		opt["temperature"] = *config.AiProvider.Options.Temperature
	}
	if config.AiProvider.Options.TopP != nil {
		opt["top_p"] = *config.AiProvider.Options.TopP
	}
	if config.AiProvider.Options.TopK != nil {
		opt["top_k"] = *config.AiProvider.Options.TopK
	}
	for k, v := range config.AiProvider.Options.Extra {
		opt[k] = v
	}
	return opt
}
