package nim

import (
	"errors"
	"github.com/gobwas/ws"
	"github.com/sirupsen/logrus"
	"neptune-im/nim/websocket"
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
	options ServerOptions
	once    sync.Once
}

func (server *DefaultServer) SendMessage(channelId string, message []byte) error {
	//TODO implement me
	panic("implement me")
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
		server.ChannelMap = NewChannelMap()
	}
	// 5. http 服务器绑定处理逻辑
	mux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		// 1. http 协议升级为 websocket 协议
		rawconn, _, _, err := ws.UpgradeHTTP(request, writer)
		if err != nil {

		}
		// 2. 封装连接
		conn := websocket.NewConn(rawconn)
		// 3. 建立连接
		channelId, err := server.Accept(conn, server.options.connectWait)
		if err != nil {
			// 3.1 如果建立连接失败, 回写连接建立失败的消息
			_ = conn.WriteFrame(OpClose, []byte(err.Error()))
			// 3.2 关闭连接
			_ = conn.Close()
			return
		}
		// 4. 封装管道
		channel := NewChannel(channelId, conn)
		channel.SetReadWait(server.options.readWait)
		channel.SetWriteWait(server.options.writeWait)
		// 5. 保存连接
		server.Put(channel)
		// 6. 启动协程异步读取管道中的消息
		go func(channel Channel) {
			// 6.1 处理消息
			if err := channel.ReceiveMessage(server.MessageListener); err != nil {

			}
			// 6.2 处理消息出现异常就移除管道
			server.Remove(channel.GetChannelID())
			// 6.3 断开连接
			if err := server.Disconnect(channel.GetChannelID()); err != nil {

			}
			// 6.3 关闭管道
			_ = channel.Close()
		}(channel)
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
