package main

import (
	"context"
	"flag"
	"github.com/FantasyRL/go-mcp-demo/api/handler/api"
	"github.com/FantasyRL/go-mcp-demo/api/router"
	"github.com/FantasyRL/go-mcp-demo/config"
	"github.com/FantasyRL/go-mcp-demo/pkg/constant"
	"github.com/FantasyRL/go-mcp-demo/pkg/logger"
	"github.com/FantasyRL/go-mcp-demo/pkg/utils"
	"time"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/flow"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/middlewares/server/recovery"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/hertz-contrib/cors"
	"github.com/hertz-contrib/gzip"
	"github.com/hertz-contrib/opensergo/sentinel/adapter"
)

var (
	serviceName = constant.ServiceNameAPI
	configPath  = flag.String("cfg", "config/config.yaml", "config file path")
)

func init() {
	flag.Parse()
	config.Load(*configPath, serviceName)
	logger.Init(serviceName, config.GetLoggerLevel())
	api.Init()
}

func main() {
	var err error

	// get available port from config set
	listenAddr, err := utils.GetAvailablePort()
	if err != nil {
		logger.Errorf("Api: get available port failed, err: %v", err)
		return
	}

	h := server.New(
		server.WithHostPorts(listenAddr),
		server.WithHandleMethodNotAllowed(true),
		server.WithMaxRequestBodySize(1<<31),
	)

	// Recovery
	h.Use(recovery.Recovery(recovery.WithRecoveryHandler(recoveryHandler)))

	// Cors
	h.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowCredentials: true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		MaxAge:           12 * time.Hour,
		ExposeHeaders:    []string{"Content-Length"},
	}))

	// gzip
	h.Use(gzip.Gzip(gzip.BestSpeed))

	// Sentinel
	initSentinel()
	h.Use(adapter.SentinelServerMiddleware(
		adapter.WithServerResourceExtractor(func(c context.Context, ctx *app.RequestContext) string {
			return "api"
		}),
		adapter.WithServerBlockFallback(func(ctx context.Context, c *app.RequestContext) {
			logger.Errorf("frequent requests have been rejected by the gateway. clientIP: %v\n", c.ClientIP())
			c.AbortWithStatusJSON(consts.StatusOK, map[string]interface{}{
				"code":    500,
				"message": "服务器当前处于请求高峰，请稍后再试",
			})
		}),
	))

	router.Register(h)
	h.Spin()
}

func recoveryHandler(ctx context.Context, c *app.RequestContext, err interface{}, stack []byte) {
	logger.Errorf("[Recovery] InternalServiceError err=%v\n stack=%s\n", err, stack)
	c.JSON(consts.StatusInternalServerError, map[string]interface{}{
		"code":    500,
		"message": "内部服务错误，请稍后再试",
	})
}

func initSentinel() {
	err := sentinel.InitDefault()
	if err != nil {
		logger.Fatalf("Unexpected error: %+v", err)
	}

	// limit QPS to 100
	_, err = flow.LoadRules([]*flow.Rule{
		{
			Resource:               "api",
			Threshold:              100,
			TokenCalculateStrategy: flow.Direct,
			ControlBehavior:        flow.Reject,
			StatIntervalInMs:       1000,
		},
	})
	if err != nil {
		logger.Fatalf("Unexpected error: %+v", err)
		return
	}
}
