# go-mcp-demo
a demo to learn how to use mcp in go

# 项目结构
- 直接将hertz HTTP server用于host接收外部请求
- mcp-host与mcp-server通过[**streamableHTTP**通信](https://www.51cto.com/article/826884.html)
- mcp-host与ollama通过http通信

# quick start
- copy `config.example.yaml` to `config.yaml` (`config.stdio.yaml`同理)
- windows需要安装`makefile`相关工具
## stdio
```bash
make stdio # windows需要修改config.stdio.yaml中的mcp.stdio.server_cmd 为./bin/mcp-server.exe
```

## http单点通信
- registry.provider为none
- mcp.transport为http
- mcp.http.base_url为mcp_server的url
### 本地启动
```bash
make mcp_local
```
```bash
make host #在另一个终端运行
```
### docker启动(network=host模式)
```bash
make docker-run-mcp_local
make docker-run-host
```

## 基于consul集群启动
### 本地启动
```bash
make env
make host
make mcp_local#在另一个终端运行
make mcp_remote#在另一个终端运行
```
### docker启动(network=host模式)
```bash
make env
make docker-run-host
make docker-run-mcp_local
make docker-run-mcp_remote
```

通过使用API管理平台(apifox/postman等)导入swagger/openapi.yaml，配置环境为host url，访问接口进行对话

记忆临时通过map来保存在内存中，重启host会丢失