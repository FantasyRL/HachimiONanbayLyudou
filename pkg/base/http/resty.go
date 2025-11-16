package http

import (
	"github.com/go-resty/resty/v2"
	"time"
)

func InitHTTPClient(baseURL string, timeout time.Duration) *resty.Client {
	c := resty.New()
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	c.SetBaseURL(baseURL).
		SetTimeout(timeout).
		SetHeader("Accept", "text/html,application/xhtml+xml;q=0.9,*/*;q=0.1").
		SetHeader("User-Agent", "go-mcp/external-html/1.0")
	return c
}
