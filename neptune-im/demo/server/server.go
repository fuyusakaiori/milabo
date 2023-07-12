package server

import (
	"github.com/gobwas/ws"
	"net"
	"net/http"
	"sync"

	"github.com/sirupsen/logrus"
)

const (
	global string = "/"
	module string = "module"
	listen string = "listen"
	id     string = "id"
)

var (
	log    *logrus.Entry
	router *http.ServeMux
)

// Server 定义服务器
type Server struct {
	// 服务器 id
	id string
	// 服务器 名称
	name string
	// 服务器地址: ip + port
	address string
	// 服务器维护的连接: 字符串简单代替用户
	connections map[string]net.Conn
	// 锁同步
	mutex sync.Mutex
}

// NewServer 创建服务器实例
func NewServer(id, name, address string) *Server {
	return newServer(id, name, address)
}

func newServer(id, name, address string) *Server {
	return &Server{
		id:          id,
		name:        name,
		address:     address,
		connections: make(map[string]net.Conn),
	}
}

// Start 启动服务器实例
func (server *Server) Start() error {
	// 1. 创建请求路由器
	router = http.NewServeMux()
	// 2. 定义日志的属性
	log = logrus.WithFields(logrus.Fields{
		module: server.name,
		listen: server.address,
		id:     server.id,
	})
	// 3. 定义路径和对应的函数
	router.HandleFunc(global, func(writer http.ResponseWriter, request *http.Request) {
		// 3.1 建立 websocket 长连接
		websocketConn, _, _, err := ws.UpgradeHTTP(request, writer)
		if err != nil {
			log.Error("websocket create fail : %v", err)
			err := websocketConn.Close()
			if err != nil {
				log.Error("websocket close fail : %v", err)
				return
			}
			return
		}
		// 3.2 从请求中获取用户信息
		user := request.URL.Query().Get("user")
		// 3.3 判断用户信息是否为空
		if user == "" {
			log.Error("user is empty")
			if err := websocketConn.Close(); err != nil {
				log.Error("old websocket close fail : %v", err)
				return
			}
			return
		}
		// 3.4 维护用户的连接
		oldConn, exists := server.createConn(user, websocketConn)
		// 3.5 如果存在旧的连接就断开
		if exists {
			if err := oldConn.Close(); err != nil {
				log.Error("old websocket close fail : %v", err)
				return
			}
		}
		log.Infof("%s user connect to server", user)
		// 3.6 启动协程读取客户端发送的消息
		go func(user string, conn net.Conn) {
			// 3.6.1 读取消息
			if err := server.readMessage(user, conn); err != nil {
				log.Error("%s user conn read message occurred error : %v", err)
			}
			// 3.6.2 出现异常断开连接
			if err := websocketConn.Close(); err != nil {
				log.Error("websocket close fail in read message : %v", err)
			}
			// 3.6.3 删除用户连接
			server.removeConn(user)
			log.Infof("%s user conn close ", user)

		}(user, websocketConn)

	})

	defer log.Infof("server start succuss")
	// 4. 监听端口运行服务器
	return http.ListenAndServe(server.address, router)

}

func (server *Server) createConn(user string, conn net.Conn) (net.Conn, bool) {
	// 1. 上锁
	server.mutex.Lock()
	// 2. 释放锁
	defer server.mutex.Unlock()
	// 3. 存入新的连接
	oldConn, exists := server.connections[user]
	// 4. 放入新的连接
	server.connections[user] = conn
	log.Printf("user = %s, create connet to server\n", user)
	// 5.返回旧的连接
	return oldConn, exists
}

func (server *Server) removeConn(user string) {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	delete(server.connections, user)
}

func (server *Server) readMessage(user string, conn net.Conn) error {

	return nil
}
