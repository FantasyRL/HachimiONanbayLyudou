package tool_set

import (
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"sync"
)

var (
	instance *ToolSet
	once     sync.Once
)

type ToolSet struct {
	// tools
	Tools []*mcp.Tool
	// map[t.Name]HandlerFunc
	HandlerFunc map[string]server.ToolHandlerFunc
}

// Option 定义了一个参数为toolSet的函数，具体实现为在函数内对toolSet进行append
type Option func(toolSet *ToolSet)

func NewToolSet(opt ...Option) *ToolSet {
	once.Do(func() {
		var options []Option
		instance = new(ToolSet)
		instance.HandlerFunc = make(map[string]server.ToolHandlerFunc)
		options = append(options, opt...)
		for _, opt := range options {
			opt(instance)
		}
	})
	return instance
}
