namespace go api
include "model.thrift"
include "openapi.thrift"

struct ChatRequest{
    1: string message(api.body="message", openapi.property='{
        title: "用户消息",
        description: "用户发送的消息内容",
        type: "string"
    }')
    2: optional binary image(api.form="image", api.file_name="image", openapi.property='{
        title: "图片文件",
        description: "可选的图片文件，支持上传图片给AI分析",
        type: "string",
        format: "binary"
    }')
    // 新增：是否启用联网搜索（仅在该会话首次消息生效）
    3: optional bool enableWebSearch(api.body="enableWebSearch", openapi.property='{
        title: "启用联网搜索",
        description: "是否允许在本次新对话中调用web.search工具",
        type: "boolean"
    }')
    // 新增：覆盖模型（仅新对话时生效）
    4: optional string model(api.body="model", openapi.property='{
        title: "模型覆盖",
        description: "仅新对话时可覆盖默认模型名称",
        type: "string"
    }')
    5: optional double temperature(api.body="temperature", openapi.property='{
        title: "采样温度",
        description: "仅新对话时覆盖温度",
        type: "number",
        format: "double"
    }')
    6: optional double top_p(api.body="top_p", openapi.property='{
        title: "Top-P",
        description: "仅新对话时覆盖核采样参数",
        type: "number",
        format: "double"
    }')
    7: optional i32 top_k(api.body="top_k", openapi.property='{
        title: "Top-K",
        description: "仅新对话时覆盖Top-K参数",
        type: "integer",
        format: "int32"
    }')
    8: optional i32 max_tokens(api.body="max_tokens", openapi.property='{
        title: "最大Token数",
        description: "仅新对话时限制最大生成长度",
        type: "integer",
        format: "int32"
    }')
}(
    openapi.schema='{
        title: "聊天请求",
        description: "包含用户消息的聊天请求，可选覆盖AI配置与启用联网搜索",
        required: ["message"]
    }'
)

struct ChatResponse{
    1: string response(api.body="response", openapi.property='{
        title: "AI回复",
        description: "AI生成的回复内容",
        type: "string"
    }')
}(
    openapi.schema='{
        title: "聊天响应",
        description: "包含AI回复的聊天响应",
        required: ["response"]
    }'
)

struct ChatSSEHandlerRequest{
    1: string message(api.query="message",openapi.property='{
        title: "用户消息",
        description: "用户发送的消息内容",
        type: "string"
    }')
    2: optional binary image(api.form="image", api.file_name="image", openapi.property='{
        title: "图片文件",
        description: "可选的图片文件，支持上传图片给AI分析",
        type: "file"
    }')
    3: optional bool enableWebSearch(api.query="enableWebSearch", openapi.property='{
        title: "启用联网搜索",
        description: "是否允许在本次新对话中调用web.search工具",
        type: "boolean"
    }')
    4: optional string model(api.query="model", openapi.property='{
        title: "模型覆盖",
        description: "仅新对话时可覆盖默认模型名称",
        type: "string"
    }')
    5: optional double temperature(api.query="temperature", openapi.property='{
        title: "采样温度",
        description: "仅新对话时覆盖温度",
        type: "number",
        format: "double"
    }')
    6: optional double top_p(api.query="top_p", openapi.property='{
        title: "Top-P",
        description: "仅新对话时覆盖核采样参数",
        type: "number",
        format: "double"
    }')
    7: optional i32 top_k(api.query="top_k", openapi.property='{
        title: "Top-K",
        description: "仅新对话时覆盖Top-K参数",
        type: "integer",
        format: "int32"
    }')
    8: optional i32 max_tokens(api.query="max_tokens", openapi.property='{
        title: "最大Token数",
        description: "仅新对话时限制最大生成长度",
        type: "integer",
        format: "int32"
    }')
}(
     openapi.schema='{
         title: "流式聊天请求",
         description: "包含用户消息的流式聊天请求，可选覆盖AI配置与启用联网搜索",
         required: ["message"]
     }'
)

struct ChatSSEHandlerResponse{
    1: string response(api.body="response", openapi.property='{
        title: "AI回复片段",
        description: "AI生成的回复片段",
        type: "string"
    }')
}(
    openapi.schema='{
        title: "流式聊天响应",
        description: "包含AI回复片段的流式聊天响应",
        required: ["response"]
    }'
)

struct TemplateRequest{
    1: string templateId(api.body="templateId", openapi.property='{
        title: "示范用param",
        description: "示范用param",
        type: "string"
    }')
}(
    openapi.schema='{
        title: "示例请求",
        description: "示例请求",
        required: ["templateId"]
    }'
)

struct TemplateResponse{
    1: model.User user(api.body="user", openapi.property='{
        title: "示范用返回值",
        description: "示范用返回值",
        type: "string"
    }')
}(
    openapi.schema='{
        title: "示例响应",
        description: "示例响应",
        required: ["user"]
    }'
)

service ApiService {
    // 非流式对话
    ChatResponse Chat(1: ChatRequest req)(api.post="/api/v1/chat")
    // 流式对话
    ChatSSEHandlerResponse ChatSSE(1: ChatSSEHandlerRequest req)(api.post="/api/v1/chat/sse")
    // 示例接口 idl写好后运行make hertz-gen-api生成脚手架
    TemplateResponse Template(1: TemplateRequest req)(api.post="/api/v1/template")
}