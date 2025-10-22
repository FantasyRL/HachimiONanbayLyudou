package mcp_local

import (
	"github.com/FantasyRL/go-mcp-demo/internal/mcp_local/internal/ai_se_solver"
	"github.com/FantasyRL/go-mcp-demo/pkg/base/ai_provider"
)

func InjectDependencies() {
	// Inject dependencies here

	AIProviderClient := ai_provider.NewAiProviderClient()
	ai_se_solver.NewAISESolver(AIProviderClient)
}
