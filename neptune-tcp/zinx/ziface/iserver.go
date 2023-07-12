package ziface

type IServer interface {
	// Start 启动服务器
	Start()
	// Serve 运行服务器
	Serve()
	// Stop 停止服务器
	Stop()
	// AddRouter 添加处理器
	AddRouter(id uint32, handler IHandler)
	// GetConnManager 获取连接管理器
	GetConnManager() IConnManager
	// GetOnConnStart 获取开始的钩子函数
	GetOnConnStart(connection IConnection)
	// GetOnConnStop 获取关闭的钩子函数
	GetOnConnStop(connection IConnection)
	// SetOnConnStart 设置开启的钩子函数
	SetOnConnStart(func(connection IConnection))
	// SetOnConnStop 设置关闭的钩子函数
	SetOnConnStop(func(connection IConnection))
}
