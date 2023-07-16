package dto

type AdminLoginRequest struct {
	// 管理员账号
	Username string `json:"username" form:"username" comment:"账号出错~" example:"admin" validate:"required"`
	// 管理员密码
	Password string `json:"password" form:"password" comment:"密码出错~" example:"123456" validate:"required"`
}

type AdminLoginResponse struct {

	Token string `json:"token" form:"token" comment:"token" example:"token" validate:""`
}
