package nim

import (
	"errors"
	"github.com/sirupsen/logrus"
	"net/http"
	"sync"
	"time"
)

// ServerOptions 服务端配置
type ServerOptions struct {
	connectWait time.Duration
	readWait    time.Duration
	writeWait   time.Duration
}

// DefaultServer 服务端实现
type DefaultServer struct {
	Acceptor
	ChannelMap
	MessageListener
	StateListener
	// ip:port
	address string
	options sync.Once
	once    sync.Once
}

func NewServer() Server {
	return &DefaultServer{}
}

func (server *DefaultServer) Start() error {
	logger := logrus.WithFields(logrus.Fields{
		"module":   "websocket.server",
		"address":  server.address,
		"serverId": 0,
	})
	// 1. 实例化 http 服务端
	mux := http.NewServeMux()
	// 2. 设置状态监听器
	if server.StateListener == nil {
		return errors.New("state listener is nil")
	}
	// 3. 设置连接器
	if server.Acceptor == nil {
		server.Acceptor = newAcceptor()
	}
	// 4. 设置连接管理器
	if server.ChannelMap == nil {
		server.ChannelMap = newChannelMap()
	}
	// 5. http 服务器绑定处理逻辑
	mux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {

	})
	logger.Infof("server started")
	// 6. 启动监听
	return http.ListenAndServe(server.address, mux)
}

func (server *DefaultServer) Shutdown() {
	//TODO implement me
	panic("implement me")
}

func (server *DefaultServer) SetChannelMap(channelMap ChannelMap) {
	//TODO implement me
	panic("implement me")
}

func (server *DefaultServer) SetReadWait(timeout time.Duration) {
	//TODO implement me
	panic("implement me")
}

func (server *DefaultServer) SetAcceptor(acceptor Acceptor) {
	//TODO implement me
	panic("implement me")
}

func (server *DefaultServer) SetStateListener(listener StateListener) {
	//TODO implement me
	panic("implement me")
}

// defaultAcceptor 连接器默认实现
type defaultAcceptor struct {

}

func newAcceptor() Acceptor {
	return &defaultAcceptor{}
}

func (acceptor *defaultAcceptor) Accept(conn Conn, timeout time.Duration) (string, error) {
	//TODO implement me
	panic("implement me")
}

