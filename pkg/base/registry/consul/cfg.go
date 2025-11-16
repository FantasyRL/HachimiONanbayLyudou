package consul

type ConsulConfig struct {
	Address    string // 127.0.0.1:8500
	Datacenter string // 可空
	Token      string // 可空
	Service    string // 要发现的服务名（必填）
	Tag        string // 可空，按 tag 过滤
	Scheme     string // http / https，默认 http
	Path       string // 例如 "/mcp"，默认 "/mcp"
}
