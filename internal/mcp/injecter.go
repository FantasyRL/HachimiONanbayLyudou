package mcp

import (
	"github.com/FantasyRL/go-mcp-demo/internal/mcp/application"
	"github.com/FantasyRL/go-mcp-demo/pkg/base/ai_provider"
)

func InjectDependencies() {
	// Inject dependencies here

	AIProviderClient := ai_provider.NewAiProviderClient()
	application.NewAISESolver(AIProviderClient)
}
