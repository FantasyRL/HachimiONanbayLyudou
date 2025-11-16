package application

import (
	"github.com/FantasyRL/go-mcp-demo/api/model/api"
	"github.com/FantasyRL/go-mcp-demo/pkg/gorm-gen/model"
)

func (h *Host) TemplateLogic(req *api.TemplateRequest) (*model.Users, error) {
	// 这里编排你的业务逻辑
	return h.templateRepository.CreateUserByIDAndName(h.ctx, req.TemplateId, "test")
}
