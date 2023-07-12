package ziface

// IRequest 请求
type IRequest interface {
	// GetMessage 获取数据
	GetMessage() IMessage
	// GetConn 获取连接
	GetConn() IConnection
}
