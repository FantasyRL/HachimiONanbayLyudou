package consul

import (
	"fmt"
	"github.com/FantasyRL/go-mcp-demo/pkg/logger"
	"strings"

	"github.com/FantasyRL/go-mcp-demo/config"
	"github.com/hashicorp/consul/api"
)

type Resolver struct {
	cfg *ConsulConfig
}

func NewResolver() *Resolver {
	cfg := &ConsulConfig{
		Address:    config.Registry.Consul.Address,
		Datacenter: config.Registry.Consul.Datacenter,
		Token:      config.Registry.Consul.Token,
		Tag:        config.Registry.Consul.Tag,
		Scheme:     config.Registry.Consul.Scheme,
		Path:       config.Registry.Consul.Path,
	}
	if cfg.Address == "" {
		return nil
	}
	return &Resolver{
		cfg: cfg,
	}
}

// Resolve 返回所有通过健康检查的实例 URL
func (r *Resolver) Resolve(services []string) (map[string][]string, error) {
	// 基本校验
	if len(services) == 0 {
		return nil, fmt.Errorf("no Services provided")
	}
	if r.cfg.Address == "" {
		return nil, fmt.Errorf("consul: empty address")
	}

	// 构造 Consul 客户端
	conf := api.DefaultConfig()
	conf.Address = r.cfg.Address
	if r.cfg.Datacenter != "" {
		conf.Datacenter = r.cfg.Datacenter
	}
	if r.cfg.Token != "" {
		conf.Token = r.cfg.Token
	}
	cl, err := api.NewClient(conf)
	if err != nil {
		return nil, fmt.Errorf("consul client: %w", err)
	}

	q := &api.QueryOptions{
		Datacenter: r.cfg.Datacenter,
		Token:      r.cfg.Token,
	}
	// 结果集合与去重
	out := make(map[string][]string)

	// 依次查询每个服务名
	for _, svc := range services {
		if svc == "" {
			continue
		}
		entries, _, err := cl.Health().Service(svc, r.cfg.Tag, true, q)
		if err != nil {
			return nil, fmt.Errorf("consul discover %q: %w", svc, err)
		}
		// 没有健康实例不视为整体错误，继续查下一个服务
		for _, inst := range entries {
			// 使用注册时写入的完整 addr
			if inst.Service != nil && inst.Service.Meta != nil {
				if u := strings.TrimSpace(inst.Service.Meta["addr"]); u != "" {
					out[svc] = append(out[svc], u)
				} else {
					logger.Errorf("consul: service %s no metadata", inst.Service.Service)
				}
			}
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("consul: no healthy instances for %v", services)
	}
	return out, nil
}
