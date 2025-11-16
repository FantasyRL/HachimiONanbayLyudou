package base

import (
	"github.com/FantasyRL/go-mcp-demo/pkg/base/ai_provider"
	"github.com/FantasyRL/go-mcp-demo/pkg/base/mcp_client"
	"github.com/FantasyRL/go-mcp-demo/pkg/base/registry"
	"gorm.io/gorm"
	"sync"
)

var (
	instance *ClientSet
	once     sync.Once
)

// ClientSet storage various client objects
// Notice: some or all of them maybe nil, we should check obj when use
type ClientSet struct {
	MCPCli           mcp_client.ToolClient
	AiProviderCli    *ai_provider.Client
	RegistryResolver registry.Resolver
	ActualDB         *gorm.DB
	cleanups         []func()
}

type Option func(clientSet *ClientSet)

// NewClientSet will be protected by sync.Once for ensure only 1 instance could be created in 1 lifecycle
func NewClientSet(opt ...Option) *ClientSet {
	once.Do(func() {
		var options []Option
		instance = &ClientSet{}
		options = append(options, opt...)
		for _, opt := range options {
			opt(instance)
		}
	})
	return instance
}
func (cs *ClientSet) Close() {
	for _, cleanup := range cs.cleanups {
		cleanup()
	}
}
