package constant

import "time"

const (
	RegistryProviderConsul = "consul"
	RegistryProviderEtcd   = "etcd"
	RegistryProviderNacos  = "nacos"
	RegistryProviderNone   = "none"

	RegistryMCPTag         = "mcp"
	RegistryMCPDefaultPath = "/mcp"

	RegistryCheckInterval                  = 5 * time.Second
	RegistryDeregisterAfter                = 15 * time.Second
	RegistryResolverDefaultRefreshInterval = 10 * time.Second
)
