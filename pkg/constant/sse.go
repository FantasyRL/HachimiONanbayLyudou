package constant

const (
	SSEEventDelta         = "delta"           // 模型内容增量
	SSEEventDone          = "done"            // 流结束事件
	SSEEventStartToolCall = "start_tool_call" // 开始工具调用
	SSEEventToolCall      = "tool_call"       // 工具调用
	SSEEventToolResult    = "tool_result"     // 工具调用结果
)
