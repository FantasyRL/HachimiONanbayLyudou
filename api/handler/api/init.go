package api

import (
	"github.com/FantasyRL/go-mcp-demo/pkg/base"
	"github.com/FantasyRL/go-mcp-demo/pkg/constant"
)

var clientSet *base.ClientSet

func Init() {
	clientSet = base.NewClientSet(
		base.WithMCPClient([]string{constant.ServiceNameMCPLocal, constant.ServiceNameMCPRemote}),
		base.WithAiProviderClient(),
		base.WithDB(),
	)
}
