package router

import (
	"github.com/FantasyRL/HachimiONanbayLyudou/api/handler"
	"github.com/cloudwego/hertz/pkg/app/server"
)

func customizedRegister(r *server.Hertz) {
	r.GET("/ping", handler.Ping)
}
