package repository

import (
	"context"
	"github.com/FantasyRL/go-mcp-demo/pkg/gorm-gen/model"
)

// TemplateRepository 根据实际需求定义上层访问的接口，在下层infra做具体方法的实现
type TemplateRepository interface {
	// CreateUserByIDAndName 这里定义一些示例方法，这样就可以先编排逻辑了，后面再去下层实现接口
	CreateUserByIDAndName(ctx context.Context, id string, name string) (*model.Users, error)
}
