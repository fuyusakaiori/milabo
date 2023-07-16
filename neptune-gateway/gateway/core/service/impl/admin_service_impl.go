package impl

import (
	"errors"
	"github.com/fuyusakaiori/gateway/core/dao"
	"github.com/fuyusakaiori/gateway/core/dao/entity"
	"github.com/fuyusakaiori/gateway/core/dto"
	"github.com/fuyusakaiori/gateway/core/util/crypto"
	"github.com/gin-gonic/gin"
)

type AdminService struct {
}

var (
	adminMapper mapper.AdminMapper = mapper.AdminMapper{}
)

// AdminLogin 管理员登陆逻辑
func (service *AdminService) AdminLogin(context *gin.Context, request *dto.AdminLoginRequest) (*entity.Admin, error) {
	// 0. 获取请求中的参数
	username, password := request.Username, request.Password
	// 1. 获取管理员信息
	admin, err := adminMapper.AdminLogin(context, username, password)
	if admin == nil || err != nil {
		return nil, errors.New("管理员账号的用户名输入错误")
	}
	// 2. 判定密码是否正确
	if admin.Password != crypto.GenerateSaltPassword(admin.Salt, password) {
		return nil, errors.New("管理员账号的密码输入错误")
	}
	// 3. 返回信息
	return admin, nil
}
