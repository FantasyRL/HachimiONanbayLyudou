package registry

import (
	"context"
	"time"
)

// Resolver 抽象服务发现（可扩展 consul/etcd/...）
// 返回「全部」可用实例的完整 URL（如 http://127.0.0.1:8080/mcp）
type Resolver interface {
	// Resolve 解析服务名列表，返回可用实例的完整 URL 列表
	Resolve(services []string) (map[string][]string, error)
}

// Registrar 抽象服务注册
type Registrar interface {
	// Register 注册服务实例
	Register(ctx context.Context, r *Registration) (deregister func() error, err error)
}

// Registration 服务实例信息 通过此信息来注册服务
type Registration struct {
	Service string            // Service 名
	ID      string            // 实例 ID（唯一）
	Address string            // IP/Host
	Port    int               // 端口
	Tags    []string          // 标签（可选）
	Meta    map[string]string // 附加元信息（建议包含完整 URL）
	Scheme  string            // http/https（可选）
	Path    string            // 例如 /mcp（可选）

	// 健康检查（可选）
	CheckInterval   time.Duration
	DeregisterAfter time.Duration
}
