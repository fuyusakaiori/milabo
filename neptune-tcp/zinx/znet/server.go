package znet

import (
	"fmt"
	"neptune-golang/neptune-tcp/zinx/utils"
	"neptune-golang/neptune-tcp/zinx/ziface"
	"net"
)

type Server struct {
	// 服务器名称
	Name string
	// IP 版本
	IPVersion string
	// IP 地址
	IP string
	// 端口号
	Port uint32
	// 路由器
	Router ziface.IRouter
	// 连接管理器
	ConnManager ziface.IConnManager
	// 钩子函数
	OnConnStart func(connection ziface.IConnection)
	OnConnStop  func(connection ziface.IConnection)
}

// Start 在方法名前声明接受者的方法, 是属于结构体方法
func (server *Server) Start() {
	// 最外层添加异步处理, 避免同步阻塞建立连接
	go func() {
		// 服务器正式启动
		fmt.Printf("[%s] Server Listener at IP :%s, Port :%d\n", server.Name, server.IP, server.Port)
		// 0. 启动线程池
		server.Router.StartWorkerPool()
		// 1. 获取 TCP 对象
		addr, err := net.ResolveTCPAddr(server.IPVersion, fmt.Sprintf("%s:%d", server.IP, server.Port))
		// 错误处理
		if err != nil {
			fmt.Println("resolve tcp addr error: ", err)
			return
		}

		// 2. 获取监听器对象
		listener, err := net.ListenTCP(server.IPVersion, addr)
		if err != nil {
			fmt.Println("listen ", server.IPVersion, " err", err)
			return
		}

		fmt.Println("start Zinx server", server.Name, " success, Listening...")
		// 3. 阻塞等待客户端的连接
		var connID uint32 = 0
		for {
			connID++
			connection, err := listener.AcceptTCP()
			if err != nil {
				fmt.Println("Accept err", err)
				continue
			}
			// 4 判断是否已经超过连接上限
			if server.ConnManager.GetConnectionCount() >= utils.Config.ZinxMaxConn {
				connection.Close()
				fmt.Println("[zinx] conn count already up to max ", utils.Config.ZinxMaxConn, ", must close some conn")
				continue
			}
			// 5. 处理业务逻辑: go 声明方法异步执行 协程
			go NewConn(connID, connection, server.Router, server).StartConn()
		}
	}()

}

func (server *Server) Serve() {
	// 1. 启动服务器
	server.Start()

	// TODO 服务器启动后的额外状态

	// 2. 阻塞服务器, 避免主进程结束导致整个服务器停止
	select {}
}

func (server *Server) Stop() {
	// 服务器关闭前释放相应的资源
	server.ConnManager.CloseConnections()
	fmt.Println("[zinx] server close, will release all connections")
}

func (server *Server) AddRouter(id uint32, handler ziface.IHandler) {
	server.Router.AddHandler(id, handler)
}

func (server *Server) GetConnManager() ziface.IConnManager {
	return server.ConnManager
}

func (server *Server) GetOnConnStart(connection ziface.IConnection) {
	if server.OnConnStart != nil {
		server.OnConnStart(connection)
	}
}

func (server *Server) GetOnConnStop(connection ziface.IConnection) {
	if server.OnConnStop != nil {
		server.OnConnStop(connection)
	}
}

func (server *Server) SetOnConnStart(onConnStart func(connection ziface.IConnection)) {
	server.OnConnStart = onConnStart
}

func (server *Server) SetOnConnStop(onConnStop func(connection ziface.IConnection)) {
	server.OnConnStop = onConnStop
}

// NewServer 1. 返回值是 IServer 2. 在方法名前没有声明接受者的, 属于公共的方法
func NewServer() ziface.IServer {
	// 变量的声明
	server := &Server{
		Name:        utils.Config.Name,
		IP:          utils.Config.IP,
		IPVersion:   utils.Config.IPVersion,
		Port:        utils.Config.Port,
		Router:      NewRouter(),
		ConnManager: NewConnManager(),
	}
	// 接口方法的入参是指针类型, 就需要传入地址, 所以对象需要取址
	return server
}
