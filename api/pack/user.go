package pack

import (
	thrift "github.com/FantasyRL/go-mcp-demo/api/model/model"
	"github.com/FantasyRL/go-mcp-demo/pkg/gorm-gen/model"
)

func BuildUserResp(users *model.Users) *thrift.User {
	return &thrift.User{
		ID:   users.ID,
		Name: users.Name,
	}
}
