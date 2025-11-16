package application

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/FantasyRL/go-mcp-demo/pkg/base/tool_set"
	"github.com/mark3labs/mcp-go/mcp"
)

// DuckDuckGo Instant Answer API 结果部分结构
type ddgInstant struct {
	Heading       string `json:"Heading"`
	AbstractURL   string `json:"AbstractURL"`
	AbstractText  string `json:"AbstractText"`
	RelatedTopics []struct {
		Text     string `json:"Text"`
		FirstURL string `json:"FirstURL"`
	} `json:"RelatedTopics"`
}

// Option：注册 web.search 工具
func WithWebSearchTool() tool_set.Option {
	return func(ts *tool_set.ToolSet) {
		tool := mcp.NewTool(
			"web.search",
			mcp.WithDescription("Use DuckDuckGo Instant Answer API to fetch public information. Required arg: query"),
			mcp.WithString("query", mcp.Required(), mcp.Description("Search query keywords")),
		)
		ts.Tools = append(ts.Tools, &tool)
		ts.HandlerFunc[tool.Name] = WebSearchHandler
	}
}

// 处理函数
func WebSearchHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	rawQuery, ok := args["query"]
	if !ok {
		return mcp.NewToolResultError("missing required arg: query"), nil
	}
	query, _ := rawQuery.(string)
	query = strings.TrimSpace(query)
	if query == "" {
		return mcp.NewToolResultError("query is empty"), nil
	}

	// 请求 DuckDuckGo
	api := "https://api.duckduckgo.com/"
	params := url.Values{}
	params.Set("q", query)
	params.Set("format", "json")
	params.Set("no_html", "1")
	params.Set("skip_disambig", "1")

	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Get(api + "?" + params.Encode())
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("request error: %v", err)), nil
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var data ddgInstant
	if err := json.Unmarshal(body, &data); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("decode error: %v", err)), nil
	}

	type item struct {
		Title   string `json:"title"`
		URL     string `json:"url"`
		Snippet string `json:"snippet"`
	}

	results := make([]item, 0, 5)
	for _, rt := range data.RelatedTopics {
		if len(results) >= 5 {
			break
		}
		results = append(results, item{
			Title:   rt.Text,
			URL:     rt.FirstURL,
			Snippet: rt.Text,
		})
	}
	if len(results) == 0 && (data.AbstractURL != "" || data.AbstractText != "") {
		results = append(results, item{
			Title:   data.Heading,
			URL:     data.AbstractURL,
			Snippet: data.AbstractText,
		})
	}
	if len(results) == 0 {
		return mcp.NewToolResultText("no results"), nil
	}

	out, _ := json.Marshal(map[string]any{
		"query":   query,
		"results": results,
	})
	return mcp.NewToolResultText(string(out)), nil
}
