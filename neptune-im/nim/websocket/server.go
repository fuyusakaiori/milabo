package websocket

import (
	"errors"
	"fmt"
	"github.com/gobwas/ws"
	"github.com/segmentio/ksuid"
	"github.com/sirupsen/logrus"
	"neptune-im/nim"
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

// Server 服务端实现
type Server struct {
	nim.Acceptor
	nim.ChannelMap
	nim.MessageListener
	nim.StateListener
	// ip:port
	address string
	options ServerOptions
	once    sync.Once
}


func NewServer(address string) nim.Server {
	return &Server{
		address: address,
	}
}

func (server *Server) Start() error {
	logger := logrus.WithFields(logrus.Fields{
		"module":   "websocket_server",
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
		server.ChannelMap = nim.NewChannelMap()
	}
	// 5. http 服务器绑定处理逻辑
	mux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		// 1. http 协议升级为 websocket 协议
		rawconn, _, _, err := ws.UpgradeHTTP(request, writer)
		if err != nil {

		}
		// 2. 封装连接
		conn := NewConn(rawconn)
		// 3. 建立连接
		channelId, err := server.Accept(conn, server.options.connectWait)
		if err != nil {
			// 3.1 如果建立连接失败, 回写连接建立失败的消息
			_ = conn.WriteFrame(nim.OpClose, []byte(err.Error()))
			// 3.2 关闭连接
			_ = conn.Close()
			return
		}
		// 4. 封装管道
		channel := nim.NewChannel(channelId, conn)
		channel.SetReadWait(server.options.readWait)
		channel.SetWriteWait(server.options.writeWait)
		// 5. 保存连接
		server.Put(channel)
		// 6. 启动协程异步读取管道中的消息
		go func(channel nim.Channel) {
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

func (server *Server) Shutdown() {
	logrus.WithFields(logrus.Fields{
		"module": "websocket_server",
		"address": server.address,
		"serverId": 0,
	})
	server.once.Do(func() {
		// 1. 获取所有连接
		channels := server.List()
		// 2. 遍历所有连接
		for _, channel := range channels {
			// 3. 移除连接
			server.Remove(channel.GetChannelID())
			// 4. 关闭连接
			_ = channel.Close()
		}
	})
}

func (server *Server) SetChannelMap(channelMap nim.ChannelMap) {
	server.ChannelMap = channelMap
}

func (server *Server) SetReadWait(timeout time.Duration) {
	server.options.readWait = readWaitTime
}

func (server *Server) SetAcceptor(acceptor nim.Acceptor) {
	server.Acceptor = acceptor
}

func (server *Server) SetStateListener(listener nim.StateListener) {
	server.StateListener = listener
}

func (server *Server) SendMessage(channelId string, message []byte) error {
	// 1. 获取连接
	if channel, ok := server.Get(channelId); ok {
		if err := channel.SendMessage(message); err != nil {
			return err
		}
		return nil
	}
	return errors.New(fmt.Sprintf("channel is not exist, channel id = %v", channelId))
}

// defaultAcceptor 连接器默认实现
type defaultAcceptor struct {
}

func newAcceptor() nim.Acceptor {
	return &defaultAcceptor{}
}

func (acceptor *defaultAcceptor) Accept(conn nim.Conn, timeout time.Duration) (string, error) {
	// 1. 生成连接 id
	return ksuid.New().String(), nil
}

