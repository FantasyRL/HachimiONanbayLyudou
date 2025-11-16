namespace go model
include "openapi.thrift"
struct BaseResp {
    1: i64 code (api.body="code", openapi.property='{
        title: "状态码",
        description: "响应状态码",
        type: "integer"
    }')
    2: string msg (api.body="msg", openapi.property='{
        title: "消息",
        description: "响应消息",
        type: "string"
    }')
}(
    openapi.schema='{
        title: "基础响应",
        description: "所有响应的基础结构",
        required: ["code", "msg"]
    }'
)

struct User {
    1: string id (api.body="id", openapi.property='{
        title: "用户ID",
        description: "唯一标识用户的ID",
        type: "string"
    }')
    2: string name (api.body="name", openapi.property='{
        title: "用户名",
        description: "用户的显示名称",
        type: "string"
    }')
}(
    openapi.schema='{
        title: "用户信息",
        description: "包含用户基本信息的结构",
        required: ["id", "name"]
    }'
)