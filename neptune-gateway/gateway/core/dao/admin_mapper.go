package mapper

import (
	"github.com/fuyusakaiori/gateway/core/dao/entity"
	"github.com/fuyusakaiori/gateway/core/util/golang_common/lib"
	"github.com/fuyusakaiori/gateway/core/util/log"
	"github.com/gin-gonic/gin"
)


type AdminMapper struct {
}

func (mapper *AdminMapper) AdminLogin(context *gin.Context, username string, password string) (*entity.Admin, error) {
	// 1. 获取数据库连接
	database, err := lib.GetGormPool("default")
	if err != nil {
		return nil, err
	}
	// 2. 查询管理员信息
	var admin entity.Admin
	if err := database.SetCtx(log.GetGinTraceContext(context)).
		Where(&entity.Admin{Username: username, Password: password}).Take(&admin).Error; err != nil {
		return nil, err
	}
	return &admin, nil
}
