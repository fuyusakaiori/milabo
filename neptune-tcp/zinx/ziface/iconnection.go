package ziface

import "net"

// IConnection 客户端连接处理器
type IConnection interface {
	// StartConn 建立连接
	StartConn()
	// StopConn 断开连接
	StopConn()
	// GetTCPConn 获取连接
	GetTCPConn() *net.TCPConn
	// GetConnID 获取连接 ID
	GetConnID() uint32
	// RemoteAddr 获取客户端状态: 连接装填、IP 地址、端口号
	RemoteAddr() net.Addr
	// SendMessage 发送数据
	SendMessage(id uint32, data []byte) error
	// SetConnectionProperty 设置参数
	SetConnectionProperty(key string, value interface{})
	// GetConnectionProperty 获取参数
	GetConnectionProperty(key string) (value interface{})
	// RemoveConnectionProperty 移除参数
	RemoveConnectionProperty(key string)
}
