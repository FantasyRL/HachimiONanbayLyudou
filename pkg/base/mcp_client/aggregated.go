package mcp_client

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/FantasyRL/go-mcp-demo/config"
	"github.com/FantasyRL/go-mcp-demo/pkg/constant"
	"go.uber.org/zap"
	"sort"
	"sync"
	"time"

	"github.com/FantasyRL/go-mcp-demo/pkg/base/registry"
	"github.com/FantasyRL/go-mcp-demo/pkg/logger"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/packages/param"
)

// AggregatedClient 面向 Host 的多路聚合客户端，实现 ToolClient 接口
type AggregatedClient struct {
	resolver        registry.Resolver
	refreshInterval time.Duration

	mu               sync.RWMutex
	discoverServices []string
	clients          map[string]*MCPClient // url -> client
	toolIndex        map[string]string     // toolName -> url
	toolSnapshot     map[string]mcp.Tool   // 聚合后的 tool 定义

	stopCh   chan struct{}
	stopOnce sync.Once
}

func NewAggregatedClient(resolver registry.Resolver, services []string) *AggregatedClient {
	if config.Registry.RefreshInterval <= constant.RegistryResolverDefaultRefreshInterval {
		config.Registry.RefreshInterval = constant.RegistryResolverDefaultRefreshInterval
	}
	ac := &AggregatedClient{
		resolver:         resolver,
		discoverServices: services,
		refreshInterval:  config.Registry.RefreshInterval,
		clients:          make(map[string]*MCPClient),
		toolIndex:        make(map[string]string),
		toolSnapshot:     make(map[string]mcp.Tool),
		stopCh:           make(chan struct{}),
	}
	// 启动定时刷新goroutine
	go ac.updateAggregatedClient()
	return ac
}

func (a *AggregatedClient) updateAggregatedClient() {
	a.refresh()
	ticker := time.NewTicker(a.refreshInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			a.refresh()
		case <-a.stopCh:
			return
		}
	}
}

// refresh 刷新registryCli与注册中心的连接，来更新可用的MCP服务实例列表
func (a *AggregatedClient) refresh() {
	if a.resolver == nil {
		return
	}

	// 服务发现
	serviceToUrls, err := a.resolver.Resolve(a.discoverServices)
	if err != nil {
		logger.Warn("registry resolve:", zap.Error(err))
		return
	}
	// 转化为set
	target := make(map[string]struct{})
	for _, urls := range serviceToUrls {
		for _, url := range urls {
			target[url] = struct{}{}
		}
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	// 删除已关闭的连接
	for u, cli := range a.clients {
		if _, ok := target[u]; ok {
			continue
		}
		if u == "fzuhelper-mcp" {
			continue
		}
		cli.Close()
		delete(a.clients, u)
		logger.Info("mcp disconnected: ", zap.String("url", u))
	}
	// 新增连接
	for u := range target {
		if _, ok := a.clients[u]; ok {
			continue
		}
		cli, err := NewMCPClient("http://" + u + "/mcp")
		if err != nil {
			logger.Errorf("mcp dial %s: %v", u, err)
			continue
		}
		a.clients[u] = cli
		logger.Infof("mcp connected: %s (tools=%d)", u, len(cli.Tools))
	}
	// 建立fzu-helper-mcp连接
	fzuCli, err := NewMCPClient(constant.FzuHelperServerMCPUrl)
	if err != nil {
		logger.Errorf("mcp dial %s: %v", constant.FzuHelperServerMCPUrl, err)
	}
	a.clients["fzuhelper-mcp"] = fzuCli
	logger.Infof("fzu-mcp connected: %s (tools=%d)", constant.FzuHelperServerMCPUrl, len(fzuCli.Tools))

	a.rebuildIndex()
}

// rebuildIndex 重建MCPClient映射
func (a *AggregatedClient) rebuildIndex() {
	// tool -> 候选 url 列表
	candidates := map[string][]string{}
	toolDef := map[string]mcp.Tool{}
	for url, cli := range a.clients {
		if cli == nil {
			logger.Warn("mcp unexpect disconnection: ", zap.String("url", url))
			continue
		}
		for _, t := range cli.Tools {
			candidates[t.Name] = append(candidates[t.Name], url)
			// 记录一个定义（相同名称一般结构一致）
			if _, ok := toolDef[t.Name]; !ok {
				toolDef[t.Name] = t
			}
		}
	}
	index := make(map[string]string, len(candidates))
	for name, urls := range candidates {
		//sort.Slice(urls, func(i, j int) bool {
		//	si := scoreURL(urls[i])
		//	sj := scoreURL(urls[j])
		//	if si != sj {
		//		return si > sj
		//	}
		//	return urls[i] < urls[j]
		//})
		index[name] = urls[0]
	}
	a.toolIndex = index
	a.toolSnapshot = toolDef
}

// scoreURL URL权重，不是很必要
func scoreURL(u string) int {
	return 0
}

func (a *AggregatedClient) ConvertToolsToOllama() []map[string]any {
	a.mu.RLock()
	defer a.mu.RUnlock()
	out := make([]map[string]any, 0, len(a.toolSnapshot))
	for _, t := range a.toolSnapshot {
		var params map[string]any
		if b, _ := json.Marshal(t.InputSchema); len(b) != 0 {
			_ = json.Unmarshal(b, &params)
		}
		out = append(out, map[string]any{
			"type": "function",
			"function": map[string]any{
				"name":        t.Name,
				"description": t.Description,
				"parameters":  params,
			},
		})
	}
	return out
}

func (a *AggregatedClient) ConvertToolsToOpenAI() []openai.ChatCompletionToolUnionParam {
	a.mu.RLock()
	defer a.mu.RUnlock()
	out := make([]openai.ChatCompletionToolUnionParam, 0, len(a.toolSnapshot))
	// 稳定顺序（按名称）
	names := make([]string, 0, len(a.toolSnapshot))
	for name := range a.toolSnapshot {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		t := a.toolSnapshot[name]
		var paramsMap map[string]any
		if b, _ := json.Marshal(t.InputSchema); len(b) != 0 {
			_ = json.Unmarshal(b, &paramsMap)
		}
		var fp openai.FunctionParameters
		if b, err := json.Marshal(paramsMap); err == nil {
			_ = json.Unmarshal(b, &fp)
		}
		fn := openai.FunctionDefinitionParam{
			Name:        t.Name,
			Description: param.Opt[string]{Value: t.Description},
			Parameters:  fp,
		}
		tool := &openai.ChatCompletionFunctionToolParam{
			Type:     "function",
			Function: fn,
		}
		out = append(out, openai.ChatCompletionToolUnionParam{OfFunction: tool})
	}
	return out
}

func (a *AggregatedClient) CallTool(ctx context.Context, name string, args any) (string, error) {
	a.mu.RLock()
	url, ok := a.toolIndex[name]
	cli := a.clients[url]
	a.mu.RUnlock()
	if !ok || cli == nil {
		return "", fmt.Errorf("tool %q not found (no connected MCP server provides it)", name)
	}
	return cli.CallTool(ctx, name, args)
}

func (a *AggregatedClient) Close() {
	a.stopOnce.Do(func() { close(a.stopCh) })
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, cli := range a.clients {
		cli.Close()
	}
}
