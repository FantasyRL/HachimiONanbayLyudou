package prompt_set

import (
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"sync"
)

var (
	instance *PromptSet
	once     sync.Once
)

type PromptSet struct {
	Prompts     []*mcp.Prompt
	HandlerFunc map[string]server.PromptHandlerFunc
}

type Option func(promptSet *PromptSet)

func NewPromptSet(opt ...Option) *PromptSet {
	once.Do(func() {
		var options []Option
		instance = new(PromptSet)
		instance.HandlerFunc = make(map[string]server.PromptHandlerFunc)
		options = append(options, opt...)
		for _, opt := range options {
			opt(instance)
		}
	})
	return instance
}
