package controller

import (
	"github.com/fuyusakaiori/gateway/core/dto"
	"github.com/fuyusakaiori/gateway/core/middleware"
	"github.com/fuyusakaiori/gateway/core/service/impl"
	"github.com/fuyusakaiori/gateway/core/util/param"
	"github.com/gin-gonic/gin"
	"net/http"
)

type AdminController struct {
}

var (
	adminService = impl.AdminService{}
)

func AdminRegister(group *gin.RouterGroup) {
	// 1. 创建 controller
	controller := &AdminController{}
	// 2. 绑定路由
	group.POST("/admin/login", controller.AdminLogin)
}

func (controller *AdminController) AdminLogin(context *gin.Context) {
	var req dto.AdminLoginRequest
	// 1. 获取参数
	if err := param.ShouldValidBind(context, &req); err != nil {
		middleware.ResponseError(context, http.StatusInternalServerError, err)
		return
	}
	// 2. 调用管理员登陆逻辑
	admin, err := adminService.AdminLogin(context, &req)
	if admin == nil || err != nil {
		middleware.ResponseError(context, http.StatusInternalServerError, err)
		return
	}
	// 3. 返回响应信息
	middleware.ResponseSuccess(context, &dto.AdminLoginResponse{Token: admin.Username})
}
