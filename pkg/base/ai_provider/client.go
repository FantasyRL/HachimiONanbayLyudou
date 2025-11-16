package ai_provider

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/FantasyRL/go-mcp-demo/config"
	"github.com/FantasyRL/go-mcp-demo/pkg/constant"
	"github.com/FantasyRL/go-mcp-demo/pkg/errno"
	"github.com/FantasyRL/go-mcp-demo/pkg/logger"
	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/option"
)

type Client struct {
	mode         string
	baseURL      string
	httpClient   *http.Client
	openaiClient *openai.Client
}

// 扩展可选覆盖参数（保持向后兼容：不传则回落到 config）
type ClientOptions struct {
	BaseURL       string
	RequestTimout time.Duration
	Mode          string // "local" | "remote"
	APIKey        string // remote 模式用
	Model         string // 兼容层中有时需要记录（此处不强制）
}

// NewAiProviderClient 支持可选覆盖参数，不传仍使用 config
func NewAiProviderClient(opts ...ClientOptions) *Client {
	var ov ClientOptions
	if len(opts) > 0 {
		ov = opts[0]
	}

	to := config.AiProvider.Options.RequestTimout
	if ov.RequestTimout > 0 {
		to = ov.RequestTimout
	}
	if to <= 0 {
		to = 60 * time.Second
	}

	mode := config.AiProvider.Mode
	if ov.Mode != "" {
		mode = ov.Mode
	}

	switch mode {
	case constant.AiProviderModeLocal:
		base := config.AiProvider.BaseURL
		if ov.BaseURL != "" {
			base = ov.BaseURL
		}
		// OpenAI 兼容前缀
		baseCompat := strings.TrimRight(base, "/") + "/v1"

		openaiCli := openai.NewClient(
			option.WithAPIKey("ollama"),
			option.WithBaseURL(baseCompat),
		)
		return &Client{
			mode:    constant.AiProviderModeLocal,
			baseURL: base, // Chat 接口直接用 base
			httpClient: &http.Client{
				Timeout: to,
				Transport: &http.Transport{
					Proxy: http.ProxyFromEnvironment,
					DialContext: (&net.Dialer{
						Timeout:   10 * time.Second,
						KeepAlive: 30 * time.Second,
					}).DialContext,
					ForceAttemptHTTP2:     true,
					MaxIdleConns:          100,
					IdleConnTimeout:       90 * time.Second,
					TLSHandshakeTimeout:   10 * time.Second,
					ExpectContinueTimeout: 1 * time.Second,
				},
			},
			openaiClient: &openaiCli,
		}
	case constant.AiProviderModeRemote:
		base := config.AiProvider.Remote.BaseURL
		key := config.AiProvider.Remote.APIKey
		if ov.BaseURL != "" {
			base = ov.BaseURL
		}
		if ov.APIKey != "" {
			key = ov.APIKey
		}
		openaiCli := openai.NewClient(
			option.WithAPIKey(key),
			option.WithBaseURL(base),
		)
		return &Client{
			mode:         constant.AiProviderModeRemote,
			baseURL:      base,
			openaiClient: &openaiCli,
		}
	default:
		logger.Errorf("unsupported mode: %s", mode)
		return nil
	}
}

// Chat 调用 /api/chat，非流式
func (c *Client) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	endpoint := fmt.Sprintf("%s/api/chat", c.baseURL)
	req.Stream = false

	b, _ := json.Marshal(req)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(b))
	if err != nil {
		logger.Errorf("ollama.Chat NewRequestWithContext error: %v", err)
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		logger.Errorf("ollama.Chat Do request error: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logger.Errorf("ollama.Chat error response: %s", string(body))
		return nil, fmt.Errorf("ollama chat failed: %s - %s", resp.Status, string(body))
	}
	var cr ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&cr); err != nil {
		return nil, err
	}
	return &cr, nil
}

// ChatStream api/chat，流式
func (c *Client) ChatStream(ctx context.Context, req ChatRequest, onChunk func(*ChatResponse) error) error {
	endpoint := fmt.Sprintf("%s/api/chat", c.baseURL)
	req.Stream = true

	b, _ := json.Marshal(req)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(b))
	if err != nil {
		logger.Errorf("ollama.ChatStream NewRequestWithContext error: %v", err)
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		logger.Errorf("ollama.ChatStream Do request error: %v", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		all, _ := io.ReadAll(resp.Body)
		logger.Errorf("ollama chat stream error response: %s", string(all))
		return fmt.Errorf("ollama chat stream failed: %s - %s", resp.Status, string(all))
	}

	sc := bufio.NewScanner(resp.Body)
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 10*1024*1024)

	for sc.Scan() {
		line := sc.Bytes()
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		chunk := new(ChatResponse)
		if err := json.Unmarshal(line, chunk); err != nil {
			logger.Errorf("ollama.ChatStream json unmarshal chunk error: %v", err)
			return err
		}
		if err := onChunk(chunk); err != nil {
			if errors.Is(err, errno.OllamaInternalStopStream) {
				return nil
			}
			return err
		}
		if chunk.Done {
			break
		}
	}
	return sc.Err()
}

// ChatStreamOpenAI 使用 OpenAI 兼容层流式聊天
func (c *Client) ChatStreamOpenAI(
	ctx context.Context,
	req openai.ChatCompletionNewParams,
	onChunk func(*openai.ChatCompletionChunk) error,
) error {
	stream := c.openaiClient.Chat.Completions.NewStreaming(ctx, req)
	defer stream.Close()
	for stream.Next() {
		chunk := stream.Current()
		if err := onChunk(&chunk); err != nil {
			if errors.Is(err, errno.OllamaInternalStopStream) {
				return nil
			}
			return err
		}
	}
	if err := stream.Err(); err != nil {
		logger.Errorf("openai.ChatStreamOpenAI stream error: %v", err)
		return err
	}
	return nil
}

func (c *Client) ChatOpenAI(
	ctx context.Context,
	req openai.ChatCompletionNewParams,
) (*openai.ChatCompletion, error) {
	resp, err := c.openaiClient.Chat.Completions.New(ctx, req)
	if err != nil {
		logger.Errorf("openai.ChatOpenAI error: %v", err)
		return nil, err
	}
	return resp, nil
}
