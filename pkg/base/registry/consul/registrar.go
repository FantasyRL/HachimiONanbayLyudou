package consul

import (
	"fmt"
	"github.com/FantasyRL/go-mcp-demo/config"
	"github.com/FantasyRL/go-mcp-demo/pkg/base/registry"
	"github.com/FantasyRL/go-mcp-demo/pkg/constant"
	"github.com/hashicorp/consul/api"
)

type Registrar struct {
	cfg *ConsulConfig
}

func NewRegistrar(serviceName string) *Registrar {
	cfg := &ConsulConfig{
		Address:    config.Registry.Consul.Address,
		Datacenter: config.Registry.Consul.Datacenter,
		Token:      config.Registry.Consul.Token,
		Tag:        config.Registry.Consul.Tag,
		Service:    serviceName,
		Scheme:     config.Registry.Consul.Scheme,
		Path:       config.Registry.Consul.Path,
	}
	if cfg.Address == "" {
		return nil
	}
	return &Registrar{cfg: cfg}
}

// Register 注册服务实例，用于mcp_server将自己注册到consul
func (r *Registrar) Register(reg *registry.Registration) (func() error, error) {
	// 创建 Consul 客户端配置
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

	// 心跳
	interval := reg.CheckInterval
	if interval <= 0 {
		interval = constant.RegistryCheckInterval
	}
	// TTL 后自动注销
	deregister := reg.DeregisterAfter
	if deregister <= 0 {
		deregister = constant.RegistryDeregisterAfter
	}

	asr := &api.AgentServiceRegistration{
		ID:      reg.ID,      // 唯一ID
		Name:    reg.Service, // 服务名称（同一服务下可能有多个实例）
		Address: reg.Address,
		Port:    reg.Port,
		Tags:    reg.Tags, // 标签，用于标识环境/版本/分区等
		Meta:    reg.Meta, // 元信息，可存储额外属性，这里存储了url
		Check: &api.AgentServiceCheck{
			TCP:     fmt.Sprintf("%s", reg.Address),
			CheckID: reg.ID,
			// 心跳间隔
			Interval: interval.String(),
			// 自动注销时间
			DeregisterCriticalServiceAfter: deregister.String(),
		},
	}
	if err := cl.Agent().ServiceRegister(asr); err != nil {
		return nil, fmt.Errorf("consul register: %w", err)
	}
	return func() error {
		return cl.Agent().ServiceDeregister(asr.ID)
	}, nil
}
