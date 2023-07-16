package entity

import "time"

const (
	gatewayAdminTableName = "gateway_admin"
)

// Admin 管理员
type Admin struct {
	ID int `json:"id" gorm:"primary_key;admin_id" description:"管理员 ID"`
	// 管理员账号用户名
	Username string `json:"username" gorm:"column:admin_username" description:"管理员账号"`
	// 管理员账号密码
	Password string `json:"password" gorm:"column:admin_password" description:"管理员账号密码"`
	// 管理员账号密码加盐
	Salt string `json:"salt" gorm:"column:admin_salt" description:"管理员账号密码加盐"`
	// 创建时间戳
	CreateTime time.Time `json:"create_time" gorm:"column:create_time" description:"账号创建时间"`
	// 更新时间
	UpdateTime time.Time `json:"update_time" gorm:"column:update_time" description:"账号更新时间"`
	// 账号是否注销
	IsDelete int `json:"is_delete" gorm:"column:is_delete" description:"账号软删除"`
}

func (admin *Admin) TableName() string {
	return gatewayAdminTableName
}
