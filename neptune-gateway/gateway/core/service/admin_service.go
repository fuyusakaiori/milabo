package service

import (
	"github.com/fuyusakaiori/gateway/core/dao/entity"
	"github.com/fuyusakaiori/gateway/core/dto"
	"github.com/gin-gonic/gin"
)

type IAdminService interface {
	// AdminLogin 管理员登陆逻辑
	AdminLogin(context *gin.Context, request *dto.AdminLoginRequest) (*entity.Admin, error)
}
